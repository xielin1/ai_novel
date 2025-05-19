package service

import (
	"fmt"
	"gin-template/model"

	"github.com/google/uuid"
)

// ReferralService 推荐码服务
type ReferralService struct{}

// GetReferralCode 获取用户推荐码信息
func (s *ReferralService) GetReferralCode(userId uint) (map[string]interface{}, error) {
	// 获取用户的推荐码
	referral, err := model.GetReferralByUserId(userId)
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}

	// 如果用户没有推荐码，则创建一个
	if referral == nil || err != nil {
		referral, err = model.GenerateNewReferralCode(userId)
		if err != nil {
			return nil, err
		}
	}

	// 获取推荐统计信息
	totalReferred, totalTokensEarned, err := model.GetReferralStat(userId)
	if err != nil {
		return nil, err
	}

	// 构造分享URL
	shareURL := "https://example.com/register?ref=" + referral.Code

	return map[string]interface{}{
		"referral_code":       referral.Code,
		"total_referred":      totalReferred,
		"total_tokens_earned": totalTokensEarned,
		"share_url":           shareURL,
	}, nil
}

// GetReferrals 获取用户的推荐记录
func (s *ReferralService) GetReferrals(userId uint, page, limit int) (map[string]interface{}, error) {
	// 获取推荐记录
	referrals, totalCount, err := model.GetReferrals(userId, page, limit)
	if err != nil {
		return nil, err
	}

	// 获取统计信息
	totalReferred, totalTokensEarned, err := model.GetReferralStat(userId)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"referrals": referrals,
		"statistics": map[string]interface{}{
			"total_referred":      totalReferred,
			"total_tokens_earned": totalTokensEarned,
		},
		"pagination": map[string]interface{}{
			"total": totalCount,
			"page":  page,
			"limit": limit,
			"pages": (totalCount + limit - 1) / limit,
		},
	}, nil
}

// GenerateNewCode 为用户生成新的推荐码
func (s *ReferralService) GenerateNewCode(userId uint) (map[string]interface{}, error) {
	// 获取用户旧的推荐码
	oldReferral, err := model.GetReferralByUserId(userId)
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}

	// 保存旧的推荐码
	var previousCode string
	if oldReferral != nil {
		previousCode = oldReferral.Code
	}

	// 生成新的推荐码
	newReferral, err := model.GenerateNewReferralCode(userId)
	if err != nil {
		return nil, err
	}

	// 构造分享URL
	shareURL := "https://example.com/register?ref=" + newReferral.Code

	return map[string]interface{}{
		"previous_code": previousCode,
		"new_code":      newReferral.Code,
		"share_url":     shareURL,
	}, nil
}

// UseReferralCode 使用他人的推荐码
func (s *ReferralService) UseReferralCode(userId uint, code string) (map[string]interface{}, error) {
	// 使用推荐码，此函数会检查推荐码是否有效
	_, err := model.UseReferralCode(userId, code)
	if err != nil {
		return nil, err
	}
	
	// 奖励固定的Token数量 - 实际应该从配置中获取
	tokensRewarded := 200
	
	// 生成唯一交易ID，用于幂等性控制
	userTransactionUUID := uuid.New().String()
	
	// 为当前用户（被推荐人）增加Token - 使用TokenService确保一致性
	userToken, err := tokenService.CreditToken(
		userId,
		int64(tokensRewarded),
		userTransactionUUID,
		"referral_credit_used",
		"使用推荐码奖励",
		"referral_code",
		code,
	)
	
	if err != nil {
		return nil, fmt.Errorf("奖励Token失败: %v", err)
	}
	
	// 如果需要同时奖励推荐人，可以在这里添加相应代码
	// 为了简化，这里假设奖励推荐人的逻辑已经在model.UseReferralCode中处理
	
	return map[string]interface{}{
		"tokens_rewarded": tokensRewarded,
		"new_balance":     userToken.Balance,
	}, nil
} 