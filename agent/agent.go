// Package agent 提供Agent框架的核心实现
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"golangchain/models"
)

// Tool 定义Agent可以使用的工具
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolExecutor 定义工具执行函数
type ToolExecutor func(ctx context.Context, input map[string]interface{}) (interface{}, error)

// Agent 定义Agent结构
type Agent struct {
	model     models.ChatModel
	tools     map[string]Tool
	executors map[string]ToolExecutor
	maxSteps  int
}

// AgentResponse 定义Agent的响应
type AgentResponse struct {
	Text      string
	ToolCalls []ToolCall
	Done      bool
}

// ToolCall 定义工具调用
type ToolCall struct {
	ToolName string                 `json:"tool_name"`
	Input    map[string]interface{} `json:"input"`
}

// NewAgent 创建新的Agent实例
func NewAgent(model models.ChatModel, maxSteps int) *Agent {
	return &Agent{
		model:     model,
		tools:     make(map[string]Tool),
		executors: make(map[string]ToolExecutor),
		maxSteps:  maxSteps,
	}
}

// RegisterTool 注册一个工具
func (a *Agent) RegisterTool(tool Tool, executor ToolExecutor) error {
	if tool.Name == "" {
		return errors.New("tool name cannot be empty")
	}
	if executor == nil {
		return errors.New("executor cannot be nil")
	}
	a.tools[tool.Name] = tool
	a.executors[tool.Name] = executor
	return nil
}

// Run 运行Agent
func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
	messages := []models.Message{
		models.NewSystemMessage(a.buildSystemPrompt()),
		models.NewUserMessage(prompt),
	}

	for step := 0; step < a.maxSteps; step++ {
		// 调用模型
		response, err := a.model.Chat(ctx, messages)
		if err != nil {
			return "", fmt.Errorf("model chat error: %w", err)
		}

		// 检查是否需要调用工具
		toolCalls, text := a.parseResponse(response.Message.Content)

		if len(toolCalls) == 0 {
			// 没有工具调用，返回最终答案
			return text, nil
		}

		// 执行工具调用
		messages = append(messages, response.Message)

		for _, call := range toolCalls {
			result, err := a.executeTool(ctx, call)
			if err != nil {
				messages = append(messages, models.NewFunctionMessage(
					fmt.Sprintf("Error: %v", err),
					call.ToolName,
				))
			} else {
				resultJSON, _ := json.Marshal(result)
				messages = append(messages, models.NewFunctionMessage(
					string(resultJSON),
					call.ToolName,
				))
			}
		}
	}

	return "", errors.New("max steps exceeded")
}

// buildSystemPrompt 构建系统提示
func (a *Agent) buildSystemPrompt() string {
	prompt := "You are a helpful assistant with access to the following tools:\n\n"

	for _, tool := range a.tools {
		prompt += fmt.Sprintf("Tool: %s\n", tool.Name)
		prompt += fmt.Sprintf("Description: %s\n", tool.Description)
		prompt += fmt.Sprintf("Parameters: %v\n\n", tool.Parameters)
	}

	prompt += "When you need to use a tool, respond with JSON in this format:\n"
	prompt += `{"tool_calls": [{"tool_name": "tool_name", "input": {...}}], "text": "explanation"}`
	prompt += "\n\nIf you don't need to use any tools, just respond with your answer."

	return prompt
}

// parseResponse 解析模型响应
func (a *Agent) parseResponse(content string) ([]ToolCall, string) {
	// 尝试解析JSON格式的工具调用
	var response struct {
		ToolCalls []ToolCall `json:"tool_calls"`
		Text      string     `json:"text"`
	}

	if err := json.Unmarshal([]byte(content), &response); err == nil && len(response.ToolCalls) > 0 {
		return response.ToolCalls, response.Text
	}

	// 如果不是JSON格式，直接返回文本
	return nil, content
}

// executeTool 执行工具
func (a *Agent) executeTool(ctx context.Context, call ToolCall) (interface{}, error) {
	executor, exists := a.executors[call.ToolName]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", call.ToolName)
	}

	return executor(ctx, call.Input)
}

// GetTools 获取所有已注册的工具
func (a *Agent) GetTools() map[string]Tool {
	return a.tools
}
