package model

// TokenReconciliationRecord Token对账记录
type TokenReconciliationRecord struct {
	ID                int64 `gorm:"primaryKey;autoIncrement"`
	UserID            int64 `gorm:"index;not null"`
	CurrentBalance    int64 `gorm:"not null"`
	CalculatedBalance int64 `gorm:"not null"`
	Discrepancy       int64 `gorm:"not null"` // 当前余额 - 计算余额
	IsFixed           bool  `gorm:"default:false"`
	FixedAt           int64
	Description       string `gorm:"type:text"`
	CreatedAt         int64
}

// TableName 指定表名
func (TokenReconciliationRecord) TableName() string {
	return "token_reconciliation_records"
}
