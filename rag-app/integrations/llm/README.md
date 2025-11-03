# LLM 集成指南

本目录包含真实 LLM 的集成实现。

## 🚀 支持的 LLM

### 1. OpenAI
- ✅ **GPT-4o-mini**：推荐用于 RAG（性能好，成本低）
- ✅ **GPT-4o**：多模态支持（图片理解）
- ✅ **text-embedding-3-small**：Embedding 模型

### 2. DeepSeek
- ✅ **deepseek-chat**：DeepSeek V3（国产，性价比高）
- ✅ **多模态支持**：DeepSeek V2.5（图片理解）

---

## 📦 使用方法

### OpenAI

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "rag-app/integrations/llm"
)

func main() {
    // 1. 创建 OpenAI 客户端
    client := llm.NewOpenAIClient(llm.OpenAIConfig{
        APIKey: "sk-xxx", // 你的 OpenAI API Key
        Model:  "gpt-4o-mini", // 可选，默认 gpt-4o-mini
    })
    
    // 2. 生成文本
    prompt := "请解释什么是 RAG？"
    answer, err := client.Generate(context.Background(), prompt)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Answer:", answer)
    
    // 3. 生成 Embedding
    text := "RAG 是检索增强生成的缩写"
    embedding, err := client.Embed(context.Background(), text)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Embedding dimension: %d\n", len(embedding))
    
    // 4. 图片理解（GPT-4V）
    imageURL := "https://example.com/image.jpg"
    description, err := client.DescribeImage(context.Background(), imageURL, "")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Image Description:", description)
}
```

### DeepSeek

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "rag-app/integrations/llm"
)

func main() {
    // 1. 创建 DeepSeek 客户端
    client := llm.NewDeepSeekClient(llm.DeepSeekConfig{
        APIKey: "sk-xxx", // 你的 DeepSeek API Key
        Model:  "deepseek-chat", // 可选，默认 deepseek-chat
    })
    
    // 2. 生成文本
    prompt := "请解释什么是 RAG？"
    answer, err := client.Generate(context.Background(), prompt)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Answer:", answer)
    
    // 3. 图片理解（DeepSeek V2.5）
    imageURL := "https://example.com/image.jpg"
    description, err := client.DescribeImage(context.Background(), imageURL, "")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Image Description:", description)
}
```

---

## 🔧 集成到 RAG-App

### Step 1: 更新 `rag_service.go`

```go
// 替换 MockLLMService 为真实 LLM
import "rag-app/integrations/llm"

func main() {
    // ... 数据库初始化 ...
    
    // 创建 OpenAI 客户端
    openaiClient := llm.NewOpenAIClient(llm.OpenAIConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })
    
    // 创建 RAG 服务
    ragService := NewRAGService(repo, openaiClient, openaiClient)
    
    // 创建 Agentic RAG 服务
    agenticService := NewAgenticRAGService(ragService)
    
    // ... 启动 HTTP 服务 ...
}
```

### Step 2: 设置环境变量

```bash
# OpenAI
export OPENAI_API_KEY="sk-xxx"

# 或 DeepSeek
export DEEPSEEK_API_KEY="sk-xxx"
```

### Step 3: 运行

```bash
go run main.go
```

---

## 💰 成本对比

### OpenAI（2024年价格）

| 模型 | 输入 | 输出 | 推荐场景 |
|------|------|------|---------|
| gpt-4o-mini | $0.15/1M tokens | $0.6/1M tokens | ⭐ RAG 生成 |
| gpt-4o | $2.5/1M tokens | $10/1M tokens | 图片理解 |
| text-embedding-3-small | $0.02/1M tokens | - | ⭐ Embedding |

### DeepSeek（2024年价格）

| 模型 | 输入 | 输出 | 推荐场景 |
|------|------|------|---------|
| deepseek-chat | ¥1/1M tokens | ¥2/1M tokens | ⭐ RAG 生成（性价比） |
| deepseek-coder | ¥1/1M tokens | ¥2/1M tokens | 代码理解 |

**推荐组合**：
- **Embedding**: OpenAI text-embedding-3-small
- **生成**: DeepSeek deepseek-chat（国内用户）或 OpenAI gpt-4o-mini（国际用户）
- **多模态**: OpenAI gpt-4o 或 DeepSeek V2.5

---

## 🎯 高级用法

### 1. 带选项生成

```go
options := map[string]interface{}{
    "temperature":   0.7,  // 温度（0-2）
    "max_tokens":    1000, // 最大 token 数
    "top_p":         0.9,  // 核采样
}

answer, err := client.GenerateWithOptions(ctx, prompt, options)
```

### 2. 流式生成（TODO）

```go
// 未来版本将支持流式输出
stream, err := client.GenerateStream(ctx, prompt)
for chunk := range stream {
    fmt.Print(chunk)
}
```

### 3. 批量 Embedding

```go
texts := []string{
    "文本1",
    "文本2",
    "文本3",
}

var embeddings [][]float32
for _, text := range texts {
    emb, err := client.Embed(ctx, text)
    if err != nil {
        log.Printf("Embed failed for %s: %v", text, err)
        continue
    }
    embeddings = append(embeddings, emb)
}
```

---

## 🔒 安全最佳实践

### 1. 使用环境变量

```bash
# .env 文件
OPENAI_API_KEY=sk-xxx
DEEPSEEK_API_KEY=sk-xxx
```

```go
import "github.com/joho/godotenv"

func init() {
    godotenv.Load()
}

apiKey := os.Getenv("OPENAI_API_KEY")
```

### 2. 速率限制

```go
import "golang.org/x/time/rate"

// 创建限流器（例如：每秒 10 个请求）
limiter := rate.NewLimiter(10, 1)

// 在调用 LLM 前
if err := limiter.Wait(ctx); err != nil {
    return "", err
}

answer, err := client.Generate(ctx, prompt)
```

### 3. 重试机制

```go
import "github.com/avast/retry-go"

var answer string
err := retry.Do(
    func() error {
        var err error
        answer, err = client.Generate(ctx, prompt)
        return err
    },
    retry.Attempts(3),
    retry.Delay(time.Second),
)
```

---

## 📚 相关链接

- [OpenAI API 文档](https://platform.openai.com/docs/api-reference)
- [DeepSeek API 文档](https://platform.deepseek.com/docs)
- [OpenAI Pricing](https://openai.com/pricing)
- [DeepSeek Pricing](https://platform.deepseek.com/pricing)

