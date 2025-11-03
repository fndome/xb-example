package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// AgenticRAGService 第三代 Agentic RAG 服务
// 核心能力：
// 1. 问题拆解（Decomposition）
// 2. 多轮召回（Multi-Round Retrieval）
// 3. 智能规划（Planning）
// 4. 结果综合（Synthesis）
type AgenticRAGService struct {
	baseRAG  *RAGService
	planner  *QueryPlanner
	executor *QueryExecutor
}

func NewAgenticRAGService(baseRAG *RAGService) *AgenticRAGService {
	return &AgenticRAGService{
		baseRAG:  baseRAG,
		planner:  NewQueryPlanner(baseRAG.llm),
		executor: NewQueryExecutor(baseRAG),
	}
}

// Query 第三代 Agentic RAG 查询
func (s *AgenticRAGService) Query(ctx context.Context, req RAGQueryRequest) (*RAGQueryResponse, error) {
	// === 阶段 1：问题分析与规划 ===
	plan, err := s.planner.Plan(ctx, req.Question)
	if err != nil {
		return nil, fmt.Errorf("planning failed: %w", err)
	}

	// 如果是简单问题，直接使用第一代 RAG
	if plan.IsSimple {
		return s.baseRAG.Query(ctx, req)
	}

	// === 阶段 2：多轮检索执行 ===
	results, err := s.executor.Execute(ctx, plan, req)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	// === 阶段 3：结果去重与重排 ===
	uniqueChunks := s.dedup(results.AllChunks)
	rerankedChunks := s.rerank(ctx, req.Question, uniqueChunks, s.getTopK(req))

	// === 阶段 4：综合生成答案 ===
	prompt := s.buildAgenticPrompt(req.Question, plan, results, rerankedChunks)
	answer, err := s.baseRAG.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	return &RAGQueryResponse{
		Answer:  answer,
		Sources: rerankedChunks,
		Metadata: map[string]interface{}{
			"mode":            "agentic_rag_v3",
			"is_simple":       plan.IsSimple,
			"question_type":   plan.QuestionType,
			"sub_queries":     plan.SubQueries,
			"total_retrieved": len(uniqueChunks),
			"final_selected":  len(rerankedChunks),
			"rounds":          results.Rounds,
		},
	}, nil
}

// getTopK 获取 TopK 参数
func (s *AgenticRAGService) getTopK(req RAGQueryRequest) int {
	if req.TopK != nil && *req.TopK > 0 {
		return *req.TopK
	}
	return 5
}

// dedup 去重（基于内容相似度）
func (s *AgenticRAGService) dedup(chunks []*DocumentChunk) []*DocumentChunk {
	if len(chunks) == 0 {
		return chunks
	}

	// 简单去重：基于 ID
	seen := make(map[int64]bool)
	unique := make([]*DocumentChunk, 0)

	for _, chunk := range chunks {
		if !seen[chunk.ID] {
			seen[chunk.ID] = true
			unique = append(unique, chunk)
		}
	}

	return unique
}

// rerank 重排序（基于相关性）
func (s *AgenticRAGService) rerank(ctx context.Context, question string, chunks []*DocumentChunk, topK int) []*DocumentChunk {
	// 简化版：保留前 topK 个
	// 实际应用中应该使用：
	// 1. LLM 重排（最准确但慢）
	// 2. BGE-Reranker（平衡性能和准确性）
	// 3. Cross-Encoder 模型

	if len(chunks) <= topK {
		return chunks
	}

	return chunks[:topK]
}

// buildAgenticPrompt 构建 Agentic RAG 提示词
func (s *AgenticRAGService) buildAgenticPrompt(
	question string,
	plan *QueryPlan,
	results *ExecutionResults,
	chunks []*DocumentChunk,
) string {
	var sb strings.Builder

	sb.WriteString("# 任务说明\n")
	sb.WriteString("你是一个专业的 RAG 助手，需要基于检索到的文档回答用户问题。\n\n")

	// 展示问题拆解过程
	if len(plan.SubQueries) > 0 {
		sb.WriteString("## 问题分析\n")
		sb.WriteString(fmt.Sprintf("原问题：%s\n", question))
		sb.WriteString(fmt.Sprintf("问题类型：%s\n", plan.QuestionType))
		sb.WriteString("\n已将问题拆解为以下子问题：\n")
		for i, subQ := range plan.SubQueries {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, subQ))
		}
		sb.WriteString("\n")
	}

	// 展示检索到的文档
	sb.WriteString("## 检索到的相关文档\n\n")
	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("### [文档 %d]\n", i+1))
		sb.WriteString(chunk.Content)
		sb.WriteString("\n\n")
	}

	// 回答要求
	sb.WriteString("## 回答要求\n")
	sb.WriteString(fmt.Sprintf("请基于上述文档回答原问题：%s\n\n", question))
	sb.WriteString("要求：\n")
	sb.WriteString("1. 回答应该全面、准确、有逻辑\n")
	sb.WriteString("2. 如果子问题的答案相关，请综合组织\n")
	sb.WriteString("3. 如果文档中没有足够信息，请明确说明\n")
	sb.WriteString("4. 回答应该自然流畅，不要生硬地罗列信息\n")

	return sb.String()
}

