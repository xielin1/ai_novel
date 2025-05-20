package service

import (
	"fmt"
	"gin-template/common"
	"gin-template/model"
	"gin-template/repository"
	"sync"
	"time"
)

// 日志前缀，方便区分不同服务的日志
const reconciliationLogPrefix = "[TokenReconciliation] "

// 记录信息日志
func reconciliationLogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	common.SysLog(reconciliationLogPrefix + message)
}

// 记录错误日志
func reconciliationLogError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	common.SysError(reconciliationLogPrefix + message)
}

// TokenReconciliationService Token对账服务
type TokenReconciliationService struct {
	running   bool
	mutex     sync.Mutex
	stopChan  chan struct{}
	interval  time.Duration // 对账间隔
	batchSize int           // 每批处理的用户数量
	tokenRepo *repository.TokenRepository
	reconRepo *repository.TokenReconciliationRepository
}

// NewTokenReconciliationService 创建新的Token对账服务
func NewTokenReconciliationService(interval time.Duration, batchSize int, tokenRepo *repository.TokenRepository, reconRepo *repository.TokenReconciliationRepository) *TokenReconciliationService {
	if interval < time.Minute {
		interval = time.Hour // 默认每小时对账一次
	}

	if batchSize <= 0 {
		batchSize = 100 // 默认每批处理100个用户
	}

	return &TokenReconciliationService{
		running:   false,
		mutex:     sync.Mutex{},
		stopChan:  make(chan struct{}),
		interval:  interval,
		batchSize: batchSize,
		tokenRepo: tokenRepo,
		reconRepo: reconRepo,
	}
}

// Start 启动定期对账服务
func (s *TokenReconciliationService) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return
	}

	s.running = true
	s.stopChan = make(chan struct{})

	go s.reconciliationLoop()

	reconciliationLogInfo("Token对账服务已启动，间隔：%v", s.interval)
}

// Stop 停止定期对账服务
func (s *TokenReconciliationService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopChan)

	reconciliationLogInfo("Token对账服务已停止")
}

// reconciliationLoop 对账循环
func (s *TokenReconciliationService) reconciliationLoop() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// 启动后立即执行一次
	s.performReconciliation()

	for {
		select {
		case <-ticker.C:
			s.performReconciliation()
		case <-s.stopChan:
			return
		}
	}
}

// performReconciliation 执行一次完整的对账操作
func (s *TokenReconciliationService) performReconciliation() {
	reconciliationLogInfo("开始执行Token对账...")
	startTime := time.Now()

	// 1. 获取所有用户ID
	userIDs, err := s.getAllUserIDs()
	if err != nil {
		reconciliationLogError("获取用户列表失败: %v", err)
		return
	}

	reconciliationLogInfo("共找到 %d 个用户账户需要对账", len(userIDs))

	// 2. 分批处理用户
	var wg sync.WaitGroup
	discrepancies := make(chan string, 100) // 存储发现的不匹配记录

	// 收集不匹配信息的goroutine
	go func() {
		for discrepancy := range discrepancies {
			reconciliationLogError("对账不匹配: %s", discrepancy)
			// 这里可以添加将不匹配信息写入数据库或发送告警的代码
		}
	}()

	// 分批处理用户
	for i := 0; i < len(userIDs); i += s.batchSize {
		end := i + s.batchSize
		if end > len(userIDs) {
			end = len(userIDs)
		}

		userBatch := userIDs[i:end]
		wg.Add(1)

		go func(users []uint) {
			defer wg.Done()
			s.reconcileUserBatch(users, discrepancies)
		}(userBatch)
	}

	wg.Wait()
	close(discrepancies)

	duration := time.Since(startTime)
	reconciliationLogInfo("Token对账完成，耗时: %v", duration)
}

