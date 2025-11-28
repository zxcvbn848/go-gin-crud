package tests

import (
	"encoding/json"
	"go-gin-crud/internal/dto"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRegister 測試註冊功能
func TestRegister(t *testing.T) {
	// 測試成功註冊
	req := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	w := makeRequest("POST", "/register", req, "")
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "註冊成功", response["message"])

	// 測試重複註冊
	w = makeRequest("POST", "/register", req, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 測試無效的 email
	invalidReq := map[string]string{
		"email":    "invalid-email",
		"password": "password123",
	}
	w = makeRequest("POST", "/register", invalidReq, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 測試密碼太短
	shortPasswordReq := map[string]string{
		"email":    "test2@example.com",
		"password": "12345", // 少於 6 個字元
	}
	w = makeRequest("POST", "/register", shortPasswordReq, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestLogin 測試登入功能
func TestLogin(t *testing.T) {
	// 先註冊一個用戶
	registerReq := map[string]string{
		"email":    "login@example.com",
		"password": "password123",
	}
	makeRequest("POST", "/register", registerReq, "")

	// 測試成功登入
	loginReq := map[string]string{
		"email":    "login@example.com",
		"password": "password123",
	}
	w := makeRequest("POST", "/login", loginReq, "")
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["access_token"])
	assert.NotEmpty(t, response["refresh_token"])

	// 測試錯誤密碼
	wrongPasswordReq := map[string]string{
		"email":    "login@example.com",
		"password": "wrongpassword",
	}
	w = makeRequest("POST", "/login", wrongPasswordReq, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 測試不存在的用戶
	nonExistentReq := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}
	w = makeRequest("POST", "/login", nonExistentReq, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestProfile 測試取得用戶資料
func TestProfile(t *testing.T) {
	// 註冊並登入獲取 token
	token := registerTestUser(t, "profile@example.com", "password123")

	// 測試成功取得資料
	w := makeRequest("GET", "/auth/profile", nil, token)
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.UserResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "profile@example.com", response.Email)

	// 測試無 token
	w = makeRequest("GET", "/auth/profile", nil, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 測試無效 token
	w = makeRequest("GET", "/auth/profile", nil, "invalid_token")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestLogout 測試登出功能
func TestLogout(t *testing.T) {
	// 註冊並登入獲取 token
	token := registerTestUser(t, "logout@example.com", "password123")

	// 測試成功登出
	w := makeRequest("POST", "/auth/logout", nil, token)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "登出成功", response["message"])

	// 測試登出後無法使用 token
	w = makeRequest("GET", "/auth/profile", nil, token)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

