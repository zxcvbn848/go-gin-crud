package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetCounterValue 測試獲取計數值
func TestGetCounterValue(t *testing.T) {
	// 測試 mutex 計數器
	w := makeRequest("GET", "/counters?type=mutex", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]int64
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "value")

	// 測試 atomic 計數器
	w = makeRequest("GET", "/counters?type=atomic", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "value")
}

// TestIncrementCounter 測試增加計數
func TestIncrementCounter(t *testing.T) {
	// 測試 mutex 計數器
	req := map[string]int64{
		"amount": 10,
	}
	w := makeRequest("POST", "/counters/increment?type=mutex", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]int64
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, int64(10), response["value"])

	// 再次增加
	w = makeRequest("POST", "/counters/increment?type=mutex", req, "")
	assert.Equal(t, http.StatusOK, w.Code)
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, int64(20), response["value"])

	// 測試 atomic 計數器
	w = makeRequest("POST", "/counters/increment?type=atomic", req, "")
	assert.Equal(t, http.StatusOK, w.Code)
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, int64(10), response["value"])
}

// TestDecrementCounter 測試減少計數
func TestDecrementCounter(t *testing.T) {
	// 先設置一個值
	setReq := map[string]int64{
		"value": 100,
	}
	makeRequest("POST", "/counters/set?type=mutex", setReq, "")

	// 減少計數
	req := map[string]int64{
		"amount": 30,
	}
	w := makeRequest("POST", "/counters/decrement?type=mutex", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]int64
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, int64(70), response["value"])
}

// TestSetCounterValue 測試設置計數值
func TestSetCounterValue(t *testing.T) {
	req := map[string]int64{
		"value": 50,
	}
	w := makeRequest("POST", "/counters/set?type=mutex", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]int64
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, int64(50), response["value"])

	// 驗證值已設置
	w = makeRequest("GET", "/counters?type=mutex", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, int64(50), response["value"])
}

// TestResetCounter 測試重置計數器
func TestResetCounter(t *testing.T) {
	// 先設置一個值
	setReq := map[string]int64{
		"value": 100,
	}
	makeRequest("POST", "/counters/set?type=mutex", setReq, "")

	// 重置
	w := makeRequest("POST", "/counters/reset?type=mutex", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]int64
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, int64(0), response["value"])
}

// TestGetCounterInfo 測試獲取計數器資訊
func TestGetCounterInfo(t *testing.T) {
	w := makeRequest("GET", "/counters/info?type=mutex", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "mutex", response["type"])
	assert.NotEmpty(t, response["description"])

	w = makeRequest("GET", "/counters/info?type=atomic", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "atomic", response["type"])
}

// TestCounterPerformance 測試性能比較
func TestCounterPerformance(t *testing.T) {
	w := makeRequest("GET", "/counters/performance?iterations=100", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "mutex")
	assert.Contains(t, response, "atomic")
	assert.Contains(t, response, "winner")
}

