package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Go 併發編程 Demo")
	fmt.Println("展示: Goroutine + Channel + Context.Context")
	fmt.Println("========================================")

	// 解析命令行參數
	args := os.Args[1:]

	if len(args) == 0 {
		// 沒有參數，執行所有 demo
		fmt.Println("\n未指定參數，執行所有 Demo...")
		runAllDemos()
	} else {
		// 有參數，執行指定的 demo
		runSelectedDemos(args)
	}

	fmt.Println("\n========================================")
	fmt.Println("所有 Demo 執行完成！")
	fmt.Println("========================================")
}

// runAllDemos 執行所有 demo
func runAllDemos() {
	demo1()
	demo2()
	demo3()
	demo4()
	demo5()
	demo6()
	demo7()
	demo8()
	demo9()
	demo10()
}

// runSelectedDemos 根據參數執行指定的 demo
func runSelectedDemos(args []string) {
	// 創建 demo 映射表
	demoMap := map[int]func(){
		1:  demo1,
		2:  demo2,
		3:  demo3,
		4:  demo4,
		5:  demo5,
		6:  demo6,
		7:  demo7,
		8:  demo8,
		9:  demo9,
		10: demo10,
	}

	// 記錄已執行的 demo，避免重複執行
	executed := make(map[int]bool)

	for _, arg := range args {
		// 支援多個參數，例如: go run main.go 1 2 3
		// 也支援範圍，例如: go run main.go 1-3
		if strings.Contains(arg, "-") {
			// 處理範圍參數，例如 "1-3"
			parts := strings.Split(arg, "-")
			if len(parts) == 2 {
				start, err1 := strconv.Atoi(parts[0])
				end, err2 := strconv.Atoi(parts[1])
				if err1 == nil && err2 == nil && start > 0 && end > 0 && start <= end {
					for i := start; i <= end; i++ {
						if demo, exists := demoMap[i]; exists && !executed[i] {
							demo()
							executed[i] = true
						}
					}
					continue
				}
			}
		}

		// 處理單個數字參數
		num, err := strconv.Atoi(arg)
		if err != nil {
			fmt.Printf("警告: 無效的參數 '%s'，將被忽略\n", arg)
			continue
		}

		if demo, exists := demoMap[num]; exists {
			if !executed[num] {
				demo()
				executed[num] = true
			}
		} else {
			fmt.Printf("警告: Demo %d 不存在（可用的 Demo: 1-10）\n", num)
		}
	}
}
