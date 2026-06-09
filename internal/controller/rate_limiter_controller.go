package controller

import (
	"net/http"
	"sync"
	"time"

	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

// RateLimiterController 限流器控制器
type RateLimiterController struct {
	rateLimiterService service.RateLimiterService
}

// NewRateLimiterController 創建限流器控制器
func NewRateLimiterController(rateLimiterService service.RateLimiterService) *RateLimiterController {
	return &RateLimiterController{
		rateLimiterService: rateLimiterService,
	}
}

// GetStatus 獲取限流器狀態
// @Summary 獲取限流器狀態
// @Description 獲取指定 key 的限流器狀態
// @Tags rate-limiter
// @Param key query string true "限流鍵（如：IP、用戶ID等）"
// @Success 200 {object} dto.RateLimiterStatus
// @Router /api/rate-limiter/status [get]
func (ctrl *RateLimiterController) GetStatus(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 key 參數"})
		return
	}

	status := ctrl.rateLimiterService.GetStatus(key)
	c.JSON(http.StatusOK, status)
}

// SetConfig 設置限流器配置
// @Summary 設置限流器配置
// @Description 為指定 key 設置限流器配置
// @Tags rate-limiter
// @Param request body dto.RateLimiterConfig true "限流器配置"
// @Success 200 {object} dto.RateLimiterStatus
// @Router /api/rate-limiter/config [post]
func (ctrl *RateLimiterController) SetConfig(c *gin.Context) {
	var req dto.RateLimiterConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	window, err := time.ParseDuration(req.Window)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的時間窗口格式，請使用如 1s, 1m, 1h 等格式"})
		return
	}

	ctrl.rateLimiterService.SetConfig(req.Key, req.Limit, window)
	status := ctrl.rateLimiterService.GetStatus(req.Key)
	c.JSON(http.StatusOK, status)
}

// Reset 重置限流器
// @Summary 重置限流器
// @Description 重置指定 key 的限流器
// @Tags rate-limiter
// @Param key query string true "限流鍵"
// @Success 200 {object} map[string]interface{}
// @Router /api/rate-limiter/reset [post]
func (ctrl *RateLimiterController) Reset(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 key 參數"})
		return
	}

	ctrl.rateLimiterService.Reset(key)
	c.JSON(http.StatusOK, gin.H{"message": "限流器已重置", "key": key})
}

// GetStats 獲取統計資訊
// @Summary 獲取統計資訊
// @Description 獲取限流器的全域統計資訊
// @Tags rate-limiter
// @Success 200 {object} dto.RateLimiterStats
// @Router /api/rate-limiter/stats [get]
func (ctrl *RateLimiterController) GetStats(c *gin.Context) {
	stats := ctrl.rateLimiterService.GetStats()
	c.JSON(http.StatusOK, stats)
}

// TestRateLimiter 測試限流器
// @Summary 測試限流器
// @Description 併發測試限流器的功能
// @Tags rate-limiter
// @Param request body dto.RateLimiterTestRequest true "測試請求"
// @Success 200 {object} dto.RateLimiterTestResponse
// @Router /api/rate-limiter/test [post]
func (ctrl *RateLimiterController) TestRateLimiter(c *gin.Context) {
	var req dto.RateLimiterTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	window, err := time.ParseDuration(req.Window)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的時間窗口格式，請使用如 1s, 1m, 1h 等格式"})
		return
	}

	// 設置限流器配置
	ctrl.rateLimiterService.SetConfig(req.Key, req.Limit, window)

	// 設置併發數，預設為 1
	concurrent := req.Concurrent
	if concurrent <= 0 {
		concurrent = 1
	}
	if concurrent > 100 {
		concurrent = 100 // 限制最大併發數
	}

	startTime := time.Now()
	details := make([]dto.RateLimiterStatus, 0, req.Requests)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 計算每個 goroutine 需要處理的請求數
	requestsPerGoroutine := req.Requests / concurrent
	remainingRequests := req.Requests % concurrent

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			requests := requestsPerGoroutine
			if goroutineID < remainingRequests {
				requests++
			}

			for j := 0; j < requests; j++ {
				allowed, status := ctrl.rateLimiterService.Allow(req.Key)
				status.IsAllowed = allowed

				mu.Lock()
				details = append(details, *status)
				mu.Unlock()

				// 添加小延遲，模擬真實請求
				time.Sleep(time.Millisecond * 10)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 統計結果
	allowedCount := 0
	blockedCount := 0
	for _, detail := range details {
		if detail.IsAllowed {
			allowedCount++
		} else {
			blockedCount++
		}
	}

	successRate := float64(allowedCount) / float64(len(details)) * 100

	response := dto.RateLimiterTestResponse{
		TotalRequests:   len(details),
		AllowedRequests: allowedCount,
		BlockedRequests: blockedCount,
		SuccessRate:     successRate,
		Duration:        duration.String(),
		Details:         details,
	}

	c.JSON(http.StatusOK, response)
}

// TestAllow 測試單個請求是否允許
// @Summary 測試單個請求
// @Description 測試指定 key 的單個請求是否允許
// @Tags rate-limiter
// @Param key query string true "限流鍵"
// @Success 200 {object} dto.RateLimiterStatus
// @Router /api/rate-limiter/test/allow [get]
func (ctrl *RateLimiterController) TestAllow(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		// 如果沒有提供 key，使用 IP 地址
		key = c.ClientIP()
	}

	allowed, status := ctrl.rateLimiterService.Allow(key)
	status.IsAllowed = allowed

	if !allowed {
		c.JSON(http.StatusTooManyRequests, status)
		return
	}

	c.JSON(http.StatusOK, status)
}
