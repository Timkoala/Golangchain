# GoLangChain

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

一个用 Go 语言编写的轻量级、高性能 LLM 应用框架，灵感来自 [LangChain](https://github.com/langchain-ai/langchain)。

[English](README_EN.md)

## 为什么选择 GoLangChain？

Python LangChain 功能强大，但在生产环境中存在显著的性能开销。GoLangChain 解决了这些痛点：

| 对比项 | Python LangChain | GoLangChain |
|--------|------------------|-------------|
| 内存占用 | 约 200MB+ | 约 80MB（降低 60%）|
| 冷启动时间 | 2-3 秒 | < 200ms |
| 并发请求 | 需要 asyncio/threading | 原生 goroutine |
| 依赖管理 | 复杂的依赖树 | 零外部依赖 |
| 部署方式 | 需要虚拟环境 | 单一二进制文件 |

### 核心优势

- **原生并发**：基于 goroutines 和 channels 构建，并行 LLM 请求吞吐量提升 4 倍以上，无需额外的异步框架
- **零依赖**：仅使用 Go 标准库，消除供应链风险，简化部署流程
- **生产就绪**：单一二进制编译，跨平台支持，资源消耗极低
- **类型安全**：通过 Go 静态类型系统实现编译时错误检测

## 项目结构

```
golangchain/
├── models/                 # 模型抽象层
│   ├── llm.go             # 基础 LLM 接口定义
│   ├── chat.go            # 聊天模型接口定义
│   ├── openai/            # OpenAI 提供商实现
│   │   └── model.go
│   ├── anthropic/         # Anthropic Claude 提供商实现
│   │   └── model.go
│   └── google/            # Google Gemini 提供商实现
│       └── model.go
├── agent/                  # Agent 框架
│   └── agent.go           # Agent 核心实现（支持工具调用）
├── prompt/                 # Prompt 模板系统
│   └── template.go        # 模板引擎和构建器
├── memory/                 # 记忆系统
│   └── memory.go          # 缓冲记忆、摘要记忆、对话记忆
├── rag/                    # RAG 检索增强生成
│   └── retriever.go       # 简单检索器、向量检索器
├── utils/                  # 工具函数
│   └── concurrency.go     # BatchProcess 并行引擎（含重试逻辑）
└── go.mod                 # 零外部依赖
```

## 核心组件

### LLM 接口

框架定义了与提供商无关的 LLM 交互接口：

```go
type LLM interface {
    Generate(ctx context.Context, prompts []string, options ...Option) ([]Completion, error)
    GenerateStream(ctx context.Context, prompt string, options ...Option) (<-chan CompletionChunk, error)
}
```

### 聊天模型接口

```go
type ChatModel interface {
    Chat(ctx context.Context, messages []Message, options ...Option) (ChatResponse, error)
    ChatStream(ctx context.Context, messages []Message, options ...Option) (<-chan ChatChunk, error)
}
```

### BatchProcess 并行处理引擎

高性能并行处理，支持可控并发和智能重试：

```go
// 并发处理多个提示
results := utils.BatchProcess(ctx, llm, prompts, 
    models.WithMaxTokens(100),
    utils.BatchProcessOptions{
        MaxConcurrent: 10,
        MaxRetries:    3,
        RetryDelay:    time.Second,
    },
)
```

## 快速开始

### 安装

```bash
go get github.com/Timkoala/Golangchain
```

### 基础用法

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
    // 创建模型实例
    model := openai.NewModel(
        "your-api-key",
        "gpt-3.5-turbo-instruct",
        models.WithMaxTokens(100),
        models.WithTemperature(0.7),
    )
    
    // 单个请求
    ctx := context.Background()
    completions, err := model.Generate(ctx, []string{"解释量子计算"})
    if err != nil {
        panic(err)
    }
    fmt.Println(completions[0].Text)
    
    // 批量并发处理
    prompts := []string{
        "解释机器学习",
        "什么是区块链？",
        "描述云计算",
    }
    
    results := utils.BatchProcess(ctx, model, prompts, 
        models.WithMaxTokens(100),
        utils.BatchProcessOptions{MaxConcurrent: 5},
    )
    
    for i, result := range results {
        if result.Error != nil {
            fmt.Printf("提示 %d 失败: %v\n", i, result.Error)
        } else {
            fmt.Printf("提示 %d: %s\n", i, result.Completion.Text)
        }
    }
}
```

### 聊天 API

```go
messages := []models.Message{
    models.NewSystemMessage("你是一个有帮助的助手。"),
    models.NewUserMessage("什么是 Go 语言？"),
}

response, err := model.Chat(ctx, messages)
if err != nil {
    panic(err)
}
fmt.Println(response.Message.Content)
```

### Agent 框架（工具调用）

```go
import "golangchain/agent"

// 定义工具
tools := []agent.Tool{
    {
        Name:        "calculator",
        Description: "执行数学计算",
        Execute: func(input string) (string, error) {
            // 实现计算逻辑
            return "结果", nil
        },
    },
}

// 创建 Agent
ag := agent.NewAgent(model, tools)

// 执行任务
result, err := ag.Execute(ctx, "计算 2+2 的结果")
if err != nil {
    panic(err)
}
fmt.Println(result)
```

### Prompt 模板系统

```go
import "golangchain/prompt"

