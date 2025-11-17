package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// REFRAGService REFRAG 风格的 RAG 服务
// 核心思路：压缩 + 智能选择 + 混合输入
// 参考：Meta REFRAG (Rethinking RAG based Decoding)
type REFRAGService struct {
	repo     ChunkRepository
	embedder EmbeddingService
	llm      LLMService
}

func NewREFRAGService(repo ChunkRepository, embedder EmbeddingService, llm LLMService) *REFRAGService {
	return &REFRAGService{
		repo:     repo,
		embedder: embedder,
		llm:      llm,
	}
}

// REFRAGQueryRequest REFRAG 查询请求
type REFRAGQueryRequest struct {
	Question         string `json:"question" binding:"required"`
	DocType          string `json:"doc_type"`
	Language         string `json:"language"`
	OverFetchK       int    `json:"over_fetch_k"`      // 过度获取数量（默认 100）
	ExpandK          int    `json:"expand_k"`          // 解压还原数量（默认 5）
	CompressionRatio int    `json:"compression_ratio"` // 压缩比例（默认 16，即每 16 个 token 压缩成 1 个）
}

// REFRAGQueryResponse REFRAG 查询响应
type REFRAGQueryResponse struct {
	Answer           string                 `json:"answer"`
	ExpandedChunks   []*CompressedChunk     `json:"expanded_chunks"`   // 解压的完整 chunks
	CompressedChunks []*CompressedChunk     `json:"compressed_chunks"` // 保持压缩的 chunks
	Metadata         map[string]interface{} `json:"metadata"`
}

// CompressedChunk 压缩后的文档块
type CompressedChunk struct {
	OriginalChunk        *DocumentChunk `json:"original_chunk"`
	CompressedVector     []float32      `json:"compressed_vector"`      // 压缩后的向量（极短）
	Score                float64        `json:"score"`                  // 相关性评分
	IsExpanded           bool           `json:"is_expanded"`            // 是否已解压
	TokenCount           int            `json:"token_count"`            // 原始 token 数
	CompressedTokenCount int            `json:"compressed_token_count"` // 压缩后 token 数
}

// Query REFRAG 查询
func (s *REFRAGService) Query(ctx context.Context, req REFRAGQueryRequest) (*REFRAGQueryResponse, error) {
	// 1. 将问题转换为向量
	queryVector, err := s.embedder.Embed(ctx, req.Question)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	// 2. 设置默认参数
	overFetchK := req.OverFetchK
	if overFetchK == 0 {
		overFetchK = 100 // 默认过度获取 100 个 chunks
	}

	expandK := req.ExpandK
	if expandK == 0 {
		expandK = 5 // 默认解压 5 个最相关的 chunks
	}

	compressionRatio := req.CompressionRatio
	if compressionRatio == 0 {
		compressionRatio = 16 // 默认每 16 个 token 压缩成 1 个
	}

	// 3. 过度获取大量 chunks（传统 RAG 只取 Top-K，这里取更多）
	chunks, err := s.repo.VectorSearch(queryVector, req.DocType, req.Language, overFetchK)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(chunks) == 0 {
		return &REFRAGQueryResponse{
			Answer:           "抱歉，没有找到相关文档。",
			ExpandedChunks:   []*CompressedChunk{},
			CompressedChunks: []*CompressedChunk{},
			Metadata: map[string]interface{}{
				"chunks_found": 0,
				"over_fetch_k": overFetchK,
			},
		}, nil
	}

	// 4. 压缩所有 chunks（生成块向量）
	compressedChunks := s.compressChunks(chunks, queryVector, compressionRatio)

	// 5. 使用策略网络评分（这里简化实现，使用向量相似度 + 关键词匹配）
	s.scoreChunks(compressedChunks, req.Question)

	// 6. 选择 Top-K 最相关的 chunks 进行解压
	expandedChunks, remainingCompressed := s.selectAndExpand(compressedChunks, expandK)

	// 7. 构建混合提示词（完整文本 + 压缩向量）
	prompt := s.buildHybridPrompt(req.Question, expandedChunks, remainingCompressed)

	// 8. 调用 LLM 生成答案
	answer, err := s.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm generation failed: %w", err)
	}

	// 9. 计算统计信息
	totalTokens := 0
	compressedTokens := 0
	for _, chunk := range expandedChunks {
		totalTokens += chunk.TokenCount
	}
	for _, chunk := range remainingCompressed {
		compressedTokens += chunk.CompressedTokenCount
	}

	return &REFRAGQueryResponse{
		Answer:           answer,
		ExpandedChunks:   expandedChunks,
		CompressedChunks: remainingCompressed,
		Metadata: map[string]interface{}{
			"chunks_found":      len(chunks),
			"expanded_count":    len(expandedChunks),
			"compressed_count":  len(remainingCompressed),
			"over_fetch_k":      overFetchK,
			"expand_k":          expandK,
			"compression_ratio": compressionRatio,
			"total_tokens":      totalTokens,
			"compressed_tokens": compressedTokens,
			"token_reduction":   fmt.Sprintf("%.1f%%", float64(compressedTokens)/float64(totalTokens+compressedTokens)*100),
		},
	}, nil
}

