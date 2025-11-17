# REFRAG 风格 RAG 示例

## 📖 什么是 REFRAG？

**REFRAG**（Rethinking RAG based Decoding）是 Meta 提出的 RAG 优化技术，核心思路是：

> **在把文本塞给 LLM 之前，就把 99% 的噪音干掉**

### 核心优势

- ⚡ **速度提升 30 倍**：首 token 延迟大幅降低
- 📈 **上下文扩大 16 倍**：从 4k/8k 扩展到 64k+
- 💰 **成本降低 50-75%**：Token 消耗减少 2-4 倍
- 🎯 **准确率提升**：在多个 RAG 基准测试上全面超越

## 🔄 传统 RAG vs REFRAG

### 传统 RAG 的问题

```
查询 → 向量检索 Top-K (100个) → 全部塞给 LLM → 模型读垃圾
```

**痛点**：
- 90% 的 chunk 其实没用
- 上下文窗口被灌满，速度暴跌
- 算力账单爆炸

### REFRAG 的解决方案

```
查询 → 过度获取 (100个) → 压缩成块向量 → 智能评分 → 选择 Top-K 解压 → 混合输入 LLM
```

**优势**：
- 只解压最相关的几个 chunks（完整文本）
- 其余保持压缩态（几乎不占 token）
- 模型看到：高质量完整文本 + 海量压缩向量

## 🏗️ 实现架构

### 1. 压缩阶段（Compression）

```go
// 每 16 个 token 压缩成 1 个块向量
compressedChunks := s.compressChunks(chunks, queryVector, 16)
```

- 使用轻量级编码器压缩文档块
- 生成极短的压缩向量（几乎不占 token）
- 保留原始 embedding 用于评分

### 2. 评分阶段（Scoring）

```go
// 使用策略网络评分（简化实现：向量相似度 + 关键词匹配）
s.scoreChunks(compressedChunks, question)
```

**评分维度**：
- 向量相似度（50%）
- 关键词匹配度（30%）
- 信息密度（10%）
- 元数据相关性（10%）

### 3. 选择阶段（Selection）

```go
// 选择 Top-K 最相关的 chunks 进行解压
expanded, compressed := s.selectAndExpand(compressedChunks, expandK)
```

- 按评分排序
- Top-K 解压成完整文本
- 其余保持压缩态

### 4. 混合输入（Hybrid Input）

```go
// 构建混合提示词：完整文本 + 压缩向量摘要
prompt := s.buildHybridPrompt(question, expanded, compressed)
```

**输入结构**：
- **核心文档**：完整文本（解压的 chunks）
- **背景文档**：压缩摘要（保持压缩的 chunks）

## 🚀 使用示例

### 基本用法

```go
service := NewREFRAGService(repo, embedder, llm)

req := REFRAGQueryRequest{
    Question:        "Go 和 Rust 在并发编程上有什么区别？",
    DocType:         "article",
    Language:        "zh",
    OverFetchK:      100,  // 过度获取 100 个 chunks
    ExpandK:         5,    // 只解压 5 个最相关的
    CompressionRatio: 16,  // 每 16 个 token 压缩成 1 个
}

resp, err := service.Query(ctx, req)
```

### 参数说明

| 参数 | 说明 | 默认值 | 建议值 |
|------|------|--------|--------|
| `OverFetchK` | 过度获取数量 | 100 | 50-200 |
| `ExpandK` | 解压还原数量 | 5 | 3-10 |
| `CompressionRatio` | 压缩比例 | 16 | 8-32 |

### 响应结构

```go
type REFRAGQueryResponse struct {
    Answer          string                 // LLM 生成的答案
    ExpandedChunks  []*CompressedChunk     // 解压的完整 chunks
    CompressedChunks []*CompressedChunk    // 保持压缩的 chunks
    Metadata        map[string]interface{} // 统计信息
}
```

**Metadata 包含**：
- `chunks_found`: 检索到的总 chunks 数
- `expanded_count`: 解压的 chunks 数
- `compressed_count`: 保持压缩的 chunks 数
- `total_tokens`: 完整文本的 token 数
- `compressed_tokens`: 压缩文本的 token 数
- `token_reduction`: Token 减少比例

## 📊 性能对比

### 示例场景

假设检索到 100 个 chunks，每个 200 tokens：

| 方案 | 输入 Token 数 | 速度 | 成本 |
|------|--------------|------|------|
| **传统 RAG** | 20,000 (100 × 200) | 1x | 100% |
| **REFRAG** | 1,250 (5 × 200 + 95 × 1.25) | 30x | 6.25% |

**Token 减少**：93.75%  
**速度提升**：30 倍  
**成本降低**：93.75%

## 🎯 最佳实践

### 1. 合理设置 OverFetchK

```go
// 简单问题：50-100
// 复杂问题：100-200
// 长文档问答：200-500
overFetchK := 100
```

### 2. 根据问题复杂度调整 ExpandK

```go
// 简单问题：3-5
// 复杂问题：5-10
// 多轮对话：10-20
expandK := 5
```

### 3. 优化压缩比例

```go
// 短文档（< 100 tokens）：8-16
// 中等文档（100-500 tokens）：16-24
// 长文档（> 500 tokens）：24-32
compressionRatio := 16
```

### 4. 结合 Rerank 模型

```go
// 在评分阶段使用 Rerank 模型提升准确性
// 参见 integrations/rerank/
```

## 🔧 扩展实现

### 1. 使用真实的压缩编码器

当前实现使用简化的压缩方法，生产环境应使用：
- 专门的压缩编码器（如 REFRAG 论文中的方法）
- 强化学习训练的策略网络

### 2. 集成 Rerank 模型

```go
// 在 scoreChunks 中使用 Rerank 模型
// 参见 integrations/rerank/bge.go
```

### 3. 动态调整参数

```go
// 根据问题复杂度动态调整 OverFetchK 和 ExpandK
if isComplexQuestion(question) {
    overFetchK = 200
    expandK = 10
}
```

## 📚 相关资源

- **论文**：[REFRAG: Rethinking RAG based Decoding](https://arxiv.org/pdf/2509.01092)
- **xb 文档**：[RAG Best Practices](../../xb/doc/ai_application/RAG_BEST_PRACTICES.md)
- **向量检索**：[Vector Guide](../../xb/doc/VECTOR_GUIDE.md)
- **混合搜索**：[Hybrid Search](../../xb/doc/ai_application/HYBRID_SEARCH.md)

## 🧪 测试

```bash
cd rag-app
go test -v -run TestREFRAGService
```

## 📝 总结

REFRAG 风格的 RAG 通过**压缩 + 智能选择 + 混合输入**，实现了：

1. ✅ **速度提升**：输入序列缩短，注意力计算减少
2. ✅ **成本降低**：Token 消耗大幅减少
3. ✅ **准确率提升**：只给模型看真正有用的内容
4. ✅ **可扩展性**：支持更大规模的文档处理

**未来属于会"精打细算"的 RAG，而 REFRAG 就是第一个真正做到的人。**

---

**Version**: v1.0.0  
**Last Updated**: 2025-01-XX

