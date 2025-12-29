package controller

import (
	"errors"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"
	"go-gin-crud/internal/validator"
	"net/http"
	"strings"

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
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// Register 用戶註冊
// @Summary 用戶註冊
// @Description 註冊一個新用戶帳號
// @Tags auth
// @Param request body RegisterRequest true "註冊資訊"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Router /register [post]
func (ctrl *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleValidationError(c, err)
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

// Login 用戶登入
// @Summary 用戶登入
// @Description 使用 email 和密碼登入，獲取 access token 和 refresh token
// @Tags auth
// @Param request body dto.LoginRequest true "登入資訊"
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleValidationError(c, err)
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

// Refresh 刷新 Access Token
// @Summary 刷新 Access Token
// @Description 使用 refresh token 獲取新的 access token
// @Tags auth
// @Param refresh_token formData string true "Refresh Token"
// @Success 200 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /refresh [post]
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

// Logout 用戶登出
// @Summary 用戶登出
// @Description 將 access token 加入黑名單，使其失效
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Router /auth/logout [post]
func (ctrl *AuthController) Logout(c *gin.Context) {
	// 從 Header 取得 token
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 Token"})
		return
	}

	// 移除 Bearer 前綴
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// 將 token 加入黑名單
	if err := ctrl.authService.Logout(tokenString); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// Profile 獲取用戶資料
// @Summary 獲取用戶資料
// @Description 獲取當前登入用戶的詳細資訊
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} gin.H
// @Router /auth/profile [get]
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

	// 轉換為 UserResponse 以包含時間欄位
	userResponse := dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, userResponse)
}

// ChangePassword 修改密碼
// @Summary 修改密碼
// @Description 修改當前登入用戶的密碼
// @Tags auth
// @Security BearerAuth
// @Param request body dto.ChangePasswordRequest true "密碼修改資訊"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /auth/change-password [post]
func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "無法取得使用者資訊"})
		return
	}

	var userID uint
	switch v := userIDVal.(type) {
	case float64:
		userID = uint(v)
	case uint:
		userID = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "使用者 ID 格式錯誤"})
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleValidationError(c, err)
		return
	}

	if err := ctrl.authService.ChangePassword(userID, req); err != nil {
		if errors.Is(err, service.ErrInvalidCurrentPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "目前密碼不正確"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "修改失敗"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密碼已更新"})
}
