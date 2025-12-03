package utils

import (
	"context"
	"sync"
	"time"

	"golangchain/models"
)

// BatchProcessOptions 配置批处理操作的选项
type BatchProcessOptions struct {
	// MaxConcurrent 控制最大并发数
	MaxConcurrent int
	// Timeout 为整个批处理操作设置超时
	Timeout time.Duration
}

// DefaultBatchOptions 返回默认的批处理选项
func DefaultBatchOptions() BatchProcessOptions {
	return BatchProcessOptions{
		MaxConcurrent: 5,
		Timeout:       30 * time.Second,
	}
}

// BatchCompletionResult 包含批处理的结果和错误
type BatchCompletionResult struct {
	Completion models.Completion
	Error      error
	Index      int
}

// BatchProcess 并行处理多个LLM请求
func BatchProcess(
	ctx context.Context,
	llm models.LLM,
	prompts []string,
	options models.Option,
	batchOptions ...BatchProcessOptions,
) []BatchCompletionResult {
	// 使用默认选项
	opts := DefaultBatchOptions()
	if len(batchOptions) > 0 {
		opts = batchOptions[0]
	}

	// 创建带有超时的上下文
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// 准备结果切片
	results := make([]BatchCompletionResult, len(prompts))

	// 如果没有提示，直接返回
	if len(prompts) == 0 {
		return results
	}

	// 创建信号量控制并发
	var sem chan struct{}
	if opts.MaxConcurrent > 0 {
		sem = make(chan struct{}, opts.MaxConcurrent)
	}

	// 使用WaitGroup等待所有goroutine完成
	var wg sync.WaitGroup

	// 启动所有请求
	for i, prompt := range prompts {
		wg.Add(1)

		// 封装参数，以便在goroutine中使用
		i, prompt := i, prompt

		go func() {
			defer wg.Done()

			// 如果设置了最大并发数，获取信号量
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

			// 执行LLM请求
			completions, err := llm.Generate(ctx, []string{prompt}, options)

			// 存储结果
			if err != nil {
				results[i] = BatchCompletionResult{
					Error: err,
					Index: i,
				}
				return
			}

			if len(completions) > 0 {
				results[i] = BatchCompletionResult{
					Completion: completions[0],
					Index:      i,
				}
			} else {
				results[i] = BatchCompletionResult{
					Error: nil, // 没有错误但也没有结果
					Index: i,
				}
			}
		}()
	}

	// 等待所有请求完成
	wg.Wait()
	return results
}

// PerformanceComparison 比较串行和并行处理的性能差异
func PerformanceComparison(
	ctx context.Context,
	llm models.LLM,
	prompts []string,
	options models.Option,
) (serialTime, parallelTime time.Duration, speedup float64) {
	// 串行处理
	serialStart := time.Now()
	for _, prompt := range prompts {
		_, _ = llm.Generate(ctx, []string{prompt}, options)
	}
	serialTime = time.Since(serialStart)

	// 并行处理
	parallelStart := time.Now()
	BatchProcess(ctx, llm, prompts, options)
	parallelTime = time.Since(parallelStart)

	// 计算加速比
	if serialTime > 0 {
		speedup = float64(serialTime) / float64(parallelTime)
	}

	return serialTime, parallelTime, speedup
}
