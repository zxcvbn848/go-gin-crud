# GitHub Actions 工作流說明

本目錄包含用於 CI/CD 的 GitHub Actions 工作流配置。

## 工作流文件

### 1. `ci.yml` - 持續集成
**觸發條件：**
- Push 到 `main` 或 `dev` 分支
- 創建 Pull Request

**執行任務：**
- ✅ 運行測試（包含 MySQL 服務）
- ✅ 構建應用
- ✅ 代碼檢查（golangci-lint）
- ✅ 上傳測試覆蓋率

### 2. `deploy.yml` - 生產環境部署
**觸發條件：**
- Push 到 `main` 分支
- 創建版本標籤（`v*`）
- 手動觸發（workflow_dispatch）

**執行任務：**
- 🐳 構建並推送 Docker 映像到 GitHub Container Registry
- 🚀 部署到服務器（SSH）
- ✅ 健康檢查
- 📢 發送通知（Slack，可選）

### 3. `docker-compose-deploy.yml` - Docker Compose 部署
**觸發條件：**
- Push 到 `main` 分支
- 手動觸發

**執行任務：**
- 🐳 構建 Docker 映像
- 📦 上傳配置文件到服務器
- 🚀 使用 Docker Compose 部署
- ✅ 健康檢查

## 設置 GitHub Secrets

在 GitHub 倉庫設置中添加以下 Secrets：

### 必需
- `DEPLOY_HOST` - 部署服務器地址
- `DEPLOY_USER` - SSH 用戶名
- `DEPLOY_SSH_KEY` - SSH 私鑰
- `DEPLOY_URL` - 應用訪問 URL（用於健康檢查）

### 可選
- `DEPLOY_PORT` - SSH 端口（默認 22）
- `DEPLOY_PASSWORD` - Docker Registry 密碼（如果需要）
- `SLACK_WEBHOOK_URL` - Slack Webhook URL（用於通知）

## 設置步驟

### 1. 生成 SSH 密鑰對
```bash
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/deploy_key
```

### 2. 將公鑰添加到服務器
```bash
cat ~/.ssh/deploy_key.pub >> ~/.ssh/authorized_keys
```

### 3. 在 GitHub 設置 Secrets
1. 進入倉庫 Settings → Secrets and variables → Actions
2. 添加以下 Secrets：
   - `DEPLOY_HOST`: 你的服務器 IP 或域名
   - `DEPLOY_USER`: SSH 用戶名（如 `root` 或 `deploy`）
   - `DEPLOY_SSH_KEY`: 私鑰內容（`~/.ssh/deploy_key`）
   - `DEPLOY_URL`: 應用 URL（如 `https://api.example.com`）

### 4. 在服務器上準備環境
```bash
# 安裝 Docker 和 Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# 創建部署目錄
mkdir -p /opt/go-gin-crud
cd /opt/go-gin-crud

# 創建 .env 文件
cp .env.example .env
# 編輯 .env 文件，填入正確的配置
```

## 使用方式

### 自動部署
推送到 `main` 分支會自動觸發部署：
```bash
git push origin main
```

### 手動部署
1. 進入 GitHub Actions 頁面
2. 選擇 "Deploy to Production" workflow
3. 點擊 "Run workflow"
4. 選擇環境（production 或 staging）
5. 點擊 "Run workflow"

### 版本標籤部署
創建版本標籤會觸發部署：
```bash
git tag v1.0.0
git push origin v1.0.0
```

## 部署流程

1. **構建階段**
   - 構建 Docker 映像
   - 推送到 GitHub Container Registry

2. **部署階段**
   - SSH 連接到服務器
   - 拉取最新映像
   - 停止舊容器
   - 啟動新容器
   - 執行健康檢查

3. **驗證階段**
   - 檢查 `/health` 端點
   - 發送部署通知

## 故障排除

### 部署失敗
1. 檢查 Secrets 是否正確設置
2. 確認服務器 SSH 連接正常
3. 檢查服務器上的 Docker 是否運行
4. 查看 GitHub Actions 日誌

### 健康檢查失敗
1. 確認應用已正確啟動
2. 檢查端口是否正確暴露
3. 確認環境變數已正確設置

## 安全建議

1. **使用專用部署用戶**
   - 不要使用 root 用戶
   - 創建專用的部署用戶並限制權限

2. **保護 SSH 密鑰**
   - 不要將私鑰提交到倉庫
   - 定期輪換 SSH 密鑰

3. **環境變數管理**
   - 使用 `.env` 文件管理敏感信息
   - 不要將 `.env` 文件提交到倉庫

4. **網絡安全**
   - 使用防火牆限制訪問
   - 考慮使用 VPN 或私有網絡



