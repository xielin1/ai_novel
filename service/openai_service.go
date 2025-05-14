package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gin-template/define"
	"gin-template/model"
	"io/ioutil"
	"net/http"
)

// OpenAI API相关配置
const (
	// OpenAI API端点
	OpenAIAPIURL = "http://1.12.219.175:3001/v1/chat/completions"
	OpenAIModelsURL = "http://1.12.219.175:3001/v1/models"
)



// GenerateAICompletion 调用OpenAI API生成补全
func GenerateAICompletion(req define.GenerateAIPromptRequest) (define.GenerateResponse, error) {
	var result define.GenerateResponse

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
		return result, fmt.Errorf("无法序列化请求: %v", err)
	}

	// 获取OpenAI API密钥
	apiKey := model.GetSetting("openai_api_key")
	if apiKey == "" {
		return result, fmt.Errorf("未配置OpenAI API密钥")
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", OpenAIAPIURL, bytes.NewBuffer(requestJSON))
	if err != nil {
		return result, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return result, fmt.Errorf("调用OpenAI API失败: %v", err)
	}
	defer resp.Body.Close()

	// 处理响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("读取API响应失败: %v", err)
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
			return result, fmt.Errorf("API错误: %s", errResponse.Error.Message)
		} else {
			return result, fmt.Errorf("API返回错误: %d", resp.StatusCode)
		}
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
		return result, fmt.Errorf("解析API响应失败: %v", err)
	}

	// 处理空响应
	if len(response.Choices) == 0 {
		return result, fmt.Errorf("API没有返回有效的回复")
	}

	result = define.GenerateResponse{
		Content:    response.Choices[0].Message.Content,
		TokensUsed: response.Usage.TotalTokens,
		Model:      response.Model,
		RequestID:  response.ID,
		StatusCode: resp.StatusCode,
	}

	return result, nil
}

// GetAvailableModels 获取可用的模型列表
func GetAvailableModels() (define.ModelsResponse, error) {
	var result define.ModelsResponse

	// 获取OpenAI API密钥
	apiKey := model.GetSetting("openai_api_key")
	if apiKey == "" {
		return result, fmt.Errorf("未配置OpenAI API密钥")
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("GET", OpenAIModelsURL, nil)
	if err != nil {
		return result, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return result, fmt.Errorf("调用OpenAI API失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("读取API响应失败: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("API返回错误: %d", resp.StatusCode)
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
		return result, fmt.Errorf("解析API响应失败: %v", err)
	}

	result.Data = modelsResponse.Data
	result.StatusCode = resp.StatusCode

	return result, nil
} 