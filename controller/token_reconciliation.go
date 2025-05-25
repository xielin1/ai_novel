package controller

import (
	"fmt"
	"gin-template/common"
	"gin-template/model"
	"gin-template/repository"
	"gin-template/service"
	"github.com/gin-gonic/gin"
)

type ReconciliationController struct{}

func NewReconciliationController() *ReconciliationController {
	return &ReconciliationController{}
}

// StartService 启动对账服务接口
func (c *ReconciliationController) StartService(ctx *gin.Context) {
	if service.GetTokenReconciliation() == nil {
		service.InitReconciliationService(
			repository.NewTokenRepository(model.DB),
			repository.NewTokenReconciliationRepository(model.DB),
		)
	}

	ResponseOKWithMessage(ctx, "Reconciliation service started", nil)
}

// StopService 停止对账服务接口
func (c *ReconciliationController) StopService(ctx *gin.Context) {

	service.StopReconciliationService()
	ResponseOKWithMessage(ctx, "Reconciliation service stopped", nil)
}

// FullReconciliation 触发全量对账接口
func (c *ReconciliationController) FullReconciliation(ctx *gin.Context) {

	err := service.ReconcileAllTokens()
	if err != nil {
		common.SysError("[reconciliation] Full reconciliation failed")
		ResponseErrorWithData(ctx, "Full reconciliation failed", err.Error())
		return
	}

	ResponseOKWithMessage(ctx, "Full reconciliation triggered", nil)
}

// UserReconciliation 对特定用户进行对账接口
func (c *ReconciliationController) UserReconciliation(ctx *gin.Context) {
	userId := ctx.GetInt64("id")
	err := service.ReconcileUserToken(userId)
	if err != nil {
		common.SysError(fmt.Sprintf("[reconciliation] User %d reconciliation failed", userId))
		ResponseErrorWithData(ctx, "User reconciliation failed", err.Error())
		return
	}
	ResponseOKWithMessage(ctx, "User reconciliation completed", nil)
}
