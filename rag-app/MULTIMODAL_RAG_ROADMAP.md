# 第四代多模态 RAG 升级路线图

## 🎯 目标：从纯文本到全模态

基于 RAG-Anything 论文的启发，将 `rag-app` 从第三代 Agentic RAG 升级到**第四代多模态 RAG**。

---

## 📊 当前状态 vs RAG-Anything

| 特性 | 第三代 Agentic RAG | RAG-Anything | 差距 |
|------|-------------------|--------------|------|
| **问题拆解** | ✅ LLM 规划 | ✅ 查询分析 | - |
| **多轮召回** | ✅ 多轮检索 | ✅ 混合检索 | - |
| **模态支持** | ❌ 纯文本 | ✅ 图像/表格/公式 | **大** |
| **知识表示** | ❌ 向量 only | ✅ 双图谱 | **大** |
| **结构理解** | ❌ 线性文本 | ✅ 层次结构 | **大** |
| **检索策略** | ✅ 向量搜索 | ✅ 结构导航 + 向量 | **中** |

---

## 🛣️ 三阶段升级路径

### 阶段 1：多模态内容解析（1-2 周）

#### 目标
支持 PDF、图片、表格的解析和存储

#### 技术选型

**1. PDF 解析**
```go
// 使用 unidoc (UniPDF) - 商业友好的 Go PDF 库
import "github.com/unidoc/unipdf/v3"

// 或使用 pdfcpu - 开源方案
import "github.com/pdfcpu/pdfcpu"
```

**2. 图像理解**
```go
// 集成多模态 LLM (OpenAI GPT-4V, Claude 3, DeepSeek V2)
type MultimodalLLM interface {
    DescribeImage(ctx context.Context, imageData []byte) (string, error)
    ExtractTableData(ctx context.Context, imageData []byte) (*TableData, error)
    AnalyzeFormula(ctx context.Context, imageData []byte) (string, error)
}
```

**3. 表格解析**
```go
// 使用 excelize - Excel 解析
import "github.com/xuri/excelize/v2"

// PDF 表格提取
import "github.com/unidoc/unipdf/v3/extractor"
```

#### 数据模型扩展

```go
// 原子内容单元（Atomic Content Unit）
type ContentUnit struct {
    ID          int64         `json:"id" db:"id"`
    DocID       int64         `json:"doc_id" db:"doc_id"`
    Type        ContentType   `json:"type" db:"type"` // text, image, table, formula
    Content     string        `json:"content" db:"content"`
    RawData     []byte        `json:"raw_data" db:"raw_data"` // 原始二进制数据
    
    // 多模态字段
    ImageURL    *string       `json:"image_url" db:"image_url"`
    TableData   string        `json:"table_data" db:"table_data"` // JSONB
    
    // AI 生成的文本表示（用于检索）
    DetailedDesc string       `json:"detailed_desc" db:"detailed_desc"` // 详细描述
    EntitySummary string      `json:"entity_summary" db:"entity_summary"` // 实体摘要
    
    Embedding   xb.Vector     `json:"embedding" db:"embedding"`
    
    // 上下文关系
    ParentID    *int64        `json:"parent_id" db:"parent_id"` // 所属章节
    Position    int           `json:"position" db:"position"` // 文档中的位置
    Metadata    string        `json:"metadata" db:"metadata"` // JSONB
    CreatedAt   time.Time     `json:"created_at" db:"created_at"`
}

type ContentType string

const (
    ContentTypeText    ContentType = "text"
    ContentTypeImage   ContentType = "image"
    ContentTypeTable   ContentType = "table"
    ContentTypeFormula ContentType = "formula"
)

// 表格数据结构
type TableData struct {
    Headers []string        `json:"headers"`
    Rows    [][]string      `json:"rows"`
    Caption string          `json:"caption"`
}
```

#### 实现步骤

