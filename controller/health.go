package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type HealthController struct{}

// NewHealthController 创建健康检查控制器
func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthCheck godoc
// @Summary      健康检查接口
// @Description  用于检查服务是否正常运行
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (c *HealthController) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "service is running",
	})
}
