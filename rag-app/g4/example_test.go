package g4

import (
	"testing"
	"time"

	"github.com/fndome/xb"
)

// TestBasicVectorSearch 示例 1：基础向量检索
// 展示：最简单的向量检索用法
func TestBasicVectorSearch(t *testing.T) {
	t.Log("=== 示例 1：基础向量检索 ===\n")

	// 模拟场景：搜索与"机器学习"相关的内容
	queryVector := generateMockEmbedding("机器学习")

	// ⭐ 使用 xb 进行向量检索
	sql, args := xb.Of(&ContentUnit{}).
		VectorSearch("embedding", queryVector, 5).
		Build().
		SqlOfVectorSearch()

	t.Logf("生成的 SQL:\n%s\n", sql)
	t.Logf("参数数量: %d\n\n", len(args))

	// 实际使用时：
	// var units []*ContentUnit
	// err := db.Select(&units, sql, args...)
}

// TestMultimodalSearch 示例 2：多模态检索
// 展示：向量检索 + 模态过滤
func TestMultimodalSearch(t *testing.T) {
	t.Log("=== 示例 2：多模态检索（图片） ===\n")

	// 模拟场景：只搜索图片
	queryVector := generateMockEmbedding("销售趋势图表")

	// ⭐ 向量检索 + 模态过滤
	sql, args := xb.Of(&ContentUnit{}).
		VectorSearch("embedding", queryVector, 10).
		Eq("type", string(ContentTypeImage)). // 只要图片
		Build().
		SqlOfVectorSearch()

	t.Logf("生成的 SQL:\n%s\n", sql)
	t.Logf("参数数量: %d\n", len(args))
	t.Logf("过滤模态: 图片\n")
	t.Log("说明：如需多个类型，可以分别查询后合并，或使用原生 SQL 的 IN 子句\n\n")
	_ = args // 避免未使用警告
}

// TestHybridSearch 示例 3：混合检索
// 展示：向量 + 多条件过滤
func TestHybridSearch(t *testing.T) {
	t.Log("=== 示例 3：混合检索（向量 + 标量）===\n")

	// 模拟场景：搜索特定文档中的最近一周的图表
	queryVector := generateMockEmbedding("2024年第一季度财报图表")
	docID := int64(100)
	lastWeek := time.Now().AddDate(0, 0, -7)

	// ⭐ 复杂的混合检索
	sql, args := xb.Of(&ContentUnit{}).
		VectorSearch("embedding", queryVector, 20).
		Eq("doc_id", docID).                  // 文档过滤
		Eq("type", string(ContentTypeImage)). // 模态过滤
		Gte("created_at", lastWeek).          // 时间过滤
		Like("content", "%图表%").              // 关键词过滤
		Build().
		SqlOfVectorSearch()

	t.Logf("生成的 SQL:\n%s\n", sql)
	t.Logf("过滤条件:\n")
	t.Logf("  - 文档ID: %d\n", docID)
	t.Logf("  - 类型: 图片\n")
	t.Logf("  - 时间: 最近一周\n")
	t.Logf("  - 关键词: 图表\n\n")
	_ = args // 避免未使用警告
}

// TestKnowledgeGraphInsert 示例 4：知识图谱构建
// 展示：图节点和边的创建
func TestKnowledgeGraphInsert(t *testing.T) {
	t.Log("=== 示例 4：知识图谱构建 ===\n")

	// 场景：构建一个简单的知识图谱
	// 实体："VAE模型" -> describes -> 内容单元（图片）

	// 1. 创建实体节点
	entityNode := &KnowledgeNode{
		Type:      NodeTypeEntity,
		Name:      "VAE模型",
		Embedding: generateMockEmbedding("VAE模型"),
	}

	sql1, args1 := xb.Of(&KnowledgeNode{}).
		Insert(func(ib *xb.InsertBuilder) {
			ib.Set("type", entityNode.Type).
				Set("name", entityNode.Name).
				Set("embedding", entityNode.Embedding)
		}).
		Build().
		SqlOfInsert()

	t.Logf("1. 创建实体节点 SQL:\n%s\n", sql1)

	// 2. 创建内容单元节点
	contentNode := &KnowledgeNode{
		Type:      NodeTypeContentUnit,
		Name:      "图5：VAE架构示意图",
		ContentID: ptr(int64(200)),
		Embedding: generateMockEmbedding("VAE架构示意图"),
	}

	sql2, args2 := xb.Of(&KnowledgeNode{}).
		Insert(func(ib *xb.InsertBuilder) {
			ib.Set("type", contentNode.Type).
				Set("name", contentNode.Name).
				Set("content_id", contentNode.ContentID).
				Set("embedding", contentNode.Embedding)
		}).
		Build().
		SqlOfInsert()

	t.Logf("2. 创建内容节点 SQL:\n%s\n", sql2)

	// 3. 创建边：实体 -> 内容
	sourceID := int64(1)
	targetID := int64(2)
	weight := 1.0
	edge := &KnowledgeEdge{
		SourceID: &sourceID, // ⭐ 指针类型
		TargetID: &targetID, // ⭐ 指针类型
		Relation: "describes",
		Weight:   &weight, // ⭐ 指针类型
	}

	sql3, args3 := xb.Of(&KnowledgeEdge{}).
		Insert(func(ib *xb.InsertBuilder) {
			ib.Set("source_id", edge.SourceID).
				Set("target_id", edge.TargetID).
				Set("relation", edge.Relation).
				Set("weight", edge.Weight)
		}).
		Build().
		SqlOfInsert()

	t.Logf("3. 创建边 SQL:\n%s\n\n", sql3)

	_ = args1
	_ = args2
	_ = args3
}

