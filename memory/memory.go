// Package memory 提供对话记忆系统
package memory

import (
	"errors"
	"sync"

	"golangchain/models"
)

// Memory 定义记忆接口
type Memory interface {
	Add(message models.Message) error
	Get() []models.Message
	Clear() error
	GetSummary() string
}

// BufferMemory 定义缓冲记忆
type BufferMemory struct {
	messages []models.Message
	maxSize  int
	mu       sync.RWMutex
}

// NewBufferMemory 创建新的缓冲记忆
func NewBufferMemory(maxSize int) *BufferMemory {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &BufferMemory{
		messages: make([]models.Message, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add 添加消息到记忆
func (bm *BufferMemory) Add(message models.Message) error {
	if message.Content == "" {
		return errors.New("message content cannot be empty")
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.messages = append(bm.messages, message)

	// 如果超过最大大小，删除最早的消息
	if len(bm.messages) > bm.maxSize {
		bm.messages = bm.messages[1:]
	}

	return nil
}

// Get 获取所有消息
func (bm *BufferMemory) Get() []models.Message {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	// 返回副本以避免外部修改
	messages := make([]models.Message, len(bm.messages))
	copy(messages, bm.messages)
	return messages
}

// Clear 清空记忆
func (bm *BufferMemory) Clear() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.messages = make([]models.Message, 0, bm.maxSize)
	return nil
}

// GetSummary 获取记忆摘要
func (bm *BufferMemory) GetSummary() string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if len(bm.messages) == 0 {
		return "No messages in memory"
	}

	summary := "Memory Summary:\n"
	for i, msg := range bm.messages {
		summary += "- " + string(msg.Role) + ": " + msg.Content + "\n"
		if i >= 4 { // 只显示最后5条消息
			break
		}
	}
	return summary
}

// SummaryMemory 定义摘要记忆
type SummaryMemory struct {
	messages []models.Message
	summary  string
	maxSize  int
	mu       sync.RWMutex
}

// NewSummaryMemory 创建新的摘要记忆
func NewSummaryMemory(maxSize int) *SummaryMemory {
	if maxSize <= 0 {
		maxSize = 50
	}
	return &SummaryMemory{
		messages: make([]models.Message, 0, maxSize),
		maxSize:  maxSize,
		summary:  "",
	}
}

// Add 添加消息到摘要记忆
func (sm *SummaryMemory) Add(message models.Message) error {
	if message.Content == "" {
		return errors.New("message content cannot be empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.messages = append(sm.messages, message)

	// 如果超过最大大小，更新摘要并清空消息
	if len(sm.messages) > sm.maxSize {
		sm.updateSummary()
		sm.messages = make([]models.Message, 0, sm.maxSize)
	}

	return nil
}

// Get 获取所有消息
func (sm *SummaryMemory) Get() []models.Message {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	messages := make([]models.Message, len(sm.messages))
	copy(messages, sm.messages)
	return messages
}

// Clear 清空记忆
func (sm *SummaryMemory) Clear() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.messages = make([]models.Message, 0, sm.maxSize)
	sm.summary = ""
	return nil
}

// GetSummary 获取摘要
func (sm *SummaryMemory) GetSummary() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.summary
}

// updateSummary 更新摘要
func (sm *SummaryMemory) updateSummary() {
	if len(sm.messages) == 0 {
		return
	}

	summary := "Previous conversation summary:\n"
	for _, msg := range sm.messages {
		summary += "- " + string(msg.Role) + ": " + msg.Content + "\n"
	}
	sm.summary = summary
}

// ConversationMemory 定义对话记忆
type ConversationMemory struct {
	buffer  *BufferMemory
	summary *SummaryMemory
	mu      sync.RWMutex
}

// NewConversationMemory 创建新的对话记忆
func NewConversationMemory(bufferSize, summarySize int) *ConversationMemory {
	return &ConversationMemory{
		buffer:  NewBufferMemory(bufferSize),
		summary: NewSummaryMemory(summarySize),
	}
}

// Add 添加消息
func (cm *ConversationMemory) Add(message models.Message) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := cm.buffer.Add(message); err != nil {
		return err
	}
	return cm.summary.Add(message)
}

// Get 获取所有消息
func (cm *ConversationMemory) Get() []models.Message {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.buffer.Get()
}

// Clear 清空记忆
func (cm *ConversationMemory) Clear() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := cm.buffer.Clear(); err != nil {
		return err
	}
	return cm.summary.Clear()
}

// GetSummary 获取摘要
func (cm *ConversationMemory) GetSummary() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.summary.GetSummary()
}

// GetBufferMessages 获取缓冲消息
func (cm *ConversationMemory) GetBufferMessages() []models.Message {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.buffer.Get()
}

// GetMessageCount 获取消息数量
func (cm *ConversationMemory) GetMessageCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.buffer.Get())
}
