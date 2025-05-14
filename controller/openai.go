package controller

import (
	"gin-template/define"
	"gin-template/service"

	"github.com/gin-gonic/gin"
)

// AIPrompt 智能组装提示并调用OpenAI API
func AIPrompt(c *gin.Context) {
	var req define.GenerateAIPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, "请求格式不正确: " + err.Error())
		return
	}

	// 调用服务层生成AI回复
	result, err := service.GenerateAICompletion(req)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}

	// 返回成功结果
	ResponseOK(c, result)
}

// 获取可用的模型列表
func GetAIModels(c *gin.Context) {
	// 调用服务层获取模型列表
	models, err := service.GetAvailableModels()
	if err != nil {
		ResponseError(c, err.Error())
		return
	}

	// 返回成功结果
	ResponseOK(c, models.Data)
} 