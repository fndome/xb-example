# xb 使用技巧（第四代多模态 RAG）

本文档总结在第四代多模态 RAG 中使用 xb 的关键技巧和最佳实践。

---

## ⭐ 核心要点

### 1. In() 方法使用可变参数

**错误示例** ❌：
```go
// 错误：直接传入切片
ids := []int64{1, 2, 3}
xb.Of(&Node{}).In("id", ids)  // ❌ 类型错误！
```

**正确示例** ✅：
```go
// 方案 1：展开切片
ids := []interface{}{int64(1), int64(2), int64(3)}
xb.Of(&Node{}).In("id", ids...)  // ✅ 使用 ... 展开

// 方案 2：直接传入参数
xb.Of(&Node{}).In("id", int64(1), int64(2), int64(3))  // ✅ 可变参数

// 方案 3：循环查询（如果 ID 很多）
for _, id := range ids {
    xb.Of(&Node{}).Eq("id", id)  // ✅ 逐个查询
}
```

**为什么？**
```go
// xb 的 In() 定义
func (x *BuilderX) In(k string, vs ...interface{}) *BuilderX
//                                 ^^^^^^^^^^^^^^ 可变参数，不是切片
```

---

## 🎯 多模态场景最佳实践

### 1. 存储多模态内容

```go
// ⭐ 使用 xb.Vector 类型存储向量
// ⭐ 重要：数值字段使用指针类型（xb 要求）
type ContentUnit struct {
    ID        int64       `db:"id"`
    DocID     *int64      `db:"doc_id"`     // ⭐ 指针类型（可选字段）
    Type      ContentType `db:"type"`
    Position  *int        `db:"position"`   // ⭐ 指针类型（可选字段）
    Embedding xb.Vector   `db:"embedding"`  // ⭐ 自动序列化
    ParentID  *int64      `db:"parent_id"`  // ⭐ 指针类型（可选字段）
    // ... 其他字段
}

// 插入
sql, args := xb.Of(&ContentUnit{}).
    Insert(func(ib *xb.InsertBuilder) {
        ib.Set("doc_id", unit.DocID).       // ⭐ 指针类型，可以是 nil
          Set("type", unit.Type).
          Set("position", unit.Position).   // ⭐ 指针类型，可以是 nil
          Set("embedding", unit.Embedding)   // ⭐ 无需手动转换
    }).
    Build().
    SqlOfInsert()

db.Exec(sql, args...)
```

**优势**：
- ✅ `xb.Vector` 实现了 `driver.Valuer` 和 `sql.Scanner`
- ✅ 自动转换为 pgvector 格式 `'[0.1, 0.2, ...]'`
- ✅ 从数据库读取时自动反序列化为 `[]float32`
- ✅ **指针类型正确处理 NULL 值**
- ✅ **可选字段语义清晰（nil = NULL）**

### 2. 向量检索 + 模态过滤

```go
// 基础向量检索
xb.Of(&ContentUnit{}).
    VectorSearch("embedding", queryVector, 10)

// + 单个模态过滤
xb.Of(&ContentUnit{}).
    VectorSearch("embedding", queryVector, 10).
    Eq("type", "image")  // ✅ 简单清晰

// + 多个模态（方案 1：分别查询后合并）
images := vectorSearchByType(queryVector, "image", 5)
tables := vectorSearchByType(queryVector, "table", 5)
combined := append(images, tables...)

// + 多个模态（方案 2：不过滤类型，在应用层筛选）
all := vectorSearch(queryVector, 20)  // Over-fetch
filtered := filterByTypes(all, []string{"image", "table"})
```

### 3. 混合检索（向量 + 标量）

```go
// ⭐ xb 的强项：灵活组合多个条件
sql, args := xb.Of(&ContentUnit{}).
    VectorSearch("embedding", queryVector, 20).
    Eq("doc_id", docID).              // 文档过滤
    Eq("type", "image").              // 模态过滤
    Gte("created_at", lastWeek).      // 时间过滤
    Like("content", "%关键词%").        // 文本过滤
    Build().
    SqlOfVectorSearch()
```

**生成的 SQL**：
```sql
SELECT *, embedding <-> $1 AS distance 
FROM content_units 
WHERE doc_id = $2 
  AND type = $3 
  AND created_at >= $4 
  AND content LIKE $5 
ORDER BY distance 
LIMIT 20
```

