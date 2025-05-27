package define

// PaymentStatus 定义支付状态
type PaymentStatus int64

const (
	PaymentStatusPending PaymentStatus = 0 // 待支付
	PaymentStatusSuccess PaymentStatus = 1 // 支付成功
	PaymentStatusFailed  PaymentStatus = 2 // 支付失败
	PaymentStatusClosed  PaymentStatus = 3 // 交易关闭
)

// CreatePaymentRequest 定义创建支付请求的结构体
type CreatePaymentRequest struct {
	PackageID int64 `json:"package_id"` // 套餐 ID
}

// CreatePaymentResponse 定义创建支付响应的结构体
// 这里需要包含调用支付宝接口后返回的支付信息，例如：
// - 网页支付的跳转 URL
// - 扫码支付的二维码数据
// - App 支付的支付参数
type CreatePaymentResponse struct {
	OrderID       string `json:"order_id"`          // 内部订单号
	AlipayTradeNo string `json:"alipay_trade_no"`   // 支付宝交易号 (创建交易时可能没有，回调时会有)
	PayURL        string `json:"pay_url,omitempty"` // 支付跳转 URL 或二维码数据等，根据支付方式而定
	// ... 其他支付宝返回的支付信息
}

// AlipayCallbackRequest 定义支付宝回调请求的结构体
// 这里的字段需要根据支付宝的异步通知参数来定义
// 例如：
// - trade_status 交易状态
// - out_trade_no 商户订单号 (即内部订单号)
// - trade_no 支付宝交易号
// - total_amount 交易金额
// - seller_id 卖家支付宝用户号
// - app_id 支付宝分配给开发者的应用 ID
// - sign 签名
// - sign_type 签名类型
// ... 其他支付宝回调参数
type AlipayCallbackRequest struct {
	TradeStatus string `form:"trade_status"`
	OutTradeNo  string `form:"out_trade_no"`
	TradeNo     string `form:"trade_no"`
	TotalAmount string `form:"total_amount"` // 注意支付宝回调的金额是字符串
	SellerID    string `form:"seller_id"`
	AppID       string `form:"app_id"`
	Sign        string `form:"sign"`
	SignType    string `form:"sign_type"`
	// ... 添加其他重要的回调参数
}

// Payment 定义内部支付订单的结构体，用于存储支付信息到数据库
type Payment struct {
	ID            int64         `gorm:"column:id;primarykey"`
	OrderID       string        `gorm:"column:order_id;uniqueIndex"` // 内部订单号，唯一索引
	UserID        int64         `gorm:"column:user_id"`
	PackageID     int64         `gorm:"column:package_id"`
	Amount        int64         `gorm:"column:amount"` // 订单金额，以"分"为单位存储
	Status        PaymentStatus `gorm:"column:status"`
	AlipayTradeNo string        `gorm:"column:alipay_trade_no"` // 支付宝交易号
	CreateTime    int64         `gorm:"column:create_time"`     // 创建时间 (通过 GORM 插件自动处理)
	UpdateTime    int64         `gorm:"column:update_time"`     // 更新时间 (通过 GORM 插件自动处理)
	// ... 可能还需要其他字段，例如支付渠道、支付时间等
}

// TableName 指定 Payment 结构体对应的数据库表名
func (Payment) TableName() string {
	return "payments" // 假设支付表名为 'payments'
}
