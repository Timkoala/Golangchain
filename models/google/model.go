// Package google 提供Google Gemini API的LLM接口实现
package google

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
	defaultChatURL = "https://generativelanguage.googleapis.com/v1beta/models"
	defaultTimeout = 30 * time.Second
)

// Model 实现Google Gemini API的ChatModel接口
type Model struct {
	apiKey   string
	modelID  string
	client   *http.Client
	defaults models.Options
}

var (
	_ models.ChatModel = (*Model)(nil)
)

// NewModel 创建新的Google Gemini模型实例
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

// geminiContent 是Gemini内容格式
type geminiContent struct {
	Role  string `json:"role"`
	Parts []struct {
		Text string `json:"text"`
	} `json:"parts"`
}

// geminiRequest 是Gemini请求体
type geminiRequest struct {
	Contents         []geminiContent `json:"contents"`
	GenerationConfig struct {
		MaxOutputTokens int      `json:"maxOutputTokens"`
		Temperature     float64  `json:"temperature,omitempty"`
		TopP            float64  `json:"topP,omitempty"`
		StopSequences   []string `json:"stopSequences,omitempty"`
	} `json:"generationConfig"`
}

// geminiResponse 是Gemini响应体
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// Chat 实现ChatModel接口
func (m *Model) Chat(ctx context.Context, messages []models.Message, options ...models.Option) (models.ChatResponse, error) {
	opts := m.defaults
	for _, opt := range options {
		opt(&opts)
	}

	if len(messages) == 0 {
		return models.ChatResponse{}, errors.New("no messages provided")
	}

	// 转换消息格式
	var contents []geminiContent
	for _, msg := range messages {
		role := "user"
		if msg.Role == models.RoleAssistant {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role: role,
			Parts: []struct {
				Text string `json:"text"`
			}{
				{Text: msg.Content},
			},
		})
	}

	// 准备请求
	req := geminiRequest{
		Contents: contents,
	}
	req.GenerationConfig.MaxOutputTokens = opts.MaxTokens
	if opts.Temperature > 0 {
		req.GenerationConfig.Temperature = opts.Temperature
	}
	if opts.TopP > 0 {
		req.GenerationConfig.TopP = opts.TopP
	}
	req.GenerationConfig.StopSequences = opts.Stop

	// 编码请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return models.ChatResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", defaultChatURL, m.modelID, m.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return models.ChatResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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
	var respBody geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return models.ChatResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// 处理结果
	if len(respBody.Candidates) == 0 {
		return models.ChatResponse{}, errors.New("no response candidates received")
	}

	candidate := respBody.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return models.ChatResponse{}, errors.New("no response content received")
	}

	return models.ChatResponse{
		Message: models.Message{
			Role:    models.RoleAssistant,
			Content: candidate.Content.Parts[0].Text,
		},
		TokensUsed:   respBody.UsageMetadata.TotalTokenCount,
		FinishReason: candidate.FinishReason,
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