// getAllUserIDs 获取所有拥有Token账户的用户ID
func (s *TokenReconciliationService) getAllUserIDs() ([]uint, error) {
	var userIDs []uint
	err := model.DB.Model(&model.UserToken{}).Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// reconcileUserBatch 对一批用户进行对账
func (s *TokenReconciliationService) reconcileUserBatch(userIDs []uint, discrepancies chan<- string) {
	for _, userID := range userIDs {
		// 1. 获取用户当前Token余额
		userToken, err := s.tokenRepo.GetUserToken(userID)
		if err != nil {
			reconciliationLogError("获取用户 %d 的Token余额失败: %v", userID, err)
			continue
		}

		currentBalance := userToken.Balance

		// 2. 计算用户交易记录总和
		calculatedBalance, err := s.calculateUserBalanceFromTransactions(userID)
		if err != nil {
			reconciliationLogError("计算用户 %d 的交易记录总和失败: %v", userID, err)
			continue
		}

		// 3. 比较余额
		if currentBalance != calculatedBalance {
			// 构造不匹配信息
			discrepancy := fmt.Sprintf("用户 %d 的余额不匹配: 当前余额=%d, 计算余额=%d, 差额=%d",
				userID, currentBalance, calculatedBalance, currentBalance-calculatedBalance)

			// 输出到日志
			reconciliationLogError("对账发现: %s", discrepancy)

			// 发送到channel，可能用于告警
			discrepancies <- discrepancy

			// 保存到数据库
			description := fmt.Sprintf("系统定期对账发现差异: 当前余额=%d, 基于交易记录计算余额=%d",
				currentBalance, calculatedBalance)

			record, err := s.reconRepo.SaveReconciliationRecord(userID, currentBalance, calculatedBalance, description)
			if err != nil {
				reconciliationLogError("保存用户 %d 的对账记录失败: %v", userID, err)
			} else {
				reconciliationLogInfo("已保存用户 %d 的对账记录，ID=%d", userID, record.ID)
			}

			// 可选: 自动修复不匹配
			if autoFix := false; autoFix { // 设为true启用自动修复
				if s.fixBalanceDiscrepancy(userID, currentBalance, calculatedBalance) {
					// 如果修复成功，更新对账记录为已修复
					if record != nil {
						err := s.reconRepo.UpdateReconciliationRecordAsFixed(record.ID)
						if err != nil {
							reconciliationLogError("更新对账记录状态失败: %v", err)
						}
					}
				}
			}
		}
	}
}

// calculateUserBalanceFromTransactions 根据交易记录计算用户余额
func (s *TokenReconciliationService) calculateUserBalanceFromTransactions(userID uint) (int64, error) {
	var totalAmount int64

	// 这里我们不使用分页，直接获取所有记录
	// 对于交易记录特别多的用户，可能需要修改为分批查询
	var transactions []model.TokenTransaction
	err := model.DB.Where("user_id = ? AND status = ?", userID, "completed").Find(&transactions).Error
	if err != nil {
		return 0, err
	}

	for _, tx := range transactions {
		totalAmount += tx.Amount
	}

	return totalAmount, nil
}

// fixBalanceDiscrepancy 修复用户余额不匹配
// 警告: 谨慎启用此功能，最好在确认有问题的情况下手动修复
func (s *TokenReconciliationService) fixBalanceDiscrepancy(userID uint, currentBalance, calculatedBalance int64) bool {
	reconciliationLogInfo("开始修复用户 %d 的余额不匹配: 当前=%d, 计算=%d", userID, currentBalance, calculatedBalance)

	// 创建一个调整交易记录
	diff := calculatedBalance - currentBalance

	// 生成唯一的交易UUID
	transactionUUID := fmt.Sprintf("recon-%d-%d", userID, time.Now().Unix())
	transactionType := "reconciliation_adjustment"
	description := fmt.Sprintf("系统对账调整: %d -> %d", currentBalance, calculatedBalance)

	// 使用TokenService执行调整
	// 如果是增加余额
	if diff > 0 {
		_, err := GetTokenService().CreditToken(userID, diff, transactionUUID, transactionType, description, "system", "reconciliation")
		if err != nil {
			reconciliationLogError("修复用户 %d 余额失败 (增加): %v", userID, err)
			return false
		}
	} else if diff < 0 { // 如果是减少余额
		_, err := GetTokenService().DebitToken(userID, -diff, transactionUUID, transactionType, description, "system", "reconciliation")
		if err != nil {
			reconciliationLogError("修复用户 %d 余额失败 (减少): %v", userID, err)
			return false
		}
	}

	reconciliationLogInfo("已修复用户 %d 的余额不匹配", userID)
	return true
}

// 定义一个方便调用的全局对账服务实例
var tokenReconciliationService *TokenReconciliationService

// InitReconciliationService 初始化Token对账服务
func InitReconciliationService(tokenRepo *repository.TokenRepository, reconRepo *repository.TokenReconciliationRepository) {
	// 确保TokenReconciliationRecord表存在
	err := reconRepo.EnsureTokenReconciliationTable()
	if err != nil {
		reconciliationLogError("迁移TokenReconciliationRecord表失败: %v", err)
	}

	// 每天凌晨3点执行对账
	interval := 24 * time.Hour
	batchSize := 500

	tokenReconciliationService = NewTokenReconciliationService(interval, batchSize, tokenRepo, reconRepo)
	tokenReconciliationService.Start()

	reconciliationLogInfo("Token对账服务初始化完成，设置为每 %v 执行一次，批处理大小: %d", interval, batchSize)
}

// StopReconciliationService 停止Token对账服务
func StopReconciliationService() {
	if tokenReconciliationService != nil {
		tokenReconciliationService.Stop()
		reconciliationLogInfo("Token对账服务已停止")
	}
}

// ReconcileAllTokens 手动触发全量对账
func ReconcileAllTokens() error {
	if tokenReconciliationService == nil {
		return fmt.Errorf("Token对账服务尚未初始化")
	}

	reconciliationLogInfo("手动触发全量对账")
	tokenReconciliationService.performReconciliation()
	return nil
}

// ReconcileUserToken 对特定用户进行对账
func ReconcileUserToken(userID uint) error {
	if tokenReconciliationService == nil {
		return fmt.Errorf("Token对账服务尚未初始化")
	}

	reconciliationLogInfo("对用户 %d 进行手动对账", userID)

	// 模拟批处理，但只包含一个用户
	discrepancies := make(chan string, 10)

	// 收集不匹配信息
	var issues []string
	go func() {
		for d := range discrepancies {
			issues = append(issues, d)
		}
	}()

	// 执行对账
	tokenReconciliationService.reconcileUserBatch([]uint{userID}, discrepancies)
	close(discrepancies)

	if len(issues) > 0 {
		reconciliationLogError("用户 %d 对账发现问题: %v", userID, issues)
		return fmt.Errorf("发现对账问题: %v", issues)
	}

	reconciliationLogInfo("用户 %d 对账完成，未发现问题", userID)
	return nil
}
