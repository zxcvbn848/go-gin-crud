package routes

import (
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/middleware"
	"go-gin-crud/internal/repository"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.Engine) service.AuthService {
	// 初始化依賴
	userRepo := repository.NewUserRepository()
	authRepo := repository.NewAuthRepository()
	authService := service.NewAuthService(userRepo, authRepo)
	authController := controller.NewAuthController(authService)

	// 註冊路由
	r.POST("/register", authController.Register)
	r.POST("/login", authController.Login)
	r.POST("/refresh", authController.Refresh)

	auth := r.Group("/auth")
	auth.Use(middleware.AuthMiddleware(authService))
	auth.POST("/logout", authController.Logout)
	auth.GET("/profile", authController.Profile)

	return authService
}
