package g4

import (
	"time"

	"github.com/fndome/xb"
)

// ContentType 内容类型
type ContentType string

const (
	ContentTypeText    ContentType = "text"
	ContentTypeImage   ContentType = "image"
	ContentTypeTable   ContentType = "table"
	ContentTypeFormula ContentType = "formula"
	ContentTypeVideo   ContentType = "video" // 未来扩展
	ContentTypeAudio   ContentType = "audio" // 未来扩展
)

// ContentUnit 多模态内容单元
// 这是第四代 RAG 的核心数据结构，使用 xb.Vector 存储向量
type ContentUnit struct {
	// 基础字段
	ID       int64       `json:"id" db:"id"`
	DocID    *int64      `json:"doc_id" db:"doc_id"`     // 所属文档
	Type     ContentType `json:"type" db:"type"`         // 内容类型
	Position *int        `json:"position" db:"position"` // ⭐ 文档中的位置（指针类型）

	// 内容字段
	Content string `json:"content" db:"content"`   // 文本内容/描述
	RawData []byte `json:"raw_data" db:"raw_data"` // 原始二进制数据

	// 多模态专属字段
	ImageURL  *string `json:"image_url" db:"image_url"`   // 图片 URL
	TableData string  `json:"table_data" db:"table_data"` // 表格数据（JSONB）

	// AI 生成的文本表示（用于检索）
	DetailedDesc  string `json:"detailed_desc" db:"detailed_desc"`   // 详细描述
	EntitySummary string `json:"entity_summary" db:"entity_summary"` // 实体摘要

	// ⭐ 向量字段（使用 xb.Vector）
	Embedding xb.Vector `json:"embedding" db:"embedding"`

	// 层次结构
	ParentID *int64 `json:"parent_id" db:"parent_id"` // 父节点（如章节）

	// 元数据
	Metadata  string    `json:"metadata" db:"metadata"` // JSONB
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (*ContentUnit) TableName() string {
	return "content_units"
}

// KnowledgeNode 知识图谱节点
type KnowledgeNode struct {
	ID        int64    `json:"id" db:"id"`
	Type      NodeType `json:"type" db:"type"`             // entity 或 content_unit
	Name      string   `json:"name" db:"name"`             // 节点名称
	ContentID *int64   `json:"content_id" db:"content_id"` // 关联的 ContentUnit

	// ⭐ 向量字段
	Embedding xb.Vector `json:"embedding" db:"embedding"`

	Metadata  string    `json:"metadata" db:"metadata"` // JSONB
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (*KnowledgeNode) TableName() string {
	return "knowledge_nodes"
}

// NodeType 节点类型
type NodeType string

const (
	NodeTypeEntity      NodeType = "entity"       // 实体节点
	NodeTypeContentUnit NodeType = "content_unit" // 内容单元节点
)

// KnowledgeEdge 知识图谱边
type KnowledgeEdge struct {
	ID       int64    `json:"id" db:"id"`
	SourceID *int64   `json:"source_id" db:"source_id"` // ⭐ 源节点（指针类型）
	TargetID *int64   `json:"target_id" db:"target_id"` // ⭐ 目标节点（指针类型）
	Relation string   `json:"relation" db:"relation"`   // 关系类型
	Weight   *float64 `json:"weight" db:"weight"`       // ⭐ 权重（指针类型）

	Metadata  string    `json:"metadata" db:"metadata"` // JSONB
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (*KnowledgeEdge) TableName() string {
	return "knowledge_edges"
}

// Document 文档元信息
type Document struct {
	ID       int64  `json:"id" db:"id"`
	Title    string `json:"title" db:"title"`
	Filename string `json:"filename" db:"filename"`
	FileType string `json:"file_type" db:"file_type"` // pdf, docx, jpg, etc.
	FileSize *int64 `json:"file_size" db:"file_size"` // ⭐ 文件大小（指针类型）

	// 统计信息
	TotalUnits *int `json:"total_units" db:"total_units"` // ⭐ 总内容单元数（指针类型）
	TextUnits  *int `json:"text_units" db:"text_units"`   // ⭐ 文本单元数（指针类型）
	ImageUnits *int `json:"image_units" db:"image_units"` // ⭐ 图片单元数（指针类型）
	TableUnits *int `json:"table_units" db:"table_units"` // ⭐ 表格单元数（指针类型）

	Metadata  string    `json:"metadata" db:"metadata"` // JSONB
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (*Document) TableName() string {
	return "documents"
}

// TableData 表格数据结构
type TableData struct {
	Caption string     `json:"caption"` // 表格标题
	Headers []string   `json:"headers"` // 列标题
	Rows    [][]string `json:"rows"`    // 数据行
	Summary string     `json:"summary"` // AI 生成的摘要
}

// ImageMetadata 图片元数据
type ImageMetadata struct {
	Width           int      `json:"width"`
	Height          int      `json:"height"`
	Format          string   `json:"format"`   // jpg, png, etc.
	Caption         string   `json:"caption"`  // 图片标题
	ObjectsDetected []string `json:"objects"`  // 检测到的对象
	OCRText         string   `json:"ocr_text"` // OCR 提取的文本
}

// HybridQuery 混合查询请求
type HybridQuery struct {
	Text           string                  `json:"text"`            // 查询文本
	QueryVector    []float32               `json:"-"`               // 查询向量
	ModalityPrefer map[ContentType]float64 `json:"modality_prefer"` // 模态偏好
	DocID          *int64                  `json:"doc_id"`          // 文档过滤
	AllowedTypes   []ContentType           `json:"allowed_types"`   // 允许的类型
	TimeRange      *TimeRange              `json:"time_range"`      // 时间范围
	TopK           int                     `json:"top_k"`           // 返回数量
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// RetrievalResult 检索结果
type RetrievalResult struct {
	Unit     *ContentUnit `json:"unit"`     // 内容单元
	Score    float64      `json:"score"`    // 综合得分
	Distance float64      `json:"distance"` // 向量距离
	Source   string       `json:"source"`   // 来源（vector/graph）
}

// MultimodalRAGResponse 多模态 RAG 响应
type MultimodalRAGResponse struct {
	Answer   string                 `json:"answer"`   // 文本答案
	Sources  []*RetrievalResult     `json:"sources"`  // 来源内容
	Images   []string               `json:"images"`   // 相关图片URL
	Tables   []*TableData           `json:"tables"`   // 相关表格
	Metadata map[string]interface{} `json:"metadata"` // 元数据
}
