package controller

import (
	"net/http"
	"time"

	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

// WorkerPoolController 工作池控制器
type WorkerPoolController struct {
	workerPoolService service.WorkerPoolService
}

// NewWorkerPoolController 創建工作池控制器
func NewWorkerPoolController(workerPoolService service.WorkerPoolService) *WorkerPoolController {
	return &WorkerPoolController{
		workerPoolService: workerPoolService,
	}
}

// CreatePool 創建工作池
// @Summary 創建工作池
// @Description 創建一個新的工作池，用於處理任務
// @Tags worker-pool
// @Param request body dto.CreateWorkerPoolRequest true "創建工作池請求"
// @Success 200 {object} dto.WorkerPoolStatus
// @Router /api/worker-pool/create [post]
func (ctrl *WorkerPoolController) CreatePool(c *gin.Context) {
	var req dto.CreateWorkerPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 設置預設值
	queueSize := req.QueueSize
	if queueSize <= 0 {
		queueSize = 0 // 0 表示無限制
	}

	err := ctrl.workerPoolService.CreatePool(req.PoolID, req.WorkerCount, queueSize)
	if err != nil {
		if err == service.ErrPoolAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 獲取狀態
	status, err := ctrl.workerPoolService.GetStatus(req.PoolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// SubmitTask 提交任務
// @Summary 提交任務
// @Description 向工作池提交一個任務
// @Tags worker-pool
// @Param request body dto.SubmitTaskRequest true "提交任務請求"
// @Success 200 {object} dto.SubmitTaskResponse
// @Router /api/worker-pool/submit [post]
func (ctrl *WorkerPoolController) SubmitTask(c *gin.Context) {
	var req dto.SubmitTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 設置預設值
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30 // 預設 30 秒
	}

	task := dto.WorkerPoolTask{
		ID:        req.TaskID,
		Type:      req.Type,
		Data:      req.Data,
		Timeout:   timeout,
		Priority:  req.Priority,
		CreatedAt: time.Now(),
	}

	err := ctrl.workerPoolService.SubmitTask(req.PoolID, task)
	response := dto.SubmitTaskResponse{
		TaskID:      req.TaskID,
		PoolID:      req.PoolID,
		SubmittedAt: time.Now(),
	}

	if err != nil {
		if err == service.ErrQueueFull {
			response.Status = "rejected"
			response.Message = "任務佇列已滿"
			c.JSON(http.StatusTooManyRequests, response)
			return
		}
		if err == service.ErrPoolNotFound {
			response.Status = "rejected"
			response.Message = "工作池不存在"
			c.JSON(http.StatusNotFound, response)
			return
		}
		if err == service.ErrPoolNotRunning {
			response.Status = "rejected"
			response.Message = "工作池未運行"
			c.JSON(http.StatusBadRequest, response)
			return
		}
		response.Status = "rejected"
		response.Message = err.Error()
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response.Status = "submitted"
	response.Message = "任務已提交"
	c.JSON(http.StatusOK, response)
}

// GetResult 獲取任務結果
// @Summary 獲取任務結果
// @Description 獲取指定任務的執行結果
// @Tags worker-pool
// @Param pool_id query string true "工作池 ID"
// @Param task_id query string true "任務 ID"
// @Success 200 {object} dto.WorkerPoolResult
// @Router /api/worker-pool/result [get]
func (ctrl *WorkerPoolController) GetResult(c *gin.Context) {
	poolID := c.Query("pool_id")
	taskID := c.Query("task_id")

	if poolID == "" || taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 pool_id 或 task_id 參數"})
		return
	}

	result, err := ctrl.workerPoolService.GetResult(poolID, taskID)
	if err != nil {
		if err == service.ErrPoolNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStatus 獲取工作池狀態
// @Summary 獲取工作池狀態
// @Description 獲取指定工作池的狀態資訊
// @Tags worker-pool
// @Param pool_id query string true "工作池 ID"
// @Success 200 {object} dto.WorkerPoolStatus
// @Router /api/worker-pool/status [get]
func (ctrl *WorkerPoolController) GetStatus(c *gin.Context) {
	poolID := c.Query("pool_id")
	if poolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 pool_id 參數"})
		return
	}

	status, err := ctrl.workerPoolService.GetStatus(poolID)
	if err != nil {
		if err == service.ErrPoolNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Shutdown 優雅關閉工作池
// @Summary 優雅關閉工作池
// @Description 優雅關閉指定工作池，等待所有任務完成
// @Tags worker-pool
// @Param pool_id query string true "工作池 ID"
// @Param timeout query int false "超時時間（秒，預設 30 秒）"
// @Success 200 {object} map[string]interface{}
// @Router /api/worker-pool/shutdown [post]
func (ctrl *WorkerPoolController) Shutdown(c *gin.Context) {
	poolID := c.Query("pool_id")
	if poolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 pool_id 參數"})
		return
	}

	timeout := 30 * time.Second
	if timeoutStr := c.Query("timeout"); timeoutStr != "" {
		if timeoutSec, err := time.ParseDuration(timeoutStr + "s"); err == nil {
			timeout = timeoutSec
		}
	}

	err := ctrl.workerPoolService.Shutdown(poolID, timeout)
	if err != nil {
		if err == service.ErrPoolNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "工作池已優雅關閉",
		"pool_id": poolID,
	})
}

// ListPools 列出所有工作池
// @Summary 列出所有工作池
// @Description 列出所有已創建的工作池
// @Tags worker-pool
// @Success 200 {object} map[string]interface{}
// @Router /api/worker-pool/list [get]
func (ctrl *WorkerPoolController) ListPools(c *gin.Context) {
	pools := ctrl.workerPoolService.ListPools()
	c.JSON(http.StatusOK, gin.H{
		"pools": pools,
		"count": len(pools),
	})
}

// BatchSubmitTasks 批量提交任務
// @Summary 批量提交任務
// @Description 向工作池批量提交多個任務
// @Tags worker-pool
// @Param request body dto.BatchSubmitTaskRequest true "批量提交任務請求"
// @Success 200 {object} dto.BatchSubmitTaskResponse
// @Router /api/worker-pool/batch-submit [post]
func (ctrl *WorkerPoolController) BatchSubmitTasks(c *gin.Context) {
	var req dto.BatchSubmitTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 轉換為 WorkerPoolTask
	tasks := make([]dto.WorkerPoolTask, 0, len(req.Tasks))
	for _, taskReq := range req.Tasks {
		timeout := taskReq.Timeout
		if timeout <= 0 {
			timeout = 30
		}

		task := dto.WorkerPoolTask{
			ID:        taskReq.TaskID,
			Type:      taskReq.Type,
			Data:      taskReq.Data,
			Timeout:   timeout,
			Priority:  taskReq.Priority,
			CreatedAt: time.Now(),
		}
		tasks = append(tasks, task)
	}

	submitted, rejected, results := ctrl.workerPoolService.BatchSubmitTasks(req.PoolID, tasks)

	response := dto.BatchSubmitTaskResponse{
		PoolID:      req.PoolID,
		TotalTasks:  len(tasks),
		Submitted:   submitted,
		Rejected:    rejected,
		Results:     results,
		SubmittedAt: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}
