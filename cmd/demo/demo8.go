package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Demo8: 批量處理（Batch Processing）
// 詳細說明請參考 cmd/concurrency/README.md
func demo8() {
	fmt.Println("\n=== Demo 8: 批量處理 ===")

	fmt.Println("\n--- 場景 1: 無批量處理（逐個處理）---")
	withoutBatchProcessing()

	fmt.Println("\n--- 場景 2: 基於數量的批量處理 ---")
	batchProcessingByCount()

	fmt.Println("\n--- 場景 3: 基於時間的批量處理 ---")
	batchProcessingByTime()

	fmt.Println("\n--- 場景 4: 混合觸發（數量 + 時間）批量處理 ---")
	batchProcessingHybrid()

	fmt.Println("\nDemo 8 完成")
}

// ==================== 場景 1: 無批量處理 ====================

// withoutBatchProcessing 演示無批量處理的問題
func withoutBatchProcessing() {
	// 模擬處理任務
	processTask := func(taskID int) {
		fmt.Printf("  [無批量] 處理任務 #%d\n", taskID)
		time.Sleep(50 * time.Millisecond) // 模擬處理耗時
	}

	// 模擬 20 個任務，逐個處理
	startTime := time.Now()
	for i := 0; i < 20; i++ {
		processTask(i)
	}
	elapsed := time.Since(startTime)
	fmt.Printf("  [無批量] ⚠️  總耗時: %v，處理了 20 個任務（逐個處理）\n", elapsed)
}

// ==================== 場景 2: 基於數量的批量處理 ====================

// batchProcessingByCount 基於數量的批量處理
func batchProcessingByCount() {
	// 任務類型
	type Task struct {
		ID   int
		Data string
	}

	// 批量處理器
	type BatchProcessor struct {
		taskChan    chan Task
		batchSize   int
		batch       []Task
		mu          sync.Mutex
		wg          sync.WaitGroup
		ctx         context.Context
		cancel      context.CancelFunc
		processed   int
		processedMu sync.Mutex
	}

	// 處理批量任務（內部函數，不獲取鎖）
	processBatchInternal := func(processor *BatchProcessor) {
		if len(processor.batch) == 0 {
			return
		}

		// 複製當前批次
		batch := make([]Task, len(processor.batch))
		copy(batch, processor.batch)
		processor.batch = processor.batch[:0] // 清空批次

		// 處理批量任務
		fmt.Printf("  [數量] 📦 開始處理批量任務（共 %d 個）\n", len(batch))
		for _, task := range batch {
			fmt.Printf("  [數量]   處理任務 #%d: %s\n", task.ID, task.Data)
			time.Sleep(10 * time.Millisecond) // 模擬處理耗時
		}

		processor.processedMu.Lock()
		processor.processed += len(batch)
		processor.processedMu.Unlock()

		fmt.Printf("  [數量] ✅ 批量處理完成（共 %d 個任務）\n", len(batch))
	}

	// 創建批量處理器
	createProcessor := func(batchSize int) *BatchProcessor {
		ctx, cancel := context.WithCancel(context.Background())
		processor := &BatchProcessor{
			taskChan:  make(chan Task, 100),
			batchSize: batchSize,
			batch:     make([]Task, 0, batchSize),
			ctx:       ctx,
			cancel:    cancel,
		}

		// 啟動批量處理 goroutine
		processor.wg.Add(1)
		go func() {
			defer processor.wg.Done()
			for {
				select {
				case <-processor.ctx.Done():
					// 處理剩餘任務
					processor.mu.Lock()
					if len(processor.batch) > 0 {
						processBatchInternal(processor)
					}
					processor.mu.Unlock()
					return

				case task := <-processor.taskChan:
					processor.mu.Lock()
					processor.batch = append(processor.batch, task)
					currentBatchSize := len(processor.batch)
					processor.mu.Unlock()

					// 達到批量大小，觸發處理
					if currentBatchSize >= processor.batchSize {
						processor.mu.Lock()
						processBatchInternal(processor)
						processor.mu.Unlock()
					}
				}
			}
		}()

		return processor
	}

	// 創建處理器
	processor := createProcessor(5) // 每 5 個任務觸發一次批量處理

	// 提交任務
	startTime := time.Now()
	for i := 0; i < 20; i++ {
		task := Task{
			ID:   i,
			Data: fmt.Sprintf("任務數據 %d", i),
		}
		processor.taskChan <- task
		fmt.Printf("  [數量] 📥 提交任務 #%d\n", i)
		time.Sleep(20 * time.Millisecond) // 模擬任務提交間隔
	}

	// 等待所有任務處理完成
	time.Sleep(500 * time.Millisecond) // 等待最後一批處理完成
	processor.cancel()
	processor.wg.Wait()

	elapsed := time.Since(startTime)
	processor.processedMu.Lock()
	processed := processor.processed
	processor.processedMu.Unlock()

	fmt.Printf("  [數量] ✅ 總耗時: %v，處理了 %d 個任務（批量處理，每批 5 個）\n", elapsed, processed)
}

