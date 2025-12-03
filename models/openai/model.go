package openai

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
	openaiCompletionURL = "https://api.openai.com/v1/completions"
	openaiChatURL       = "https://api.openai.com/v1/chat/completions"
	defaultTimeout      = 30 * time.Second
)

// Model 实现OpenAI API的模型接口
type Model struct {
	apiKey   string
	modelID  string
	client   *http.Client
	defaults models.Options
}

// NewModel 创建新的OpenAI模型实例
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

// completionRequest 是OpenAI完成接口的请求体
type completionRequest struct {
	Model            string   `json:"model"`
	Prompt           string   `json:"prompt"`
	MaxTokens        int      `json:"max_tokens"`
	Temperature      float64  `json:"temperature"`
	TopP             float64  `json:"top_p"`
	FrequencyPenalty float64  `json:"frequency_penalty"`
	PresencePenalty  float64  `json:"presence_penalty"`
	Stop             []string `json:"stop,omitempty"`
	Stream           bool     `json:"stream,omitempty"`
}

// completionResponse 是OpenAI完成接口的响应体
type completionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string      `json:"text"`
		Index        int         `json:"index"`
		LogProbs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Generate 实现LLM接口，调用OpenAI API生成文本
func (m *Model) Generate(ctx context.Context, prompts []string, options ...models.Option) ([]models.Completion, error) {
	opts := m.defaults
	for _, opt := range options {
		opt(&opts)
	}

	// 这里简化，只处理第一个提示
	if len(prompts) == 0 {
		return nil, errors.New("no prompts provided")
	}

	// 准备请求
	req := completionRequest{
		Model:            m.modelID,
		Prompt:           prompts[0],
		MaxTokens:        opts.MaxTokens,
		Temperature:      opts.Temperature,
		TopP:             opts.TopP,
		FrequencyPenalty: opts.FrequencyPenalty,
		PresencePenalty:  opts.PresencePenalty,
		Stop:             opts.Stop,
	}

	// 编码请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", openaiCompletionURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)

	// 发送请求
	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	// 解析响应
	var respBody completionResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 处理结果
	completions := make([]models.Completion, len(respBody.Choices))
	for i, choice := range respBody.Choices {
		completions[i] = models.Completion{
			Text:         choice.Text,
			FinishReason: choice.FinishReason,
			TokensUsed:   respBody.Usage.TotalTokens,
		}
	}

	return completions, nil
}

// GenerateStream 实现流式生成接口
func (m *Model) GenerateStream(ctx context.Context, prompt string, options ...models.Option) (<-chan models.CompletionChunk, error) {
	// 这里仅做接口实现，简化处理，实际需要支持流式传输
	resultChan := make(chan models.CompletionChunk, 1)

	go func() {
		defer close(resultChan)

		// 实际实现中应该使用SSE流式解析
		completions, err := m.Generate(ctx, []string{prompt}, options...)
		if err != nil {
			resultChan <- models.CompletionChunk{
				Text:         "",
				FinishReason: "error",
				Done:         true,
			}
			return
		}

		if len(completions) > 0 {
			resultChan <- models.CompletionChunk{
				Text:         completions[0].Text,
				FinishReason: completions[0].FinishReason,
				Done:         true,
			}
		}
	}()

	return resultChan, nil
}
