-- 第四代多模态 RAG 示例数据
-- 展示如何在 PostgreSQL 中存储和查询多模态内容

-- ==================== 插入示例文档 ====================

INSERT INTO documents (title, filename, file_type, file_size, total_units, text_units, image_units, table_units, metadata) VALUES
('深度学习研究论文', 'deep_learning_2024.pdf', 'pdf', 2048000, 15, 10, 3, 2, '{"author": "张三", "year": 2024}'),
('2024年第一季度财报', 'Q1_2024_report.pdf', 'pdf', 1024000, 8, 5, 2, 1, '{"department": "财务部", "quarter": "Q1"}'),
('机器学习教程', 'ml_tutorial.docx', 'docx', 512000, 20, 18, 1, 1, '{"category": "教程", "level": "初级"}');

-- ==================== 插入示例内容单元 ====================

-- 文档 1：深度学习论文

-- 文本内容
INSERT INTO content_units (doc_id, type, position, content, embedding, detailed_desc, metadata) VALUES
(1, 'text', 1, 'VAE（变分自编码器）是一种生成模型，通过学习数据的潜在分布来生成新样本。', 
 '[' || string_agg(random()::text, ',') || ']'::vector,
 'VAE模型的基本介绍',
 '{"section": "Introduction", "page": 1}')
FROM generate_series(1, 768);

INSERT INTO content_units (doc_id, type, position, content, embedding, detailed_desc, metadata) VALUES
(1, 'text', 2, 'VAE由编码器和解码器两部分组成，编码器将输入映射到潜在空间，解码器从潜在空间重构输入。',
 '[' || string_agg(random()::text, ',') || ']'::vector,
 'VAE的架构说明',
 '{"section": "Architecture", "page": 2}')
FROM generate_series(1, 768);

-- 图片内容
INSERT INTO content_units (doc_id, type, position, content, image_url, embedding, detailed_desc, entity_summary, metadata) VALUES
(1, 'image', 3, '图1：VAE模型架构示意图', 
 'https://example.com/images/vae_architecture.png',
 '[' || string_agg(random()::text, ',') || ']'::vector,
 '该图展示了VAE的完整架构，包括编码器（Encoder）、潜在空间（Latent Space）和解码器（Decoder）三个主要部分。编码器将输入图像压缩为均值和方差，解码器从采样的潜在向量重构图像。',
 'VAE模型, 编码器, 解码器, 潜在空间',
 '{"caption": "图1：VAE架构", "page": 2, "size": "800x600"}')
FROM generate_series(1, 768);

INSERT INTO content_units (doc_id, type, position, content, image_url, embedding, detailed_desc, entity_summary, metadata) VALUES
(1, 'image', 4, '图2：训练损失曲线',
 'https://example.com/images/loss_curve.png',
 '[' || string_agg(random()::text, ',') || ']'::vector,
 '该图展示了VAE在MNIST数据集上的训练损失曲线。横轴表示训练轮次（Epochs），纵轴表示损失值（Loss）。曲线显示损失在前10个epoch快速下降，然后趋于平稳。',
 '训练损失, MNIST数据集, 损失曲线',
 '{"caption": "图2：损失曲线", "page": 5, "dataset": "MNIST"}')
FROM generate_series(1, 768);

-- 表格内容
INSERT INTO content_units (doc_id, type, position, content, table_data, embedding, detailed_desc, entity_summary, metadata) VALUES
(1, 'table', 5, '表1：不同模型在MNIST上的性能对比',
 '{
   "caption": "表1：模型性能对比",
   "headers": ["模型", "准确率", "训练时间(分钟)", "参数量(M)"],
   "rows": [
     ["VAE", "92.5%", "45", "2.3"],
     ["GAN", "93.8%", "60", "3.1"],
     ["Diffusion", "95.2%", "120", "5.8"]
   ],
   "summary": "Diffusion模型准确率最高但训练时间最长"
 }'::jsonb,
 '[' || string_agg(random()::text, ',') || ']'::vector,
 '该表对比了VAE、GAN和Diffusion三种生成模型在MNIST数据集上的性能。Diffusion模型准确率最高达95.2%，但训练时间也最长需要120分钟。VAE模型虽然准确率略低为92.5%，但训练速度快且参数量少。',
 'VAE, GAN, Diffusion, MNIST, 性能对比',
 '{"page": 6, "table_type": "comparison"}')
FROM generate_series(1, 768);

-- 公式内容
INSERT INTO content_units (doc_id, type, position, content, embedding, detailed_desc, metadata) VALUES
(1, 'formula', 6, '公式1：VAE的损失函数 L = -E[log p(x|z)] + KL(q(z|x)||p(z))',
 '[' || string_agg(random()::text, ',') || ']'::vector,
 'VAE的损失函数包含两部分：重构损失和KL散度。重构损失衡量重构质量，KL散度约束潜在分布接近标准正态分布。',
 '{"formula_type": "loss_function", "page": 3}')
FROM generate_series(1, 768);

-- 文档 2：财报

INSERT INTO content_units (doc_id, type, position, content, embedding, detailed_desc, metadata) VALUES
(2, 'text', 1, '2024年第一季度，公司实现营收10.5亿元，同比增长25%。',
 '[' || string_agg(random()::text, ',') || ']'::vector,
 'Q1季度营收概述',
 '{"section": "财务概要", "page": 1}')
FROM generate_series(1, 768);

