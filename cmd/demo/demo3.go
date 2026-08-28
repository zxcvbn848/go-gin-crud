package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Demo3: Worker Pool 模式（結合 Goroutine + Channel + Context）
func demo3() {
	fmt.Println("\n=== Demo 3: Worker Pool 模式 ===")

	// 創建帶超時的 context（整個 demo 最多運行 10 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 創建 channels
	taskChan := make(chan Task, 10)
	resultChan := make(chan Result, 10)

	// 創建並啟動工作池（3 個 worker）
	pool := NewWorkerPool(3, taskChan, resultChan)
	pool.Start(ctx)

	// 啟動結果收集器
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		successCount := 0
		timeoutCount := 0

		for result := range resultChan {
			if result.Error == nil {
				successCount++
			} else {
				timeoutCount++
			}
		}

		fmt.Printf("\n=== 統計結果 ===\n")
		fmt.Printf("成功: %d 個任務\n", successCount)
		fmt.Printf("超時: %d 個任務\n", timeoutCount)
	}()

	// 發送任務
	go func() {
		defer close(taskChan)
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		for i := 1; i <= 10; i++ {
			// 隨機生成任務執行時間（1-4 秒）
			duration := time.Duration(r.Intn(3000)+1000) * time.Millisecond
			task := Task{
				ID:       i,
				Duration: duration,
			}

			select {
			case taskChan <- task:
				fmt.Printf("已發送任務 %d (預計耗時: %v)\n", task.ID, task.Duration)
			case <-ctx.Done():
				fmt.Println("Context 已取消，停止發送任務")
				return
			}

			// 稍微延遲，模擬任務到達的間隔
			time.Sleep(200 * time.Millisecond)
		}
	}()

	// 等待所有 worker 完成
	pool.Wait()
	wg.Wait()

	fmt.Println("Demo 3 完成")
}
