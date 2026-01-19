package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Demo9: 並行查詢聚合（Parallel Query Aggregation）
// 詳細說明請參考 cmd/concurrency/README.md
func demo9() {
	fmt.Println("\n=== Demo 9: 並行查詢聚合 ===")

	fmt.Println("\n--- 場景 1: 串行查詢（對比）---")
	serialQuery()

	fmt.Println("\n--- 場景 2: 並行查詢（基本實現）---")
	parallelQuery()

	fmt.Println("\n--- 場景 3: 並行查詢 + 超時控制 ---")
	parallelQueryWithTimeout()

	fmt.Println("\n--- 場景 4: 並行查詢 + 部分失敗處理 ---")
	parallelQueryWithPartialFailure()

	fmt.Println("\nDemo 9 完成")
}

// ==================== 場景 1: 串行查詢 ====================

// serialQuery 演示串行查詢的問題
func serialQuery() {
	// 模擬多個數據源
	dataSources := []string{
		"API-用戶服務",
		"API-訂單服務",
		"API-商品服務",
		"API-支付服務",
		"API-物流服務",
	}

	// 模擬查詢函數
	queryDataSource := func(source string) (map[string]interface{}, error) {
		fmt.Printf("  [串行] 🔍 查詢數據源: %s\n", source)
		time.Sleep(200 * time.Millisecond) // 模擬網絡延遲
		return map[string]interface{}{
			"source": source,
			"data":   fmt.Sprintf("來自 %s 的數據", source),
		}, nil
	}

	// 串行查詢
	startTime := time.Now()
	results := make(map[string]interface{})
	for _, source := range dataSources {
		data, err := queryDataSource(source)
		if err != nil {
			fmt.Printf("  [串行] ❌ 查詢失敗: %s - %v\n", source, err)
			continue
		}
		results[source] = data
	}
	elapsed := time.Since(startTime)

	fmt.Printf("  [串行] ✅ 查詢完成，總耗時: %v，獲取 %d 個數據源\n", elapsed, len(results))
	fmt.Printf("  [串行] ⚠️  串行查詢耗時較長，因為每個查詢都要等待前一個完成\n")
}

// ==================== 場景 2: 並行查詢 ====================

// parallelQuery 並行查詢基本實現
func parallelQuery() {
	// 模擬多個數據源
	dataSources := []string{
		"API-用戶服務",
		"API-訂單服務",
		"API-商品服務",
		"API-支付服務",
		"API-物流服務",
	}

	// 查詢結果類型
	type QueryResult struct {
		Source string
		Data   map[string]interface{}
		Error  error
	}

	// 模擬查詢函數
	queryDataSource := func(source string) (map[string]interface{}, error) {
		fmt.Printf("  [並行] 🔍 查詢數據源: %s\n", source)
		time.Sleep(200 * time.Millisecond) // 模擬網絡延遲
		return map[string]interface{}{
			"source": source,
			"data":   fmt.Sprintf("來自 %s 的數據", source),
		}, nil
	}

	// 使用 Channel 收集結果
	resultChan := make(chan QueryResult, len(dataSources))
	var wg sync.WaitGroup

	// 啟動多個 goroutine 並行查詢
	startTime := time.Now()
	for _, source := range dataSources {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			data, err := queryDataSource(src)
			resultChan <- QueryResult{
				Source: src,
				Data:   data,
				Error:  err,
			}
		}(source)
	}

	// 等待所有查詢完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集結果
	results := make(map[string]interface{})
	successCount := 0
	failureCount := 0

	for result := range resultChan {
		if result.Error != nil {
			fmt.Printf("  [並行] ❌ 查詢失敗: %s - %v\n", result.Source, result.Error)
			failureCount++
		} else {
			results[result.Source] = result.Data
			successCount++
		}
	}

	elapsed := time.Since(startTime)

	fmt.Printf("  [並行] ✅ 查詢完成，總耗時: %v，成功: %d，失敗: %d\n", elapsed, successCount, failureCount)
	fmt.Printf("  [並行] 🚀 並行查詢大幅縮短了總耗時（從串行的 %v 縮短到 %v）\n", time.Duration(len(dataSources))*200*time.Millisecond, elapsed)
}

// ==================== 場景 3: 並行查詢 + 超時控制 ====================

