package routes

import (
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/middleware"
	"go-gin-crud/internal/repository"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterBookRoutes(r *gin.Engine, authService service.AuthService) {
	// 初始化依賴
	bookRepo := repository.NewBookRepository()
	bookService := service.NewBookService(bookRepo)
	bookController := controller.NewBookController(bookService)

	// 需要認證的路由組
	books := r.Group("/books")
	books.Use(middleware.AuthMiddleware(authService))

	// 註冊路由
	books.POST("", bookController.CreateBook)
	books.GET("", bookController.GetBooks)
	books.GET("/:id", bookController.GetBook)
	books.PUT("/:id", bookController.UpdateBook)
	books.DELETE("/:id", bookController.DeleteBook)
}
