package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 推荐记录结构
type ReferralRecord struct {
	ID             uint   `json:"id"`
	UserID         uint   `json:"user_id"`
	Username       string `json:"username"`
	RegisteredAt   string `json:"registered_at"`
	TokensRewarded int    `json:"tokens_rewarded"`
}

// 获取个人推荐码
func GetReferralCode(c *gin.Context) {
	// 获取当前用户ID
	// 在实际代码中使用这个ID查询数据库
	_ = c.GetUint("id")

	// 从数据库中获取用户的推荐码信息
	// 这里应该有从数据库获取用户推荐码的逻辑...

	// 构造分享URL
	referralCode := "RF7XYZ9" // 实际应从数据库获取
	shareURL := "https://example.com/register?ref=" + referralCode

	// 获取推荐统计信息
	totalReferred := 5      // 实际应从数据库统计
	totalTokensEarned := 1000 // 实际应从数据库统计

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"referral_code":       referralCode,
			"total_referred":      totalReferred,
			"total_tokens_earned": totalTokensEarned,
			"share_url":           shareURL,
		},
	})
}

// 获取推荐记录
func GetReferrals(c *gin.Context) {
	// 获取当前用户ID
	// 在实际代码中使用这个ID查询数据库
	_ = c.GetUint("id")

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 从数据库查询推荐记录
	// 这里应该有从数据库查询推荐记录的逻辑...

	// 示例数据
	referrals := []ReferralRecord{
		{
			ID:             34,
			UserID:         156,
			Username:       "user***56",
			RegisteredAt:   "2023-05-10T09:25:00Z",
			TokensRewarded: 200,
		},
	}

	// 获取统计信息
	totalReferred := 5
	totalTokensEarned := 1000

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"referrals": referrals,
			"statistics": gin.H{
				"total_referred":      totalReferred,
				"total_tokens_earned": totalTokensEarned,
			},
			"pagination": gin.H{
				"total": totalReferred,
				"page":  page,
				"limit": limit,
				"pages": (totalReferred + limit - 1) / limit,
			},
		},
	})
}

// 生成新的推荐码
func GenerateReferralCode(c *gin.Context) {
	// 获取当前用户ID
	// 在实际代码中使用这个ID查询和更新数据库
	_ = c.GetUint("id")

	// 生成新的推荐码
	// 这里应该有生成唯一推荐码并保存到数据库的逻辑...

	previousCode := "RF7XYZ9" // 实际应从数据库获取旧的推荐码
	newCode := "RF8ABC3"      // 实际应生成新的推荐码
	shareURL := "https://example.com/register?ref=" + newCode

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "推荐码已重新生成",
		"data": gin.H{
			"previous_code": previousCode,
			"new_code":      newCode,
			"share_url":     shareURL,
		},
	})
}

// 使用推荐码请求
type UseReferralRequest struct {
	ReferralCode string `json:"referralCode" binding:"required"`
}

// 使用他人的推荐码
func UseReferral(c *gin.Context) {
	var req UseReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
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

	// 验证推荐码是否有效
	// 这里应该有验证推荐码的逻辑...

	// 检查用户是否已经使用过推荐码
	// 这里应该有检查逻辑...

	// 奖励token给用户
	tokensRewarded := 200   // 实际应根据系统设置确定奖励数量
	newBalance := 1050      // 实际应计算新余额

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "推荐码使用成功",
		"data": gin.H{
			"tokens_rewarded": tokensRewarded,
			"new_balance":     newBalance,
		},
	})
} 