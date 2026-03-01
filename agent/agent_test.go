package agent

import (
	"context"
	"testing"

	"golangchain/models"
)

// MockChatModel 用于测试的模拟ChatModel
type MockChatModel struct {
	response string
}

func (m *MockChatModel) Chat(ctx context.Context, messages []models.Message, options ...models.Option) (models.ChatResponse, error) {
	return models.ChatResponse{
		Message: models.Message{
			Role:    models.RoleAssistant,
			Content: m.response,
		},
		TokensUsed:   10,
		FinishReason: "stop",
	}, nil
}

func (m *MockChatModel) ChatStream(ctx context.Context, messages []models.Message, options ...models.Option) (<-chan models.ChatChunk, error) {
	resultChan := make(chan models.ChatChunk, 1)
	resultChan <- models.ChatChunk{
		Delta: models.Message{
			Role:    models.RoleAssistant,
			Content: m.response,
		},
		Done:         true,
		FinishReason: "stop",
	}
	close(resultChan)
	return resultChan, nil
}

// TestNewAgent 测试Agent创建
func TestNewAgent(t *testing.T) {
	mockModel := &MockChatModel{}
	agent := NewAgent(mockModel, 5)

	if agent.model != mockModel {
		t.Error("Expected model to be set")
	}

	if agent.maxSteps != 5 {
		t.Errorf("Expected maxSteps to be 5, got %d", agent.maxSteps)
	}

	if len(agent.tools) != 0 {
		t.Errorf("Expected no tools initially, got %d", len(agent.tools))
	}
}

// TestRegisterTool 测试工具注册
func TestRegisterTool(t *testing.T) {
	mockModel := &MockChatModel{}
	agent := NewAgent(mockModel, 5)

	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]interface{}{
			"param1": "string",
		},
	}

	executor := func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	err := agent.RegisterTool(tool, executor)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(agent.tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(agent.tools))
	}

	if _, exists := agent.tools["test_tool"]; !exists {
		t.Error("Expected tool to be registered")
	}
}

// TestRegisterToolWithEmptyName 测试注册空名称的工具
func TestRegisterToolWithEmptyName(t *testing.T) {
	mockModel := &MockChatModel{}
	agent := NewAgent(mockModel, 5)

	tool := Tool{
		Name:        "",
		Description: "A test tool",
	}

	executor := func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	err := agent.RegisterTool(tool, executor)
	if err == nil {
		t.Error("Expected error for empty tool name")
	}
}

// TestRegisterToolWithNilExecutor 测试注册nil执行器
func TestRegisterToolWithNilExecutor(t *testing.T) {
	mockModel := &MockChatModel{}
	agent := NewAgent(mockModel, 5)

	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
	}

	err := agent.RegisterTool(tool, nil)
	if err == nil {
		t.Error("Expected error for nil executor")
	}
}

// TestParseResponse 测试响应解析
func TestParseResponse(t *testing.T) {
	mockModel := &MockChatModel{}
	agent := NewAgent(mockModel, 5)

	// 测试纯文本响应
	toolCalls, text := agent.parseResponse("This is a plain text response")
	if len(toolCalls) != 0 {
		t.Errorf("Expected no tool calls, got %d", len(toolCalls))
	}
	if text != "This is a plain text response" {
		t.Errorf("Expected text to be preserved, got %s", text)
	}

	// 测试JSON格式的工具调用
	jsonResponse := `{"tool_calls": [{"tool_name": "test", "input": {"key": "value"}}], "text": "Using tool"}`
	toolCalls, text = agent.parseResponse(jsonResponse)
	if len(toolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(toolCalls))
	}
	if toolCalls[0].ToolName != "test" {
		t.Errorf("Expected tool name to be 'test', got %s", toolCalls[0].ToolName)
	}
}

// TestGetTools 测试获取工具列表
func TestGetTools(t *testing.T) {
	mockModel := &MockChatModel{}
	agent := NewAgent(mockModel, 5)

	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
	}

	executor := func(ctx context.Context, input map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	agent.RegisterTool(tool, executor)

	tools := agent.GetTools()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	if _, exists := tools["test_tool"]; !exists {
		t.Error("Expected test_tool to exist")
	}
}

// BenchmarkNewAgent 基准测试Agent创建
func BenchmarkNewAgent(b *testing.B) {
	mockModel := &MockChatModel{}
	for i := 0; i < b.N; i++ {
		NewAgent(mockModel, 5)
	}
}
