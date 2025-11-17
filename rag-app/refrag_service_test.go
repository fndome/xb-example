package main

import (
	"context"
	"testing"
)

func TestREFRAGService_Query(t *testing.T) {
	// 创建模拟服务
	repo := &MockChunkRepository{
		chunks: []*DocumentChunk{
			{
				ID:      1,
				Content: "Go 语言是一种开源的编程语言，由 Google 开发。它专注于简洁性、并发性和性能。",
				Embedding: []float32{0.1, 0.2, 0.3, 0.4, 0.5},
				DocType: "article",
				Language: "zh",
			},
			{
				ID:      2,
				Content: "Rust 是一种系统编程语言，注重内存安全和并发性。它使用所有权系统来保证内存安全。",
				Embedding: []float32{0.2, 0.3, 0.4, 0.5, 0.6},
				DocType: "article",
				Language: "zh",
			},
			{
				ID:      3,
				Content: "Python 是一种高级编程语言，广泛用于数据科学和机器学习。",
				Embedding: []float32{0.3, 0.4, 0.5, 0.6, 0.7},
				DocType: "article",
				Language: "zh",
			},
		},
	}

	embedder := &MockEmbeddingService{}
	llm := &MockLLMService{}

	service := NewREFRAGService(repo, embedder, llm)

	// 测试查询
	req := REFRAGQueryRequest{
		Question:        "Go 和 Rust 在并发编程上有什么区别？",
		DocType:         "article",
		Language:        "zh",
		OverFetchK:      10,
		ExpandK:        2,
		CompressionRatio: 16,
	}

	resp, err := service.Query(context.Background(), req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// 验证结果
	if resp.Answer == "" {
		t.Error("Answer should not be empty")
	}

	if len(resp.ExpandedChunks) == 0 {
		t.Error("Should have expanded chunks")
	}

	if resp.Metadata == nil {
		t.Error("Metadata should not be nil")
	}

	// 验证压缩效果
	expandedCount := resp.Metadata["expanded_count"].(int)
	compressedCount := resp.Metadata["compressed_count"].(int)
	
	t.Logf("Expanded chunks: %d", expandedCount)
	t.Logf("Compressed chunks: %d", compressedCount)
	t.Logf("Token reduction: %s", resp.Metadata["token_reduction"])
}

func TestREFRAGService_CompressChunks(t *testing.T) {
	service := &REFRAGService{}

	chunks := []*DocumentChunk{
		{
			ID:      1,
			Content: "这是一个测试文档，包含一些内容用于测试压缩功能。",
			Embedding: make([]float32, 768), // 模拟 768 维向量
		},
	}

	queryVector := make([]float32, 768)
	compressed := service.compressChunks(chunks, queryVector, 16)

	if len(compressed) != 1 {
		t.Fatalf("Expected 1 compressed chunk, got %d", len(compressed))
	}

	chunk := compressed[0]
	if chunk.CompressedTokenCount >= chunk.TokenCount {
		t.Error("Compressed token count should be less than original")
	}

	if len(chunk.CompressedVector) == 0 {
		t.Error("Compressed vector should not be empty")
	}
}

func TestREFRAGService_ScoreChunks(t *testing.T) {
	service := &REFRAGService{}

	chunks := []*CompressedChunk{
		{
			OriginalChunk: &DocumentChunk{
				Content: "Go 语言是一种开源的编程语言，由 Google 开发。",
			},
			Score: 0,
		},
		{
			OriginalChunk: &DocumentChunk{
				Content: "Python 是一种高级编程语言。",
			},
			Score: 0,
		},
	}

	question := "Go 语言的特点是什么？"
	service.scoreChunks(chunks, question)

	// 第一个 chunk 应该得分更高（包含 "Go"）
	if chunks[0].Score <= chunks[1].Score {
		t.Error("First chunk should have higher score")
	}
}

func TestREFRAGService_SelectAndExpand(t *testing.T) {
	service := &REFRAGService{}

	chunks := []*CompressedChunk{
		{Score: 0.9},
		{Score: 0.8},
		{Score: 0.7},
		{Score: 0.6},
		{Score: 0.5},
	}

	expanded, compressed := service.selectAndExpand(chunks, 2)

	if len(expanded) != 2 {
		t.Fatalf("Expected 2 expanded chunks, got %d", len(expanded))
	}

	if len(compressed) != 3 {
		t.Fatalf("Expected 3 compressed chunks, got %d", len(compressed))
	}

	// 验证已解压的 chunks
	for _, chunk := range expanded {
		if !chunk.IsExpanded {
			t.Error("Expanded chunks should be marked as expanded")
		}
	}

	// 验证保持压缩的 chunks
	for _, chunk := range compressed {
		if chunk.IsExpanded {
			t.Error("Compressed chunks should not be expanded")
		}
	}
}

// MockChunkRepository 模拟仓库（用于测试）
type MockChunkRepository struct {
	chunks []*DocumentChunk
}

func (m *MockChunkRepository) Create(chunk *DocumentChunk) error {
	return nil
}

func (m *MockChunkRepository) VectorSearch(queryVector []float32, docType, language string, limit int) ([]*DocumentChunk, error) {
	result := make([]*DocumentChunk, 0, limit)
	for i, chunk := range m.chunks {
		if i >= limit {
			break
		}
		if docType == "" || chunk.DocType == docType {
			if language == "" || chunk.Language == language {
				result = append(result, chunk)
			}
		}
	}
	return result, nil
}

func (m *MockChunkRepository) HybridSearch(queryVector []float32, keyword, docType, language string, limit int) ([]*DocumentChunk, error) {
	return m.VectorSearch(queryVector, docType, language, limit)
}

