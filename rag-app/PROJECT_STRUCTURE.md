# RAG-App 项目结构

## 📁 完整目录结构

```
rag-app/
├── 📄 核心代码
│   ├── main.go                          # 主程序入口
│   ├── model.go                         # 数据模型
│   ├── repository.go                    # 数据访问层（接口化）
│   ├── rag_service.go                   # 第一代 RAG 服务
│   ├── agentic_rag.go                   # ⭐ 第三代 Agentic RAG
│   └── handler.go                       # HTTP 处理器
│
├── 🔌 生产集成
│   ├── integrations/
│   │   ├── llm/                        # ⭐ LLM 集成
│   │   │   ├── openai.go              # OpenAI 客户端
│   │   │   ├── deepseek.go            # DeepSeek 客户端
│   │   │   └── README.md              # LLM 集成指南
│   │   │
│   │   └── rerank/                     # ⭐ Rerank 集成
│   │       ├── cohere.go              # Cohere Rerank
│   │       ├── bge.go                 # BGE Reranker
│   │       └── README.md              # Rerank 集成指南
│   │
│   └── examples/
│       └── prompts/                    # ⭐ Prompt 优化
│           ├── planning_prompt.go      # 问题规划 Prompt
│           ├── generation_prompt.go    # 答案生成 Prompt
│           └── README.md               # Prompt 优化指南
│
├── 🧪 测试
│   ├── agentic_rag_test.go             # Agentic RAG 测试
│   ├── rag_service_test.go             # RAG 服务测试
│   └── repository_test.go              # 数据访问层测试
│
├── 📚 核心文档
│   ├── README.md                        # ⭐ 项目主 README
│   ├── QUICK_START_GUIDE.md            # ⭐ 快速开始（20分钟）
│   ├── AGENTIC_RAG_V3.md               # 第三代文档
│   ├── IMPLEMENTATION_SUMMARY.md        # 实现总结
│   ├── UPDATE_SUMMARY.md                # 更新总结
│   └── PROJECT_STRUCTURE.md             # 本文档
│
├── 🗺️ RAG 演进
│   ├── RAG_EVOLUTION.md                 # 第一到第四代对比
│   └── MULTIMODAL_RAG_ROADMAP.md        # 第四代路线图
│
├── 🔗 高级集成
│   └── LLAMAINDEX_INTEGRATION.md        # LlamaIndex 集成
│
├── ⚙️ 配置
│   ├── go.mod                           # Go 依赖
│   ├── go.sum                           # Go 依赖锁定
│   └── .env                             # 环境变量（需创建）
│
└── 🏗️ 构建产物
    └── rag-app.exe                      # 可执行文件
```

---

## 🎯 文件分类

### 核心代码（7 个文件）

| 文件 | 功能 | 行数 | 优先级 |
|------|------|------|--------|
| `main.go` | 程序入口 | ~50 | ⭐⭐⭐ |
| `model.go` | 数据模型 | ~50 | ⭐⭐⭐ |
| `repository.go` | 数据访问 | ~90 | ⭐⭐⭐ |
| `rag_service.go` | 第一代 RAG | ~150 | ⭐⭐ |
| `agentic_rag.go` | 第三代 RAG | ~330 | ⭐⭐⭐ |
| `handler.go` | HTTP 路由 | ~100 | ⭐⭐ |

### 生产集成（6 个文件）✨ 新增

| 文件 | 功能 | 状态 |
|------|------|------|
| `integrations/llm/openai.go` | OpenAI 集成 | ✅ 新增 |
| `integrations/llm/deepseek.go` | DeepSeek 集成 | ✅ 新增 |
| `integrations/rerank/cohere.go` | Cohere Rerank | ✅ 新增 |
| `integrations/rerank/bge.go` | BGE Rerank | ✅ 新增 |
| `examples/prompts/planning_prompt.go` | 规划 Prompt | ✅ 新增 |
| `examples/prompts/generation_prompt.go` | 生成 Prompt | ✅ 新增 |

### 文档（12 个文件）

#### 快速开始
- ✅ `README.md` - 项目主 README
- ⭐ `QUICK_START_GUIDE.md` - 20 分钟快速入门

#### 集成指南
- ⭐ `integrations/llm/README.md` - LLM 集成详解
- ⭐ `integrations/rerank/README.md` - Rerank 集成详解
- ⭐ `examples/prompts/README.md` - Prompt 优化详解

#### RAG 演进
- `RAG_EVOLUTION.md` - 四代 RAG 完整对比
- `AGENTIC_RAG_V3.md` - 第三代详细文档
- `MULTIMODAL_RAG_ROADMAP.md` - 第四代路线图

#### 实现细节
- `IMPLEMENTATION_SUMMARY.md` - 第三代实现总结
- `UPDATE_SUMMARY.md` - 最新更新说明
- `PROJECT_STRUCTURE.md` - 项目结构（本文档）

#### 高级话题
- `LLAMAINDEX_INTEGRATION.md` - Python/LlamaIndex 集成

---

## 🔍 关键文件详解

### 1. `main.go` - 程序入口

**职责**：
- 初始化数据库连接
- 创建 LLM、Embedder、Reranker 客户端
- 创建 RAG 服务
- 注册 HTTP 路由
- 启动服务器

**依赖**：
```go
import (
    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "rag-app/integrations/llm"      // ⭐ 新增
    "rag-app/integrations/rerank"   // ⭐ 新增
)
```

### 2. `agentic_rag.go` - 第三代核心

**职责**：
- 问题分析和拆解（QueryPlanner）
- 多轮检索执行（QueryExecutor）
- 结果去重和重排
- 综合生成答案

