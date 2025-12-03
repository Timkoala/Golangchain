package models

import (
	"context"
)

// Role 表示聊天消息的角色
type Role string

const (
	// RoleSystem 表示系统消息
	RoleSystem Role = "system"
	// RoleUser 表示用户消息
	RoleUser Role = "user"
	// RoleAssistant 表示助手消息
	RoleAssistant Role = "assistant"
	// RoleFuction 表示函数调用结果
	RoleFunction Role = "function"
)

// Message 表示聊天对话中的单条消息
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ChatResponse 表示聊天模型的响应
type ChatResponse struct {
	Message      Message
	TokensUsed   int
	FinishReason string
}

// ChatChunk 表示流式聊天响应的片段
type ChatChunk struct {
	Delta       Message
	Done        bool
	FinishReason string
}

// ChatModel 定义聊天模型接口
type ChatModel interface {
	// Chat 发送消息并获取回复
	Chat(ctx context.Context, messages []Message, options ...Option) (ChatResponse, error)

	// ChatStream 发送消息并获取流式回复
	ChatStream(ctx context.Context, messages []Message, options ...Option) (<-chan ChatChunk, error)
}

// NewSystemMessage 创建一条系统消息
func NewSystemMessage(content string) Message {
	return Message{
		Role:    RoleSystem,
		Content: content,
	}
}

// NewUserMessage 创建一条用户消息
func NewUserMessage(content string) Message {
	return Message{
		Role:    RoleUser,
		Content: content,
	}
}

// NewAssistantMessage 创建一条助手消息
func NewAssistantMessage(content string) Message {
	return Message{
		Role:    RoleAssistant,
		Content: content,
	}
}

// NewFunctionMessage 创建一条函数调用结果消息
func NewFunctionMessage(content, name string) Message {
	return Message{
		Role:    RoleFunction,
		Content: content,
		Name:    name,
	}
}