### 4. 知识图谱操作

#### 创建节点

```go
sql, args := xb.Of(&KnowledgeNode{}).
    Insert(func(ib *xb.InsertBuilder) {
        ib.Set("type", "entity").
          Set("name", "VAE模型").
          Set("content_id", node.ContentID).  // ⭐ 指针类型，可以是 nil
          Set("embedding", nodeEmbedding)      // ⭐ 节点也可以有向量
    }).
    Build().
    SqlOfInsert()
```

#### 创建边

```go
// ⭐ 重要：SourceID, TargetID, Weight 都是指针类型
sourceID := int64(1)
targetID := int64(2)
weight := 1.0

sql, args := xb.Of(&KnowledgeEdge{}).
    Insert(func(ib *xb.InsertBuilder) {
        ib.Set("source_id", &sourceID).     // ⭐ 指针类型
          Set("target_id", &targetID).      // ⭐ 指针类型
          Set("relation", "describes").
          Set("weight", &weight)            // ⭐ 指针类型
    }).
    Build().
    SqlOfInsert()

// 或者直接使用字段（如果已经是指针）
edge := &KnowledgeEdge{
    SourceID: ptr(int64(1)),   // 辅助函数创建指针
    TargetID: ptr(int64(2)),
    Weight:   ptr(1.0),
}
sql, args := xb.Of(&KnowledgeEdge{}).
    Insert(func(ib *xb.InsertBuilder) {
        ib.Set("source_id", edge.SourceID).
          Set("target_id", edge.TargetID).
          Set("weight", edge.Weight)
    }).
    Build().
    SqlOfInsert()
```

#### 查询邻居

```go
// 1. 查询出边
sql, args, _ := xb.Of(&KnowledgeEdge{}).
    Eq("source_id", nodeID).
    Build().
    SqlOfSelect()

var edges []*KnowledgeEdge
db.Select(&edges, sql, args...)

// 2. 提取目标 ID 并展开查询
targetIDs := make([]interface{}, len(edges))
for i, edge := range edges {
    targetIDs[i] = edge.TargetID
}

sql, args, _ = xb.Of(&KnowledgeNode{}).
    In("id", targetIDs...).  // ⭐ 使用 ... 展开
    Build().
    SqlOfSelect()
```

---

## 🚀 性能优化技巧

### 1. 批量插入

```go
tx, _ := db.Begin()
defer tx.Rollback()

for _, unit := range units {
    sql, args := xb.Of(&ContentUnit{}).
        Insert(func(ib *xb.InsertBuilder) {
            ib.Set("type", unit.Type).
              Set("embedding", unit.Embedding)
        }).
        Build().
        SqlOfInsert()
    
    tx.Exec(sql, args...)
}

tx.Commit()
```

**优势**：
- ✅ 事务保证原子性
- ✅ 减少网络往返
- ✅ PostgreSQL 自动批量优化

### 2. 向量索引

```sql
-- IVFFlat（适合中大规模）
CREATE INDEX idx_content_units_embedding 
ON content_units 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- HNSW（更快但占用内存）
CREATE INDEX idx_content_units_embedding_hnsw 
ON content_units 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
```

### 3. 复合索引

```sql
-- 常见查询模式：doc_id + type
CREATE INDEX idx_content_units_doc_type 
ON content_units (doc_id, type);
```

---

## 🔍 常见模式

### 模式 1：纯向量检索

```go
// 最简单的向量检索
sql, args := xb.Of(&ContentUnit{}).
    VectorSearch("embedding", queryVector, topK).
    Build().
    SqlOfVectorSearch()

var units []*ContentUnit
db.Select(&units, sql, args...)
```

### 模式 2：向量 + 单条件过滤

```go
// 向量 + 文档过滤
sql, args := xb.Of(&ContentUnit{}).
    VectorSearch("embedding", queryVector, topK).
    Eq("doc_id", docID).
    Build().
    SqlOfVectorSearch()
```

### 模式 3：向量 + 多条件过滤

