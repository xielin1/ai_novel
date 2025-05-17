package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 套餐数据结构
type Package struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Price         float64  `json:"price"`
	MonthlyTokens int      `json:"monthly_tokens"`
	Duration      string   `json:"duration"`    // monthly, yearly, permanent
	Features      []string `json:"features"`
}

// 订阅数据结构
type Subscription struct {
	PackageID    uint   `json:"package_id"`
	UserID       uint   `json:"user_id"`
	Status       string `json:"status"`        // active, expired, cancelled
	StartDate    string `json:"start_date"`
	ExpiryDate   string `json:"expiry_date"`
	AutoRenew    bool   `json:"auto_renew"`
	NextRenewal  string `json:"next_renewal_date,omitempty"`
}

// 获取套餐列表
func GetPackages(c *gin.Context) {
	// 这里应从数据库中获取套餐信息
	packages := []Package{
		{
			ID:            1,
			Name:          "基础版",
			Description:   "适合轻度使用的创作者",
			Price:         19.9,
			MonthlyTokens: 5000,
			Duration:      "monthly",
			Features:      []string{"基础AI续写", "历史版本保存"},
		},
		{
			ID:            2,
			Name:          "升级版",
			Description:   "适合中度创作需求",
			Price:         49.9,
			MonthlyTokens: 15000,
			Duration:      "monthly",
			Features:      []string{"高级AI续写", "历史版本保存", "优先客服支持"},
		},
		{
			ID:            3,
			Name:          "永久版会员",
			Description:   "适合专业创作者",
			Price:         199.9,
			MonthlyTokens: 50000,
			Duration:      "permanent",
			Features:      []string{"高级AI续写", "无限历史版本", "专属客服", "高级导出格式"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    packages,
	})
}

// 购买/订阅套餐请求
type SubscribeRequest struct {
	PackageID     uint   `json:"package_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// 购买/订阅套餐
func SubscribePackage(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误",
			"errors": []gin.H{
				{
					"field":   "package_id/payment_method",
					"message": "必填字段",
				},
			},
			"code": 400,
		})
		return
	}

	// 验证套餐ID是否存在
	// 这里应该有检查套餐是否存在的逻辑...

	// 检查支付方式是否有效
	validPaymentMethods := []string{"alipay", "wechat", "creditcard"}
	validPayment := false
	for _, method := range validPaymentMethods {
		if req.PaymentMethod == method {
			validPayment = true
			break
		}
	}

	if !validPayment {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不支持的支付方式",
			"code":    400,
		})
		return
	}

	// 处理支付和订阅逻辑
	// 这里应该有处理支付和创建订阅的逻辑...

	// 示例响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "订阅成功",
		"data": gin.H{
			"order_id":       "ORD20230605001",
			"package_name":   "升级版",
			"amount":         49.9,
			"payment_status": "completed",
			"valid_until":    "2023-07-05T23:59:59Z",
			"tokens_awarded": 15000,
		},
	})
}

// 获取当前套餐信息
func GetUserPackage(c *gin.Context) {
	// 获取当前用户ID
	// 在实际代码中使用这个ID查询数据库
	_ = c.GetUint("id")

	// 从数据库获取用户的订阅信息
	// 这里应该有从数据库获取用户订阅信息的逻辑...

	// 示例响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"package": gin.H{
				"id":              2,
				"name":            "升级版",
				"monthly_tokens":  15000,
			},
			"subscription_status": "active",
			"start_date":          "2023-06-05T14:30:00Z",
			"expiry_date":         "2023-07-05T23:59:59Z",
			"auto_renew":          true,
			"next_renewal_date":   "2023-07-05T00:00:00Z",
		},
	})
}

// 取消自动续费
func CancelRenewal(c *gin.Context) {
	// 获取当前用户ID
	// 在实际代码中使用这个ID更新数据库
	_ = c.GetUint("id")

	// 更新用户的订阅状态
	// 这里应该有更新用户订阅状态的逻辑...

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "自动续费已取消",
		"data": gin.H{
			"package_name": "升级版",
			"expiry_date":  "2023-07-05T23:59:59Z",
			"auto_renew":   false,
		},
	})
} 