// Package openai provides an OpenAI API implementation of the LLM interfaces.
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
	defaultCompletionURL = "https://api.openai.com/v1/completions"
	defaultChatURL       = "https://api.openai.com/v1/chat/completions"
	defaultTimeout       = 30 * time.Second
)

// Model implements the LLM and ChatModel interfaces for OpenAI API.
type Model struct {
	apiKey   string
	modelID  string
	client   *http.Client
	defaults models.Options
}

// Compile-time interface compliance check.
var (
	_ models.LLM       = (*Model)(nil)
	_ models.ChatModel = (*Model)(nil)
)

// NewModel creates a new OpenAI model instance.
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

// completionRequest represents the OpenAI completion API request body.
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

// completionResponse represents the OpenAI completion API response body.
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

// Generate implements the LLM interface.
func (m *Model) Generate(ctx context.Context, prompts []string, options ...models.Option) ([]models.Completion, error) {
	opts := m.defaults
	for _, opt := range options {
		opt(&opts)
	}

	if len(prompts) == 0 {
		return nil, errors.New("no prompts provided")
	}

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

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultCompletionURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var respBody completionResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

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

// GenerateStream implements streaming generation.
func (m *Model) GenerateStream(ctx context.Context, prompt string, options ...models.Option) (<-chan models.CompletionChunk, error) {
	resultChan := make(chan models.CompletionChunk, 1)

	go func() {
		defer close(resultChan)

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

// chatRequest represents the OpenAI chat API request body.
type chatRequest struct {
	Model            string           `json:"model"`
	Messages         []models.Message `json:"messages"`
	MaxTokens        int              `json:"max_tokens"`
	Temperature      float64          `json:"temperature"`
	TopP             float64          `json:"top_p"`
	FrequencyPenalty float64          `json:"frequency_penalty"`
	PresencePenalty  float64          `json:"presence_penalty"`
	Stop             []string         `json:"stop,omitempty"`
	Stream           bool             `json:"stream,omitempty"`
}

// chatResponse represents the OpenAI chat API response body.
type chatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int            `json:"index"`
		Message      models.Message `json:"message"`
		FinishReason string         `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Chat implements the ChatModel interface.
func (m *Model) Chat(ctx context.Context, messages []models.Message, options ...models.Option) (models.ChatResponse, error) {
	opts := m.defaults
	for _, opt := range options {
		opt(&opts)
	}

	if len(messages) == 0 {
		return models.ChatResponse{}, errors.New("no messages provided")
	}

	req := chatRequest{
		Model:            m.modelID,
		Messages:         messages,
		MaxTokens:        opts.MaxTokens,
		Temperature:      opts.Temperature,
		TopP:             opts.TopP,
		FrequencyPenalty: opts.FrequencyPenalty,
		PresencePenalty:  opts.PresencePenalty,
		Stop:             opts.Stop,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return models.ChatResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultChatURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return models.ChatResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return models.ChatResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.ChatResponse{}, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var respBody chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return models.ChatResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(respBody.Choices) == 0 {
		return models.ChatResponse{}, errors.New("no response choices received")
	}

	choice := respBody.Choices[0]
	return models.ChatResponse{
		Message:      choice.Message,
		TokensUsed:   respBody.Usage.TotalTokens,
		FinishReason: choice.FinishReason,
	}, nil
}

// ChatStream implements streaming chat.
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
