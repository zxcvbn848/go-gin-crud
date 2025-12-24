package routes

import (
	"go-gin-crud/internal/controller"

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes 註冊健康檢查相關路由
func RegisterHealthRoutes(r *gin.Engine) {
	healthController := controller.NewHealthController()
	r.GET("/health", healthController.GetHealth)

	// Swagger 文檔路由（需要先安裝: go get -u github.com/swaggo/files github.com/swaggo/gin-swagger）
	// 取消註釋以下代碼以啟用 Swagger UI：
	/*
		swaggerFiles "github.com/swaggo/files"
		ginSwagger "github.com/swaggo/gin-swagger"
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	*/
}