```go
// 向量 + 多个标量条件
sql, args := xb.Of(&ContentUnit{}).
    VectorSearch("embedding", queryVector, topK).
    Eq("doc_id", docID).
    Eq("type", "image").
    Gte("created_at", startTime).
    Like("content", "%关键词%").
    Build().
    SqlOfVectorSearch()
```

### 模式 4：更新向量

```go
// 重新生成向量后更新
sql, args := xb.Of(&ContentUnit{}).
    Update(func(ub *xb.UpdateBuilder) {
        ub.Set("embedding", newEmbedding).
          Set("detailed_desc", newDesc)
    }).
    Eq("id", unitID).
    Build().
    SqlOfUpdate()

db.Exec(sql, args...)
```

### 模式 5：图遍历

```go
// 查询节点的邻居
sql, args, _ := xb.Of(&KnowledgeEdge{}).
    Eq("source_id", nodeID).
    Build().
    SqlOfSelect()

var edges []*KnowledgeEdge
db.Select(&edges, sql, args...)
```

---

## 🎨 高级技巧

### 1. 模态偏好实现

```go
// 方案 1：分别查询不同模态
func QueryWithModalityPreference(
    queryVector []float32,
    preferences map[ContentType]int,
) []*ContentUnit {
    var allUnits []*ContentUnit
    
    for contentType, count := range preferences {
        sql, args := xb.Of(&ContentUnit{}).
            VectorSearch("embedding", queryVector, count).
            Eq("type", string(contentType)).
            Build().
            SqlOfVectorSearch()
        
        var units []*ContentUnit
        db.Select(&units, sql, args...)
        allUnits = append(allUnits, units...)
    }
    
    return allUnits
}

// 使用
units := QueryWithModalityPreference(queryVector, map[ContentType]int{
    ContentTypeImage: 5,  // 要 5 张图片
    ContentTypeTable: 3,  // 要 3 个表格
    ContentTypeText:  2,  // 要 2 段文本
})
```

### 2. 动态条件构建

```go
func BuildDynamicQuery(
    queryVector []float32,
    filters map[string]interface{},
) (string, []interface{}) {
    builder := xb.Of(&ContentUnit{}).
        VectorSearch("embedding", queryVector, 20)
    
    // 动态添加条件
    if docID, ok := filters["doc_id"]; ok {
        builder = builder.Eq("doc_id", docID)
    }
    
    if contentType, ok := filters["type"]; ok {
        builder = builder.Eq("type", contentType)
    }
    
    if keyword, ok := filters["keyword"]; ok {
        builder = builder.Like("content", "%"+keyword.(string)+"%")
    }
    
    return builder.Build().SqlOfVectorSearch()
}
```

### 3. 分页检索

```go
// 第四代 RAG 中的分页向量检索
func VectorSearchWithPagination(
    queryVector []float32,
    page, pageSize int,
) []*ContentUnit {
    offset := (page - 1) * pageSize
    
    // xb 的 VectorSearch 自动处理 LIMIT
    sql, args := xb.Of(&ContentUnit{}).
        VectorSearch("embedding", queryVector, pageSize).
        Build().
        SqlOfVectorSearch()
    
    // 如需 OFFSET，可以手动添加到 SQL
    sql += fmt.Sprintf(" OFFSET %d", offset)
    
    var units []*ContentUnit
    db.Select(&units, sql, args...)
    return units
}
```

---

## 📊 API 对比

### xb vs 原生 SQL

| 操作 | 原生 SQL | xb | 优势 |
|------|---------|----|----|
| **向量检索** | 手动拼接 | `VectorSearch()` | ✅ 类型安全 |
| **条件过滤** | WHERE 子句 | `Eq()`, `Like()` | ✅ 链式调用 |
| **向量类型** | 手动转换 | `xb.Vector` | ✅ 自动序列化 |
| **批量查询** | IN 子句 | `In(...)`  | ✅ 可变参数 |
| **更新** | UPDATE SET | `Update(func)` | ✅ 函数式 |

---

## 🎯 第四代 RAG 中的 xb 优势

### 1. 统一的向量接口

**所有模态使用相同的 API**：

```go
// 文本
xb.Of(&ContentUnit{}).VectorSearch("embedding", vec, 10).Eq("type", "text")

// 图片
xb.Of(&ContentUnit{}).VectorSearch("embedding", vec, 10).Eq("type", "image")

// 表格
xb.Of(&ContentUnit{}).VectorSearch("embedding", vec, 10).Eq("type", "table")

// 公式
xb.Of(&ContentUnit{}).VectorSearch("embedding", vec, 10).Eq("type", "formula")
```