**Step 1: PDF 解析器**
```go
// pdf_parser.go
type PDFParser struct {
    mllm MultimodalLLM
}

func (p *PDFParser) Parse(pdfPath string) ([]*ContentUnit, error) {
    // 1. 提取文本块、图片、表格
    // 2. 保留位置和层次信息
    // 3. 为非文本内容生成描述
    // 4. 返回原子内容单元列表
}
```

**Step 2: 多模态 Embedding**
```go
// multimodal_embedder.go
type MultimodalEmbedder struct {
    textEmbedder  EmbeddingService      // 文本 Embedding
    imageEmbedder ImageEmbeddingService // 图像 Embedding (CLIP)
}

func (e *MultimodalEmbedder) EmbedUnit(ctx context.Context, unit *ContentUnit) ([]float32, error) {
    switch unit.Type {
    case ContentTypeText:
        return e.textEmbedder.Embed(ctx, unit.Content)
    case ContentTypeImage:
        // 方案1: 用 CLIP 直接 embed 图片
        // 方案2: 用 AI 生成的描述 embed
        return e.embedDescription(ctx, unit.DetailedDesc)
    case ContentTypeTable:
        // 转为文本描述后 embed
        return e.embedTableAsText(ctx, unit)
    }
}
```

**Step 3: 存储层扩展**
```sql
-- 扩展 document_chunks 表
ALTER TABLE document_chunks 
    ADD COLUMN content_type VARCHAR(20) DEFAULT 'text',
    ADD COLUMN raw_data BYTEA,
    ADD COLUMN image_url TEXT,
    ADD COLUMN table_data JSONB,
    ADD COLUMN detailed_desc TEXT,
    ADD COLUMN entity_summary TEXT,
    ADD COLUMN parent_id BIGINT,
    ADD COLUMN position INT;

-- 为非文本内容创建索引
CREATE INDEX idx_content_type ON document_chunks(content_type);
CREATE INDEX idx_parent_id ON document_chunks(parent_id);
```

#### API 扩展

```go
// 上传 PDF 文档
type UploadPDFRequest struct {
    Title    string `json:"title" binding:"required"`
    FilePath string `json:"file_path" binding:"required"`
}

// POST /api/documents/pdf
func UploadPDFHandler(service *MultimodalRAGService) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req UploadPDFRequest
        // 1. 解析 PDF
        units, err := service.pdfParser.Parse(req.FilePath)
        // 2. 为每个单元生成 embedding
        // 3. 存储到数据库
        // 4. 返回统计信息
    }
}
```

---

### 阶段 2：双图谱构建（2-3 周）

#### 目标
构建跨模态知识图谱 + 文本知识图谱

#### 数据模型

```go
// 知识图谱节点
type KnowledgeNode struct {
    ID          int64      `json:"id" db:"id"`
    Type        NodeType   `json:"type" db:"type"` // entity, content_unit
    Name        string     `json:"name" db:"name"`
    ContentID   *int64     `json:"content_id" db:"content_id"` // 关联 ContentUnit
    Embedding   xb.Vector  `json:"embedding" db:"embedding"`
    Metadata    string     `json:"metadata" db:"metadata"` // JSONB
}

type NodeType string

const (
    NodeTypeEntity      NodeType = "entity"
    NodeTypeContentUnit NodeType = "content_unit"
)

// 知识图谱边
type KnowledgeEdge struct {
    ID        int64    `json:"id" db:"id"`
    SourceID  int64    `json:"source_id" db:"source_id"`
    TargetID  int64    `json:"target_id" db:"target_id"`
    Relation  string   `json:"relation" db:"relation"` // belongs_to, references, describes
    Weight    float64  `json:"weight" db:"weight"`
}
```

#### 图谱构建器

