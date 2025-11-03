package g4

import (
	"fmt"

	"github.com/fndome/xb"
	"github.com/jmoiron/sqlx"
)

// MultimodalRepository 多模态内容仓库
// 展示如何使用 xb 处理多模态数据
type MultimodalRepository struct {
	db *sqlx.DB
}

func NewMultimodalRepository(db *sqlx.DB) *MultimodalRepository {
	return &MultimodalRepository{db: db}
}

// ==================== 内容单元操作 ====================

// CreateUnit 创建内容单元
// ⭐ 展示：使用 xb 插入多模态内容
func (r *MultimodalRepository) CreateUnit(unit *ContentUnit) error {
	sql, args := xb.Of(&ContentUnit{}).
		Insert(func(ib *xb.InsertBuilder) {
			ib.Set("doc_id", unit.DocID).
				Set("type", unit.Type).
				Set("position", unit.Position).
				Set("content", unit.Content).
				Set("raw_data", unit.RawData).
				Set("image_url", unit.ImageURL).
				Set("table_data", unit.TableData).
				Set("detailed_desc", unit.DetailedDesc).
				Set("entity_summary", unit.EntitySummary).
				Set("embedding", unit.Embedding). // ⭐ xb.Vector 自动序列化
				Set("parent_id", unit.ParentID).
				Set("metadata", unit.Metadata)
		}).
		Build().
		SqlOfInsert()

	_, err := r.db.Exec(sql, args...)
	return err
}

