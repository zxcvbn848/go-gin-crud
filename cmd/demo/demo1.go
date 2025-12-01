package main

import (
	"fmt"
	"time"
)

// Demo1: 基本 Goroutine + Channel 使用
func demo1() {
	fmt.Println("\n=== Demo 1: 基本 Goroutine + Channel ===")

	// 創建 channel
	ch := make(chan int, 3)

	// 啟動多個 goroutine 發送數據
	for i := 0; i < 3; i++ {
		go func(id int) {
			time.Sleep(time.Duration(id) * 100 * time.Millisecond)
			ch <- id
			fmt.Printf("Goroutine %d 發送了數據\n", id)
		}(i)
	}

	// 接收數據
	for i := 0; i < 3; i++ {
		value := <-ch
		fmt.Printf("接收到數據: %d\n", value)
	}

	fmt.Println("Demo 1 完成")
}
