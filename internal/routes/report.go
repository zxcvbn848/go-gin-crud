package routes

import (
	"go-gin-crud/internal/cache"
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/middleware"
	"go-gin-crud/internal/redis"
	"go-gin-crud/internal/repository"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterReportRoutes 註冊報表相關路由。
//
// 全部限管理員：報表會揭露平台整體的營運數字，不該對一般使用者開放。
func RegisterReportRoutes(r *gin.Engine, authService service.AuthService, redisClient *redis.Client) {
	// 報表跨多個 model，所以注入四個 repository —— 查詢留在各自的 repo，
	// 組裝在 service，沒有 ReportRepository
	reportService := service.NewReportService(
		repository.NewUserRepository(),
		repository.NewPostRepository(),
		repository.NewProductRepository(),
		repository.NewBookRepository(),
		cache.NewReportCache(redisClient),
	)
	reportController := controller.NewReportController(reportService)

	reports := r.Group("/reports")
	reports.Use(middleware.AuthMiddleware(authService))
	reports.Use(middleware.RoleMiddleware("admin"))

	reports.GET("/overview", reportController.GetOverview)
	reports.GET("/daily", reportController.GetDaily)
	reports.GET("/authors", reportController.GetTopAuthors)
}
