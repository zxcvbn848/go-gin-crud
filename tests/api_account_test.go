package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetAccountBalance 測試獲取帳戶餘額
func TestGetAccountBalance(t *testing.T) {
	w := makeRequest("GET", "/accounts/balance", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]int64
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "balance")
	// 初始餘額應該是 100（根據 routes/account.go 中的設置）
	assert.Equal(t, int64(100), response["balance"])
}

// TestDeposit 測試存款
func TestDeposit(t *testing.T) {
	req := map[string]int64{
		"amount": 50,
	}
	w := makeRequest("POST", "/accounts/deposit", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "deposit", response["operation"])
	assert.Equal(t, true, response["success"])
	assert.Contains(t, response, "before")
	assert.Contains(t, response, "after")
}

// TestWithdraw 測試取款
func TestWithdraw(t *testing.T) {
	// 先設置餘額
	setReq := map[string]int64{
		"balance": 200,
	}
	makeRequest("POST", "/accounts/balance", setReq, "")

	// 取款
	req := map[string]int64{
		"amount": 50,
	}
	w := makeRequest("POST", "/accounts/withdraw", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "withdraw", response["operation"])
	assert.Equal(t, true, response["success"])

	// 驗證餘額
	w = makeRequest("GET", "/accounts/balance", nil, "")
	var balanceResp map[string]int64
	json.Unmarshal(w.Body.Bytes(), &balanceResp)
	assert.Equal(t, int64(150), balanceResp["balance"])
}

// TestWithdrawInsufficientBalance 測試餘額不足
func TestWithdrawInsufficientBalance(t *testing.T) {
	// 設置較小的餘額
	setReq := map[string]int64{
		"balance": 10,
	}
	makeRequest("POST", "/accounts/balance", setReq, "")

	// 嘗試取款超過餘額
	req := map[string]int64{
		"amount": 100,
	}
	w := makeRequest("POST", "/accounts/withdraw", req, "")
	assert.Equal(t, http.StatusOK, w.Code) // 即使失敗也返回 200，但 success 為 false

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "withdraw", response["operation"])
	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["message"], "餘額不足")
}

// TestSetBalance 測試設置餘額
func TestSetBalance(t *testing.T) {
	req := map[string]int64{
		"balance": 500,
	}
	w := makeRequest("POST", "/accounts/balance", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]int64
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, int64(500), response["balance"])

	// 驗證餘額已設置
	w = makeRequest("GET", "/accounts/balance", nil, "")
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, int64(500), response["balance"])
}

// TestResetAccount 測試重置帳戶
func TestResetAccount(t *testing.T) {
	// 先設置一個值
	setReq := map[string]int64{
		"balance": 1000,
	}
	makeRequest("POST", "/accounts/balance", setReq, "")

	// 重置
	w := makeRequest("POST", "/accounts/reset", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]int64
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, int64(0), response["balance"])
}

// TestBatchTransactions 測試批量執行交易
func TestBatchTransactions(t *testing.T) {
	// 設置初始餘額
	setReq := map[string]int64{
		"balance": 100,
	}
	makeRequest("POST", "/accounts/balance", setReq, "")

	// 批量交易：1 (存款), -1 (取款), 0 (查詢)
	req := map[string]interface{}{
		"operations": []int{1, -1, 0, 2, -1},
	}
	w := makeRequest("POST", "/accounts/batch", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	// JSON 解析時數字會變成 float64
	assert.Equal(t, float64(100), response["initial_balance"])
	assert.Contains(t, response, "final_balance")
	assert.Contains(t, response, "transactions")

	// 驗證交易數量
	transactions := response["transactions"].([]interface{})
	assert.Equal(t, 5, len(transactions))
}

// TestRandomBatchTransactions 測試隨機批量交易
func TestRandomBatchTransactions(t *testing.T) {
	// 設置初始餘額
	setReq := map[string]int64{
		"balance": 100,
	}
	makeRequest("POST", "/accounts/balance", setReq, "")

	// 執行隨機批量交易
	w := makeRequest("POST", "/accounts/batch/random?count=5", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "initial_balance")
	assert.Contains(t, response, "final_balance")
	assert.Contains(t, response, "transactions")

	// 驗證交易數量
	transactions := response["transactions"].([]interface{})
	assert.Equal(t, 5, len(transactions))
}

// TestRandomBatchTransactionsWithDelay 測試帶延遲的隨機批量交易
func TestRandomBatchTransactionsWithDelay(t *testing.T) {
	// 設置初始餘額
	setReq := map[string]int64{
		"balance": 100,
	}
	makeRequest("POST", "/accounts/balance", setReq, "")

	// 執行帶延遲的隨機批量交易
	w := makeRequest("POST", "/accounts/batch/random?count=3&delay=true", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "transactions")
}

