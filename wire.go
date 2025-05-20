//go:build wireinject
// +build wireinject

package main

import (
	"gin-template/repository"
	"gin-template/service"
	"gorm.io/gorm"

	"github.com/google/wire"
)

// 新增 app_struct.go
type AppStruct struct {
	TokenService    service.TokenService
	OutlineService  service.OutlineService
	ProjectService  service.ProjectService
	ReferralService service.ReferralService
}

func InitializeAllServices(db *gorm.DB) *AppStruct {
	panic(wire.Build(
		// 依赖集合
		RepositorySet,
		TokenServiceSet,
		OutlineServiceSet,
		ProjectServiceSet,
		ReferralServiceSet,

		// 创建聚合结构体（自动绑定所有字段）
		wire.Struct(new(AppStruct), "*"),
	))
}

// repository.RepositorySet 基础仓库集合
var RepositorySet = wire.NewSet(
	repository.NewTokenRepository,
	repository.NewTokenReconciliationRepository,
	repository.NewOutlineRepository,
	repository.NewProjectRepository,
	repository.NewReferralRepository,
)

// TokenServiceSet Token服务集合
var TokenServiceSet = wire.NewSet(
	service.NewTokenService,
)

// OutlineServiceSet 大纲服务集合
var OutlineServiceSet = wire.NewSet(
	service.NewOutlineService,
)

// ProjectServiceSet 项目服务集合
var ProjectServiceSet = wire.NewSet(
	service.NewProjectService)

// ReferralServiceSet 推荐码服务集合
var ReferralServiceSet = wire.NewSet(
	service.NewReferralService,
)

//// setGlobalServices 设置全局服务实例（如果需要）
//func setGlobalServices(
//	tokenService service.TokenService,
//	projectService service.ProjectService,
//) {
//	service.SetTokenService(tokenService)
//	service.SetProjectService(projectService)
//}
