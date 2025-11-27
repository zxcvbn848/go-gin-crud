package routes

import (
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/repository"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterBookRoutes(r *gin.Engine) {
	// 初始化依賴
	bookRepo := repository.NewBookRepository()
	bookService := service.NewBookService(bookRepo)
	bookController := controller.NewBookController(bookService)

	// 註冊路由
	r.POST("/books", bookController.CreateBook)
	r.GET("/books", bookController.GetBooks)
	r.GET("/books/:id", bookController.GetBook)
	r.PUT("/books/:id", bookController.UpdateBook)
	r.DELETE("/books/:id", bookController.DeleteBook)
}
