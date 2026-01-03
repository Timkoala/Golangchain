package models

import (
	"context"
)

// Role represents the role of a chat message.
type Role string

const (
	// RoleSystem represents a system message.
	RoleSystem Role = "system"
	// RoleUser represents a user message.
	RoleUser Role = "user"
	// RoleAssistant represents an assistant message.
	RoleAssistant Role = "assistant"
	// RoleFunction represents a function call result.
	RoleFunction Role = "function"
)

// Message represents a single message in a chat conversation.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ChatResponse represents a chat model response.
type ChatResponse struct {
	Message      Message
	TokensUsed   int
	FinishReason string
}

// ChatChunk represents a chunk of a streaming chat response.
type ChatChunk struct {
	Delta        Message
	Done         bool
	FinishReason string
}

// ChatModel defines the interface for chat models.
type ChatModel interface {
	// Chat sends messages and receives a response.
	Chat(ctx context.Context, messages []Message, options ...Option) (ChatResponse, error)

	// ChatStream sends messages and receives a streaming response.
	ChatStream(ctx context.Context, messages []Message, options ...Option) (<-chan ChatChunk, error)
}

// NewSystemMessage creates a system message.
func NewSystemMessage(content string) Message {
	return Message{
		Role:    RoleSystem,
		Content: content,
	}
}

// NewUserMessage creates a user message.
func NewUserMessage(content string) Message {
	return Message{
		Role:    RoleUser,
		Content: content,
	}
}

// NewAssistantMessage creates an assistant message.
func NewAssistantMessage(content string) Message {
	return Message{
		Role:    RoleAssistant,
		Content: content,
	}
}

// NewFunctionMessage creates a function call result message.
func NewFunctionMessage(content, name string) Message {
	return Message{
		Role:    RoleFunction,
		Content: content,
		Name:    name,
	}
}
