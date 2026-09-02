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
  - [x] GitHub Actions
  - [ ] 架設 EC2
  - [ ] 架設 DB
- [ ] 開發簡易的前端頁面

## Resilience Patterns 🛡️

> 流程圖與邊界守衛說明見 [`docs/RESILIENCE_PATTERNS.md`](docs/RESILIENCE_PATTERNS.md)

- [x] Timeout（請求/任務執行超時控制）
  - 任務層：`ExecuteTask` 用 `context.WithTimeout`
  - 請求層：`internal/middleware/timeout.go`，把帶 deadline 的 ctx 換進 `c.Request`，
  讓 GORM / go-redis 提早失敗；`/stream/`、`/socket.io/`、`/tasks/` 豁免
  - 逾時長度由 `REQUEST_TIMEOUT` 環境變數控制，預設 10s
- [x] Retry with Backoff（指數退避重試機制）
- [x] Circuit Breaker（熔斷器，防止雪崩效應）
  - 包在 Redis 快取讀取上，見 `internal/redis/breaker.go`（用 go-redis Hook，快取層未改動）
  - 判斷條件為**滑動視窗失敗率**（30s / 6 桶環形緩衝，失敗率 50%、最小樣本 20 筆），
  不是連續失敗計數 —— 後者在混合流量（大量成功穿插失敗）下永遠不會熔斷
- [ ] Bulkhead（艙壁隔離，限制資源佔用範圍）
  - **刻意延後**：目前沒有多個下游在搶同一個資源池，現在加會是沒有保護對象的玩具
  - 觸發條件：等「統計報表 API」做完。報表是長時間全表掃描，會跟登入等 API 搶
  `SetMaxOpenConns(100)`（`internal/database/database.go:54`）那一池連線 —— 幾個報表請求
  吃光連線就會拖垮全站，那才是真的需要隔離
  - 屆時做法：用號誌（緩衝 channel，或已是 indirect 依賴的 `golang.org/x/sync/semaphore`）
  把報表查詢限制在 5 條連線內
- [x] Fallback（降級處理，回傳預設值或快取）
  - 未新增程式碼。service 層本來就是 `if err == nil && cached != nil`，快取出錯自動回 DB
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
    - [https://blackbing.medium.com/%E6%B7%BA%E8%AB%87-server-sent-events-9c81ef21ca8e?sharedUserId=andywang_10320](https://blackbing.medium.com/%E6%B7%BA%E8%AB%87-server-sent-events-9c81ef21ca8e?sharedUserId=andywang_10320)
- [ ] 影音串流協定 MPEG-DASH / HLS，可用 CMAF 同時處理兩種協定
- [ ] 評論系統
- [ ] 統計報表 API
- [x] CI/CD 整合

詳細功能規劃請參考 [FEATURE_PLAN.md](./FEATURE_PLAN.md)
