package controller

import (
	"gin-template/define"
	"gin-template/service/agent"
	"gin-template/service/agent/service"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

// AgentController 智能体控制器
type AgentController struct {
	agentService *service.MultiUserAgentService
}

// NewAgentController 创建智能体控制器
func NewAgentController() *AgentController {
	return &AgentController{
		agentService: agent.GetGlobalAgentService(),
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	SessionID string `json:"session_id" binding:"required"` // 会话ID
	Message   string `json:"message" binding:"required"`    // 用户消息
}

// ChatResponse 聊天响应
type ChatResponse struct {
	SessionID string `json:"session_id"` // 会话ID
	Response  string `json:"response"`   // 智能体响应
}

// Chat 处理聊天请求
// @Summary 与智能体进行对话
// @Description 发送消息给智能体并获取响应
// @Tags 智能体
// @Accept json
// @Produce json
// @Param request body ChatRequest true "聊天请求"
// @Success 200 {object} ChatResponse
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/agent/chat [post]
func (c *AgentController) Chat(ctx *gin.Context) {
	var req ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ResponseError(ctx, "无效的请求参数")
		return
	}

	// 创建用户消息
	message := schema.UserMessage(req.Message)

	// 调用智能体服务生成响应
	response, err := c.agentService.Generate(ctx, &define.GenerateRequest{
		SessionID: req.SessionID,
		Messages:  []*schema.Message{message},
	})
	if err != nil {
		ResponseErrorWithStatus(ctx, 500, "生成响应失败: "+err.Error())
		return
	}

	ResponseOK(ctx, ChatResponse{
		SessionID: response.SessionID,
		Response:  response.Message.Content,
	})
}

// StreamChat 处理流式聊天请求
// @Summary 与智能体进行流式对话
// @Description 发送消息给智能体并获取流式响应
// @Tags 智能体
// @Accept json
// @Produce text/event-stream
// @Param request body ChatRequest true "聊天请求"
// @Success 200 {object} ChatResponse
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/agent/chat/stream [post]
//func (c *AgentController) StreamChat(ctx *gin.Context) {
//	var req ChatRequest
//	if err := ctx.ShouldBindJSON(&req); err != nil {
//		ResponseError(ctx, "无效的请求参数")
//		return
//	}
//
//	// 设置SSE响应头
//	ctx.Header("Content-Type", "text/event-stream")
//	ctx.Header("Cache-Control", "no-cache")
//	ctx.Header("Connection", "keep-alive")
//	ctx.Header("Transfer-Encoding", "chunked")
//
//	// 创建用户消息
//	message := schema.UserMessage(req.Message)
//
//	// 获取会话
//	agent, err := c.agentService.SessionManager.GetOrCreateAgentWithSession(ctx, req.SessionID)
//	if err != nil {
//		ResponseErrorWithStatus(ctx, 500, "获取会话失败: "+err.Error())
//		return
//	}
//
//	// 调用智能体生成流式响应
//	stream, err := agent.Stream(ctx, []*schema.Message{message})
//	if err != nil {
//		ResponseErrorWithStatus(ctx, 500, "生成流式响应失败: "+err.Error())
//		return
//	}
//
//	// 发送流式响应
//	ctx.Stream(func(w io.Writer) bool {
//		msg, err := stream.Recv()
//		if err != nil {
//			if err == io.EOF {
//				return false
//			}
//			ctx.SSEvent("error", gin.H{
//				"error": err.Error(),
//			})
//			return false
//		}
//
//		ctx.SSEvent("message", gin.H{
//			"session_id": req.SessionID,
//			"content":    msg.Content,
//		})
//		return true
//	})
//}
