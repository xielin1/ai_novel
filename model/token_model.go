package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserToken 代表用户Token余额表
type UserToken struct {
	UserID    uint      `gorm:"primaryKey"`
	Balance   int64     `gorm:"default:0"`
	Version   uint      `gorm:"default:1"` // 乐观锁版本号
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 指定表名
func (UserToken) TableName() string {
	return "user_tokens"
}

// TokenTransaction 代表Token交易流水表
type TokenTransaction struct {
	ID                uint      `gorm:"primaryKey;autoIncrement"`
	TransactionUUID   string    `gorm:"type:varchar(36);uniqueIndex;not null"` // 用于幂等性检查
	UserID            uint      `gorm:"index;not null"`
	Amount            int64     // 正数为增加，负数为减少
	BalanceBefore     int64
	BalanceAfter      int64
	Type              string    `gorm:"type:varchar(50);not null"` // 例如 "ai_generation_debit", "referral_credit"
	RelatedEntityType string    `gorm:"type:varchar(50)"`          // 例如 "project", "order"
	RelatedEntityID   string    `gorm:"type:varchar(100)"`         // 关联实体的ID
	Description       string    `gorm:"type:text"`
	Status            string    `gorm:"type:varchar(20);default:'completed'"` // 例如 "completed", "pending", "failed"
	CreatedAt         time.Time
}

// TableName 指定表名
func (TokenTransaction) TableName() string {
	return "token_transactions"
}

// GetUserToken 获取用户Token余额
func GetUserToken(userID uint) (*UserToken, error) {
	var userToken UserToken
	err := DB.Where("user_id = ?", userID).First(&userToken).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("token account not found for user %d", userID)
		}
		return nil, err
	}
	return &userToken, nil
}

// InitUserTokenAccount 初始化用户Token账户
func InitUserTokenAccount(userID uint, initialBalance int64) (*UserToken, error) {
	userToken := &UserToken{
		UserID:  userID,
		Balance: initialBalance,
		Version: 1,
	}
	
	err := DB.FirstOrCreate(userToken, UserToken{UserID: userID}).Error
	if err != nil {
		return nil, err
	}
	
	return userToken, nil
}

// GetTransactionByUUID 根据 UUID 查询交易记录
func GetTransactionByUUID(transactionUUID string) (*TokenTransaction, error) {
	var transaction TokenTransaction
	err := DB.Where("transaction_uuid = ?", transactionUUID).First(&transaction).Error
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
func ModifyTokenBalanceWithTransaction(tx *gorm.DB, userID uint, amount int64, transactionUUID string, 
	transactionType string, description string, relatedEntityType string, relatedEntityID string) (*UserToken, error) {
	
	// 1. 使用悲观锁获取用户Token信息
	var userToken UserToken
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
	result := tx.Model(&UserToken{}).
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
	transaction := TokenTransaction{
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
		CreatedAt:         time.Now(),
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
func CreditUserToken(userID uint, amount int64, transactionUUID string, 
	transactionType string, description string, relatedEntityType string, relatedEntityID string) (*UserToken, error) {
	
	if amount <= 0 {
		return nil, errors.New("增加的金额必须为正数")
	}
	
	// 如果没有提供交易UUID，则生成一个
	if transactionUUID == "" {
		transactionUUID = uuid.New().String()
	}
	
	var finalUserToken *UserToken
	
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 1. 幂等性检查
		existingTransaction, err := GetTransactionByUUID(transactionUUID)
		if err != nil {
			return err
		}
		
		if existingTransaction != nil && existingTransaction.Status == "completed" {
			// 交易已完成，获取当前用户Token状态
			var userToken UserToken
			if err := tx.Where("user_id = ?", userID).First(&userToken).Error; err != nil {
				return err
			}
			finalUserToken = &userToken
			return nil // 事务成功，不重复处理
		}
		
		// 2. 增加Token余额
		updatedToken, err := ModifyTokenBalanceWithTransaction(
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
		return GetUserToken(userID)
	}
	
	return finalUserToken, nil
}

// DebitUserToken 扣减用户Token余额
func DebitUserToken(userID uint, amount int64, transactionUUID string, 
	transactionType string, description string, relatedEntityType string, relatedEntityID string) (*UserToken, error) {
	
	if amount <= 0 {
		return nil, errors.New("扣减的金额必须为正数")
	}
	
	// 如果没有提供交易UUID，则生成一个
	if transactionUUID == "" {
		transactionUUID = uuid.New().String()
	}
	
	var finalUserToken *UserToken
	
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 1. 幂等性检查
		existingTransaction, err := GetTransactionByUUID(transactionUUID)
		if err != nil {
			return err
		}
		
		if existingTransaction != nil && existingTransaction.Status == "completed" {
			// 交易已完成，获取当前用户Token状态
			var userToken UserToken
			if err := tx.Where("user_id = ?", userID).First(&userToken).Error; err != nil {
				return err
			}
			finalUserToken = &userToken
			return nil // 事务成功，不重复处理
		}
		
		// 2. 扣减Token余额（注意这里传入负的金额值）
		updatedToken, err := ModifyTokenBalanceWithTransaction(
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
		return GetUserToken(userID)
	}
	
	return finalUserToken, nil
}

// GetTokenBalance 获取用户当前Token余额
func GetTokenBalance(userID uint) (int64, error) {
	userToken, err := GetUserToken(userID)
	if err != nil {
		return 0, err
	}
	return userToken.Balance, nil
}

// GetUserTokenTransactions 获取用户Token交易记录
func GetUserTokenTransactions(userID uint, page, limit int) ([]TokenTransaction, int64, error) {
	var transactions []TokenTransaction
	var total int64
	
	offset := (page - 1) * limit
	
	// 获取总数
	DB.Model(&TokenTransaction{}).Where("user_id = ?", userID).Count(&total)
	
	// 获取分页数据
	err := DB.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}
	
	return transactions, total, nil
} 