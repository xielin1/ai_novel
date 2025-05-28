package service

import (
	"fmt"
	"gin-template/common"
	"gin-template/define"
	"gin-template/model"
	"gin-template/repository"
	"gin-template/util"
	"sync"
	"time"
)

type TokenReconciliationService struct {
	running   bool
	mutex     sync.Mutex
	stopChan  chan struct{}
	interval  time.Duration // 对账间隔
	batchSize int           // 每批处理的用户数量
	tokenRepo *repository.TokenRepository
	reconRepo *repository.TokenReconciliationRepository
}

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

	// 直接调用common.SysLog，移除日志前缀
	common.SysLog(fmt.Sprintf("Token对账服务已启动，间隔：%v", s.interval))
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

	common.SysLog("Token对账服务已停止")
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
	common.SysLog("开始执行Token对账...")
	startTime := time.Now()

	// 1. 获取所有用户ID
	userIDs, err := s.getAllUserIDs()
	if err != nil {
		common.SysError(fmt.Sprintf("获取用户列表失败: %v", err))
		return
	}

	common.SysLog(fmt.Sprintf("共找到 %d 个用户账户需要对账", len(userIDs)))

	// 2. 分批处理用户
	var wg sync.WaitGroup
	discrepancies := make(chan string, 100) // 存储发现的不匹配记录

	// 收集不匹配信息的goroutine
	go func() {
		for discrepancy := range discrepancies {
			common.SysError(discrepancy) // 直接输出错误日志
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

		go func(users []int64) {
			defer wg.Done()
			s.reconcileUserBatch(users, discrepancies)
		}(userBatch)
	}

	wg.Wait()
	close(discrepancies)

	duration := time.Since(startTime)
	common.SysLog(fmt.Sprintf("Token对账完成，耗时: %v", duration))
}

