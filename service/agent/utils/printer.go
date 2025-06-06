package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	callbacks2 "github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino/utils/callbacks"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	White  = "\033[97m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
)

// IntermediateOutputPrinter 利用 Eino 的 callback 机制，收集多智能体各步骤的实时输出
type IntermediateOutputPrinter struct {
	ch               chan coloredString
	currentAgentName string          // 当前智能体名称
	agentReasoning   map[string]bool // 智能体处在"推理"阶段还是"最终答案"阶段
	mu               sync.Mutex
	wg               sync.WaitGroup
}

type coloredString struct {
	str  string
	code string
}

// NewIntermediateOutputPrinter 创建新的中间输出打印器
func NewIntermediateOutputPrinter() *IntermediateOutputPrinter {
	return &IntermediateOutputPrinter{
		ch: make(chan coloredString),
		agentReasoning: map[string]bool{
			"planner":  false,
			"executor": false,
			"reviser":  false,
		},
	}
}

// PrintStream 开始异步输出到控制台
func (s *IntermediateOutputPrinter) PrintStream() {
	go func() {
		for m := range s.ch {
			fmt.Print(m.code + m.str + Reset)
		}
	}()
}

// ToCallbackHandler 转化为 Eino 框架的 callback handler
func (s *IntermediateOutputPrinter) ToCallbackHandler() callbacks2.Handler {
	return callbacks.NewHandlerHelper().ChatModel(&callbacks.ModelCallbackHandler{
		OnEndWithStreamOutput: s.onChatModelEndWithStreamOutput,
	}).Tool(&callbacks.ToolCallbackHandler{
		OnStart: s.onToolStart,
		OnEnd:   s.onToolEnd,
	}).Handler()
}

// Wait 等待所有输出都处理完
func (s *IntermediateOutputPrinter) Wait() {
	s.wg.Wait()
}

// onChatModelEndWithStreamOutput 当 ChatModel 结束时，获取它的流式输出并格式化处理
func (s *IntermediateOutputPrinter) onChatModelEndWithStreamOutput(ctx context.Context, runInfo *callbacks2.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
	name := runInfo.Name
	if name != s.currentAgentName {
		s.ch <- coloredString{fmt.Sprintf("\n\n=======\n%s: \n=======\n", name), Cyan}
		s.currentAgentName = name
	}

	s.wg.Add(1)

	go func() {
		defer output.Close()
		defer s.wg.Done()

		for {
			chunk, err := output.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				log.Fatalf("internal error: %s\n", err)
			}

			if len(chunk.Message.Content) > 0 {
				if s.agentReasoning[name] { // 切换到最终答案阶段
					s.ch <- coloredString{"\nanswer begin: \n", Green}
					s.mu.Lock()
					s.agentReasoning[name] = false
					s.mu.Unlock()
				}
				s.ch <- coloredString{chunk.Message.Content, Yellow}
			} else if reasoningContent, ok := deepseek.GetReasoningContent(chunk.Message); ok {
				if !s.agentReasoning[name] { // 切换到推理阶段
					s.ch <- coloredString{"\nreasoning begin: \n", Green}
					s.mu.Lock()
					s.agentReasoning[name] = true
					s.mu.Unlock()
				}
				s.ch <- coloredString{reasoningContent, White}
			}
		}
	}()

	return ctx
}

// onToolStart 当 Tool 执行开始时，获取并输出调用信息
func (s *IntermediateOutputPrinter) onToolStart(ctx context.Context, info *callbacks2.RunInfo, input *tool.CallbackInput) context.Context {
	arguments := make(map[string]any)
	err := sonic.Unmarshal([]byte(input.ArgumentsInJSON), &arguments)
	if err != nil {
		s.ch <- coloredString{fmt.Sprintf("\ncall %s: %s\n", info.Name, input.ArgumentsInJSON), Red}
		return ctx
	}

	formatted, err := sonic.MarshalIndent(arguments, "  ", "  ")
	if err != nil {
		s.ch <- coloredString{fmt.Sprintf("\ncall %s: %s\n", info.Name, input.ArgumentsInJSON), Red}
		return ctx
	}

	s.ch <- coloredString{fmt.Sprintf("\ncall %s: %s\n", info.Name, string(formatted)), Red}
	return ctx
}

// onToolEnd 当 Tool 执行结束时，获取并输出返回结果
func (s *IntermediateOutputPrinter) onToolEnd(ctx context.Context, info *callbacks2.RunInfo, output *tool.CallbackOutput) context.Context {
	response := make(map[string]any)
	err := sonic.Unmarshal([]byte(output.Response), &response)
	if err != nil {
		s.ch <- coloredString{fmt.Sprintf("\ncall %s: %s\n", info.Name, output.Response), Blue}
		return ctx
	}

	formatted, err := sonic.MarshalIndent(response, "  ", "  ")
	if err != nil {
		s.ch <- coloredString{fmt.Sprintf("\ncall %s: %s\n", info.Name, output.Response), Blue}
		return ctx
	}

	s.ch <- coloredString{fmt.Sprintf("\ncall %s result: %s\n", info.Name, string(formatted)), Blue}
	return ctx
}
