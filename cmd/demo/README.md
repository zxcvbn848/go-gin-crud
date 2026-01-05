# Go 併發編程 Demo

本目錄包含多個 Go 併發編程的示例，展示 Goroutine、Channel 和 Context.Context 的使用。

## Demo 列表

- **Demo 1**: 基本 Goroutine + Channel 使用
- **Demo 2**: Context 取消和超時
- **Demo 3**: Worker Pool 模式（結合 Goroutine + Channel + Context）
- **Demo 4**: 使用 Context 傳遞值
- **Demo 5**: 優雅關閉（Graceful Shutdown）
- **Demo 6**: 避免緩存擊穿（Cache Penetration）
- **Demo 7**: 連接池管理（Connection Pool）
- **Demo 8**: 批量處理（Batch Processing）

## 運行方式

**重要：** 必須運行整個包，不能只運行 `main.go` 文件！

```bash
# 運行所有 demo（推薦方式）
go run ./cmd/demo

# 或者使用通配符運行所有文件
go run cmd/demo/*.go

# 運行指定的 demo
go run ./cmd/demo 1        # 只運行 demo1
go run ./cmd/demo 1 2 3    # 運行 demo1, demo2, demo3
go run ./cmd/demo 1-3      # 運行 demo1 到 demo3
go run ./cmd/demo 6        # 只運行 demo6

# ❌ 錯誤方式（會出現 undefined 錯誤）
# go run cmd/demo/main.go 6
```

**為什麼？**
- 使用 `go run cmd/demo/main.go` 時，Go 只編譯 `main.go` 文件
- 不會自動包含同目錄下的其他文件（demo1.go, demo2.go 等）
- 因此會出現 `undefined: demo1` 等錯誤
- 使用 `go run ./cmd/demo` 會編譯整個包的所有文件

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

2. **用戶身份資訊**
   ```go
   // 認證後設置用戶資訊
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
- ✅ 存儲：用戶 ID、請求 ID、追蹤 ID、認證資訊
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

## Demo 6: 避免緩存擊穿（Cache Penetration）

### 什麼是緩存擊穿？

緩存擊穿是指當熱點數據的緩存過期時，大量併發請求同時穿透緩存直接查詢數據庫，導致數據庫壓力過大。

### 問題場景

**無保護的情況：**
- 緩存過期
- 10 個併發請求同時發現緩存未命中
- 10 個請求同時查詢數據庫
- 數據庫壓力暴增

### 解決方案

使用 `sync.Once` 確保只有第一個請求去查詢數據庫，其他請求等待第一個查詢結果。

**核心技術：**
1. **`sync.Once`**: 確保查詢操作只執行一次
2. **Channel**: 用於傳遞查詢結果給等待的 goroutine
3. **Context**: 控制查詢超時
4. **錯誤處理**: 處理查詢失敗的情況

### 實現要點

```go
// 1. 使用 sync.Once 確保只查詢一次
once := &sync.Once{}
resultChan := make(chan queryResult, 1)

once.Do(func() {
    // 只有第一個 goroutine 會執行這裡
    value, err := queryDB(key)
    resultChan <- queryResult{value: value, err: err}
})

