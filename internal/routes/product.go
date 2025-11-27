package routes

import (
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/middleware"
	"go-gin-crud/internal/repository"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterProductRoutes(r *gin.Engine, authService service.AuthService) {
	// 初始化依賴
	productRepo := repository.NewProductRepository()
	productService := service.NewProductService(productRepo)
	productController := controller.NewProductController(productService)

	// 需要認證的路由組
	products := r.Group("/products")
	products.Use(middleware.AuthMiddleware(authService))

	// 註冊路由
	products.POST("", productController.CreateProduct)
	products.GET("", productController.GetProducts)
	products.GET("/:id", productController.GetProduct)
	products.PUT("/:id", productController.UpdateProduct)
	products.DELETE("/:id", productController.DeleteProduct)
}

