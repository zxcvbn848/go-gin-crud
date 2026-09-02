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
	// breakerWindow 統計視窗長度
	breakerWindow = 30 * time.Second
	// breakerBuckets 視窗切成幾個桶（環形緩衝的大小）
	breakerBuckets = 6
	// breakerFailureRate 視窗內失敗率達此比例即熔斷
	breakerFailureRate = 0.5
	// breakerMinRequests 視窗內樣本數不足時不做判斷。
	//
	// 沒有這道門檻，視窗內第一個請求失敗就是 100% 失敗率，會立刻熔斷。
	// 這是低流量下最容易誤觸的來源。
	//
	// ponytail: 這是主要的校準旋鈕。20 筆 / 30 秒約等於 0.67 QPS 才有判斷力，
	// 流量更低的環境要往下調，否則熔斷器實質上永遠不會作用。
	breakerMinRequests = 20
	// breakerCooldown 熔斷後多久放行一個探測請求
	breakerCooldown = 10 * time.Second
)

// bucketDuration 每個桶涵蓋的時間長度
const bucketDuration = breakerWindow / breakerBuckets

// state 熔斷器的三個狀態。
//
// 刻意不存成欄位 —— open → half-open 是「時間走到了」自然發生的轉換，
// 沒有任何程式碼在執行。存起來就需要背景 timer 去改，或是變成第二個
// 真相來源、與 openedAt 有機會不一致。這裡一律由 currentState 推導。
type state int

const (
	stateClosed state = iota
	stateOpen
	stateHalfOpen
)

func (s state) String() string {
	switch s {
	case stateClosed:
		return "closed"
	case stateOpen:
		return "open"
	default:
		return "half-open"
	}
}

// bucket 一個時間格內的統計。
//
// idx 是這個桶代表的絕對時間格編號，用來判定桶是否過期 —— 過期就地歸零，
// 因此不需要背景 goroutine 做清理。
type bucket struct {
	idx      int64
	total    int
	failures int
}

// breaker 三態熔斷器：closed → open → half-open → (closed | open)
//
// 狀態由 openedAt 與 probing 推導，見 currentState。
//
// 存在的理由不是保護 Redis，而是保護延遲：Redis 掛掉時 ReadTimeout 是 3 秒，
// 每個請求都要先付這 3 秒才會降級去查 DB。熔斷後直接跳過，延遲回到正常。
//
// 判斷條件是「視窗內失敗率」而非「連續失敗次數」。後者在混合流量下有漏洞：
// 大量成功穿插少量失敗時，計數永遠被成功歸零，錯誤率再高也不會熔斷。
//
// ponytail: 環形緩衝以桶為單位，所以視窗邊界是階梯式而非平滑滑動 ——
// 最舊的桶會整桶掉出，誤差最多一個桶（5 秒）。要更精確得記錄每筆時間戳，
// 記憶體會隨流量成長，不值得。
type breaker struct {
	mu       sync.Mutex
	buckets  [breakerBuckets]bucket
	openedAt time.Time // 零值代表未熔斷（closed）
	probing  bool      // half-open：已放行一個探測，其餘照擋
}

// bucketIndex 把時間換算成絕對桶號
func bucketIndex(now time.Time) int64 {
	return now.UnixNano() / int64(bucketDuration)
}

// currentBucket 取得 idx 所屬的桶。若該槽位存的是過期的桶，就地歸零後重用。
func (b *breaker) currentBucket(idx int64) *bucket {
	slot := &b.buckets[idx%breakerBuckets]
	if slot.idx != idx {
		*slot = bucket{idx: idx}
	}
	return slot
}

// window 統計視窗內（含 idx 這一桶往前 breakerBuckets 桶）的總數與失敗數
func (b *breaker) window(idx int64) (total, failures int) {
	oldest := idx - breakerBuckets + 1
	for _, bk := range b.buckets {
		if bk.idx >= oldest && bk.idx <= idx {
			total += bk.total
			failures += bk.failures
		}
	}
	return total, failures
}

// resetWindow 清空統計。熔斷與探測成功時都要清 —— 否則恢復後舊的失敗還留在
// 視窗裡，下一兩次失敗就會把失敗率再次推過門檻。
func (b *breaker) resetWindow() {
	b.buckets = [breakerBuckets]bucket{}
}

// currentState 由 openedAt 與 probing 推導出當下的狀態。
//
// 呼叫方必須已持有 b.mu。
func (b *breaker) currentState(now time.Time) state {
	switch {
	case b.openedAt.IsZero():
		return stateClosed
	case now.Sub(b.openedAt) < breakerCooldown:
		return stateOpen
	default:
		return stateHalfOpen
	}
}

// allow 回報這次呼叫是否放行。now 由呼叫方傳入，測試才不必真的等 cooldown。
func (b *breaker) allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.currentState(now) {
	case stateClosed:
		return true
	case stateOpen:
		return false
	default: // stateHalfOpen：只放行一個探測
		if b.probing {
			return false
		}
		b.probing = true
		return true
	}
}

// record 回報這次呼叫的結果，推進狀態機
func (b *breaker) record(now time.Time, failed bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// half-open 的探測用單一樣本決定去留，不進統計視窗 ——
	// 一筆結果不該被稀釋在失敗率裡。
	if b.probing {
		b.probing = false
		if failed {
			b.openedAt = now // 回到 open，cooldown 重新計時
		} else {
			b.openedAt = time.Time{} // 完全復原
			b.resetWindow()
		}
		return
	}

	bk := b.currentBucket(bucketIndex(now))
	bk.total++
	if failed {
		bk.failures++
	}

	// 已經熔斷中（allow 與 record 之間的競態才會走到這裡），不重複判斷
	if b.currentState(now) != stateClosed {
		return
	}

	total, failures := b.window(bucketIndex(now))
	if total < breakerMinRequests {
		return // 樣本不足，失敗率沒有意義
	}
	if float64(failures)/float64(total) >= breakerFailureRate {
		b.openedAt = now
		b.resetWindow()
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
