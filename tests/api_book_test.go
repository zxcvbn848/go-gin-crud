package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreateBook 測試創建書籍
func TestCreateBook(t *testing.T) {
	setupTestDBForTest()
	setupTestRouter()
	adminToken := registerTestAdmin(t)

	// 測試成功創建
	req := map[string]string{
		"title":  "測試書籍",
		"author": "測試作者",
	}
	w := makeRequest("POST", "/books", req, adminToken)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var book map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &book)
	assert.Equal(t, "測試書籍", book["title"])
	assert.Equal(t, "測試作者", book["author"])

	// 測試缺少必填欄位
	invalidReq := map[string]string{
		"title": "只有標題",
	}
	w = makeRequest("POST", "/books", invalidReq, adminToken)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 測試無 token
	w = makeRequest("POST", "/books", req, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 測試非管理員用戶
	userToken := registerTestUser(t, "user@example.com", "password123")
	w = makeRequest("POST", "/books", req, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestGetBooks 測試取得書籍列表
func TestGetBooks(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建幾本書
	for i := 1; i <= 3; i++ {
		req := map[string]string{
			"title":  "書籍 " + string(rune(i+48)),
			"author": "作者 " + string(rune(i+48)),
		}
		w := makeRequest("POST", "/books", req, adminToken)
		if w.Code != http.StatusOK {
			t.Fatalf("創建書籍失敗: %d", w.Code)
		}
	}

	// 測試取得列表
	w := makeRequest("GET", "/books?page=1&page_size=10", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["data"])

	// 測試分頁
	w = makeRequest("GET", "/books?page=1&page_size=2", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 測試搜尋
	w = makeRequest("GET", "/books?search=書籍", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 測試無 token
	w = makeRequest("GET", "/books", nil, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetBook 測試取得單一書籍
func TestGetBook(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建一本書
	createReq := map[string]string{
		"title":  "單一書籍",
		"author": "單一作者",
	}
	w := makeRequest("POST", "/books", createReq, adminToken)
	assert.Equal(t, http.StatusOK, w.Code, "創建書籍應該成功")
	
	var createdBook map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createdBook)
	assert.NoError(t, err, "應該能解析創建的書籍")
	assert.NotNil(t, createdBook["id"], "書籍應該有 ID")
	
	bookID := int(createdBook["id"].(float64))

	// 測試成功取得
	w = makeRequest("GET", "/books/"+fmt.Sprintf("%d", bookID), nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var book map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &book)
	assert.Equal(t, "單一書籍", book["title"])

	// 測試不存在的 ID
	w = makeRequest("GET", "/books/99999", nil, adminToken)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestUpdateBook 測試更新書籍
func TestUpdateBook(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建一本書
	createReq := map[string]string{
		"title":  "原始標題",
		"author": "原始作者",
	}
	w := makeRequest("POST", "/books", createReq, adminToken)
	assert.Equal(t, http.StatusOK, w.Code, "創建書籍應該成功")
	
	var createdBook map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createdBook)
	assert.NoError(t, err, "應該能解析創建的書籍")
	assert.NotNil(t, createdBook["id"], "書籍應該有 ID")
	
	bookID := int(createdBook["id"].(float64))

	// 測試成功更新
	updateReq := map[string]string{
		"title":  "更新標題",
		"author": "更新作者",
	}
	w = makeRequest("PUT", "/books/"+fmt.Sprintf("%d", bookID), updateReq, adminToken)
	assert.Equal(t, http.StatusOK, w.Code, "更新書籍應該成功")

	var book map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &book)
	assert.NoError(t, err, "應該能解析更新的書籍")
	assert.Equal(t, "更新標題", book["title"])

	// 測試非管理員用戶
	userToken := registerTestUser(t, "updateuser@example.com", "password123")
	w = makeRequest("PUT", "/books/"+fmt.Sprintf("%d", bookID), updateReq, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code, "非管理員用戶不應該能更新書籍")
}

// TestDeleteBook 測試刪除書籍
func TestDeleteBook(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建一本書
	createReq := map[string]string{
		"title":  "待刪除書籍",
		"author": "待刪除作者",
	}
	w := makeRequest("POST", "/books", createReq, adminToken)
	assert.Equal(t, http.StatusOK, w.Code, "創建書籍應該成功")
	
	var createdBook map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createdBook)
	assert.NoError(t, err, "應該能解析創建的書籍")
	assert.NotNil(t, createdBook["id"], "書籍應該有 ID")
	
	bookID := int(createdBook["id"].(float64))

	// 測試成功刪除
	w = makeRequest("DELETE", "/books/"+fmt.Sprintf("%d", bookID), nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 驗證已刪除
	w = makeRequest("GET", "/books/"+fmt.Sprintf("%d", bookID), nil, adminToken)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 測試非管理員用戶
	userToken := registerTestUser(t, "deleteuser@example.com", "password123")
	createReq2 := map[string]string{
		"title":  "另一本書",
		"author": "另一個作者",
	}
	w = makeRequest("POST", "/books", createReq2, adminToken)
	assert.Equal(t, http.StatusOK, w.Code, "創建書籍應該成功")
	
	var createdBook2 map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createdBook2)
	assert.NoError(t, err, "應該能解析創建的書籍")
	assert.NotNil(t, createdBook2["id"], "書籍應該有 ID")
	
	bookID2 := int(createdBook2["id"].(float64))

	w = makeRequest("DELETE", "/books/"+fmt.Sprintf("%d", bookID2), nil, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

