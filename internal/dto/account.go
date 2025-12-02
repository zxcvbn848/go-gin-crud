package dto

// AccountResponse 帳戶回應
type AccountResponse struct {
	Balance int64 `json:"balance"`
}

// DepositRequest 存款請求
type DepositRequest struct {
	Amount int64 `json:"amount" binding:"required,min=1"`
}

// WithdrawRequest 取款請求
type WithdrawRequest struct {
	Amount int64 `json:"amount" binding:"required,min=1"`
}

// TransactionResponse 交易回應
type TransactionResponse struct {
	Operation string `json:"operation"` // "deposit", "withdraw", "balance"
	Amount    int64  `json:"amount"`
	Before    int64  `json:"before"`
	After     int64  `json:"after"`
	Message   string `json:"message"`
	Success   bool   `json:"success"`
}

// BatchTransactionRequest 批量交易請求
type BatchTransactionRequest struct {
	Operations []int `json:"operations" binding:"required"` // -1: 取款, 0: 查詢, 1+: 存款
}

// BatchTransactionResponse 批量交易回應
type BatchTransactionResponse struct {
	InitialBalance int64                 `json:"initial_balance"`
	FinalBalance   int64                 `json:"final_balance"`
	Transactions   []TransactionResponse `json:"transactions"`
}