```go
// graph_builder.go
type GraphBuilder struct {
    llm      LLMService
    mllm     MultimodalLLM
    repo     GraphRepository
}

func (b *GraphBuilder) BuildCrossModalGraph(units []*ContentUnit) (*KnowledgeGraph, error) {
    // 1. 为非文本单元提取实体
    for _, unit := range units {
        if unit.Type != ContentTypeText {
            entities := b.mllm.ExtractEntities(unit)
            // 2. 创建节点和边
            // content_unit_node --belongs_to--> entity_node
        }
    }
    
    // 3. 返回跨模态图谱
}

func (b *GraphBuilder) BuildTextGraph(units []*ContentUnit) (*KnowledgeGraph, error) {
    // 1. 从纯文本中提取实体和关系
    // 2. 构建文本图谱
}

func (b *GraphBuilder) FuseGraphs(crossModalGraph, textGraph *KnowledgeGraph) (*KnowledgeGraph, error) {
    // 实体对齐（Entity Alignment）
    // 合并两个图谱
}
```

#### PostgreSQL 图存储

```sql
-- 知识图谱节点表
CREATE TABLE knowledge_nodes (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    name TEXT NOT NULL,
    content_id BIGINT REFERENCES document_chunks(id),
    embedding vector(768),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 知识图谱边表
CREATE TABLE knowledge_edges (
    id BIGSERIAL PRIMARY KEY,
    source_id BIGINT REFERENCES knowledge_nodes(id),
    target_id BIGINT REFERENCES knowledge_nodes(id),
    relation VARCHAR(50) NOT NULL,
    weight FLOAT DEFAULT 1.0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_node_type ON knowledge_nodes(type);
CREATE INDEX idx_edge_source ON knowledge_edges(source_id);
CREATE INDEX idx_edge_target ON knowledge_edges(target_id);
CREATE INDEX idx_edge_relation ON knowledge_edges(relation);
```

---

### 阶段 3：混合检索引擎（2-3 周）

#### 目标
实现结构导航 + 语义搜索的混合检索

#### 检索引擎

```go
// hybrid_retriever.go
type HybridRetriever struct {
    graphRepo     GraphRepository
    vectorRepo    ChunkRepository
    queryAnalyzer *QueryAnalyzer
}

// 混合检索
func (r *HybridRetriever) Retrieve(ctx context.Context, query string, topK int) ([]*ContentUnit, error) {
    // 1. 查询分析（识别模态偏好）
    analysis := r.queryAnalyzer.Analyze(query)
    
    // 2. 结构化导航（在图谱上搜索）
    structuralResults := r.structuralSearch(ctx, query, analysis)
    
    // 3. 语义相似性搜索（向量搜索）
    semanticResults := r.semanticSearch(ctx, query, topK*2)
    
    // 4. 多信号融合排序
    merged := r.fuseAndRank(structuralResults, semanticResults, analysis)
    
    return merged[:topK], nil
}

// 结构化搜索（图遍历）
func (r *HybridRetriever) structuralSearch(ctx context.Context, query string, analysis *QueryAnalysis) []*ContentUnit {
    // 1. 关键词匹配找到起始节点
    startNodes := r.graphRepo.FindNodesByKeywords(query)
    
    // 2. N-hop 邻域扩展
    expandedNodes := r.graphRepo.ExpandNeighborhood(startNodes, 2)
    
    // 3. 转换为 ContentUnit
    return r.nodesToContentUnits(expandedNodes)
}

// 语义搜索（向量搜索）
func (r *HybridRetriever) semanticSearch(ctx context.Context, query string, limit int) []*ContentUnit {
    // 使用现有的向量搜索
    queryVector, _ := r.embedder.Embed(ctx, query)
    return r.vectorRepo.VectorSearch(queryVector, "", "", limit)
}

// 多信号融合
func (r *HybridRetriever) fuseAndRank(
    structural, semantic []*ContentUnit,
    analysis *QueryAnalysis,
) []*ContentUnit {
    // 1. 去重
    // 2. 计算综合得分
    //    - 结构重要性
    //    - 语义相似度
    //    - 模态偏好
    // 3. 排序
}
```

#### 查询分析器

