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

