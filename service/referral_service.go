package service

import (
	"fmt"
	"gin-template/common"
	"gin-template/define"
	"gin-template/model"
	"strconv"

	"github.com/google/uuid"
)

// ReferralService 推荐码服务
type ReferralService struct{}

// GetReferralCode 获取用户推荐码信息
func (s *ReferralService) GetReferralCode(userId uint) (define.ReferralCodeResponse, error) {
	// 获取用户的推荐码
	referral, err := model.GetReferralByUserId(userId)
	if err != nil && err.Error() != "record not found" {
		return define.ReferralCodeResponse{}, err
	}

	// 如果用户没有推荐码，则创建一个
	if referral == nil || err != nil {
		common.SysLog("[referral]code not exist,generate referral code")
		referral, err = model.GenerateNewReferralCode(userId)
		if err != nil {
			return define.ReferralCodeResponse{}, err
		}
	}

	// 获取推荐统计信息
	totalReferred, totalTokensEarned, err := model.GetReferralStat(userId)
	if err != nil {
		return define.ReferralCodeResponse{}, err
	}

	//构造分享URL
	shareURL := "https://example.com/register?ref=" + referral.Code

	return define.ReferralCodeResponse{
		ReferralCode:      referral.Code,
		TotalReferred:     totalReferred,
		TotalTokensEarned: totalTokensEarned,
		ShareURL:          shareURL,
	}, nil
}

// GetReferrals 获取用户的推荐记录
func (s *ReferralService) GetReferrals(userId uint, page, limit int) (define.ReferralsListResponse, error) {
	// 获取推荐记录
	referrals, totalCount, err := model.GetReferrals(userId, page, limit)
	if err != nil {
		return define.ReferralsListResponse{}, err
	}

	// 获取统计信息
	totalReferred, totalTokensEarned, err := model.GetReferralStat(userId)
	if err != nil {
		return define.ReferralsListResponse{}, err
	}
	
	// 将model的referrals转换为define.ReferralDetail
	referralDetails := make([]define.ReferralDetail, len(referrals))
	for i, ref := range referrals {
		// 从map中提取字段
		id, _ := ref["id"].(uint)
		if idFloat, ok := ref["id"].(float64); ok {
			id = uint(idFloat)
		}
		
		referrerId, _ := ref["referrer_id"].(uint)
		if refIdFloat, ok := ref["referrer_id"].(float64); ok {
			referrerId = uint(refIdFloat)
		}
		
		referredId, _ := ref["referred_id"].(uint)
		if refIdFloat, ok := ref["referred_id"].(float64); ok {
			referredId = uint(refIdFloat)
		}
		
		referredName, _ := ref["referred_name"].(string)
		code, _ := ref["code"].(string)
		createdAt, _ := ref["created_at"].(string)
		
		tokensEarned := 0
		if tokensVal, ok := ref["tokens_earned"].(int); ok {
			tokensEarned = tokensVal
		} else if tokensFloat, ok := ref["tokens_earned"].(float64); ok {
			tokensEarned = int(tokensFloat)
		} else if tokensStr, ok := ref["tokens_earned"].(string); ok {
			tokensEarned, _ = strconv.Atoi(tokensStr)
		}
		
		status, _ := ref["status"].(string)
		
		referralDetails[i] = define.ReferralDetail{
			ID:           id,
			ReferrerID:   referrerId,
			ReferredID:   referredId,
			ReferredName: referredName,
			Code:         code,
			CreatedAt:    createdAt,
			TokensEarned: tokensEarned,
			Status:       status,
		}
	}

	return define.ReferralsListResponse{
		Referrals: referralDetails,
		Statistics: define.ReferralStatistics{
			TotalReferred:     totalReferred,
			TotalTokensEarned: totalTokensEarned,
		},
		Pagination: define.PaginationInfo{
			Total: totalCount,
			Page:  page,
			Limit: limit,
			Pages: (totalCount + limit - 1) / limit,
		},
	}, nil
}

// GenerateNewCode 为用户生成新的推荐码
func (s *ReferralService) GenerateNewCode(userId uint) (define.NewReferralCodeResponse, error) {
	// 获取用户旧的推荐码
	oldReferral, err := model.GetReferralByUserId(userId)
	if err != nil && err.Error() != "record not found" {
		return define.NewReferralCodeResponse{}, err
	}

	// 保存旧的推荐码
	var previousCode string
	if oldReferral != nil {
		previousCode = oldReferral.Code
	}

	// 生成新的推荐码
	newReferral, err := model.GenerateNewReferralCode(userId)
	if err != nil {
		return define.NewReferralCodeResponse{}, err
	}

	// 构造分享URL
	shareURL := "https://example.com/register?ref=" + newReferral.Code

	return define.NewReferralCodeResponse{
		PreviousCode: previousCode,
		NewCode:      newReferral.Code,
		ShareURL:     shareURL,
	}, nil
}

// UseReferralCode 使用他人的推荐码
func (s *ReferralService) UseReferralCode(userId uint, code string) (define.UseReferralCodeResponse, error) {
	// 使用推荐码，此函数会检查推荐码是否有效
	_, err := model.UseReferralCode(userId, code)
	if err != nil {
		return define.UseReferralCodeResponse{}, err
	}

	// 奖励固定的Token数量 - 实际应该从配置中获取
	tokensRewarded := 200

	// 生成唯一交易ID，用于幂等性控制
	userTransactionUUID := uuid.New().String()

	// 为当前用户（被推荐人）增加Token - 使用TokenService确保一致性
	userToken, err := GetTokenService().CreditToken(
		userId,
		int64(tokensRewarded),
		userTransactionUUID,
		"referral_credit_used",
		"使用推荐码奖励",
		"referral_code",
		code,
	)

	if err != nil {
		return define.UseReferralCodeResponse{}, fmt.Errorf("奖励Token失败: %v", err)
	}

	return define.UseReferralCodeResponse{
		TokensRewarded: int64(tokensRewarded),
		NewBalance:     userToken.Balance,
	}, nil
}
