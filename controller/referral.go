package controller

import (
	"net/http"
	"strconv"

	"gin-template/service"

	"github.com/gin-gonic/gin"
)

// ReferralController 推荐码控制器
type ReferralController struct {
	referralService *service.ReferralService
}

// NewReferralController 创建推荐码控制器实例
func NewReferralController() *ReferralController {
	return &ReferralController{
		referralService: &service.ReferralService{},
	}
}

// 推荐记录结构
type ReferralRecord struct {
	ID             uint   `json:"id"`
	UserID         uint   `json:"user_id"`
	Username       string `json:"username"`
	RegisteredAt   string `json:"registered_at"`
	TokensRewarded int    `json:"tokens_rewarded"`
}

// 获取个人推荐码
func (c *ReferralController) GetReferralCode(ctx *gin.Context) {
	// 获取当前用户ID
	userId := ctx.GetUint("id")

	// 使用服务层获取推荐码信息
	result, err := c.referralService.GetReferralCode(userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取推荐码失败: " + err.Error(),
			"code":    500,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// 获取推荐记录
func (c *ReferralController) GetReferrals(ctx *gin.Context) {
	// 获取当前用户ID
	userId := ctx.GetUint("id")

	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// 使用服务层获取推荐记录
	result, err := c.referralService.GetReferrals(userId, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取推荐记录失败: " + err.Error(),
			"code":    500,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// 生成新的推荐码
func (c *ReferralController) GenerateReferralCode(ctx *gin.Context) {
	// 获取当前用户ID
	userId := ctx.GetUint("id")

	// 使用服务层生成新的推荐码
	result, err := c.referralService.GenerateNewCode(userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "生成推荐码失败: " + err.Error(),
			"code":    500,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "推荐码已重新生成",
		"data":    result,
	})
}

// 使用推荐码请求
type UseReferralRequest struct {
	ReferralCode string `json:"referralCode" binding:"required"`
}

// 使用他人的推荐码
func (c *ReferralController) UseReferral(ctx *gin.Context) {
	var req UseReferralRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误",
			"errors": []gin.H{
				{
					"field":   "referralCode",
					"message": "推荐码不能为空",
				},
			},
			"code": 400,
		})
		return
	}

	// 获取当前用户ID
	userId := ctx.GetUint("id")

	// 使用服务层处理推荐码
	result, err := c.referralService.UseReferralCode(userId, req.ReferralCode)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
			"code":    400,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "推荐码使用成功",
		"data":    result,
	})
} 