**核心组件**：
```go
- AgenticRAGService    // 协调器
- QueryPlanner         // 规划器
- QueryExecutor        // 执行器
- QueryPlan            // 规划数据
- ExecutionResults     // 执行结果
```

### 3. `integrations/llm/openai.go` - OpenAI 集成

**职责**：
- 文本生成（Generate）
- Embedding 生成（Embed）
- 图片理解（DescribeImage）

**API**：
```go
client := llm.NewOpenAIClient(llm.OpenAIConfig{
    APIKey: "sk-xxx",
    Model:  "gpt-4o-mini",
})

answer, _ := client.Generate(ctx, prompt)
embedding, _ := client.Embed(ctx, text)
description, _ := client.DescribeImage(ctx, imageURL, "")
```

### 4. `integrations/rerank/cohere.go` - Cohere Rerank

**职责**：
- 文档重排序
- 提升检索准确性

**API**：
```go
reranker := rerank.NewCohereRerankClient(rerank.CohereConfig{
    APIKey: "xxx",
    Model:  "rerank-multilingual-v3.0",
})

results, _ := reranker.Rerank(ctx, query, documents, topK)
```

### 5. `examples/prompts/planning_prompt.go` - Few-shot Prompt

**职责**：
- 提供高质量 Prompt 模板
- 包含 Few-shot 示例
- 提升 LLM 输出质量

**API**：
```go
prompt := prompts.PlanningPrompt("Go 和 Rust 的区别？")
plan, _ := llm.Generate(ctx, prompt)
```

---

## 🚀 快速导航

### 我想...

#### 快速开始
→ 阅读 [`QUICK_START_GUIDE.md`](./QUICK_START_GUIDE.md)

#### 了解第三代 RAG
→ 阅读 [`AGENTIC_RAG_V3.md`](./AGENTIC_RAG_V3.md)

#### 集成真实 LLM
→ 阅读 [`integrations/llm/README.md`](./integrations/llm/README.md)

#### 集成 Rerank
→ 阅读 [`integrations/rerank/README.md`](./integrations/rerank/README.md)

#### 优化 Prompt
→ 阅读 [`examples/prompts/README.md`](./examples/prompts/README.md)

#### 了解 RAG 演进
→ 阅读 [`RAG_EVOLUTION.md`](./RAG_EVOLUTION.md)

#### 规划第四代
→ 阅读 [`MULTIMODAL_RAG_ROADMAP.md`](./MULTIMODAL_RAG_ROADMAP.md)

#### 集成 LlamaIndex
→ 阅读 [`LLAMAINDEX_INTEGRATION.md`](./LLAMAINDEX_INTEGRATION.md)

---

## 📊 代码统计

### 核心代码
```
总文件数：7
总行数：~770
平均行数：110 行/文件
```

### 集成代码（新增）
```
总文件数：6
总行数：~900
平均行数：150 行/文件
```

### 测试代码
```
总文件数：3
总行数：~400
覆盖率：>80%
```

### 文档
```
总文件数：12
总字数：~80,000
平均篇幅：6,600 字/文档
```

---

## 🎯 开发流程

### 1. 新功能开发

```
1. 在 integrations/ 或 examples/ 下创建新文件
2. 实现功能
3. 编写测试（*_test.go）
4. 更新 README.md
5. 提交 PR
```

### 2. Bug 修复

```
1. 定位问题（查看日志、复现）
2. 编写测试用例（证明 bug 存在）
3. 修复代码
4. 验证测试通过
5. 提交 PR
```

### 3. 文档更新

```
1. 确定需要更新的文档
2. 编辑 Markdown 文件
3. 检查链接有效性
4. 提交 PR
```

---

## 🔗 依赖关系

### 核心依赖

```
main.go
  ↓
├─ repository.go (数据访问)
├─ rag_service.go (第一代)
│    ↓
│    └─ agentic_rag.go (第三代)
│
├─ integrations/llm/ (LLM 集成)
├─ integrations/rerank/ (Rerank 集成)
└─ examples/prompts/ (Prompt 优化)
```

### 外部依赖

```go
github.com/gin-gonic/gin         // HTTP 框架
github.com/jmoiron/sqlx          // SQL 扩展
github.com/lib/pq                // PostgreSQL 驱动
github.com/fndome/xb             // 向量查询构建器
github.com/joho/godotenv         // 环境变量
```

---

## 🎨 设计模式

### 1. 接口抽象

```go
type LLMService interface {
    Generate(ctx context.Context, prompt string) (string, error)
}

// 实现
type OpenAIClient struct { ... }
type DeepSeekClient struct { ... }
```

### 2. 依赖注入

```go
func NewAgenticRAGService(
    baseRAG *RAGService,
    reranker *rerank.CohereRerankClient,
) *AgenticRAGService {
    // ...
}
```

### 3. 策略模式

```go
// 根据问题类型选择不同策略
if plan.IsSimple {
    return s.baseRAG.Query(ctx, req)  // 简单策略
}
// 复杂策略
```

---

## 📈 未来规划

### 短期（已完成）
- ✅ 第三代 Agentic RAG
- ✅ LLM 集成
- ✅ Rerank 集成
- ✅ Prompt 优化

### 中期（1-2 周）
- [ ] 缓存机制
- [ ] 监控和日志
- [ ] 性能优化
- [ ] 错误处理增强

### 长期（7 周）
- [ ] 多模态解析
- [ ] 双图谱构建
- [ ] 混合检索
- [ ] 第四代 RAG

---

## 🙏 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

---

**清晰的结构，让协作更高效！** 📁✨

