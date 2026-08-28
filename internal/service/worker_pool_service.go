package service

import (
	"context"
	"errors"
	"fmt"
	"go-gin-crud/internal/dto"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrPoolNotFound      = errors.New("工作池不存在")
	ErrPoolAlreadyExists = errors.New("工作池已存在")
	ErrPoolNotRunning    = errors.New("工作池未運行")
	ErrQueueFull         = errors.New("任務佇列已滿")
	ErrTaskNotFound      = errors.New("任務不存在")
)

// WorkerPoolService 工作池服務介面
type WorkerPoolService interface {
	// CreatePool 創建工作池
	CreatePool(poolID string, workerCount int, queueSize int) error
	// SubmitTask 提交任務
	SubmitTask(poolID string, task dto.WorkerPoolTask) error
	// GetResult 獲取任務結果
	GetResult(poolID string, taskID string) (*dto.WorkerPoolResult, error)
	// GetStatus 獲取工作池狀態
	GetStatus(poolID string) (*dto.WorkerPoolStatus, error)
	// Shutdown 優雅關閉工作池
	Shutdown(poolID string, timeout time.Duration) error
	// ListPools 列出所有工作池
	ListPools() []string
	// BatchSubmitTasks 批量提交任務
	BatchSubmitTasks(poolID string, tasks []dto.WorkerPoolTask) (int, int, []dto.SubmitTaskResponse)
}

// workerPool 工作池實現
type workerPool struct {
	poolID      string
	workerCount int
	taskChan    chan dto.WorkerPoolTask
	resultChan  chan dto.WorkerPoolResult
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex

	// 狀態統計
	activeWorkers  int64
	totalTasks     int64
	completedTasks int64
	failedTasks    int64
	queueSize      int
	queueCapacity  int
	isRunning      bool
	createdAt      time.Time

	// 結果存儲
	results   map[string]*dto.WorkerPoolResult
	resultsMu sync.RWMutex
}

// workerPoolManager 工作池管理器
type workerPoolManager struct {
	pools map[string]*workerPool
	mu    sync.RWMutex
}

var globalManager = &workerPoolManager{
	pools: make(map[string]*workerPool),
}

// NewWorkerPoolService 創建工作池服務
func NewWorkerPoolService() WorkerPoolService {
	return globalManager
}

// CreatePool 創建工作池
func (m *workerPoolManager) CreatePool(poolID string, workerCount int, queueSize int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.pools[poolID]; exists {
		return ErrPoolAlreadyExists
	}

	// 如果 queueSize <= 0，表示無限制
	if queueSize <= 0 {
		queueSize = 0 // 0 表示無緩衝 channel（無限制）
	}

	ctx, cancel := context.WithCancel(context.Background())
	pool := &workerPool{
		poolID:        poolID,
		workerCount:   workerCount,
		queueSize:     0,
		queueCapacity: queueSize,
		ctx:           ctx,
		cancel:        cancel,
		isRunning:     true,
		createdAt:     time.Now(),
		results:       make(map[string]*dto.WorkerPoolResult),
	}

	// 創建 channel
	if queueSize > 0 {
		pool.taskChan = make(chan dto.WorkerPoolTask, queueSize)
	} else {
		// 無緩衝 channel（無限制，但可能阻塞）
		pool.taskChan = make(chan dto.WorkerPoolTask)
	}
	pool.resultChan = make(chan dto.WorkerPoolResult, 100) // 結果 channel 有緩衝

	// 啟動 workers
	pool.startWorkers()

	// 啟動結果收集器
	pool.startResultCollector()

	m.pools[poolID] = pool
	return nil
}

// startWorkers 啟動所有 worker
func (p *workerPool) startWorkers() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker 工作協程
func (p *workerPool) worker(workerID int) {
	defer p.wg.Done()
	atomic.AddInt64(&p.activeWorkers, 1)
	defer atomic.AddInt64(&p.activeWorkers, -1)

	for {
		select {
		case <-p.ctx.Done():
			// Context 被取消，優雅退出
			return

		case task, ok := <-p.taskChan:
			if !ok {
				// Channel 已關閉，沒有更多任務
				return
			}

			// 處理任務
			result := p.processTask(task, workerID)
			p.resultChan <- result
		}
	}
}

// processTask 處理單個任務
func (p *workerPool) processTask(task dto.WorkerPoolTask, workerID int) dto.WorkerPoolResult {
	startTime := time.Now()
	atomic.AddInt64(&p.totalTasks, 1)

	// 設置超時時間（預設 30 秒）
	timeout := time.Duration(task.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	taskCtx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()

	// 模擬任務處理（實際應用中可以根據 task.Type 執行不同的邏輯）
	done := make(chan bool, 1)
	var err error
	var resultData map[string]interface{}

	go func() {
		// 模擬實際工作
		// 這裡可以根據 task.Type 執行不同的業務邏輯
		time.Sleep(time.Millisecond * 100) // 模擬處理時間

		// 模擬處理結果
		resultData = map[string]interface{}{
			"processed": true,
			"task_type": task.Type,
			"data":      task.Data,
		}

		done <- true
	}()

	select {
	case <-taskCtx.Done():
		// 任務超時或被取消
		err = taskCtx.Err()
		atomic.AddInt64(&p.failedTasks, 1)
		return dto.WorkerPoolResult{
			TaskID:      task.ID,
			Status:      "timeout",
			Result:      nil,
			Error:       err.Error(),
			Duration:    time.Since(startTime).Milliseconds(),
			WorkerID:    workerID,
			CompletedAt: time.Now(),
		}

	case <-done:
		// 任務完成
		atomic.AddInt64(&p.completedTasks, 1)
		return dto.WorkerPoolResult{
			TaskID:      task.ID,
			Status:      "success",
			Result:      resultData,
			Error:       "",
			Duration:    time.Since(startTime).Milliseconds(),
			WorkerID:    workerID,
			CompletedAt: time.Now(),
		}
	}
}

// startResultCollector 啟動結果收集器
func (p *workerPool) startResultCollector() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-p.ctx.Done():
				return
			case result, ok := <-p.resultChan:
				if !ok {
					return
				}
				// 存儲結果
				p.resultsMu.Lock()
				p.results[result.TaskID] = &result
				p.resultsMu.Unlock()
			}
		}
	}()
}

