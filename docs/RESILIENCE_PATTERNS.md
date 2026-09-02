# Resilience Patterns — 流程圖

對應實作：

| 模式 | 位置 |
|---|---|
| Circuit Breaker | `internal/redis/breaker.go` |
| Retry with Backoff + Jitter | `internal/service/task_executor_service.go` |

進度與 Bulkhead 的延後理由見 [`TODO.md`](../TODO.md) 的 Resilience Patterns 段落。

---

## 1. 熔斷器狀態機

三個狀態沒有用 enum，靠 `openedAt` 與 `probing` 兩個欄位隱含表達。

```mermaid
stateDiagram-v2
    [*] --> Closed

    Closed --> Closed: 成功<br/>failures 歸零
    Closed --> Open: 連續失敗達 5 次<br/>openedAt = now

    Open --> Open: cooldown 未滿<br/>全部擋掉
    Open --> HalfOpen: cooldown 10s 屆滿

    HalfOpen --> Closed: 探測成功<br/>全部歸零
    HalfOpen --> Open: 探測失敗<br/>openedAt = now 重新計時

    note right of Closed
        openedAt.IsZero() == true
    end note

    note right of Open
        now - openedAt < 10s
    end note

    note right of HalfOpen
        cooldown 已過
        probing 控制只放一個探測
    end note
```

常數定義於 `breaker.go`：`breakerThreshold = 5`、`breakerCooldown = 10 * time.Second`。

---

## 2. 請求路徑

`breakerHook` 實作 go-redis 的 `Hook` 介面攔截 `ProcessHook`，因此快取層與 service 層不需改動，降級行為自動成立。

```mermaid
sequenceDiagram
    autonumber
    participant S as Service 層
    participant C as 快取層
    participant H as breakerHook
    participant B as breaker
    participant R as Redis
    participant DB as MySQL

    S->>C: 讀快取
    C->>H: Redis 指令
    H->>B: allow(now)

    alt 熔斷中 — 直接跳過
        B-->>H: false
        H->>H: cmd.SetErr(ErrBreakerOpen)
        Note over H: 呼叫方讀 cmd.Err()<br/>只回傳 error 會被忽略
        H-->>C: ErrBreakerOpen
        C-->>S: error
        Note over S,DB: ErrBreakerOpen 不是 rd.Nil<br/>會冒成錯誤，觸發既有降級路徑
        S->>DB: 降級查 DB
        DB-->>S: 資料
    else 放行 — closed 或 half-open 探測
        B-->>H: true
        H->>R: next(ctx, cmd)
        R-->>H: 結果 or error
        H->>B: record(now, isBreakerFailure(ctx, err))
        Note over B: rd.Nil 不算失敗（快取未命中）<br/>呼叫方自己取消也不算
        H-->>C: 結果
        C-->>S: 結果
    end
```

存在的理由不是保護 Redis，而是保護延遲：Redis 掛掉時 `ReadTimeout` 是 3 秒，每個請求都要先付這 3 秒才會降級去查 DB。熔斷後直接跳過，延遲回到正常。

---

## 3. 重試主迴圈

`ExecuteTaskWithRetry`

```mermaid
flowchart TD
    A[ExecuteTaskWithRetry] --> B{"maxRetry < 0 ?"}
    B -->|是| C["maxRetry = 0<br/>防 lastResponse 為 nil"]
    B -->|否| D["attempt = 0"]
    C --> D

    D --> E{"父 ctx 已取消 ?"}
    E -->|是| F["回傳 cancelled"]
    E -->|否| G["ExecuteTask"]

    G --> H{"成功 ?"}
    H -->|是| I["回傳成功<br/>第 attempt+1 次嘗試"]
    H -->|否| J["記下 lastErr / lastResponse"]

    J --> K{"attempt < maxRetry ?"}
    K -->|否| L["回傳失敗<br/>maxRetry+1 次後仍失敗"]
    K -->|是| M["select 等待"]

    M --> N{"哪個先到 ?"}
    N -->|"ctx.Done()"| F
    N -->|"time.After(backoffDelay)"| O["attempt++"]
    O --> E

    style F fill:#4a2020,stroke:#a04040,color:#e8d0d0
    style L fill:#4a2020,stroke:#a04040,color:#e8d0d0
    style I fill:#204a2a,stroke:#40a060,color:#d0e8d8
```

等待用 `select` 同時監聽 `ctx.Done()`，而不是 `time.Sleep` —— 否則重試等待中的取消請求會被無視，白等完整段 backoff。

---

## 4. backoffDelay — 指數退避 + equal jitter

```mermaid
flowchart TD
    A["backoffDelay(base, attempt)"] --> B{"base <= 0 ?"}
    B -->|是| C["回傳 0"]
    B -->|否| D["d = base 左移 attempt 位"]

    D --> E{"d > maxBackoff<br/>或 d <= 0 ?"}
    E -->|"是（含位移溢位）"| F["d = maxBackoff"]
    E -->|否| G{"d < 2 ?"}
    F --> G

    G -->|"是（避免 rand.N(0) panic）"| H["回傳 d"]
    G -->|否| I["回傳 d/2 + rand.N(d/2)<br/>結果落在 [d/2, d)"]

    style I fill:#204a2a,stroke:#40a060,color:#d0e8d8
```

純指數退避會讓多個 client 同時失敗後一起重試，下游剛要復原就被同步打回去（thundering herd）。一半固定一半隨機，把重試攤在時間窗內，尖峰負載大致砍半。

---

## 5. 兩者的分工

| | 處理的故障 | 時間尺度 | 行為 |
|---|---|---|---|
| Retry + backoff + jitter | 網路抖動、GC 暫停、瞬間過載 | 秒 | 撐過去 |
| Circuit Breaker | 服務掛了、部署炸了 | 分鐘 | fail fast + 降級 |

兩者必須成對。**單獨用 retry 是危險的** —— 重試會放大負載，下游掛掉時一個失敗請求變成四個失敗請求。熔斷器就是那個放大的上限。

---

## 邊界守衛

實作中四個非顯而易見的守衛，移除任一個都會出問題：

| 守衛 | 位置 | 移除後的後果 |
|---|---|---|
| 等待期用 `select` 監聽 `ctx.Done()` | `task_executor_service.go` | 取消請求被無視，白等完整段 backoff |
| 位移溢位守衛 `d <= 0` | `backoffDelay` | attempt 大時 `base << attempt` 溢位成負數，出現負延遲 |
| 熔斷時 `failures = 0` 歸零 | `breaker.record` | failures 永遠 >= threshold，`b.probing` 分支變死碼，half-open 探測失敗處理不到 |
| `maxRetry < 0` 修正為 0 | `ExecuteTaskWithRetry` | 迴圈不執行，`lastResponse` 為 nil，回傳時 nil pointer dereference |

## 已知限制

熔斷判斷用「連續失敗次數」而非滑動視窗錯誤率（見 `breaker.go` 的 `ponytail:` 註解）。低流量下夠用，但混合流量（大量成功穿插少量失敗）永遠不會熔斷。要處理那個情境得換成時間視窗內的失敗比例。
