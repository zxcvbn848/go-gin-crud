package routes

import (
	"go-gin-crud/database"
	"go-gin-crud/middleware"
	"go-gin-crud/models"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("secret_key_123") // 建議改成環境變數

func RegisterAuthRoutes(r *gin.Engine) {
	r.POST("/register", Register)
	r.POST("/login", Login)

	auth := r.Group("/auth")
	auth.Use(middleware.AuthMiddleware())
	auth.GET("/profile", Profile)
}

func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": "資料格式錯誤"})
		return
	}

	// 加密密碼
	hashed, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	user.Password = string(hashed)

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(400, gin.H{"error": "Email 已存在"})
		return
	}

	c.JSON(200, gin.H{"message": "註冊成功"})
}

func Login(c *gin.Context) {
	var req models.User
	var user models.User

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "資料格式錯誤"})
		return
	}

	// 找 Email
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "帳號或密碼錯誤"})
		return
	}

	// 檢查密碼
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{"error": "帳號或密碼錯誤"})
		return
	}

	// 簽 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, _ := token.SignedString(jwtKey)

	c.JSON(200, gin.H{
		"token": tokenString,
	})
}

func Profile(c *gin.Context) {
	userId, _ := c.Get("user_id")
	c.JSON(200, gin.H{
		"message": "你是合法使用者",
		"user_id": userId,
	})
}
