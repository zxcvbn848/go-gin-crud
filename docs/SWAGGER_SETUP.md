# Swagger API 文件自動生成設置

## 介紹

本專案使用 [swaggo/swag](https://github.com/swaggo/swag) 來自動生成 API 文件。這個工具可以從 Go 代碼註釋中自動生成 Swagger/OpenAPI 規範文檔。

## 安裝

### 1. 安裝 swag CLI 工具

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. 安裝 Gin Swagger 套件

```bash
go get -u github.com/swaggo/files
go get -u github.com/swaggo/gin-swagger
```

## 使用方式

### 1. 在代碼中添加 Swagger 註釋

在 `cmd/main.go` 中已經添加了基本的 Swagger 註釋：

```go
// @title           Go Gin CRUD API
// @version         1.0
// @description     這是一個使用 Gin 框架構建的 CRUD API 服務
// @host      localhost:8080
// @BasePath  /
```

### 2. 在 Controller 中添加 API 註釋

每個 API 端點都需要添加註釋，例如：

```go
// GetStatus 獲取限流器狀態
// @Summary 獲取限流器狀態
// @Description 獲取指定 key 的限流器狀態
// @Tags rate-limiter
// @Param key query string true "限流鍵（如：IP、用戶ID等）"
// @Success 200 {object} dto.RateLimiterStatus
// @Router /api/rate-limiter/status [get]
func (ctrl *RateLimiterController) GetStatus(c *gin.Context) {
    // ...
}
```

### 3. 生成 Swagger 文檔

在專案根目錄執行：

```bash
swag init -g cmd/main.go -o docs
```

這會生成 `docs/swagger.json` 和 `docs/swagger.yaml` 文件。

### 4. 在路由中集成 Swagger UI

在 `internal/routes/health.go` 或創建新的路由文件中添加：

```go
import (
    "github.com/swaggo/files"
    "github.com/swaggo/gin-swagger"
)

// 在 RegisterHealthRoutes 或單獨的函數中添加
r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

### 5. 訪問 Swagger UI

啟動服務後，訪問：
- Swagger UI: http://localhost:8080/swagger/index.html
- Swagger JSON: http://localhost:8080/swagger/doc.json

## 常用註釋標籤

### 基本資訊
- `@title` - API 標題
- `@version` - API 版本
- `@description` - API 描述
- `@host` - 主機地址
- `@BasePath` - 基礎路徑

### 端點註釋
- `@Summary` - 端點摘要
- `@Description` - 詳細描述
- `@Tags` - 標籤（用於分組）
- `@Param` - 參數定義
- `@Success` - 成功響應
- `@Failure` - 失敗響應
- `@Router` - 路由定義

### 認證
- `@securityDefinitions.apikey` - API Key 認證定義
- `@Security` - 端點安全要求

## 範例

### 帶認證的端點

```go
// GetUser 獲取用戶資訊
// @Summary 獲取用戶資訊
// @Description 獲取指定 ID 的用戶資訊（需要認證）
// @Tags user
// @Security BearerAuth
// @Param id path int true "用戶 ID"
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} gin.H
// @Router /api/users/{id} [get]
func (ctrl *UserController) GetUser(c *gin.Context) {
    // ...
}
```

### 帶請求體的端點

```go
// CreateUser 創建用戶
// @Summary 創建用戶
// @Description 創建一個新用戶
// @Tags user
// @Param request body dto.CreateUserRequest true "用戶資訊"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} gin.H
// @Router /api/users [post]
func (ctrl *UserController) CreateUser(c *gin.Context) {
    // ...
}
```

## 自動生成腳本

可以創建一個腳本來自動生成文檔：

```bash
#!/bin/bash
# generate-docs.sh

echo "生成 Swagger 文檔..."
swag init -g cmd/main.go -o docs

if [ $? -eq 0 ]; then
    echo "✅ Swagger 文檔生成成功！"
    echo "📄 文檔位置: docs/swagger.json"
    echo "🌐 訪問地址: http://localhost:8080/swagger/index.html"
else
    echo "❌ Swagger 文檔生成失敗！"
    exit 1
fi
```

## 注意事項

1. **註釋格式**：Swagger 註釋必須緊貼在函數上方，不能有空行
2. **類型引用**：使用 `{object} dto.XXX` 時，確保 DTO 結構體有正確的 JSON 標籤
3. **路徑參數**：使用 `{id}` 格式定義路徑參數
4. **查詢參數**：使用 `query` 類型
5. **請求體**：使用 `body` 類型，並指定 DTO 結構

## 其他工具

除了 swaggo/swag，還有其他選擇：

1. **go-swagger** - 功能更強大，但配置更複雜
2. **oapi-codegen** - 從 OpenAPI 規範生成代碼
3. **godoc** - Go 官方文檔工具（主要用於代碼文檔，不是 API 文檔）

## 參考資源

- [swaggo/swag GitHub](https://github.com/swaggo/swag)
- [Swagger 註釋規範](https://github.com/swaggo/swag#declarative-comments-format)
- [Gin Swagger 範例](https://github.com/swaggo/gin-swagger)