// TestGraphTraversal 示例 5：图遍历
// 展示：查询节点的邻居
func TestGraphTraversal(t *testing.T) {
	t.Log("=== 示例 5：图遍历查询 ===\n")

	// 场景：查找"VAE模型"实体的所有关联内容
	nodeID := int64(1)

	// 1. 查询所有出边
	sql1, args1, _ := xb.Of(&KnowledgeEdge{}).
		Eq("source_id", nodeID).
		Build().
		SqlOfSelect()

	t.Logf("1. 查询出边 SQL:\n%s\n", sql1)
	t.Logf("   查找从节点 %d 出发的所有边\n\n", nodeID)

	// 2. 查询单个目标节点（示例）
	targetID := int64(2)

	sql2, args2, _ := xb.Of(&KnowledgeNode{}).
		Eq("id", targetID).
		Build().
		SqlOfSelect()

	t.Logf("2. 查询目标节点 SQL:\n%s\n", sql2)
	t.Logf("   获取节点 %d 的详细信息\n", targetID)
	t.Log("   说明：实际应用中可以循环查询多个节点，或使用原生 SQL 的 IN\n\n")

	_ = args1
	_ = args2
}

// TestBatchInsert 示例 6：批量插入
// 展示：高效的批量操作
func TestBatchInsert(t *testing.T) {
	t.Log("=== 示例 6：批量插入优化 ===\n")

	// 场景：批量插入 100 个内容单元
	units := make([]*ContentUnit, 100)
	for i := 0; i < 100; i++ {
		position := i
		units[i] = &ContentUnit{
			DocID:     ptr(int64(1)),
			Type:      ContentTypeText,
			Position:  &position, // ⭐ 指针类型
			Content:   "内容" + string(rune(i)),
			Embedding: generateMockEmbedding("内容"),
		}
	}

	t.Logf("批量插入 %d 个内容单元\n", len(units))
	t.Log("推荐方案：使用事务 + 循环插入\n\n")

	// 示例代码：
	// tx, _ := db.Begin()
	// for _, unit := range units {
	//     sql, args := xb.Of(&ContentUnit{}).Insert(...).Build().SqlOfInsert()
	//     tx.Exec(sql, args...)
	// }
	// tx.Commit()

	t.Log("优势：")
	t.Log("  - 事务保证原子性")
	t.Log("  - 减少网络往返")
	t.Log("  - PostgreSQL 自动批量优化\n\n")
}

// TestUpdateVector 示例 7：向量更新
// 展示：更新已有内容的向量
func TestUpdateVector(t *testing.T) {
	t.Log("=== 示例 7：向量更新 ===\n")

	// 场景：重新生成图片的 Embedding
	unitID := int64(200)
	newEmbedding := generateMockEmbedding("更新后的图片描述")

	// ⭐ 使用 xb 更新向量
	sql, args := xb.Of(&ContentUnit{}).
		Update(func(ub *xb.UpdateBuilder) {
			ub.Set("embedding", newEmbedding).
				Set("detailed_desc", "更新后的详细描述")
		}).
		Eq("id", unitID).
		Build().
		SqlOfUpdate()

	t.Logf("生成的 SQL:\n%s\n", sql)
	t.Logf("更新内容单元 %d 的向量\n\n", unitID)
	_ = args // 避免未使用警告
}

