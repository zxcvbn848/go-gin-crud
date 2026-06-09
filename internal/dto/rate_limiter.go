package dto

import "time"

// RateLimiterConfig 限流器配置
type RateLimiterConfig struct {
	Key       string `json:"key"`       // 限流鍵（如：IP、用戶ID等）
	Limit     int    `json:"limit"`     // 限制數量
	Window    string `json:"window"`    // 時間窗口（如：1s, 1m）
	Algorithm string `json:"algorithm"` // 演算法類型：sliding_window, fixed_window
}

// RateLimiterStatus 限流器狀態
type RateLimiterStatus struct {
	Key          string    `json:"key"`
	Limit        int       `json:"limit"`
	Window       string    `json:"window"`        // 時間窗口字串表示
	CurrentCount int       `json:"current_count"` // 當前窗口內的請求數
	Remaining    int       `json:"remaining"`     // 剩餘可用請求數
	ResetTime    time.Time `json:"reset_time"`    // 重置時間
	IsAllowed    bool      `json:"is_allowed"`    // 是否允許請求
	Algorithm    string    `json:"algorithm"`     // 演算法類型
}

// RateLimiterStats 限流器統計資訊
type RateLimiterStats struct {
	TotalRequests   int64     `json:"total_requests"`   // 總請求數
	AllowedRequests int64     `json:"allowed_requests"` // 允許的請求數
	BlockedRequests int64     `json:"blocked_requests"` // 被阻止的請求數
	ActiveKeys      int       `json:"active_keys"`      // 活躍的限流鍵數量
	LastResetTime   time.Time `json:"last_reset_time"`  // 最後重置時間
}

// RateLimiterTestRequest 限流器測試請求
type RateLimiterTestRequest struct {
	Key        string `json:"key" binding:"required"`            // 限流鍵
	Limit      int    `json:"limit" binding:"required,min=1"`    // 限制數量
	Window     string `json:"window" binding:"required"`         // 時間窗口（如：1s, 1m）
	Requests   int    `json:"requests" binding:"required,min=1"` // 測試請求數量
	Concurrent int    `json:"concurrent"`                        // 併發數（可選）
}

// RateLimiterTestResponse 限流器測試響應
type RateLimiterTestResponse struct {
	TotalRequests   int                 `json:"total_requests"`
	AllowedRequests int                 `json:"allowed_requests"`
	BlockedRequests int                 `json:"blocked_requests"`
	SuccessRate     float64             `json:"success_rate"` // 成功率
	Duration        string              `json:"duration"`     // 測試耗時
	Details         []RateLimiterStatus `json:"details"`      // 詳細結果
}
