package routes

import (
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/middleware"
	"go-gin-crud/internal/repository"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.Engine, authService service.AuthService) {
	// 初始化依賴
	userRepo := repository.NewUserRepository()
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	// 需要認證的路由組
	users := r.Group("/users")
	users.Use(middleware.AuthMiddleware(authService))

	// 註冊路由
	users.POST("", userController.CreateUser)
	users.GET("", userController.GetUsers)
	users.GET("/:id", userController.GetUser)
	users.PUT("/:id", userController.UpdateUser)
	users.DELETE("/:id", userController.DeleteUser)
}
