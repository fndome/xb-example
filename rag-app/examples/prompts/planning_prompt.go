package prompts

import "fmt"

// PlanningPrompt 问题规划提示词（Few-shot 学习）
func PlanningPrompt(question string) string {
	return fmt.Sprintf(`你是一个专业的查询规划专家，负责分析用户问题并生成检索计划。

# 任务说明
分析用户问题，判断问题类型，并将复杂问题拆解为多个子问题。

# 问题类型
- factual: 事实性问题（"什么是X"、"谁发明了Y"）
- comparison: 比较性问题（"X和Y的区别"、"比较A和B"）
- reasoning: 推理性问题（"为什么"、"如何"）
- multi_aspect: 多方面问题（"详细介绍X"、"全面分析Y"）

# Few-shot 示例

## 示例 1：简单问题
**用户问题**：什么是 Channel？

**输出**：
{
  "is_simple": true,
  "question_type": "factual",
  "sub_queries": [],
  "keywords": ["Channel"],
  "reasoning": "这是一个简单的事实性问题，可以直接检索回答"
}

## 示例 2：比较问题
**用户问题**：Go 和 Rust 在并发编程上有什么区别？

**输出**：
{
  "is_simple": false,
  "question_type": "comparison",
  "sub_queries": [
    "Go 在并发编程上的特点和机制",
    "Rust 在并发编程上的特点和机制",
    "Go 和 Rust 并发编程的主要区别"
  ],
  "keywords": ["Go", "Rust", "并发", "区别"],
  "reasoning": "这是一个比较性问题，需要分别了解两者的特点，然后进行对比"
}

## 示例 3：多方面问题
**用户问题**：详细介绍一下 xb 这个库

**输出**：
{
  "is_simple": false,
  "question_type": "multi_aspect",
  "sub_queries": [
    "xb 的核心功能是什么",
    "xb 的使用方法和 API",
    "xb 的优势和特点",
    "xb 的应用场景"
  ],
  "keywords": ["xb", "功能", "使用", "优势", "场景"],
  "reasoning": "这是一个多方面问题，需要从功能、使用、优势、场景等多个角度全面介绍"
}

## 示例 4：推理问题
**用户问题**：为什么 xb 比传统 ORM 更适合向量数据库？

**输出**：
{
  "is_simple": false,
  "question_type": "reasoning",
  "sub_queries": [
    "传统 ORM 的设计理念和局限",
    "向量数据库的特殊需求",
    "xb 的设计理念和创新点",
    "xb 如何解决向量数据库的痛点"
  ],
  "keywords": ["xb", "ORM", "向量数据库", "优势"],
  "reasoning": "这是一个推理性问题，需要分析 ORM 的局限、向量数据库的需求、xb 的设计，最后推理出原因"
}

# 现在轮到你了

**用户问题**：%s

请严格按照上述格式输出 JSON，不要有其他文字。`, question)
}

