// Package models defines core interfaces and types for LLM interactions.
package models

import (
	"context"
)

// Option defines a function type for configuring model options.
type Option func(*Options)

// Options contains all model invocation options.
type Options struct {
	MaxTokens        int
	Temperature      float64
	TopP             float64
	FrequencyPenalty float64
	PresencePenalty  float64
	Stop             []string
	Timeout          int
}

// Completion represents an LLM completion response.
type Completion struct {
	Text         string
	FinishReason string
	TokensUsed   int
}

// CompletionChunk represents a chunk of a streaming response.
type CompletionChunk struct {
	Text         string
	FinishReason string
	Done         bool
}

// LLM defines the base interface for large language models.
type LLM interface {
	// Generate produces completions for the given prompts.
	Generate(ctx context.Context, prompts []string, options ...Option) ([]Completion, error)

	// GenerateStream produces a streaming completion for a single prompt.
	GenerateStream(ctx context.Context, prompt string, options ...Option) (<-chan CompletionChunk, error)
}

// WithMaxTokens sets the maximum number of tokens to generate.
func WithMaxTokens(n int) Option {
	return func(o *Options) {
		o.MaxTokens = n
	}
}

// WithTemperature sets the sampling temperature.
func WithTemperature(t float64) Option {
	return func(o *Options) {
		o.Temperature = t
	}
}

// WithTopP sets the top-p sampling parameter.
func WithTopP(p float64) Option {
	return func(o *Options) {
		o.TopP = p
	}
}

// WithStop sets the stop sequences.
func WithStop(stop []string) Option {
	return func(o *Options) {
		o.Stop = stop
	}
}

// DefaultOptions returns the default options.
func DefaultOptions() Options {
	return Options{
		MaxTokens:   100,
		Temperature: 0.7,
		TopP:        1.0,
	}
}
