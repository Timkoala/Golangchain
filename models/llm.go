package models

import (
	"context"
)

// Option 定义模型调用的选项
type Option func(*Options)

// Options 包含所有模型调用选项
type Options struct {
	MaxTokens        int
	Temperature      float64
	TopP             float64
	FrequencyPenalty float64
	PresencePenalty  float64
	Stop             []string
	Timeout          int
}

// Completion 代表LLM的完成响应
type Completion struct {
	Text         string
	FinishReason string
	TokensUsed   int
}

// CompletionChunk 代表流式响应的一个片段
type CompletionChunk struct {
	Text         string
	FinishReason string
	Done         bool
}

// LLM 定义基础大语言模型接口
type LLM interface {
	// Generate 根据提示生成完成内容
	Generate(ctx context.Context, prompts []string, options ...Option) ([]Completion, error)

	// GenerateStream 流式生成完成内容
	GenerateStream(ctx context.Context, prompt string, options ...Option) (<-chan CompletionChunk, error)
}

// WithMaxTokens 设置最大生成标记数
func WithMaxTokens(n int) Option {
	return func(o *Options) {
		o.MaxTokens = n
	}
}

// WithTemperature 设置生成温度
func WithTemperature(t float64) Option {
	return func(o *Options) {
		o.Temperature = t
	}
}

// WithTopP 设置Top-P采样参数
func WithTopP(p float64) Option {
	return func(o *Options) {
		o.TopP = p
	}
}

// WithStop 设置停止序列
func WithStop(stop []string) Option {
	return func(o *Options) {
		o.Stop = stop
	}
}

// DefaultOptions 返回默认选项
func DefaultOptions() Options {
	return Options{
		MaxTokens:   100,
		Temperature: 0.7,
		TopP:        1.0,
	}
}
