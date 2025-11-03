# RAG 检索应用完整示例

这是一个使用 xb 构建的完整 RAG (Retrieval Augmented Generation) 应用，展示如何将文档检索与 LLM 结合。

## ⭐ 第三代 Agentic RAG 已实现！

本应用已升级到**第三代 Agentic RAG**，支持：
- ✅ **智能问题拆解**：自动将复杂问题拆解为多个子问题
- ✅ **多轮召回**：针对每个子问题分别检索
- ✅ **智能规划**：分析问题类型并生成最优策略
- ✅ **结果综合**：将多轮检索结果综合生成答案

详见：**[第三代 Agentic RAG 文档](./AGENTIC_RAG_V3.md)** 🚀

## 📋 功能

- 文档分块和向量化
- 语义检索
- 混合检索（关键词 + 向量）
- 重排序和多样性
- LLM 集成
- **⭐ 第三代 Agentic RAG**（问题拆解 + 多轮召回）

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
```

## 📁 项目结构

```
rag-app/
├── README.md
├── AGENTIC_RAG_V3.md      # ⭐ 第三代 Agentic RAG 文档
├── main.go                # 主程序
├── model.go               # 数据模型
├── repository.go          # 数据访问层
├── rag_service.go         # 第一代 RAG 服务
├── agentic_rag.go         # ⭐ 第三代 Agentic RAG 服务
├── agentic_rag_test.go    # Agentic RAG 测试
├── handler.go             # HTTP 处理器
└── go.mod
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

- **[第三代 Agentic RAG](./AGENTIC_RAG_V3.md)** ⭐ - 问题拆解 + 多轮召回
- **[LlamaIndex 集成](./LLAMAINDEX_INTEGRATION.md)** - Python + Go 集成方案
- [RAG Best Practices](../../xb/doc/ai_application/RAG_BEST_PRACTICES.md)
- [Hybrid Search](../../xb/doc/ai_application/HYBRID_SEARCH.md)
- [Vector Diversity](../../xb/doc/VECTOR_DIVERSITY_QDRANT.md)

