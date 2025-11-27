package main

import (
	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 連線 DB
	database.Connect()

	// 自動建表
	database.DB.AutoMigrate(&models.Book{})
	database.DB.AutoMigrate(&models.User{})
	database.DB.AutoMigrate(&models.RefreshToken{})
	database.DB.AutoMigrate(&models.BlacklistToken{})
	database.DB.AutoMigrate(&models.Product{})

	// 註冊路由
	routes.RegisterBookRoutes(r)
	authService := routes.RegisterAuthRoutes(r)
	routes.RegisterUserRoutes(r, authService)
	routes.RegisterProductRoutes(r, authService)

	r.Run(":8080")
}
