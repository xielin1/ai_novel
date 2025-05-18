package model

import (
	"time"
)

// TokenRecordType 定义token记录类型
const (
	TokenRecordTypePlan        = 1 // 套餐赠送
	TokenRecordTypeReferral    = 2 // 推荐奖励
	TokenRecordTypeConsumption = 3 // 续写消费
	TokenRecordTypeRecharge    = 4 // 充值
)

// TokenRecord token记录模型
type TokenRecord struct {
	Id          uint      `json:"id" gorm:"primaryKey"`
	UserId      int       `json:"user_id" gorm:"not null;index"`
	Amount      int       `json:"amount" gorm:"not null"` // 正值为增加，负值为消费
	Balance     int       `json:"balance" gorm:"not null"` // 变动后余额
	RecordType  int       `json:"record_type" gorm:"not null"` // 记录类型
	RelatedId   int       `json:"related_id"` // 相关记录ID
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// AddTokenRecord 添加token记录并更新用户余额
func AddTokenRecord(userId int, amount int, recordType int, relatedId int, description string) (int, error) {
	// 更新用户token余额
	newBalance, err := UpdateUserTokenBalance(userId, amount)
	if err != nil {
		return 0, err
	}
	
	// 创建token记录
	record := &TokenRecord{
		UserId:      userId,
		Amount:      amount,
		Balance:     newBalance,
		RecordType:  recordType,
		RelatedId:   relatedId,
		Description: description,
		CreatedAt:   time.Now(),
	}
	
	// 保存记录
	err = DB.Create(record).Error
	if err != nil {
		return 0, err
	}
	
	return newBalance, nil
}

// GetUserTokenRecords 获取用户token记录
func GetUserTokenRecords(userId int, page, pageSize int) ([]*TokenRecord, int64, error) {
	var records []*TokenRecord
	var total int64
	
	// 计算偏移量
	offset := (page - 1) * pageSize
	
	// 查询总数
	err := DB.Model(&TokenRecord{}).Where("user_id = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// 查询记录
	err = DB.Where("user_id = ?", userId).
		Order("id DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&records).Error
	
	return records, total, err
} 