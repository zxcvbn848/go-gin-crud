# go-gin-crud

## Docker 部署

### 啟動所有服務（MySQL + Go 應用）

在專案根目錄可使用下列指令啟動所有服務：

```bash
docker compose up -d
```

這會啟動：
- MySQL 資料庫（端口 3307）
- Go 應用程式（端口 8080）

### 停止服務

```bash
docker compose down
```

### 查看日誌

```bash
# 查看所有服務日誌
docker compose logs -f

# 查看 Go 應用日誌
docker compose logs -f app

# 查看 MySQL 日誌
docker compose logs -f mysql
```

### 重新構建 Go 應用

**重要：** 如果修改了程式碼，必須重新構建映像才能生效。`docker compose restart` 只會重啟容器，不會載入新的程式碼。

```bash
# 重新構建並啟動
docker compose build app
docker compose up -d app

# 或使用一行指令
docker compose up -d --build app
```

## 資料庫資訊

預設使用的資料庫資訊：

- host: `127.0.0.1` (本地開發) / `mysql` (Docker 環境)
- port: `3307` (本地) / `3306` (Docker 內部)
- database: `goGinCRUD`
- user: `gogin`
- password: `a3935522`

## 環境變數設定

### 使用 .env 檔案

專案支援使用 `.env` 檔案來設定環境變數。首先複製範例檔案：

```bash
cp .env.example .env
```

然後編輯 `.env` 檔案來設定你的資料庫連線資訊：

```env
# 資料庫連線設定（Go 應用使用）
DB_HOST=127.0.0.1
DB_PORT=3307
DB_USER=gogin
DB_PASSWORD=a3935522
DB_NAME=goGinCRUD

# MySQL Docker 容器設定
MYSQL_ROOT_PASSWORD=a3935522
MYSQL_DATABASE=goGinCRUD
MYSQL_USER=gogin
MYSQL_PASSWORD=a3935522

# 時區設定
TZ=Asia/Taipei
```

### 支援的環境變數

**Go 應用使用的環境變數：**

- `DB_HOST`: 資料庫主機（預設: 127.0.0.1）
- `DB_PORT`: 資料庫端口（預設: 3307）
- `DB_USER`: 資料庫用戶（預設: gogin）
- `DB_PASSWORD`: 資料庫密碼（預設: a3935522）
- `DB_NAME`: 資料庫名稱（預設: goGinCRUD）
- `ACCESS_SECRET`: JWT Access Token 密鑰（預設: ACCESS_SECRET，**生產環境請務必修改**）
- `REFRESH_SECRET`: JWT Refresh Token 密鑰（預設: REFRESH_SECRET，**生產環境請務必修改**）

**MySQL Docker 容器使用的環境變數：**

- `MYSQL_ROOT_PASSWORD`: MySQL root 密碼（預設: a3935522）
- `MYSQL_DATABASE`: 初始資料庫名稱（預設: goGinCRUD）
- `MYSQL_USER`: MySQL 用戶（預設: gogin）
- `MYSQL_PASSWORD`: MySQL 用戶密碼（預設: a3935522）

**共用環境變數：**

- `TZ`: 時區設定（預設: Asia/Taipei）

**注意：** `.env` 檔案不會被提交到 Git（已在 `.gitignore` 中），請使用 `.env.example` 作為範本。

## 資料庫遷移

### 自動遷移（推薦）

應用程式啟動時會自動執行資料庫遷移（`AutoMigrate()`），這會：
- 自動建立不存在的表
- 自動新增缺失的欄位（如 `created_at` 和 `updated_at`）

**重新啟動應用程式即可執行遷移：**

```bash
# 本地開發
go run cmd/main.go

# Docker 環境
docker compose restart app
```

### 手動執行 SQL（備選）

如果需要手動執行遷移，可以使用 SQL 腳本：

```bash
# 使用 Docker 執行 SQL
docker exec -i go_gin_crud_mysql mysql -ugogin -pa3935522 goGinCRUD < migrations/add_timestamps.sql
```

詳細遷移說明請參考 [docs/MIGRATION.md](./docs/MIGRATION.md)

### 為既有資料設定時間戳

遷移後，可以使用 Seeder 工具為既有記錄設定時間戳（只更新 NULL 的欄位，有值的不更新）：

```bash
go run cmd/seed/main.go
```

詳細 Seeder 說明請參考 [docs/SEEDER.md](./docs/SEEDER.md)

## API 端點

應用啟動後可訪問：

- API 端點: `http://localhost:8080`
- 健康檢查: `http://localhost:8080/health` (如果有的話)
