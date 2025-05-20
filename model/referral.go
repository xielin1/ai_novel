package model

import (
	"time"
)

// Referral 推荐码模型
type Referral struct {
	Id        uint      `json:"id" gorm:"primaryKey"`
	UserId    uint      `json:"user_id" gorm:"not null;index;unique"`
	Code      string    `json:"code" gorm:"type:varchar(20);not null;unique;index"`
	TotalUsed int       `json:"total_used" gorm:"default:0"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReferralUse 推荐码使用记录
type ReferralUse struct {
	Id             uint      `json:"id" gorm:"primaryKey"`
	ReferrerId     uint      `json:"referrer_id" gorm:"not null;index"`    // 推荐人ID
	UserId         uint      `json:"user_id" gorm:"not null;index;unique"` // 使用者ID
	ReferralCode   string    `json:"referral_code" gorm:"type:varchar(20);not null"`
	TokensRewarded int       `json:"tokens_rewarded" gorm:"not null"` // 奖励的token数量
	UsedAt         time.Time `json:"used_at" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at"`
}
