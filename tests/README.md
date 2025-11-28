# API 測試說明

## 測試架構

本專案使用 Go 的標準測試框架 `testing` 和 `testify/assert` 來進行 API 測試。

## 測試文件結構

- `api_test.go` - 測試輔助函數和設置
- `api_auth_test.go` - 認證相關 API 測試
- `api_book_test.go` - 書籍相關 API 測試
- `api_user_test.go` - 用戶相關 API 測試
- `api_product_test.go` - 產品相關 API 測試
- `api_post_test.go` - 文章相關 API 測試

## 運行測試

### 運行所有測試

```bash
go test ./tests/... -v
```

或從專案根目錄：

```bash
go test -v ./tests
```

### 運行特定測試文件

```bash
go test ./tests -v -run TestRegister
go test ./tests -v -run TestLogin
go test ./tests -v -run TestCreateBook
```

### 運行特定測試函數

```bash
go test ./tests -v -run TestRegister
go test ./tests -v -run TestLogin
go test ./tests -v -run TestCreateBook
```

## 測試覆蓋範圍

### 認證 API (`api_auth_test.go`) - 4 個測試
- ✅ 用戶註冊 (`TestRegister`) - 測試成功註冊、重複註冊、無效 email、密碼太短
- ✅ 用戶登入 (`TestLogin`) - 測試成功登入、錯誤密碼、不存在的用戶
- ✅ 取得用戶資料 (`TestProfile`) - 測試成功取得、無 token、無效 token
- ✅ 用戶登出 (`TestLogout`) - 測試成功登出、登出後無法使用 token

### 書籍 API (`api_book_test.go`) - 5 個測試
- ✅ 創建書籍 (`TestCreateBook`) - 測試成功創建、缺少欄位、無 token、非管理員
- ✅ 取得書籍列表 (`TestGetBooks`) - 測試成功取得、分頁、搜尋、無 token
- ✅ 取得單一書籍 (`TestGetBook`) - 測試成功取得、不存在 ID
- ✅ 更新書籍 (`TestUpdateBook`) - 測試成功更新、非管理員
- ✅ 刪除書籍 (`TestDeleteBook`) - 測試成功刪除、非管理員

### 用戶 API (`api_user_test.go`) - 5 個測試
- ✅ 創建用戶 (`TestCreateUser`) - 測試成功創建、創建管理員、重複 email、無效 email、密碼太短、無 token、非管理員
- ✅ 取得用戶列表 (`TestGetUsers`) - 測試成功取得、分頁、搜尋、無 token、非管理員
- ✅ 取得單一用戶 (`TestGetUser`) - 測試成功取得、不存在 ID、非管理員
- ✅ 更新用戶 (`TestUpdateUser`) - 測試成功更新 email、更新密碼、重複 email、非管理員
- ✅ 刪除用戶 (`TestDeleteUser`) - 測試成功刪除、非管理員

### 產品 API (`api_product_test.go`) - 5 個測試
- ✅ 創建產品 (`TestCreateProduct`) - 測試成功創建、缺少欄位、負數價格、無 token、非管理員
- ✅ 取得產品列表 (`TestGetProducts`) - 測試成功取得、分頁、搜尋、無 token、一般用戶可取得
- ✅ 取得單一產品 (`TestGetProduct`) - 測試成功取得、不存在 ID、一般用戶可取得
- ✅ 更新產品 (`TestUpdateProduct`) - 測試成功更新、部分更新、非管理員
- ✅ 刪除產品 (`TestDeleteProduct`) - 測試成功刪除、非管理員

### 文章 API (`api_post_test.go`) - 5 個測試
- ✅ 創建文章 (`TestCreatePost`) - 測試成功創建、缺少欄位、無 token、非管理員
- ✅ 取得文章列表 (`TestGetPosts`) - 測試成功取得、分頁、搜尋、無 token、一般用戶可取得
- ✅ 取得單一文章 (`TestGetPost`) - 測試成功取得、不存在 ID、一般用戶可取得
- ✅ 更新文章 (`TestUpdatePost`) - 測試管理員可更新、非管理員無法更新別人的文章
- ✅ 刪除文章 (`TestDeletePost`) - 測試管理員可刪除、非管理員無法刪除別人的文章

## 測試統計

- **總測試數量**: 24 個測試函數
- **API 端點覆蓋**: 所有主要 CRUD 操作
- **權限測試**: 包含認證和授權測試
- **驗證測試**: 包含輸入驗證測試

## 測試環境

測試使用 SQLite 記憶體資料庫（`:memory:`），不會影響實際資料庫。

測試環境變數：
- `ACCESS_SECRET`: `test_access_secret_key_for_testing_only`
- `REFRESH_SECRET`: `test_refresh_secret_key_for_testing_only`

## 測試輔助函數

### `setupTestRouter()`
設置測試路由器，包含所有路由和中介層。

### `makeRequest(method, url, body, token)`
發送 HTTP 請求的輔助函數。

### `registerTestUser(t, email, password)`
註冊測試用戶並返回 access token。

### `registerTestAdmin(t)`
註冊測試管理員並返回 access token。

## 注意事項

1. 每個測試都是獨立的，使用記憶體資料庫
2. 測試會自動清理資料庫
3. 測試使用測試專用的 JWT Secret
4. 測試不會影響實際的資料庫或配置

## 擴展測試

要添加新的測試：

1. 在對應的測試文件中添加新的測試函數
2. 使用 `makeRequest` 發送請求
3. 使用 `assert` 驗證結果
4. 確保測試函數名稱以 `Test` 開頭

範例：

```go
func TestNewFeature(t *testing.T) {
    token := registerTestUser(t, "test@example.com", "password123")
    
    req := map[string]string{
        "field": "value",
    }
    w := makeRequest("POST", "/api/endpoint", req, token)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

