package repository

import (
	"gin-template/model"
	"time"

	"gorm.io/gorm"
)

type TokenReconciliationRepository struct {
	DB *gorm.DB
}

func NewTokenReconciliationRepository(db *gorm.DB) *TokenReconciliationRepository {
	return &TokenReconciliationRepository{
		DB: db,
	}
}

// SaveReconciliationRecord 保存对账记录
func (r *TokenReconciliationRepository) SaveReconciliationRecord(userID int64, currentBalance, calculatedBalance int64, description string) (*model.TokenReconciliationRecord, error) {
	record := &model.TokenReconciliationRecord{
		UserID:            userID,
		CurrentBalance:    currentBalance,
		CalculatedBalance: calculatedBalance,
		Discrepancy:       currentBalance - calculatedBalance,
		Description:       description,
	}

	err := r.DB.Create(record).Error
	if err != nil {
		return nil, err
	}

	return record, nil
}

// UpdateReconciliationRecordAsFixed 更新对账记录为已修复
func (r *TokenReconciliationRepository) UpdateReconciliationRecordAsFixed(recordID int64) error {
	now := time.Now()
	return r.DB.Model(&model.TokenReconciliationRecord{}).
		Where("id = ?", recordID).
		Updates(map[string]interface{}{
			"is_fixed": true,
			"fixed_at": now,
		}).Error
}

// GetRecentReconciliationRecords 获取最近的对账记录
func (r *TokenReconciliationRepository) GetRecentReconciliationRecords(limit int) ([]model.TokenReconciliationRecord, error) {
	var records []model.TokenReconciliationRecord

	if limit <= 0 {
		limit = 100
	}

	err := r.DB.Order("created_at DESC").Limit(limit).Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetUserReconciliationRecords 获取用户的对账记录
func (r *TokenReconciliationRepository) GetUserReconciliationRecords(userID int64, limit int) ([]model.TokenReconciliationRecord, error) {
	var records []model.TokenReconciliationRecord

	if limit <= 0 {
		limit = 50
	}

	err := r.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error

	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetUnfixedReconciliationRecords 获取未修复的对账记录
func (r *TokenReconciliationRepository) GetUnfixedReconciliationRecords(limit int) ([]model.TokenReconciliationRecord, error) {
	var records []model.TokenReconciliationRecord

	if limit <= 0 {
		limit = 100
	}

	err := r.DB.Where("is_fixed = ?", false).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error

	if err != nil {
		return nil, err
	}

	return records, nil
}

// EnsureTokenReconciliationTable 确保TokenReconciliationRecord表存在
func (r *TokenReconciliationRepository) EnsureTokenReconciliationTable() error {
	return r.DB.AutoMigrate(&model.TokenReconciliationRecord{})
}
