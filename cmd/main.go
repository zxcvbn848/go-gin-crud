package main

import (
	"go-gin-crud/internal/config"
	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/logger"
	"go-gin-crud/internal/routes"

	"github.com/gin-gonic/gin"
)

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

	// 註冊路由
	routes.RegisterHealthRoutes(r)
	authService := routes.RegisterAuthRoutes(r)
	routes.RegisterBookRoutes(r, authService)
	routes.RegisterUserRoutes(r, authService)
	routes.RegisterProductRoutes(r, authService)
	routes.RegisterPostRoutes(r, authService)
	routes.RegisterCounterRoutes(r)

	r.Run(":8080")
}
