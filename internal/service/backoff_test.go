package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBackoffDelayBounds 驗證退避延遲落在 [d/2, d)，且隨 attempt 指數成長
func TestBackoffDelayBounds(t *testing.T) {
	base := 100 * time.Millisecond

	cases := []struct {
		attempt  int
		wantLow  time.Duration // 含
		wantHigh time.Duration // 不含
	}{
		{0, 50 * time.Millisecond, 100 * time.Millisecond},
		{1, 100 * time.Millisecond, 200 * time.Millisecond},
		{2, 200 * time.Millisecond, 400 * time.Millisecond},
		{3, 400 * time.Millisecond, 800 * time.Millisecond},
	}

	// 有 jitter，單次通過不算數，每個 case 跑 100 次
	for _, c := range cases {
		for i := 0; i < 100; i++ {
			got := backoffDelay(base, c.attempt)
			assert.GreaterOrEqual(t, got, c.wantLow, "attempt=%d 延遲過短", c.attempt)
			assert.Less(t, got, c.wantHigh, "attempt=%d 延遲過長", c.attempt)
		}
	}
}

// TestBackoffDelayCapped 大 attempt 不能無限成長，也不能因位移溢位變成 0 或負數
func TestBackoffDelayCapped(t *testing.T) {
	for _, attempt := range []int{10, 30, 63, 64, 100, 1000} {
		got := backoffDelay(time.Second, attempt)
		assert.Greater(t, got, time.Duration(0), "attempt=%d 不該是 0 或負數", attempt)
		assert.Less(t, got, maxBackoff, "attempt=%d 應被 maxBackoff 夾住", attempt)
	}
}

// TestBackoffDelayEdgeCases 邊界：base 為 0 或負數回傳 0，極小 base 不能 panic
func TestBackoffDelayEdgeCases(t *testing.T) {
	assert.Equal(t, time.Duration(0), backoffDelay(0, 3))
	assert.Equal(t, time.Duration(0), backoffDelay(-time.Second, 3))

	// d/2 為 0 的路徑：rand.N(0) 會 panic，必須被擋住
	assert.NotPanics(t, func() { backoffDelay(1, 0) })
	assert.Equal(t, time.Duration(1), backoffDelay(1, 0))
}
