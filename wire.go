//go:build wireinject
// +build wireinject

package main

import (
	"gin-template/controller"
	"gin-template/repository"
	"gin-template/router"
	"gin-template/service"
	"github.com/google/wire"
	"gorm.io/gorm"
)

func InitializeAllController(db *gorm.DB) (*router.APIControllers, error) {
	panic(wire.Build(
		ControllerSet,
		ServiceSet,
		RepositorySet,
		// 创建聚合结构体（自动绑定所有字段）
		wire.Struct(new(router.APIControllers), "*"),
	))
	return &router.APIControllers{}, nil
}

// ServiceSet 大纲服务集合
var ServiceSet = wire.NewSet(
	service.NewOutlineService,
	service.NewTokenService,
	service.NewProjectService,
	service.NewReferralService,
	service.NewPackageService,
)

// repository.RepositorySet 基础仓库集合
var RepositorySet = wire.NewSet(
	repository.NewTokenRepository,
	repository.NewTokenReconciliationRepository,
	repository.NewOutlineRepository,
	repository.NewProjectRepository,
	repository.NewReferralRepository,
	repository.NewPackageRepository,
)

// 控制器依赖注入集合
var ControllerSet = wire.NewSet(
	controller.NewReferralController,
	controller.NewProjectController,
	controller.NewOutlineController,
	controller.NewPackageController,
	controller.NewReconciliationController,
	controller.NewHealthController,
)
