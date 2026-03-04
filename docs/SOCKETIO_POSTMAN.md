# Socket.IO 用 Postman WebSocket 測試清單

本 API 跑在 `http://localhost:8080` 時，Socket.IO 路徑為 `/socket.io/`。

---

## 一、Postman 設定

1. 新增請求 → 選 **WebSocket Request**（不是 HTTP）。
2. **URL** 填：
   ```text
   ws://localhost:8080/socket.io/?EIO=4&transport=websocket
   ```
3. 若服務在別台或不同 port，改成對應的 `ws://<host>:<port>/socket.io/?EIO=4&transport=websocket`。
4. 點 **Connect**，連線成功後下方會出現 **Messages** 區塊，用來送/收訊息。

---

## 二、封包格式說明

Socket.IO 在 WebSocket 上用的格式為：**`<型別><內容>`**（字串）。

- **型別 `4`**：Socket.IO 層的訊息（我們只會用到這個）。
- **內容**：JSON 陣列字串。
  - 客戶端**發送事件**：`2["事件名稱", 參數1, 參數2, ...]` → 整段合起來是 **`42["事件名稱", 參數1, 參數2, ...]`**。
  - 伺服器**回傳**也會用 `42["事件名稱", 參數...]`。

也就是說，在 Postman 的 **Message** 輸入框裡，你要送的都是以 **`42`** 開頭，後面接一個 JSON 陣列。

---

## 三、連線後伺服器會先送什麼

Connect 成功後，伺服器通常會先送兩筆（順序可能因實作而異）：

| 收到內容 | 意義 |
|----------|------|
| `0{"sid":"xxx",...}` | Engine.IO 連線建立（session id） |
| `40` | 已連上預設 namespace（/） |

看到這些就代表連線正常，可以開始發下面的事件。

---

## 四、要打的完整範例清單（照順序送）

以下每一行都是你在 Postman 的 **Message** 裡要**手動輸入並送出**的內容（一字不差，含括號與逗號）。

### 1. 加入房間 `join_room`

**發送（Postman Message）：**
```text
42["join_room","room-101"]
```

**可預期伺服器回傳：**
```text
42["joined_room","room-101"]
```

其他範例（替換房間名即可）：
```text
42["join_room","lobby"]
42["join_room","chat-123"]
```

---

### 2. 發送聊天訊息 `message`

格式：`[房間名, 訊息文字]`。房間名可填 `lobby` 或你已加入的房間（例如 `room-101`）。

**發送：**
```text
42["message","lobby","你好，這是第一則訊息"]
```

**發送到指定房間：**
```text
42["message","room-101","Hello room-101"]
```

**伺服器行為：** 會對**同房間其他連線**廣播（不包含自己），格式為：
```text
42["message","<發送者的 socket id>","<你打的訊息文字>"]
```
所以若你開兩個 Postman WebSocket 連線，一個發訊息，另一個會收到這筆 `42["message", ...]`。

---

### 3. 離開房間 `leave_room`

**發送：**
```text
42["leave_room","room-101"]
```

**可預期伺服器回傳：**
```text
42["left_room","room-101"]
```

其他範例：
```text
42["leave_room","lobby"]
42["leave_room","chat-123"]
```

---

### 4. 斷線（可選）

- **方式 A**：在 Postman 直接點 **Disconnect**，伺服器會觸發 `disconnect` 邏輯。
- **方式 B**：手動發 Socket.IO 的 disconnect 封包（若套件支援）：
  ```text
  41
  ```
  表示離開預設 namespace，通常會導致連線關閉。

---

## 五、建議測試順序（複製貼上用）

1. **Connect** → 確認收到 `0...` 與 `40`。
2. 送：`42["join_room","room-101"]` → 應收到 `42["joined_room","room-101"]`。
3. 送：`42["message","room-101","測試訊息"]` → 若有另一個同房間連線會收到 `42["message", "<id>", "測試訊息"]`；只有一個連線則不會再收到（因為是廣播給「其他人」）。
4. 送：`42["leave_room","room-101"]` → 應收到 `42["left_room","room-101"]`。
5. **Disconnect** 結束。

---

## 六、常見問題

| 狀況 | 可能原因 |
|------|----------|
| Connect 失敗 | 服務未啟動、URL 或 port 錯誤、path 必須是 `/socket.io/`。 |
| 送 `42[...]` 沒反應 | 先確認有收到 `40`，且字串是 `42` 開頭、後面是**一個** JSON 陣列、括號與逗號正確（英文、無多餘空白）。 |
| 收不到 `message` 回傳 | 伺服器只對「同房間的**其他**連線」廣播，同一隻連線自己不會收到自己發的 `message`。可開第二個 Postman WebSocket 連到同一 URL、加入同一房間再發訊息測試。 |

---

以上即為用 Postman WebSocket 呼叫本專案 Socket.IO 的完整範例清單；所有「要打的」內容都已列出，可直接複製到 Postman 的 Message 使用。
