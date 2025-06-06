package define

// OpenAI API相关配置常量
const (
	// OpenAI API端点
	OpenAIAPIURL    = "http://1.12.219.175:3001/v1/chat/completions"
	OpenAIModelsURL = "http://1.12.219.175:3001/v1/models"
)

// GenerateAIPromptRequest 表示生成AI提示的请求
type GenerateAIPromptRequest struct {
	// 系统提示，定义AI的行为或背景
	SystemPrompt string `json:"system_prompt" binding:"required"`
	// 用户提示，实际要问的问题
	UserPrompt string `json:"user_prompt" binding:"required"`
	// 可选参数
	Model       string  `json:"model" binding:"omitempty"`
	Temperature float64 `json:"temperature" binding:"omitempty"`
	MaxTokens   int     `json:"max_tokens" binding:"omitempty"`
	// 额外的上下文信息（可选）
	ContextData []string `json:"context_data" binding:"omitempty"`
}

// GenerateResponse 表示AI生成的响应
type GenerateResponse struct {
	Content    string `json:"content"`
	TokensUsed int    `json:"tokens_used"`
	Model      string `json:"model"`
	RequestID  string `json:"request_id"`
	Error      string `json:"error,omitempty"`
	StatusCode int    `json:"-"`
}

// ModelsResponse 表示模型列表响应
type ModelsResponse struct {
	Data []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
	Error      string `json:"error,omitempty"`
	StatusCode int    `json:"-"`
}
