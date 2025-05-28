package define

// 套餐基本信息
type PackageInfo struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Price         float64  `json:"price"`
	MonthlyTokens int      `json:"monthly_tokens"`
	Duration      string   `json:"duration"` // monthly, yearly, permanent
	Features      []string `json:"features"`
}

// 套餐响应结构
type PackageResponse struct {
	Packages []PackageInfo `json:"packages"`
}

// 订阅信息
type SubscriptionInfo struct {
	PackageID   int64  `json:"package_id"`
	UserID      int64  `json:"user_id"`
	Status      string `json:"status"` // active, expired, cancelled
	StartDate   int64  `json:"start_date"`
	ExpiryDate  int64  `json:"expiry_date"`
	AutoRenew   bool   `json:"auto_renew"`
	NextRenewal int64  `json:"next_renewal_date,omitempty"`
}

// 订阅请求结构
type CreateSubscriptionRequest struct {
	PackageID     int64  `json:"package_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
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
	Package            PackageInfo      `json:"package"`
	Subscription       SubscriptionInfo `json:"subscription"`
	SubscriptionStatus string           `json:"subscription_status"`
	StartDate          int64            `json:"start_date"`
	ExpiryDate         int64            `json:"expiry_date"`
	AutoRenew          bool             `json:"auto_renew"`
	NextRenewalDate    int64            `json:"next_renewal_date,omitempty"`
}

// 取消自动续费响应
type CancelRenewalResponse struct {
	PackageName string `json:"package_name"`
	ExpiryDate  int64  `json:"expiry_date"`
	AutoRenew   bool   `json:"auto_renew"`
}

// PaymentRequest 支付请求
type PaymentRequest struct {
	PackageID     int    `json:"package_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=alipay wechat"`
}

// PaymentResponse 支付响应
type PaymentResponse struct {
	OrderID    string `json:"order_id"`
	PaymentURL string `json:"payment_url"`
}

// PaymentCallbackRequest 支付回调请求
type PaymentCallbackRequest struct {
	OrderID     string `json:"order_id" binding:"required"`
	PaymentID   string `json:"payment_id" binding:"required"`
	Status      string `json:"status" binding:"required,oneof=success failed"`
	PaymentTime string `json:"payment_time" binding:"required"`
	Sign        string `json:"sign" binding:"required"`
}
