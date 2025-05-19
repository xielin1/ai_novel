package define

// 推荐码请求
type ReferralCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

// 推荐码信息响应
type ReferralCodeResponse struct {
	ReferralCode      string `json:"referral_code"`
	TotalReferred     int    `json:"total_referred"`
	TotalTokensEarned int    `json:"total_tokens_earned"`
	ShareURL          string `json:"share_url"`
}

// 推荐记录列表响应
type ReferralsListResponse struct {
	Referrals  []ReferralDetail `json:"referrals"`
	Statistics ReferralStatistics `json:"statistics"`
	Pagination PaginationInfo `json:"pagination"`
}

// 推荐统计信息
type ReferralStatistics struct {
	TotalReferred     int   `json:"total_referred"`
	TotalTokensEarned int   `json:"total_tokens_earned"`
}

// 分页信息
type PaginationInfo struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Pages int `json:"pages"`
}

// 推荐明细信息
type ReferralDetail struct {
	ID          uint   `json:"id"`
	ReferrerID  uint   `json:"referrer_id"`
	ReferredID  uint   `json:"referred_id"`
	ReferredName string `json:"referred_name"`
	Code        string `json:"code"`
	CreatedAt   string `json:"created_at"`
	TokensEarned int   `json:"tokens_earned"`
	Status      string `json:"status"`
}

// 新推荐码响应
type NewReferralCodeResponse struct {
	PreviousCode string `json:"previous_code"`
	NewCode      string `json:"new_code"`
	ShareURL     string `json:"share_url"`
}

// 使用推荐码响应
type UseReferralCodeResponse struct {
	TokensRewarded int64 `json:"tokens_rewarded"`
	NewBalance     int64 `json:"new_balance"`
} 