// 2. 所有 goroutine 等待結果
result := <-resultChan
```

### Demo 6 包含三個場景

1. **場景 1**: 緩存擊穿問題演示（無保護）
   - 展示問題：10 個請求同時查詢數據庫
   - 結果：數據庫被查詢 10 次，壓力暴增

2. **場景 2**: 使用 `sync.Once` 解決方案
   - 展示解決：只有 1 個請求查詢數據庫，其他等待結果
   - 結果：數據庫只被查詢 1 次，壓力減少 90%

3. **場景 3**: 帶超時和錯誤處理
   - 正常查詢
   - 查詢失敗處理
   - 超時處理

### 為什麼兩個場景的耗時差不多？

**重要理解：** 避免緩存擊穿的主要目的不是減少總耗時，而是**減少數據庫壓力**！

#### 場景 1（無保護）：
- 10 個 goroutine **併發**執行
- 每個都查詢數據庫（500ms）
- 總耗時：約 500ms（因為併發，取最慢的那個）
- 數據庫壓力：**10 次查詢**

#### 場景 2（有保護）：
- 10 個 goroutine **併發**執行
- 只有第一個查詢數據庫（500ms）
- 其他 9 個等待第一個的結果
- 總耗時：約 500ms（等待第一個查詢完成）
- 數據庫壓力：**1 次查詢**

#### 關鍵差異：

| 指標 | 場景 1（無保護） | 場景 2（有保護） | 差異 |
|------|-----------------|----------------|------|
| 總耗時 | ~500ms | ~500ms | 相同 |
| 數據庫查詢次數 | 10 次 | 1 次 | **減少 90%** |
| 數據庫 CPU 使用 | 10 倍 | 1 倍 | **減少 90%** |
| 數據庫連接數 | 10 個 | 1 個 | **減少 90%** |
| 數據庫內存使用 | 10 倍 | 1 倍 | **減少 90%** |

**結論：** 雖然總耗時差不多，但在高併發場景下，避免緩存擊穿可以：
- ✅ 保護數據庫不被壓垮
- ✅ 減少數據庫資源消耗
- ✅ 提高系統穩定性
- ✅ 避免數據庫連接池耗盡

### 運行 Demo 6

```bash
# 只運行 demo6
go run ./cmd/demo 6
```

---

## Demo 7: 連接池管理（Connection Pool）

### 什麼是連接池？

連接池是一種資源管理技術，用於複用昂貴的資源（如數據庫連接、HTTP 連接等），避免頻繁建立和關閉連接，從而提高系統性能。

### 問題場景

**無連接池的情況：**
- 每個請求都需要建立新連接
- 請求結束後立即關閉連接
- 頻繁建立/關閉連接導致性能下降

### 解決方案

使用 Channel 作為連接池，複用連接資源。

**核心技術：**
1. **Channel**: 作為連接池存儲連接
2. **Context**: 控制連接獲取超時
3. **WaitGroup**: 等待連接歸還
4. **優雅關閉**: 確保所有連接正確關閉

### 實現要點

```go
// 1. 使用 Channel 作為連接池
pool := make(chan Connection, maxSize)

// 2. 獲取連接（帶超時）
select {
case conn := <-pool:
    // 從池中獲取連接
case <-ctx.Done():
    // 超時，創建新連接
}

// 3. 歸還連接
select {
case pool <- conn:
    // 成功歸還
default:
    // 池已滿，關閉連接
}
```

### Demo 7 包含三個場景

1. **場景 1**: 無連接池問題演示
   - 展示問題：頻繁建立/關閉連接

2. **場景 2**: 使用連接池
   - 展示解決：連接複用，性能提升

3. **場景 3**: 連接池超時和優雅關閉
   - 正常使用
   - 超時處理
   - 優雅關閉
   - 關閉後處理

### 運行 Demo 7

```bash
# 只運行 demo7
go run ./cmd/demo 7
```

---

## Demo 8: 批量處理（Batch Processing）

### 什麼是批量處理？

批量處理是一種優化技術，將多個小任務累積起來，達到一定數量或時間後再一次性處理，從而減少處理開銷，提高系統效率。

### 問題場景

**無批量處理的情況：**
- 每個任務都立即處理
- 頻繁的處理開銷（如數據庫連接、網絡請求）
- 系統資源浪費

### 解決方案

使用 Channel 累積任務，通過數量或時間觸發批量處理。

**核心技術：**
1. **Channel**: 累積任務
2. **time.Ticker**: 定時觸發批量處理
3. **Context**: 控制關閉
4. **Mutex**: 保護共享狀態

### 實現要點

```go
// 1. 基於數量的批量處理
if len(batch) >= batchSize {
    processBatch()
}

// 2. 基於時間的批量處理
ticker := time.NewTicker(interval)
case <-ticker.C:
    processBatch()

// 3. 混合觸發（數量 + 時間）
if len(batch) >= batchSize || <-ticker.C {
    processBatch()
}
```

### Demo 8 包含四個場景

1. **場景 1**: 無批量處理問題演示
   - 展示問題：逐個處理，效率低

2. **場景 2**: 基於數量的批量處理
   - 達到指定數量後觸發處理
   - 適合高頻任務場景

3. **場景 3**: 基於時間的批量處理
   - 定時觸發批量處理
   - 適合低頻但需要及時處理的場景

4. **場景 4**: 混合觸發（數量 + 時間）
   - 達到數量或時間任一條件即觸發
   - 兼顧效率和及時性

### 運行 Demo 8

```bash
# 只運行 demo8
go run ./cmd/demo 8
```

---

## 相關資源

高併發情境題和面試問題請參考：[`cmd/concurrency/README.md`](../concurrency/README.md)
