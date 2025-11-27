package middleware

import (
	"go-gin-crud/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var accessKey = []byte("ACCESS_SECRET") // 與 service 中的 accessKey 保持一致

func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// 取得 Header 的 Authorization: Bearer <token>
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少 Token"})
			c.Abort()
			return
		}

		// 自動移除 Bearer
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// 檢查 token 是否在黑名單中
		isBlacklisted, err := authService.IsTokenBlacklisted(tokenString)
		if err != nil || isBlacklisted {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 已失效"})
			c.Abort()
			return
		}

		// 解析 Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return accessKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 無效"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", claims["user_id"])

		c.Next()
	}
}
