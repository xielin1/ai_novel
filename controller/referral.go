package controller

import (
	"gin-template/define"
	"strconv"

	"gin-template/common"
	"gin-template/service"

	"github.com/gin-gonic/gin"
)

type ReferralController struct {
	referralService *service.ReferralService
}

func NewReferralController(referralService *service.ReferralService) *ReferralController {
	return &ReferralController{
		referralService: referralService,
	}
}

func (c *ReferralController) GetReferralCode(ctx *gin.Context) {
	// 获取当前用户ID
	userId := ctx.GetInt64("id")
	result, err := c.referralService.GetReferralCode(userId)
	if err != nil {
		common.SysError("[referral]获取推荐码失败")
		ResponseError(ctx, "获取推荐码失败")
		return
	}
	ResponseOK(ctx, result)
}

func (c *ReferralController) GetReferrals(ctx *gin.Context) {
	// 获取当前用户ID
	userId := ctx.GetInt64("id")

	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	// 使用服务层获取推荐记录
	result, err := c.referralService.GetReferrals(userId, page, limit)
	if err != nil {
		common.SysError("[referral]获取推荐记录失败")
		ResponseError(ctx, "获取推荐记录失败")
		return
	}

	ResponseOK(ctx, result)
}
func (c *ReferralController) UseReferral(ctx *gin.Context) {
	var req define.UseReferralRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseError(ctx, "参数错误")
	}

	// 获取当前用户ID
	userId := ctx.GetInt64("id")
	username := ctx.GetString("username")

	result, err := c.referralService.UseReferralCode(userId, username, req.ReferralCode)
	if err != nil {
		common.SysError("[referral]使用推荐码失败")
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOKWithMessage(ctx, "推荐码使用成功", result)
}
