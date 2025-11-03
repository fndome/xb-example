package main

import (
	"context"
	"testing"
)

func TestAgenticRAG_SimpleQuestion(t *testing.T) {
	// Mock 服务
	embedder := &MockEmbeddingService{}
	llm := &MockLLMService{}
	repo := &MockChunkRepositoryImpl{}

	ragService := NewRAGService(repo, embedder, llm)
	agenticService := NewAgenticRAGService(ragService)

	// 简单问题应该直接回退到第一代 RAG
	req := RAGQueryRequest{
		Question: "什么是 Go 语言？",
	}

	resp, err := agenticService.Query(context.Background(), req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if resp.Answer == "" {
		t.Error("Answer should not be empty")
	}

	t.Logf("Answer: %s", resp.Answer)
	t.Logf("Metadata: %+v", resp.Metadata)
}

func TestAgenticRAG_ComplexQuestion(t *testing.T) {
	// Mock 服务
	embedder := &MockEmbeddingService{}
	llm := &MockLLMService{}
	repo := &MockChunkRepositoryImpl{}

	ragService := NewRAGService(repo, embedder, llm)
	agenticService := NewAgenticRAGService(ragService)

	// 复杂问题应该触发 Agentic RAG
	req := RAGQueryRequest{
		Question: "Go 和 Rust 在并发编程上有什么区别？各自的优势是什么？",
	}

	resp, err := agenticService.Query(context.Background(), req)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if resp.Answer == "" {
		t.Error("Answer should not be empty")
	}

	// 检查 metadata
	metadata := resp.Metadata
	if metadata["mode"] != "agentic_rag_v3" {
		t.Error("Should use agentic_rag_v3 mode")
	}

	t.Logf("Answer: %s", resp.Answer)
	t.Logf("Metadata: %+v", metadata)
	t.Logf("Sub-queries: %+v", metadata["sub_queries"])
}

func TestQueryPlanner_SimpleQuestion(t *testing.T) {
	llm := &MockLLMService{}
	planner := NewQueryPlanner(llm)

	plan, err := planner.Plan(context.Background(), "什么是 Channel？")
	if err != nil {
		t.Fatalf("Planning failed: %v", err)
	}

	// Mock LLM 会返回 is_simple=true
	if !plan.IsSimple {
		t.Logf("Plan: %+v", plan)
	}

	t.Logf("Question Type: %s", plan.QuestionType)
	t.Logf("Sub Queries: %+v", plan.SubQueries)
}

func TestQueryPlanner_ComplexQuestion(t *testing.T) {
	llm := &MockLLMService{}
	planner := NewQueryPlanner(llm)

	plan, err := planner.Plan(context.Background(), "Go 和 Rust 的区别是什么？")
	if err != nil {
		t.Fatalf("Planning failed: %v", err)
	}

	t.Logf("Is Simple: %v", plan.IsSimple)
	t.Logf("Question Type: %s", plan.QuestionType)
	t.Logf("Sub Queries: %+v", plan.SubQueries)
	t.Logf("Keywords: %+v", plan.Keywords)
}

// MockChunkRepositoryImpl 用于测试
type MockChunkRepositoryImpl struct {
}

func (r *MockChunkRepositoryImpl) Create(chunk *DocumentChunk) error {
	return nil
}

func (r *MockChunkRepositoryImpl) VectorSearch(queryVector []float32, docType, language string, limit int) ([]*DocumentChunk, error) {
	// 返回模拟数据
	return []*DocumentChunk{
		{
			ID:      1,
			Content: "Go 语言是 Google 开发的编程语言，以并发编程见长。",
		},
		{
			ID:      2,
			Content: "Goroutine 是 Go 语言的轻量级线程，可以轻松创建数百万个并发任务。",
		},
		{
			ID:      3,
			Content: "Channel 是 Go 语言中用于 Goroutine 之间通信的机制。",
		},
	}, nil
}

func (r *MockChunkRepositoryImpl) HybridSearch(queryVector []float32, keyword, docType, language string, limit int) ([]*DocumentChunk, error) {
	return r.VectorSearch(queryVector, docType, language, limit)
}

