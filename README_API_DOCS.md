# API 文檔自動生成指南

## 快速開始

### 1. 安裝依賴

```bash
# 安裝 swag CLI 工具
go install github.com/swaggo/swag/cmd/swag@latest

# 安裝 Gin Swagger 套件
go get -u github.com/swaggo/files
go get -u github.com/swaggo/gin-swagger
```

### 2. 生成文檔

```bash
# 使用腳本（推薦）
./scripts/generate-docs.sh

# 或手動執行
swag init -g cmd/main.go -o docs
```

### 3. 啟動服務並訪問

```bash
# 啟動服務
go run cmd/main.go

# 訪問 Swagger UI
# http://localhost:8080/swagger/index.html
```

## 主要工具

### swaggo/swag

**最推薦的工具**，專為 Gin 框架設計：

- ✅ 從代碼註釋自動生成 Swagger 文檔
- ✅ 支援 OpenAPI 2.0 和 3.0
- ✅ 簡單易用，註釋格式清晰
- ✅ 與 Gin 完美整合

**GitHub**: https://github.com/swaggo/swag

### 其他選擇

1. **go-swagger** - 功能更強大，但配置複雜
2. **oapi-codegen** - 從 OpenAPI 規範生成代碼（反向）
3. **godoc** - Go 官方工具（主要用於代碼文檔）

## 註釋格式範例

### 基本端點

```go
// GetUser 獲取用戶
// @Summary 獲取用戶資訊
// @Description 根據 ID 獲取用戶詳細資訊
// @Tags user
// @Param id path int true "用戶 ID"
// @Success 200 {object} dto.UserResponse
// @Router /api/users/{id} [get]
func (ctrl *UserController) GetUser(c *gin.Context) {
    // ...
}
```

### 帶認證的端點

```go
// CreateBook 創建書籍
// @Summary 創建書籍
// @Description 創建一本新書籍（需要管理員權限）
// @Tags book
// @Security BearerAuth
// @Param request body dto.CreateBookRequest true "書籍資訊"
// @Success 201 {object} dto.BookResponse
// @Failure 401 {object} gin.H
// @Router /api/books [post]
func (ctrl *BookController) CreateBook(c *gin.Context) {
    // ...
}
```

### 查詢參數

```go
// GetProducts 獲取產品列表
// @Summary 獲取產品列表
// @Description 獲取產品列表，支援分頁和搜尋
// @Tags product
// @Param page query int false "頁碼" default(1)
// @Param page_size query int false "每頁數量" default(10)
// @Param search query string false "搜尋關鍵字"
// @Success 200 {object} dto.ProductListResponse
// @Router /api/products [get]
func (ctrl *ProductController) GetProducts(c *gin.Context) {
    // ...
}
```

## 已配置的功能

✅ Swagger UI 路由已配置在 `/swagger/index.html`  
✅ 基本 Swagger 註釋已添加到 `cmd/main.go`  
✅ 所有 Controller 都有 Swagger 註釋範例  
✅ 自動生成腳本已創建

## 下一步

1. 執行 `./scripts/generate-docs.sh` 生成文檔
2. 啟動服務：`go run cmd/main.go`
3. 訪問 http://localhost:8080/swagger/index.html
4. 在 Swagger UI 中測試 API

## 詳細文檔

更多詳細資訊請參考：`docs/SWAGGER_SETUP.md`

