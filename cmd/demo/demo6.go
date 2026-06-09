package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Demo6: 避免緩存擊穿（Cache Penetration）
// 詳細說明請參考 cmd/concurrency/README.md
func demo6() {
	fmt.Println("\n=== Demo 6: 避免緩存擊穿 ===")

	fmt.Println("\n--- 場景 1: 緩存擊穿問題（無保護）---")
	cachePenetrationProblem()

	fmt.Println("\n--- 場景 2: 使用 sync.Once 避免緩存擊穿 ---")
	cachePenetrationSolution()

	fmt.Println("\n--- 場景 3: 帶超時和錯誤處理的緩存查詢 ---")
	cacheWithTimeout()

	fmt.Println("\nDemo 6 完成")
}

// ==================== 場景 1: 緩存擊穿問題 ====================

// cachePenetrationProblem 演示緩存擊穿問題
// 當緩存過期時，大量併發請求同時查詢數據庫
func cachePenetrationProblem() {
	// 模擬緩存
	cache := make(map[string]interface{})
	cacheExpiry := make(map[string]time.Time)
	var cacheMu sync.RWMutex

	// 統計數據庫查詢次數
	var dbQueryCount int64
	var dbQueryMu sync.Mutex

	// 模擬數據庫查詢（耗時操作）
	queryDB := func(key string) (interface{}, error) {
		// 統計查詢次數
		dbQueryMu.Lock()
		dbQueryCount++
		currentCount := dbQueryCount
		dbQueryMu.Unlock()

		fmt.Printf("  [問題] 🔴 查詢數據庫: %s (第 %d 次查詢，這會導致數據庫壓力過大)\n", key, currentCount)
		time.Sleep(500 * time.Millisecond) // 模擬數據庫查詢耗時
		return fmt.Sprintf("數據: %s", key), nil
	}

	// 模擬緩存過期
	cacheMu.Lock()
	cacheExpiry["user:123"] = time.Now().Add(-1 * time.Second) // 已過期
	cacheMu.Unlock()

	// 模擬 10 個併發請求同時查詢
	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			key := "user:123"
			requestStart := time.Now()

			cacheMu.RLock()
			value, exists := cache[key]
			expiry, hasExpiry := cacheExpiry[key]
			cacheMu.RUnlock()

			// 檢查緩存是否過期
			if !exists || (hasExpiry && time.Now().After(expiry)) {
				// 緩存未命中或已過期，查詢數據庫
				result, err := queryDB(key)
				if err != nil {
					fmt.Printf("  [問題] 請求 %d: 查詢失敗: %v\n", id, err)
					return
				}

				// 更新緩存
				cacheMu.Lock()
				cache[key] = result
				cacheExpiry[key] = time.Now().Add(5 * time.Second)
				cacheMu.Unlock()

				requestElapsed := time.Since(requestStart)
				fmt.Printf("  [問題] 請求 %d: 獲取數據: %v (耗時: %v)\n", id, result, requestElapsed)
			} else {
				fmt.Printf("  [問題] 請求 %d: 從緩存獲取: %v\n", id, value)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("  [問題] ⚠️  總耗時: %v，數據庫被查詢了 %d 次（應該只查詢 1 次）\n", elapsed, dbQueryCount)
	fmt.Printf("  [問題] 💡 說明：雖然總耗時約 500ms（因為併發），但數據庫承受了 10 倍的壓力！\n")
}

// ==================== 場景 2: 使用 sync.Once 解決方案 ====================

// cachePenetrationSolution 使用 sync.Once 避免緩存擊穿
// 只有第一個請求去查詢數據庫，其他請求等待結果
func cachePenetrationSolution() {
	// 模擬緩存
	cache := make(map[string]interface{})
	cacheExpiry := make(map[string]time.Time)
	var cacheMu sync.RWMutex

	// 查詢結果類型
	type queryResult struct {
		value interface{}
		err   error
	}

	// 查詢狀態：每個 key 對應一個查詢狀態
	type queryState struct {
		once     *sync.Once
		result   *queryResult  // 共享的查詢結果
		resultMu sync.RWMutex  // 保護結果的讀寫
		ready    chan struct{} // 用於通知結果已準備好
	}

	queryStates := make(map[string]*queryState)
	statesMu := sync.Mutex{}

	// 模擬數據庫查詢（耗時操作）
	queryDB := func(key string) (interface{}, error) {
		fmt.Printf("  [解決] ✅ 查詢數據庫: %s (只有第一個請求會執行)\n", key)
		time.Sleep(500 * time.Millisecond) // 模擬數據庫查詢耗時
		return fmt.Sprintf("數據: %s", key), nil
	}

	// 獲取或創建查詢狀態
	getQueryState := func(key string) *queryState {
		statesMu.Lock()
		defer statesMu.Unlock()
		if state, exists := queryStates[key]; exists {
			return state
		}
		state := &queryState{
			once:  &sync.Once{},
			ready: make(chan struct{}),
		}
		queryStates[key] = state
		return state
	}

	// 查詢函數（帶緩存保護）
	queryWithCache := func(key string) (interface{}, error) {
		// 1. 先檢查緩存
		cacheMu.RLock()
		value, exists := cache[key]
		expiry, hasExpiry := cacheExpiry[key]
		cacheMu.RUnlock()

		if exists && (!hasExpiry || time.Now().Before(expiry)) {
			// 緩存命中
			return value, nil
		}

		// 2. 緩存未命中，使用 sync.Once 確保只查詢一次
		state := getQueryState(key)

		state.once.Do(func() {
			// 只有第一個 goroutine 會執行這裡
			value, err := queryDB(key)
			result := &queryResult{value: value, err: err}

			// 保存結果
			state.resultMu.Lock()
			state.result = result
			state.resultMu.Unlock()

			// 如果成功，更新緩存
			if err == nil {
				cacheMu.Lock()
				cache[key] = value
				cacheExpiry[key] = time.Now().Add(5 * time.Second)
				cacheMu.Unlock()
			}

			// 通知所有等待的 goroutine 結果已準備好
			close(state.ready)

			// 清理：刪除狀態（允許下次緩存過期時重新查詢）
			statesMu.Lock()
			delete(queryStates, key)
			statesMu.Unlock()
		})

		// 3. 等待查詢結果（所有 goroutine 都會等待）
		<-state.ready

		// 4. 讀取結果
		state.resultMu.RLock()
		result := state.result
		state.resultMu.RUnlock()

		return result.value, result.err
	}

	// 模擬緩存過期
	cacheMu.Lock()
	cacheExpiry["user:123"] = time.Now().Add(-1 * time.Second) // 已過期
	cacheMu.Unlock()

	// 模擬 10 個併發請求同時查詢
	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			key := "user:123"
			value, err := queryWithCache(key)
			if err != nil {
				fmt.Printf("  [解決] 請求 %d: 查詢失敗: %v\n", id, err)
				return
			}
			fmt.Printf("  [解決] 請求 %d: 獲取數據: %v\n", id, value)
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("  [解決] ✅ 總耗時: %v，數據庫只被查詢了 1 次（所有請求共享結果）\n", elapsed)
	fmt.Printf("  [解決] 💡 說明：雖然總耗時約 500ms（等待數據庫查詢），但數據庫壓力減少了 90%%！\n")
	fmt.Printf("  [解決] 💡 關鍵優勢：在高併發場景下，避免緩存擊穿可以保護數據庫不被壓垮！\n")
}

// ==================== 場景 3: 帶超時和錯誤處理 ====================

// cacheWithTimeout 帶超時和錯誤處理的緩存查詢
func cacheWithTimeout() {
	// 模擬緩存
	cache := make(map[string]interface{})
	cacheExpiry := make(map[string]time.Time)
	var cacheMu sync.RWMutex

	// 查詢結果類型
	type queryResult struct {
		value interface{}
		err   error
	}

	// 查詢狀態管理
	type queryState struct {
		once       *sync.Once
		resultChan chan queryResult
	}

	queryStates := make(map[string]*queryState)
	statesMu := sync.Mutex{}

	// 獲取或創建查詢狀態
	getQueryState := func(key string) *queryState {
		statesMu.Lock()
		defer statesMu.Unlock()
		if state, exists := queryStates[key]; exists {
			return state
		}
		state := &queryState{
			once:       &sync.Once{},
			resultChan: make(chan queryResult, 1),
		}
		queryStates[key] = state
		return state
	}

	// 模擬數據庫查詢（可能失敗）
	queryDB := func(key string) (interface{}, error) {
		fmt.Printf("  [超時] ✅ 查詢數據庫: %s\n", key)
		time.Sleep(500 * time.Millisecond) // 模擬數據庫查詢耗時

		// 模擬偶爾失敗
		if key == "user:error" {
			return nil, errors.New("數據庫連接失敗")
		}

		return fmt.Sprintf("數據: %s", key), nil
	}

	// 帶超時的查詢函數
	queryWithTimeout := func(ctx context.Context, key string, timeout time.Duration) (interface{}, error) {
		// 1. 先檢查緩存
		cacheMu.RLock()
		value, exists := cache[key]
		expiry, hasExpiry := cacheExpiry[key]
		cacheMu.RUnlock()

		if exists && (!hasExpiry || time.Now().Before(expiry)) {
			return value, nil
		}

		// 2. 創建帶超時的 context
		queryCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// 3. 獲取查詢狀態
		state := getQueryState(key)

		var result queryResult
		state.once.Do(func() {
			// 在 goroutine 中執行查詢，以便可以取消
			go func() {
				value, err := queryDB(key)
				select {
				case state.resultChan <- queryResult{value: value, err: err}:
					// 成功發送結果
					if err == nil {
						cacheMu.Lock()
						cache[key] = value
						cacheExpiry[key] = time.Now().Add(5 * time.Second)
						cacheMu.Unlock()
					}
				case <-queryCtx.Done():
					// 查詢被取消或超時
					fmt.Printf("  [超時] ⏱️  查詢 %s 被取消或超時\n", key)
				}

				// 清理狀態
				statesMu.Lock()
				delete(queryStates, key)
				statesMu.Unlock()
			}()
		})

		// 4. 等待結果或超時
		select {
		case result = <-state.resultChan:
			if result.err != nil {
				return nil, result.err
			}
			return result.value, nil
		case <-queryCtx.Done():
			return nil, fmt.Errorf("查詢超時: %v", queryCtx.Err())
		}
	}

	// 測試場景 1: 正常查詢
	fmt.Println("\n  [超時] 測試 1: 正常查詢（10 個併發請求）")
	var wg sync.WaitGroup
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			value, err := queryWithTimeout(ctx, "user:123", 2*time.Second)
			if err != nil {
				fmt.Printf("  [超時] 請求 %d: 錯誤: %v\n", id, err)
				return
			}
			fmt.Printf("  [超時] 請求 %d: 獲取數據: %v\n", id, value)
		}(i)
	}
	wg.Wait()

	// 測試場景 2: 查詢失敗
	fmt.Println("\n  [超時] 測試 2: 查詢失敗處理")
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := queryWithTimeout(ctx, "user:error", 2*time.Second)
		if err != nil {
			fmt.Printf("  [超時] ✅ 正確處理錯誤: %v\n", err)
		}
	}()
	wg.Wait()

	// 測試場景 3: 超時處理
	fmt.Println("\n  [超時] 測試 3: 超時處理（設置很短的超時時間）")
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 設置超時時間為 100ms，但查詢需要 500ms
		_, err := queryWithTimeout(ctx, "user:timeout", 100*time.Millisecond)
		if err != nil {
			fmt.Printf("  [超時] ✅ 正確處理超時: %v\n", err)
		}
	}()
	wg.Wait()
}