// TestCompleteWorkflow 示例 8：完整工作流
// 展示：从 PDF 解析到检索的完整流程
func TestCompleteWorkflow(t *testing.T) {
	t.Log("=== 示例 8：完整的多模态 RAG 工作流 ===\n")

	t.Log("步骤 1: 解析 PDF")
	t.Log("  - 提取文本块、图片、表格")
	t.Log("  - 使用 MLLM 生成描述\n")

	t.Log("步骤 2: 生成 Embedding")
	t.Log("  - 文本：OpenAI text-embedding-3-small")
	t.Log("  - 图片：CLIP 或 AI 描述\n")

	t.Log("步骤 3: 存储到数据库（使用 xb）")
	sql1, _ := xb.Of(&ContentUnit{}).
		Insert(func(ib *xb.InsertBuilder) {
			ib.Set("type", ContentTypeImage).
				Set("embedding", generateMockEmbedding("图片"))
		}).
		Build().
		SqlOfInsert()
	if len(sql1) > 50 {
		t.Logf("  SQL: %s...\n\n", sql1[:50])
	} else {
		t.Logf("  SQL: %s\n\n", sql1)
	}

	t.Log("步骤 4: 构建知识图谱")
	t.Log("  - 提取实体和关系")
	t.Log("  - 创建节点和边\n")

	t.Log("步骤 5: 混合检索（使用 xb）")
	sql2, _ := xb.Of(&ContentUnit{}).
		VectorSearch("embedding", generateMockEmbedding("查询"), 10).
		Eq("type", string(ContentTypeImage)).
		Build().
		SqlOfVectorSearch()
	if len(sql2) > 50 {
		t.Logf("  SQL: %s...\n\n", sql2[:50])
	} else {
		t.Logf("  SQL: %s\n\n", sql2)
	}

	t.Log("步骤 6: 生成答案")
	t.Log("  - LLM 综合生成")
	t.Log("  - 返回文本 + 图片/表格引用\n")
}

// TestModalityPreference 示例 9：模态偏好
// 展示：用户偏好特定类型的内容
func TestModalityPreference(t *testing.T) {
	t.Log("=== 示例 9：模态偏好检索 ===\n")

	// 场景：用户想看图表，优先返回图片和表格
	queryVector := generateMockEmbedding("2024年销售数据")

	// 1. 先检索所有类型
	sql, _ := xb.Of(&ContentUnit{}).
		VectorSearch("embedding", queryVector, 20).
		Build().
		SqlOfVectorSearch()

	t.Logf("1. 检索所有类型（Top 20）:\n%s...\n\n", sql[:50])

	// 2. 在应用层应用模态偏好
	t.Log("2. 应用模态偏好:")
	t.Log("   - 图片：权重 2.0（优先）")
	t.Log("   - 表格：权重 2.0（优先）")
	t.Log("   - 文本：权重 1.0（正常）\n")

	t.Log("3. 重排结果:")
	t.Log("   [图片1, 图片2, 表格1, 文本1, ...]\n\n")
}

// TestCrossModalRetrieval 示例 10：跨模态检索
// 展示：文本查询检索图片
func TestCrossModalRetrieval(t *testing.T) {
	t.Log("=== 示例 10：跨模态检索 ===\n")

	// 场景：用户输入文本问题，想找到相关的图表
	question := "显示2024年第一季度增长趋势的图表"
	queryVector := generateMockEmbedding(question)

	t.Logf("用户问题: \"%s\"\n\n", question)

	// 检索图片（如需检索多种类型，可以分别查询后合并）
	sql, _ := xb.Of(&ContentUnit{}).
		VectorSearch("embedding", queryVector, 5).
		Eq("type", string(ContentTypeImage)).
		Build().
		SqlOfVectorSearch()

	t.Logf("检索策略:\n")
	t.Log("  1. 将文本问题向量化")
	t.Log("  2. 在图片/表格的向量空间中搜索")
	t.Log("  3. 利用 AI 生成的描述（DetailedDesc）\n")

	if len(sql) > 50 {
		t.Logf("生成的 SQL:\n%s...\n\n", sql[:50])
	} else {
		t.Logf("生成的 SQL:\n%s\n\n", sql)
	}

	t.Log("关键点：")
	t.Log("  ✅ xb 的向量检索不区分模态")
	t.Log("  ✅ 通过 AI 描述实现跨模态理解")
	t.Log("  ✅ 统一的向量空间让跨模态检索成为可能\n\n")
}

// ==================== 辅助函数 ====================

// generateMockEmbedding 生成模拟向量
func generateMockEmbedding(text string) []float32 {
	// 实际应用中应该调用真实的 Embedding API
	vec := make([]float32, 768) // OpenAI text-embedding-3-small 的维度
	for i := range vec {
		vec[i] = 0.1
	}
	return vec
}

// ptr 辅助函数：创建指针
func ptr[T any](v T) *T {
	return &v
}

// toInterfaceSlice 转换为 interface{} 切片（xb In() 方法需要）
func toInterfaceSlice[T any](slice []T) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}

// ==================== 性能基准测试 ====================

func BenchmarkVectorSearch(b *testing.B) {
	queryVector := generateMockEmbedding("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sql, args := xb.Of(&ContentUnit{}).
			VectorSearch("embedding", queryVector, 10).
			Build().
			SqlOfVectorSearch()

		_ = sql
		_ = args
	}
}

func BenchmarkHybridSearch(b *testing.B) {
	queryVector := generateMockEmbedding("benchmark")
	docID := int64(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sql, args := xb.Of(&ContentUnit{}).
			VectorSearch("embedding", queryVector, 10).
			Eq("doc_id", docID).
			Eq("type", string(ContentTypeImage)).
			Build().
			SqlOfVectorSearch()

		_ = sql
		_ = args
	}
}
