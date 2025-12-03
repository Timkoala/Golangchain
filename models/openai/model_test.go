package openai

import (
	"context"
	"testing"
	"time"

	"golangchain/models"
)

// TestNewModel 测试模型创建
func TestNewModel(t *testing.T) {
	model := NewModel("test-api-key", "gpt-3.5-turbo",
		models.WithMaxTokens(200),
		models.WithTemperature(0.8),
	)

	if model.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got %s", model.apiKey)
	}

	if model.modelID != "gpt-3.5-turbo" {
		t.Errorf("Expected modelID to be 'gpt-3.5-turbo', got %s", model.modelID)
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
	model := NewModel("test-key", "gpt-3.5-turbo")

	// 测试是否实现了LLM接口
	var _ models.LLM = model

	// 测试是否实现了ChatModel接口
	var _ models.ChatModel = model
}

// TestGenerateWithInvalidInput 测试Generate方法的输入验证
func TestGenerateWithInvalidInput(t *testing.T) {
	model := NewModel("test-key", "gpt-3.5-turbo")
	ctx := context.Background()

	// 测试空prompts
	_, err := model.Generate(ctx, []string{})
	if err == nil {
		t.Error("Expected error for empty prompts, got nil")
	}

	expectedErr := "no prompts provided"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// TestChatWithInvalidInput 测试Chat方法的输入验证
func TestChatWithInvalidInput(t *testing.T) {
	model := NewModel("test-key", "gpt-3.5-turbo")
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
	model := NewModel("invalid-key", "gpt-3.5-turbo")

	// 创建一个很短的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// 等待确保超时
	time.Sleep(2 * time.Millisecond)

	_, err := model.Generate(ctx, []string{"test prompt"})
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

// BenchmarkNewModel 基准测试模型创建
func BenchmarkNewModel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewModel("test-key", "gpt-3.5-turbo",
			models.WithMaxTokens(100),
			models.WithTemperature(0.7),
		)
	}
}
