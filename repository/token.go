package repository

import (
	"errors"
	"fmt"
	"gin-template/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TokenRepository 提供Token相关的数据库操作
type TokenRepository struct {
	DB *gorm.DB
}

// NewTokenRepository 创建一个新的TokenRepository实例
func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{
		DB: db,
	}
}

// GetUserToken 获取用户Token余额
func (r *TokenRepository) GetUserToken(userID uint) (*model.UserToken, error) {
	var userToken model.UserToken
	err := r.DB.Where("user_id = ?", userID).First(&userToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("token account not found for user %d", userID)
		}
		return nil, err
	}
	return &userToken, nil
}

// InitUserTokenAccount 初始化用户Token账户
func (r *TokenRepository) InitUserTokenAccount(userID uint, initialBalance int64) (*model.UserToken, error) {
	userToken := &model.UserToken{
		UserID:  userID,
		Balance: initialBalance,
		Version: 1,
	}

	err := r.DB.FirstOrCreate(userToken, model.UserToken{UserID: userID}).Error
	if err != nil {
		return nil, err
	}

	return userToken, nil
}

// GetTransactionByUUID 根据UUID获取交易记录
func (r *TokenRepository) GetTransactionByUUID(transactionUUID string) (*model.TokenTransaction, error) {
	var transaction model.TokenTransaction
	err := r.DB.Where("transaction_uuid = ?", transactionUUID).First(&transaction).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 未找到不视为错误，表示新交易
		}
		return nil, err
	}
	return &transaction, nil
}

// ModifyTokenBalanceWithTransaction 在事务中修改用户Token余额
// amount 为正表示增加，为负表示减少
func (r *TokenRepository) ModifyTokenBalanceWithTransaction(tx *gorm.DB, userID uint, amount int64, transactionUUID string,
	transactionType string, description string, relatedEntityType string, relatedEntityID string) (*model.UserToken, error) {

	// 1. 使用悲观锁获取用户Token信息
	var userToken model.UserToken
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&userToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户 %d 的Token账户不存在", userID)
		}
		return nil, err
	}

	// 2. 对于扣减操作，检查余额是否充足
	if amount < 0 && userToken.Balance < -amount {
		return nil, fmt.Errorf("用户 %d 余额不足: 当前 %d, 尝试扣减 %d", userID, userToken.Balance, -amount)
	}

	balanceBefore := userToken.Balance
	balanceAfter := userToken.Balance + amount

	// 3. 使用乐观锁更新用户余额
	result := tx.Model(&model.UserToken{}).
		Where("user_id = ? AND version = ?", userID, userToken.Version).
		Updates(map[string]interface{}{
			"balance":    balanceAfter,
			"version":    userToken.Version + 1,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("更新Token余额失败，可能发生并发冲突")
	}

	// 4. 创建交易记录
	transaction := model.TokenTransaction{
		TransactionUUID:   transactionUUID,
		UserID:            userID,
		Amount:            amount,
		BalanceBefore:     balanceBefore,
		BalanceAfter:      balanceAfter,
		Type:              transactionType,
		RelatedEntityType: relatedEntityType,
		RelatedEntityID:   relatedEntityID,
		Description:       description,
		Status:            "completed",
	}

	if err := tx.Create(&transaction).Error; err != nil {
		return nil, err
	}

	// 5. 更新返回的用户Token对象
	userToken.Balance = balanceAfter
	userToken.Version = userToken.Version + 1
	userToken.UpdatedAt = time.Now()

	return &userToken, nil
}

// CreditUserToken 增加用户Token余额
func (r *TokenRepository) CreditUserToken(userID uint, amount int64, transactionUUID string,
	transactionType string, description string, relatedEntityType string, relatedEntityID string) (*model.UserToken, error) {

	if amount <= 0 {
		return nil, errors.New("增加的金额必须为正数")
	}

	// 如果没有提供交易UUID，则生成一个
	if transactionUUID == "" {
		transactionUUID = uuid.New().String()
	}

	var finalUserToken *model.UserToken

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 幂等性检查
		existingTransaction, err := r.GetTransactionByUUID(transactionUUID)
		if err != nil {
			return err
		}

		if existingTransaction != nil && existingTransaction.Status == "completed" {
			// 交易已完成，获取当前用户Token状态
			var userToken model.UserToken
			if err := tx.Where("user_id = ?", userID).First(&userToken).Error; err != nil {
				return err
			}
			finalUserToken = &userToken
			return nil // 事务成功，不重复处理
		}

		// 2. 增加Token余额
		updatedToken, err := r.ModifyTokenBalanceWithTransaction(
			tx, userID, amount, transactionUUID, transactionType, description, relatedEntityType, relatedEntityID)
		if err != nil {
			return err
		}

		finalUserToken = updatedToken
		return nil
	})

	if err != nil {
		return nil, err
	}

	if finalUserToken == nil {
		// 安全检查，正常情况下不应该发生
		return r.GetUserToken(userID)
	}

	return finalUserToken, nil
}

// DebitUserToken 扣减用户Token余额
func (r *TokenRepository) DebitUserToken(userID uint, amount int64, transactionUUID string,
	transactionType string, description string, relatedEntityType string, relatedEntityID string) (*model.UserToken, error) {

	if amount <= 0 {
		return nil, errors.New("扣减的金额必须为正数")
	}

	// 如果没有提供交易UUID，则生成一个
	if transactionUUID == "" {
		transactionUUID = uuid.New().String()
	}

	var finalUserToken *model.UserToken

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 幂等性检查
		existingTransaction, err := r.GetTransactionByUUID(transactionUUID)
		if err != nil {
			return err
		}

		if existingTransaction != nil && existingTransaction.Status == "completed" {
			// 交易已完成，获取当前用户Token状态
			var userToken model.UserToken
			if err := tx.Where("user_id = ?", userID).First(&userToken).Error; err != nil {
				return err
			}
			finalUserToken = &userToken
			return nil // 事务成功，不重复处理
		}

		// 2. 扣减Token余额（注意这里传入负的金额值）
		updatedToken, err := r.ModifyTokenBalanceWithTransaction(
			tx, userID, -amount, transactionUUID, transactionType, description, relatedEntityType, relatedEntityID)
		if err != nil {
			return err
		}

		finalUserToken = updatedToken
		return nil
	})

	if err != nil {
		return nil, err
	}

	if finalUserToken == nil {
		// 安全检查，正常情况下不应该发生
		return r.GetUserToken(userID)
	}

	return finalUserToken, nil
}

// GetTokenBalance 获取用户当前Token余额
func (r *TokenRepository) GetTokenBalance(userID uint) (int64, error) {
	userToken, err := r.GetUserToken(userID)
	if err != nil {
		return 0, err
	}
	return userToken.Balance, nil
}

// GetUserTokenTransactions 获取用户Token交易记录
func (r *TokenRepository) GetUserTokenTransactions(userID uint, page, limit int) ([]model.TokenTransaction, int64, error) {
	var transactions []model.TokenTransaction
	var total int64

	offset := (page - 1) * limit

	// 获取总数
	r.DB.Model(&model.TokenTransaction{}).Where("user_id = ?", userID).Count(&total)

	// 获取分页数据
	err := r.DB.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}