// GetUnit 获取单个内容单元
func (r *MultimodalRepository) GetUnit(id int64) (*ContentUnit, error) {
	sql, args, _ := xb.Of(&ContentUnit{}).
		Eq("id", id).
		Build().
		SqlOfSelect()

	var unit ContentUnit
	err := r.db.Get(&unit, sql, args...)
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

// ==================== 向量检索 ====================

// VectorSearch 纯向量检索
// ⭐ 展示：xb 的基础向量检索能力
func (r *MultimodalRepository) VectorSearch(
	queryVector []float32,
	limit int,
) ([]*ContentUnit, error) {
	sql, args := xb.Of(&ContentUnit{}).
		VectorSearch("embedding", queryVector, limit).
		Build().
		SqlOfVectorSearch()

	var units []*ContentUnit
	err := r.db.Select(&units, sql, args...)
	if err != nil {
		return nil, err
	}
	return units, nil
}

// VectorSearchByType 按类型的向量检索
// ⭐ 展示：向量检索 + 标量过滤
func (r *MultimodalRepository) VectorSearchByType(
	queryVector []float32,
	contentType ContentType,
	limit int,
) ([]*ContentUnit, error) {
	sql, args := xb.Of(&ContentUnit{}).
		VectorSearch("embedding", queryVector, limit).
		Eq("type", contentType). // 过滤特定模态
		Build().
		SqlOfVectorSearch()

	var units []*ContentUnit
	err := r.db.Select(&units, sql, args...)
	if err != nil {
		return nil, err
	}
	return units, nil
}

// VectorSearchByDoc 按文档的向量检索
// ⭐ 展示：向量检索 + 文档过滤
func (r *MultimodalRepository) VectorSearchByDoc(
	queryVector []float32,
	docID int64,
	limit int,
) ([]*ContentUnit, error) {
	sql, args := xb.Of(&ContentUnit{}).
		VectorSearch("embedding", queryVector, limit).
		Eq("doc_id", docID). // 限定文档范围
		Build().
		SqlOfVectorSearch()

	var units []*ContentUnit
	err := r.db.Select(&units, sql, args...)
	if err != nil {
		return nil, err
	}
	return units, nil
}

// HybridSearch 混合检索
// ⭐ 展示：向量 + 多条件过滤的复杂查询
func (r *MultimodalRepository) HybridSearch(query HybridQuery) ([]*ContentUnit, error) {
	builder := xb.Of(&ContentUnit{}).
		VectorSearch("embedding", query.QueryVector, query.TopK*2) // Over-fetch

	// 文档过滤
	if query.DocID != nil {
		builder = builder.Eq("doc_id", *query.DocID)
	}

	// 模态过滤
	if len(query.AllowedTypes) > 0 {
		// 转换为 interface{} 切片（xb 要求）
		typeInterfaces := make([]interface{}, len(query.AllowedTypes))
		for i, t := range query.AllowedTypes {
			typeInterfaces[i] = string(t)
		}
		builder = builder.In("type", typeInterfaces)
	}

	// 时间范围过滤
	if query.TimeRange != nil {
		builder = builder.
			Gte("created_at", query.TimeRange.Start).
			Lte("created_at", query.TimeRange.End)
	}

	// 关键词过滤（可选）
	if query.Text != "" {
		builder = builder.Like("content", "%"+query.Text+"%")
	}

	sql, args := builder.Build().SqlOfVectorSearch()

	var units []*ContentUnit
	err := r.db.Select(&units, sql, args...)
	if err != nil {
		return nil, err
	}

	// 应用模态偏好重排（如果指定）
	if len(query.ModalityPrefer) > 0 {
		units = r.applyModalityPreference(units, query.ModalityPrefer)
	}

	// 限制返回数量
	if len(units) > query.TopK {
		units = units[:query.TopK]
	}

	return units, nil
}

// applyModalityPreference 应用模态偏好
func (r *MultimodalRepository) applyModalityPreference(
	units []*ContentUnit,
	preferences map[ContentType]float64,
) []*ContentUnit {
	// 简化版：根据模态偏好调整顺序
	// 实际应用中可以结合距离和偏好计算综合得分

	// 按类型分组
	groups := make(map[ContentType][]*ContentUnit)
	for _, unit := range units {
		groups[unit.Type] = append(groups[unit.Type], unit)
	}

	// 按偏好权重排序
	var result []*ContentUnit
	for contentType := range groups {
		weight := 1.0
		if w, ok := preferences[contentType]; ok {
			weight = w
		}

		// 简化：权重高的类型优先
		if weight > 1.0 {
			result = append(groups[contentType], result...)
		} else {
			result = append(result, groups[contentType]...)
		}
	}

	return result
}

// ==================== 知识图谱操作 ====================

// CreateNode 创建知识图谱节点
// ⭐ 展示：使用 xb 构建知识图谱
func (r *MultimodalRepository) CreateNode(node *KnowledgeNode) error {
	sql, args := xb.Of(&KnowledgeNode{}).
		Insert(func(ib *xb.InsertBuilder) {
			ib.Set("type", node.Type).
				Set("name", node.Name).
				Set("content_id", node.ContentID).
				Set("embedding", node.Embedding). // ⭐ 节点也有向量
				Set("metadata", node.Metadata)
		}).
		Build().
		SqlOfInsert()

	result, err := r.db.Exec(sql, args...)
	if err != nil {
		return err
	}

	// 获取插入的 ID
	id, _ := result.LastInsertId()
	node.ID = id

	return nil
}

// CreateEdge 创建知识图谱边
func (r *MultimodalRepository) CreateEdge(edge *KnowledgeEdge) error {
	sql, args := xb.Of(&KnowledgeEdge{}).
		Insert(func(ib *xb.InsertBuilder) {
			ib.Set("source_id", edge.SourceID).
				Set("target_id", edge.TargetID).
				Set("relation", edge.Relation).
				Set("weight", edge.Weight).
				Set("metadata", edge.Metadata)
		}).
		Build().
		SqlOfInsert()

	_, err := r.db.Exec(sql, args...)
	return err
}

// GetNodeNeighbors 获取节点的邻居
// ⭐ 展示：图遍历查询
func (r *MultimodalRepository) GetNodeNeighbors(nodeID int64) ([]*KnowledgeNode, error) {
	// 查询所有出边
	sql, args, _ := xb.Of(&KnowledgeEdge{}).
		Eq("source_id", nodeID).
		Build().
		SqlOfSelect()

	var edges []*KnowledgeEdge
	err := r.db.Select(&edges, sql, args...)
	if err != nil {
		return nil, err
	}

	if len(edges) == 0 {
		return []*KnowledgeNode{}, nil
	}

	// 提取目标节点 ID
	targetIDs := make([]interface{}, len(edges))
	for i, edge := range edges {
		targetIDs[i] = edge.TargetID
	}

	// 批量查询节点
	// ⭐ 注意：xb 的 In() 使用可变参数，需要展开切片
	sql, args, _ = xb.Of(&KnowledgeNode{}).
		In("id", targetIDs...). // 使用 ... 展开
		Build().
		SqlOfSelect()

	var nodes []*KnowledgeNode
	err = r.db.Select(&nodes, sql, args...)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// FindPathBetweenNodes 查找两个节点之间的路径
// ⭐ 展示：复杂的图查询（BFS）
func (r *MultimodalRepository) FindPathBetweenNodes(
	startID, endID int64,
	maxDepth int,
) ([][]*KnowledgeEdge, error) {
	// 简化实现：使用递归 CTE 或多次查询
	// 这里展示基本思路

	type PathNode struct {
		NodeID int64
		Depth  int
		Path   []*KnowledgeEdge
	}

	visited := make(map[int64]bool)
	queue := []PathNode{{NodeID: startID, Depth: 0, Path: []*KnowledgeEdge{}}}
	var paths [][]*KnowledgeEdge

	for len(queue) > 0 && len(paths) < 10 { // 最多返回 10 条路径
		current := queue[0]
		queue = queue[1:]

		if current.NodeID == endID {
			paths = append(paths, current.Path)
			continue
		}

		if current.Depth >= maxDepth || visited[current.NodeID] {
			continue
		}

		visited[current.NodeID] = true

		// 获取邻居
		sql, args, _ := xb.Of(&KnowledgeEdge{}).
			Eq("source_id", current.NodeID).
			Build().
			SqlOfSelect()

		var edges []*KnowledgeEdge
		r.db.Select(&edges, sql, args...)

		for _, edge := range edges {
			newPath := make([]*KnowledgeEdge, len(current.Path)+1)
			copy(newPath, current.Path)
			newPath[len(current.Path)] = edge

			queue = append(queue, PathNode{
				NodeID: *edge.TargetID,
				Depth:  current.Depth + 1,
				Path:   newPath,
			})
		}
	}

	return paths, nil
}

// ==================== 批量操作 ====================

// BatchCreateUnits 批量创建内容单元
// ⭐ 展示：高效的批量插入
func (r *MultimodalRepository) BatchCreateUnits(units []*ContentUnit) error {
	if len(units) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, unit := range units {
		sql, args := xb.Of(&ContentUnit{}).
			Insert(func(ib *xb.InsertBuilder) {
				ib.Set("doc_id", unit.DocID).
					Set("type", unit.Type).
					Set("position", unit.Position).
					Set("content", unit.Content).
					Set("embedding", unit.Embedding).
					Set("detailed_desc", unit.DetailedDesc)
			}).
			Build().
			SqlOfInsert()

		if _, err := tx.Exec(sql, args...); err != nil {
			return fmt.Errorf("insert unit failed: %w", err)
		}
	}

	return tx.Commit()
}

// ==================== 统计查询 ====================

// GetDocumentStats 获取文档统计信息
func (r *MultimodalRepository) GetDocumentStats(docID int64) (map[ContentType]int, error) {
	// 使用 xb 的聚合查询（未来功能）
	// 当前使用原生 SQL
	query := `
		SELECT type, COUNT(*) as count
		FROM content_units
		WHERE doc_id = $1
		GROUP BY type
	`

	rows, err := r.db.Query(query, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[ContentType]int)
	for rows.Next() {
		var contentType ContentType
		var count int
		if err := rows.Scan(&contentType, &count); err != nil {
			return nil, err
		}
		stats[contentType] = count
	}

	return stats, nil
}

// DeleteUnit 删除内容单元
func (r *MultimodalRepository) DeleteUnit(id int64) error {
	sql, args := xb.Of(&ContentUnit{}).
		Eq("id", id).
		Build().
		SqlOfDelete()

	_, err := r.db.Exec(sql, args...)
	return err
}

// UpdateUnitEmbedding 更新内容单元的向量
// ⭐ 展示：向量更新操作
func (r *MultimodalRepository) UpdateUnitEmbedding(id int64, embedding xb.Vector) error {
	sql, args := xb.Of(&ContentUnit{}).
		Update(func(ub *xb.UpdateBuilder) {
			ub.Set("embedding", embedding)
		}).
		Eq("id", id).
		Build().
		SqlOfUpdate()

	_, err := r.db.Exec(sql, args...)
	return err
}
