package main

import (
	"go-gin-crud/database"
	"go-gin-crud/models"
	"go-gin-crud/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 連線 DB
	database.Connect()

	// 自動建表
	database.DB.AutoMigrate(&models.Book{})
	database.DB.AutoMigrate(&models.User{})

	// 註冊路由
	routes.RegisterBookRoutes(r)
	routes.RegisterAuthRoutes(r)

	r.Run(":8080")
}