```go
// query_analyzer.go
type QueryAnalyzer struct {
    llm LLMService
}

type QueryAnalysis struct {
    ModalityPreference map[ContentType]float64 // 模态偏好
    RequiresMultiHop   bool                     // 是否需要多跳推理
    Keywords           []string                 // 关键词
}

func (a *QueryAnalyzer) Analyze(query string) *QueryAnalysis {
    // 1. 识别模态关键词（"图X"、"表格"、"公式"）
    // 2. 判断是否需要多跳推理
    // 3. 提取关键词
    
    analysis := &QueryAnalysis{
        ModalityPreference: make(map[ContentType]float64),
    }
    
    // 检测模态偏好
    if strings.Contains(query, "图") || strings.Contains(query, "图片") {
        analysis.ModalityPreference[ContentTypeImage] = 2.0
    }
    if strings.Contains(query, "表格") || strings.Contains(query, "表") {
        analysis.ModalityPreference[ContentTypeTable] = 2.0
    }
    if strings.Contains(query, "公式") || strings.Contains(query, "计算") {
        analysis.ModalityPreference[ContentTypeFormula] = 2.0
    }
    
    return analysis
}
```

---

## 🎨 完整架构（第四代多模态 RAG）

```
┌─────────────────────────────────────────────────────────┐
│                   用户查询（可能带图片）                   │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                    QueryAnalyzer                        │
│  - 模态偏好识别                                           │
│  - 多跳推理判断                                           │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  AgenticRAGService                      │
│  - 问题拆解（第三代能力）                                 │
│  - 多轮召回规划                                           │
└─────────────────────────────────────────────────────────┘
                          │
         ┌────────────────┴────────────────┐
         ▼                                  ▼
┌──────────────────┐              ┌──────────────────┐
│StructuralSearch  │              │ SemanticSearch   │
│ (图谱导航)        │              │ (向量检索)        │
└──────────────────┘              └──────────────────┘
         │                                  │
         └────────────────┬─────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  FusionRanker                           │
│  - 去重                                                  │
│  - 多信号融合（结构 + 语义 + 模态）                       │
│  - 排序                                                  │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                    LLM Generator                        │
│  - 综合多模态内容                                         │
│  - 生成答案（可能引用图片/表格）                          │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              响应（文本 + 图片/表格引用）                  │
└─────────────────────────────────────────────────────────┘
```

---

## 📊 性能预期

### 阶段 1（多模态解析）

| 指标 | 第三代 | 第四代（阶段1） | 提升 |
|------|-------|----------------|------|
| **支持格式** | 纯文本 | PDF + 图片 + 表格 | +200% |
| **信息保留** | 60% | 85% | +25% |
| **延迟** | 2.8s | 3.5s | +25% |

### 阶段 2（图谱构建）

| 指标 | 第三代 | 第四代（阶段2） | 提升 |
|------|-------|----------------|------|
| **多跳推理** | 有限 | 强 | +100% |
| **结构理解** | 无 | 有 | - |
| **准确性** | 87% | 92% | +5% |

### 阶段 3（混合检索）

| 指标 | 第三代 | 第四代（阶段3） | 提升 |
|------|-------|----------------|------|
| **长文档准确性** | 87% | 95%+ | +8% |
| **超长文档（>200页）** | 70% | 85%+ | +15% |
| **非文本检索** | 0% | 90% | +90% |

---

## 🔧 技术栈

### 现有基础
- ✅ Go + xb + pgvector
- ✅ 第三代 Agentic RAG
- ✅ 问题拆解 + 多轮召回

### 新增依赖

**PDF 解析**
```go
github.com/unidoc/unipdf/v3  // PDF 解析（推荐，商业友好）
github.com/pdfcpu/pdfcpu     // 备选开源方案
```

**表格处理**
```go
github.com/xuri/excelize/v2  // Excel 解析
```

**图像理解**
```go
// 多模态 LLM API
- OpenAI GPT-4V
- Claude 3 (Anthropic)
- DeepSeek V2
- Qwen-VL
```

