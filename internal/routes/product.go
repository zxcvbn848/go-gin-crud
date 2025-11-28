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
	products.GET("", productController.GetProducts)
	products.GET("/:id", productController.GetProduct)

	adminProducts := products.Group("")
	adminProducts.Use(middleware.RoleMiddleware("admin"))
	adminProducts.POST("", productController.CreateProduct)
	adminProducts.PUT("/:id", productController.UpdateProduct)
	adminProducts.DELETE("/:id", productController.DeleteProduct)
}

