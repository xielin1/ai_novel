package controller

import (
	"gin-template/common"
	"gin-template/define"
	"gin-template/service"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PackageController handles HTTP requests for package operations.
type PackageController struct {
	packageService *service.PackageService
}

// NewPackageController creates a new instance of PackageController.
func NewPackageController(ps *service.PackageService) *PackageController {
	return &PackageController{packageService: ps}
}

// GetPackages handles the request to get all available packages.
// @Summary Get all packages
// @Description Retrieves a list of all available packages, including the free tier.
// @Tags Package
// @Accept json
// @Produce json
// @Success 200 {object} Response{data=define.PackageResponse}
// @Failure 500 {object} Response
// @Router /packages [get]
func (pc *PackageController) GetPackages(c *gin.Context) {
	packages, err := pc.packageService.GetAllPackages()
	if err != nil {
		common.SysError(fmt.Sprintf("[PackageController.GetPackages] failed to get all packages: %v", err))
		ResponseError(c, "获取套餐列表失败")
		return
	}
	ResponseOK(c, packages)
}

// SubscribePackage handles the request to subscribe to a package.
// @Summary Subscribe to a package
// @Description Creates a new subscription for the logged-in user to the specified package.
// @Tags Package
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Param body body define.CreateSubscriptionRequest true "Subscription Request"
// @Success 200 {object} Response{data=define.SubscriptionResponse}
// @Failure 400 {object} Response "Invalid request format or parameters"
// @Failure 500 {object} Response "Internal server error"
// @Router /packages/subscribe [post]
func (pc *PackageController) SubscribePackage(c *gin.Context) {
	var req define.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseErrorWithData(c, "参数错误", err.Error()) // Provide more specific error
		return
	}

	userID := c.GetUint("id") // Assuming user ID is set in context by auth middleware
	if userID == 0 {
		ResponseError(c, "用户未登录或无法获取用户信息")
		return
	}

	// Validation is now handled in the service layer, but basic checks can remain or be enhanced here.
	// For example, checking if PackageID is non-zero if that's a requirement not covered by `binding:"required"` for uint.

	result, err := pc.packageService.CreateSubscription(userID, req)
	if err != nil {
		common.SysError(fmt.Sprintf("[PackageController.SubscribePackage] failed for user %d, package %d: %v", userID, req.PackageID, err))
		// Distinguish between client errors (e.g., invalid package) and server errors
		// For now, a generic error message.
		ResponseError(c, fmt.Sprintf("订阅失败: %s", err.Error()))
		return
	}

	ResponseOKWithMessage(c, "订阅成功", result)
}

// GetUserPackage handles the request to get the current user's package information.
// @Summary Get user's current package
// @Description Retrieves the current package and subscription details for the logged-in user.
// @Tags Package
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} Response{data=define.CurrentPackageResponse}
// @Failure 500 {object} Response
// @Router /packages/current [get]
func (pc *PackageController) GetUserPackage(c *gin.Context) {
	userID := c.GetUint("id")
	if userID == 0 {
		ResponseError(c, "用户未登录或无法获取用户信息")
		return
	}

	packageInfo, err := pc.packageService.GetUserCurrentPackageInfo(userID)
	if err != nil {
		common.SysError(fmt.Sprintf("[PackageController.GetUserPackage] failed for user %d: %v", userID, err))
		ResponseError(c, "获取套餐信息失败")
		return
	}

	ResponseOK(c, packageInfo)
}

// CancelRenewal handles the request to cancel auto-renewal for the user's current package.
// @Summary Cancel subscription auto-renewal
// @Description Cancels the automatic renewal for the logged-in user's active subscription.
// @Tags Package
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} Response{data=define.CancelRenewalResponse}
// @Failure 400 {object} Response "e.g., No active subscription or renewal already cancelled"
// @Failure 500 {object} Response
// @Router /packages/cancel-renewal [post]
func (pc *PackageController) CancelRenewal(c *gin.Context) {
	userID := c.GetUint("id")
	if userID == 0 {
		ResponseError(c, "用户未登录或无法获取用户信息")
		return
	}

	result, err := pc.packageService.CancelSubscriptionRenewal(userID)
	if err != nil {
		common.SysError(fmt.Sprintf("[PackageController.CancelRenewal] failed for user %d: %v", userID, err))
		// Check if the error indicates a condition that should be a 400 (e.g., already cancelled, no active sub)
		if err.Error() == "auto-renewal is already disabled" || err.Error() == "no active subscription found to cancel renewal for" {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: err.Error(),
				Data:    result, // result might contain current state even on "error"
			})
		} else {
			ResponseError(c, "取消自动续费失败")
		}
		return
	}
	ResponseOKWithMessage(c, "成功取消自动续费", result)
}
