package rerank

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CohereRerankClient Cohere Rerank 客户端
// Cohere 提供强大的 Rerank 模型
type CohereRerankClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// CohereConfig Cohere 配置
type CohereConfig struct {
	APIKey  string // Cohere API Key
	BaseURL string // 可选，默认 https://api.cohere.ai/v1
	Model   string // 默认 rerank-english-v3.0
}

func NewCohereRerankClient(config CohereConfig) *CohereRerankClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.cohere.ai/v1"
	}
	if config.Model == "" {
		config.Model = "rerank-english-v3.0" // 或 rerank-multilingual-v3.0 for 中文
	}

	return &CohereRerankClient{
		apiKey:  config.APIKey,
		baseURL: config.BaseURL,
		model:   config.Model,
		client:  &http.Client{},
	}
}

// RerankResult Rerank 结果
type RerankResult struct {
	Index          int     `json:"index"`           // 原始索引
	RelevanceScore float64 `json:"relevance_score"` // 相关性分数
	Document       string  `json:"-"`               // 文档内容
}

// Rerank 重排序
func (c *CohereRerankClient) Rerank(ctx context.Context, query string, documents []string, topK int) ([]RerankResult, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model":     c.model,
		"query":     query,
		"documents": documents,
		"top_n":     topK,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/rerank", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cohere api error (status %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result struct {
		Results []struct {
			Index          int     `json:"index"`
			RelevanceScore float64 `json:"relevance_score"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// 构建结果
	var results []RerankResult
	for _, r := range result.Results {
		if r.Index < len(documents) {
			results = append(results, RerankResult{
				Index:          r.Index,
				RelevanceScore: r.RelevanceScore,
				Document:       documents[r.Index],
			})
		}
	}

	return results, nil
}
