package wire

import (
	"gin-template/service"
)

// 初始化Token服务
func InitializeTokenService() {
	deps, err := InitializeApp()
	if err != nil {
		panic(err)
	}

	// 全局设置TokenService实例
	service.SetTokenService(deps.TokenService)
}

// 初始化Token对账服务
func InitializeTokenReconciliationService() {
	deps, err := InitializeApp()
	if err != nil {
		panic(err)
	}

	// 初始化Token对账服务
	service.InitReconciliationService(deps.TokenRepo, deps.ReconRepo)
}

// 初始化所有服务
func InitializeAllServices() {
	InitializeTokenService()
	InitializeTokenReconciliationService()
} 