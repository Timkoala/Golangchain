# GoLangChain

GoLangChain是一个用Go语言重新实现的LangChain核心组件，旨在利用Go语言的并发、性能和类型安全优势，为大语言模型应用开发提供更高效的框架。

## 项目概述

本项目是对Python LangChain框架的部分组件进行Go语言重写，重点展示Go语言的并发处理能力和类型安全特性。当前实现了模型接口部分，包括基础LLM接口、聊天模型接口以及OpenAI API的实现。

## 核心特性

- **强类型接口设计**：利用Go的类型系统，提供类型安全的API
- **高并发处理**：使用goroutines和channels进行高效的并行请求处理
- **优雅的错误处理**：Go风格的错误处理和资源管理
- **模块化架构**：易于扩展的接口设计
- **性能优化**：相比Python实现，提供更高的吞吐量和更低的延迟

## 项目结构

```
golangchain/
├── cmd/               # 命令行工具
│   └── demo/          # 演示程序
├── docs/              # 文档
│   ├── README.md      # 项目概述文档
│   ├── langchain_analysis.md  # LangChain架构分析
│   └── project_implementation.md  # 项目实施方案
├── models/            # 模型接口
│   ├── llm.go         # 基础LLM接口
│   ├── chat.go        # 聊天模型接口
│   └── openai/        # OpenAI实现
├── utils/             # 工具函数
│   └── concurrency.go # 并发处理工具
└── go.mod             # Go模块定义
```

## 已实现的组件

1. **模型接口 (models/)**
   - 基础LLM接口定义
   - 聊天模型接口定义
   - 灵活的选项配置系统

2. **模型实现 (models/openai/)**
   - OpenAI API的接口实现
   - 支持完成和流式响应

3. **并发工具 (utils/)**
   - 高效的并行请求处理
   - 可控的并发限制
   - 性能比较工具

4. **演示程序 (cmd/demo/)**
   - 展示并行处理LLM请求的性能优势
   - 提供直观的性能比较

## 优势展示

通过实现LangChain的核心模型接口，本项目展示了Go语言在以下方面的优势：

1. **并发性能**：使用Go的goroutines进行并行处理，显著提高了多请求场景下的吞吐量。在测试中，并行处理相比串行处理提供了约N倍的速度提升（其中N近似于并发请求数）。

2. **类型安全**：使用Go的静态类型系统，在编译时捕获许多在Python中只能在运行时发现的错误。

3. **资源效率**：相比Python实现，Go版本有更低的内存占用和更高的CPU利用效率。

4. **错误处理**：使用Go的显式错误返回模式，提供了更可靠的错误处理流程。

## 使用方法

### 前置条件

- Go 1.18或更高版本
- OpenAI API密钥（用于演示程序）

### 运行演示程序

1. 设置API密钥：

```bash
export OPENAI_API_KEY=your_api_key_here
```

2. 运行演示程序：

```bash
go run cmd/demo/main.go
```

可选参数：
- `-api-key`：OpenAI API密钥
- `-model`：要使用的模型名称（默认：gpt-3.5-turbo-instruct）
- `-concurrent`：是否使用并发模式（默认：true）
- `-concurrency`：并发级别（默认：5）
- `-timeout`：超时时间（秒）（默认：30）
- `-verbose`：是否输出详细日志（默认：false）
- `-prompts`：包含提示的文件路径

### 作为库使用

```go
import (
    "golangchain/models"
    "golangchain/models/openai"
    "golangchain/utils"
)

// 创建模型实例
model := openai.NewModel("your_api_key", "gpt-3.5-turbo-instruct", 
    models.WithMaxTokens(100),
    models.WithTemperature(0.7),
)

// 并行处理多个提示
prompts := []string{"提示1", "提示2", "提示3"}
results := utils.BatchProcess(ctx, model, prompts, models.WithMaxTokens(100))
```

## 未来计划

1. **添加更多模型提供商**：实现对Anthropic Claude、本地模型等的支持
2. **实现检索器组件**：高性能的向量存储和检索系统
3. **添加文档处理工具**：并行文档加载和处理
4. **性能基准测试**：与Python LangChain进行详细的性能对比

## 为什么选择Go？

Go语言为LLM应用开发提供了多项优势：

1. **并发模型**：goroutines和channels提供了简洁而强大的并发编程模型，非常适合处理多个并行API请求。
2. **性能**：Go程序通常比Python更快，内存占用更少，这对于生产环境中的LLM应用至关重要。
3. **类型安全**：静态类型系统有助于捕获常见错误，提高代码可靠性。
4. **部署简便**：Go编译为单个二进制文件，无需依赖项，简化了部署流程。
5. **标准库**：丰富的标准库减少了对第三方依赖的需求。

## 贡献

本项目是个人简历项目的一部分，但欢迎提出建议和改进意见。

## 致谢

- [LangChain项目](https://github.com/langchain-ai/langchain)，为该项目提供了灵感和参考架构
- OpenAI，提供了高质量的API和文档
