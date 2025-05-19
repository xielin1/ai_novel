package service

import (
	"fmt"
	"gin-template/common"
	"gin-template/model"
	"gin-template/repository/db"
)

// TokenService 定义了Token管理的核心接口
type TokenService interface {
	// CreditToken 给用户增加Token
	CreditToken(userID uint, amount int64, transactionUUID string, transactionType string, description string, relatedEntityType string, relatedEntityID string) (*model.UserToken, error)
	
	// DebitToken 扣除用户Token
	DebitToken(userID uint, amount int64, transactionUUID string, transactionType string, description string, relatedEntityType string, relatedEntityID string) (*model.UserToken, error)
	
	// GetBalance 获取用户当前Token余额
	GetBalance(userID uint) (int64, error)
	
	// GetUserToken 获取用户Token账户详情
	GetUserToken(userID uint) (*model.UserToken, error)
	
	// InitUserTokenAccount 初始化用户的Token账户
	InitUserTokenAccount(userID uint, initialBalance int64) (*model.UserToken, error)
	
	// GetTransactionByUUID 根据UUID获取交易记录
	GetTransactionByUUID(transactionUUID string) (*model.TokenTransaction, error)
	
	// GetUserTransactions 获取用户交易记录
	GetUserTransactions(userID uint, page, limit int) ([]model.TokenTransaction, int64, error)
}

// tokenServiceImpl 是 TokenService 的具体实现
type tokenServiceImpl struct{
	tokenRepo *db.TokenRepository
}

// 日志前缀，方便区分不同服务的日志
const tokenServiceLogPrefix = "[TokenService] "

// logInfo 记录信息日志
func logInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	common.SysLog(tokenServiceLogPrefix + message)
}

// logError 记录错误日志
func logError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	common.SysError(tokenServiceLogPrefix + message)
}

// NewTokenService 创建一个新的 TokenService 实例
func NewTokenService(tokenRepo *db.TokenRepository) TokenService {
	logInfo("初始化TokenService")
	return &tokenServiceImpl{
		tokenRepo: tokenRepo,
	}
}

// InitUserTokenAccount 初始化用户的Token账户
func (s *tokenServiceImpl) InitUserTokenAccount(userID uint, initialBalance int64) (*model.UserToken, error) {
	logInfo("初始化用户 %d 的Token账户，初始余额: %d", userID, initialBalance)
	userToken, err := s.tokenRepo.InitUserTokenAccount(userID, initialBalance)
	if err != nil {
		logError("初始化用户 %d 的Token账户失败: %v", userID, err)
		return nil, err
	}
	logInfo("用户 %d 的Token账户初始化成功，余额: %d", userID, userToken.Balance)
	return userToken, nil
}

// GetBalance 获取用户当前Token余额
func (s *tokenServiceImpl) GetBalance(userID uint) (int64, error) {
	balance, err := s.tokenRepo.GetTokenBalance(userID)
	if err != nil {
		logError("获取用户 %d 的Token余额失败: %v", userID, err)
		return 0, err
	}
	return balance, nil
}

// GetUserToken 获取用户Token账户详情
func (s *tokenServiceImpl) GetUserToken(userID uint) (*model.UserToken, error) {
	userToken, err := s.tokenRepo.GetUserToken(userID)
	if err != nil {
		logError("获取用户 %d 的Token账户失败: %v", userID, err)
		return nil, err
	}
	return userToken, nil
}

// GetTransactionByUUID 获取交易记录
func (s *tokenServiceImpl) GetTransactionByUUID(transactionUUID string) (*model.TokenTransaction, error) {
	return s.tokenRepo.GetTransactionByUUID(transactionUUID)
}

// CreditToken 给用户增加Token
func (s *tokenServiceImpl) CreditToken(userID uint, amount int64, transactionUUID string, transactionType string, description string, relatedEntityType string, relatedEntityID string) (*model.UserToken, error) {
	if amount <= 0 {
		logError("增加Token金额必须为正数，用户: %d, 金额: %d", userID, amount)
		return nil, fmt.Errorf("增加金额必须为正数")
	}
	
	logInfo("尝试为用户 %d 增加 %d Token, 交易ID: %s, 类型: %s", userID, amount, transactionUUID, transactionType)
	
	userToken, err := s.tokenRepo.CreditUserToken(userID, amount, transactionUUID, transactionType, description, relatedEntityType, relatedEntityID)
	if err != nil {
		logError("为用户 %d 增加Token失败: %v", userID, err)
		return nil, err
	}
	
	logInfo("用户 %d Token增加成功，新余额: %d", userID, userToken.Balance)
	return userToken, nil
}

// DebitToken 扣除用户Token
func (s *tokenServiceImpl) DebitToken(userID uint, amount int64, transactionUUID string, transactionType string, description string, relatedEntityType string, relatedEntityID string) (*model.UserToken, error) {
	if amount <= 0 {
		logError("扣减Token金额必须为正数，用户: %d, 金额: %d", userID, amount)
		return nil, fmt.Errorf("扣减金额必须为正数")
	}
	
	logInfo("尝试从用户 %d 扣减 %d Token, 交易ID: %s, 类型: %s", userID, amount, transactionUUID, transactionType)
	
	userToken, err := s.tokenRepo.DebitUserToken(userID, amount, transactionUUID, transactionType, description, relatedEntityType, relatedEntityID)
	if err != nil {
		logError("从用户 %d 扣减Token失败: %v", userID, err)
		return nil, err
	}
	
	logInfo("用户 %d Token扣减成功，新余额: %d", userID, userToken.Balance)
	return userToken, nil
}

// GetUserTransactions 获取用户交易记录
func (s *tokenServiceImpl) GetUserTransactions(userID uint, page, limit int) ([]model.TokenTransaction, int64, error) {
	logInfo("获取用户 %d 的交易记录，页码: %d, 每页数量: %d", userID, page, limit)
	transactions, total, err := s.tokenRepo.GetUserTokenTransactions(userID, page, limit)
	if err != nil {
		logError("获取用户 %d 的交易记录失败: %v", userID, err)
		return nil, 0, err
	}
	logInfo("成功获取用户 %d 的交易记录，共 %d 条", userID, total)
	return transactions, total, nil
} 