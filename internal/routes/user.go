package routes

import (
	"go-gin-crud/internal/cache"
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/middleware"
	"go-gin-crud/internal/redis"
	"go-gin-crud/internal/repository"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.Engine, authService service.AuthService, redisClient *redis.Client) {
	// 初始化依賴
	userRepo := repository.NewUserRepository()
	userCache := cache.NewUserCache(redisClient)
	userService := service.NewUserService(userRepo, userCache)
	userController := controller.NewUserController(userService)

	// 需要認證的路由組
	users := r.Group("/users")
	users.Use(middleware.AuthMiddleware(authService))

	// 需要 admin 權限的路由組
	adminUsers := users.Group("")
	adminUsers.Use(middleware.RoleMiddleware("admin"))

	// 註冊路由（所有操作都需要 admin 權限）
	adminUsers.GET("", userController.GetUsers)
	adminUsers.GET("/:id", userController.GetUser)
	adminUsers.POST("", userController.CreateUser)
	adminUsers.PUT("/:id", userController.UpdateUser)
	adminUsers.DELETE("/:id", userController.DeleteUser)
}
