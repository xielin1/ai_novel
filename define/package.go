package define

// 套餐基本信息
type PackageInfo struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Price         float64  `json:"price"`
	MonthlyTokens int      `json:"monthly_tokens"`
	Duration      string   `json:"duration"`    // monthly, yearly, permanent
	Features      []string `json:"features"`
}

// 套餐响应结构
type PackageResponse struct {
	Packages []PackageInfo `json:"packages"`
}

// 订阅信息
type SubscriptionInfo struct {
	PackageID    uint   `json:"package_id"`
	UserID       uint   `json:"user_id"`
	Status       string `json:"status"`        // active, expired, cancelled
	StartDate    string `json:"start_date"`
	ExpiryDate   string `json:"expiry_date"`
	AutoRenew    bool   `json:"auto_renew"`
	NextRenewal  string `json:"next_renewal_date,omitempty"`
}

// 订阅请求结构
type CreateSubscriptionRequest struct {
	PackageID      uint   `json:"package_id" binding:"required"`
	PaymentMethod  string `json:"payment_method" binding:"required"`
}

// 订阅响应结构
type SubscriptionResponse struct {
	OrderID       string  `json:"order_id"`
	PackageName   string  `json:"package_name"`
	Amount        float64 `json:"amount"`
	PaymentStatus string  `json:"payment_status"`
	ValidUntil    string  `json:"valid_until"`
	TokensAwarded int     `json:"tokens_awarded"`
	TokenBalance  int64   `json:"token_balance"`
}

// 用户当前套餐信息响应
type CurrentPackageResponse struct {
	Package          PackageInfo     `json:"package"`
	Subscription     SubscriptionInfo `json:"subscription"`
	SubscriptionStatus string         `json:"subscription_status"`
	StartDate         string         `json:"start_date"`
	ExpiryDate        string         `json:"expiry_date"`
	AutoRenew         bool           `json:"auto_renew"`
	NextRenewalDate   string         `json:"next_renewal_date,omitempty"`
}

// 取消自动续费响应
type CancelRenewalResponse struct {
	PackageName string `json:"package_name"`
	ExpiryDate  string `json:"expiry_date"`
	AutoRenew   bool   `json:"auto_renew"`
} 