**图数据库（可选）**
```go
// 方案1: PostgreSQL + AGE (Apache AGE 图扩展)
// 方案2: Neo4j + Go Driver
// 方案3: 自建图存储（基于 PostgreSQL 关系表）
```

**向量 + 图混合**
```go
// 推荐：继续使用 PostgreSQL
// pgvector (向量) + 关系表 (图)
// 统一数据库，降低复杂度
```

---

## 🛠️ 实施计划

### 第 1 周：多模态解析基础
- [ ] 集成 unipdf
- [ ] 实现 PDFParser
- [ ] 扩展 ContentUnit 模型
- [ ] 数据库 schema 升级

### 第 2 周：多模态 Embedding
- [ ] 集成多模态 LLM
- [ ] 实现 MultimodalEmbedder
- [ ] 图片描述生成
- [ ] 表格文本化

### 第 3-4 周：图谱构建
- [ ] 实体提取（文本 + 非文本）
- [ ] 跨模态图谱构建
- [ ] 文本图谱构建
- [ ] 图谱融合

### 第 5-6 周：混合检索
- [ ] 结构化搜索实现
- [ ] 语义搜索集成
- [ ] 多信号融合排序
- [ ] QueryAnalyzer

### 第 7 周：集成与测试
- [ ] 集成到 AgenticRAG
- [ ] 端到端测试
- [ ] 性能优化
- [ ] 文档更新

---

## 🎯 里程碑

### Milestone 1: 多模态基础（2 周）
- ✅ 支持 PDF、图片、表格上传
- ✅ 生成 AI 描述
- ✅ 存储到 PostgreSQL

**Demo**: 上传一篇包含图表的论文，查询"图3展示了什么？"

### Milestone 2: 图谱能力（4 周）
- ✅ 双图谱构建
- ✅ 结构化搜索
- ✅ 多跳推理

**Demo**: 查询"X概念和Y图表有什么关系？"（需要图遍历）

### Milestone 3: 完整系统（7 周）
- ✅ 混合检索
- ✅ 模态偏好识别
- ✅ 多信号融合

**Demo**: 查询复杂的多模态问题，返回文本+图片+表格

---

## 📚 参考资源

### 论文
- **RAG-Anything**: All-in-One RAG Framework
- **GraphRAG**: From Local to Global
- **RAPTOR**: Recursive Abstractive Processing

### 开源项目
- **LlamaIndex**: Multi-modal RAG
- **LangChain**: Document Loaders
- **Unstructured**: PDF/Image Parsing

### API
- **OpenAI GPT-4V**: 多模态理解
- **Claude 3**: 图像和文档理解
- **DeepSeek V2**: 中文多模态

---

## 🤔 设计决策

### 为什么不直接用 Neo4j？
- **PostgreSQL 统一存储**：向量 + 图 + 关系数据
- **降低运维复杂度**：单一数据库
- **xb 完美集成**：继续用 xb 操作
- **性能足够**：中小规模图谱（<100万节点）性能可接受

### 为什么不用专门的图数据库？
- **如果需要复杂图算法**（PageRank, 社区发现），考虑 Neo4j
- **如果图规模巨大**（>1000万节点），考虑专业图数据库
- **当前场景**：文档级图谱，PostgreSQL + 关系表足够

---

## 🎉 总结

### 核心价值
1. **突破纯文本限制**：真正理解图片、表格、公式
2. **结构化知识表示**：双图谱捕捉显式关系
3. **智能混合检索**：结构导航 + 语义搜索
4. **保持第三代优势**：问题拆解 + 多轮召回

### 竞争优势
- ✅ **Go 生态独一份**：多模态 RAG + xb + pgvector
- ✅ **性能优越**：Go 原生性能 + 统一存储
- ✅ **易于部署**：单一 PostgreSQL 数据库
- ✅ **渐进升级**：不破坏现有功能

---

**第四代多模态 RAG - 让 AI 真正"看懂"文档！** 🖼️📊📈

