package main

import (
	"context"
	"fmt"
)

// contextKey 定義 context 的 key 類型，避免使用內建類型作為 key
type contextKey string

const (
	userIDKey    contextKey = "userID"
	requestIDKey contextKey = "requestID"
)

// Demo4: 使用 Context 傳遞值
func demo4() {
	fmt.Println("\n=== Demo 4: Context 傳遞值 ===")

	// 創建帶值的 context
	ctx := context.WithValue(context.Background(), userIDKey, "12345")
	ctx = context.WithValue(ctx, requestIDKey, "req-abc-123")

	// 在 goroutine 中使用 context 的值
	ch := make(chan string, 1)
	go func(ctx context.Context) {
		userID := ctx.Value(userIDKey).(string)
		requestID := ctx.Value(requestIDKey).(string)
		ch <- fmt.Sprintf("UserID: %s, RequestID: %s", userID, requestID)
	}(ctx)

	result := <-ch
	fmt.Println(result)

	fmt.Println("Demo 4 完成")
}

