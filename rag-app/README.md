# RAG 检索应用完整示例

这是一个使用 xb 构建的完整 RAG (Retrieval Augmented Generation) 应用，展示如何将文档检索与 LLM 结合。

## ⭐ 第三代 Agentic RAG 已实现！第四代示例已完成！REFRAG 风格示例已添加！

本应用已升级到**第三代 Agentic RAG**，支持：
- ✅ **智能问题拆解**：自动将复杂问题拆解为多个子问题
- ✅ **多轮召回**：针对每个子问题分别检索
- ✅ **智能规划**：分析问题类型并生成最优策略
- ✅ **结果综合**：将多轮检索结果综合生成答案

**第四代多模态 RAG** 示例已完成：
- ✅ **10 个完整示例**：展示 xb 在多模态场景的用法
- ✅ **多模态数据模型**：图像、表格、公式、文本
- ✅ **知识图谱示例**：节点、边、遍历
- ✅ **xb 使用技巧**：指针类型、In() 方法、最佳实践
- ✅ **数据库 Schema**：包含索引优化和示例数据

**⭐ REFRAG 风格 RAG** 示例已添加：
- ✅ **压缩 + 智能选择**：在送入 LLM 前过滤 99% 噪音
- ✅ **混合输入**：完整文本 + 压缩向量
- ✅ **性能优化**：速度提升 30 倍，成本降低 50-75%
- ✅ **完整实现**：包含压缩、评分、选择、混合输入全流程

详见：
- **[第三代 Agentic RAG 文档](./AGENTIC_RAG_V3.md)** 🚀（已实现）
- **[第四代多模态 RAG 路线图](./MULTIMODAL_RAG_ROADMAP.md)** 📋（规划中）
- **[第四代示例代码](./g4/README.md)** 💎（已完成 - 10 个示例）
- **[REFRAG 风格指南](./REFRAG_GUIDE.md)** ⚡（新增 - 压缩优化）
- **[RAG 演进史](./RAG_EVOLUTION.md)** 📚（完整对比）

## 📋 功能

### 核心功能
- ✅ 文档分块和向量化
- ✅ 语义检索
- ✅ 混合检索（关键词 + 向量）
- ✅ **第三代 Agentic RAG**（问题拆解 + 多轮召回）
- ✅ **⭐ REFRAG 风格 RAG**（压缩 + 智能选择 + 混合输入）

### 生产就绪集成
- ✅ **真实 LLM**：OpenAI, DeepSeek（参见 `integrations/llm/`）
- ✅ **Rerank 模型**：Cohere, BGE-Reranker（参见 `integrations/rerank/`）
- ✅ **优化 Prompt**：Few-shot 学习（参见 `examples/prompts/`）

⭐ **[快速开始指南](./QUICK_START_GUIDE.md)** - 20 分钟完成三大集成！

## 🏗️ 架构

```
用户查询 → 向量化 → xb 检索 → 重排序 → LLM 生成 → 回答
            ↓           ↓          ↓
         Embedding   PostgreSQL  Application
                     或 Qdrant    Layer
```

## 🚀 快速开始

### 1. 安装依赖

```bash
go get github.com/fndome/xb
go get github.com/jmoiron/sqlx
go get github.com/lib/pq
go get github.com/gin-gonic/gin
```

### 2. 创建数据库

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE document_chunks (
    id BIGSERIAL PRIMARY KEY,
    doc_id BIGINT,
    chunk_id INT,
    content TEXT,
    embedding vector(768),
    doc_type VARCHAR(50),
    language VARCHAR(10),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX ON document_chunks USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX ON document_chunks (doc_type);
CREATE INDEX ON document_chunks (language);
```

### 3. 运行应用

```bash
cd examples/rag-app
go run *.go
```

### 4. 测试 API

```bash
# 上传文档
curl -X POST http://localhost:8080/api/documents \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Go语言并发编程",
    "content": "Goroutine和Channel是Go语言并发编程的核心...",
    "doc_type": "article",
    "language": "zh"
  }'

# RAG 查询（默认使用第三代 Agentic RAG）
curl -X POST http://localhost:8080/api/rag/query \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Go 和 Rust 在并发编程上有什么区别？",
    "doc_type": "article",
    "top_k": 5
  }'

