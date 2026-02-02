package routes

import (
	"go-gin-crud/internal/cache"
	"go-gin-crud/internal/config"
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/logger"
	"go-gin-crud/internal/middleware"
	"go-gin-crud/internal/redis"
	"go-gin-crud/internal/repository"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterBookRoutes(r *gin.Engine, authService service.AuthService) {
	// 初始化依賴
	bookRepo := repository.NewBookRepository()

	var bookCache cache.BookCache
	if config.RedisAddr != "" {
		redisClient, err := redis.NewClient(config.RedisAddr)
		if err != nil {
			logger.Log.WithError(err).Warn("Redis 連線失敗，Book 快取停用")
		} else {
			bookCache = cache.NewBookCache(redisClient)
		}
	}

	bookService := service.NewBookService(bookRepo, bookCache)
	bookController := controller.NewBookController(bookService)

	// 需要認證的路由組
	books := r.Group("/books")
	books.Use(middleware.AuthMiddleware(authService))

	// 註冊路由
	books.GET("", bookController.GetBooks)
	books.GET("/:id", bookController.GetBook)

	adminBooks := books.Group("")
	adminBooks.Use(middleware.RoleMiddleware("admin"))
	adminBooks.POST("", bookController.CreateBook)
	adminBooks.PUT("/:id", bookController.UpdateBook)
	adminBooks.DELETE("/:id", bookController.DeleteBook)
}
