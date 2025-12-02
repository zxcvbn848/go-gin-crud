# Go 併發編程 Demo

本目錄包含多個 Go 併發編程的示例，展示 Goroutine、Channel 和 Context.Context 的使用。

## Demo 列表

- **Demo 1**: 基本 Goroutine + Channel 使用
- **Demo 2**: Context 取消和超時
- **Demo 3**: Worker Pool 模式（結合 Goroutine + Channel + Context）
- **Demo 4**: 使用 Context 傳遞值
- **Demo 5**: 優雅關閉（Graceful Shutdown）

## 運行方式

```bash
# 運行所有 demo
go run cmd/demo/main.go

# 運行指定的 demo
go run cmd/demo/main.go 1        # 只運行 demo1
go run cmd/demo/main.go 1 2 3    # 運行 demo1, demo2, demo3
go run cmd/demo/main.go 1-3      # 運行 demo1 到 demo3
```

## Demo 4: 使用 Context 傳遞值

### 為什麼要用 Context 傳值？

在 Go 的併發編程中，Context 傳值是一個重要的模式，用於在請求鏈中傳遞請求範圍的數據。

#### 1. 為什麼不能直接用 string 作為 key？

**錯誤示範：**
```go
// ❌ 不推薦：使用 string 作為 key
ctx := context.WithValue(context.Background(), "userID", "12345")
```

**問題：**
- 不同包可能使用相同的 string key，導致 key 衝突
- 無法保證類型安全
- 容易造成數據覆蓋或讀取錯誤

**正確示範：**
```go
// ✅ 推薦：使用自定義類型作為 key
type contextKey string

const (
    userIDKey contextKey = "userID"
)

ctx := context.WithValue(context.Background(), userIDKey, "12345")
```

**優點：**
- 類型安全：編譯時就能發現錯誤
- 避免衝突：不同包定義的 contextKey 類型不同，不會衝突
- 代碼清晰：明確表示這是 context 的 key

#### 2. 為什麼要用 Context 傳值而不是函數參數？

**場景對比：**

| 方式 | 函數參數 | Context 傳值 |
|:----|:--------|:------------|
| 適用場景 | 簡單的函數調用 | 跨多層函數、goroutine 傳遞 |
| 參數數量 | 參數列表會很長 | 不需要修改函數簽名 |
| 中間層處理 | 需要每層都傳遞 | 自動傳遞，中間層不需要關心 |
| 併發安全 | 需要額外處理 | 每個請求獨立的 context |
| 使用場景 | 簡單調用鏈 | HTTP 請求、RPC 調用、日誌追蹤 |

**實際應用場景：**

1. **HTTP 請求追蹤**
   ```go
   // 在 middleware 中設置
   ctx := context.WithValue(r.Context(), requestIDKey, generateRequestID())
   
   // 在任意深度的函數中獲取
   requestID := ctx.Value(requestIDKey).(string)
   ```

2. **用戶身份信息**
   ```go
   // 認證後設置用戶信息
   ctx := context.WithValue(ctx, userIDKey, user.ID)
   
   // 在業務邏輯中直接使用，不需要每層都傳 userID
   userID := ctx.Value(userIDKey).(string)
   ```

3. **日誌追蹤**
   ```go
   // 設置追蹤 ID
   ctx := context.WithValue(ctx, traceIDKey, traceID)
   
   // 所有日誌自動包含追蹤 ID
   logger.WithContext(ctx).Info("處理請求")
   ```

#### 3. Context 傳值的限制和注意事項

**限制：**
- Context 不是數據庫，不應該存儲大量數據
- 只適合存儲請求範圍的數據（request-scoped data）
- 數據應該是不可變的（immutable）

**最佳實踐：**
- ✅ 存儲：用戶 ID、請求 ID、追蹤 ID、認證信息
- ❌ 不要存儲：大量數據、可變對象、業務邏輯狀態

**類型斷言安全：**
```go
// 安全的類型斷言
if userID, ok := ctx.Value(userIDKey).(string); ok {
    // 使用 userID
} else {
    // 處理 key 不存在的情況
}
```

### 實現示例

```go
// 定義 context key 類型
type contextKey string

const (
    userIDKey    contextKey = "userID"
    requestIDKey contextKey = "requestID"
)

// 設置值
ctx := context.WithValue(context.Background(), userIDKey, "12345")
ctx = context.WithValue(ctx, requestIDKey, "req-abc-123")

// 獲取值
userID := ctx.Value(userIDKey).(string)
requestID := ctx.Value(requestIDKey).(string)
```

## Demo 5: 優雅關閉 vs 普通關閉

### 關鍵區別對比

| 特性                    | 普通關閉                    | 優雅關閉                           |
| :---------------------- | :------------------------- | :--------------------------------- |
| 等待任務完成            | ❌ 不等待                  | ✅ 等待 (wg.Wait())                |
| 處理剩餘數據            | ❌ 可能丟失                | ✅ 處理完所有剩餘消息              |
| 資源清理                | ❌ 不清理                  | ✅ 清理資源                        |
| 數據一致性              | ❌ 可能不一致              | ✅ 保證一致性                      |
| 使用場景                | 開發/測試                  | 生產環境                           |

### 優雅關閉的核心要素

1. **Context 取消信號**：通知所有 goroutine 準備關閉
2. **WaitGroup**：等待所有 goroutine 完成
3. **資源清理**：在退出前清理所有資源
4. **剩餘任務處理**：確保不丟失正在處理的數據

### 為什麼需要優雅關閉？

在生產環境中，直接關閉程序可能導致：
- 正在處理的請求被中斷
- 數據庫事務未提交
- 文件未正確保存
- 連接未正確關閉（資源泄漏）
- 用戶體驗差（請求失敗）

優雅關閉可以確保：
- 所有正在處理的任務完成
- 所有數據正確保存
- 所有資源正確清理
- 服務平滑停止，不影響用戶

### 實現要點

```go
// 1. 創建 context 用於取消信號
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// 2. 使用 WaitGroup 追蹤 goroutine
var wg sync.WaitGroup

// 3. 在 goroutine 中監聽取消信號
go func() {
    defer wg.Done()
    for {
        select {
        case <-ctx.Done():
            // 清理資源
            cleanup()
            return
        case <-ticker.C:
            // 正常處理
        }
    }
}()

// 4. 發送關閉信號
cancel()

// 5. 等待所有 goroutine 完成
wg.Wait()
```

---

## 相關資源

高併發情境題和面試問題請參考：[`cmd/concurrency/README.md`](../concurrency/README.md)
