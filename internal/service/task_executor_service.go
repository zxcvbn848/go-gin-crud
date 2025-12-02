package service

import (
	"context"
	"errors"
	"fmt"
	"go-gin-crud/internal/dto"
	"sync"
	"time"
)

var (
	ErrTaskTimeout   = errors.New("任務超時")
	ErrTaskCancelled = errors.New("任務被取消")
	ErrTaskFailed    = errors.New("任務執行失敗")
)

// TaskExecutorService 任務執行器服務接口
type TaskExecutorService interface {
	// ExecuteTask 執行單個任務（帶超時）
	ExecuteTask(ctx context.Context, task dto.TaskRequest, timeout time.Duration) (*dto.TaskResponse, error)
	// ExecuteTaskWithRetry 執行任務（帶重試機制）
	ExecuteTaskWithRetry(ctx context.Context, task dto.TaskRequest, timeout time.Duration, maxRetry int, retryDelay time.Duration) (*dto.TaskResponse, error)
	// BatchExecuteTasks 批量執行任務（並發）
	BatchExecuteTasks(ctx context.Context, tasks []dto.TaskRequest, timeout time.Duration) (*dto.BatchExecuteResponse, error)
}

// taskExecutorService 任務執行器服務實現
type taskExecutorService struct {
	// 可以添加資源池、連接池等
}

// NewTaskExecutorService 創建任務執行器服務
func NewTaskExecutorService() TaskExecutorService {
	return &taskExecutorService{}
}

// ExecuteTask 執行單個任務（帶超時）
func (s *taskExecutorService) ExecuteTask(ctx context.Context, task dto.TaskRequest, timeout time.Duration) (*dto.TaskResponse, error) {
	startTime := time.Now()

	// 創建帶超時的 context
	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // 確保資源清理

	// 創建結果 channel
	resultChan := make(chan *dto.TaskResponse, 1)
	errorChan := make(chan error, 1)

	// 啟動任務 goroutine
	go func() {
		defer func() {
			// 資源清理：關閉 channel
			close(resultChan)
			close(errorChan)
		}()

		// 模擬任務執行
		select {
		case <-taskCtx.Done():
			// Context 被取消或超時
			errorChan <- taskCtx.Err()
			return
		default:
			// 執行任務（模擬工作）
			done := make(chan bool, 1)
			go func() {
				// 模擬實際工作（將毫秒轉換為 Duration）
				time.Sleep(time.Duration(task.Duration) * time.Millisecond)
				done <- true
			}()

			select {
			case <-taskCtx.Done():
				// 超時或被取消
				errorChan <- taskCtx.Err()
				return
			case <-done:
				// 任務完成
				duration := time.Since(startTime)
				resultChan <- &dto.TaskResponse{
					ID:          task.ID,
					Status:      "success",
					Duration:    duration.Milliseconds(),
					Message:     fmt.Sprintf("任務 %s 執行成功", task.ID),
					CompletedAt: time.Now(),
				}
				return
			}
		}
	}()

	// 等待結果或超時
	select {
	case <-ctx.Done():
		// 父 context 被取消
		cancel() // 清理資源
		duration := time.Since(startTime)
		return &dto.TaskResponse{
			ID:       task.ID,
			Status:   "cancelled",
			Duration: duration.Milliseconds(),
			Message:  "任務被父 context 取消",
			Error:    ctx.Err().Error(),
		}, ErrTaskCancelled

	case result := <-resultChan:
		// 任務完成
		return result, nil

	case err := <-errorChan:
		// 任務失敗或超時
		status := "timeout"
		if errors.Is(err, context.Canceled) {
			status = "cancelled"
		}

		duration := time.Since(startTime)
		return &dto.TaskResponse{
			ID:       task.ID,
			Status:   status,
			Duration: duration.Milliseconds(),
			Message:  fmt.Sprintf("任務 %s 失敗: %v", task.ID, err),
			Error:    err.Error(),
		}, err

	case <-time.After(timeout + 100*time.Millisecond):
		// 額外的安全超時（防止 goroutine 泄漏）
		cancel()
		duration := time.Since(startTime)
		return &dto.TaskResponse{
			ID:       task.ID,
			Status:   "timeout",
			Duration: duration.Milliseconds(),
			Message:  "任務執行超時（安全機制觸發）",
			Error:    ErrTaskTimeout.Error(),
		}, ErrTaskTimeout
	}
}

