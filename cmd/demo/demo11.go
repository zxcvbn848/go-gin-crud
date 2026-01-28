package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Demo11: 資源競爭控制（Resource Contention）
// 詳細說明請參考 cmd/concurrency/README.md
//
// 目標：
// - 演示多個 goroutine 競爭同一資源時如何保證互斥訪問
// - 對比 Mutex / RWMutex / Channel 作為鎖
// - 演示如何使用 Context 做超時控制，避免 goroutine 永遠阻塞
func demo11() {
	fmt.Println("\n=== Demo 11: 資源競爭控制（Resource Contention）===")

	fmt.Println("\n--- 場景 1: 不加鎖的資源競爭（錯誤示範）---")
	raceConditionWithoutLock()

	fmt.Println("\n--- 場景 2: 使用 Mutex 保證互斥 ---")
	raceConditionWithMutex()

	fmt.Println("\n--- 場景 3: 使用 RWMutex 提高併發讀性能 ---")
	raceConditionWithRWMutex()

	fmt.Println("\n--- 場景 4: 使用 Channel 作為鎖 + Context 超時 ---")
	raceConditionWithChannelLockAndTimeout()

	fmt.Println("\nDemo 11 完成")
}

// ==================== 場景 1: 不加鎖的資源競爭（錯誤示範） ====================

// raceConditionWithoutLock 演示多個 goroutine 併發寫共享資源但不加鎖，導致資料錯亂
func raceConditionWithoutLock() {
	counter := 0
	var wg sync.WaitGroup

	workerCount := 5
	incrementPerWorker := 1000

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < incrementPerWorker; j++ {
				// ❌ 未加鎖的共享變數寫入，存在資料競爭
				counter++
			}
		}(i)
	}

	wg.Wait()

	expected := workerCount * incrementPerWorker
	fmt.Printf("  [未加鎖] counter = %d，預期 = %d（結果通常會小於預期，因為發生競態條件）\n",
		counter, expected)
}

// ==================== 場景 2: 使用 Mutex 保證互斥 ====================

// raceConditionWithMutex 使用 sync.Mutex 保證對共享資源的互斥訪問
func raceConditionWithMutex() {
	counter := 0
	var wg sync.WaitGroup
	var mu sync.Mutex

	workerCount := 5
	incrementPerWorker := 1000

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < incrementPerWorker; j++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	expected := workerCount * incrementPerWorker
	fmt.Printf("  [Mutex] counter = %d，預期 = %d（使用互斥鎖，結果應該一致）\n",
		counter, expected)
}

// ==================== 場景 3: 使用 RWMutex 提高併發讀性能 ====================

// raceConditionWithRWMutex 使用 sync.RWMutex 讓多個讀操作併發、寫操作互斥
func raceConditionWithRWMutex() {
	type Config struct {
		Value int
	}

	var (
		cfg = &Config{Value: 0}
		mu  sync.RWMutex
		wg  sync.WaitGroup
	)

	// 模擬多個讀者和少量寫者
	readerCount := 5
	writerCount := 2
	writesPerWriter := 3

	readConfig := func(id int) {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			mu.RLock()
			v := cfg.Value
			fmt.Printf("  [RWMutex][讀者-%d] 讀取配置 Value = %d\n", id, v)
			mu.RUnlock()
			time.Sleep(50 * time.Millisecond)
		}
	}

	writeConfig := func(id int) {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			mu.Lock()
			cfg.Value++
			fmt.Printf("  [RWMutex][寫者-%d] 更新配置 Value => %d\n", id, cfg.Value)
			mu.Unlock()
			time.Sleep(120 * time.Millisecond)
		}
	}

	// 啟動讀者
	for i := 0; i < readerCount; i++ {
		wg.Add(1)
		go readConfig(i)
	}

	// 啟動寫者
	for i := 0; i < writerCount; i++ {
		wg.Add(1)
		go writeConfig(i)
	}

	wg.Wait()
	expected := writerCount * writesPerWriter
	fmt.Printf("  [RWMutex] 最終配置 Value = %d，預期 = %d（所有寫入都在鎖保護下，結果應一致）\n",
		cfg.Value, expected)
}

// ==================== 場景 4: 使用 Channel 作為鎖 + Context 超時 ====================

// resource 表示需要互斥訪問的資源，這裡用簡單的切片模擬文件寫入
type resource struct {
	data []string
	mu   sync.Mutex // 只在 demo 中保護打印，實際鎖由 channel 控制
}

func (r *resource) write(id int, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data = append(r.data, value)
	fmt.Printf("  [Channel 鎖][Worker-%d] 寫入: %s\n", id, value)
}

// lockWithContext 嘗試在 ctx 規定的時間內獲取鎖（channel），如果超時則返回 false
func lockWithContext(ctx context.Context, lockChan chan struct{}) bool {
	select {
	case lockChan <- struct{}{}:
		// 獲取鎖成功
		return true
	case <-ctx.Done():
		// 超時或被取消
		return false
	}
}

// unlock 釋放鎖
func unlock(lockChan chan struct{}) {
	select {
	case <-lockChan:
	default:
	}
}

// raceConditionWithChannelLockAndTimeout 使用帶緩衝為 1 的 channel 作為鎖，配合 Context 控制超時
func raceConditionWithChannelLockAndTimeout() {
	var (
		res      = &resource{data: make([]string, 0)}
		lockChan = make(chan struct{}, 1) // 緩衝為 1 的 channel 作為互斥鎖
		wg       sync.WaitGroup
	)

	workerCount := 5

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 每個 worker 嘗試多次寫入
			for j := 0; j < 3; j++ {
				// 為每次寫入設置超時，避免永遠阻塞
				ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
				ok := lockWithContext(ctx, lockChan)
				cancel()

				if !ok {
					fmt.Printf("  [Channel 鎖][Worker-%d] ❌ 獲取鎖超時，放棄此次寫入（避免阻塞）\n", id)
					time.Sleep(50 * time.Millisecond)
					continue
				}

				// 獲取鎖成功，安全訪問共享資源
				res.write(id, fmt.Sprintf("worker-%d: write-%d", id, j))

				// 模擬寫入耗時
				time.Sleep(80 * time.Millisecond)

				// 釋放鎖
				unlock(lockChan)
			}
		}(i)
	}

	wg.Wait()

	// 最後輸出資源內容
	fmt.Println("  [Channel 鎖] 最終資源內容：")
	for i, v := range res.data {
		fmt.Printf("    #%d -> %s\n", i, v)
	}

	fmt.Println("  [Channel 鎖] 💡 關鍵點：")
	fmt.Println("    - 使用緩衝為 1 的 channel 作為互斥鎖")
	fmt.Println("    - 結合 Context 控制獲取鎖的超時，避免 goroutine 永遠阻塞")
	fmt.Println("    - 適用於需要嘗試鎖 / 超時退避的場景，例如高併發下的資源訪問")
}
