package model

import (
	"time"
)

// FreePackage 免费版套餐常量
var FreePackage = Package{
	Id:            0,
	Name:          "免费版",
	Description:   "基础功能免费体验",
	Price:         0,
	MonthlyTokens: 500,
	Duration:      "monthly",
	Features:      `["基础AI续写功能", "每月500个免费Token", "社区支持"]`,
}

// Package 套餐模型
type Package struct {
	Id            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"type:varchar(50);not null"`
	Description   string    `json:"description" gorm:"type:varchar(255)"`
	Price         float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	MonthlyTokens int       `json:"monthly_tokens" gorm:"not null"`
	Duration      string    `json:"duration" gorm:"type:varchar(20);not null"` // monthly, yearly, permanent
	Features      string    `json:"features" gorm:"type:text"`                 // 存储为JSON字符串
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToResponsePackage 转换为API响应结构
// 注意：此处的 Features 解析逻辑可能需要根据实际存储的 JSON 格式进行调整
func (p *Package) ToResponsePackage() map[string]interface{} {
	// 实际项目中，应该用 json.Unmarshal 解析 p.Features
	// 这里为了简化，暂时返回原始字符串或一个空slice
	// var features []string
	// if p.Features != "" {
	// 	 if err := json.Unmarshal([]byte(p.Features), &features); err != nil {
	// 		 // log error or handle
	// 	 }
	// }

	return map[string]interface{}{
		"id":             p.Id,
		"name":           p.Name,
		"description":    p.Description,
		"price":          p.Price,
		"monthly_tokens": p.MonthlyTokens,
		"duration":       p.Duration,
		"features":       p.Features, // 或者返回解析后的 features slice
	}
}

// Subscription 订阅模型
type Subscription struct {
	Id          uint      `json:"id" gorm:"primaryKey"`
	UserId      uint      `json:"user_id" gorm:"not null;index"`
	PackageId   uint      `json:"package_id" gorm:"not null;index"`
	Status      string    `json:"status" gorm:"type:varchar(20);not null;default:'active'"` // active, expired, cancelled
	StartDate   time.Time `json:"start_date" gorm:"not null"`
	ExpiryDate  time.Time `json:"expiry_date" gorm:"not null"`
	AutoRenew   bool      `json:"auto_renew" gorm:"default:true"`
	NextRenewal time.Time `json:"next_renewal,omitempty" gorm:"null"` // 允许 NextRenewal 为空
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TokenDistribution 每月Token分发记录
type TokenDistribution struct {
	Id             uint      `json:"id" gorm:"primaryKey"`
	UserId         uint      `json:"user_id" gorm:"not null;index"`
	SubscriptionId uint      `json:"subscription_id" gorm:"index"`
	PackageId      uint      `json:"package_id" gorm:"index"`
	Amount         int       `json:"amount" gorm:"not null"`
	DistributedAt  time.Time `json:"distributed_at" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at"`
}
