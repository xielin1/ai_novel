package model

import (
	"time"
)

// Package 套餐模型
type Package struct {
	Id            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"type:varchar(50);not null"`
	Description   string    `json:"description" gorm:"type:varchar(255)"`
	Price         float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	MonthlyTokens int       `json:"monthly_tokens" gorm:"not null"`
	Duration      string    `json:"duration" gorm:"type:varchar(20);not null"`  // monthly, yearly, permanent
	Features      string    `json:"features" gorm:"type:text"`  // 存储为JSON字符串
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToResponsePackage 转换为API响应结构
func (p *Package) ToResponsePackage() map[string]interface{} {
	features := make([]string, 0)
	// 实际应解析JSON字符串到切片

	return map[string]interface{}{
		"id":              p.Id,
		"name":            p.Name,
		"description":     p.Description,
		"price":           p.Price,
		"monthly_tokens":  p.MonthlyTokens,
		"duration":        p.Duration,
		"features":        features,
	}
}

// Subscription 订阅模型
type Subscription struct {
	Id            uint      `json:"id" gorm:"primaryKey"`
	UserId        uint      `json:"user_id" gorm:"not null;index"`
	PackageId     uint      `json:"package_id" gorm:"not null;index"`
	Status        string    `json:"status" gorm:"type:varchar(20);not null;default:'active'"` // active, expired, cancelled
	StartDate     time.Time `json:"start_date" gorm:"not null"`
	ExpiryDate    time.Time `json:"expiry_date" gorm:"not null"`
	AutoRenew     bool      `json:"auto_renew" gorm:"default:true"`
	NextRenewal   time.Time `json:"next_renewal,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GetUserCurrentPackage 获取用户当前套餐
func GetUserCurrentPackage(userId uint) (*Subscription, *Package, error) {
	// 这里实现从数据库查询用户当前有效的订阅和套餐信息
	// 返回订阅和套餐信息，如果发生错误则返回错误
	return nil, nil, nil
}

// CreateSubscription 创建新的订阅
func CreateSubscription(userId uint, packageId uint, autoRenew bool) (*Subscription, error) {
	// 这里实现创建新订阅的逻辑
	// 包括计算开始时间、结束时间等
	return nil, nil
}

// CancelRenewal 取消自动续费
func (s *Subscription) CancelRenewal() error {
	// 更新订阅的自动续费状态
	s.AutoRenew = false
	// 更新数据库
	return nil
}

// TokenDistribution 每月Token分发记录
type TokenDistribution struct {
	Id           uint      `json:"id" gorm:"primaryKey"`
	UserId       uint      `json:"user_id" gorm:"not null;index"`
	SubscriptionId uint    `json:"subscription_id" gorm:"index"`
	PackageId    uint      `json:"package_id" gorm:"index"`
	Amount       int       `json:"amount" gorm:"not null"`
	DistributedAt time.Time `json:"distributed_at" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
}

// DistributeMonthlyTokens 为用户分发月度Token
func DistributeMonthlyTokens(userId uint, subscriptionId uint, packageId uint, amount int) error {
	// 实现分发Token的逻辑
	// 创建Token分发记录并更新用户Token余额
	return nil
} 