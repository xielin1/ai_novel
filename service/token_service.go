package service

import (
	"fmt"
	"gin-template/model"
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
type tokenServiceImpl struct{}

// NewTokenService 创建一个新的 TokenService 实例
func NewTokenService() TokenService {
	return &tokenServiceImpl{}
}

// InitUserTokenAccount 初始化用户的Token账户
func (s *tokenServiceImpl) InitUserTokenAccount(userID uint, initialBalance int64) (*model.UserToken, error) {
	return model.InitUserTokenAccount(userID, initialBalance)
}

// GetBalance 获取用户当前Token余额
func (s *tokenServiceImpl) GetBalance(userID uint) (int64, error) {
	return model.GetTokenBalance(userID)
}

// GetUserToken 获取用户Token账户详情
func (s *tokenServiceImpl) GetUserToken(userID uint) (*model.UserToken, error) {
	return model.GetUserToken(userID)
}

// GetTransactionByUUID 获取交易记录
func (s *tokenServiceImpl) GetTransactionByUUID(transactionUUID string) (*model.TokenTransaction, error) {
	return model.GetTransactionByUUID(transactionUUID)
}

// CreditToken 给用户增加Token
func (s *tokenServiceImpl) CreditToken(userID uint, amount int64, transactionUUID string, transactionType string, description string, relatedEntityType string, relatedEntityID string) (*model.UserToken, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("增加金额必须为正数")
	}
	
	return model.CreditUserToken(userID, amount, transactionUUID, transactionType, description, relatedEntityType, relatedEntityID)
}

// DebitToken 扣除用户Token
func (s *tokenServiceImpl) DebitToken(userID uint, amount int64, transactionUUID string, transactionType string, description string, relatedEntityType string, relatedEntityID string) (*model.UserToken, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("扣减金额必须为正数")
	}
	
	return model.DebitUserToken(userID, amount, transactionUUID, transactionType, description, relatedEntityType, relatedEntityID)
}

// GetUserTransactions 获取用户交易记录
func (s *tokenServiceImpl) GetUserTransactions(userID uint, page, limit int) ([]model.TokenTransaction, int64, error) {
	return model.GetUserTokenTransactions(userID, page, limit)
} 