package service

import (
	"context"
	"errors"
	"fmt"
	"go-gin-crud/internal/dto"
	"math/rand/v2"
	"sync"
	"time"
)

// maxBackoff 退避延遲上限，避免重試次數大時等到天荒地老
const maxBackoff = 30 * time.Second

var (
	ErrTaskTimeout   = errors.New("任務超時")
	ErrTaskCancelled = errors.New("任務被取消")
	ErrTaskFailed    = errors.New("任務執行失敗")
)

// TaskExecutorService 任務執行器服務介面
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
//
// 保證：永遠不會回傳 (nil, nil)，呼叫端可以無條件使用 response。
func (s *taskExecutorService) ExecuteTask(ctx context.Context, task dto.TaskRequest, timeout time.Duration) (*dto.TaskResponse, error) {
	startTime := time.Now()

	// taskCtx 同時涵蓋「自身超時」與「父 context 被取消」兩種結束條件
	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// done 只由工作 goroutine 關閉、只由這裡讀取，不會出現「已關閉 channel 永遠 ready」的問題
	done := make(chan struct{})
	go func() {
		// ponytail: 模擬實際工作。sleeper 最多殘留 task.Duration 就自行結束，是有界的，不需要額外的取消機制
		time.Sleep(time.Duration(task.Duration) * time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
		duration := time.Since(startTime)
		return &dto.TaskResponse{
			ID:          task.ID,
			Status:      "success",
			Duration:    duration.Milliseconds(),
			Message:     fmt.Sprintf("任務 %s 執行成功", task.ID),
			CompletedAt: time.Now(),
		}, nil

	case <-taskCtx.Done():
		duration := time.Since(startTime)

		// 父 context 已結束 → 是被取消，不是自身超時
		if ctx.Err() != nil {
			return &dto.TaskResponse{
				ID:       task.ID,
				Status:   "cancelled",
				Duration: duration.Milliseconds(),
				Message:  fmt.Sprintf("任務 %s 被父 context 取消", task.ID),
				Error:    ctx.Err().Error(),
			}, ErrTaskCancelled
		}

		return &dto.TaskResponse{
			ID:       task.ID,
			Status:   "timeout",
			Duration: duration.Milliseconds(),
			Message:  fmt.Sprintf("任務 %s 超時: %v", task.ID, taskCtx.Err()),
			Error:    taskCtx.Err().Error(),
		}, ErrTaskTimeout
	}
}

// backoffDelay 計算第 attempt 次重試前的等待時間：指數退避 + equal jitter。
//
// 純指數（base << attempt）的問題是多個 client 同時失敗會一起重試、一起再打爆
// 下游（thundering herd），所以一半固定、一半隨機打散，結果落在 [d/2, d)。
func backoffDelay(base time.Duration, attempt int) time.Duration {
	if base <= 0 {
		return 0
	}

	d := base << attempt
	// d <= 0 代表位移已溢位（attempt 過大）
	if d > maxBackoff || d <= 0 {
		d = maxBackoff
	}
	// d/2 為 0 時 rand.N 會 panic
	if d < 2 {
		return d
	}

	return d/2 + rand.N(d/2)
}

// ExecuteTaskWithRetry 執行任務（帶重試機制）
func (s *taskExecutorService) ExecuteTaskWithRetry(ctx context.Context, task dto.TaskRequest, timeout time.Duration, maxRetry int, retryDelay time.Duration) (*dto.TaskResponse, error) {
	var lastErr error
	var lastResponse *dto.TaskResponse

	// 負數會讓下面的迴圈完全不執行，導致 lastResponse 為 nil
	if maxRetry < 0 {
		maxRetry = 0
	}

	for attempt := 0; attempt <= maxRetry; attempt++ {
		// 檢查父 context 是否已取消
		if ctx.Err() != nil {
			return &dto.TaskResponse{
				ID:      task.ID,
				Status:  "cancelled",
				Message: "任務在重試過程中被取消",
				Error:   ctx.Err().Error(),
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
					ID:      task.ID,
					Status:  "cancelled",
					Message: "任務在重試等待過程中被取消",
					Error:   ctx.Err().Error(),
				}, ErrTaskCancelled
			case <-time.After(backoffDelay(retryDelay, attempt)):
				// 繼續重試
			}
		}
	}

	// 所有重試都失敗
	lastResponse.Message = fmt.Sprintf("任務 %s 在 %d 次嘗試後仍然失敗", task.ID, maxRetry+1)
	return lastResponse, lastErr
}

// BatchExecuteTasks 批量執行任務（並發）
//
// results 的緩衝等於 len(tasks)，每個 goroutine 必定送出剛好一筆結果，
// 因此送出永不阻塞、wg.Wait() 必定返回；整體耗時由 batchCtx 綁住。
func (s *taskExecutorService) BatchExecuteTasks(ctx context.Context, tasks []dto.TaskRequest, timeout time.Duration) (*dto.BatchExecuteResponse, error) {
	startTime := time.Now()

	// 整體超時
	batchCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan *dto.TaskResponse, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(t dto.TaskRequest) {
			defer wg.Done()
			result, _ := s.ExecuteTask(batchCtx, t, timeout)
			results <- result
		}(task)
	}

	wg.Wait()
	close(results)

	taskResults := make([]dto.TaskResponse, 0, len(tasks))
	successCount := 0
	timeoutCount := 0
	errorCount := 0

	for result := range results {
		taskResults = append(taskResults, *result)
		switch result.Status {
		case "success":
			successCount++
		case "timeout":
			timeoutCount++
		default:
			errorCount++
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