### 2. 灵活的标量过滤

**轻松组合多个条件**：

```go
xb.Of(&ContentUnit{}).
    VectorSearch("embedding", vec, 20).
    Eq("doc_id", docID).        // 文档
    Eq("type", "image").         // 模态
    Gte("created_at", time).     // 时间
    Like("content", keyword)     // 关键词
```

### 3. 图 + 向量的统一存储

**节点也可以有向量**：

```go
type KnowledgeNode struct {
    ID        int64     `db:"id"`
    Name      string    `db:"name"`
    Embedding xb.Vector `db:"embedding"`  // ⭐ 节点向量
}

// 语义搜索实体
xb.Of(&KnowledgeNode{}).
    VectorSearch("embedding", entityVector, 10).
    Eq("type", "entity")
```

### 4. 类型安全

```go
// ⭐ xb.Vector 提供类型安全
unit.Embedding = xb.Vector{0.1, 0.2, 0.3}  // ✅ 类型检查
unit.Embedding = "wrong"                    // ❌ 编译错误

// ⭐ 指针类型正确处理 NULL
unit.Position = ptr(1)        // ✅ 有值
unit.Position = nil           // ✅ NULL
unit.DocID = ptr(int64(100))  // ✅ 有值
unit.DocID = nil              // ✅ NULL

// 从数据库读取时自动转换
var unit ContentUnit
db.Get(&unit, "SELECT * FROM content_units WHERE id = ?", id)
// unit.Embedding 已经是 []float32 类型，可以直接使用
// unit.Position 是指针，nil = NULL，指针 = 有值
// unit.DocID 是指针，nil = NULL，指针 = 有值
```

---

## 💡 实用示例

### 示例 1：跨模态检索

```go
// 用户输入文本问题，想找图表
question := "2024年第一季度增长趋势"
queryVector, _ := embedder.Embed(ctx, question)

// 检索图片
sql, args := xb.Of(&ContentUnit{}).
    VectorSearch("embedding", queryVector, 5).
    Eq("type", "image").
    Build().
    SqlOfVectorSearch()

var images []*ContentUnit
db.Select(&images, sql, args...)

// 关键：图片的 Embedding 来自 AI 生成的 DetailedDesc
// 所以文本问题可以匹配到图片！
```

### 示例 2：文档内检索

```go
// 在特定文档中搜索相关内容
sql, args := xb.Of(&ContentUnit{}).
    VectorSearch("embedding", queryVector, 10).
    Eq("doc_id", docID).  // ⭐ 限定文档范围
    Build().
    SqlOfVectorSearch()

// 用途：
// - 文档内问答
// - 引用查找
// - 上下文检索
```

### 示例 3：时间范围检索

```go
// 查找最近一周的内容
lastWeek := time.Now().AddDate(0, 0, -7)

sql, args := xb.Of(&ContentUnit{}).
    VectorSearch("embedding", queryVector, 20).
    Gte("created_at", lastWeek).  // ⭐ 时间过滤
    Build().
    SqlOfVectorSearch()

// 用途：
// - 最新资讯检索
// - 时序分析
// - 变化追踪
```

### 示例 4：层次结构检索

```go
// 查找特定章节下的所有内容
sql, args := xb.Of(&ContentUnit{}).
    VectorSearch("embedding", queryVector, 10).
    Eq("parent_id", chapterID).  // ⭐ 层次过滤
    Build().
    SqlOfVectorSearch()

// 用途：
// - 章节内检索
// - 结构化导航
// - 上下文保留
```

---

## ⚠️ 常见陷阱

### 陷阱 0：忘记使用指针类型

