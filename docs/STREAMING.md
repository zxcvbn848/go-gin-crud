# Streaming API 說明

本專案提供三種串流端點，皆為 **GET**，不需認證。

---

## 1. Server-Sent Events（SSE）

**路徑：** `GET /stream/sse`  
**用途：** 伺服器持續推送事件（即時通知、即時報表、日誌尾端等）。

**Query 參數：**

| 參數     | 說明           | 預設 |
|----------|----------------|------|
| `seconds` | 串流持續秒數（1–300） | 30   |

**回應：** `Content-Type: text/event-stream`，每秒一筆 JSON 事件，最後一筆為 `event: done`。

**範例：**

```bash
# 預設約 30 秒
curl -N http://localhost:8080/stream/sse

# 只跑 5 秒
curl -N "http://localhost:8080/stream/sse?seconds=5"
```

**事件格式：** 每行 `data: <JSON>`，例如：

```text
data: {"seq":1,"time":"2025-03-04T12:00:00Z","message":"event #1"}

data: {"seq":2,"time":"2025-03-04T12:00:01Z","message":"event #2"}
```

---

## 2. Chunked 串流（NDJSON）

**路徑：** `GET /stream/chunked`  
**用途：** 大量資料或逐筆產生的結果，以「每行一筆 JSON」串流輸出（NDJSON）。

**Query 參數：**

| 參數   | 說明           | 預設 |
|--------|----------------|------|
| `count` | 輸出筆數（1–1000） | 10   |

**回應：** `Content-Type: application/x-ndjson`，每行一筆 JSON。

**範例：**

```bash
curl -N "http://localhost:8080/stream/chunked?count=5"
```

**輸出範例：**

```json
{"index":1,"value":"item-1","at":"2025-03-04T12:00:00Z"}
{"index":2,"value":"item-2","at":"2025-03-04T12:00:00.2Z"}
```

---

## 3. 進度串流（SSE）

**路徑：** `GET /stream/progress`  
**用途：** 以 SSE 模擬長時間任務進度（0% → 100%），可用於前端進度條。

**回應：** `Content-Type: text/event-stream`，每筆 `data` 含 `progress` 與 `message`。

**範例：**

```bash
curl -N http://localhost:8080/stream/progress
```

---

## 前端範例（JavaScript）

### SSE（EventSource）

```javascript
const es = new EventSource("http://localhost:8080/stream/sse?seconds=10");
es.onmessage = (e) => console.log(JSON.parse(e.data));
es.addEventListener("done", () => { es.close(); });
```

### Chunked NDJSON（fetch + ReadableStream）

```javascript
const res = await fetch("http://localhost:8080/stream/chunked?count=10");
const reader = res.body.getReader();
const dec = new TextDecoder();
let buf = "";
for (;;) {
  const { done, value } = await reader.read();
  if (done) break;
  buf += dec.decode(value, { stream: true });
  const lines = buf.split("\n");
  buf = lines.pop() || "";
  for (const line of lines) if (line) console.log(JSON.parse(line));
}
```

---

## 實作重點（後端）

- **SSE：** 設定 `Content-Type: text/event-stream`、`Cache-Control: no-cache`、`Connection: keep-alive`，每寫一筆就 `Flush()`，並用 `c.Request.Context().Done()` 處理客戶端斷線。
- **Chunked：** 設定 `Content-Type: application/x-ndjson`，每寫一行就 `Flush()`，可依 `context.Done()` 提早結束。
- 長時間串流務必尊重 `context` 取消（例如客戶端關閉連線），避免 goroutine 與資源殘留。
