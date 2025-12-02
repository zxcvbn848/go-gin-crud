package routes

import (
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterTaskExecutorRoutes 註冊任務執行器相關路由
func RegisterTaskExecutorRoutes(r *gin.Engine) {
	taskExecutorService := service.NewTaskExecutorService()
	taskExecutorController := controller.NewTaskExecutorController(taskExecutorService)

	api := r.Group("/tasks")
	{
		// 執行單個任務（帶超時）
		api.POST("/execute", taskExecutorController.ExecuteTask)

		// 執行任務（帶重試機制）
		api.POST("/execute/retry", taskExecutorController.ExecuteTaskWithRetry)

		// 批量執行任務（並發）
		api.POST("/batch", taskExecutorController.BatchExecuteTasks)
	}
}
