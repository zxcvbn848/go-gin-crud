package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-gin-crud/internal/logger"

	"github.com/gin-gonic/gin"
)

// StreamingController 提供 SSE 與 Chunked 串流 API
type StreamingController struct{}

// NewStreamingController 建立串流控制器
func NewStreamingController() *StreamingController {
	return &StreamingController{}
}

// StreamSSE 以 Server-Sent Events 推送事件（例如即時通知、進度）
// GET /stream/sse?seconds=10 可指定秒數後結束，預設 30 秒
// @Summary SSE 串流
// @Description Server-Sent Events 串流，每秒推送一筆事件，可用 Query seconds 指定秒數（預設 30）
// @Tags streaming
// @Produce text/event-stream
// @Param seconds query int false "串流持續秒數" default(30)
// @Success 200 {string} string "text/event-stream"
// @Router /stream/sse [get]
func (ctrl *StreamingController) StreamSSE(c *gin.Context) {
	seconds := 30
	if s := c.Query("seconds"); s != "" {
		if n, err := parseIntParam(s, 1, 300); err == nil {
			seconds = n
		}
	}

	w := c.Writer
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // 關閉 nginx 等代理的 buffer
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx := c.Request.Context()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for i := 0; i < seconds; i++ {
		select {
		case <-ctx.Done():
			logger.Log.WithField("reason", "client_gone").Debug("SSE client disconnected")
			return
		case <-ticker.C:
			event := map[string]interface{}{
				"seq":    i + 1,
				"time":   time.Now().Format(time.RFC3339),
				"message": fmt.Sprintf("event #%d", i+1),
			}
			payload, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
	// 結束前送一筆 done 事件
	_, _ = fmt.Fprint(w, "event: done\ndata: {\"message\":\"stream ended\"}\n\n")
	flusher.Flush()
}

// StreamChunked 以 Chunked 編碼逐筆輸出（例如 NDJSON、大量資料）
// GET /stream/chunked?count=10 可指定筆數，預設 10
// @Summary Chunked 串流
// @Description 以 chunked 編碼逐筆回傳 JSON 行（NDJSON），可用 Query count 指定筆數（預設 10）
// @Tags streaming
// @Produce application/x-ndjson
// @Param count query int false "輸出筆數" default(10)
// @Success 200 {string} string "application/x-ndjson"
// @Router /stream/chunked [get]
func (ctrl *StreamingController) StreamChunked(c *gin.Context) {
	count := 10
	if s := c.Query("count"); s != "" {
		if n, err := parseIntParam(s, 1, 1000); err == nil {
			count = n
		}
	}

	w := c.Writer
	w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx := c.Request.Context()
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			row := map[string]interface{}{
				"index":   i + 1,
				"value":  fmt.Sprintf("item-%d", i+1),
				"at":     time.Now().Format(time.RFC3339),
			}
			line, _ := json.Marshal(row)
			_, _ = w.Write(line)
			_, _ = w.Write([]byte("\n"))
			flusher.Flush()
			time.Sleep(200 * time.Millisecond) // 模擬逐筆產生
		}
	}
}

// StreamProgress 以 SSE 模擬長時間任務進度（0% -> 100%）
// GET /stream/progress
// @Summary 進度串流（SSE）
// @Description 以 SSE 推送模擬任務進度 0%~100%
// @Tags streaming
// @Produce text/event-stream
// @Success 200 {string} string "text/event-stream"
// @Router /stream/progress [get]
func (ctrl *StreamingController) StreamProgress(c *gin.Context) {
	w := c.Writer
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx := c.Request.Context()
	for p := 0; p <= 100; p += 10 {
		select {
		case <-ctx.Done():
			return
		default:
			payload, _ := json.Marshal(map[string]interface{}{"progress": p, "message": fmt.Sprintf("%d%%", p)})
			_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
			time.Sleep(300 * time.Millisecond)
		}
	}
	_, _ = fmt.Fprint(w, "event: done\ndata: {\"progress\":100,\"message\":\"complete\"}\n\n")
	flusher.Flush()
}

// parseIntParam 解析整數 query 參數，限制在 [min, max]
func parseIntParam(s string, min, max int) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, err
	}
	if n < min {
		n = min
	}
	if n > max {
		n = max
	}
	return n, nil
}
