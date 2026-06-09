# 部署指南

本指南說明如何使用 GitHub Actions 將應用部署到正式環境。

## 快速開始

### 1. 準備服務器

確保服務器已安裝：
- Docker (20.10+)
- Docker Compose (2.0+)

```bash
# 檢查 Docker 版本
docker --version
docker-compose --version
```

### 2. 設置 GitHub Secrets

在 GitHub 倉庫中設置以下 Secrets：

1. 進入倉庫：`Settings` → `Secrets and variables` → `Actions`
2. 點擊 `New repository secret`
3. 添加以下 Secrets：

| Secret 名稱 | 說明 | 範例 |
|------------|------|------|
| `DEPLOY_HOST` | 服務器 IP 或域名 | `192.168.1.100` 或 `api.example.com` |
| `DEPLOY_USER` | SSH 用戶名 | `deploy` 或 `root` |
| `DEPLOY_SSH_KEY` | SSH 私鑰內容 | 見下方說明 |
| `DEPLOY_URL` | 應用訪問 URL | `https://api.example.com` |
| `DEPLOY_PORT` | SSH 端口（可選） | `22` |

### 3. 生成 SSH 密鑰對

```bash
# 生成 SSH 密鑰對
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/deploy_key

# 查看公鑰（需要添加到服務器）
cat ~/.ssh/deploy_key.pub

# 查看私鑰（需要添加到 GitHub Secrets）
cat ~/.ssh/deploy_key
```

### 4. 配置服務器

#### 4.1 添加 SSH 公鑰到服務器

```bash
# 在服務器上執行
mkdir -p ~/.ssh
chmod 700 ~/.ssh
echo "你的公鑰內容" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

#### 4.2 創建部署目錄

```bash
# 在服務器上執行
mkdir -p /opt/go-gin-crud
cd /opt/go-gin-crud
```

#### 4.3 創建環境變數文件

```bash
# 在服務器上創建 .env 文件
cat > /opt/go-gin-crud/.env << EOF
# 數據庫配置
DB_HOST=mysql
DB_PORT=3306
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name

# JWT 配置
ACCESS_SECRET=your_access_secret_key
REFRESH_SECRET=your_refresh_secret_key

# 應用配置
PORT=8080
TZ=Asia/Taipei
EOF

chmod 600 /opt/go-gin-crud/.env
```

#### 4.4 準備 Docker Compose 文件（可選）

如果使用 `docker-compose-deploy.yml`，需要將 `docker-compose.yml` 上傳到服務器：

```bash
# 從本地複製到服務器
scp docker-compose.yml deploy@your-server:/opt/go-gin-crud/
```

## 部署方式

### 方式一：自動部署（推薦）

當代碼推送到 `main` 分支時，會自動觸發部署：

```bash
git add .
git commit -m "準備部署"
git push origin main
```

### 方式二：手動觸發

1. 進入 GitHub Actions 頁面
2. 選擇 "Deploy to Production" workflow
3. 點擊 "Run workflow"
4. 選擇環境（production 或 staging）
5. 點擊 "Run workflow"

### 方式三：版本標籤部署

創建版本標籤會觸發部署並創建 Release：

```bash
git tag v1.0.0
git push origin v1.0.0
```

## 部署流程

### 1. CI 階段（自動執行）

- ✅ 運行測試
- ✅ 代碼檢查
- ✅ 構建應用

### 2. 構建 Docker 映像

- 🐳 構建 Docker 映像
- 📦 推送到 GitHub Container Registry

### 3. 部署到服務器

- 🔐 SSH 連接到服務器
- 📥 拉取最新 Docker 映像
- 🛑 停止舊容器
- 🚀 啟動新容器
- ✅ 執行健康檢查

## 驗證部署

### 檢查容器狀態

```bash
# SSH 到服務器
ssh deploy@your-server

# 檢查容器
docker ps | grep go-gin-crud

# 查看日誌
docker logs go-gin-crud-app
```

### 測試 API

```bash
# 健康檢查
curl http://your-server:8080/health

# 應該返回
# {"status":"ok"}
```

## 故障排除

### 問題 1: SSH 連接失敗

**症狀：** GitHub Actions 無法連接到服務器

**解決方案：**
1. 檢查 `DEPLOY_HOST` 和 `DEPLOY_PORT` 是否正確
2. 確認服務器防火牆允許 SSH 連接
3. 驗證 SSH 密鑰是否正確添加到 GitHub Secrets
4. 測試手動 SSH 連接：`ssh -i ~/.ssh/deploy_key deploy@your-server`

### 問題 2: Docker 映像拉取失敗

**症狀：** 無法從 GitHub Container Registry 拉取映像

**解決方案：**
1. 確認 GitHub Token 有權限訪問 Container Registry
2. 檢查映像是否成功構建並推送
3. 在服務器上手動測試：`docker pull ghcr.io/username/repo:latest`

### 問題 3: 容器啟動失敗

**症狀：** 容器無法啟動或立即退出

**解決方案：**
1. 查看容器日誌：`docker logs go-gin-crud-app`
2. 檢查環境變數是否正確設置
3. 確認數據庫連接是否正常
4. 檢查端口是否被占用：`netstat -tulpn | grep 8080`

### 問題 4: 健康檢查失敗

**症狀：** 部署後健康檢查返回錯誤

**解決方案：**
1. 確認應用已正確啟動：`docker ps`
2. 檢查應用日誌：`docker logs go-gin-crud-app`
3. 驗證端口映射是否正確
4. 測試本地健康檢查：`curl http://localhost:8080/health`

## 回滾部署

如果需要回滾到之前的版本：

```bash
# SSH 到服務器
ssh deploy@your-server

# 查看可用的映像
docker images | grep go-gin-crud

# 停止當前容器
docker stop go-gin-crud-app
docker rm go-gin-crud-app

# 使用之前的映像啟動
docker run -d \
  --name go-gin-crud-app \
  --restart unless-stopped \
  -p 8080:8080 \
  --env-file /opt/go-gin-crud/.env \
  ghcr.io/username/repo:previous-tag
```

## 監控和日誌

### 查看實時日誌

```bash
docker logs -f go-gin-crud-app
```

### 查看資源使用

```bash
docker stats go-gin-crud-app
```

### 設置日誌輪轉

在服務器上配置 Docker 日誌輪轉：

```bash
# 編輯 Docker daemon 配置
sudo nano /etc/docker/daemon.json

# 添加以下內容
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}

# 重啟 Docker
sudo systemctl restart docker
```

## 安全建議

1. **使用專用部署用戶**
   - 創建專用的 `deploy` 用戶
   - 限制該用戶的權限
   - 使用 `sudo` 僅授予必要的權限

2. **保護敏感信息**
   - 不要將 `.env` 文件提交到倉庫
   - 使用強密碼和密鑰
   - 定期輪換密鑰和密碼

3. **網絡安全**
   - 使用防火牆限制訪問
   - 考慮使用 VPN 或私有網絡
   - 啟用 HTTPS（使用 Nginx 反向代理）

4. **定期更新**
   - 定期更新 Docker 映像
   - 更新系統和依賴
   - 監控安全公告

## 進階配置

### 使用 Nginx 反向代理

```nginx
server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 設置 SSL/TLS

使用 Let's Encrypt 獲取免費 SSL 證書：

```bash
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d api.example.com
```

## 相關文檔

- [GitHub Actions 工作流說明](.github/workflows/README.md)
- [Docker 配置](Dockerfile)
- [Docker Compose 配置](docker-compose.yml)

