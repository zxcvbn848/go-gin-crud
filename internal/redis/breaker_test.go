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

// recordN 回報 n 次同樣的結果
func recordN(b *breaker, now time.Time, n int, failed bool) {
	for i := 0; i < n; i++ {
		b.record(now, failed)
	}
}

// trip 用最少的樣本讓熔斷器開啟：剛好 minRequests 次全失敗
func trip(b *breaker, now time.Time) {
	recordN(b, now, breakerMinRequests, true)
}

// TestBreakerClosedAllows 初始狀態放行
func TestBreakerClosedAllows(t *testing.T) {
	b := &breaker{}
	assert.True(t, b.allow(t0))
}

// TestBreakerBelowMinRequests 樣本不足時不判斷，即使失敗率 100%
//
// 沒有這道門檻，視窗內第一個請求失敗就是 100% 失敗率，會立刻熔斷。
func TestBreakerBelowMinRequests(t *testing.T) {
	b := &breaker{}
	recordN(b, t0, breakerMinRequests-1, true)
	assert.True(t, b.allow(t0), "失敗率 100% 但樣本不足，不該熔斷")

	b.record(t0, true) // 補到門檻
	assert.False(t, b.allow(t0), "達到 minRequests 後應熔斷")
}

// TestBreakerBelowFailureRate 樣本夠但失敗率未達門檻
func TestBreakerBelowFailureRate(t *testing.T) {
	b := &breaker{}
	// 25% 失敗率，遠低於 50%
	recordN(b, t0, 5, true)
	recordN(b, t0, 15, false)

	assert.True(t, b.allow(t0), "失敗率 25% 不該熔斷")
}

// TestBreakerMixedTrafficTrips 混合流量：大量成功穿插失敗，達到失敗率就該熔斷
//
// 這是換掉「連續失敗計數」的理由。舊實作在這個情境下永遠不會熔斷，
// 因為每次成功都會把連續計數歸零。
func TestBreakerMixedTrafficTrips(t *testing.T) {
	b := &breaker{}

	// 成功失敗交錯，剛好 50%
	for i := 0; i < breakerMinRequests/2; i++ {
		b.record(t0, false)
		b.record(t0, true)
	}

	assert.False(t, b.allow(t0), "失敗率達 50% 應熔斷，不論失敗是否連續")
}

// TestBreakerWindowExpiry 舊的失敗掉出視窗後不再計入
func TestBreakerWindowExpiry(t *testing.T) {
	b := &breaker{}

	// 先累積到差一筆就熔斷
	recordN(b, t0, breakerMinRequests-1, true)
	assert.True(t, b.allow(t0))

	// 整個視窗過去之後，那些失敗應該全部掉出
	later := t0.Add(breakerWindow + bucketDuration)
	recordN(b, later, breakerMinRequests-1, true)
	assert.True(t, b.allow(later), "舊失敗已掉出視窗，樣本數應重新計算")
}

// TestBreakerOpenBlocksDuringCooldown cooldown 內一律擋
func TestBreakerOpenBlocksDuringCooldown(t *testing.T) {
	b := &breaker{}
	trip(b, t0)

	assert.False(t, b.allow(t0))
	assert.False(t, b.allow(t0.Add(breakerCooldown-time.Nanosecond)), "cooldown 內應持續擋")
}

// TestBreakerHalfOpenAllowsOneProbe cooldown 過後只放行一個探測請求
func TestBreakerHalfOpenAllowsOneProbe(t *testing.T) {
	b := &breaker{}
	trip(b, t0)

	after := t0.Add(breakerCooldown)
	assert.True(t, b.allow(after), "cooldown 過後應放行探測")
	assert.False(t, b.allow(after), "同時只該有一個探測在飛")
	assert.False(t, b.allow(after), "後續仍應擋住")
}

// TestBreakerProbeSuccessCloses 探測成功 → 完全復原，且視窗已清空
func TestBreakerProbeSuccessCloses(t *testing.T) {
	b := &breaker{}
	trip(b, t0)

	after := t0.Add(breakerCooldown)
	assert.True(t, b.allow(after))

	b.record(after, false) // 探測成功
	assert.True(t, b.allow(after), "應回到 closed")
	assert.True(t, b.allow(after), "closed 不該有探測名額限制")

	// 視窗若沒清空，恢復後少數幾次失敗就會把失敗率再次推過門檻
	recordN(b, after, breakerMinRequests-1, true)
	assert.True(t, b.allow(after), "視窗應已清空，樣本數需重新累積")
}

// TestBreakerProbeFailureReopens 探測失敗 → 立刻回到 open 並重新計時
//
// 探測是單一樣本，不進統計視窗；若誤把它丟進失敗率計算，一筆結果會被
// 稀釋掉而無法立即反應。
func TestBreakerProbeFailureReopens(t *testing.T) {
	b := &breaker{}
	trip(b, t0)

	after := t0.Add(breakerCooldown)
	assert.True(t, b.allow(after))

	b.record(after, true) // 探測失敗
	assert.False(t, b.allow(after), "探測失敗應立刻回到 open")
	assert.False(t, b.allow(after.Add(breakerCooldown-time.Nanosecond)), "cooldown 應從探測失敗當下重新計時")
	assert.True(t, b.allow(after.Add(breakerCooldown)), "新的 cooldown 過後才再放行")
}

// TestBucketReuse 環形緩衝：桶過期後就地歸零重用，不會累加到舊統計上
func TestBucketReuse(t *testing.T) {
	b := &breaker{}

	idx := bucketIndex(t0)
	bk := b.currentBucket(idx)
	bk.total = 99
	bk.failures = 99

	// 剛好繞一圈回到同一個槽位，但桶號不同
	nextRound := idx + breakerBuckets
	reused := b.currentBucket(nextRound)

	assert.Equal(t, nextRound, reused.idx, "應該是新的桶號")
	assert.Zero(t, reused.total, "過期桶必須歸零，否則舊統計會被算進新視窗")
	assert.Zero(t, reused.failures)
}

// TestCurrentStateTransitions 直接斷言狀態，而不是從 allow() 的行為反推。
//
// 狀態機的測試斷在狀態上比斷在行為上精準 —— 行為相同的兩個狀態
// （例如 open 與「已有探測在飛的 half-open」都回 false）才分得開。
func TestCurrentStateTransitions(t *testing.T) {
	b := &breaker{}
	assert.Equal(t, stateClosed, b.currentState(t0))

	trip(b, t0)
	assert.Equal(t, stateOpen, b.currentState(t0))
	assert.Equal(t, stateOpen, b.currentState(t0.Add(breakerCooldown-time.Nanosecond)))
	assert.Equal(t, stateHalfOpen, b.currentState(t0.Add(breakerCooldown)), "cooldown 屆滿即進入 half-open，不需要任何程式碼去改狀態")

	after := t0.Add(breakerCooldown)
	assert.True(t, b.allow(after))
	assert.Equal(t, stateHalfOpen, b.currentState(after), "探測在飛時仍是 half-open")

	b.record(after, false)
	assert.Equal(t, stateClosed, b.currentState(after), "探測成功應回到 closed")
}

// TestStateString 狀態名稱，日誌與 metrics 會用到
func TestStateString(t *testing.T) {
	assert.Equal(t, "closed", stateClosed.String())
	assert.Equal(t, "open", stateOpen.String())
	assert.Equal(t, "half-open", stateHalfOpen.String())
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