// ==================== 場景 3: 基於時間的批量處理 ====================

// batchProcessingByTime 基於時間的批量處理
func batchProcessingByTime() {
	// 任務類型
	type Task struct {
		ID   int
		Data string
	}

	// 批量處理器
	type BatchProcessor struct {
		taskChan    chan Task
		batch       []Task
		mu          sync.Mutex
		wg          sync.WaitGroup
		ctx         context.Context
		cancel      context.CancelFunc
		ticker      *time.Ticker
		interval    time.Duration
		processed   int
		processedMu sync.Mutex
	}

	// 處理批量任務（內部函數，不獲取鎖）
	processBatchInternal := func(processor *BatchProcessor) {
		if len(processor.batch) == 0 {
			return
		}

		// 複製當前批次
		batch := make([]Task, len(processor.batch))
		copy(batch, processor.batch)
		processor.batch = processor.batch[:0] // 清空批次

		// 處理批量任務
		fmt.Printf("  [時間] ⏰ 定時觸發批量處理（共 %d 個任務）\n", len(batch))
		for _, task := range batch {
			fmt.Printf("  [時間]   處理任務 #%d: %s\n", task.ID, task.Data)
			time.Sleep(10 * time.Millisecond) // 模擬處理耗時
		}

		processor.processedMu.Lock()
		processor.processed += len(batch)
		processor.processedMu.Unlock()

		fmt.Printf("  [時間] ✅ 批量處理完成（共 %d 個任務）\n", len(batch))
	}

	// 創建批量處理器
	createProcessor := func(interval time.Duration) *BatchProcessor {
		ctx, cancel := context.WithCancel(context.Background())
		processor := &BatchProcessor{
			taskChan: make(chan Task, 100),
			batch:    make([]Task, 0, 10),
			ctx:      ctx,
			cancel:   cancel,
			ticker:   time.NewTicker(interval),
			interval: interval,
		}

		// 啟動批量處理 goroutine
		processor.wg.Add(1)
		go func() {
			defer processor.wg.Done()
			defer processor.ticker.Stop()

			for {
				select {
				case <-processor.ctx.Done():
					// 處理剩餘任務
					processor.mu.Lock()
					if len(processor.batch) > 0 {
						processBatchInternal(processor)
					}
					processor.mu.Unlock()
					return

				case <-processor.ticker.C:
					// 定時觸發批量處理
					processor.mu.Lock()
					if len(processor.batch) > 0 {
						processBatchInternal(processor)
					}
					processor.mu.Unlock()

				case task := <-processor.taskChan:
					// 接收任務
					processor.mu.Lock()
					processor.batch = append(processor.batch, task)
					processor.mu.Unlock()
				}
			}
		}()

		return processor
	}

	// 創建處理器
	processor := createProcessor(200 * time.Millisecond) // 每 200ms 觸發一次

	// 提交任務
	startTime := time.Now()
	for i := 0; i < 15; i++ {
		task := Task{
			ID:   i,
			Data: fmt.Sprintf("任務數據 %d", i),
		}
		processor.taskChan <- task
		fmt.Printf("  [時間] 📥 提交任務 #%d\n", i)
		time.Sleep(30 * time.Millisecond) // 模擬任務提交間隔
	}

	// 等待所有任務處理完成
	time.Sleep(500 * time.Millisecond) // 等待最後一批處理完成
	processor.cancel()
	processor.wg.Wait()

	elapsed := time.Since(startTime)
	processor.processedMu.Lock()
	processed := processor.processed
	processor.processedMu.Unlock()

	fmt.Printf("  [時間] ✅ 總耗時: %v，處理了 %d 個任務（定時批量處理，每 200ms 觸發）\n", elapsed, processed)
}

// ==================== 場景 4: 混合觸發批量處理 ====================

