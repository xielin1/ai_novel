package agent

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cloudwego/eino-examples/flow/agent/multiagent/plan_execute/debug"
	"github.com/cloudwego/eino-examples/flow/agent/multiagent/plan_execute/tools"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"

	"gin-template/service/agent/config"
	"gin-template/service/agent/core"
	"gin-template/service/agent/service"
	"gin-template/service/agent/session"
	"gin-template/service/agent/utils"
)

var (
	// 全局的智能体服务实例
	globalAgentService *service.MultiUserAgentService
)

func InitAgent() {
	ctx := context.Background()

	//初始化model
	deepSeekModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		Model:   os.Getenv("DEEPSEEK_MODEL_NAME"),
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		BaseURL: os.Getenv("DEEPSEEK_BASE_URL"),
	})
	if err != nil {
		log.Fatalf("new DeepSeek model failed: %v", err)
	}

	arkModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  os.Getenv("ARK_MODEL_NAME"),
	})
	if err != nil {
		log.Fatalf("new Ark model failed: %v", err)
	}

	toolsConfig, err := tools.GetTools(ctx)
	if err != nil {
		log.Fatalf("get tools config failed: %v", err)
	}

	// 创建多智能体的配置，system prompt 都用默认值
	agentConfig := &config.Config{
		// planner 在调试时大部分场景不需要真的去生成，可以用 mock 输出替代
		PlannerModel: &debug.ChatModelDebugDecorator{
			Model: deepSeekModel,
		},
		ExecutorModel: arkModel,
		ToolsConfig:   compose.ToolsNodeConfig{Tools: toolsConfig},
		ReviserModel: &debug.ChatModelDebugDecorator{
			Model: deepSeekModel,
		},
		ReviserSystemPrompt:  config.DefaultReviserPrompt,
		ExecutorSystemPrompt: config.DefaultExecutorPrompt,
		PlannerSystemPrompt:  config.DefaultPlannerPrompt,
	}

	// 创建智能体池配置
	poolConfig := &config.AgentPoolConfig{
		MinIdle:     2,           // 最小空闲实例数
		MaxActive:   10,          // 最大活跃实例数
		IdleTimeout: 300,         // 空闲超时时间
		AgentConfig: agentConfig, // 智能体配置
	}

	// 初始化全局智能体池
	agentPool, err := core.NewAgentPool(ctx, poolConfig)
	if err != nil {
		log.Fatalf("new agent pool failed: %v", err)
	}

	// 创建会话管理器
	sessionManager := session.NewSessionManager(agentPool, 30*time.Minute)

	// 创建智能体服务
	globalAgentService = service.NewMultiUserAgentService(sessionManager)

	// 创建一个测试用的智能体实例
	planExecuteAgent, err := core.NewMultiAgent(ctx, agentConfig)
	if err != nil {
		log.Fatalf("new plan execute multi agent failed: %v", err)
	}

	printer := utils.NewIntermediateOutputPrinter() // 创建一个中间结果打印器
	printer.PrintStream()                           // 开始异步输出到 console
	handler := printer.ToCallbackHandler()          // 转化为 Eino 框架的 callback handler

	// 以流式方式调用多智能体，实际的 OutputStream 不再需要关注，因为所有输出都由 intermediateOutputPrinter 处理了
	_, err = planExecuteAgent.Stream(ctx, []*schema.Message{schema.UserMessage("我们一家三口去乐园玩，孩子身高 120 cm，预算 2000 元，希望能尽可能多的看表演，游乐设施则比较偏爱刺激项目，希望能在一天内尽可能多体验不同的活动，请帮忙规划一个可操作的一日行程。我们会在乐园开门的时候入场，玩到晚上闭园的时候。")},
		agent.WithComposeOptions(compose.WithCallbacks(handler)), // 将中间结果打印的 callback handler 注入进来
	)
	if err != nil {
		log.Fatalf("stream error: %v", err)
	}

	printer.Wait() // 等待所有输出都处理完再结束
}

// GetGlobalAgentService 获取全局智能体服务实例
func GetGlobalAgentService() *service.MultiUserAgentService {
	return globalAgentService
}
