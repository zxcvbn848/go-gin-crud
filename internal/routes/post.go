package routes

import (
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/middleware"
	"go-gin-crud/internal/repository"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterPostRoutes(r *gin.Engine, authService service.AuthService) {
	// 初始化依賴
	postRepo := repository.NewPostRepository()
	postService := service.NewPostService(postRepo)
	postController := controller.NewPostController(postService)

	// 需要認證的路由組
	posts := r.Group("/posts")
	posts.Use(middleware.AuthMiddleware(authService))

	// 註冊路由
	posts.GET("", postController.GetPosts)
	posts.GET("/:id", postController.GetPost)

	adminPosts := posts.Group("")
	adminPosts.Use(middleware.RoleMiddleware("admin"))
	adminPosts.POST("", postController.CreatePost)
	adminPosts.PUT("/:id", postController.UpdatePost)
	adminPosts.DELETE("/:id", postController.DeletePost)
}

