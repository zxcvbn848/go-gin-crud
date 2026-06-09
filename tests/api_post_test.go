package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreatePost 測試創建文章
func TestCreatePost(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 測試成功創建
	req := map[string]string{
		"title":   "測試文章",
		"content": "這是測試文章的內容",
	}
	w := makeRequest("POST", "/posts", req, adminToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var post map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &post)
	assert.Equal(t, "測試文章", post["title"])
	assert.Equal(t, "這是測試文章的內容", post["content"])

	// 測試缺少必填欄位
	invalidReq := map[string]string{
		"title": "只有標題",
	}
	w = makeRequest("POST", "/posts", invalidReq, adminToken)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 測試無 token
	w = makeRequest("POST", "/posts", req, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 測試非管理員用戶
	userToken := registerTestUser(t, "postuser@example.com", "password123")
	w = makeRequest("POST", "/posts", req, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestGetPosts 測試取得文章列表
func TestGetPosts(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建幾篇文章
	for i := 1; i <= 3; i++ {
		req := map[string]string{
			"title":   fmt.Sprintf("文章 %d", i),
			"content": fmt.Sprintf("文章 %d 的內容", i),
		}
		makeRequest("POST", "/posts", req, adminToken)
	}

	// 測試取得列表
	w := makeRequest("GET", "/posts?page=1&page_size=10", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["data"])

	// 測試分頁
	w = makeRequest("GET", "/posts?page=1&page_size=2", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 測試搜尋
	w = makeRequest("GET", "/posts?search=文章", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 測試無 token
	w = makeRequest("GET", "/posts", nil, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 測試一般用戶可以取得列表
	userToken := registerTestUser(t, "getposts@example.com", "password123")
	w = makeRequest("GET", "/posts", nil, userToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetPost 測試取得單一文章
func TestGetPost(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建一篇文章
	createReq := map[string]string{
		"title":   "單一文章",
		"content": "單一文章的內容",
	}
	w := makeRequest("POST", "/posts", createReq, adminToken)
	var createdPost map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createdPost)
	postID := int(createdPost["id"].(float64))

	// 測試成功取得
	w = makeRequest("GET", "/posts/"+fmt.Sprintf("%d", postID), nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var post map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &post)
	assert.Equal(t, "單一文章", post["title"])

	// 測試不存在的 ID
	w = makeRequest("GET", "/posts/99999", nil, adminToken)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 測試一般用戶可以取得
	userToken := registerTestUser(t, "getpost@example.com", "password123")
	w = makeRequest("GET", "/posts/"+fmt.Sprintf("%d", postID), nil, userToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestUpdatePost 測試更新文章
func TestUpdatePost(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建一篇文章
	createReq := map[string]string{
		"title":   "原始文章",
		"content": "原始內容",
	}
	w := makeRequest("POST", "/posts", createReq, adminToken)
	var createdPost map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createdPost)
	postID := int(createdPost["id"].(float64))

	// 測試管理員可以更新
	updateReq := map[string]string{
		"title":   "更新文章",
		"content": "更新內容",
	}
	w = makeRequest("PUT", "/posts/"+fmt.Sprintf("%d", postID), updateReq, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var post map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &post)
	assert.Equal(t, "更新文章", post["title"])

	// 測試非管理員用戶無法更新別人的文章
	userToken := registerTestUser(t, "updatepost@example.com", "password123")
	w = makeRequest("PUT", "/posts/"+fmt.Sprintf("%d", postID), updateReq, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestDeletePost 測試刪除文章
func TestDeletePost(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建一篇文章
	createReq := map[string]string{
		"title":   "待刪除文章",
		"content": "待刪除內容",
	}
	w := makeRequest("POST", "/posts", createReq, adminToken)
	var createdPost map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createdPost)
	postID := int(createdPost["id"].(float64))

	// 測試管理員可以刪除
	w = makeRequest("DELETE", "/posts/"+fmt.Sprintf("%d", postID), nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 驗證已刪除
	w = makeRequest("GET", "/posts/"+fmt.Sprintf("%d", postID), nil, adminToken)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 測試非管理員用戶無法刪除別人的文章
	userToken := registerTestUser(t, "deletepost@example.com", "password123")
	createReq2 := map[string]string{
		"title":   "另一篇文章",
		"content": "另一篇文章的內容",
	}
	w = makeRequest("POST", "/posts", createReq2, adminToken)
	var createdPost2 map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createdPost2)
	postID2 := int(createdPost2["id"].(float64))

	w = makeRequest("DELETE", "/posts/"+fmt.Sprintf("%d", postID2), nil, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