// compressChunks 压缩文档块
// 将每个 chunk 压缩成一个极短的向量（模拟 REFRAG 的压缩过程）
func (s *REFRAGService) compressChunks(chunks []*DocumentChunk, queryVector []float32, ratio int) []*CompressedChunk {
	compressed := make([]*CompressedChunk, 0, len(chunks))

	for _, chunk := range chunks {
		// 估算 token 数（简化：按字符数 / 4）
		tokenCount := len(chunk.Content) / 4
		compressedTokenCount := tokenCount / ratio
		if compressedTokenCount < 1 {
			compressedTokenCount = 1
		}

		// 生成压缩向量（简化：使用原始 embedding 的降维版本）
		// 实际 REFRAG 使用专门的编码器，这里用前 N 维作为压缩向量
		compressedVector := make([]float32, min(32, len(chunk.Embedding)))
		copy(compressedVector, chunk.Embedding[:min(32, len(chunk.Embedding))])

		compressed = append(compressed, &CompressedChunk{
			OriginalChunk:        chunk,
			CompressedVector:     compressedVector,
			Score:                0, // 稍后评分
			IsExpanded:           false,
			TokenCount:           tokenCount,
			CompressedTokenCount: compressedTokenCount,
		})
	}

	return compressed
}

// scoreChunks 对压缩后的 chunks 进行评分
// REFRAG 使用强化学习训练的策略网络，这里简化实现
func (s *REFRAGService) scoreChunks(chunks []*CompressedChunk, question string) {
	// 提取问题关键词
	questionWords := extractKeywords(question)

	for _, chunk := range chunks {
		score := 0.0

		// 1. 向量相似度（使用原始 embedding）
		// 这里简化，实际应该使用 queryVector 与 chunk.Embedding 的相似度
		// 假设已经在 VectorSearch 中按相似度排序，这里给基础分
		score += 0.5

		// 2. 关键词匹配度
		contentWords := extractKeywords(chunk.OriginalChunk.Content)
		keywordMatch := calculateKeywordMatch(questionWords, contentWords)
		score += keywordMatch * 0.3

		// 3. 信息密度（内容长度适中得分更高）
		contentLen := len(chunk.OriginalChunk.Content)
		if contentLen >= 100 && contentLen <= 500 {
			score += 0.1
		} else if contentLen < 50 {
			score -= 0.1 // 太短的内容可能信息不足
		}

		// 4. 元数据相关性（如果有）
		if chunk.OriginalChunk.Metadata != "" {
			score += 0.1
		}

		chunk.Score = score
	}
}

