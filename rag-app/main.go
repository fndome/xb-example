package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// 初始化数据库
	db, err := sqlx.Connect("postgres", "postgres://user:password@localhost/rag_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 创建服务
	repo := NewChunkRepository(db)
	embedder := &MockEmbeddingService{}
	llm := &MockLLMService{}
	ragService := NewRAGService(repo, embedder, llm)

	// ⭐ 创建第三代 Agentic RAG 服务
	agenticService := NewAgenticRAGService(ragService)

	// ⭐ 创建 REFRAG 风格 RAG 服务
	refragService := NewREFRAGService(repo, embedder, llm)

	// 创建 HTTP 服务
	r := gin.Default()

	// 注册路由
	api := r.Group("/api")
	{
		api.POST("/documents", CreateDocumentHandler(ragService))
		api.POST("/rag/query", RAGQueryHandler(ragService, agenticService))
		api.POST("/rag/refrag", REFRAGQueryHandler(refragService)) // ⭐ REFRAG 查询
	}

	// 启动服务
	log.Println("RAG Server (v3 Agentic + REFRAG) starting on :8080")
	log.Println("Endpoints:")
	log.Println("  POST /api/documents - 上传文档")
	log.Println("  POST /api/rag/query - RAG 查询（默认使用第三代 Agentic RAG）")
	log.Println("  POST /api/rag/refrag - REFRAG 风格查询（压缩 + 智能选择）")
	
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

