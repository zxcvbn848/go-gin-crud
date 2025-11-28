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
- [ ] 結構化日誌系統（logrus 或 zap）
- [x] 健康檢查端點（/health）
- [ ] 使用者密碼修改功能

## 中優先級 📋
- [ ] Rate Limiting（API 限流）
- [ ] 檔案上傳功能（圖片）
- [ ] 軟刪除（Soft Delete）
- [ ] 操作日誌記錄（Audit Log）
- [ ] Email 驗證（註冊確認信）
- [ ] 瞭解 Goroutine + channel
- [ ] 知道 context.Context 用法
- [ ] 寫一個簡單的併發 demo（面試常考）

## 進階功能 🚀
- [ ] Redis 快取整合
- [ ] 全文搜尋（Elasticsearch）
- [ ] 評論系統
- [ ] 統計報表 API
- [ ] CI/CD 整合

詳細功能規劃請參考 [FEATURE_PLAN.md](./FEATURE_PLAN.md)
