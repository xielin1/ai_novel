package model

import (
	"errors"
)

type Order struct {
	ID                uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo           string  `gorm:"type:varchar(32);uniqueIndex;not null" json:"order_no"`
	UserID            uint64  `gorm:"index;not null" json:"user_id"`
	PackageID         uint64  `gorm:"index;not null" json:"package_id"`
	SubscriptionID    *uint64 `gorm:"index" json:"subscription_id,omitempty"`
	OriginalAmount    float64 `gorm:"type:decimal(10,2);not null" json:"original_amount"`
	PayAmount         float64 `gorm:"type:decimal(10,2);not null" json:"pay_amount"`
	Currency          string  `gorm:"type:char(3);default:'CNY'" json:"currency"`
	DiscountInfo      string  `gorm:"type:json" json:"discount_info,omitempty"`
	PaymentMethod     string  `gorm:"type:varchar(20);not null" json:"payment_method"`
	ThirdpartyTradeNo string  `gorm:"type:varchar(64)" json:"thirdparty_trade_no,omitempty"`
	PaymentTime       *int64  `gorm:"type:bigint" json:"payment_time,omitempty"`
	OrderStatus       string  `gorm:"type:varchar(20);default:'pending'" json:"order_status"`
	StatusReason      string  `gorm:"type:varchar(255)" json:"status_reason,omitempty"`
	AutoRenew         bool    `gorm:"default:false" json:"auto_renew"`
	SubscriptionCycle string  `gorm:"type:varchar(20);not null" json:"subscription_cycle"`
	CreatedAt         int64   `gorm:"type:bigint;not null" json:"created_at"`
	UpdatedAt         int64   `gorm:"type:bigint;not null" json:"updated_at"`
	EffectiveTime     *int64  `gorm:"type:bigint" json:"effective_time,omitempty"`
	ExpiryTime        *int64  `gorm:"type:bigint;index" json:"expiry_time,omitempty"`
	NextRenewalTime   *int64  `gorm:"type:bigint" json:"next_renewal_time,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

// Validate 订单校验
func (o *Order) Validate() error {
	// 状态校验
	validStatus := map[string]bool{
		"pending":   true,
		"paid":      true,
		"cancelled": true,
		"refunding": true,
		"refunded":  true,
		"expired":   true,
	}
	if !validStatus[o.OrderStatus] {
		return errors.New("invalid order status")
	}

	// 支付方式校验
	validPaymentMethods := map[string]bool{
		"alipay": true,
		"wechat": true,
		"paypal": true,
		"stripe": true,
	}
	if !validPaymentMethods[o.PaymentMethod] {
		return errors.New("invalid payment method")
	}

	return nil
}