// selectAndExpand 选择 Top-K chunks 进行解压
func (s *REFRAGService) selectAndExpand(chunks []*CompressedChunk, expandK int) ([]*CompressedChunk, []*CompressedChunk) {
	// 按评分排序
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Score > chunks[j].Score
	})

	// 选择 Top-K 解压
	expanded := make([]*CompressedChunk, 0, expandK)
	compressed := make([]*CompressedChunk, 0)

	for i, chunk := range chunks {
		if i < expandK {
			chunk.IsExpanded = true
			expanded = append(expanded, chunk)
		} else {
			compressed = append(compressed, chunk)
		}
	}

	return expanded, compressed
}

// buildHybridPrompt 构建混合提示词
// 包含：完整文本（解压的 chunks）+ 压缩向量摘要（保持压缩的 chunks）
func (s *REFRAGService) buildHybridPrompt(question string, expanded []*CompressedChunk, compressed []*CompressedChunk) string {
	var sb strings.Builder

	sb.WriteString("请根据以下文档内容回答问题。\n\n")

	// 1. 完整文档（解压的 chunks）
	if len(expanded) > 0 {
		sb.WriteString("【核心文档】（完整内容）：\n")
		for i, chunk := range expanded {
			sb.WriteString(fmt.Sprintf("\n[文档 %d] (相关性: %.2f)\n", i+1, chunk.Score))
			sb.WriteString(chunk.OriginalChunk.Content)
			sb.WriteString("\n")
		}
	}

	// 2. 压缩文档摘要（保持压缩的 chunks）
	if len(compressed) > 0 {
		sb.WriteString(fmt.Sprintf("\n【背景文档】（已压缩，共 %d 个）：\n", len(compressed)))
		sb.WriteString("以下文档与问题相关，但已压缩以节省 token。如需详细信息，请参考核心文档。\n")

		// 只显示前 10 个压缩文档的摘要
		maxCompressed := min(10, len(compressed))
		for i := 0; i < maxCompressed; i++ {
			chunk := compressed[i]
			// 显示前 50 个字符作为摘要
			summary := chunk.OriginalChunk.Content
			if len(summary) > 50 {
				summary = summary[:50] + "..."
			}
			sb.WriteString(fmt.Sprintf("- [文档 %d] (相关性: %.2f) %s\n", i+1, chunk.Score, summary))
		}

		if len(compressed) > maxCompressed {
			sb.WriteString(fmt.Sprintf("... 还有 %d 个相关文档已压缩\n", len(compressed)-maxCompressed))
		}
	}

	sb.WriteString(fmt.Sprintf("\n问题：%s\n\n", question))
	sb.WriteString("请基于上述文档内容进行回答。优先使用【核心文档】中的信息，【背景文档】提供补充上下文。")

	return sb.String()
}

// 辅助函数

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func extractKeywords(text string) map[string]bool {
	// 简化实现：提取常见的中文和英文单词
	words := make(map[string]bool)

	// 移除标点符号，转换为小写
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, ",", " ")
	text = strings.ReplaceAll(text, ".", " ")
	text = strings.ReplaceAll(text, "?", " ")
	text = strings.ReplaceAll(text, "!", " ")
	text = strings.ReplaceAll(text, "。", " ")
	text = strings.ReplaceAll(text, "，", " ")
	text = strings.ReplaceAll(text, "？", " ")
	text = strings.ReplaceAll(text, "！", " ")

	parts := strings.Fields(text)
	for _, part := range parts {
		if len(part) > 1 { // 忽略单字符
			words[part] = true
		}
	}

	return words
}

func calculateKeywordMatch(queryWords, contentWords map[string]bool) float64 {
	if len(queryWords) == 0 {
		return 0.0
	}

	matchCount := 0
	for word := range queryWords {
		if contentWords[word] {
			matchCount++
		}
	}

	return float64(matchCount) / float64(len(queryWords))
}
