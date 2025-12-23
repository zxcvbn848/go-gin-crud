package service

import (
	"go-gin-crud/internal/dto"
	"sync"
	"sync/atomic"
)

// CounterService 計數器服務介面
type CounterService interface {
	// GetValue 獲取當前計數值
	GetValue() int64
	// Increment 增加計數
	Increment(amount int64) int64
	// Decrement 減少計數
	Decrement(amount int64) int64
	// SetValue 設置計數值
	SetValue(value int64) int64
	// Reset 重置計數器
	Reset() int64
}

// CounterServiceType 計數器實現類型
type CounterServiceType string

const (
	// CounterTypeMutex 使用 Mutex 實現
	CounterTypeMutex CounterServiceType = "mutex"
	// CounterTypeAtomic 使用 atomic 實現
	CounterTypeAtomic CounterServiceType = "atomic"
)

// NewCounterService 創建計數器服務
func NewCounterService(counterType CounterServiceType) CounterService {
	switch counterType {
	case CounterTypeAtomic:
		return NewAtomicCounter()
	case CounterTypeMutex:
		fallthrough
	default:
		return NewMutexCounter()
	}
}

// ==================== Mutex 實現 ====================

// mutexCounter 使用 sync.Mutex 實現的計數器
type mutexCounter struct {
	value int64
	mu    sync.Mutex
}

// NewMutexCounter 創建使用 Mutex 的計數器
func NewMutexCounter() CounterService {
	return &mutexCounter{
		value: 0,
	}
}

func (c *mutexCounter) GetValue() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func (c *mutexCounter) Increment(amount int64) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += amount
	return c.value
}

func (c *mutexCounter) Decrement(amount int64) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value -= amount
	return c.value
}

func (c *mutexCounter) SetValue(value int64) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = value
	return c.value
}

func (c *mutexCounter) Reset() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = 0
	return c.value
}

// ==================== Atomic 實現 ====================

// atomicCounter 使用 atomic 實現的計數器
type atomicCounter struct {
	value int64
}

// NewAtomicCounter 創建使用 atomic 的計數器
func NewAtomicCounter() CounterService {
	return &atomicCounter{
		value: 0,
	}
}

func (c *atomicCounter) GetValue() int64 {
	return atomic.LoadInt64(&c.value)
}

func (c *atomicCounter) Increment(amount int64) int64 {
	return atomic.AddInt64(&c.value, amount)
}

func (c *atomicCounter) Decrement(amount int64) int64 {
	return atomic.AddInt64(&c.value, -amount)
}

func (c *atomicCounter) SetValue(value int64) int64 {
	atomic.StoreInt64(&c.value, value)
	return value
}

func (c *atomicCounter) Reset() int64 {
	atomic.StoreInt64(&c.value, 0)
	return 0
}

// ==================== 輔助函數 ====================

// GetCounterServiceInfo 獲取計數器服務資訊
func GetCounterServiceInfo(counterType CounterServiceType) dto.CounterServiceInfo {
	return dto.CounterServiceInfo{
		Type:        string(counterType),
		Description: getCounterTypeDescription(counterType),
	}
}

func getCounterTypeDescription(counterType CounterServiceType) string {
	switch counterType {
	case CounterTypeMutex:
		return "使用 sync.Mutex 實現，適合需要複雜操作的場景"
	case CounterTypeAtomic:
		return "使用 atomic 操作實現，性能更高，適合簡單的計數操作"
	default:
		return "未知類型"
	}
}
