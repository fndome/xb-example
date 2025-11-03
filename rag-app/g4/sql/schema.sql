-- 第四代多模态 RAG 数据库 Schema
-- 展示如何使用 PostgreSQL + pgvector 存储多模态内容

-- 启用 pgvector 扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- ==================== 文档表 ====================

CREATE TABLE IF NOT EXISTS documents (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    filename VARCHAR(500) NOT NULL,
    file_type VARCHAR(50) NOT NULL,  -- pdf, docx, jpg, etc.
    file_size BIGINT DEFAULT 0,
    
    -- 统计信息
    total_units INT DEFAULT 0,
    text_units INT DEFAULT 0,
    image_units INT DEFAULT 0,
    table_units INT DEFAULT 0,
    
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_documents_created_at ON documents(created_at);

-- ==================== 内容单元表（核心） ====================

CREATE TABLE IF NOT EXISTS content_units (
    id BIGSERIAL PRIMARY KEY,
    doc_id BIGINT REFERENCES documents(id) ON DELETE CASCADE,
    
    -- 类型和位置
    type VARCHAR(20) NOT NULL,  -- text, image, table, formula
    position INT DEFAULT 0,     -- 在文档中的位置
    
    -- 内容
    content TEXT,               -- 文本内容或描述
    raw_data BYTEA,             -- 原始二进制数据
    
    -- 多模态字段
    image_url TEXT,             -- 图片 URL
    table_data JSONB,           -- 表格数据（结构化）
    
    -- AI 生成的文本表示
    detailed_desc TEXT,         -- 详细描述（用于展示）
    entity_summary TEXT,        -- 实体摘要（用于图谱）
    
    -- ⭐ 向量字段（使用 pgvector）
    embedding vector(768),      -- 768 维向量（OpenAI text-embedding-3-small）
    
    -- 层次结构
    parent_id BIGINT REFERENCES content_units(id) ON DELETE CASCADE,
    
    -- 元数据
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- 索引优化
CREATE INDEX idx_content_units_doc_id ON content_units(doc_id);
CREATE INDEX idx_content_units_type ON content_units(type);
CREATE INDEX idx_content_units_doc_type ON content_units(doc_id, type);
CREATE INDEX idx_content_units_parent_id ON content_units(parent_id);
CREATE INDEX idx_content_units_created_at ON content_units(created_at);

-- ⭐ 向量索引（IVFFlat，适合中大规模数据）
CREATE INDEX idx_content_units_embedding ON content_units 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- 说明：
-- - lists = 100 适合 10万 ~ 100万 数据
-- - lists = sqrt(total_rows) 是通用公式
-- - 可以根据实际数据量调整

-- ⭐ 或使用 HNSW 索引（更快但占用更多内存）
-- CREATE INDEX idx_content_units_embedding_hnsw ON content_units 
-- USING hnsw (embedding vector_cosine_ops)
-- WITH (m = 16, ef_construction = 64);

-- ==================== 知识图谱表 ====================

-- 图节点表
CREATE TABLE IF NOT EXISTS knowledge_nodes (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,  -- entity, content_unit
    name VARCHAR(500) NOT NULL,
    content_id BIGINT REFERENCES content_units(id) ON DELETE CASCADE,
    
    -- 节点向量（用于语义搜索）
    embedding vector(768),
    
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_knowledge_nodes_type ON knowledge_nodes(type);
CREATE INDEX idx_knowledge_nodes_name ON knowledge_nodes(name);
CREATE INDEX idx_knowledge_nodes_content_id ON knowledge_nodes(content_id);

-- 向量索引
CREATE INDEX idx_knowledge_nodes_embedding ON knowledge_nodes 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- 图边表
CREATE TABLE IF NOT EXISTS knowledge_edges (
    id BIGSERIAL PRIMARY KEY,
    source_id BIGINT REFERENCES knowledge_nodes(id) ON DELETE CASCADE,
    target_id BIGINT REFERENCES knowledge_nodes(id) ON DELETE CASCADE,
    relation VARCHAR(100) NOT NULL,  -- describes, belongs_to, references, etc.
    weight FLOAT DEFAULT 1.0,
    
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_knowledge_edges_source ON knowledge_edges(source_id);
CREATE INDEX idx_knowledge_edges_target ON knowledge_edges(target_id);
CREATE INDEX idx_knowledge_edges_relation ON knowledge_edges(relation);
CREATE INDEX idx_knowledge_edges_source_relation ON knowledge_edges(source_id, relation);

-- ==================== 辅助视图 ====================

-- 多模态统计视图
CREATE OR REPLACE VIEW v_document_stats AS
SELECT 
    d.id,
    d.title,
    d.total_units,
    COUNT(CASE WHEN cu.type = 'text' THEN 1 END) AS text_count,
    COUNT(CASE WHEN cu.type = 'image' THEN 1 END) AS image_count,
    COUNT(CASE WHEN cu.type = 'table' THEN 1 END) AS table_count,
    COUNT(CASE WHEN cu.type = 'formula' THEN 1 END) AS formula_count
FROM documents d
LEFT JOIN content_units cu ON d.id = cu.doc_id
GROUP BY d.id, d.title, d.total_units;

-- 图谱统计视图
CREATE OR REPLACE VIEW v_graph_stats AS
SELECT 
    COUNT(DISTINCT kn.id) AS total_nodes,
    COUNT(DISTINCT CASE WHEN kn.type = 'entity' THEN kn.id END) AS entity_nodes,
    COUNT(DISTINCT CASE WHEN kn.type = 'content_unit' THEN kn.id END) AS content_nodes,
    COUNT(ke.id) AS total_edges,
    COUNT(DISTINCT ke.relation) AS unique_relations
FROM knowledge_nodes kn
LEFT JOIN knowledge_edges ke ON kn.id = ke.source_id OR kn.id = ke.target_id;

-- ==================== 有用的函数 ====================

-- 计算向量余弦相似度
CREATE OR REPLACE FUNCTION cosine_similarity(a vector, b vector)
RETURNS float AS $$
BEGIN
    RETURN 1 - (a <=> b);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- 查找节点的 N 跳邻居（递归 CTE）
CREATE OR REPLACE FUNCTION find_neighbors(
    start_node_id BIGINT,
    max_depth INT DEFAULT 2
)
RETURNS TABLE(node_id BIGINT, depth INT, path BIGINT[]) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE neighbors AS (
        -- 起始节点
        SELECT 
            start_node_id AS node_id,
            0 AS depth,
            ARRAY[start_node_id] AS path
        
        UNION ALL
        
        -- 递归：查找邻居
        SELECT 
            ke.target_id AS node_id,
            n.depth + 1 AS depth,
            n.path || ke.target_id AS path
        FROM neighbors n
        JOIN knowledge_edges ke ON n.node_id = ke.source_id
        WHERE n.depth < max_depth
          AND NOT (ke.target_id = ANY(n.path))  -- 避免环
    )
    SELECT DISTINCT n.node_id, n.depth, n.path
    FROM neighbors n
    WHERE n.depth > 0
    ORDER BY n.depth, n.node_id;
END;
$$ LANGUAGE plpgsql;

-- ==================== 示例查询（注释） ====================

-- 1. 向量检索（Top 10）
-- SELECT id, type, content, embedding <=> '[0.1, 0.2, ...]'::vector AS distance
-- FROM content_units
-- ORDER BY embedding <=> '[0.1, 0.2, ...]'::vector
-- LIMIT 10;

-- 2. 混合检索（向量 + 标量过滤）
-- SELECT id, type, content, embedding <=> $1 AS distance
-- FROM content_units
-- WHERE doc_id = $2
--   AND type IN ('image', 'table')
--   AND created_at >= $3
-- ORDER BY embedding <=> $1
-- LIMIT $4;

-- 3. 跨模态检索（文本查询检索图片）
-- SELECT id, image_url, detailed_desc, embedding <=> $1 AS distance
-- FROM content_units
-- WHERE type = 'image'
-- ORDER BY embedding <=> $1
-- LIMIT 5;

-- 4. 图遍历（查找节点的邻居）
-- SELECT * FROM find_neighbors(100, 2);

-- 5. 实体检索（在图谱中搜索）
-- SELECT kn.id, kn.name, kn.embedding <=> $1 AS distance
-- FROM knowledge_nodes kn
-- WHERE kn.type = 'entity'
-- ORDER BY kn.embedding <=> $1
-- LIMIT 10;

-- ==================== 性能优化建议 ====================

-- 1. 定期 VACUUM（保持索引效率）
-- VACUUM ANALYZE content_units;
-- VACUUM ANALYZE knowledge_nodes;

-- 2. 更新统计信息
-- ANALYZE content_units;

-- 3. 监控查询性能
-- EXPLAIN ANALYZE
-- SELECT ... FROM content_units
-- WHERE ... ORDER BY embedding <=> ...;

-- 4. 调整 pgvector 参数
-- SET ivfflat.probes = 10;  -- 增加精度但降低速度
-- SET hnsw.ef_search = 40;   -- HNSW 搜索参数

-- ==================== 清理 ====================

-- DROP TABLE IF EXISTS knowledge_edges CASCADE;
-- DROP TABLE IF EXISTS knowledge_nodes CASCADE;
-- DROP TABLE IF EXISTS content_units CASCADE;
-- DROP TABLE IF EXISTS documents CASCADE;
-- DROP FUNCTION IF EXISTS find_neighbors;
-- DROP FUNCTION IF EXISTS cosine_similarity;

