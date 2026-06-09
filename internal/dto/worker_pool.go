package dto

import "time"

// WorkerPoolTask 工作池任務
type WorkerPoolTask struct {
	ID        string                 `json:"id"`         // 任務 ID
	Type      string                 `json:"type"`       // 任務類型
	Data      map[string]interface{} `json:"data"`       // 任務數據
	Timeout   int                    `json:"timeout"`    // 超時時間（秒）
	Priority  int                    `json:"priority"`   // 優先級（數字越大優先級越高）
	CreatedAt time.Time              `json:"created_at"` // 創建時間
}

// WorkerPoolResult 工作池任務結果
type WorkerPoolResult struct {
	TaskID      string                 `json:"task_id"`      // 任務 ID
	Status      string                 `json:"status"`       // 狀態：success, failed, timeout, cancelled
	Result      map[string]interface{} `json:"result"`       // 執行結果
	Error       string                 `json:"error"`        // 錯誤訊息
	Duration    int64                  `json:"duration"`     // 執行時間（毫秒）
	WorkerID    int                    `json:"worker_id"`    // 處理該任務的 Worker ID
	CompletedAt time.Time              `json:"completed_at"` // 完成時間
}

// CreateWorkerPoolRequest 創建工作池請求
type CreateWorkerPoolRequest struct {
	PoolID      string `json:"pool_id" binding:"required"`                    // 工作池 ID
	WorkerCount int    `json:"worker_count" binding:"required,min=1,max=100"` // Worker 數量
	QueueSize   int    `json:"queue_size" binding:"min=1"`                    // 任務佇列大小（可選，預設無限制）
}

// SubmitTaskRequest 提交任務請求
type SubmitTaskRequest struct {
	PoolID   string                 `json:"pool_id" binding:"required"` // 工作池 ID
	TaskID   string                 `json:"task_id" binding:"required"` // 任務 ID
	Type     string                 `json:"type" binding:"required"`    // 任務類型
	Data     map[string]interface{} `json:"data"`                       // 任務數據
	Timeout  int                    `json:"timeout"`                    // 超時時間（秒，預設 30 秒）
	Priority int                    `json:"priority"`                   // 優先級（預設 0）
}

// SubmitTaskResponse 提交任務響應
type SubmitTaskResponse struct {
	TaskID      string    `json:"task_id"`
	PoolID      string    `json:"pool_id"`
	Status      string    `json:"status"` // submitted, rejected
	Message     string    `json:"message"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// GetResultRequest 獲取結果請求
type GetResultRequest struct {
	PoolID string `json:"pool_id" binding:"required"` // 工作池 ID
	TaskID string `json:"task_id" binding:"required"` // 任務 ID
}

// WorkerPoolStatus 工作池狀態
type WorkerPoolStatus struct {
	PoolID         string    `json:"pool_id"`         // 工作池 ID
	WorkerCount    int       `json:"worker_count"`    // Worker 數量
	ActiveWorkers  int       `json:"active_workers"`  // 活躍的 Worker 數量
	QueueSize      int       `json:"queue_size"`      // 當前佇列大小
	QueueCapacity  int       `json:"queue_capacity"`  // 佇列容量（-1 表示無限制）
	TotalTasks     int64     `json:"total_tasks"`     // 總任務數
	CompletedTasks int64     `json:"completed_tasks"` // 已完成任務數
	FailedTasks    int64     `json:"failed_tasks"`    // 失敗任務數
	IsRunning      bool      `json:"is_running"`      // 是否運行中
	CreatedAt      time.Time `json:"created_at"`      // 創建時間
}

// BatchSubmitTaskRequest 批量提交任務請求
type BatchSubmitTaskRequest struct {
	PoolID string              `json:"pool_id" binding:"required"`             // 工作池 ID
	Tasks  []SubmitTaskRequest `json:"tasks" binding:"required,min=1,max=100"` // 任務列表
}

// BatchSubmitTaskResponse 批量提交任務響應
type BatchSubmitTaskResponse struct {
	PoolID      string               `json:"pool_id"`
	TotalTasks  int                  `json:"total_tasks"`
	Submitted   int                  `json:"submitted"` // 成功提交的任務數
	Rejected    int                  `json:"rejected"`  // 被拒絕的任務數
	Results     []SubmitTaskResponse `json:"results"`   // 每個任務的提交結果
	SubmittedAt time.Time            `json:"submitted_at"`
}
