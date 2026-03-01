// Package anthropic 提供Anthropic Claude API的LLM接口实现
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golangchain/models"
)

const (
	defaultChatURL = "https://api.anthropic.com/v1/messages"
	defaultTimeout = 30 * time.Second
	apiVersion     = "2024-01-15"
)

// Model 实现Anthropic Claude API的ChatModel接口
type Model struct {
	apiKey   string
	modelID  string
	client   *http.Client
	defaults models.Options
}

// 编译时接口合规性检查
var (
	_ models.ChatModel = (*Model)(nil)
)

// NewModel 创建新的Anthropic Claude模型实例
func NewModel(apiKey, modelID string, options ...models.Option) *Model {
	defaults := models.DefaultOptions()
	for _, opt := range options {
		opt(&defaults)
	}

	return &Model{
		apiKey:   apiKey,
		modelID:  modelID,
		client:   &http.Client{Timeout: defaultTimeout},
		defaults: defaults,
	}
}

// messageRequest 是Anthropic消息接口的请求体
type messageRequest struct {
	Model       string         `json:"model"`
	Messages    []anthropicMsg `json:"messages"`
	MaxTokens   int            `json:"max_tokens"`
	Temperature float64        `json:"temperature,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
	StopSeqs    []string       `json:"stop_sequences,omitempty"`
	System      string         `json:"system,omitempty"`
}

// anthropicMsg 是Anthropic消息格式
type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// messageResponse 是Anthropic消息接口的响应体
type messageResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	StopSeq    string `json:"stop_sequence"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Chat 实现ChatModel接口，进行聊天对话
func (m *Model) Chat(ctx context.Context, messages []models.Message, options ...models.Option) (models.ChatResponse, error) {
	opts := m.defaults
	for _, opt := range options {
		opt(&opts)
	}

	if len(messages) == 0 {
		return models.ChatResponse{}, errors.New("no messages provided")
	}

	// 分离系统消息和其他消息
	var systemMsg string
	var chatMessages []anthropicMsg
	for _, msg := range messages {
		if msg.Role == models.RoleSystem {
			systemMsg = msg.Content
		} else {
			role := string(msg.Role)
			if role == "assistant" {
				role = "assistant"
			} else {
				role = "user"
			}
			chatMessages = append(chatMessages, anthropicMsg{
				Role:    role,
				Content: msg.Content,
			})
		}
	}

	// 准备请求
	req := messageRequest{
		Model:     m.modelID,
		Messages:  chatMessages,
		MaxTokens: opts.MaxTokens,
		System:    systemMsg,
		StopSeqs:  opts.Stop,
	}

	if opts.Temperature > 0 {
		req.Temperature = opts.Temperature
	}
	if opts.TopP > 0 {
		req.TopP = opts.TopP
	}

	// 编码请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return models.ChatResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultChatURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return models.ChatResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", m.apiKey)
	httpReq.Header.Set("anthropic-version", apiVersion)

	// 发送请求
	resp, err := m.client.Do(httpReq)
	if err != nil {
		return models.ChatResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.ChatResponse{}, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	// 解析响应
	var respBody messageResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return models.ChatResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// 处理结果
	if len(respBody.Content) == 0 {
		return models.ChatResponse{}, errors.New("no response content received")
	}

	return models.ChatResponse{
		Message: models.Message{
			Role:    models.RoleAssistant,
			Content: respBody.Content[0].Text,
		},
		TokensUsed:   respBody.Usage.InputTokens + respBody.Usage.OutputTokens,
		FinishReason: respBody.StopReason,
	}, nil
}

// ChatStream 实现流式聊天接口
func (m *Model) ChatStream(ctx context.Context, messages []models.Message, options ...models.Option) (<-chan models.ChatChunk, error) {
	resultChan := make(chan models.ChatChunk, 1)

	go func() {
		defer close(resultChan)

		response, err := m.Chat(ctx, messages, options...)
		if err != nil {
			resultChan <- models.ChatChunk{
				Delta:        models.Message{},
				Done:         true,
				FinishReason: "error",
			}
			return
		}

		resultChan <- models.ChatChunk{
			Delta:        response.Message,
			Done:         true,
			FinishReason: response.FinishReason,
		}
	}()

	return resultChan, nil
}
