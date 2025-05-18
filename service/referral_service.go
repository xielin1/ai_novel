package service

import (
	"gin-template/model"
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
	// 使用推荐码，此函数现在会返回更新后的token余额
	newBalance, err := model.UseReferralCode(userId, code)
	if err != nil {
		return nil, err
	}

	// 获取奖励token数量 - 实际中应该从系统配置中获取
	tokensRewarded := 200

	return map[string]interface{}{
		"tokens_rewarded": tokensRewarded,
		"new_balance":     newBalance,
	}, nil
} 