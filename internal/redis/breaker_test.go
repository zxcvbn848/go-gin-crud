package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	rd "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// t0 固定起點。now 由參數傳入，所以整份測試不需要真的等 cooldown。
var t0 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// fail 連續回報 n 次失敗
func fail(b *breaker, now time.Time, n int) {
	for i := 0; i < n; i++ {
		b.record(now, true)
	}
}

// TestBreakerClosedAllows 初始狀態放行
func TestBreakerClosedAllows(t *testing.T) {
	b := &breaker{}
	assert.True(t, b.allow(t0))
}

// TestBreakerSuccessResetsFailures 失敗未達門檻時，一次成功就歸零
func TestBreakerSuccessResetsFailures(t *testing.T) {
	b := &breaker{}

	fail(b, t0, breakerThreshold-1)
	assert.True(t, b.allow(t0), "未達門檻不該熔斷")

	b.record(t0, false) // 成功
	fail(b, t0, breakerThreshold-1)
	assert.True(t, b.allow(t0), "計數應已被成功歸零，不該累加後熔斷")
}

// TestBreakerOpensAtThreshold 連續失敗達門檻即熔斷，cooldown 內一律擋
func TestBreakerOpensAtThreshold(t *testing.T) {
	b := &breaker{}
	fail(b, t0, breakerThreshold)

	assert.False(t, b.allow(t0), "達門檻應熔斷")
	assert.False(t, b.allow(t0.Add(breakerCooldown-time.Nanosecond)), "cooldown 內應持續擋")
}

// TestBreakerHalfOpenAllowsOneProbe cooldown 過後只放行一個探測請求
func TestBreakerHalfOpenAllowsOneProbe(t *testing.T) {
	b := &breaker{}
	fail(b, t0, breakerThreshold)

	after := t0.Add(breakerCooldown)
	assert.True(t, b.allow(after), "cooldown 過後應放行探測")
	assert.False(t, b.allow(after), "同時只該有一個探測在飛")
	assert.False(t, b.allow(after), "後續仍應擋住")
}

// TestBreakerProbeSuccessCloses 探測成功 → 完全復原
func TestBreakerProbeSuccessCloses(t *testing.T) {
	b := &breaker{}
	fail(b, t0, breakerThreshold)

	after := t0.Add(breakerCooldown)
	assert.True(t, b.allow(after))

	b.record(after, false) // 探測成功
	assert.True(t, b.allow(after), "應回到 closed")
	assert.True(t, b.allow(after), "closed 不該有探測名額限制")
}

// TestBreakerProbeFailureReopens 探測失敗 → 立刻回到 open 並重新計時
//
// 這是最容易寫錯的一格：若只看 failures >= threshold，探測失敗時 failures
// 已被歸零過或未達門檻，就會誤放行，變成每次 cooldown 都放一個請求進去慢慢等。
func TestBreakerProbeFailureReopens(t *testing.T) {
	b := &breaker{}
	fail(b, t0, breakerThreshold)

	after := t0.Add(breakerCooldown)
	assert.True(t, b.allow(after))

	b.record(after, true) // 探測失敗
	assert.False(t, b.allow(after), "探測失敗應立刻回到 open")
	assert.False(t, b.allow(after.Add(breakerCooldown-time.Nanosecond)), "cooldown 應從探測失敗當下重新計時")
	assert.True(t, b.allow(after.Add(breakerCooldown)), "新的 cooldown 過後才再放行")
}

// TestIsBreakerFailure 什麼該算 Redis 的失敗
func TestIsBreakerFailure(t *testing.T) {
	ctx := context.Background()

	assert.False(t, isBreakerFailure(ctx, nil), "成功")
	assert.False(t, isBreakerFailure(ctx, rd.Nil), "rd.Nil 是快取未命中，Redis 是健康的")
	assert.True(t, isBreakerFailure(ctx, errors.New("connection refused")), "連線失敗")

	// 呼叫方自己取消（例如 HTTP client 斷線）不該推開熔斷器
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	assert.False(t, isBreakerFailure(cancelled, context.Canceled), "呼叫方取消不是 Redis 的錯")
}
