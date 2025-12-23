package service

import (
	"errors"
	"fmt"
	"go-gin-crud/internal/dto"
	"sync"
	"time"
)

var (
	ErrInsufficientBalance = errors.New("餘額不足")
	ErrInvalidAmount       = errors.New("金額無效")
)

// AccountService 帳戶服務介面
type AccountService interface {
	// GetBalance 獲取餘額
	GetBalance() int64
	// Deposit 存款
	Deposit(amount int64) (*dto.TransactionResponse, error)
	// Withdraw 取款
	Withdraw(amount int64) (*dto.TransactionResponse, error)
	// SetBalance 設置餘額（用於初始化）
	SetBalance(balance int64)
	// Reset 重置帳戶
	Reset()
	// ExecuteBatchTransactions 批量執行交易
	ExecuteBatchTransactions(operations []int) (*dto.BatchTransactionResponse, error)
}

// accountService 帳戶服務實現
type accountService struct {
	balance int64
	mu      sync.Mutex
	results chan dto.TransactionResponse
}

// NewAccountService 創建帳戶服務
func NewAccountService(initialBalance int64) AccountService {
	return &accountService{
		balance: initialBalance,
		results: make(chan dto.TransactionResponse, 100), // 緩衝 channel
	}
}

// GetBalance 獲取餘額
func (s *accountService) GetBalance() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.balance
}

// Deposit 存款
func (s *accountService) Deposit(amount int64) (*dto.TransactionResponse, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	before := s.balance
	s.balance += amount
	after := s.balance

	return &dto.TransactionResponse{
		Operation: "deposit",
		Amount:    amount,
		Before:    before,
		After:     after,
		Message:   fmt.Sprintf("%d + %d = %d", before, amount, after),
		Success:   true,
	}, nil
}

// Withdraw 取款
func (s *accountService) Withdraw(amount int64) (*dto.TransactionResponse, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.balance < amount {
		return &dto.TransactionResponse{
			Operation: "withdraw",
			Amount:    amount,
			Before:    s.balance,
			After:     s.balance,
			Message:   fmt.Sprintf("餘額不足：嘗試取款 %d，但餘額只有 %d", amount, s.balance),
			Success:   false,
		}, ErrInsufficientBalance
	}

	before := s.balance
	s.balance -= amount
	after := s.balance

	return &dto.TransactionResponse{
		Operation: "withdraw",
		Amount:    amount,
		Before:    before,
		After:     after,
		Message:   fmt.Sprintf("%d - %d = %d", before, amount, after),
		Success:   true,
	}, nil
}

// SetBalance 設置餘額
func (s *accountService) SetBalance(balance int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.balance = balance
}

// Reset 重置帳戶
func (s *accountService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.balance = 0
}

// ExecuteBatchTransactions 批量執行交易（優化版本）
func (s *accountService) ExecuteBatchTransactions(operations []int) (*dto.BatchTransactionResponse, error) {
	initialBalance := s.GetBalance()
	transactions := make([]dto.TransactionResponse, 0, len(operations))

	// 使用 WaitGroup 等待所有 goroutine 完成
	var wg sync.WaitGroup
	results := make(chan dto.TransactionResponse, len(operations))

	// 並發執行所有操作
	for _, op := range operations {
		wg.Add(1)
		go func(operation int) {
			defer wg.Done()

			var resp dto.TransactionResponse

			switch {
			case operation == 0:
				// 查詢餘額
				balance := s.GetBalance()
				resp = dto.TransactionResponse{
					Operation: "balance",
					Amount:    0,
					Before:    balance,
					After:     balance,
					Message:   fmt.Sprintf("餘額: %d", balance),
					Success:   true,
				}
			case operation < 0:
				// 取款
				tx, err := s.Withdraw(int64(-operation))
				if err != nil {
					resp = *tx // 即使失敗也返回結果
				} else {
					resp = *tx
				}
			default:
				// 存款
				tx, err := s.Deposit(int64(operation))
				if err != nil {
					resp = dto.TransactionResponse{
						Operation: "deposit",
						Amount:    int64(operation),
						Before:    s.GetBalance(),
						After:     s.GetBalance(),
						Message:   fmt.Sprintf("存款失敗: %v", err),
						Success:   false,
					}
				} else {
					resp = *tx
				}
			}

			results <- resp
		}(op)
	}

	// 等待所有操作完成
	wg.Wait()
	close(results)

	// 收集結果（按執行順序）
	for resp := range results {
		transactions = append(transactions, resp)
	}

	finalBalance := s.GetBalance()

	return &dto.BatchTransactionResponse{
		InitialBalance: initialBalance,
		FinalBalance:   finalBalance,
		Transactions:   transactions,
	}, nil
}

// ExecuteBatchTransactionsWithDelay 批量執行交易（帶延遲版本，用於演示）
func (s *accountService) ExecuteBatchTransactionsWithDelay(operations []int, delay time.Duration) (*dto.BatchTransactionResponse, error) {
	initialBalance := s.GetBalance()
	transactions := make([]dto.TransactionResponse, 0, len(operations))

	var wg sync.WaitGroup
	results := make(chan dto.TransactionResponse, len(operations))

	for _, op := range operations {
		wg.Add(1)
		go func(operation int) {
			defer wg.Done()

			// 模擬處理延遲（用於演示併發問題）
			time.Sleep(delay)

			var resp dto.TransactionResponse

			switch {
			case operation == 0:
				balance := s.GetBalance()
				resp = dto.TransactionResponse{
					Operation: "balance",
					Amount:    0,
					Before:    balance,
					After:     balance,
					Message:   fmt.Sprintf("餘額: %d", balance),
					Success:   true,
				}
			case operation < 0:
				tx, err := s.Withdraw(int64(-operation))
				if err != nil {
					resp = *tx
				} else {
					resp = *tx
				}
			default:
				tx, err := s.Deposit(int64(operation))
				if err != nil {
					resp = dto.TransactionResponse{
						Operation: "deposit",
						Amount:    int64(operation),
						Before:    s.GetBalance(),
						After:     s.GetBalance(),
						Message:   fmt.Sprintf("存款失敗: %v", err),
						Success:   false,
					}
				} else {
					resp = *tx
				}
			}

			results <- resp
		}(op)
	}

	wg.Wait()
	close(results)

	for resp := range results {
		transactions = append(transactions, resp)
	}

	finalBalance := s.GetBalance()

	return &dto.BatchTransactionResponse{
		InitialBalance: initialBalance,
		FinalBalance:   finalBalance,
		Transactions:   transactions,
	}, nil
}
