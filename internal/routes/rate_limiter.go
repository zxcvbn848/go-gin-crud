package routes

import (
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRateLimiterRoutes 註冊限流器相關路由
func RegisterRateLimiterRoutes(r *gin.Engine) {
	// 創建限流器服務（預設：100 請求/秒）
	rateLimiterService := service.NewRateLimiterService(100, time.Second)
	rateLimiterController := controller.NewRateLimiterController(rateLimiterService)

	api := r.Group("/api/rate-limiter")
	{
		// 獲取限流器狀態
		api.GET("/status", rateLimiterController.GetStatus)

		// 設置限流器配置
		api.POST("/config", rateLimiterController.SetConfig)

		// 重置限流器
		api.POST("/reset", rateLimiterController.Reset)

		// 獲取統計資訊
		api.GET("/stats", rateLimiterController.GetStats)

		// 測試限流器
		api.POST("/test", rateLimiterController.TestRateLimiter)

		// 測試單個請求
		api.GET("/test/allow", rateLimiterController.TestAllow)
	}
}

// GetRateLimiterService 獲取限流器服務實例（供其他路由使用）
func GetRateLimiterService() service.RateLimiterService {
	// 預設配置：100 請求/秒
	return service.NewRateLimiterService(100, time.Second)
}
