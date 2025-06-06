package service

import (
	"context"
	"gin-template/define"

	"gin-template/service/agent/session"
)

// MultiUserAgentService 多用户智能体服务
type MultiUserAgentService struct {
	sessionManager *session.SessionManager
}

// NewMultiUserAgentService 创建新的多用户智能体服务
func NewMultiUserAgentService(sessionManager *session.SessionManager) *MultiUserAgentService {
	return &MultiUserAgentService{
		sessionManager: sessionManager,
	}
}

// Generate 生成回复
func (s *MultiUserAgentService) Generate(ctx context.Context, req *define.GenerateRequest) (*define.GenerateResponseForAgent, error) {
	// 获取会话
	session, err := s.sessionManager.GetOrCreateSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	// 更新会话消息
	allMessages := append(session.Messages, req.Messages...)

	// 借用智能体实例
	agent, err := s.sessionManager.GetAgentPool().BorrowAgent(ctx)
	if err != nil {
		return nil, err
	}
	defer s.sessionManager.GetAgentPool().ReturnAgent(agent)

	// 调用智能体生成回复
	result, err := agent.Generate(ctx, allMessages)
	if err != nil {
		return nil, err
	}

	// 更新会话状态
	session.Messages = append(allMessages, result)

	return &define.GenerateResponseForAgent{
		SessionID: session.ID,
		Message:   result,
	}, nil
}
