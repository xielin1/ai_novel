package service

import (
	"fmt"
	"gin-template/model"
	"log"
	"sync"
	"time"
)

// TokenReconciliationService Token对账服务
type TokenReconciliationService struct {
	running   bool
	mutex     sync.Mutex
	stopChan  chan struct{}
	interval  time.Duration // 对账间隔
	batchSize int           // 每批处理的用户数量
}

// NewTokenReconciliationService 创建新的Token对账服务
func NewTokenReconciliationService(interval time.Duration, batchSize int) *TokenReconciliationService {
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
	
	log.Printf("Token对账服务已启动，间隔：%v\n", s.interval)
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
	
	log.Println("Token对账服务已停止")
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
	log.Println("开始执行Token对账...")
	startTime := time.Now()
	
	// 1. 获取所有用户ID
	userIDs, err := s.getAllUserIDs()
	if err != nil {
		log.Printf("获取用户列表失败: %v", err)
		return
	}
	
	log.Printf("共找到 %d 个用户账户需要对账", len(userIDs))
	
	// 2. 分批处理用户
	var wg sync.WaitGroup
	discrepancies := make(chan string, 100) // 存储发现的不匹配记录
	
	// 收集不匹配信息的goroutine
	go func() {
		for discrepancy := range discrepancies {
			log.Printf("对账不匹配: %s", discrepancy)
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
	log.Printf("Token对账完成，耗时: %v", duration)
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
		userToken, err := model.GetUserToken(userID)
		if err != nil {
			log.Printf("获取用户 %d 的Token余额失败: %v", userID, err)
			continue
		}
		
		currentBalance := userToken.Balance
		
		// 2. 计算用户交易记录总和
		calculatedBalance, err := s.calculateUserBalanceFromTransactions(userID)
		if err != nil {
			log.Printf("计算用户 %d 的交易记录总和失败: %v", userID, err)
			continue
		}
		
		// 3. 比较余额
		if currentBalance != calculatedBalance {
			// 构造不匹配信息
			discrepancy := fmt.Sprintf("用户 %d 的余额不匹配: 当前余额=%d, 计算余额=%d, 差额=%d", 
				userID, currentBalance, calculatedBalance, currentBalance-calculatedBalance)
			
			// 输出到日志
			log.Printf("对账发现: %s", discrepancy)
			
			// 发送到channel，可能用于告警
			discrepancies <- discrepancy
			
			// 保存到数据库
			description := fmt.Sprintf("系统定期对账发现差异: 当前余额=%d, 基于交易记录计算余额=%d", 
				currentBalance, calculatedBalance)
			
			record, err := model.SaveReconciliationRecord(userID, currentBalance, calculatedBalance, description)
			if err != nil {
				log.Printf("保存用户 %d 的对账记录失败: %v", userID, err)
			} else {
				log.Printf("已保存用户 %d 的对账记录，ID=%d", userID, record.ID)
			}
			
			// 可选: 自动修复不匹配
			if autoFix := false; autoFix { // 设为true启用自动修复
				if s.fixBalanceDiscrepancy(userID, currentBalance, calculatedBalance) {
					// 如果修复成功，更新对账记录为已修复
					if record != nil {
						err := model.UpdateReconciliationRecordAsFixed(record.ID)
						if err != nil {
							log.Printf("更新对账记录状态失败: %v", err)
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
	log.Printf("开始修复用户 %d 的余额不匹配: 当前=%d, 计算=%d", userID, currentBalance, calculatedBalance)
	
	// 创建一个调整交易记录
	diff := calculatedBalance - currentBalance
	
	// 生成唯一的交易UUID
	transactionUUID := fmt.Sprintf("recon-%d-%d", userID, time.Now().Unix())
	transactionType := "reconciliation_adjustment"
	description := fmt.Sprintf("系统对账调整: %d -> %d", currentBalance, calculatedBalance)
	
	// 使用TokenService执行调整
	// 如果是增加余额
	if diff > 0 {
		_, err := tokenService.CreditToken(userID, diff, transactionUUID, transactionType, description, "system", "reconciliation")
		if err != nil {
			log.Printf("修复用户 %d 余额失败 (增加): %v", userID, err)
			return false
		}
	} else if diff < 0 { // 如果是减少余额
		_, err := tokenService.DebitToken(userID, -diff, transactionUUID, transactionType, description, "system", "reconciliation")
		if err != nil {
			log.Printf("修复用户 %d 余额失败 (减少): %v", userID, err)
			return false
		}
	}
	
	log.Printf("已修复用户 %d 的余额不匹配", userID)
	return true
}

// ReconcileAllTokens 手动触发全量对账
func ReconcileAllTokens() error {
	service := NewTokenReconciliationService(0, 0) // 使用默认批处理大小
	service.performReconciliation()
	return nil
}

// ReconcileUserToken 对特定用户进行对账
func ReconcileUserToken(userID uint) error {
	service := NewTokenReconciliationService(0, 0)
	
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
	service.reconcileUserBatch([]uint{userID}, discrepancies)
	close(discrepancies)
	
	if len(issues) > 0 {
		return fmt.Errorf("发现对账问题: %v", issues)
	}
	
	return nil
}

// 以下代码用于应用启动时初始化对账服务
var reconciliationService *TokenReconciliationService

// InitReconciliationService 初始化Token对账服务
func InitReconciliationService() {
	// 确保TokenReconciliationRecord表存在
	err := model.DB.AutoMigrate(&model.TokenReconciliationRecord{})
	if err != nil {
		log.Printf("迁移TokenReconciliationRecord表失败: %v", err)
	}
	
	// 每天凌晨3点执行对账
	interval := 24 * time.Hour
	batchSize := 500
	
	reconciliationService = NewTokenReconciliationService(interval, batchSize)
	reconciliationService.Start()
}

// StopReconciliationService 停止Token对账服务
func StopReconciliationService() {
	if reconciliationService != nil {
		reconciliationService.Stop()
	}
} 