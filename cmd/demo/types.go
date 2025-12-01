package main

import "time"

// Task 代表一個需要處理的任務
type Task struct {
	ID       int
	Duration time.Duration // 任務執行時間
}

// Result 代表任務處理結果
type Result struct {
	TaskID   int
	Status   string
	Duration time.Duration
	Error    error
}

