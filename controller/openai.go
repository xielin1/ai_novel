package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gin-template/model"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// OpenAI API相关配置
const (
	// OpenAI API端点
	OpenAIAPIURL = "http://1.12.219.175:3001/v1/chat/completions"
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
	Content     string `json:"content"`
	TokensUsed  int    `json:"tokens_used"`
	Model       string `json:"model"`
	RequestID   string `json:"request_id"`
	Error       string `json:"error,omitempty"`
	StatusCode  int    `json:"-"`
}

// AIPrompt 智能组装提示并调用OpenAI API
func AIPrompt(c *gin.Context) {
	var req GenerateAIPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式不正确", "details": err.Error()})
		return
	}

	// 设置默认值
	if req.Model == "" {
		req.Model = "gpt-3.5-turbo"
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 1000
	}

	// 组装消息数组
	messages := []map[string]string{
		{
			"role":    "system",
			"content": req.SystemPrompt,
		},
	}

	// 添加上下文信息（如果有）
	for _, ctx := range req.ContextData {
		messages = append(messages, map[string]string{
			"role":    "assistant",
			"content": ctx,
		})
	}

	// 添加用户消息
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": req.UserPrompt,
	})

	// 准备OpenAI API请求
	requestBody := map[string]interface{}{
		"model":       req.Model,
		"messages":    messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法序列化请求", "details": err.Error()})
		return
	}

	// 获取OpenAI API密钥
	apiKey := model.GetSetting("openai_api_key")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "未配置OpenAI API密钥"})
		return
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", OpenAIAPIURL, bytes.NewBuffer(requestJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建HTTP请求失败", "details": err.Error()})
		return
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用OpenAI API失败", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	// 处理响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取API响应失败", "details": err.Error()})
		return
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		errResponse := struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}{}
		if err := json.Unmarshal(body, &errResponse); err == nil && errResponse.Error.Message != "" {
			c.JSON(resp.StatusCode, gin.H{"error": errResponse.Error.Message, "type": errResponse.Error.Type})
		} else {
			c.JSON(resp.StatusCode, gin.H{"error": fmt.Sprintf("API返回错误: %d", resp.StatusCode), "response": string(body)})
		}
		return
	}

	// 解析响应
	var response struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		Model   string `json:"model"`
		Usage   struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
			Index        int    `json:"index"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析API响应失败", "details": err.Error()})
		return
	}

	// 处理空响应
	if len(response.Choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API没有返回有效的回复"})
		return
	}

	result := GenerateResponse{
		Content:    response.Choices[0].Message.Content,
		TokensUsed: response.Usage.TotalTokens,
		Model:      response.Model,
		RequestID:  response.ID,
		StatusCode: resp.StatusCode,
	}

	c.JSON(http.StatusOK, result)
}

// 获取可用的模型列表
func GetAIModels(c *gin.Context) {
	// 获取OpenAI API密钥
	apiKey := model.GetSetting("openai_api_key")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "未配置OpenAI API密钥"})
		return
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("GET", "http://1.12.219.175:3001/v1/models", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建HTTP请求失败", "details": err.Error()})
		return
	}

	// 设置请求头
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用OpenAI API失败", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取API响应失败", "details": err.Error()})
		return
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": fmt.Sprintf("API返回错误: %d", resp.StatusCode), "response": string(body)})
		return
	}

	// 解析响应
	var modelsResponse struct {
		Object string `json:"object"`
		Data   []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(body, &modelsResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析API响应失败", "details": err.Error()})
		return
	}
	
	// 返回格式化的响应，与前端期望的结构匹配
	c.JSON(http.StatusOK, gin.H{
		"data": modelsResponse.Data,
	})
} 