// ================== QueryPlanner ==================

// QueryPlanner 查询规划器
type QueryPlanner struct {
	llm LLMService
}

func NewQueryPlanner(llm LLMService) *QueryPlanner {
	return &QueryPlanner{llm: llm}
}

// Plan 生成查询计划
func (p *QueryPlanner) Plan(ctx context.Context, question string) (*QueryPlan, error) {
	prompt := p.buildPlanningPrompt(question)

	response, err := p.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return p.parsePlan(question, response)
}

// buildPlanningPrompt 构建规划提示词
func (p *QueryPlanner) buildPlanningPrompt(question string) string {
	return fmt.Sprintf(`你是一个查询规划专家，需要分析用户问题并生成检索计划。

用户问题：%s

请分析这个问题并输出 JSON 格式的规划：

{
  "is_simple": false,
  "question_type": "comparison|factual|reasoning|multi_aspect",
  "sub_queries": [
    "子问题1",
    "子问题2"
  ],
  "keywords": ["关键词1", "关键词2"],
  "reasoning": "为什么要这样拆解"
}

规则：
1. is_simple: 如果问题可以直接回答（如"什么是X"），设为 true
2. question_type: 
   - factual: 事实性问题（"什么是"、"谁发明了"）
   - comparison: 比较性问题（"X和Y的区别"）
   - reasoning: 推理性问题（"为什么"、"如何"）
   - multi_aspect: 多方面问题（"详细介绍X"）
3. sub_queries: 如果是复杂问题，拆解为 2-4 个子问题
4. keywords: 提取 2-5 个关键词用于辅助检索

只返回 JSON，不要有其他文字。`, question)
}

// parsePlan 解析规划结果
func (p *QueryPlanner) parsePlan(question, response string) (*QueryPlan, error) {
	// 提取 JSON（可能包含在 markdown 代码块中）
	jsonStr := extractJSON(response)

	var plan QueryPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		// 如果解析失败，返回简单规划
		return &QueryPlan{
			IsSimple:      true,
			QuestionType:  "factual",
			SubQueries:    []string{question},
			Keywords:      []string{},
			Reasoning:     "JSON parsing failed, fallback to simple mode",
		}, nil
	}

	return &plan, nil
}

// QueryPlan 查询计划
type QueryPlan struct {
	IsSimple     bool     `json:"is_simple"`
	QuestionType string   `json:"question_type"`
	SubQueries   []string `json:"sub_queries"`
	Keywords     []string `json:"keywords"`
	Reasoning    string   `json:"reasoning"`
}

// ================== QueryExecutor ==================

// QueryExecutor 查询执行器
type QueryExecutor struct {
	baseRAG *RAGService
}

func NewQueryExecutor(baseRAG *RAGService) *QueryExecutor {
	return &QueryExecutor{baseRAG: baseRAG}
}

// Execute 执行多轮检索
func (e *QueryExecutor) Execute(ctx context.Context, plan *QueryPlan, req RAGQueryRequest) (*ExecutionResults, error) {
	results := &ExecutionResults{
		AllChunks: make([]*DocumentChunk, 0),
		Rounds:    0,
	}

	// 每个子问题执行一轮检索
	for _, subQuery := range plan.SubQueries {
		chunks, err := e.executeRound(ctx, subQuery, req)
		if err != nil {
			// 单轮失败不影响整体
			continue
		}

		results.AllChunks = append(results.AllChunks, chunks...)
		results.Rounds++
	}

	return results, nil
}

// executeRound 执行单轮检索
func (e *QueryExecutor) executeRound(ctx context.Context, query string, req RAGQueryRequest) ([]*DocumentChunk, error) {
	// 1. 向量化子问题
	queryVector, err := e.baseRAG.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	// 2. 检索（每轮检索 3-5 个结果）
	roundTopK := 3
	if req.TopK != nil && *req.TopK > 0 {
		roundTopK = *req.TopK / 2
		if roundTopK < 3 {
			roundTopK = 3
		}
	}

	chunks, err := e.baseRAG.repo.VectorSearch(
		queryVector,
		req.DocType,
		req.Language,
		roundTopK,
	)

	return chunks, err
}

// ExecutionResults 执行结果
type ExecutionResults struct {
	AllChunks []*DocumentChunk
	Rounds    int
}

// ================== 辅助函数 ==================

// extractJSON 从文本中提取 JSON
func extractJSON(text string) string {
	// 移除 markdown 代码块标记
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	return text
}

