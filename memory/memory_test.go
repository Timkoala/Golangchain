package memory

import (
	"testing"

	"golangchain/models"
)

// TestNewBufferMemory 测试缓冲记忆创建
func TestNewBufferMemory(t *testing.T) {
	mem := NewBufferMemory(10)
	if mem == nil {
		t.Error("Expected non-nil memory")
	}
	if mem.maxSize != 10 {
		t.Errorf("Expected maxSize to be 10, got %d", mem.maxSize)
	}
}

// TestBufferMemoryAdd 测试添加消息
func TestBufferMemoryAdd(t *testing.T) {
	mem := NewBufferMemory(10)
	msg := models.NewUserMessage("Hello")

	err := mem.Add(msg)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	messages := mem.Get()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

// TestBufferMemoryAddEmpty 测试添加空消息
func TestBufferMemoryAddEmpty(t *testing.T) {
	mem := NewBufferMemory(10)
	msg := models.Message{Role: models.RoleUser, Content: ""}

	err := mem.Add(msg)
	if err == nil {
		t.Error("Expected error for empty content")
	}
}

// TestBufferMemoryMaxSize 测试最大大小限制
func TestBufferMemoryMaxSize(t *testing.T) {
	mem := NewBufferMemory(3)

	for i := 0; i < 5; i++ {
		mem.Add(models.NewUserMessage("Message"))
	}

	messages := mem.Get()
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}
}

// TestBufferMemoryClear 测试清空记忆
func TestBufferMemoryClear(t *testing.T) {
	mem := NewBufferMemory(10)
	mem.Add(models.NewUserMessage("Hello"))
	mem.Add(models.NewUserMessage("World"))

	err := mem.Clear()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	messages := mem.Get()
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(messages))
	}
}

// TestBufferMemoryGetSummary 测试获取摘要
func TestBufferMemoryGetSummary(t *testing.T) {
	mem := NewBufferMemory(10)
	mem.Add(models.NewUserMessage("Hello"))

	summary := mem.GetSummary()
	if summary == "" {
		t.Error("Expected non-empty summary")
	}
}

// TestSummaryMemory 测试摘要记忆
func TestSummaryMemory(t *testing.T) {
	mem := NewSummaryMemory(5)
	if mem == nil {
		t.Error("Expected non-nil memory")
	}

	for i := 0; i < 3; i++ {
		mem.Add(models.NewUserMessage("Message"))
	}

	messages := mem.Get()
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}
}

// TestConversationMemory 测试对话记忆
func TestConversationMemory(t *testing.T) {
	mem := NewConversationMemory(10, 5)
	if mem == nil {
		t.Error("Expected non-nil memory")
	}

	mem.Add(models.NewUserMessage("Hello"))
	mem.Add(models.NewAssistantMessage("Hi there!"))

	count := mem.GetMessageCount()
	if count != 2 {
		t.Errorf("Expected 2 messages, got %d", count)
	}
}

// TestConversationMemoryClear 测试清空对话记忆
func TestConversationMemoryClear(t *testing.T) {
	mem := NewConversationMemory(10, 5)
	mem.Add(models.NewUserMessage("Hello"))

	err := mem.Clear()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	count := mem.GetMessageCount()
	if count != 0 {
		t.Errorf("Expected 0 messages, got %d", count)
	}
}

// TestMemoryInterface 测试接口实现
func TestMemoryInterface(t *testing.T) {
	var _ Memory = NewBufferMemory(10)
	var _ Memory = NewSummaryMemory(10)
	var _ Memory = NewConversationMemory(10, 5)
}

// BenchmarkBufferMemoryAdd 基准测试添加
func BenchmarkBufferMemoryAdd(b *testing.B) {
	mem := NewBufferMemory(1000)
	msg := models.NewUserMessage("Test message")

	for i := 0; i < b.N; i++ {
		mem.Add(msg)
	}
}
