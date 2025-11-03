package prompts

import (
	"fmt"
	"strings"
)

// DocumentChunk 文档分块（为了避免循环依赖，这里重新定义）
type DocumentChunk struct {
	ID      int64
	Content string
}

// QueryPlan 查询计划
type QueryPlan struct {
	IsSimple     bool
	QuestionType string
	SubQueries   []string
	Keywords     []string
	Reasoning    string
}

// GenerationPrompt RAG 生成提示词（Few-shot 学习）
func GenerationPrompt(question string, plan *QueryPlan, chunks []DocumentChunk) string {
	var sb strings.Builder

	// === 系统角色 ===
	sb.WriteString("# 角色\n")
	sb.WriteString("你是一个专业的 RAG 助手，擅长基于检索到的文档回答用户问题。\n\n")

	// === Few-shot 示例 ===
	sb.WriteString("# Few-shot 示例\n\n")

	sb.WriteString("## 示例 1：简单事实问题\n")
	sb.WriteString("**问题**：什么是 Channel？\n")
	sb.WriteString("**检索到的文档**：\n")
	sb.WriteString("[文档 1] Channel 是 Go 语言中用于 Goroutine 之间通信的机制。\n")
	sb.WriteString("[文档 2] Channel 可以是有缓冲或无缓冲的。\n\n")
	sb.WriteString("**回答**：\n")
	sb.WriteString("Channel 是 Go 语言中用于 Goroutine 之间通信的核心机制。它可以分为两种类型：\n\n")
	sb.WriteString("1. **无缓冲 Channel**：发送和接收操作会阻塞，直到另一端准备好\n")
	sb.WriteString("2. **有缓冲 Channel**：在缓冲区未满时，发送操作不会阻塞\n\n")
	sb.WriteString("Channel 保证了 Goroutine 之间的线程安全通信。\n\n")

	sb.WriteString("## 示例 2：比较问题\n")
	sb.WriteString("**问题**：Go 和 Rust 在并发编程上有什么区别？\n")
	sb.WriteString("**子问题**：\n")
	sb.WriteString("1. Go 在并发编程上的特点\n")
	sb.WriteString("2. Rust 在并发编程上的特点\n")
	sb.WriteString("3. Go 和 Rust 并发编程的主要区别\n\n")
	sb.WriteString("**检索到的文档**：\n")
	sb.WriteString("[文档 1] Go 使用 Goroutine 和 Channel 实现并发，语法简单。\n")
	sb.WriteString("[文档 2] Rust 通过所有权系统在编译时保证线程安全。\n")
	sb.WriteString("[文档 3] Go 的并发模型基于 CSP 理论。\n\n")
	sb.WriteString("**回答**：\n")
	sb.WriteString("Go 和 Rust 在并发编程上有显著的设计理念和实现方式的差异：\n\n")
	sb.WriteString("**Go 的并发特点**：\n")
	sb.WriteString("- 基于 CSP（Communicating Sequential Processes）理论\n")
	sb.WriteString("- 使用 Goroutine（轻量级线程）和 Channel（通信机制）\n")
	sb.WriteString("- 语法简单，易于使用：`go func()` 即可启动并发\n")
	sb.WriteString("- 运行时调度，GC 管理内存\n\n")
	sb.WriteString("**Rust 的并发特点**：\n")
	sb.WriteString("- 基于所有权（Ownership）和借用（Borrowing）系统\n")
	sb.WriteString("- 在编译时保证线程安全（\"Fearless Concurrency\"）\n")
	sb.WriteString("- 无运行时开销，性能接近 C++\n")
	sb.WriteString("- 学习曲线较陡\n\n")
	sb.WriteString("**主要区别**：\n")
	sb.WriteString("1. **安全保证时机**：Rust 在编译时保证，Go 在运行时检查\n")
	sb.WriteString("2. **易用性**：Go 更简单，Rust 更严格\n")
	sb.WriteString("3. **性能**：Rust 无 GC，性能更高；Go 有 GC，但更易用\n")
	sb.WriteString("4. **适用场景**：Go 适合快速开发高并发服务，Rust 适合系统编程和性能敏感场景\n\n")

	// === 当前任务 ===
	sb.WriteString("# 现在轮到你了\n\n")

	// 展示问题分析
	if plan != nil && !plan.IsSimple && len(plan.SubQueries) > 0 {
		sb.WriteString("## 问题分析\n")
		sb.WriteString(fmt.Sprintf("**原问题**：%s\n", question))
		sb.WriteString(fmt.Sprintf("**问题类型**：%s\n", plan.QuestionType))
		sb.WriteString("\n**已将问题拆解为以下子问题**：\n")
		for i, subQ := range plan.SubQueries {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, subQ))
		}
		sb.WriteString("\n")
	}

	// 展示检索到的文档
	sb.WriteString("## 检索到的文档\n\n")
	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("**[文档 %d]**\n", i+1))
		sb.WriteString(chunk.Content)
		sb.WriteString("\n\n")
	}

	// 回答要求
	sb.WriteString("## 回答要求\n")
	sb.WriteString(fmt.Sprintf("请基于上述检索到的文档回答问题：**%s**\n\n", question))
	sb.WriteString("要求：\n")
	sb.WriteString("1. **准确性**：回答应基于文档内容，不要编造信息\n")
	sb.WriteString("2. **完整性**：如果有多个子问题，请综合回答\n")
	sb.WriteString("3. **结构化**：使用标题、列表等格式，使回答清晰易读\n")
	sb.WriteString("4. **引用文档**：重要观点可以引用'根据文档X'\n")
	sb.WriteString("5. **坦诚不足**：如果文档中没有足够信息，请明确说明\n\n")
	sb.WriteString("请开始回答：\n")

	return sb.String()
}

// SimpleGenerationPrompt 简单问题的生成提示词
func SimpleGenerationPrompt(question string, chunks []DocumentChunk) string {
	var sb strings.Builder

	sb.WriteString("请根据以下文档内容回答问题。\n\n")
	sb.WriteString("**检索到的文档**：\n\n")

	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("[文档 %d] %s\n\n", i+1, chunk.Content))
	}

	sb.WriteString(fmt.Sprintf("**问题**：%s\n\n", question))
	sb.WriteString("请基于上述文档内容进行回答。如果文档中没有相关信息，请明确说明。\n")

	return sb.String()
}
