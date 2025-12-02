package controller

import (
	"net/http"
	"strconv"
	"time"

	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

// CounterController 計數器控制器
type CounterController struct {
	mutexCounter  service.CounterService
	atomicCounter service.CounterService
}

// NewCounterController 創建計數器控制器
func NewCounterController() *CounterController {
	return &CounterController{
		mutexCounter:  service.NewCounterService(service.CounterTypeMutex),
		atomicCounter: service.NewCounterService(service.CounterTypeAtomic),
	}
}

// getCounter 根據類型獲取對應的計數器服務
func (ctrl *CounterController) getCounter(counterType string) service.CounterService {
	if counterType == "atomic" {
		return ctrl.atomicCounter
	}
	return ctrl.mutexCounter
}

// GetValue 獲取計數值
// @Summary 獲取計數值
// @Description 獲取指定類型計數器的當前值
// @Tags counter
// @Param type query string false "計數器類型 (mutex/atomic)" default(mutex)
// @Success 200 {object} dto.CounterResponse
// @Router /api/counter [get]
func (ctrl *CounterController) GetValue(c *gin.Context) {
	counterType := c.DefaultQuery("type", "mutex")
	counter := ctrl.getCounter(counterType)

	value := counter.GetValue()
	c.JSON(http.StatusOK, dto.CounterResponse{
		Value: value,
	})
}

// Increment 增加計數
// @Summary 增加計數
// @Description 增加指定類型計數器的值
// @Tags counter
// @Param type query string false "計數器類型 (mutex/atomic)" default(mutex)
// @Param request body dto.CounterIncrementRequest true "增加數量"
// @Success 200 {object} dto.CounterResponse
// @Router /api/counter/increment [post]
func (ctrl *CounterController) Increment(c *gin.Context) {
	counterType := c.DefaultQuery("type", "mutex")
	counter := ctrl.getCounter(counterType)

	var req dto.CounterIncrementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value := counter.Increment(req.Amount)
	c.JSON(http.StatusOK, dto.CounterResponse{
		Value: value,
	})
}

// Decrement 減少計數
// @Summary 減少計數
// @Description 減少指定類型計數器的值
// @Tags counter
// @Param type query string false "計數器類型 (mutex/atomic)" default(mutex)
// @Param request body dto.CounterDecrementRequest true "減少數量"
// @Success 200 {object} dto.CounterResponse
// @Router /api/counter/decrement [post]
func (ctrl *CounterController) Decrement(c *gin.Context) {
	counterType := c.DefaultQuery("type", "mutex")
	counter := ctrl.getCounter(counterType)

	var req dto.CounterDecrementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value := counter.Decrement(req.Amount)
	c.JSON(http.StatusOK, dto.CounterResponse{
		Value: value,
	})
}

// SetValue 設置計數值
// @Summary 設置計數值
// @Description 設置指定類型計數器的值
// @Tags counter
// @Param type query string false "計數器類型 (mutex/atomic)" default(mutex)
// @Param request body dto.CounterSetRequest true "計數值"
// @Success 200 {object} dto.CounterResponse
// @Router /api/counter/set [post]
func (ctrl *CounterController) SetValue(c *gin.Context) {
	counterType := c.DefaultQuery("type", "mutex")
	counter := ctrl.getCounter(counterType)

	var req dto.CounterSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value := counter.SetValue(req.Value)
	c.JSON(http.StatusOK, dto.CounterResponse{
		Value: value,
	})
}

// Reset 重置計數器
// @Summary 重置計數器
// @Description 重置指定類型計數器的值為 0
// @Tags counter
// @Param type query string false "計數器類型 (mutex/atomic)" default(mutex)
// @Success 200 {object} dto.CounterResponse
// @Router /api/counter/reset [post]
func (ctrl *CounterController) Reset(c *gin.Context) {
	counterType := c.DefaultQuery("type", "mutex")
	counter := ctrl.getCounter(counterType)

	value := counter.Reset()
	c.JSON(http.StatusOK, dto.CounterResponse{
		Value: value,
	})
}

// GetInfo 獲取計數器服務信息
// @Summary 獲取計數器服務信息
// @Description 獲取計數器服務的類型和描述
// @Tags counter
// @Param type query string false "計數器類型 (mutex/atomic)" default(mutex)
// @Success 200 {object} dto.CounterServiceInfo
// @Router /api/counter/info [get]
func (ctrl *CounterController) GetInfo(c *gin.Context) {
	counterType := c.DefaultQuery("type", "mutex")

	var serviceType service.CounterServiceType
	if counterType == "atomic" {
		serviceType = service.CounterTypeAtomic
	} else {
		serviceType = service.CounterTypeMutex
	}

	info := service.GetCounterServiceInfo(serviceType)
	c.JSON(http.StatusOK, info)
}

// ComparePerformance 性能比較（用於測試）
// @Summary 性能比較
// @Description 比較 Mutex 和 Atomic 實現的性能
// @Tags counter
// @Param iterations query int false "迭代次數" default(1000)
// @Success 200 {object} map[string]interface{}
// @Router /api/counter/performance [get]
func (ctrl *CounterController) ComparePerformance(c *gin.Context) {
	iterationsStr := c.DefaultQuery("iterations", "1000")
	iterations, err := strconv.Atoi(iterationsStr)
	if err != nil || iterations <= 0 {
		iterations = 1000
	}

	// 測試 Mutex 性能
	mutexStart := time.Now()
	for i := 0; i < iterations; i++ {
		ctrl.mutexCounter.Increment(1)
	}
	mutexDuration := time.Since(mutexStart)

	// 重置
	ctrl.mutexCounter.Reset()

	// 測試 Atomic 性能
	atomicStart := time.Now()
	for i := 0; i < iterations; i++ {
		ctrl.atomicCounter.Increment(1)
	}
	atomicDuration := time.Since(atomicStart)

	// 重置
	ctrl.atomicCounter.Reset()

	c.JSON(http.StatusOK, gin.H{
		"iterations": iterations,
		"mutex": gin.H{
			"duration_ms": mutexDuration.Milliseconds(),
			"ops_per_sec": float64(iterations) / mutexDuration.Seconds(),
		},
		"atomic": gin.H{
			"duration_ms": atomicDuration.Milliseconds(),
			"ops_per_sec": float64(iterations) / atomicDuration.Seconds(),
		},
		"winner": func() string {
			if atomicDuration < mutexDuration {
				return "atomic"
			}
			return "mutex"
		}(),
	})
}
