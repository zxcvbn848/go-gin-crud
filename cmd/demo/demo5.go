package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Demo5: 優雅關閉（Graceful Shutdown）
// 詳細說明請參考 README.md
func demo5() {
	fmt.Println("\n=== Demo 5: 優雅關閉 vs 普通關閉 ===")

	fmt.Println("\n--- 場景 1: 普通關閉（直接退出）---")
	normalShutdown()

	fmt.Println("\n--- 場景 2: 優雅關閉（等待完成）---")
	gracefulShutdown()

	fmt.Println("\nDemo 5 完成")
}

// normalShutdown 模擬普通關閉：直接退出，不等待任務完成
func normalShutdown() {
	serviceChan := make(chan string, 10)
	var wg sync.WaitGroup

	// 啟動服務 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			time.Sleep(300 * time.Millisecond)
			serviceChan <- fmt.Sprintf("消息 %d", i+1)
			fmt.Printf("  [普通] 已產生消息 %d\n", i+1)
		}
		fmt.Println("  [普通] 服務 goroutine 完成")
	}()

	// 啟動消息處理器
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			time.Sleep(200 * time.Millisecond)
			select {
			case msg := <-serviceChan:
				fmt.Printf("  [普通] 處理消息: %s\n", msg)
			default:
				fmt.Println("  [普通] 沒有消息可處理")
			}
		}
		fmt.Println("  [普通] 處理器 goroutine 完成")
	}()

	// 模擬主程序在 1 秒後直接退出（不等待）
	time.Sleep(1 * time.Second)
	fmt.Println("  [普通] 主程序直接退出（不等待 goroutine 完成）")
	fmt.Println("  [普通] ⚠️  問題：可能有未處理的消息和未清理的資源")
	// 注意：這裡沒有 wg.Wait()，所以主程序會直接退出
}

// gracefulShutdown 模擬優雅關閉：等待任務完成並清理資源
func gracefulShutdown() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 模擬一個長時間運行的服務
	serviceChan := make(chan string, 10)
	var wg sync.WaitGroup

	// 啟動服務 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("  [優雅] 服務收到關閉信號，正在清理資源...")
				time.Sleep(200 * time.Millisecond) // 模擬清理工作（關閉連接、保存數據等）
				fmt.Println("  [優雅] 服務資源已清理完成")
				return
			case <-ticker.C:
				msg := fmt.Sprintf("服務消息: %s", time.Now().Format("15:04:05"))
				serviceChan <- msg
				fmt.Printf("  [優雅] 已產生: %s\n", msg)
			}
		}
	}()

	// 啟動消息處理器
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				// 處理剩餘消息（確保不丟失數據）
				fmt.Println("  [優雅] 開始處理剩餘消息...")
				remainingCount := 0
				for {
					select {
					case msg := <-serviceChan:
						fmt.Printf("  [優雅] 處理剩餘消息: %s\n", msg)
						remainingCount++
					default:
						fmt.Printf("  [優雅] 已處理 %d 條剩餘消息\n", remainingCount)
						return
					}
				}
			case msg := <-serviceChan:
				fmt.Printf("  [優雅] 處理消息: %s\n", msg)
			}
		}
	}()

	// 運行 3 秒後發送關閉信號
	time.Sleep(3 * time.Second)
	fmt.Println("\n  [優雅] 發送關閉信號...")
	cancel()

	// 等待所有 goroutine 完成（關鍵：優雅關閉的核心）
	fmt.Println("  [優雅] 等待所有 goroutine 完成...")
	wg.Wait()
	fmt.Println("  [優雅] ✅ 所有任務已完成，資源已清理，安全退出")
}
