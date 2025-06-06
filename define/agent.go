package define

import (
	"time"

	"github.com/cloudwego/eino/schema"
)

type SessionState struct {
	ID        string
	Messages  []*schema.Message // 历史消息
	UserInfo  map[string]string // 用户信息
	CreatedAt time.Time         // 创建时间
	LastUsed  time.Time         // 最后访问时间
}

type GenerateRequest struct {
	SessionID string            `json:"session_id"` // 为空则创建新会话
	Messages  []*schema.Message `json:"messages"`
	UserInfo  map[string]string `json:"user_info,omitempty"`
}

type GenerateResponseForAgent struct {
	SessionID string          `json:"session_id"`
	Message   *schema.Message `json:"message"`
}
