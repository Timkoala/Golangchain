package anthropic

import (
	"context"
	"testing"
	"time"

	"golangchain/models"
)

// TestNewModel 测试模型创建
func TestNewModel(t *testing.T) {
	model := NewModel("test-api-key", "claude-3-opus-20240229",
		models.WithMaxTokens(200),
		models.WithTemperature(0.8),
	)

	if model.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got %s", model.apiKey)
	}

	if model.modelID != "claude-3-opus-20240229" {
		t.Errorf("Expected modelID to be 'claude-3-opus-20240229', got %s", model.modelID)
	}

	if model.defaults.MaxTokens != 200 {
		t.Errorf("Expected MaxTokens to be 200, got %d", model.defaults.MaxTokens)
	}

	if model.defaults.Temperature != 0.8 {
		t.Errorf("Expected Temperature to be 0.8, got %f", model.defaults.Temperature)
	}
}

// TestInterfaceCompliance 测试接口合规性
func TestInterfaceCompliance(t *testing.T) {
	model := NewModel("test-key", "claude-3-opus-20240229")

	// 测试是否实现了ChatModel接口
	var _ models.ChatModel = model
}

// TestChatWithInvalidInput 测试Chat方法的输入验证
func TestChatWithInvalidInput(t *testing.T) {
	model := NewModel("test-key", "claude-3-opus-20240229")
	ctx := context.Background()

	// 测试空messages
	_, err := model.Chat(ctx, []models.Message{})
	if err == nil {
		t.Error("Expected error for empty messages, got nil")
	}

	expectedErr := "no messages provided"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestContextTimeout 测试上下文超时
func TestContextTimeout(t *testing.T) {
	model := NewModel("invalid-key", "claude-3-opus-20240229")

	// 创建一个很短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// 等待确保超时
	time.Sleep(2 * time.Millisecond)

	_, err := model.Chat(ctx, []models.Message{
		models.NewUserMessage("test"),
	})
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

// BenchmarkNewModel 基准测试模型创建
func BenchmarkNewModel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewModel("test-key", "claude-3-opus-20240229",
			models.WithMaxTokens(100),
			models.WithTemperature(0.7),
		)
	}
}
