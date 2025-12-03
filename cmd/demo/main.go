package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golangchain/models"
	"golangchain/models/openai"
	"golangchain/utils"
)

func main() {
	// 命令行参数
	apiKey := flag.String("api-key", "", "OpenAI API密钥")
	modelName := flag.String("model", "gpt-3.5-turbo-instruct", "要使用的模型名称")
	promptsFile := flag.String("prompts", "", "包含提示的文件路径，每行一个提示")
	concurrent := flag.Bool("concurrent", true, "是否使用并发模式")
	concurrencyLevel := flag.Int("concurrency", 5, "并发级别")
	timeout := flag.Int("timeout", 30, "超时时间（秒）")
	verbose := flag.Bool("verbose", false, "是否输出详细日志")

	flag.Parse()

	// 检查API密钥
	if *apiKey == "" {
		*apiKey = os.Getenv("OPENAI_API_KEY")
		if *apiKey == "" {
			log.Fatal("必须提供OpenAI API密钥，通过-api-key参数或OPENAI_API_KEY环境变量")
		}
	}

	// 从文件加载提示或使用默认提示
	var prompts []string
	if *promptsFile != "" {
		data, err := os.ReadFile(*promptsFile)
		if err != nil {
			log.Fatalf("读取提示文件失败: %v", err)
		}
		prompts = strings.Split(string(data), "\n")
		// 过滤空行
		var filtered []string
		for _, p := range prompts {
			if strings.TrimSpace(p) != "" {
				filtered = append(filtered, p)
			}
		}
		prompts = filtered
	} else {
		// 默认提示
		prompts = []string{
			"解释量子计算的基本原理。",
			"列出5种常见的机器学习算法及其应用场景。",
			"用Go语言编写一个快速排序算法。",
			"解释P vs NP问题及其在计算机科学中的重要性。",
			"分析区块链技术的优势和局限性。",
		}
	}

	if len(prompts) == 0 {
		log.Fatal("没有提供有效的提示")
	}

	if *verbose {
		fmt.Printf("加载了%d个提示\n", len(prompts))
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	// 创建模型实例
	model := openai.NewModel(*apiKey, *modelName,
		models.WithMaxTokens(100),
		models.WithTemperature(0.7),
	)

	// 设置批处理选项
	batchOpts := utils.BatchProcessOptions{
		MaxConcurrent: *concurrencyLevel,
		Timeout:       time.Duration(*timeout) * time.Second,
	}

	fmt.Println("GoLangChain演示：展示Go的并发优势")
	fmt.Println("===================================")
	fmt.Printf("使用模型: %s\n", *modelName)
	fmt.Printf("提示数量: %d\n", len(prompts))
	fmt.Printf("并发级别: %d\n", *concurrencyLevel)
	fmt.Println("===================================")

	// 运行性能比较
	serialTime, parallelTime, speedup := utils.PerformanceComparison(ctx, model, prompts, models.WithMaxTokens(100))

	fmt.Println("\n性能比较结果:")
	fmt.Printf("串行处理时间: %v\n", serialTime)
	fmt.Printf("并行处理时间: %v\n", parallelTime)
	fmt.Printf("加速比: %.2f倍\n", speedup)
	fmt.Println("===================================")

	// 根据用户选择运行串行或并行处理并显示结果
	fmt.Println("\n提示处理结果:")

	var results []utils.BatchCompletionResult

	if *concurrent {
		fmt.Println("使用并行处理模式...")
		results = utils.BatchProcess(ctx, model, prompts, models.WithMaxTokens(100), batchOpts)
	} else {
		fmt.Println("使用串行处理模式...")
		results = make([]utils.BatchCompletionResult, len(prompts))
		for i, prompt := range prompts {
			completions, err := model.Generate(ctx, []string{prompt}, models.WithMaxTokens(100))
			if err != nil {
				results[i] = utils.BatchCompletionResult{Error: err, Index: i}
			} else if len(completions) > 0 {
				results[i] = utils.BatchCompletionResult{Completion: completions[0], Index: i}
			}
		}
	}

	// 显示结果
	for i, result := range results {
		fmt.Printf("\n提示 #%d: %s\n", i+1, truncateString(prompts[i], 50))
		if result.Error != nil {
			fmt.Printf("错误: %v\n", result.Error)
		} else {
			response := result.Completion.Text
			fmt.Printf("响应: %s\n", truncateString(response, 100))
			if *verbose {
				fmt.Printf("完整响应: %s\n", response)
				fmt.Printf("结束原因: %s\n", result.Completion.FinishReason)
				fmt.Printf("使用的令牌数: %d\n", result.Completion.TokensUsed)
			}
		}
	}
}

// truncateString 截断字符串并添加省略号
func truncateString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
