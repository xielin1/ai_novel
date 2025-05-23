package define

type ReferralCodeRequest struct {
	Code string `json:"code" binding:"required"`
}
type UseReferralRequest struct {
	ReferralCode string `json:"referralCode" binding:"required"`
}

type ReferralCodeResponse struct {
	ReferralCode      string `json:"referral_code"`
	TotalReferred     int    `json:"total_referred"`
	TotalTokensEarned int    `json:"total_tokens_earned"`
	ShareURL          string `json:"share_url"`
}

type ReferralsListResponse struct {
	Referrals  []ReferralDetail   `json:"referrals"`
	Statistics ReferralStatistics `json:"statistics"`
	Pagination PaginationInfo     `json:"pagination"`
}

type ReferralStatistics struct {
	TotalReferred     int `json:"total_referred"`
	TotalTokensEarned int `json:"total_tokens_earned"`
}

type PaginationInfo struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Pages int `json:"pages"`
}

type ReferralDetail struct {
	ID           int64  `json:"id"`
	ReferrerID   int64  `json:"referrer_id"`
	ReferredID   int64  `json:"referred_id"`
	ReferredName string `json:"referred_name"`
	Code         string `json:"code"`
	CreatedAt    string `json:"created_at"`
	TokensEarned int    `json:"tokens_earned"`
	Status       string `json:"status"`
}

type NewReferralCodeResponse struct {
	NewCode  string `json:"new_code"`
	ShareURL string `json:"share_url"`
}

type UseReferralCodeResponse struct {
	TokensRewarded int64 `json:"tokens_rewarded"`
	NewBalance     int64 `json:"new_balance"`
}
