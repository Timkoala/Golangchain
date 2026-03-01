# GoLangChain

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

A lightweight, high-performance LLM application framework written in Go, inspired by [LangChain](https://github.com/langchain-ai/langchain).

[中文文档](README_CN.md)

## Why GoLangChain?

While Python LangChain is powerful, it comes with significant overhead in production environments. GoLangChain addresses these pain points:

| Aspect | Python LangChain | GoLangChain |
|--------|------------------|-------------|
| Memory Footprint | ~200MB+ | ~80MB (60% reduction) |
| Cold Start Time | 2-3 seconds | < 200ms |
| Concurrent Requests | Requires asyncio/threading | Native goroutines |
| Dependencies | Complex dependency tree | Zero external dependencies |
| Deployment | Virtual environment required | Single binary |

### Key Advantages

- **Native Concurrency**: Built on goroutines and channels, achieving 4x+ throughput improvement for parallel LLM requests without external async frameworks
- **Zero Dependencies**: Uses only Go standard library, eliminating supply chain risks and simplifying deployment
- **Production Ready**: Single binary compilation, cross-platform support, minimal resource consumption
- **Type Safety**: Compile-time error detection through Go's static type system

## Architecture

```
golangchain/
├── models/                 # Model abstraction layer
│   ├── llm.go             # Base LLM interface
│   ├── chat.go            # Chat model interface
│   ├── openai/            # OpenAI provider implementation
│   │   └── model.go
│   ├── anthropic/         # Anthropic Claude provider
│   │   └── model.go
│   └── google/            # Google Gemini provider
│       └── model.go
├── agent/                  # Agent framework
│   └── agent.go           # Agent core with tool support
├── prompt/                 # Prompt template system
│   └── template.go        # Template engine and builder
├── memory/                 # Memory system
│   └── memory.go          # Buffer, summary, and conversation memory
├── rag/                    # RAG retrieval augmented generation
│   └── retriever.go       # Simple and vector retrievers
├── utils/                  # Utilities
│   └── concurrency.go     # BatchProcess engine with retry logic
└── go.mod                 # Zero external dependencies
```

## Core Components

### LLM Interface

The framework defines a provider-agnostic interface for LLM interactions:

```go
type LLM interface {
    Generate(ctx context.Context, prompts []string, options ...Option) ([]Completion, error)
    GenerateStream(ctx context.Context, prompt string, options ...Option) (<-chan CompletionChunk, error)
}
```

### Chat Model Interface

```go
type ChatModel interface {
    Chat(ctx context.Context, messages []Message, options ...Option) (ChatResponse, error)
    ChatStream(ctx context.Context, messages []Message, options ...Option) (<-chan ChatChunk, error)
}
```

### BatchProcess Engine

High-performance parallel processing with controlled concurrency and intelligent retry:

```go
// Process multiple prompts concurrently
results := utils.BatchProcess(ctx, llm, prompts, 
    models.WithMaxTokens(100),
    utils.BatchProcessOptions{
        MaxConcurrent: 10,
        MaxRetries:    3,
        RetryDelay:    time.Second,
    },
)
```

## Quick Start

### Installation

```bash
go get github.com/Timkoala/Golangchain
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    
    "golangchain/models"
    "golangchain/models/openai"
    "golangchain/utils"
)

func main() {
    // Create model instance
    model := openai.NewModel(
        "your-api-key",
        "gpt-3.5-turbo-instruct",
        models.WithMaxTokens(100),
        models.WithTemperature(0.7),
    )
    
    // Single request
    ctx := context.Background()
    completions, err := model.Generate(ctx, []string{"Explain quantum computing"})
    if err != nil {
        panic(err)
    }
    fmt.Println(completions[0].Text)
    
    // Batch processing with concurrency control
    prompts := []string{
        "Explain machine learning",
        "What is blockchain?",
        "Describe cloud computing",
    }
    
    results := utils.BatchProcess(ctx, model, prompts, 
        models.WithMaxTokens(100),
        utils.BatchProcessOptions{MaxConcurrent: 5},
    )
    
    for i, result := range results {
        if result.Error != nil {
            fmt.Printf("Prompt %d failed: %v\n", i, result.Error)
        } else {
            fmt.Printf("Prompt %d: %s\n", i, result.Completion.Text)
        }
    }
}
```

### Chat API

```go
messages := []models.Message{
    models.NewSystemMessage("You are a helpful assistant."),
    models.NewUserMessage("What is Go?"),
}

response, err := model.Chat(ctx, messages)
if err != nil {
    panic(err)
}
fmt.Println(response.Message.Content)
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithMaxTokens(n)` | Maximum tokens to generate | 100 |
| `WithTemperature(t)` | Sampling temperature (0.0-2.0) | 0.7 |
| `WithTopP(p)` | Top-p sampling parameter | 1.0 |
| `WithStop([]string)` | Stop sequences | nil |

## Performance Benchmarks

Benchmark comparing serial vs parallel processing of 10 LLM requests:

```
Serial Processing:  ~15.2s
Parallel Processing: ~3.1s  (5 concurrent)
Speedup: 4.9x
```

The speedup approaches the concurrency limit, demonstrating efficient goroutine utilization.

## Roadmap

- [x] Additional providers (Anthropic Claude, Google Gemini)
- [ ] Vector store integration
- [ ] Document loaders
- [x] Agent framework with tool support
- [x] Prompt template system
- [ ] Local model support
- [x] Memory system (buffer, summary, conversation)
- [x] RAG retrieval augmented generation (simple and vector retrievers)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [LangChain](https://github.com/langchain-ai/langchain) for the architectural inspiration
- The Go community for excellent standard library support
