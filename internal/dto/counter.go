package dto

// CounterResponse 計數器回應
type CounterResponse struct {
	Value int64 `json:"value"`
}

// CounterIncrementRequest 增加計數請求
type CounterIncrementRequest struct {
	Amount int64 `json:"amount" binding:"required,min=1"`
}

// CounterDecrementRequest 減少計數請求
type CounterDecrementRequest struct {
	Amount int64 `json:"amount" binding:"required,min=1"`
}

// CounterSetRequest 設置計數值請求
type CounterSetRequest struct {
	Value int64 `json:"value" binding:"required"`
}

// CounterServiceInfo 計數器服務信息
type CounterServiceInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
