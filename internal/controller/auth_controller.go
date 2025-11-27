package controller

import (
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService service.AuthService
}

func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "資料格式錯誤"})
		return
	}

	if err := ctrl.authService.Register(req.Email, req.Password); err != nil {
		if err.Error() == "email 已存在" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email 已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "註冊失敗"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "註冊成功"})
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "資料格式錯誤"})
		return
	}

	accessToken, refreshToken, err := ctrl.authService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (ctrl *AuthController) Refresh(ctx *gin.Context) {
	refreshToken := ctx.PostForm("refresh_token")
	if refreshToken == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token 為必填"})
		return
	}

	token, err := ctrl.authService.Refresh(refreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"access_token": token})
}

func (ctrl *AuthController) Profile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "無法取得使用者資訊"})
		return
	}

	// JWT claims 中的數字會被解析為 float64
	var userIDUint uint
	switch v := userID.(type) {
	case float64:
		userIDUint = uint(v)
	case uint:
		userIDUint = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "使用者 ID 格式錯誤"})
		return
	}

	user, err := ctrl.authService.GetUserProfile(userIDUint)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到使用者"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "你是合法使用者",
		"user_id": user.ID,
		"email":   user.Email,
	})
}
