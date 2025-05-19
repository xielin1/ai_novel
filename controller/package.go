package controller

import (
	"gin-template/common"
	"gin-template/service"
	"github.com/gin-gonic/gin"
)

// 购买/订阅套餐请求
type SubscribeRequest struct {
	PackageID     uint   `json:"package_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// 获取套餐列表
func GetPackages(c *gin.Context) {
	packages, err := service.GetAllPackages()
	if err != nil {
		common.SysError("[package]获取套餐列表失败")
		ResponseError(c, "获取套餐列表失败")
		return
	}

	ResponseOK(c, packages)
}

// 购买/订阅套餐
func SubscribePackage(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseErrorWithData(c, "参数错误", []gin.H{
			{
				"field":   "package_id/payment_method",
				"message": "必填字段",
			},
		})
		return
	}

	// 验证套餐ID是否存在
	if !service.ValidatePackageID(req.PackageID) {
		ResponseError(c, "套餐不存在")
		return
	}

	// 检查支付方式是否有效
	if !service.ValidatePaymentMethod(req.PaymentMethod) {
		ResponseError(c, "不支持的支付方式")
		return
	}

	// 获取当前用户ID
	userID := c.GetUint("id")

	// 调用服务层创建订阅
	result, err := service.CreatePackageSubscription(userID, req.PackageID, req.PaymentMethod)
	if err != nil {
		common.SysError("[package]订阅失败")
		ResponseError(c, "订阅失败")
		return
	}

	ResponseOKWithMessage(c, "订阅成功", result)
}

// 获取当前套餐信息
func GetUserPackage(c *gin.Context) {
	// 获取当前用户ID
	userID := c.GetUint("id")

	// 调用服务层获取套餐信息
	packageInfo, err := service.GetUserCurrentPackageInfo(userID)
	if err != nil {
		common.SysError("[package]获取套餐信息失败")
		ResponseError(c, "获取套餐信息失败")
		return
	}

	ResponseOK(c, packageInfo)
}

// 取消自动续费
func CancelRenewal(c *gin.Context) {
	// 获取当前用户ID
	userID := c.GetUint("id")

	// 调用服务层取消自动续费
	result, err := service.CancelPackageRenewal(userID)
	if err != nil {
		common.SysError("[package]取消自动续费失败")
		ResponseError(c, "cancel package renewal fail")
		return
	}
	ResponseOKWithMessage(c, "success cancel", result)
}