INSERT INTO content_units (doc_id, type, position, content, image_url, embedding, detailed_desc, entity_summary, metadata) VALUES
(2, 'image', 2, '图1：2024年Q1营收趋势图',
 'https://example.com/images/revenue_trend.png',
 '[' || string_agg(random()::text, ',') || ']'::vector,
 '该图表展示了2024年第一季度的月度营收趋势。1月营收3.2亿，2月营收3.5亿，3月营收3.8亿，呈现稳定增长态势。柱状图显示营收逐月递增。',
 '营收趋势, 季度数据, 柱状图',
 '{"caption": "图1：营收趋势", "chart_type": "bar", "page": 2}')
FROM generate_series(1, 768);

INSERT INTO content_units (doc_id, type, position, content, table_data, embedding, detailed_desc, entity_summary, metadata) VALUES
(2, 'table', 3, '表1：各部门营收占比',
 '{
   "caption": "表1：部门营收占比",
   "headers": ["部门", "营收(亿元)", "占比", "同比增长"],
   "rows": [
     ["云计算", "4.5", "43%", "+35%"],
     ["电商", "3.2", "30%", "+18%"],
     ["广告", "2.8", "27%", "+22%"]
   ],
   "summary": "云计算部门贡献最大，占比43%，增长最快达35%"
 }'::jsonb,
 '[' || string_agg(random()::text, ',') || ']'::vector,
 '该表详细列出了各部门的营收数据。云计算部门表现突出，营收4.5亿占总营收43%，同比增长35%。电商和广告部门也保持稳定增长。',
 '部门营收, 云计算, 电商, 广告',
 '{"page": 3, "table_type": "breakdown"}')
FROM generate_series(1, 768);

-- 文档 3：教程

INSERT INTO content_units (doc_id, type, position, content, embedding, detailed_desc, metadata) VALUES
(3, 'text', 1, '机器学习是人工智能的一个分支，通过算法使计算机能够从数据中学习。',
 '[' || string_agg(random()::text, ',') || ']'::vector,
 '机器学习基础定义',
 '{"chapter": 1, "page": 1}')
FROM generate_series(1, 768);

-- ==================== 插入知识图谱 ====================

-- 实体节点
INSERT INTO knowledge_nodes (type, name, embedding, metadata) VALUES
('entity', 'VAE模型', '[' || string_agg(random()::text, ',') || ']'::vector, '{"category": "模型"}'),
('entity', 'GAN模型', '[' || string_agg(random()::text, ',') || ']'::vector, '{"category": "模型"}'),
('entity', 'MNIST数据集', '[' || string_agg(random()::text, ',') || ']'::vector, '{"category": "数据集"}'),
('entity', '云计算业务', '[' || string_agg(random()::text, ',') || ']'::vector, '{"category": "业务线"}')
FROM generate_series(1, 768);

-- 内容单元节点（关联到实际内容）
INSERT INTO knowledge_nodes (type, name, content_id, embedding) 
SELECT 
    'content_unit',
    'ContentUnit_' || id,
    id,
    embedding
FROM content_units
WHERE type IN ('image', 'table')
LIMIT 5;

-- 图边（实体关系）
INSERT INTO knowledge_edges (source_id, target_id, relation, weight) VALUES
-- VAE模型 describes 图片
(1, 6, 'describes', 1.0),  -- VAE -> 图1（VAE架构）
-- MNIST数据集 used_in 表格
(3, 8, 'used_in', 1.0),    -- MNIST -> 表1（性能对比）
-- VAE模型 compared_with GAN模型
(1, 2, 'compared_with', 0.8),  -- VAE <-> GAN
-- 云计算业务 shows_in 图表
(4, 9, 'shows_in', 1.0);   -- 云计算 -> 营收趋势图

-- ==================== 示例查询 ====================

-- 查询 1：查看所有多模态内容统计
SELECT 
    d.title,
    COUNT(cu.id) AS total_units,
    SUM(CASE WHEN cu.type = 'text' THEN 1 ELSE 0 END) AS text_count,
    SUM(CASE WHEN cu.type = 'image' THEN 1 ELSE 0 END) AS image_count,
    SUM(CASE WHEN cu.type = 'table' THEN 1 ELSE 0 END) AS table_count,
    SUM(CASE WHEN cu.type = 'formula' THEN 1 ELSE 0 END) AS formula_count
FROM documents d
LEFT JOIN content_units cu ON d.id = cu.doc_id
GROUP BY d.id, d.title;

-- 查询 2：查看知识图谱统计
SELECT 
    (SELECT COUNT(*) FROM knowledge_nodes WHERE type = 'entity') AS entity_nodes,
    (SELECT COUNT(*) FROM knowledge_nodes WHERE type = 'content_unit') AS content_nodes,
    (SELECT COUNT(*) FROM knowledge_edges) AS total_edges;

-- 查询 3：查看某个实体的所有关联内容
SELECT 
    kn1.name AS entity_name,
    ke.relation,
    kn2.name AS related_content,
    cu.type AS content_type,
    cu.content
FROM knowledge_nodes kn1
JOIN knowledge_edges ke ON kn1.id = ke.source_id
JOIN knowledge_nodes kn2 ON ke.target_id = kn2.id
LEFT JOIN content_units cu ON kn2.content_id = cu.id
WHERE kn1.name = 'VAE模型';

-- ==================== 注释 ====================

-- 说明：
-- 1. 示例数据使用 random() 生成模拟向量，实际应用中应该使用真实的 Embedding
-- 2. 向量维度为 768（OpenAI text-embedding-3-small）
-- 3. 图片 URL 为示例地址，实际应用中应该指向真实的图片存储服务
-- 4. 表格数据使用 JSONB 格式存储，便于查询和展示
-- 5. 知识图谱展示了实体和内容之间的关联关系

-- 清理示例数据：
-- DELETE FROM knowledge_edges;
-- DELETE FROM knowledge_nodes;
-- DELETE FROM content_units;
-- DELETE FROM documents;

