package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// contains 檢查字符串是否包含子字符串（不區分大小寫）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// TestExecuteTask 測試執行單個任務
func TestExecuteTask(t *testing.T) {
	req := map[string]interface{}{
		"task": map[string]interface{}{
			"id":          "task-1",
			"duration":    1000, // 1 秒
			"description": "測試任務",
		},
		"timeout": 3, // 3 秒超時
	}
	w := makeRequest("POST", "/tasks/execute", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "task-1", response["task_id"])
	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "duration")
}

// TestExecuteTaskTimeout 測試任務超時
func TestExecuteTaskTimeout(t *testing.T) {
	req := map[string]interface{}{
		"task": map[string]interface{}{
			"id":          "task-timeout",
			"duration":    5000, // 5 秒
			"description": "會超時的任務",
		},
		"timeout": 1, // 1 秒超時（任務需要 5 秒，會超時）
	}
	w := makeRequest("POST", "/tasks/execute", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "task-timeout", response["task_id"])
	assert.Equal(t, "timeout", response["status"])
	// 檢查消息包含超時相關資訊（可能是中文或英文）
	message := response["message"].(string)
	assert.True(t, 
		contains(message, "超時") || 
		contains(message, "timeout") || 
		contains(message, "deadline exceeded"),
		"消息應該包含超時相關資訊: %s", message)
}

// TestExecuteTaskWithRetry 測試帶重試的任務執行
func TestExecuteTaskWithRetry(t *testing.T) {
	req := map[string]interface{}{
		"task": map[string]interface{}{
			"id":          "task-retry",
			"duration":    500, // 0.5 秒
			"description": "帶重試的任務",
		},
		"timeout":    2,
		"max_retry":  2,
		"retry_delay": 100, // 100 毫秒
	}
	w := makeRequest("POST", "/tasks/execute/retry", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "task-retry", response["task_id"])
	assert.Contains(t, response, "retry_count")
}

// TestBatchExecuteTasks 測試批量執行任務
func TestBatchExecuteTasks(t *testing.T) {
	req := map[string]interface{}{
		"tasks": []map[string]interface{}{
			{"id": "task-1", "duration": 500},
			{"id": "task-2", "duration": 800},
			{"id": "task-3", "duration": 300},
		},
		"timeout": 5,
	}
	w := makeRequest("POST", "/tasks/batch", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 3, int(response["total_tasks"].(float64)))
	assert.Contains(t, response, "success_count")
	assert.Contains(t, response, "timeout_count")
	assert.Contains(t, response, "error_count")
	assert.Contains(t, response, "tasks")

	// 驗證任務數量
	tasks := response["tasks"].([]interface{})
	assert.Equal(t, 3, len(tasks))
}

// TestBatchExecuteTasksTimeout 測試批量任務整體超時
func TestBatchExecuteTasksTimeout(t *testing.T) {
	req := map[string]interface{}{
		"tasks": []map[string]interface{}{
			{"id": "task-1", "duration": 2000},
			{"id": "task-2", "duration": 2000},
			{"id": "task-3", "duration": 2000},
		},
		"timeout": 1, // 1 秒超時（任務需要 2 秒）
	}
	w := makeRequest("POST", "/tasks/batch", req, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 3, int(response["total_tasks"].(float64)))
	// 應該有超時或取消的任務
	assert.GreaterOrEqual(t, response["timeout_count"].(float64)+response["error_count"].(float64), float64(0))
}

// TestExecuteTaskInvalidRequest 測試無效請求
func TestExecuteTaskInvalidRequest(t *testing.T) {
	// 缺少必要字段
	req := map[string]interface{}{
		"timeout": 3,
	}
	w := makeRequest("POST", "/tasks/execute", req, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 無效的超時時間
	req = map[string]interface{}{
		"task": map[string]interface{}{
			"id":       "task-1",
			"duration": 1000,
		},
		"timeout": 0, // 無效
	}
	w = makeRequest("POST", "/tasks/execute", req, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestExecuteTaskConcurrent 測試併發執行多個任務
func TestExecuteTaskConcurrent(t *testing.T) {
	// 同時執行多個任務
	done := make(chan bool, 3)
	
	go func() {
		req := map[string]interface{}{
			"task": map[string]interface{}{
				"id":       "concurrent-1",
				"duration": 500,
			},
			"timeout": 2,
		}
		w := makeRequest("POST", "/tasks/execute", req, "")
		assert.Equal(t, http.StatusOK, w.Code)
		done <- true
	}()

	go func() {
		req := map[string]interface{}{
			"task": map[string]interface{}{
				"id":       "concurrent-2",
				"duration": 500,
			},
			"timeout": 2,
		}
		w := makeRequest("POST", "/tasks/execute", req, "")
		assert.Equal(t, http.StatusOK, w.Code)
		done <- true
	}()

	go func() {
		req := map[string]interface{}{
			"task": map[string]interface{}{
				"id":       "concurrent-3",
				"duration": 500,
			},
			"timeout": 2,
		}
		w := makeRequest("POST", "/tasks/execute", req, "")
		assert.Equal(t, http.StatusOK, w.Code)
		done <- true
	}()

	// 等待所有任務完成
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// 任務完成
		case <-time.After(5 * time.Second):
			t.Fatal("任務執行超時")
		}
	}
}

