// Package utils provides utility functions for concurrent processing of LLM requests.
package utils

import (
	"context"
	"sync"
	"time"

	"golangchain/models"
)

// BatchProcessOptions configures batch processing operations.
type BatchProcessOptions struct {
	// MaxConcurrent controls maximum concurrent requests.
	MaxConcurrent int
	// Timeout sets the timeout for the entire batch operation.
	Timeout time.Duration
	// MaxRetries specifies the maximum number of retry attempts for failed requests.
	MaxRetries int
	// RetryDelay specifies the delay between retry attempts.
	RetryDelay time.Duration
}

// DefaultBatchOptions returns default batch processing options.
func DefaultBatchOptions() BatchProcessOptions {
	return BatchProcessOptions{
		MaxConcurrent: 5,
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryDelay:    time.Second,
	}
}

// BatchCompletionResult contains the result and error of a batch operation.
type BatchCompletionResult struct {
	Completion models.Completion
	Error      error
	Index      int
	Retries    int
}

// BatchProcess processes multiple LLM requests concurrently with retry support.
func BatchProcess(
	ctx context.Context,
	llm models.LLM,
	prompts []string,
	options models.Option,
	batchOptions ...BatchProcessOptions,
) []BatchCompletionResult {
	// Use default options
	opts := DefaultBatchOptions()
	if len(batchOptions) > 0 {
		opts = batchOptions[0]
	}

	// Create context with timeout
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Prepare results slice
	results := make([]BatchCompletionResult, len(prompts))

	// Return early if no prompts
	if len(prompts) == 0 {
		return results
	}

	// Create semaphore for concurrency control
	var sem chan struct{}
	if opts.MaxConcurrent > 0 {
		sem = make(chan struct{}, opts.MaxConcurrent)
	}

	// Use WaitGroup to wait for all goroutines
	var wg sync.WaitGroup

	// Launch all requests
	for i, prompt := range prompts {
		wg.Add(1)

		// Capture variables for goroutine
		i, prompt := i, prompt

		go func() {
			defer wg.Done()

			// Acquire semaphore if concurrency limit is set
			if sem != nil {
				select {
				case sem <- struct{}{}:
					defer func() { <-sem }()
				case <-ctx.Done():
					results[i] = BatchCompletionResult{
						Error: ctx.Err(),
						Index: i,
					}
					return
				}
			}

			// Execute with retry logic
			result := executeWithRetry(ctx, llm, prompt, options, opts.MaxRetries, opts.RetryDelay)
			result.Index = i
			results[i] = result
		}()
	}

	// Wait for all requests to complete
	wg.Wait()
	return results
}

// executeWithRetry executes a single LLM request with retry support.
func executeWithRetry(
	ctx context.Context,
	llm models.LLM,
	prompt string,
	options models.Option,
	maxRetries int,
	retryDelay time.Duration,
) BatchCompletionResult {
	var lastErr error
	retries := 0

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context before each attempt
		select {
		case <-ctx.Done():
			return BatchCompletionResult{
				Error:   ctx.Err(),
				Retries: retries,
			}
		default:
		}

		// Execute LLM request
		completions, err := llm.Generate(ctx, []string{prompt}, options)
		if err == nil && len(completions) > 0 {
			return BatchCompletionResult{
				Completion: completions[0],
				Retries:    retries,
			}
		}

		lastErr = err
		retries = attempt

		// Don't sleep after the last attempt
		if attempt < maxRetries {
			// Exponential backoff
			delay := retryDelay * time.Duration(1<<uint(attempt))
			select {
			case <-ctx.Done():
				return BatchCompletionResult{
					Error:   ctx.Err(),
					Retries: retries,
				}
			case <-time.After(delay):
			}
		}
	}

	return BatchCompletionResult{
		Error:   lastErr,
		Retries: retries,
	}
}

// PerformanceComparison compares serial and parallel processing performance.
func PerformanceComparison(
	ctx context.Context,
	llm models.LLM,
	prompts []string,
	options models.Option,
) (serialTime, parallelTime time.Duration, speedup float64) {
	// Serial processing
	serialStart := time.Now()
	for _, prompt := range prompts {
		_, _ = llm.Generate(ctx, []string{prompt}, options)
	}
	serialTime = time.Since(serialStart)

	// Parallel processing
	parallelStart := time.Now()
	BatchProcess(ctx, llm, prompts, options)
	parallelTime = time.Since(parallelStart)

	// Calculate speedup
	if parallelTime > 0 {
		speedup = float64(serialTime) / float64(parallelTime)
	}

	return serialTime, parallelTime, speedup
}
