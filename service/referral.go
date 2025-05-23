package service

import (
	"fmt"
	"gin-template/common"
	"gin-template/define"
	"gin-template/repository"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
	"strconv"
)

type ReferralService struct {
	referralRepo *repository.ReferralRepository
	tokenService *TokenService
}

func NewReferralService(referralRepo *repository.ReferralRepository, tokenService *TokenService) *ReferralService {
	return &ReferralService{
		referralRepo: referralRepo,
		tokenService: tokenService,
	}
}

// GetReferralCode 获取用户推荐码信息
func (s *ReferralService) GetReferralCode(userId int64) (define.ReferralCodeResponse, error) {
	resp := define.ReferralCodeResponse{}
	// 获取用户推荐码
	referral, err := s.referralRepo.GetReferralCodeByUserId(userId)
	if err != nil {
		common.SysError(fmt.Sprintf("get referral code err,%v", err))
		return resp, err
	}

	// 如果推荐码不存在，生成新的
	if referral == "" {
		common.SysLog("[referral] code not exist, generating new referral code")
		newReferralCodeResp, err := s.GenerateNewCode(userId)
		if err != nil {
			common.SysError(fmt.Sprintf("generating new referral code,%v", err))
			return resp, err
		}
		resp.ReferralCode = newReferralCodeResp.NewCode
		resp.ShareURL = newReferralCodeResp.ShareURL

	} else {
		resp.ReferralCode = referral
		//todo 生成推荐链接

		//已经存在，获取推荐统计信息
		totalReferred, totalTokens, err := s.referralRepo.GetReferralStat(userId)
		if err != nil {
			return define.ReferralCodeResponse{}, err
		}
		resp.TotalReferred = totalReferred
		resp.TotalTokensEarned = totalTokens
	}
	return resp, nil
}

// GetReferrals 获取用户的推荐记录
func (s *ReferralService) GetReferrals(userId int64, page, limit int) (define.ReferralsListResponse, error) {
	// 获取推荐记录
	referralUses, totalCount, err := s.referralRepo.GetReferrals(userId, page, limit)
	if err != nil {
		return define.ReferralsListResponse{}, err
	}

	// 转换结果为定义的结构体
	referralDetails := make([]define.ReferralDetail, 0, len(referralUses))
	for _, use := range referralUses {
		detail := define.ReferralDetail{
			ID:           use["id"].(int64),
			ReferredName: use["username"].(string),
			CreatedAt:    use["used_at"].(string),
			TokensEarned: int(use["tokens_rewarded"].(float64)),
		}
		referralDetails = append(referralDetails, detail)
	}

	// 获取统计信息
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
func (s *ReferralService) GenerateNewCode(userId int64) (define.NewReferralCodeResponse, error) {
	resp := define.NewReferralCodeResponse{}
	// 获取旧推荐码（如果存在）
	oldReferral, _ := s.referralRepo.GetReferralCodeByUserId(userId)
	if oldReferral != "" {
		resp.NewCode = oldReferral
	} else {
		// 生成新推荐码
		code := s.generateRandomCode(8)
		resp.NewCode = code
		s.referralRepo.GenerateNewReferralCode(userId, code)
	}
	// 构造分享URL,TODO优化分享链接
	resp.ShareURL = fmt.Sprintf("https://example.com/register?ref=%s", resp.NewCode)

	return resp, nil
}

// UseReferralCode 使用他人的推荐码
func (s *ReferralService) UseReferralCode(userId int64, username, code string) (define.UseReferralCodeResponse, error) {
	// 使用推荐码
	tokensRewarded, err := s.referralRepo.UseReferralCode(userId, username, code, 200)
	if err != nil {
		return define.UseReferralCodeResponse{}, err
	}

	// 生成交易ID
	transactionUUID := uuid.New().String()

	// Todo 优化成异步重试，为被推荐人增加Token
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

// generateRandomCode 生成随机推荐码
func (s *ReferralService) generateRandomCode(length int) string {
	sid := shortuuid.New()[:length]
	return sid
}
