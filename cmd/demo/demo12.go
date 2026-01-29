package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Demo12: 優雅降級（Graceful Degradation）
// 詳細說明請參考 cmd/concurrency/README.md
//
// 目標：
// - 當系統負載過高時，不是整體崩潰，而是「關閉非核心功能 / 降級行為」
// - 演示如何檢測負載（goroutine 數量）
// - 使用 Context 傳遞「是否處於降級模式」的標誌
// - 使用 Channel 控制流量（限制併發請求數）
func demo12() {
	fmt.Println("\n=== Demo 12: 優雅降級（Graceful Degradation）===")

	fmt.Println("\n--- 場景 1: 無降級策略（錯誤示範）---")
	noDegradeUnderHighLoad()

	fmt.Println("\n--- 場景 2: 基於 goroutine 數量的簡單降級 ---")
	degradeByGoroutineCount()

	fmt.Println("\n--- 場景 3: 使用 Context 傳遞降級標誌 ---")
	degradeWithContextFlag()

	fmt.Println("\n--- 場景 4: 使用 Channel 控制流量（限流 + 降級） ---")
	degradeWithChannelLimit()

	fmt.Println("\nDemo 12 完成")
}

// ==================== 場景 1: 無降級策略（錯誤示範） ====================

// noDegradeUnderHighLoad 模擬在高併發下，系統沒有任何降級策略，可能導致 goroutine 瘋狂增長
func noDegradeUnderHighLoad() {
	var wg sync.WaitGroup

	handleRequest := func(id int) {
		defer wg.Done()
		// 模擬一些耗時操作（例如：外部 API 調用）
		fmt.Printf("  [無降級] 處理請求 #%d，中...\n", id)
		time.Sleep(150 * time.Millisecond)
	}

	requestCount := 50

	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go handleRequest(i)
	}

	// 稍等一段時間後觀察 goroutine 數量
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("  [無降級] 當前 goroutine 數量: %d（請求數量: %d）\n", runtime.NumGoroutine(), requestCount)

	wg.Wait()
	fmt.Println("  [無降級] ✅ 所有請求處理完成，但在高負載情況下可能導致系統不穩定")
}

// ==================== 場景 2: 基於 goroutine 數量的簡單降級 ====================

// degradeByGoroutineCount 根據當前 goroutine 數量決定是否降級：超過閾值就拒絕非核心請求
func degradeByGoroutineCount() {
	var wg sync.WaitGroup

	const (
		totalRequests       = 50
		degradeThreshold    = 100 // 當前 goroutine 超過此值時觸發降級
		coreRequestModulo   = 5   // 每 5 個請求中，1 個是「核心請求」
		simulatedHandleTime = 150 * time.Millisecond
	)

	handleCoreRequest := func(id int) {
		defer wg.Done()
		fmt.Printf("  [降級-核心] ✅ 處理核心請求 #%d\n", id)
		time.Sleep(simulatedHandleTime)
	}

	handleNonCoreRequest := func(id int) {
		defer wg.Done()

		currentG := runtime.NumGoroutine()
		if currentG > degradeThreshold {
			// 觸發降級：直接拒絕 / 快速返回緩存 / 簡化處理
			fmt.Printf("  [降級-非核心] ⚠️ 高負載（goroutine=%d），降級處理，拒絕非核心請求 #%d\n", currentG, id)
			return
		}

		fmt.Printf("  [降級-非核心] 處理非核心請求 #%d\n", id)
		time.Sleep(simulatedHandleTime)
	}

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		// 每 coreRequestModulo 個請求中，只有第一個視為核心請求
		if i%coreRequestModulo == 0 {
			go handleCoreRequest(i)
		} else {
			go handleNonCoreRequest(i)
		}
	}

	wg.Wait()
	fmt.Println("  [降級] ✅ 基於 goroutine 數量的簡單降級示例完成")
}

// ==================== 場景 3: 使用 Context 傳遞降級標誌 ====================

type degradationKey struct{}

// withDegradationFlag 在 Context 中標記當前是否為降級模式
func withDegradationFlag(ctx context.Context, degraded bool) context.Context {
	return context.WithValue(ctx, degradationKey{}, degraded)
}

// isDegraded 從 Context 中讀取是否處於降級模式
func isDegraded(ctx context.Context) bool {
	if v := ctx.Value(degradationKey{}); v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// degradeWithContextFlag 演示如何用 Context 在整條調用鏈中傳遞「降級模式」資訊
func degradeWithContextFlag() {
	// 模擬一個後端服務，根據降級模式決定行為
	type Response struct {
		Source  string
		Message string
		From    string
	}

	callBackend := func(ctx context.Context, reqID int) Response {
		if isDegraded(ctx) {
			// 降級模式下：返回緩存 / 簡化結果 / 跳過一些昂貴操作
			return Response{
				Source:  "BackendService",
				Message: fmt.Sprintf("降級模式：返回緩存數據（req #%d）", reqID),
				From:    "cache",
			}
		}

		// 正常模式：做完整處理
		time.Sleep(80 * time.Millisecond)
		return Response{
			Source:  "BackendService",
			Message: fmt.Sprintf("正常模式：完整計算結果（req #%d）", reqID),
			From:    "primary",
		}
	}

	handleRequest := func(ctx context.Context, reqID int) {
		resp := callBackend(ctx, reqID)
		mode := "正常"
		if isDegraded(ctx) {
			mode = "降級"
		}
		fmt.Printf("  [Context 降級] 模式=%s，請求 #%d，結果: %s（from=%s）\n",
			mode, reqID, resp.Message, resp.From)
	}

	// 模擬兩種模式：正常 + 降級
	ctxNormal := context.Background()
	ctxDegraded := withDegradationFlag(context.Background(), true)

	handleRequest(ctxNormal, 1)
	handleRequest(ctxNormal, 2)

	handleRequest(ctxDegraded, 3)
	handleRequest(ctxDegraded, 4)

	fmt.Println("  [Context 降級] ✅ 使用 Context 傳遞降級標誌示例完成")
}

// ==================== 場景 4: 使用 Channel 控制流量（限流 + 降級） ====================

// degradeWithChannelLimit 使用帶容量的 Channel 作為「令牌桶 / 併發閥門」，超過容量時觸發降級
func degradeWithChannelLimit() {
	var wg sync.WaitGroup

	const (
		totalRequests   = 40
		maxConcurrent   = 5 // 最多允許同時處理的請求數（核心能力）
		degradeWaitTime = 50 * time.Millisecond
	)

	// semaphore channel 控制同時處理的核心請求數
	sem := make(chan struct{}, maxConcurrent)

	handleRequest := func(id int) {
		defer wg.Done()

		select {
		case sem <- struct{}{}:
			// 成功獲取「令牌」，作為核心請求處理
			fmt.Printf("  [Channel 限流] ✅ 核心處理請求 #%d\n", id)
			time.Sleep(150 * time.Millisecond)
			<-sem
		default:
			// 無法即時獲取令牌，啟用降級策略：快速返回 / 排隊等待 / 返回緩存
			fmt.Printf("  [Channel 限流] ⚠️ 併發已達上限，對請求 #%d 進行降級處理（快速返回）\n", id)
			time.Sleep(degradeWaitTime)
		}
	}

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go handleRequest(i)
	}

	wg.Wait()
	fmt.Println("  [Channel 限流] ✅ 使用 Channel 控制流量 + 優雅降級示例完成")
}