// SubmitTask 提交任務
func (m *workerPoolManager) SubmitTask(poolID string, task dto.WorkerPoolTask) error {
	m.mu.RLock()
	pool, exists := m.pools[poolID]
	m.mu.RUnlock()

	if !exists {
		return ErrPoolNotFound
	}

	if !pool.isRunning {
		return ErrPoolNotRunning
	}

	// 設置創建時間
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	// 嘗試提交任務（非阻塞）
	select {
	case pool.taskChan <- task:
		return nil
	default:
		// 佇列已滿
		return ErrQueueFull
	}
}

// GetResult 獲取任務結果
func (m *workerPoolManager) GetResult(poolID string, taskID string) (*dto.WorkerPoolResult, error) {
	m.mu.RLock()
	pool, exists := m.pools[poolID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrPoolNotFound
	}

	pool.resultsMu.RLock()
	defer pool.resultsMu.RUnlock()

	result, exists := pool.results[taskID]
	if !exists {
		return nil, ErrTaskNotFound
	}

	return result, nil
}

// GetStatus 獲取工作池狀態
func (m *workerPoolManager) GetStatus(poolID string) (*dto.WorkerPoolStatus, error) {
	m.mu.RLock()
	pool, exists := m.pools[poolID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrPoolNotFound
	}

	pool.mu.RLock()
	defer pool.mu.RUnlock()

	queueSize := len(pool.taskChan)
	queueCapacity := pool.queueCapacity
	if queueCapacity == 0 {
		queueCapacity = -1 // -1 表示無限制
	}

	return &dto.WorkerPoolStatus{
		PoolID:         pool.poolID,
		WorkerCount:    pool.workerCount,
		ActiveWorkers:  int(atomic.LoadInt64(&pool.activeWorkers)),
		QueueSize:      queueSize,
		QueueCapacity:  queueCapacity,
		TotalTasks:     atomic.LoadInt64(&pool.totalTasks),
		CompletedTasks: atomic.LoadInt64(&pool.completedTasks),
		FailedTasks:    atomic.LoadInt64(&pool.failedTasks),
		IsRunning:      pool.isRunning,
		CreatedAt:      pool.createdAt,
	}, nil
}

// Shutdown 優雅關閉工作池
func (m *workerPoolManager) Shutdown(poolID string, timeout time.Duration) error {
	m.mu.Lock()
	pool, exists := m.pools[poolID]
	if !exists {
		m.mu.Unlock()
		return ErrPoolNotFound
	}
	m.mu.Unlock()

	// 標記為停止
	pool.mu.Lock()
	pool.isRunning = false
	pool.mu.Unlock()

	// 發送取消信號
	pool.cancel()

	// 關閉任務 channel（不再接受新任務）
	close(pool.taskChan)

	// 等待所有 worker 完成，但有超時限制
	done := make(chan struct{})
	go func() {
		pool.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 所有 worker 已完成
		close(pool.resultChan)
	case <-time.After(timeout):
		// 超時，強制關閉
		close(pool.resultChan)
		return fmt.Errorf("工作池關閉超時（%v）", timeout)
	}

	// 從管理器中移除
	m.mu.Lock()
	delete(m.pools, poolID)
	m.mu.Unlock()

	return nil
}

// ListPools 列出所有工作池
func (m *workerPoolManager) ListPools() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pools := make([]string, 0, len(m.pools))
	for poolID := range m.pools {
		pools = append(pools, poolID)
	}
	return pools
}

// BatchSubmitTasks 批量提交任務
func (m *workerPoolManager) BatchSubmitTasks(poolID string, tasks []dto.WorkerPoolTask) (int, int, []dto.SubmitTaskResponse) {
	submitted := 0
	rejected := 0
	results := make([]dto.SubmitTaskResponse, 0, len(tasks))

	for _, task := range tasks {
		err := m.SubmitTask(poolID, task)
		response := dto.SubmitTaskResponse{
			TaskID:      task.ID,
			PoolID:      poolID,
			SubmittedAt: time.Now(),
		}

		if err != nil {
			rejected++
			response.Status = "rejected"
			response.Message = err.Error()
		} else {
			submitted++
			response.Status = "submitted"
			response.Message = "任務已提交"
		}

		results = append(results, response)
	}

	return submitted, rejected, results
}
