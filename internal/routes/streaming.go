package routes

import (
	"go-gin-crud/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterStreamingRoutes 註冊串流相關路由
func RegisterStreamingRoutes(r *gin.Engine) {
	ctrl := controller.NewStreamingController()
	g := r.Group("/stream")
	{
		g.GET("/sse", ctrl.StreamSSE)           // Server-Sent Events
		g.GET("/chunked", ctrl.StreamChunked)   // Chunked NDJSON
		g.GET("/progress", ctrl.StreamProgress) // SSE 進度範例
	}
}
