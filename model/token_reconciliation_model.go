package model

import (
	"time"
)

// TokenReconciliationRecord Token对账记录
type TokenReconciliationRecord struct {
	ID               uint      `gorm:"primaryKey;autoIncrement"`
	UserID           uint      `gorm:"index;not null"`
	CurrentBalance   int64     `gorm:"not null"`
	CalculatedBalance int64     `gorm:"not null"`
	Discrepancy      int64     `gorm:"not null"` // 当前余额 - 计算余额
	IsFixed          bool      `gorm:"default:false"`
	FixedAt          *time.Time
	Description      string    `gorm:"type:text"`
	CreatedAt        time.Time
}

// TableName 指定表名
func (TokenReconciliationRecord) TableName() string {
	return "token_reconciliation_records"
}

// SaveReconciliationRecord 保存对账记录
func SaveReconciliationRecord(userID uint, currentBalance, calculatedBalance int64, description string) (*TokenReconciliationRecord, error) {
	record := &TokenReconciliationRecord{
		UserID:            userID,
		CurrentBalance:    currentBalance,
		CalculatedBalance: calculatedBalance,
		Discrepancy:       currentBalance - calculatedBalance,
		Description:       description,
		CreatedAt:         time.Now(),
	}
	
	err := DB.Create(record).Error
	if err != nil {
		return nil, err
	}
	
	return record, nil
}

// UpdateReconciliationRecordAsFixed 更新对账记录为已修复
func UpdateReconciliationRecordAsFixed(recordID uint) error {
	now := time.Now()
	return DB.Model(&TokenReconciliationRecord{}).
		Where("id = ?", recordID).
		Updates(map[string]interface{}{
			"is_fixed": true,
			"fixed_at": now,
		}).Error
}

// GetRecentReconciliationRecords 获取最近的对账记录
func GetRecentReconciliationRecords(limit int) ([]TokenReconciliationRecord, error) {
	var records []TokenReconciliationRecord
	
	if limit <= 0 {
		limit = 100
	}
	
	err := DB.Order("created_at DESC").Limit(limit).Find(&records).Error
	if err != nil {
		return nil, err
	}
	
	return records, nil
}

// GetUserReconciliationRecords 获取用户的对账记录
func GetUserReconciliationRecords(userID uint, limit int) ([]TokenReconciliationRecord, error) {
	var records []TokenReconciliationRecord
	
	if limit <= 0 {
		limit = 50
	}
	
	err := DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error
	
	if err != nil {
		return nil, err
	}
	
	return records, nil
}

// GetUnfixedReconciliationRecords 获取未修复的对账记录
func GetUnfixedReconciliationRecords(limit int) ([]TokenReconciliationRecord, error) {
	var records []TokenReconciliationRecord
	
	if limit <= 0 {
		limit = 100
	}
	
	err := DB.Where("is_fixed = ?", false).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error
	
	if err != nil {
		return nil, err
	}
	
	return records, nil
} 