package controller

import "gorm.io/gorm"

// Controllers 所有控制器的集合
type Controllers struct {
	Health *HealthController
	// ... 其他控制器
}

// InitializeAllController 初始化所有控制器
func InitializeAllController(db *gorm.DB) (*Controllers, error) {
	return &Controllers{
		Health: NewHealthController(),
		// ... 初始化其他控制器
	}, nil
}