// 创建模板
template := prompt.NewTemplate(
    "你是一个{{role}}。用户问题：{{question}}",
)

// 渲染模板
rendered, err := template.Render(map[string]interface{}{
    "role":     "Python 专家",
    "question": "如何优化列表推导式？",
})
if err != nil {
    panic(err)
}
fmt.Println(rendered)

// 使用模板构建器
builder := prompt.NewPromptBuilder()
builder.AddSystemMessage("你是一个有帮助的助手")
builder.AddUserMessage("{{question}}")
prompt := builder.Build()
```

### 记忆系统

```go
import "golangchain/memory"

// 缓冲记忆：保留最近 N 条消息
bufferMem := memory.NewBufferMemory(10)
bufferMem.Add(models.NewUserMessage("你好"))
bufferMem.Add(models.NewAssistantMessage("你好！有什么我可以帮助的吗？"))

// 对话记忆：支持消息计数和摘要
convMem := memory.NewConversationMemory(100, 50)
convMem.Add(models.NewUserMessage("什么是 Go？"))
count := convMem.GetMessageCount()
fmt.Printf("对话消息数：%d\n", count)

// 获取历史消息
messages := bufferMem.Get()
for _, msg := range messages {
    fmt.Printf("[%s]: %s\n", msg.Role, msg.Content)
}
```

### RAG 检索增强生成

```go
import "golangchain/rag"

// 创建检索器
retriever := rag.NewSimpleRetriever()

// 添加文档
docs := []rag.Document{
    {
        ID:      "doc1",
        Content: "Go 是一门编译型编程语言，具有高效的并发能力",
    },
    {
        ID:      "doc2",
        Content: "Python 是一门解释型编程语言，易于学习和使用",
    },
}

for _, doc := range docs {
    retriever.Add(doc)
}

// 搜索相关文档
results, err := retriever.Search("Go 编程语言", 5)
if err != nil {
    panic(err)
}

for _, result := range results {
    fmt.Printf("文档 %s (相似度: %.2f): %s\n", 
        result.Document.ID, result.Score, result.Document.Content)
}

// 使用向量检索器获得更好的语义搜索
vectorRetriever := rag.NewVectorRetriever()
for _, doc := range docs {
    vectorRetriever.Add(doc)
}

results, _ = vectorRetriever.Search("编程", 5)
```

### 多提供商支持

```go
import (
    "golangchain/models/openai"
    "golangchain/models/anthropic"
    "golangchain/models/google"
)

// OpenAI
openaiModel := openai.NewModel("your-openai-key", "gpt-4")

// Anthropic Claude
claudeModel := anthropic.NewModel("your-anthropic-key", "claude-3-opus-20240229")

// Google Gemini
geminiModel := google.NewModel("your-google-key", "gemini-pro")

// 统一接口调用
ctx := context.Background()
response, _ := openaiModel.Chat(ctx, messages)
response, _ = claudeModel.Chat(ctx, messages)
response, _ = geminiModel.Chat(ctx, messages)
```

## 技术特色

### 🚀 高性能并发引擎

基于 Go 原生 goroutines 和 channels 构建，无需外部异步框架：

- **信号量控制**：精确控制并发数量，避免 API 限流
- **智能重试**：指数退避策略，自动处理临时错误
- **上下文传播**：完整支持 context 取消和超时

### 🔧 模块化架构

- **提供商无关**：统一 LLM/ChatModel 接口，轻松切换模型
- **可插拔工具**：Agent 工具系统支持自定义扩展
- **灵活记忆**：多种记忆策略满足不同场景需求

### 🛡️ 生产级特性

- **零外部依赖**：仅使用 Go 标准库，消除供应链风险
- **线程安全**：所有组件均支持并发访问
- **类型安全**：编译时错误检测，减少运行时异常

### 📦 部署友好

- **单一二进制**：无需安装运行时环境
- **跨平台支持**：Linux、macOS、Windows 全平台兼容
- **容器优化**：最小镜像体积，快速启动

## 配置选项

| 选项 | 描述 | 默认值 |
|------|------|--------|
| `WithMaxTokens(n)` | 最大生成 token 数 | 100 |
| `WithTemperature(t)` | 采样温度 (0.0-2.0) | 0.7 |
| `WithTopP(p)` | Top-p 采样参数 | 1.0 |
| `WithStop([]string)` | 停止序列 | nil |

## 性能基准测试

10 个 LLM 请求的串行与并行处理对比：

```
串行处理:  约 15.2s
并行处理:  约 3.1s  (5 并发)
加速比: 4.9x
```

加速比接近并发限制，证明了 goroutine 的高效利用。

## 发展路线

- [x] 更多提供商支持（Anthropic Claude、Google Gemini）
- [ ] 向量存储集成
- [ ] 文档加载器
- [x] Agent 框架（支持工具调用）
- [x] Prompt 模板系统
- [ ] 本地模型支持
- [x] 记忆系统（缓冲记忆、摘要记忆、对话记忆）
- [x] RAG 检索增强生成（简单检索器、向量检索器）

## 贡献

欢迎贡献！请随时提交 Issue 和 Pull Request。

## 许可证

MIT 许可证 - 详见 [LICENSE](LICENSE)。

## 致谢

- [LangChain](https://github.com/langchain-ai/langchain) 提供的架构灵感
- Go 社区提供的优秀标准库支持
