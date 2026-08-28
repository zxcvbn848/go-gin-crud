package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkerPool 工作池結構
type WorkerPool struct {
	workerCount int
	taskChan    chan Task
	resultChan  chan Result
	wg          sync.WaitGroup
}

// NewWorkerPool 創建新的工作池
func NewWorkerPool(workerCount int, taskChan chan Task, resultChan chan Result) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		taskChan:    taskChan,
		resultChan:  resultChan,
	}
}

// Start 啟動工作池
func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

// worker 工作協程
func (wp *WorkerPool) worker(ctx context.Context, workerID int) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			// Context 被取消，優雅退出
			fmt.Printf("Worker %d: 收到取消信號，正在退出...\n", workerID)
			return

		case task, ok := <-wp.taskChan:
			if !ok {
				// Channel 已關閉，沒有更多任務
				fmt.Printf("Worker %d: 任務通道已關閉，退出\n", workerID)
				return
			}

			// 處理任務
			result := wp.processTask(ctx, task, workerID)
			wp.resultChan <- result
		}
	}
}

// processTask 處理單個任務
func (wp *WorkerPool) processTask(ctx context.Context, task Task, workerID int) Result {
	startTime := time.Now()

	// 創建一個帶超時的 context（每個任務最多執行 3 秒）
	taskCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 模擬任務處理
	done := make(chan bool, 1)
	var err error

	go func() {
		// 模擬實際工作（可能是 I/O 操作、計算等）
		time.Sleep(task.Duration)
		done <- true
	}()

	select {
	case <-taskCtx.Done():
		// 任務超時或被取消
		err = taskCtx.Err()
		fmt.Printf("Worker %d: 任務 %d 超時或被取消 (%v)\n", workerID, task.ID, err)
		return Result{
			TaskID:   task.ID,
			Status:   "timeout",
			Duration: time.Since(startTime),
			Error:    err,
		}

	case <-done:
		// 任務完成
		fmt.Printf("Worker %d: 任務 %d 完成 (耗時: %v)\n", workerID, task.ID, time.Since(startTime))
		return Result{
			TaskID:   task.ID,
			Status:   "success",
			Duration: time.Since(startTime),
			Error:    nil,
		}
	}
}

// Wait 等待所有 worker 完成
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
	close(wp.resultChan)
}
