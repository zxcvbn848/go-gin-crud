package routes

import (
	"go-gin-crud/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes 註冊健康檢查相關路由
func RegisterHealthRoutes(r *gin.Engine) {
	healthController := controller.NewHealthController()
	r.GET("/health", healthController.GetHealth)
}


