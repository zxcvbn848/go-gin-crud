# Go 語言常見／特有名詞整理

這份文件整理了 Go 語言中常見、或相對「Go 特有」的名詞，並依照不同主題分表列出，方便快速查詢與複習。

---

## 併發相關（Concurrency）

| 名詞 | 說明（中文） | 英文關鍵字 |
|------|--------------|------------|
| goroutine | Go 內建的輕量級執行單元，比作業系統 thread 更便宜 | goroutine |
| channel | 在 goroutine 之間安全傳遞資料的管道，可同步／非同步 | channel |
| select | 同時等待多個 channel 操作的語法，類似多路復用 | select statement |
| context | 傳遞取消／超時／請求範圍資料的物件 | context.Context |

---

## 語言特性（Language Features）

| 名詞 | 說明（中文） | 英文關鍵字 |
|------|--------------|------------|
| defer | 延遲執行，在函式回傳前才呼叫，多用於釋放資源 | defer |
| iota | 在 const 區塊中自動遞增的列舉輔助常數 | iota |
| embedding | 結構體／介面內嵌，達成類似「繼承／組合」效果 | type embedding |
| init 函式 | 每個檔案可定義 `init()`，在 `main()` 前自動執行 | init function |
| make / new | `make` 建立 slice/map/chan，`new` 配置零值指標 | make / new |

---

## 錯誤處理（Error Handling）

| 名詞 | 說明（中文） | 英文關鍵字 |
|------|--------------|------------|
| panic | 發生嚴重錯誤時中斷執行的機制，類似「崩潰」 | panic |
| recover | 從 panic 中「撈回來」，避免程式直接崩潰 | recover |

---

## 型別與資料結構（Types & Data Structures）

| 名詞 | 說明（中文） | 英文關鍵字 |
|------|--------------|------------|
| zero value | 變數未初始化時的預設值（如 int=0, string=""） | zero value |
| struct | 自訂結構型別，類似其他語言的 class（不含繼承） | struct |
| interface | 行為契約，只定義方法，不關心實作 | interface |
| 空介面 | 可代表任意型別，早期寫成 `interface{}`，現在常用 `any` | empty interface |
| type assertion | 從 interface 取回具體型別的動作，如 `v.(T)` | type assertion |
| method receiver | 綁在型別上的方法接收者，分值接收者／指標接收者 | method receiver |
| slice | 動態長度的序列，底層指向 array，常用的集合型別 | slice |
| map | Key–Value 雜湊表，內建型別 | map |

---

## 模組、工具與編譯相關（Modules, Tools & Compiler）

| 名詞 | 說明（中文） | 英文關鍵字 |
|------|--------------|------------|
| package | 原始碼組織單位，對應資料夾 | package |
| module | 一個 Go 專案單位，由 `go.mod` 描述 | module |
| go mod | Go 的依賴與版本管理機制 | Go modules |
| gofmt | 官方程式碼自動排版工具，統一程式風格 | gofmt |
| go vet | 靜態分析工具，找出可疑的程式碼模式 | go vet |
| race detector | 偵測資料競爭的工具，`go run -race` | race detector |
| escape analysis | 決定變數配置在 stack 還是 heap 的分析 | escape analysis |

