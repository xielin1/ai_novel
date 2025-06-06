package session

import (
	"fmt"
	"gin-template/define"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"gin-template/service/agent/core"
)

// SessionManager 会话管理器
type SessionManager struct {
	sessions  map[string]*define.SessionState
	agentPool *core.AgentPool
	mutex     sync.RWMutex
	ttl       time.Duration // 会话过期时间
}

// NewSessionManager 创建新的会话管理器
func NewSessionManager(agentPool *core.AgentPool, ttl time.Duration) *SessionManager {
	return &SessionManager{
		sessions:  make(map[string]*define.SessionState),
		agentPool: agentPool,
		ttl:       ttl,
	}
}

// GetOrCreateSession 获取或创建会话
func (s *SessionManager) GetOrCreateSession(id string) (*define.SessionState, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if id == "" {
		// 创建新会话
		newID := uuid.New().String()
		session := &define.SessionState{
			ID:        newID,
			Messages:  make([]*schema.Message, 0),
			UserInfo:  make(map[string]string),
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
		}
		s.sessions[newID] = session
		return session, nil
	}

	// 获取现有会话
	session, exists := s.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	session.LastUsed = time.Now()
	return session, nil
}

// GetAgentPool 获取智能体池
func (s *SessionManager) GetAgentPool() *core.AgentPool {
	return s.agentPool
}
