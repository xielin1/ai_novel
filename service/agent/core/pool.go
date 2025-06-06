package core

import (
	"context"
	"sync"

	"gin-template/service/agent/config"
)

// AgentPool 智能体实例池
type AgentPool struct {
	agents      chan *PlanExecuteMultiAgent
	config      *config.AgentPoolConfig
	mutex       sync.Mutex
	createCount int
}

// NewAgentPool 创建新的智能体池
func NewAgentPool(ctx context.Context, poolConfig *config.AgentPoolConfig) (*AgentPool, error) {
	pool := &AgentPool{
		agents: make(chan *PlanExecuteMultiAgent, poolConfig.MaxActive),
		config: poolConfig,
	}

	// 预创建最小空闲实例
	for i := 0; i < poolConfig.MinIdle; i++ {
		agent, err := NewMultiAgent(ctx, poolConfig.AgentConfig)
		if err != nil {
			return nil, err
		}
		pool.agents <- agent
		pool.createCount++
	}

	return pool, nil
}

// BorrowAgent 借用智能体实例
func (p *AgentPool) BorrowAgent(ctx context.Context) (*PlanExecuteMultiAgent, error) {
	// 尝试从池中获取
	select {
	case agent := <-p.agents:
		return agent, nil
	default:
		// 池为空，检查是否可以创建新实例
		p.mutex.Lock()
		defer p.mutex.Unlock()

		if p.createCount < p.config.MaxActive {
			agent, err := NewMultiAgent(ctx, p.config.AgentConfig)
			if err != nil {
				return nil, err
			}
			p.createCount++
			return agent, nil
		}

		// 等待可用实例
		select {
		case agent := <-p.agents:
			return agent, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// ReturnAgent 将使用完的智能体实例返回到池中
func (p *AgentPool) ReturnAgent(agent *PlanExecuteMultiAgent) {
	select {
	case p.agents <- agent:
		// 成功将智能体返回到池中
	default:
		// 如果池已满，则减少创建计数
		p.mutex.Lock()
		p.createCount--
		p.mutex.Unlock()
	}
}
