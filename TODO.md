# TODO

## 已完成 ✅

- [x] 黃金架構（分層：controller/service/repository）
- [x] JWT Refresh Token（雙 token）
- [x] Blacklist Token（登出）
- [x] 分頁搜尋 API
- [x] 全套 CRUD
  - [x] User
  - [x] Product
  - [x] Posts
- [x] RBAC 權限角色
- [x] 服務部署（Docker）
- [x] 環境變數管理（.env 檔案）

## 高優先級 🔥

- [x] JWT Secret 從環境變數讀取（目前硬編碼）
- [x] 輸入驗證（使用 validator 套件）
- [x] API 測試
- [x] 結構化日誌系統（logrus 或 zap）
- [x] 健康檢查端點（/health）
- [x] 使用者密碼修改功能

## 中優先級 📋

- [x] Rate Limiting （API 限流）
- [x] 檔案上傳功能（圖片）
- [x] 軟刪除（Soft Delete）
- [ ] 操作日誌記錄（Audit Log）
- [ ] Email 驗證（註冊確認信）
- [x] 瞭解 Goroutine + channel
- [x] 知道 context.Context 用法
- [x] 寫一個簡單的併發 demo（面試常考）
- [x] 高併發情境題：文件在 [cmd/concurrency/README.md](./cmd/concurrency/README.md)
- [x] 部署到正式站
  - [x] 事前準備 [.github/workflows/README.md](.github/workflows/README.md)
  - [ ] GitHub Actions
  - [ ] 架設 EC2
  - [ ] 架設 DB
- [ ] 開發簡易的前端頁面

## Resilience Patterns 🛡️

- [ ] Timeout（請求/任務執行超時控制）
- [x] Retry with Backoff（指數退避重試機制）
- [ ] Circuit Breaker（熔斷器，防止雪崩效應）
- [ ] Bulkhead（艙壁隔離，限制資源佔用範圍）
- [ ] Fallback（降級處理，回傳預設值或快取）
- [ ] Health Check（依賴服務健康偵測）

## 進階功能 🚀

- [x] Redis 快取整合
  - [x] Books APIs
  - [x] Posts APIs
  - [x] Users APIs
  - [x] Products APIs
- [ ] 全文搜尋（Elasticsearch）
- [x] Socket io
- [x] Streaming
  - 技術文件
    - https://blackbing.medium.com/淺談-server-sent-events-9c81ef21ca8e
- [ ] 影音串流協定 MPEG-DASH / HLS，可用 CMAF 同時處理兩種協定
- [ ] 評論系統
- [ ] 統計報表 API
- [ ] CI/CD 整合

詳細功能規劃請參考 [FEATURE_PLAN.md](./FEATURE_PLAN.md)