// batchProcessingHybrid 混合觸發（數量 + 時間）批量處理
func batchProcessingHybrid() {
	// 任務類型
	type Task struct {
		ID   int
		Data string
	}

	// 批量處理器
	type BatchProcessor struct {
		taskChan    chan Task
		batchSize   int
		batch       []Task
		mu          sync.Mutex
		wg          sync.WaitGroup
		ctx         context.Context
		cancel      context.CancelFunc
		ticker      *time.Ticker
		interval    time.Duration
		processed   int
		processedMu sync.Mutex
	}

	// 處理批量任務（內部函數，不獲取鎖）
	processBatchInternal := func(processor *BatchProcessor) {
		if len(processor.batch) == 0 {
			return
		}

		// 複製當前批次
		batch := make([]Task, len(processor.batch))
		copy(batch, processor.batch)
		processor.batch = processor.batch[:0] // 清空批次

		// 處理批量任務
		fmt.Printf("  [混合] 🔄 開始處理批量任務（共 %d 個）\n", len(batch))
		for _, task := range batch {
			fmt.Printf("  [混合]   處理任務 #%d: %s\n", task.ID, task.Data)
			time.Sleep(10 * time.Millisecond) // 模擬處理耗時
		}

		processor.processedMu.Lock()
		processor.processed += len(batch)
		processor.processedMu.Unlock()

		fmt.Printf("  [混合] ✅ 批量處理完成（共 %d 個任務）\n", len(batch))
	}

	// 創建批量處理器
	createProcessor := func(batchSize int, interval time.Duration) *BatchProcessor {
		ctx, cancel := context.WithCancel(context.Background())
		processor := &BatchProcessor{
			taskChan:  make(chan Task, 100),
			batchSize: batchSize,
			batch:     make([]Task, 0, batchSize),
			ctx:       ctx,
			cancel:    cancel,
			ticker:    time.NewTicker(interval),
			interval:  interval,
		}

		// 啟動批量處理 goroutine
		processor.wg.Add(1)
		go func() {
			defer processor.wg.Done()
			defer processor.ticker.Stop()

			for {
				select {
				case <-processor.ctx.Done():
					// 處理剩餘任務
					processor.mu.Lock()
					if len(processor.batch) > 0 {
						processBatchInternal(processor)
					}
					processor.mu.Unlock()
					return

				case <-processor.ticker.C:
					// 定時觸發批量處理
					processor.mu.Lock()
					if len(processor.batch) > 0 {
						fmt.Printf("  [混合] ⏰ 定時觸發批量處理\n")
						processBatchInternal(processor)
					}
					processor.mu.Unlock()

				case task := <-processor.taskChan:
					// 接收任務
					processor.mu.Lock()
					processor.batch = append(processor.batch, task)
					currentBatchSize := len(processor.batch)
					processor.mu.Unlock()

					// 達到批量大小，立即觸發處理
					if currentBatchSize >= processor.batchSize {
						fmt.Printf("  [混合] 📦 達到批量大小（%d），立即觸發處理\n", processor.batchSize)
						processor.mu.Lock()
						processBatchInternal(processor)
						processor.mu.Unlock()
					}
				}
			}
		}()

		return processor
	}

	// 創建處理器
	processor := createProcessor(5, 300*time.Millisecond) // 每 5 個或每 300ms 觸發

	// 提交任務（模擬不同提交速度）
	startTime := time.Now()
	fmt.Println("\n  [混合] 階段 1: 快速提交任務（觸發數量限制）")
	for i := 0; i < 6; i++ {
		task := Task{
			ID:   i,
			Data: fmt.Sprintf("任務數據 %d", i),
		}
		processor.taskChan <- task
		fmt.Printf("  [混合] 📥 提交任務 #%d\n", i)
		time.Sleep(20 * time.Millisecond)
	}

	fmt.Println("\n  [混合] 階段 2: 慢速提交任務（觸發時間限制）")
	for i := 6; i < 12; i++ {
		task := Task{
			ID:   i,
			Data: fmt.Sprintf("任務數據 %d", i),
		}
		processor.taskChan <- task
		fmt.Printf("  [混合] 📥 提交任務 #%d\n", i)
		time.Sleep(80 * time.Millisecond)
	}

	// 等待所有任務處理完成
	time.Sleep(500 * time.Millisecond) // 等待最後一批處理完成
	processor.cancel()
	processor.wg.Wait()

	elapsed := time.Since(startTime)
	processor.processedMu.Lock()
	processed := processor.processed
	processor.processedMu.Unlock()

	fmt.Printf("\n  [混合] ✅ 總耗時: %v，處理了 %d 個任務（混合觸發：每 5 個或每 300ms）\n", elapsed, processed)
}
