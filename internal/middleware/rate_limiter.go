package middleware

import (
	"fmt"
	"go-gin-crud/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiterMiddleware 限流器中介層
// 支援按 IP 或用戶 ID 進行限流
func RateLimiterMiddleware(rateLimiterService service.RateLimiterService, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if key == "" {
			// 如果無法獲取 key，則允許通過
			c.Next()
			return
		}

		allowed, status := rateLimiterService.Allow(key)
		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "請求過於頻繁，請稍後再試",
				"status":      status,
				"retry_after": time.Until(status.ResetTime).Seconds(),
			})
			c.Abort()
			return
		}

		// 將限流狀態資訊添加到響應頭
		c.Header("X-RateLimit-Limit", strconv.Itoa(status.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(status.Remaining))
		c.Header("X-RateLimit-Reset", status.ResetTime.Format(time.RFC3339))

		c.Next()
	}
}

// RateLimiterByIP 按 IP 地址限流
func RateLimiterByIP(rateLimiterService service.RateLimiterService) gin.HandlerFunc {
	return RateLimiterMiddleware(rateLimiterService, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

// RateLimiterByUserID 按用戶 ID 限流
func RateLimiterByUserID(rateLimiterService service.RateLimiterService) gin.HandlerFunc {
	return RateLimiterMiddleware(rateLimiterService, func(c *gin.Context) string {
		userID, exists := c.Get("user_id")
		if !exists {
			return ""
		}

		userIDStr, ok := userID.(string)
		if !ok {
			// 嘗試轉換為 float64（JWT claims 可能是 float64）
			if userIDFloat, ok := userID.(float64); ok {
				return fmt.Sprintf("%.0f", userIDFloat)
			}
			// 嘗試轉換為 int（JWT claims 可能是 int）
			if userIDInt, ok := userID.(int); ok {
				return strconv.Itoa(userIDInt)
			}
			return ""
		}

		return userIDStr
	})
}

// RateLimiterByKey 按自訂 key 限流
func RateLimiterByKey(rateLimiterService service.RateLimiterService, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return RateLimiterMiddleware(rateLimiterService, keyFunc)
}
