package service

import (
	"fmt"
	"gin-template/common"
	"gin-template/model"
	"gin-template/repository"
)

type TokenService struct {
	tokenRepo *repository.TokenRepository
}

const tokenServiceLogPrefix = "[TokenService] "

var tokenService *TokenService

func SetTokenService(service *TokenService) {
	tokenService = service
	common.SysLog(tokenServiceLogPrefix + "TokenService has been set via dependency injection")
}

func GetTokenService() *TokenService {
	return tokenService
}

func NewTokenService(tokenRepo *repository.TokenRepository) *TokenService {
	return &TokenService{
		tokenRepo: tokenRepo,
	}
}

// InitUserTokenAccount initializes a user's token account
func (s *TokenService) InitUserTokenAccount(userID int64, initialBalance int64) (*model.UserToken, error) {
	common.SysLog(tokenServiceLogPrefix + fmt.Sprintf("Initializing token account for user %d with initial balance: %d", userID, initialBalance))
	userToken, err := s.tokenRepo.InitUserTokenAccount(userID, initialBalance)
	if err != nil {
		common.SysError(tokenServiceLogPrefix + fmt.Sprintf("Failed to initialize token account for user %d: %v", userID, err))
		return nil, err
	}
	common.SysLog(tokenServiceLogPrefix + fmt.Sprintf("Token account initialized successfully for user %d, balance: %d", userID, userToken.Balance))
	return userToken, nil
}

// GetBalance retrieves the current token balance of a user
func (s *TokenService) GetBalance(userID int64) (int64, error) {
	balance, err := s.tokenRepo.GetTokenBalance(userID)
	if err != nil {
		common.SysError(tokenServiceLogPrefix + fmt.Sprintf("Failed to get token balance for user %d: %v", userID, err))
		return 0, err
	}
	return balance, nil
}

// GetUserToken retrieves the token account details of a user
func (s *TokenService) GetUserToken(userID int64) (*model.UserToken, error) {
	userToken, err := s.tokenRepo.GetUserToken(userID)
	if err != nil {
		common.SysError(tokenServiceLogPrefix + fmt.Sprintf("Failed to get token account for user %d: %v", userID, err))
		return nil, err
	}
	return userToken, nil
}

// GetTransactionByUUID retrieves a transaction record by UUID
func (s *TokenService) GetTransactionByUUID(transactionUUID string) (*model.TokenTransaction, error) {
	return s.tokenRepo.GetTransactionByUUID(transactionUUID)
}

// CreditToken adds tokens to a user's account
func (s *TokenService) CreditToken(userID int64, amount int64, transactionUUID string, transactionType string, description string, relatedEntityType string, relatedEntityID string) (*model.UserToken, error) {
	if amount <= 0 {
		common.SysError(tokenServiceLogPrefix + fmt.Sprintf("Token credit amount must be positive, user: %d, amount: %d", userID, amount))
		return nil, fmt.Errorf("credit amount must be positive")
	}

	common.SysLog(tokenServiceLogPrefix + fmt.Sprintf("Attempting to credit %d tokens to user %d, transaction ID: %s, type: %s", amount, userID, transactionUUID, transactionType))

	userToken, err := s.tokenRepo.CreditUserToken(userID, amount, transactionUUID, transactionType, description, relatedEntityType, relatedEntityID)
	if err != nil {
		common.SysError(tokenServiceLogPrefix + fmt.Sprintf("Failed to credit tokens to user %d: %v", userID, err))
		return nil, err
	}

	common.SysLog(tokenServiceLogPrefix + fmt.Sprintf("Tokens credited successfully to user %d, new balance: %d", userID, userToken.Balance))
	return userToken, nil
}

// DebitToken deducts tokens from a user's account
func (s *TokenService) DebitToken(userID int64, amount int64, transactionUUID string, transactionType string, description string, relatedEntityType string, relatedEntityID string) (*model.UserToken, error) {
	if amount <= 0 {
		common.SysError(tokenServiceLogPrefix + fmt.Sprintf("Token debit amount must be positive, user: %d, amount: %d", userID, amount))
		return nil, fmt.Errorf("debit amount must be positive")
	}

	common.SysLog(tokenServiceLogPrefix + fmt.Sprintf("Attempting to debit %d tokens from user %d, transaction ID: %s, type: %s", amount, userID, transactionUUID, transactionType))

	userToken, err := s.tokenRepo.DebitUserToken(userID, amount, transactionUUID, transactionType, description, relatedEntityType, relatedEntityID)
	if err != nil {
		common.SysError(tokenServiceLogPrefix + fmt.Sprintf("Failed to debit tokens from user %d: %v", userID, err))
		return nil, err
	}

	common.SysLog(tokenServiceLogPrefix + fmt.Sprintf("Tokens debited successfully from user %d, new balance: %d", userID, userToken.Balance))
	return userToken, nil
}

// GetUserTransactions retrieves a user's transaction history
func (s *TokenService) GetUserTransactions(userID int64, page, limit int) ([]model.TokenTransaction, int64, error) {
	common.SysLog(tokenServiceLogPrefix + fmt.Sprintf("Retrieving transactions for user %d, page: %d, limit: %d", userID, page, limit))
	transactions, total, err := s.tokenRepo.GetUserTokenTransactions(userID, page, limit)
	if err != nil {
		common.SysError(tokenServiceLogPrefix + fmt.Sprintf("Failed to retrieve transactions for user %d: %v", userID, err))
		return nil, 0, err
	}
	common.SysLog(tokenServiceLogPrefix + fmt.Sprintf("Successfully retrieved transactions for user %d, total: %d", userID, total))
	return transactions, total, nil
}
