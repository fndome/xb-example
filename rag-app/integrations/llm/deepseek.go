package llm

import (
	"context"
	"fmt"
)

// DeepSeekClient DeepSeek LLM 客户端
// DeepSeek API 完全兼容 OpenAI 接口
type DeepSeekClient struct {
	*OpenAIClient
}

// DeepSeekConfig DeepSeek 配置
type DeepSeekConfig struct {
	APIKey string // DeepSeek API Key
	Model  string // 默认 deepseek-chat
}

func NewDeepSeekClient(config DeepSeekConfig) *DeepSeekClient {
	if config.Model == "" {
		config.Model = "deepseek-chat" // DeepSeek V3
	}

	// DeepSeek 使用自己的 API 端点
	openaiConfig := OpenAIConfig{
		APIKey:  config.APIKey,
		BaseURL: "https://api.deepseek.com/v1",
		Model:   config.Model,
	}

	return &DeepSeekClient{
		OpenAIClient: NewOpenAIClient(openaiConfig),
	}
}

// Generate 生成文本
func (c *DeepSeekClient) Generate(ctx context.Context, prompt string) (string, error) {
	return c.OpenAIClient.Generate(ctx, prompt)
}

// GenerateWithOptions 生成文本（带选项）
func (c *DeepSeekClient) GenerateWithOptions(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	return c.OpenAIClient.GenerateWithOptions(ctx, prompt, options)
}

// Embed DeepSeek 暂不支持 Embedding，建议使用 OpenAI 或其他 Embedding 服务
func (c *DeepSeekClient) Embed(ctx context.Context, text string) ([]float32, error) {
	return nil, fmt.Errorf("deepseek does not support embedding yet, please use openai or other embedding services")
}

// DescribeImage DeepSeek V2.5 支持多模态
func (c *DeepSeekClient) DescribeImage(ctx context.Context, imageURL string, prompt string) (string, error) {
	if prompt == "" {
		prompt = "请详细描述这张图片的内容，包括关键元素、结构和重要信息。"
	}

	// DeepSeek 多模态接口与 OpenAI 兼容
	return c.OpenAIClient.DescribeImage(ctx, imageURL, prompt)
}

