package service

import (
	"go-gin-crud/internal/dto"
	"sync"
	"sync/atomic"
	"time"
)

// RateLimiterService 限流器服務介面
type RateLimiterService interface {
	// Allow 檢查是否允許請求
	Allow(key string) (bool, *dto.RateLimiterStatus)
	// GetStatus 獲取限流器狀態
	GetStatus(key string) *dto.RateLimiterStatus
	// SetConfig 設置限流器配置
	SetConfig(key string, limit int, window time.Duration)
	// Reset 重置指定鍵的限流器
	Reset(key string)
	// GetStats 獲取統計資訊
	GetStats() *dto.RateLimiterStats
	// Cleanup 清理過期的限流器
	Cleanup()
}

// slidingWindowLimiter 滑動時間窗口限流器
type slidingWindowLimiter struct {
	limiters map[string]*limiterEntry
	mu       sync.RWMutex
	// 預設配置
	defaultLimit  int
	defaultWindow time.Duration
	// 統計資訊
	totalRequests   int64
	allowedRequests int64
	blockedRequests int64
	lastCleanup     time.Time
}

// limiterEntry 限流器條目
type limiterEntry struct {
	mu         sync.Mutex
	key        string
	limit      int
	window     time.Duration
	requests   []time.Time // 請求時間戳佇列
	lastAccess time.Time
	// 統計
	totalRequests   int64
	allowedRequests int64
	blockedRequests int64
}

// NewRateLimiterService 創建限流器服務
func NewRateLimiterService(defaultLimit int, defaultWindow time.Duration) RateLimiterService {
	rl := &slidingWindowLimiter{
		limiters:      make(map[string]*limiterEntry),
		defaultLimit:  defaultLimit,
		defaultWindow: defaultWindow,
		lastCleanup:   time.Now(),
	}

	// 啟動定期清理任務
	go rl.periodicCleanup()

	return rl
}

// Allow 檢查是否允許請求
func (rl *slidingWindowLimiter) Allow(key string) (bool, *dto.RateLimiterStatus) {
	atomic.AddInt64(&rl.totalRequests, 1)

	entry := rl.getOrCreateEntry(key)
	entry.mu.Lock()
	defer entry.mu.Unlock()

	now := time.Now()
	entry.lastAccess = now

	// 清理過期的請求記錄
	rl.cleanupExpiredRequests(entry, now)

	// 檢查是否超過限制
	allowed := len(entry.requests) < entry.limit

	if allowed {
		// 記錄請求時間
		entry.requests = append(entry.requests, now)
		atomic.AddInt64(&entry.allowedRequests, 1)
		atomic.AddInt64(&rl.allowedRequests, 1)
	} else {
		atomic.AddInt64(&entry.blockedRequests, 1)
		atomic.AddInt64(&rl.blockedRequests, 1)
	}

	atomic.AddInt64(&entry.totalRequests, 1)

	// 計算重置時間（最早請求的時間 + 窗口時間）
	var resetTime time.Time
	if len(entry.requests) > 0 {
		resetTime = entry.requests[0].Add(entry.window)
	} else {
		resetTime = now.Add(entry.window)
	}

	status := &dto.RateLimiterStatus{
		Key:          key,
		Limit:        entry.limit,
		Window:       entry.window.String(),
		CurrentCount: len(entry.requests),
		Remaining:    entry.limit - len(entry.requests),
		ResetTime:    resetTime,
		IsAllowed:    allowed,
		Algorithm:    "sliding_window",
	}

	return allowed, status
}

// GetStatus 獲取限流器狀態
func (rl *slidingWindowLimiter) GetStatus(key string) *dto.RateLimiterStatus {
	entry := rl.getEntry(key)
	if entry == nil {
		return &dto.RateLimiterStatus{
			Key:       key,
			Limit:     rl.defaultLimit,
			Window:    rl.defaultWindow.String(),
			Algorithm: "sliding_window",
		}
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	now := time.Now()
	rl.cleanupExpiredRequests(entry, now)

	var resetTime time.Time
	if len(entry.requests) > 0 {
		resetTime = entry.requests[0].Add(entry.window)
	} else {
		resetTime = now.Add(entry.window)
	}

	return &dto.RateLimiterStatus{
		Key:          key,
		Limit:        entry.limit,
		Window:       entry.window.String(),
		CurrentCount: len(entry.requests),
		Remaining:    entry.limit - len(entry.requests),
		ResetTime:    resetTime,
		IsAllowed:    len(entry.requests) < entry.limit,
		Algorithm:    "sliding_window",
	}
}

// SetConfig 設置限流器配置
func (rl *slidingWindowLimiter) SetConfig(key string, limit int, window time.Duration) {
	entry := rl.getOrCreateEntry(key)
	entry.mu.Lock()
	defer entry.mu.Unlock()

	entry.limit = limit
	entry.window = window
}

// Reset 重置指定鍵的限流器
func (rl *slidingWindowLimiter) Reset(key string) {
	entry := rl.getEntry(key)
	if entry == nil {
		return
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	entry.requests = entry.requests[:0]
}

// GetStats 獲取統計資訊
func (rl *slidingWindowLimiter) GetStats() *dto.RateLimiterStats {
	rl.mu.RLock()
	activeKeys := len(rl.limiters)
	rl.mu.RUnlock()

	return &dto.RateLimiterStats{
		TotalRequests:   atomic.LoadInt64(&rl.totalRequests),
		AllowedRequests: atomic.LoadInt64(&rl.allowedRequests),
		BlockedRequests: atomic.LoadInt64(&rl.blockedRequests),
		ActiveKeys:      activeKeys,
		LastResetTime:   rl.lastCleanup,
	}
}

// Cleanup 清理過期的限流器
func (rl *slidingWindowLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, entry := range rl.limiters {
		entry.mu.Lock()
		// 如果超過 1 小時沒有訪問，則刪除
		if now.Sub(entry.lastAccess) > time.Hour {
			expiredKeys = append(expiredKeys, key)
		}
		entry.mu.Unlock()
	}

	for _, key := range expiredKeys {
		delete(rl.limiters, key)
	}

	rl.lastCleanup = now
}

// periodicCleanup 定期清理
func (rl *slidingWindowLimiter) periodicCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.Cleanup()
	}
}

// getOrCreateEntry 獲取或創建限流器條目
func (rl *slidingWindowLimiter) getOrCreateEntry(key string) *limiterEntry {
	rl.mu.RLock()
	entry, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		return entry
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 雙重檢查
	if entry, exists := rl.limiters[key]; exists {
		return entry
	}

	entry = &limiterEntry{
		key:        key,
		limit:      rl.defaultLimit,
		window:     rl.defaultWindow,
		requests:   make([]time.Time, 0),
		lastAccess: time.Now(),
	}

	rl.limiters[key] = entry
	return entry
}

// getEntry 獲取限流器條目
func (rl *slidingWindowLimiter) getEntry(key string) *limiterEntry {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.limiters[key]
}

// cleanupExpiredRequests 清理過期的請求記錄
func (rl *slidingWindowLimiter) cleanupExpiredRequests(entry *limiterEntry, now time.Time) {
	cutoff := now.Add(-entry.window)
	validRequests := make([]time.Time, 0, len(entry.requests))

	for _, reqTime := range entry.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	entry.requests = validRequests
}
