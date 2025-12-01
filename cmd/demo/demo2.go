package main

import (
	"context"
	"fmt"
	"time"
)

// Demo2: Context 取消和超時
func demo2() {
	fmt.Println("\n=== Demo 2: Context 取消和超時 ===")

	// 記錄開始時間
	startTime := time.Now()
	fmt.Printf("開始時間: %s\n", startTime.Format("15:04:05.000"))

	// 創建一個 2 秒後取消的 context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 啟動一個長時間運行的 goroutine
	// 注意：如果 sleep 時間 < 2 秒，會走到 case <-done
	//       如果 sleep 時間 > 2 秒，會走到 case <-ctx.Done()
	done := make(chan bool)
	go func() {
		// 修改這裡的時間來測試不同的情況：
		// - 1 秒：任務會在超時前完成，走到 case <-done
		// - 5 秒：任務會超時，走到 case <-ctx.Done()
		time.Sleep(10 * time.Second) // 改為 1 秒，讓任務在超時前完成
		done <- true
	}()

	select {
	case <-ctx.Done():
		// 記錄取消時間
		cancelTime := time.Now()
		elapsed := cancelTime.Sub(startTime)
		fmt.Printf("取消時間: %s\n", cancelTime.Format("15:04:05.000"))
		fmt.Printf("經過時間: %v\n", elapsed)
		fmt.Printf("Context 已取消: %v\n", ctx.Err())
		fmt.Println("說明: 任務執行時間超過 context 超時時間，所以走到這個分支")
	case <-done:
		finishTime := time.Now()
		elapsed := finishTime.Sub(startTime)
		fmt.Printf("完成時間: %s\n", finishTime.Format("15:04:05.000"))
		fmt.Printf("經過時間: %v\n", elapsed)
		fmt.Println("任務完成")
		fmt.Println("說明: 任務在 context 超時前完成，所以走到這個分支")
	}

	fmt.Println("Demo 2 完成")
}
