package routes

import (
	"go-gin-crud/internal/controller"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterHealthRoutes 註冊健康檢查相關路由
func RegisterHealthRoutes(r *gin.Engine) {
	healthController := controller.NewHealthController()
	r.GET("/health", healthController.GetHealth)

	// Swagger 文檔路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
