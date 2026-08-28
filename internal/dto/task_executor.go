package dto

import "time"

// TaskRequest 任務請求
type TaskRequest struct {
	ID          string `json:"id" binding:"required"`
	Duration    int    `json:"duration"` // 任務執行時間（毫秒）
	Description string `json:"description"`
}

// TaskResponse 任務響應
type TaskResponse struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`   // "success", "timeout", "cancelled", "error"
	Duration    int64     `json:"duration"` // 實際執行時間（毫秒）
	Message     string    `json:"message"`
	Error       string    `json:"error,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// ExecuteTaskRequest 執行任務請求
type ExecuteTaskRequest struct {
	Task       TaskRequest `json:"task" binding:"required"`
	Timeout    int         `json:"timeout" binding:"required,min=1"` // 超時時間（秒）
	MaxRetry   int         `json:"max_retry"`                        // 最大重試次數
	RetryDelay int         `json:"retry_delay"`                      // 重試延遲（毫秒）
}

// ExecuteTaskResponse 執行任務響應
type ExecuteTaskResponse struct {
	TaskID      string    `json:"task_id"`
	Status      string    `json:"status"`
	Duration    int64     `json:"duration"` // 執行時間（毫秒）
	Message     string    `json:"message"`
	RetryCount  int       `json:"retry_count"`
	CompletedAt time.Time `json:"completed_at"`
}

// BatchExecuteRequest 批量執行請求
type BatchExecuteRequest struct {
	Tasks   []TaskRequest `json:"tasks" binding:"required"`
	Timeout int           `json:"timeout" binding:"required,min=1"` // 整體超時時間（秒）
}

// BatchExecuteResponse 批量執行響應
type BatchExecuteResponse struct {
	TotalTasks    int            `json:"total_tasks"`
	SuccessCount  int            `json:"success_count"`
	TimeoutCount  int            `json:"timeout_count"`
	ErrorCount    int            `json:"error_count"`
	TotalDuration int64          `json:"total_duration"` // 總執行時間（毫秒）
	Tasks         []TaskResponse `json:"tasks"`
}
