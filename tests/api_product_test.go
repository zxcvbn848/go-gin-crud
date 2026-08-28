package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreateProduct 測試創建產品
func TestCreateProduct(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 測試成功創建
	req := map[string]interface{}{
		"name":        "測試產品",
		"description": "這是一個測試產品",
		"price":       99.99,
		"stock":       100,
	}
	w := makeRequest("POST", "/products", req, adminToken)

	assert.Equal(t, http.StatusOK, w.Code)

	var product map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &product)
	assert.Equal(t, "測試產品", product["name"])
	assert.Equal(t, 99.99, product["price"])
	assert.Equal(t, float64(100), product["stock"])

	// 測試缺少必填欄位
	invalidReq := map[string]interface{}{
		"name": "只有名稱",
	}
	w = makeRequest("POST", "/products", invalidReq, adminToken)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 測試價格為負數
	negativePriceReq := map[string]interface{}{
		"name":  "負數價格產品",
		"price": -10.0,
	}
	w = makeRequest("POST", "/products", negativePriceReq, adminToken)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 測試無 token
	w = makeRequest("POST", "/products", req, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 測試非管理員用戶
	userToken := registerTestUser(t, "productuser@example.com", "password123")
	w = makeRequest("POST", "/products", req, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestGetProducts 測試取得產品列表
func TestGetProducts(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建幾個產品
	for i := 1; i <= 3; i++ {
		req := map[string]interface{}{
			"name":        fmt.Sprintf("產品 %d", i),
			"description": fmt.Sprintf("產品 %d 的描述", i),
			"price":       float64(10.0 * float64(i)),
			"stock":       i * 10,
		}
		w := makeRequest("POST", "/products", req, adminToken)
		if w.Code != http.StatusOK {
			t.Fatalf("創建產品失敗: %d", w.Code)
		}
	}

	// 測試取得列表
	w := makeRequest("GET", "/products?page=1&page_size=10", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["data"])

	// 測試分頁
	w = makeRequest("GET", "/products?page=1&page_size=2", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 測試搜尋
	w = makeRequest("GET", "/products?search=產品", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 測試無 token
	w = makeRequest("GET", "/products", nil, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 測試一般用戶可以取得列表
	userToken := registerTestUser(t, "getproducts@example.com", "password123")
	w = makeRequest("GET", "/products", nil, userToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetProduct 測試取得單一產品
func TestGetProduct(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建一個產品
	createReq := map[string]interface{}{
		"name":        "單一產品",
		"description": "單一產品的描述",
		"price":       50.0,
		"stock":       25,
	}
	w := makeRequest("POST", "/products", createReq, adminToken)
	var createdProduct map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createdProduct)
	productID := int(createdProduct["id"].(float64))

	// 測試成功取得
	w = makeRequest("GET", "/products/"+fmt.Sprintf("%d", productID), nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var product map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &product)
	assert.Equal(t, "單一產品", product["name"])

	// 測試不存在的 ID
	w = makeRequest("GET", "/products/99999", nil, adminToken)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 測試一般用戶可以取得
	userToken := registerTestUser(t, "getproduct@example.com", "password123")
	w = makeRequest("GET", "/products/"+fmt.Sprintf("%d", productID), nil, userToken)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestUpdateProduct 測試更新產品
func TestUpdateProduct(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建一個產品
	createReq := map[string]interface{}{
		"name":        "原始產品",
		"description": "原始描述",
		"price":       100.0,
		"stock":       50,
	}
	w := makeRequest("POST", "/products", createReq, adminToken)
	var createdProduct map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createdProduct)
	productID := int(createdProduct["id"].(float64))

	// 測試成功更新
	updateReq := map[string]interface{}{
		"name":        "更新產品",
		"description": "更新描述",
		"price":       150.0,
		"stock":       75,
	}
	w = makeRequest("PUT", "/products/"+fmt.Sprintf("%d", productID), updateReq, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var product map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &product)
	assert.Equal(t, "更新產品", product["name"])
	assert.Equal(t, 150.0, product["price"])

	// 測試部分更新
	partialUpdateReq := map[string]interface{}{
		"price": 200.0,
	}
	w = makeRequest("PUT", "/products/"+fmt.Sprintf("%d", productID), partialUpdateReq, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 測試非管理員用戶
	userToken := registerTestUser(t, "updateproduct@example.com", "password123")
	w = makeRequest("PUT", "/products/"+fmt.Sprintf("%d", productID), updateReq, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestDeleteProduct 測試刪除產品
func TestDeleteProduct(t *testing.T) {
	adminToken := registerTestAdmin(t)

	// 先創建一個產品
	createReq := map[string]interface{}{
		"name":        "待刪除產品",
		"description": "待刪除描述",
		"price":       99.0,
		"stock":       10,
	}
	w := makeRequest("POST", "/products", createReq, adminToken)
	var createdProduct map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createdProduct)
	productID := int(createdProduct["id"].(float64))

	// 測試成功刪除
	w = makeRequest("DELETE", "/products/"+fmt.Sprintf("%d", productID), nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	// 驗證已刪除
	w = makeRequest("GET", "/products/"+fmt.Sprintf("%d", productID), nil, adminToken)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 測試非管理員用戶
	userToken := registerTestUser(t, "deleteproduct@example.com", "password123")
	createReq2 := map[string]interface{}{
		"name":  "另一產品",
		"price": 50.0,
		"stock": 20,
	}
	w = makeRequest("POST", "/products", createReq2, adminToken)
	var createdProduct2 map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &createdProduct2)
	productID2 := int(createdProduct2["id"].(float64))

	w = makeRequest("DELETE", "/products/"+fmt.Sprintf("%d", productID2), nil, userToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
