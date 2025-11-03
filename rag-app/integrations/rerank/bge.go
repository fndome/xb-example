package rerank

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// BGERerankClient BGE-Reranker 客户端
// 可以使用本地部署或云端服务
type BGERerankClient struct {
	baseURL string
	client  *http.Client
}

// BGEConfig BGE 配置
type BGEConfig struct {
	BaseURL string // BGE 服务 URL（本地或云端）
}

func NewBGERerankClient(config BGEConfig) *BGERerankClient {
	if config.BaseURL == "" {
		// 默认本地服务
		config.BaseURL = "http://localhost:8000"
	}

	return &BGERerankClient{
		baseURL: config.BaseURL,
		client:  &http.Client{},
	}
}

// Rerank 重排序
func (c *BGERerankClient) Rerank(ctx context.Context, query string, documents []string, topK int) ([]RerankResult, error) {
	// 构建请求体（兼容 FastAPI 格式）
	requestBody := map[string]interface{}{
		"query":     query,
		"documents": documents,
		"top_k":     topK,
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
		return nil, fmt.Errorf("bge api error (status %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result struct {
		Results []struct {
			Index int     `json:"index"`
			Score float64 `json:"score"`
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
				RelevanceScore: r.Score,
				Document:       documents[r.Index],
			})
		}
	}

	return results, nil
}