# 如需使用第一代 RAG（简单问题）
curl -X POST http://localhost:8080/api/rag/query \
  -H "Content-Type: application/json" \
  -d '{
    "question": "什么是 Channel？",
    "use_agentic": false
  }'

# ⭐ REFRAG 风格查询（压缩 + 智能选择）
curl -X POST http://localhost:8080/api/rag/refrag \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Go 和 Rust 在并发编程上有什么区别？",
    "doc_type": "article",
    "over_fetch_k": 100,
    "expand_k": 5,
    "compression_ratio": 16
  }'
```

## 📁 项目结构

```
rag-app/
├── 核心代码
│   ├── main.go                # 主程序
│   ├── model.go               # 数据模型
│   ├── repository.go          # 数据访问层
│   ├── rag_service.go         # 第一代 RAG 服务
│   ├── agentic_rag.go         # ⭐ 第三代 Agentic RAG 服务
│   ├── refrag_service.go      # ⭐ REFRAG 风格 RAG 服务
│   └── handler.go             # HTTP 处理器
│
├── 生产集成
│   ├── integrations/llm/      # ⭐ LLM 集成（OpenAI, DeepSeek）
│   ├── integrations/rerank/   # ⭐ Rerank 集成（Cohere, BGE）
│   └── examples/prompts/      # ⭐ Prompt 优化（Few-shot）
│
├── 第四代示例
│   └── g4/                    # ⭐ 多模态 RAG 示例
│       ├── model.go          # 多模态数据模型
│       ├── multimodal_repository.go  # 数据访问层
│       ├── example_test.go   # 10 个完整示例
│       ├── XB_USAGE_TIPS.md  # xb 使用技巧
│       └── sql/              # 数据库 Schema
│
└── 文档
    ├── README.md                    # 本文档
    ├── QUICK_START_GUIDE.md         # 快速开始
    ├── RAG_EVOLUTION.md             # RAG 演进史
    ├── AGENTIC_RAG_V3.md            # 第三代文档
    └── MULTIMODAL_RAG_ROADMAP.md    # 第四代路线图
```

## 🔗 LlamaIndex 集成

xb 可以作为 LlamaIndex 的向量存储后端，提供高性能检索：

- **[LlamaIndex 集成指南](./LLAMAINDEX_INTEGRATION.md)** ⭐
- Python/LlamaIndex（AI 层）+ Go/xb（检索层）
- 向量检索性能提升 20 倍

**优势**：
- ✅ Go 后端：高性能向量检索
- ✅ Python 前端：丰富的 LLM 生态
- ✅ 最佳组合：各自发挥所长

---

## 📚 相关文档

### 快速开始
- **[快速开始指南](./QUICK_START_GUIDE.md)** ⚡ - 20 分钟完成三大集成

### 集成指南（生产就绪）
- **[LLM 集成](./integrations/llm/README.md)** 🤖 - OpenAI, DeepSeek
- **[Rerank 集成](./integrations/rerank/README.md)** 🎯 - Cohere, BGE-Reranker
- **[Prompt 优化](./examples/prompts/README.md)** 💡 - Few-shot 学习

### RAG 演进
- **[RAG 演进史](./RAG_EVOLUTION.md)** 📚 - 第一代到第四代完整对比
- **[第三代 Agentic RAG](./AGENTIC_RAG_V3.md)** ⭐ - 问题拆解 + 多轮召回（已实现）
- **[第四代多模态 RAG 路线图](./MULTIMODAL_RAG_ROADMAP.md)** 🚀 - 双图谱 + 多模态（规划中）
- **[第四代示例代码](./g4/README.md)** 💎 - 10 个完整示例（已完成）
- **[REFRAG 风格指南](./REFRAG_GUIDE.md)** ⚡ - 压缩优化方案（新增）
- **[实现总结](./IMPLEMENTATION_SUMMARY.md)** 📝 - 第三代实现详情
- **[更新总结](./UPDATE_SUMMARY.md)** 📋 - 最新更新内容

### 高级话题
- **[LlamaIndex 集成](./LLAMAINDEX_INTEGRATION.md)** - Python + Go 集成方案
- [RAG Best Practices](../../xb/doc/ai_application/RAG_BEST_PRACTICES.md)
- [Hybrid Search](../../xb/doc/ai_application/HYBRID_SEARCH.md)
- [Vector Diversity](../../xb/doc/VECTOR_DIVERSITY_QDRANT.md)

