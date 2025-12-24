# 安裝 Swagger 文檔工具

## 步驟 1: 安裝 swag CLI

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## 步驟 2: 安裝 Gin Swagger 套件

```bash
go get -u github.com/swaggo/files
go get -u github.com/swaggo/gin-swagger
```

## 步驟 3: 啟用 Swagger 路由

編輯 `internal/routes/health.go`，取消註釋 Swagger 相關代碼：

```go
package routes

import (
	"go-gin-crud/internal/controller"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterHealthRoutes(r *gin.Engine) {
	healthController := controller.NewHealthController()
	r.GET("/health", healthController.GetHealth)

	// Swagger 文檔路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
```

## 步驟 4: 生成文檔

```bash
# 使用腳本
./scripts/generate-docs.sh

# 或手動執行
swag init -g cmd/main.go -o docs
```

## 步驟 5: 啟動服務並訪問

```bash
go run cmd/main.go
```

訪問：http://localhost:8080/swagger/index.html

## 驗證安裝

執行以下命令檢查是否安裝成功：

```bash
# 檢查 swag
swag version

# 檢查 Go 模組
go list -m github.com/swaggo/files
go list -m github.com/swaggo/gin-swagger
```