// ExecuteTaskWithRetry 執行任務（帶重試機制）
func (s *taskExecutorService) ExecuteTaskWithRetry(ctx context.Context, task dto.TaskRequest, timeout time.Duration, maxRetry int, retryDelay time.Duration) (*dto.TaskResponse, error) {
	var lastErr error
	var lastResponse *dto.TaskResponse

	for attempt := 0; attempt <= maxRetry; attempt++ {
		// 檢查父 context 是否已取消
		if ctx.Err() != nil {
			return &dto.TaskResponse{
				ID:       task.ID,
				Status:   "cancelled",
				Message:  "任務在重試過程中被取消",
				Error:    ctx.Err().Error(),
			}, ErrTaskCancelled
		}

		// 執行任務
		response, err := s.ExecuteTask(ctx, task, timeout)

		if err == nil {
			// 成功
			response.Message = fmt.Sprintf("任務 %s 在第 %d 次嘗試時成功", task.ID, attempt+1)
			return response, nil
		}

		lastResponse = response
		lastErr = err

		// 如果不是最後一次嘗試，等待後重試
		if attempt < maxRetry {
			select {
			case <-ctx.Done():
				return &dto.TaskResponse{
					ID:       task.ID,
					Status:   "cancelled",
					Message:  "任務在重試等待過程中被取消",
					Error:    ctx.Err().Error(),
				}, ErrTaskCancelled
			case <-time.After(retryDelay):
				// 繼續重試
			}
		}
	}

	// 所有重試都失敗
	if lastResponse != nil {
		lastResponse.Message = fmt.Sprintf("任務 %s 在 %d 次嘗試後仍然失敗", task.ID, maxRetry+1)
		return lastResponse, lastErr
	}

	return &dto.TaskResponse{
		ID:       task.ID,
		Status:   "error",
		Message:  fmt.Sprintf("任務 %s 執行失敗", task.ID),
		Error:    lastErr.Error(),
	}, lastErr
}

// BatchExecuteTasks 批量執行任務（並發）
func (s *taskExecutorService) BatchExecuteTasks(ctx context.Context, tasks []dto.TaskRequest, timeout time.Duration) (*dto.BatchExecuteResponse, error) {
	startTime := time.Now()

	// 創建帶超時的 context（整體超時）
	batchCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // 確保資源清理

	// 使用 WaitGroup 管理 goroutine
	var wg sync.WaitGroup
	results := make(chan *dto.TaskResponse, len(tasks))

	// 並發執行所有任務
	for _, task := range tasks {
		wg.Add(1)
		go func(t dto.TaskRequest) {
			defer wg.Done()

			// 檢查整體超時
			if batchCtx.Err() != nil {
				results <- &dto.TaskResponse{
					ID:       t.ID,
					Status:   "cancelled",
					Message:  "批量任務整體超時",
					Error:    batchCtx.Err().Error(),
				}
				return
			}

			// 執行任務（每個任務有自己的超時時間）
			result, err := s.ExecuteTask(batchCtx, t, timeout)
			if err != nil {
				// 即使失敗也記錄結果
				if result == nil {
					result = &dto.TaskResponse{
						ID:       t.ID,
						Status:   "error",
						Message:  fmt.Sprintf("任務執行失敗: %v", err),
						Error:    err.Error(),
					}
				}
			}
			results <- result
		}(task)
	}

	// 等待所有任務完成或超時
	done := make(chan bool, 1)
	go func() {
		wg.Wait()
		close(results)
		done <- true
	}()

	// 收集結果
	taskResults := make([]dto.TaskResponse, 0, len(tasks))
	successCount := 0
	timeoutCount := 0
	errorCount := 0

	select {
	case <-batchCtx.Done():
		// 整體超時，取消所有任務
		cancel()
		// 收集已完成的任務
		for {
			select {
			case result := <-results:
				if result != nil {
					taskResults = append(taskResults, *result)
					if result.Status == "success" {
						successCount++
					} else if result.Status == "timeout" {
						timeoutCount++
					} else {
						errorCount++
					}
				}
			default:
				goto done
			}
		}
	done:
		// 為未完成的任務創建取消響應
		for _, task := range tasks {
			found := false
			for _, r := range taskResults {
				if r.ID == task.ID {
					found = true
					break
				}
			}
			if !found {
				taskResults = append(taskResults, dto.TaskResponse{
					ID:       task.ID,
					Status:   "cancelled",
					Message:  "批量任務整體超時，任務被取消",
					Error:    batchCtx.Err().Error(),
				})
				errorCount++
			}
		}

	case <-done:
		// 所有任務完成
		for result := range results {
			if result != nil {
				taskResults = append(taskResults, *result)
				if result.Status == "success" {
					successCount++
				} else if result.Status == "timeout" {
					timeoutCount++
				} else {
					errorCount++
				}
			}
		}
	}

	totalDuration := time.Since(startTime)
	return &dto.BatchExecuteResponse{
		TotalTasks:    len(tasks),
		SuccessCount:  successCount,
		TimeoutCount:  timeoutCount,
		ErrorCount:    errorCount,
		TotalDuration: totalDuration.Milliseconds(),
		Tasks:         taskResults,
	}, nil
}

