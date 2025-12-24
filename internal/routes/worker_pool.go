package routes

import (
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterWorkerPoolRoutes 註冊工作池相關路由
func RegisterWorkerPoolRoutes(r *gin.Engine) {
	// 創建工作池服務
	workerPoolService := service.NewWorkerPoolService()
	workerPoolController := controller.NewWorkerPoolController(workerPoolService)

	api := r.Group("/api/worker-pool")
	{
		// 創建工作池
		api.POST("/create", workerPoolController.CreatePool)

		// 提交任務
		api.POST("/submit", workerPoolController.SubmitTask)

		// 批量提交任務
		api.POST("/batch-submit", workerPoolController.BatchSubmitTasks)

		// 獲取任務結果
		api.GET("/result", workerPoolController.GetResult)

		// 獲取工作池狀態
		api.GET("/status", workerPoolController.GetStatus)

		// 列出所有工作池
		api.GET("/list", workerPoolController.ListPools)

		// 優雅關閉工作池
		api.POST("/shutdown", workerPoolController.Shutdown)
	}
}

