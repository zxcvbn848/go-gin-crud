package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newRouter 建一個掛好逾時中介層的路由，handler 由測試決定
func newRouter(d time.Duration, path string, handler gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(TimeoutMiddleware(d))
	r.GET(path, handler)
	return r
}

// do 發一個 GET 請求並回傳 recorder
func do(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

// TestTimeoutPassesThrough 沒逾時的請求不受影響
func TestTimeoutPassesThrough(t *testing.T) {
	r := newRouter(time.Second, "/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := do(r, "/ok")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "true")
}

// TestTimeoutInjectsDeadline handler 收到的 ctx 必須帶 deadline，
// 這樣下游的 GORM / go-redis 才會提早失敗
func TestTimeoutInjectsDeadline(t *testing.T) {
	var hasDeadline bool
	var remaining time.Duration

	r := newRouter(500*time.Millisecond, "/ctx", func(c *gin.Context) {
		dl, ok := c.Request.Context().Deadline()
		hasDeadline = ok
		remaining = time.Until(dl)
		c.Status(http.StatusOK)
	})

	do(r, "/ctx")
	assert.True(t, hasDeadline, "ctx 應帶 deadline")
	assert.Positive(t, remaining)
	assert.LessOrEqual(t, remaining, 500*time.Millisecond)
}

// TestTimeoutReturns504 handler 等到逾時且什麼都沒寫 → 補 504
func TestTimeoutReturns504(t *testing.T) {
	r := newRouter(20*time.Millisecond, "/slow", func(c *gin.Context) {
		<-c.Request.Context().Done() // 模擬吃 ctx 的下游呼叫被 deadline 打斷
	})

	w := do(r, "/slow")
	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Contains(t, w.Body.String(), "逾時")
}

// TestTimeoutDoesNotOverwriteWrittenResponse handler 已經寫過回應時不覆寫。
//
// 覆寫會造成 superfluous WriteHeader，而且會把 handler 已經回給呼叫方的
// 內容硬改成 504，實際上那個回應已經送出去了。
func TestTimeoutDoesNotOverwriteWrittenResponse(t *testing.T) {
	r := newRouter(20*time.Millisecond, "/partial", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"done": true}) // 先寫
		<-c.Request.Context().Done()                    // 再逾時
	})

	w := do(r, "/partial")
	assert.Equal(t, http.StatusCreated, w.Code, "已寫過的狀態碼不該被改成 504")
	assert.Contains(t, w.Body.String(), "done")
}

// TestTimeoutSkipsStreamingPaths 長連線路徑不套用逾時，ctx 不該帶 deadline
func TestTimeoutSkipsStreamingPaths(t *testing.T) {
	for _, path := range []string{"/stream/sse", "/socket.io/", "/tasks/execute"} {
		var hasDeadline bool

		r := gin.New()
		r.Use(TimeoutMiddleware(time.Millisecond))
		r.GET(path, func(c *gin.Context) {
			_, hasDeadline = c.Request.Context().Deadline()
			c.Status(http.StatusOK)
		})

		w := do(r, path)
		assert.Equal(t, http.StatusOK, w.Code, path)
		assert.False(t, hasDeadline, "%s 應豁免逾時，ctx 不該被換掉", path)
	}
}

// TestSkipTimeout 前綴比對
func TestSkipTimeout(t *testing.T) {
	assert.True(t, skipTimeout("/stream/sse"))
	assert.True(t, skipTimeout("/socket.io/"))
	assert.True(t, skipTimeout("/tasks/execute/retry"))

	assert.False(t, skipTimeout("/users"), "一般 API 不該被豁免")
	assert.False(t, skipTimeout("/stream"), "少了尾斜線就不是 stream 群組，不該誤判")
	assert.False(t, skipTimeout("/books/stream/"), "前綴比對，不是子字串比對")
}
