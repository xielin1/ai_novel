package config

import (
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// Config 多智能体的配置
type Config struct {
	// Planner 智能体使用的模型
	PlannerModel model.ChatModel
	// Executor 智能体使用的模型
	ExecutorModel model.ChatModel
	// Reviser 智能体使用的模型
	ReviserModel model.ChatModel
	// 工具配置
	ToolsConfig compose.ToolsNodeConfig
	// Planner 智能体的 system prompt
	PlannerSystemPrompt string
	// Executor 智能体的 system prompt
	ExecutorSystemPrompt string
	// Reviser 智能体的 system prompt
	ReviserSystemPrompt string
	// 最大执行步骤数量
	MaxStep int
}

// AgentPoolConfig 智能体池的配置
type AgentPoolConfig struct {
	// 最小空闲实例数
	MinIdle int
	// 最大活跃实例数
	MaxActive int
	// 空闲超时时间
	IdleTimeout int64
	// 智能体配置
	AgentConfig *Config
}
