package routes

import (
	"go-gin-crud/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterCounterRoutes 註冊計數器相關路由
func RegisterCounterRoutes(r *gin.Engine) {
	counterController := controller.NewCounterController()

	api := r.Group("/counters")
	{
		// 獲取計數值
		api.GET("", counterController.GetValue)

		// 增加計數
		api.POST("/increment", counterController.Increment)

		// 減少計數
		api.POST("/decrement", counterController.Decrement)

		// 設置計數值
		api.POST("/set", counterController.SetValue)

		// 重置計數器
		api.POST("/reset", counterController.Reset)

		// 獲取服務信息
		api.GET("/info", counterController.GetInfo)

		// 性能比較
		api.GET("/performance", counterController.ComparePerformance)
	}
}
