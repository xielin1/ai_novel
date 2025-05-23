package service

import (
	"fmt"
	"gin-template/common"
	"gin-template/define"
	"gin-template/repository"
	"strconv"

	"github.com/google/uuid"
)

type ReferralService struct {
	referralRepo *repository.ReferralRepository // 注入推荐码仓库
	tokenService *TokenService                  // 注入Token服务
}

// NewReferralService 创建推荐码服务实例
func NewReferralService(referralRepo *repository.ReferralRepository, tokenService *TokenService) *ReferralService {
	return &ReferralService{
		referralRepo: referralRepo,
		tokenService: tokenService,
	}
}

// GetReferralCode 获取用户推荐码信息
func (s *ReferralService) GetReferralCode(userId uint) (define.ReferralCodeResponse, error) {
	// 获取用户的推荐码（通过仓库）
	referral, err := s.referralRepo.GetReferralByUserId(userId)
	if err != nil {
		return define.ReferralCodeResponse{}, err
	}

	// 如果推荐码不存在，生成新的
	if referral == nil {
		common.SysLog("[referral] code not exist, generating new referral code")
		newReferral, err := s.referralRepo.GenerateNewReferralCode(userId)
		if err != nil {
			return define.ReferralCodeResponse{}, err
		}
		referral = newReferral
	}

	// 获取推荐统计信息（通过仓库）
	totalReferred, totalTokens, err := s.referralRepo.GetReferralStat(userId)
	if err != nil {
		return define.ReferralCodeResponse{}, err
	}

	// 构造分享URL
	shareURL := fmt.Sprintf("https://example.com/register?ref=%s", referral.Code)

	return define.ReferralCodeResponse{
		ReferralCode:      referral.Code,
		TotalReferred:     totalReferred,
		TotalTokensEarned: totalTokens,
		ShareURL:          shareURL,
	}, nil
}

// GetReferrals 获取用户的推荐记录
func (s *ReferralService) GetReferrals(userId uint, page, limit int) (define.ReferralsListResponse, error) {
	// 获取推荐记录（通过仓库）
	referralUses, totalCount, err := s.referralRepo.GetReferrals(userId, page, limit)
	if err != nil {
		return define.ReferralsListResponse{}, err
	}

	// 转换结果为定义的结构体
	referralDetails := make([]define.ReferralDetail, 0, len(referralUses))
	for _, use := range referralUses {
		detail := define.ReferralDetail{
			ID:           uint(use["id"].(float64)), // 处理类型转换（假设仓库返回float64）
			ReferrerID:   uint(use["referrer_id"].(float64)),
			ReferredID:   uint(use["user_id"].(float64)), // 假设user_id是被推荐人ID
			ReferredName: use["username"].(string),
			Code:         use["referral_code"].(string), // 假设仓库返回推荐码
			CreatedAt:    use["used_at"].(string),       // 假设使用时间字段名
			TokensEarned: int(use["tokens_rewarded"].(float64)),
			Status:       "active", // 假设默认状态
		}
		referralDetails = append(referralDetails, detail)
	}

	// 获取统计信息（通过仓库）
	totalReferred, totalTokens, err := s.referralRepo.GetReferralStat(userId)
	if err != nil {
		return define.ReferralsListResponse{}, err
	}

	return define.ReferralsListResponse{
		Referrals: referralDetails,
		Statistics: define.ReferralStatistics{
			TotalReferred:     totalReferred,
			TotalTokensEarned: totalTokens,
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
	// 生成新推荐码（通过仓库）
	newReferral, err := s.referralRepo.GenerateNewReferralCode(userId)
	if err != nil {
		return define.NewReferralCodeResponse{}, err
	}

	// 获取旧推荐码（如果存在）
	oldReferral, _ := s.referralRepo.GetReferralByUserId(userId)
	var previousCode string
	if oldReferral != nil {
		previousCode = oldReferral.Code
	}

	// 构造分享URL
	shareURL := fmt.Sprintf("https://example.com/register?ref=%s", newReferral.Code)

	return define.NewReferralCodeResponse{
		PreviousCode: previousCode,
		NewCode:      newReferral.Code,
		ShareURL:     shareURL,
	}, nil
}

// UseReferralCode 使用他人的推荐码
func (s *ReferralService) UseReferralCode(userId uint, code string) (define.UseReferralCodeResponse, error) {
	// 使用推荐码（通过仓库，返回奖励的Token数量）
	tokensRewarded, err := s.referralRepo.UseReferralCode(userId, code)
	if err != nil {
		return define.UseReferralCodeResponse{}, err
	}

	// 生成交易ID
	transactionUUID := uuid.New().String()

	// 为被推荐人增加Token（通过TokenService）
	userToken, err := s.tokenService.CreditToken(
		userId,
		int64(tokensRewarded),
		transactionUUID,
		"referral_credit",
		"使用推荐码获得奖励",
		"referral",
		strconv.Itoa(int(userId)),
	)
	if err != nil {
		return define.UseReferralCodeResponse{}, fmt.Errorf("发放奖励失败: %v", err)
	}

	return define.UseReferralCodeResponse{
		TokensRewarded: int64(tokensRewarded),
		NewBalance:     userToken.Balance,
	}, nil
}
