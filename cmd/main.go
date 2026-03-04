package main

import (
	"go-gin-crud/internal/config"
	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/logger"
	"go-gin-crud/internal/redis"
	"go-gin-crud/internal/routes"

	_ "go-gin-crud/docs" // swagger docs

	"github.com/gin-gonic/gin"
)

// @title           Go Gin CRUD API
// @version         1.0
// @description     這是一個使用 Gin 框架構建的 CRUD API 服務，包含用戶管理、書籍管理、產品管理、文章管理、認證、限流器、工作池等功能。
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 請在值前加上 "Bearer " 前綴，例如：Bearer {token}
func main() {
	r := gin.Default()

	// 載入配置（包含 JWT Secrets）
	config.Load()

	// 連線 DB
	database.Connect()

	// 自動遷移資料庫結構
	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Book{},
		&models.Product{},
		&models.Post{},
		&models.RefreshToken{},
		&models.BlacklistToken{},
	); err != nil {
		logger.Log.WithError(err).Fatal("資料庫遷移失敗")
	}
	logger.Log.Info("資料庫遷移完成")

	// Redis 連線
	var redisClient *redis.Client
	if config.RedisAddr != "" {
		var err error
		redisClient, err = redis.NewClient(config.RedisAddr)
		if err != nil {
			logger.Log.WithError(err).Warn("Redis 連線失敗，所有快取停用")
		}
	}

	// 註冊路由
	routes.RegisterHealthRoutes(r)
	authService := routes.RegisterAuthRoutes(r)
	routes.RegisterBookRoutes(r, authService, redisClient)
	routes.RegisterUserRoutes(r, authService, redisClient)
	routes.RegisterProductRoutes(r, authService, redisClient)
	routes.RegisterPostRoutes(r, authService, redisClient)
	routes.RegisterCounterRoutes(r)
	routes.RegisterAccountRoutes(r)
	routes.RegisterTaskExecutorRoutes(r)
	routes.RegisterRateLimiterRoutes(r)
	routes.RegisterWorkerPoolRoutes(r)
	routes.RegisterSocketRoutes(r)

	r.Run(":8080")
}
