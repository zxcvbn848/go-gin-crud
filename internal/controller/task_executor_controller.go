package controller

import (
	"net/http"
	"time"

	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

// TaskExecutorController 任務執行器控制器
type TaskExecutorController struct {
	taskExecutorService service.TaskExecutorService
}

// NewTaskExecutorController 創建任務執行器控制器
func NewTaskExecutorController(taskExecutorService service.TaskExecutorService) *TaskExecutorController {
	return &TaskExecutorController{
		taskExecutorService: taskExecutorService,
	}
}

// ExecuteTask 執行單個任務
// @Summary 執行單個任務（帶超時）
// @Description 執行一個任務，支援超時控制和資源清理
// @Tags task-executor
// @Param request body dto.ExecuteTaskRequest true "任務請求"
// @Success 200 {object} dto.ExecuteTaskResponse
// @Router /api/tasks/execute [post]
func (ctrl *TaskExecutorController) ExecuteTask(c *gin.Context) {
	var req dto.ExecuteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 創建 context（可以從請求中獲取，這裡使用 background）
	ctx := c.Request.Context()
	timeout := time.Duration(req.Timeout) * time.Second

	// 執行任務
	result, err := ctrl.taskExecutorService.ExecuteTask(ctx, req.Task, timeout)
	if err != nil {
		// 即使失敗也返回結果（包含錯誤資訊）
		if result != nil {
			c.JSON(http.StatusOK, dto.ExecuteTaskResponse{
				TaskID:      result.ID,
				Status:      result.Status,
				Duration:    result.Duration,
				Message:     result.Message,
				RetryCount:  0,
				CompletedAt: result.CompletedAt,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ExecuteTaskResponse{
		TaskID:      result.ID,
		Status:      result.Status,
		Duration:    result.Duration,
		Message:     result.Message,
		RetryCount:  0,
		CompletedAt: result.CompletedAt,
	})
}

// ExecuteTaskWithRetry 執行任務（帶重試）
// @Summary 執行任務（帶重試機制）
// @Description 執行任務，支援超時、重試和資源清理
// @Tags task-executor
// @Param request body dto.ExecuteTaskRequest true "任務請求（包含重試參數）"
// @Success 200 {object} dto.ExecuteTaskResponse
// @Router /api/tasks/execute/retry [post]
func (ctrl *TaskExecutorController) ExecuteTaskWithRetry(c *gin.Context) {
	var req dto.ExecuteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	timeout := time.Duration(req.Timeout) * time.Second
	retryDelay := time.Duration(req.RetryDelay) * time.Millisecond
	if retryDelay == 0 {
		retryDelay = 100 * time.Millisecond // 默認重試延遲
	}

	result, err := ctrl.taskExecutorService.ExecuteTaskWithRetry(ctx, req.Task, timeout, req.MaxRetry, retryDelay)
	if err != nil {
		if result != nil {
			c.JSON(http.StatusOK, dto.ExecuteTaskResponse{
				TaskID:      result.ID,
				Status:      result.Status,
				Duration:    result.Duration,
				Message:     result.Message,
				RetryCount:  req.MaxRetry + 1,
				CompletedAt: result.CompletedAt,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ExecuteTaskResponse{
		TaskID:      result.ID,
		Status:      result.Status,
		Duration:    result.Duration,
		Message:     result.Message,
		RetryCount:  0,
		CompletedAt: result.CompletedAt,
	})
}

// BatchExecuteTasks 批量執行任務
// @Summary 批量執行任務（並發）
// @Description 並發執行多個任務，支援整體超時控制
// @Tags task-executor
// @Param request body dto.BatchExecuteRequest true "批量任務請求"
// @Success 200 {object} dto.BatchExecuteResponse
// @Router /api/tasks/batch [post]
func (ctrl *TaskExecutorController) BatchExecuteTasks(c *gin.Context) {
	var req dto.BatchExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	timeout := time.Duration(req.Timeout) * time.Second

	result, err := ctrl.taskExecutorService.BatchExecuteTasks(ctx, req.Tasks, timeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