// getAllUserIDs 获取所有拥有Token账户的用户ID
func (s *TokenReconciliationService) getAllUserIDs() ([]int64, error) {
	var userIDs []int64
	err := model.DB.Model(&model.UserToken{}).Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// reconcileUserBatch 对一批用户进行对账
func (s *TokenReconciliationService) reconcileUserBatch(userIDs []int64, discrepancies chan<- string) {
	for _, userID := range userIDs {
		// 1. 获取用户当前Token余额
		userToken, err := s.tokenRepo.GetUserToken(userID)
		if err != nil {
			common.SysError(fmt.Sprintf("获取用户 %d 的Token余额失败: %v", userID, err))
			continue
		}

		currentBalance := userToken.Balance

		// 2. 计算用户交易记录总和
		calculatedBalance, err := s.calculateUserBalanceFromTransactions(userID)
		if err != nil {
			common.SysError(fmt.Sprintf("计算用户 %d 的交易记录总和失败: %v", userID, err))
			continue
		}

		// 3. 比较余额
		if currentBalance != calculatedBalance {
			// 构造不匹配信息
			discrepancy := fmt.Sprintf("对账发现: 用户 %d 的余额不匹配: 当前余额=%d, 计算余额=%d, 差额=%d",
				userID, currentBalance, calculatedBalance, currentBalance-calculatedBalance)

			discrepancies <- discrepancy // 发送到错误日志收集channel

			// 保存到数据库
			description := fmt.Sprintf("系统定期对账发现差异: 当前余额=%d, 基于交易记录计算余额=%d",
				currentBalance, calculatedBalance)

			record, err := s.reconRepo.SaveReconciliationRecord(userID, currentBalance, calculatedBalance, description)
			if err != nil {
				common.SysError(fmt.Sprintf("保存用户 %d 的对账记录失败: %v", userID, err))
			} else {
				common.SysLog(fmt.Sprintf("已保存用户 %d 的对账记录，ID=%d", userID, record.ID))
			}

			// 可选: 自动修复不匹配,todo 设置成配置项
			if autoFix := true; autoFix { // 设为true启用自动修复
				if s.fixBalanceDiscrepancy(userID, currentBalance, calculatedBalance) {
					// 如果修复成功，更新对账记录为已修复
					if record != nil {
						err := s.reconRepo.UpdateReconciliationRecordAsFixed(record.ID)
						if err != nil {
							common.SysError(fmt.Sprintf("更新对账记录状态失败: %v", err))
						}
					}
				}
			}
		}

		common.SysLog(fmt.Sprintf("用户 %d 的余额已对账完成", userID))
	}
}

// calculateUserBalanceFromTransactions 根据交易记录计算用户余额
func (s *TokenReconciliationService) calculateUserBalanceFromTransactions(userID int64) (int64, error) {
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
func (s *TokenReconciliationService) fixBalanceDiscrepancy(userID int64, currentBalance, calculatedBalance int64) bool {
	common.SysLog(fmt.Sprintf("开始修复用户 %d 的余额不匹配: 当前=%d, 计算=%d", userID, currentBalance, calculatedBalance))

	// 创建一个调整交易记录
	diff := calculatedBalance - currentBalance

	// 生成唯一的交易UUID

	transactionUUID := util.GetUUIDGenerator().Generate(util.BusinessReconciliation)
	transactionType := define.TokenTransactionTypeReconciliationAdjustment
	description := fmt.Sprintf("系统对账调整: %d -> %d", currentBalance, calculatedBalance)

	// 使用TokenService执行调整
	// 如果是增加余额
	if diff > 0 {
		_, err := GetTokenService().CreditToken(userID, diff, transactionUUID, transactionType, description, "system", "reconciliation")
		if err != nil {
			common.SysError(fmt.Sprintf("修复用户 %d 余额失败 (增加): %v", userID, err))
			return false
		}
	} else if diff < 0 { // 如果是减少余额
		_, err := GetTokenService().DebitToken(userID, -diff, transactionUUID, transactionType, description, "system", "reconciliation")
		if err != nil {
			common.SysError(fmt.Sprintf("修复用户 %d 余额失败 (减少): %v", userID, err))
			return false
		}
	}

	common.SysLog(fmt.Sprintf("已修复用户 %d 的余额不匹配", userID))
	return true
}

var tokenReconciliationService *TokenReconciliationService

func GetTokenReconciliation() *TokenReconciliationService {
	//todo 优化下，单例模式
	return tokenReconciliationService
}

// InitReconciliationService 初始化Token对账服务
func InitReconciliationService(tokenRepo *repository.TokenRepository, reconRepo *repository.TokenReconciliationRepository) {
	// 确保TokenReconciliationRecord表存在
	err := reconRepo.EnsureTokenReconciliationTable()
	if err != nil {
		common.SysError(fmt.Sprintf("迁移TokenReconciliationRecord表失败: %v", err))
	}

	// 每天凌晨3点执行对账
	interval := 24 * time.Hour
	batchSize := 500

	tokenReconciliationService = NewTokenReconciliationService(interval, batchSize, tokenRepo, reconRepo)
	tokenReconciliationService.Start()

	common.SysLog(fmt.Sprintf("Token对账服务初始化完成，设置为每 %v 执行一次，批处理大小: %d", interval, batchSize))
}

// StopReconciliationService 停止Token对账服务
func StopReconciliationService() {
	if tokenReconciliationService != nil {
		tokenReconciliationService.Stop()
		common.SysLog("Token对账服务已停止")
	}
}

// ReconcileAllTokens 手动触发全量对账
func ReconcileAllTokens() error {
	if tokenReconciliationService == nil {
		return fmt.Errorf("token对账服务尚未初始化")
	}

	common.SysLog("手动触发全量对账")
	tokenReconciliationService.performReconciliation()
	return nil
}

// ReconcileUserToken 对特定用户进行对账
func ReconcileUserToken(userID int64) error {
	if tokenReconciliationService == nil {
		return fmt.Errorf("token对账服务尚未初始化")
	}

	common.SysLog(fmt.Sprintf("对用户 %d 进行手动对账", userID))

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
	tokenReconciliationService.reconcileUserBatch([]int64{userID}, discrepancies)
	close(discrepancies)

	if len(issues) > 0 {
		common.SysError(fmt.Sprintf("用户 %d 对账发现问题: %v", userID, issues))
		return fmt.Errorf("发现对账问题: %v", issues)
	}

	common.SysLog(fmt.Sprintf("用户 %d 对账完成，未发现问题", userID))
	return nil
}
