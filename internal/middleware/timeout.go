package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// timeoutSkipPrefixes 不套用請求逾時的路徑前綴：
//
//	/stream/     SSE 與 chunked 回應本質上是長連線，套逾時會直接切斷
//	/socket.io/  WebSocket 同上
//	/tasks/      任務層已用 context.WithTimeout 自行控制，且重試鏈的總長
//	             由請求參數決定（maxRetry × backoff），固定逾時會誤砍
var timeoutSkipPrefixes = []string{"/stream/", "/socket.io/", "/tasks/"}

// TimeoutMiddleware 為每個請求加上逾時。
//
// 做法是把帶 deadline 的 context 換進 c.Request —— GORM 與 go-redis 都會
// respect 傳入的 ctx，因此真正耗時的 DB / 快取呼叫會提早失敗，請求不會
// 一直佔著 goroutine 與連線。這也是 Redis 熔斷器降級路徑的另一道保險：
// 熔斷器讓失敗變快，這裡讓「沒有熔斷器保護的慢」有上限。
//
// ponytail: 只注入 context，不另開 goroutine 攔截回應。因此純 CPU 迴圈或
// 不吃 ctx 的第三方呼叫仍然不會被硬性切斷。要硬上限得把 handler 丟進
// goroutine 並緩衝 ResponseWriter，那會引入雙寫競態與記憶體緩衝 ——
// 代價大於這裡要解的問題（幾乎所有等待都花在 DB / 快取 I/O 上）。
func TimeoutMiddleware(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipTimeout(c.Request.URL.Path) {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// handler 已經寫過回應就不能再改狀態碼（會是 superfluous WriteHeader），
		// 只有它自己什麼都沒寫時才補 504
		if errors.Is(ctx.Err(), context.DeadlineExceeded) && !c.Writer.Written() {
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error": "請求處理逾時",
			})
		}
	}
}

// skipTimeout 判斷這個路徑是否豁免逾時
func skipTimeout(path string) bool {
	for _, prefix := range timeoutSkipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
