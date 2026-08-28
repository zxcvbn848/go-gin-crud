package redis

import (
	"context"
	"errors"
	"sync"
	"time"

	rd "github.com/redis/go-redis/v9"
)

// ErrBreakerOpen 熔斷期間直接回傳，不會真的去打 Redis。
//
// 快取層對它的處理與其他 Redis 錯誤相同（不是 rd.Nil，所以會往上回報 error），
// 而 service 層本來就會在快取出錯時回 DB 查，因此降級行為自動成立。
var ErrBreakerOpen = errors.New("redis 熔斷中，跳過本次呼叫")

const (
	// breakerThreshold 連續失敗幾次後熔斷
	breakerThreshold = 5
	// breakerCooldown 熔斷後多久放行一個探測請求
	breakerCooldown = 10 * time.Second
)

// breaker 三態熔斷器：closed → open → half-open → (closed | open)
//
// 存在的理由不是保護 Redis，而是保護延遲：Redis 掛掉時 ReadTimeout 是 3 秒，
// 每個請求都要先付這 3 秒才會降級去查 DB。熔斷後直接跳過，延遲回到正常。
//
// ponytail: 以「連續失敗次數」判斷，不是滑動視窗錯誤率。低流量下夠用，
// 但混合流量（大量成功穿插少量失敗）永遠不會熔斷；真要那樣得換成
// 時間視窗內的失敗比例。
type breaker struct {
	mu       sync.Mutex
	failures int
	openedAt time.Time // 零值代表未熔斷（closed）
	probing  bool      // half-open：已放行一個探測，其餘照擋
}

// allow 回報這次呼叫是否放行。now 由呼叫方傳入，測試才不必真的等 cooldown。
func (b *breaker) allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.openedAt.IsZero() {
		return true // closed
	}
	if now.Sub(b.openedAt) < breakerCooldown {
		return false // open
	}
	if b.probing {
		return false // half-open，已經有人在探測了
	}

	b.probing = true
	return true
}

// record 回報這次呼叫的結果，推進狀態機
func (b *breaker) record(now time.Time, failed bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !failed {
		// 成功即完全復原（含 half-open 探測成功）
		b.failures = 0
		b.openedAt = time.Time{}
		b.probing = false
		return
	}

	b.failures++
	// half-open 的探測失敗 → 立刻回到 open，重新計時
	if b.probing || b.failures >= breakerThreshold {
		b.openedAt = now
		b.probing = false
		// 計數歸零，否則它永遠 >= threshold，上面的 b.probing 條件會變成死碼
		b.failures = 0
	}
}

// breakerHook 用 go-redis 官方的 Hook 攔截所有指令，
// 因此快取層與 service 層完全不需要改動。
type breakerHook struct {
	b *breaker
}

func (h breakerHook) DialHook(next rd.DialHook) rd.DialHook { return next }

func (h breakerHook) ProcessPipelineHook(next rd.ProcessPipelineHook) rd.ProcessPipelineHook {
	return next
}

func (h breakerHook) ProcessHook(next rd.ProcessHook) rd.ProcessHook {
	return func(ctx context.Context, cmd rd.Cmder) error {
		if !h.b.allow(time.Now()) {
			// 呼叫方是讀 cmd.Err()（例如 .Bytes()），所以錯誤要同時掛在 cmd 上
			cmd.SetErr(ErrBreakerOpen)
			return ErrBreakerOpen
		}

		err := next(ctx, cmd)
		h.b.record(time.Now(), isBreakerFailure(ctx, err))
		return err
	}
}

// isBreakerFailure 判斷這個錯誤該不該算成 Redis 的失敗
func isBreakerFailure(ctx context.Context, err error) bool {
	if err == nil || errors.Is(err, rd.Nil) {
		return false // rd.Nil 是快取未命中，代表 Redis 活得很好
	}
	// 呼叫方自己取消或逾時（例如 HTTP client 斷線）不是 Redis 的錯，
	// 不該讓它把熔斷器推開
	return ctx.Err() == nil
}