// parallelQueryWithTimeout 並行查詢 + 超時控制
func parallelQueryWithTimeout() {
	// 模擬多個數據源（有些可能響應較慢）
	dataSources := []string{
		"API-用戶服務",
		"API-訂單服務",
		"API-商品服務",
		"API-支付服務（慢）", // 模擬慢響應
		"API-物流服務",
	}

	// 查詢結果類型
	type QueryResult struct {
		Source string
		Data   map[string]interface{}
		Error  error
	}

	// 模擬查詢函數（有些數據源可能很慢）
	queryDataSource := func(ctx context.Context, source string) (map[string]interface{}, error) {
		// 模擬慢響應的數據源
		sleepDuration := 200 * time.Millisecond
		if source == "API-支付服務（慢）" {
			sleepDuration = 800 * time.Millisecond // 模擬慢響應
		}

		fmt.Printf("  [超時] 🔍 查詢數據源: %s\n", source)

		// 使用 select 實現超時控制
		done := make(chan struct{})
		var data map[string]interface{}

		go func() {
			time.Sleep(sleepDuration)
			data = map[string]interface{}{
				"source": source,
				"data":   fmt.Sprintf("來自 %s 的數據", source),
			}
			close(done)
		}()

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("查詢超時: %s", ctx.Err())
		case <-done:
			return data, nil
		}
	}

	// 設置超時時間（500ms）
	timeout := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 使用 Channel 收集結果
	resultChan := make(chan QueryResult, len(dataSources))
	var wg sync.WaitGroup

	// 啟動多個 goroutine 並行查詢
	startTime := time.Now()
	for _, source := range dataSources {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			data, err := queryDataSource(ctx, src)
			resultChan <- QueryResult{
				Source: src,
				Data:   data,
				Error:  err,
			}
		}(source)
	}

	// 等待所有查詢完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集結果
	results := make(map[string]interface{})
	successCount := 0
	timeoutCount := 0

	for result := range resultChan {
		if result.Error != nil {
			if errors.Is(result.Error, context.DeadlineExceeded) {
				fmt.Printf("  [超時] ⏰ 查詢超時: %s\n", result.Source)
				timeoutCount++
			} else {
				fmt.Printf("  [超時] ❌ 查詢失敗: %s - %v\n", result.Source, result.Error)
			}
		} else {
			results[result.Source] = result.Data
			successCount++
		}
	}

	elapsed := time.Since(startTime)

	fmt.Printf("  [超時] ✅ 查詢完成，總耗時: %v，成功: %d，超時: %d\n", elapsed, successCount, timeoutCount)
	fmt.Printf("  [超時] 🎯 超時控制確保了查詢不會無限等待，即使部分數據源響應慢\n")
}

// ==================== 場景 4: 並行查詢 + 部分失敗處理 ====================

// parallelQueryWithPartialFailure 並行查詢 + 部分失敗處理
func parallelQueryWithPartialFailure() {
	// 模擬多個數據源（有些可能失敗）
	dataSources := []string{
		"API-用戶服務",
		"API-訂單服務（失敗）", // 模擬失敗
		"API-商品服務",
		"API-支付服務",
		"API-物流服務（失敗）", // 模擬失敗
		"API-評論服務",
	}

	// 查詢結果類型
	type QueryResult struct {
		Source string
		Data   map[string]interface{}
		Error  error
	}

	// 模擬查詢函數（有些數據源可能失敗）
	queryDataSource := func(ctx context.Context, source string) (map[string]interface{}, error) {
		fmt.Printf("  [部分失敗] 🔍 查詢數據源: %s\n", source)

		// 模擬部分數據源失敗
		if source == "API-訂單服務（失敗）" || source == "API-物流服務（失敗）" {
			time.Sleep(100 * time.Millisecond) // 模擬網絡延遲
			return nil, fmt.Errorf("數據源暫時不可用")
		}

		// 正常查詢
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return map[string]interface{}{
				"source": source,
				"data":   fmt.Sprintf("來自 %s 的數據", source),
			}, nil
		}
	}

	// 設置超時時間
	timeout := 1 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 使用 Channel 收集結果
	resultChan := make(chan QueryResult, len(dataSources))
	var wg sync.WaitGroup

	// 啟動多個 goroutine 並行查詢
	startTime := time.Now()
	for _, source := range dataSources {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			data, err := queryDataSource(ctx, src)
			resultChan <- QueryResult{
				Source: src,
				Data:   data,
				Error:  err,
			}
		}(source)
	}

	// 等待所有查詢完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集結果並聚合
	results := make(map[string]interface{})
	successCount := 0
	failureCount := 0
	errors := make([]string, 0)

	for result := range resultChan {
		if result.Error != nil {
			fmt.Printf("  [部分失敗] ❌ 查詢失敗: %s - %v\n", result.Source, result.Error)
			failureCount++
			errors = append(errors, fmt.Sprintf("%s: %v", result.Source, result.Error))
		} else {
			results[result.Source] = result.Data
			successCount++
		}
	}

	elapsed := time.Since(startTime)

	// 聚合結果
	aggregatedResult := map[string]interface{}{
		"success_count": successCount,
		"failure_count": failureCount,
		"total_sources": len(dataSources),
		"data":          results,
		"errors":        errors,
	}

	fmt.Printf("  [部分失敗] ✅ 查詢完成，總耗時: %v\n", elapsed)
	fmt.Printf("  [部分失敗] 📊 成功: %d，失敗: %d，總數: %d\n", successCount, failureCount, len(dataSources))
	fmt.Printf("  [部分失敗] 🎯 聚合結果: 即使部分數據源失敗，也能返回可用數據\n")
	fmt.Printf("  [部分失敗] 📦 聚合數據: %v\n", aggregatedResult)
}