```go
// ❌ 错误：直接使用值类型
type ContentUnit struct {
    Position  int      `db:"position"`    // 无法区分 NULL 和 0
    DocID     int64    `db:"doc_id"`      // 无法区分 NULL 和 0
    Weight    float64  `db:"weight"`      // 无法区分 NULL 和 0.0
}

// 问题：
// 1. 数据库中的 NULL 会被读取为 0
// 2. 无法判断字段是否真的为 0 还是不存在
// 3. xb 的条件构建可能不正确

// ✅ 正确：使用指针类型
type ContentUnit struct {
    Position  *int      `db:"position"`    // nil = NULL, 指针 = 值
    DocID     *int64    `db:"doc_id"`      // nil = NULL, 指针 = 值
    Weight    *float64  `db:"weight"`      // nil = NULL, 指针 = 值
}

// 优势：
// 1. nil 明确表示 NULL
// 2. 指针明确表示有值
// 3. xb 可以正确处理 WHERE 条件
```

**何时使用指针类型（xb 要求）**：
- ✅ **数值字段**（int, int64, float64 等）：必须是指针
- ✅ **布尔字段**（bool）：必须是指针
- ✅ **可选字段**（如 `doc_id`, `parent_id`）：必须是指针
- ✅ **外键字段**（如 `source_id`, `target_id`）：必须是指针
- ✅ **统计字段**（如 `total_units`, `text_units`）：必须是指针
- ❌ **主键**（如 `id BIGSERIAL PRIMARY KEY`）：可以是值类型

**规则总结**：
```go
type Model struct {
    // ❌ 主键：值类型
    ID int64 `db:"id"`
    
    // ✅ 其他数值：指针类型
    DocID    *int64   `db:"doc_id"`
    Position *int     `db:"position"`
    Count    *int     `db:"count"`
    Weight   *float64 `db:"weight"`
    
    // ✅ 布尔：指针类型
    IsActive *bool `db:"is_active"`
    IsPublic *bool `db:"is_public"`
    
    // ✅ 字符串：通常是值类型（除非可选）
    Name     string  `db:"name"`
    Content  string  `db:"content"`
    ImageURL *string `db:"image_url"`  // 可选则用指针
    
    // ✅ 时间：通常是值类型
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
    
    // ✅ 向量：xb.Vector（值类型）
    Embedding xb.Vector `db:"embedding"`
}
```

### 陷阱 1：In() 传入切片

```go
// ❌ 错误
ids := []int64{1, 2, 3}
xb.Of(&Node{}).In("id", ids)

// ✅ 正确
ids := []interface{}{int64(1), int64(2), int64(3)}
xb.Of(&Node{}).In("id", ids...)  // 使用 ... 展开
```

### 陷阱 2：向量维度不匹配

```go
// ❌ 错误：查询向量 512 维，数据库向量 768 维
queryVector := make([]float32, 512)  // ❌ 维度不对

// ✅ 正确：维度必须匹配
queryVector := make([]float32, 768)  // ✅ 与数据库一致
```

### 陷阱 3：忘记类型转换

```go
// ❌ 错误：直接用枚举类型
xb.Of(&ContentUnit{}).Eq("type", ContentTypeImage)  // ❌ 类型不对

// ✅ 正确：转换为字符串
xb.Of(&ContentUnit{}).Eq("type", string(ContentTypeImage))  // ✅
```

---

## 📚 参考资源

### xb 文档
- [xb GitHub](https://github.com/fndome/xb)
- [向量检索指南](../../xb/doc/ai_application/VECTOR_SEARCH.md)
- [混合检索](../../xb/doc/ai_application/HYBRID_SEARCH.md)

### 示例代码
- `example_test.go` - 10 个完整示例
- `multimodal_repository.go` - 生产级实现
- `sql/schema.sql` - 数据库 Schema

---

## 🎉 总结

### xb 在第四代 RAG 中的核心价值

1. **统一的向量接口**：文本、图片、表格、公式用相同 API
2. **灵活的组合查询**：向量 + 标量条件轻松组合
3. **类型安全**：`xb.Vector` 自动处理序列化
4. **性能优越**：自动生成优化 SQL
5. **易于扩展**：添加新模态只需扩展数据模型

### 最佳实践

- ✅ **数值字段使用指针类型**（Position, DocID, Weight, SourceID, TargetID 等）
- ✅ 使用 `xb.Vector` 存储向量
- ✅ `In()` 方法使用可变参数展开 `...`
- ✅ 枚举类型转换为字符串
- ✅ 批量操作使用事务
- ✅ 创建合适的索引

---

**xb - 让多模态 RAG 开发变得简单！** 🚀💎
