package tests

import (
	"encoding/json"
	"fmt"
	"go-gin-crud/internal/dto"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreateUser 測試創建用戶
func TestCreateUser(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 測試成功創建
	req := map[string]string{
		"email":    "newuser@example.com",
		"password": "password123",
		"role":     "user",
	}
	w := makeRequest("POST", "/users", req, adminToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var user dto.UserResponse
	_ = json.Unmarshal(w.Body.Bytes(), &user)
	assert.Equal(t, "newuser@example.com", user.Email)
	assert.Equal(t, "user", user.Role)

	// 測試創建管理員
	adminReq := map[string]string{
		"email":    "newadmin@example.com",
		"password": "password123",
		"role":     "admin",
	}
	w = makeRequest("POST", "/users", adminReq, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var adminUser dto.UserResponse
	_ = json.Unmarshal(w.Body.Bytes(), &adminUser)
	assert.Equal(t, "admin", adminUser.Role)

	// 測試重複 email
	w = makeRequest("POST", "/users", req, adminToken)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 測試無效 email
	invalidReq := map[string]string{
		"email":    "invalid-email",
		"password": "password123",
	}
	w = makeRequest("POST", "/users", invalidReq, adminToken)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 測試密碼太短
	shortPasswordReq := map[string]string{
		"email":    "shortpass@example.com",
		"password": "12345",
	}
	w = makeRequest("POST", "/users", shortPasswordReq, adminToken)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 測試無 token
	w = makeRequest("POST", "/users", req, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 測試非管理員用戶
	userToken := registerTestUser(t, "regularuser@example.com", "password123")
	w = makeRequest("POST", "/users", req, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestGetUsers 測試取得用戶列表
func TestGetUsers(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建幾個用戶
	for i := 1; i <= 3; i++ {
		req := map[string]string{
			"email":    fmt.Sprintf("user%d@example.com", i),
			"password": "password123",
		}
		makeRequest("POST", "/users", req, adminToken)
	}

	// 測試取得列表
	w := makeRequest("GET", "/users?page=1&page_size=10", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["data"])

	// 測試分頁
	w = makeRequest("GET", "/users?page=1&page_size=2", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 測試搜尋
	w = makeRequest("GET", "/users?search=user1", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 測試無 token
	w = makeRequest("GET", "/users", nil, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 測試非管理員用戶
	userToken := registerTestUser(t, "getusers@example.com", "password123")
	w = makeRequest("GET", "/users", nil, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestGetUser 測試取得單一用戶
func TestGetUser(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建一個用戶
	createReq := map[string]string{
		"email":    "getuser@example.com",
		"password": "password123",
	}
	w := makeRequest("POST", "/users", createReq, adminToken)
	var createdUser dto.UserResponse
	_ = json.Unmarshal(w.Body.Bytes(), &createdUser)
	userID := createdUser.ID

	// 測試成功取得
	w = makeRequest("GET", "/users/"+fmt.Sprintf("%d", userID), nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var user dto.UserResponse
	_ = json.Unmarshal(w.Body.Bytes(), &user)
	assert.Equal(t, "getuser@example.com", user.Email)

	// 測試不存在的 ID
	w = makeRequest("GET", "/users/99999", nil, adminToken)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 測試非管理員用戶
	userToken := registerTestUser(t, "getuser2@example.com", "password123")
	w = makeRequest("GET", "/users/"+fmt.Sprintf("%d", userID), nil, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestUpdateUser 測試更新用戶
func TestUpdateUser(t *testing.T) {
	setupTestDBForTest()
	setupTestRouter()
	adminToken := registerTestAdmin(t)

	// 先創建一個用戶
	createReq := map[string]string{
		"email":    "updateuser@example.com",
		"password": "password123",
	}
	w := makeRequest("POST", "/users", createReq, adminToken)
	var createdUser dto.UserResponse
	_ = json.Unmarshal(w.Body.Bytes(), &createdUser)
	userID := createdUser.ID

	// 測試成功更新 email
	updateReq := map[string]string{
		"email": "updated@example.com",
	}
	w = makeRequest("PUT", "/users/"+fmt.Sprintf("%d", userID), updateReq, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var user dto.UserResponse
	_ = json.Unmarshal(w.Body.Bytes(), &user)
	assert.Equal(t, "updated@example.com", user.Email)

	// 測試更新密碼
	updatePasswordReq := map[string]string{
		"password": "newpassword123",
	}
	w = makeRequest("PUT", "/users/"+fmt.Sprintf("%d", userID), updatePasswordReq, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 測試重複 email（使用另一個已存在的 email）
	duplicateReq := map[string]string{
		"email": "admin@test.com", // 使用已存在的 email
	}
	w = makeRequest("PUT", "/users/"+fmt.Sprintf("%d", userID), duplicateReq, adminToken)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 測試非管理員用戶
	userToken := registerTestUser(t, "updateuser2@example.com", "password123")
	w = makeRequest("PUT", "/users/"+fmt.Sprintf("%d", userID), updateReq, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestDeleteUser 測試刪除用戶
func TestDeleteUser(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建一個用戶
	createReq := map[string]string{
		"email":    "deleteuser@example.com",
		"password": "password123",
	}
	w := makeRequest("POST", "/users", createReq, adminToken)
	var createdUser dto.UserResponse
	_ = json.Unmarshal(w.Body.Bytes(), &createdUser)
	userID := createdUser.ID

	// 測試成功刪除
	w = makeRequest("DELETE", "/users/"+fmt.Sprintf("%d", userID), nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 驗證已刪除
	w = makeRequest("GET", "/users/"+fmt.Sprintf("%d", userID), nil, adminToken)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 測試非管理員用戶
	userToken := registerTestUser(t, "deleteuser2@example.com", "password123")
	createReq2 := map[string]string{
		"email":    "deleteuser3@example.com",
		"password": "password123",
	}
	w = makeRequest("POST", "/users", createReq2, adminToken)
	var createdUser2 dto.UserResponse
	_ = json.Unmarshal(w.Body.Bytes(), &createdUser2)
	userID2 := createdUser2.ID

	w = makeRequest("DELETE", "/users/"+fmt.Sprintf("%d", userID2), nil, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

