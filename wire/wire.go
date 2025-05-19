//go:build wireinject
// +build wireinject

package wire

import (
	"gin-template/model"
	"gin-template/repository/db"
	"gin-template/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// 数据库Provider
func ProvideDB() *gorm.DB {
	return model.DB
}

// Repository Providers
func ProvideTokenRepository(dbConn *gorm.DB) *db.TokenRepository {
	return db.NewTokenRepository(dbConn)
}

func ProvideTokenReconciliationRepository(dbConn *gorm.DB) *db.TokenReconciliationRepository {
	return db.NewTokenReconciliationRepository(dbConn)
}

// Service Providers
func ProvideTokenService(tokenRepo *db.TokenRepository) service.TokenService {
	return service.NewTokenService(tokenRepo)
}

// RepositorySet 仓库层依赖集合
var RepositorySet = wire.NewSet(
	ProvideDB,
	ProvideTokenRepository,
	ProvideTokenReconciliationRepository,
)

// ServiceSet 服务层依赖集合
var ServiceSet = wire.NewSet(
	ProvideTokenService,
)

// AppSet 应用程序依赖集合
var AppSet = wire.NewSet(
	RepositorySet,
	ServiceSet,
)

// InitializeApp 初始化应用程序依赖
func InitializeApp() (*AppDependencies, error) {
	wire.Build(
		AppSet,
		wire.Struct(new(AppDependencies), "*"),
	)
	return &AppDependencies{}, nil
}

// AppDependencies 应用程序依赖
type AppDependencies struct {
	// 数据库
	DB *gorm.DB

	// Repositories
	TokenRepo  *db.TokenRepository
	ReconRepo  *db.TokenReconciliationRepository

	// Services
	TokenService service.TokenService
} 