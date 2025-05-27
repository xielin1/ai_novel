package define

import "time"

const (
	TokenTransactionTypeInitial                  = "initial"
	TokenTransactionTypeReconciliationAdjustment = "reconciliation_adjustment"
	TokenTransactionTypeOutlineDebit             = "outline_debit"
	TokenTransactionTypeContentDebit             = "content_debit"
	TokenTransactionTypePackageCredit            = "package_credit"
	TokenTransactionTypeReferralCredit           = "referral_credit"
)

const (
	TransactionStatusCompleted = "completed"
)

// TokenBalance 用户Token余额信息
type TokenBalance struct {
	UserID    uint      `json:"user_id"`
	Balance   int64     `json:"balance"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TokenTransaction 代表Token交易记录
type TokenTransaction struct {
	ID                uint      `json:"id"`
	TransactionUUID   string    `json:"transaction_uuid"`
	UserID            uint      `json:"user_id"`
	Amount            int64     `json:"amount"` // 正数为增加，负数为减少
	BalanceBefore     int64     `json:"balance_before"`
	BalanceAfter      int64     `json:"balance_after"`
	Type              string    `json:"type"`                          // 例如 "ai_generation_debit", "referral_credit"
	RelatedEntityType string    `json:"related_entity_type,omitempty"` // 例如 "project", "order"
	RelatedEntityID   string    `json:"related_entity_id,omitempty"`   // 关联实体的ID
	Description       string    `json:"description"`
	Status            string    `json:"status"` // 例如 "completed", "pending", "failed"
	CreatedAt         time.Time `json:"created_at"`
}

// TokenTransactionRequest 创建Token交易的请求
type TokenTransactionRequest struct {
	UserID            uint   `json:"user_id" binding:"required"`
	Amount            int64  `json:"amount" binding:"required"`
	TransactionUUID   string `json:"transaction_uuid,omitempty"` // 可选，用于幂等性
	Type              string `json:"type" binding:"required"`
	Description       string `json:"description"`
	RelatedEntityType string `json:"related_entity_type,omitempty"`
	RelatedEntityID   string `json:"related_entity_id,omitempty"`
}

// TokenTransactionList Token交易记录列表及分页信息
type TokenTransactionList struct {
	Transactions []TokenTransaction `json:"transactions"`
	Total        int64              `json:"total"`
	Page         int                `json:"page"`
	Limit        int                `json:"limit"`
	Pages        int                `json:"pages"`
}

// TokenInitRequest 初始化用户Token账户的请求
type TokenInitRequest struct {
	UserID         uint  `json:"user_id" binding:"required"`
	InitialBalance int64 `json:"initial_balance"`
}

// TokenError Token操作相关错误
type TokenError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
