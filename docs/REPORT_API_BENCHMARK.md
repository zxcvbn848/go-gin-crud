# 報表 API 效能基準

優化前的量測紀錄。目的是讓後續的索引與查詢改寫有可對照的數字，而不是憑感覺說「變快了」。

## 量測環境

| 項目 | 值 |
|---|---|
| 資料庫 | MySQL 容器 `go_gin_crud_mysql`，port 3307 |
| Schema | `goGinCRUD_bench`（獨立於開發用的 `goGinCRUD`） |
| 資料來源 | `go run ./cmd/seed -bulk`（`internal/database/bulk_seeder.go`） |
| 量測方式 | MySQL `SET profiling = 1` + `SHOW PROFILES`，不含 client 往返 |
| 日期 | 2026-09-02 |

重建方式：

```bash
docker exec go_gin_crud_mysql mysql -uroot -p<pw> -e \
  "CREATE DATABASE IF NOT EXISTS goGinCRUD_bench CHARACTER SET utf8mb4; \
   GRANT ALL ON goGinCRUD_bench.* TO 'gogin'@'%'; FLUSH PRIVILEGES;"

DB_NAME=goGinCRUD_bench go run ./cmd/seed -bulk
```

**不要在 `goGinCRUD` 上量。** 開發資料量太小，而且量測會被自己的操作污染。
也不要用 sqlite —— 查詢規劃器、索引選擇與連線池語意都與 MySQL 不同，
`:memory:` 沒有磁碟 IO，40 萬筆全表掃描在裡面是毫秒級，量出來的數字沒有意義。

## 資料量

| 表 | 筆數 |
|---|---|
| users | 50,000 |
| posts | 400,000 |
| products | 50,000 |
| books | 50,000 |

`created_at` 散布在過去 365 天內（`SpreadDays`）。報表按日期分組，資料擠在同一天的話
`GROUP BY` 只會產生一列，量出來的東西跟正式環境完全不同。

## 起始索引狀態

只有 GORM 自動建立的那些，**沒有任何 `created_at` 索引**：

| 表 | 索引 | 欄位 |
|---|---|---|
| users | PRIMARY / idx_users_email_deleted_at / idx_users_deleted_at | id / email / deleted_at |
| posts | PRIMARY / idx_posts_author_id / idx_posts_deleted_at | id / author_id / deleted_at |
| products | PRIMARY / idx_products_deleted_at | id / deleted_at |
| books | PRIMARY / idx_books_deleted_at | id / deleted_at |

## 基準數字

| # | 查詢 | 耗時 | 瓶頸 |
|---|---|---|---|
| 1 | overview — 各表 `COUNT` + `SUM(price*stock)` | **59 ms** | 走 `deleted_at` 索引，本來就不慢 |
| 2 | daily — `GROUP BY DATE(created_at)`，近 90 天 | **260 ms** | 沒有 `created_at` 索引 |
| 3 | authors — `JOIN` + `GROUP BY` + `ORDER BY count` | **535 ms** | 掃 20 萬列後建暫存表排序 |

### 查詢 1 — overview

```sql
SELECT (SELECT COUNT(*) FROM users    WHERE deleted_at IS NULL) u,
       (SELECT COUNT(*) FROM posts    WHERE deleted_at IS NULL) p,
       (SELECT COUNT(*) FROM products WHERE deleted_at IS NULL) pr,
       (SELECT COUNT(*) FROM books    WHERE deleted_at IS NULL) b,
       (SELECT COALESCE(SUM(price*stock),0) FROM products WHERE deleted_at IS NULL) v;
```

59 ms。這支不是優化重點，但適合展示快取 —— 數字變動慢、命中率高。

### 查詢 2 — daily

```sql
SELECT DATE(created_at) d, COUNT(*) n FROM posts
 WHERE deleted_at IS NULL AND created_at >= NOW() - INTERVAL 90 DAY
 GROUP BY DATE(created_at) ORDER BY d;
```

```
type  key                    rows     Extra
ref   idx_posts_deleted_at   198544   Using index condition; Using where;
                                      Using temporary; Using filesort
```

**只能用 `deleted_at` 索引過濾，`created_at` 的範圍條件沒有索引可用**，估算掃 198,544 列。
`Using temporary; Using filesort` 代表分組與排序都在暫存表上做。

值得注意的是 **`GROUP BY DATE(created_at)` 是函數包住欄位**，即使加上 `created_at` 索引，
`GROUP BY` 那段仍然用不到 —— 索引能幫的只有 `WHERE` 的範圍過濾。這一支不是「加個索引就好」。

### 查詢 3 — authors

```sql
SELECT u.id, COUNT(p.id) n FROM posts p JOIN users u ON u.id = p.author_id
 WHERE p.deleted_at IS NULL GROUP BY u.id ORDER BY n DESC LIMIT 10;
```

```
table  type     key                    rows     Extra
p      ref      idx_posts_deleted_at   198544   Using index condition;
                                                Using temporary; Using filesort
u      eq_ref   PRIMARY                1        Using index
```

`users` 側是 `eq_ref` 走主鍵，沒問題。成本全在 `posts` 側：掃 198,544 列，
分組與排序都落在暫存表。`LIMIT 10` 幫不上忙 —— 要先算完所有作者的計數才知道前十名是誰。

**這支是後續 Bulkhead 的保護對象。** 535 ms 意味著幾個這種請求同時進來就會佔住連線池
（`SetMaxOpenConns(100)`，見 `internal/database/database.go`）好幾百毫秒，
而登入等一般 API 也在搶同一池連線。

## API 層基準

端點實作完成後的量測，同一個 bench schema、**Redis 停用**（`REDIS_ADDR=` 空值），
所以是未快取的路徑：

```bash
DB_NAME=goGinCRUD_bench APP_PORT=8090 REDIS_ADDR= go run ./cmd
```

| 端點 | API 耗時 | 對應 SQL 基準 |
|---|---|---|
| `GET /reports/overview` | 51 ms | 59 ms |
| `GET /reports/daily` | 190 ms | 260 ms |
| `GET /reports/authors` | **798 ms** | 535 ms |

`authors` 的 API 比手寫 SQL 慢了約 260 ms，原因不是 HTTP 或 JSON —— 而是
**`GROUP BY` 多了一個 varchar 欄位**：

```sql
-- 手寫基準：只 group by 主鍵
GROUP BY u.id

-- 實際實作：因為 SELECT 了 users.email，ONLY_FULL_GROUP_BY 要求它也進 GROUP BY
GROUP BY posts.author_id, users.email
```

在 varchar 上分組比在整數主鍵上貴。這是實作揭露出來的成本，手寫 SQL 時看不到。

## 待驗證的優化方向

尚未實作，這裡只記方向，實作後回填實測數字：

1. **`posts (created_at)` 或 `(deleted_at, created_at)` 複合索引** —— 幫查詢 2 的範圍過濾。
   要比較單欄與複合的差異，以及 `deleted_at` 放前面是否真的有幫助（區分度極低）
2. **`GROUP BY DATE(created_at)` 改寫** —— 函數包欄位用不到索引。可能的方向是產生欄
   （generated column）加索引，或改由應用層分組
3. **`posts (author_id)` 已存在** —— 查詢 3 的 `GROUP BY` 是否能改走 `author_id` 索引
   避開暫存表，需要實測
4. **`authors` 拆成兩段查詢** —— 先只 `GROUP BY posts.author_id`（整數主鍵，不含
   varchar）取出前 N 名，再用 `WHERE id IN (...)` 一次撈那 N 個 email。
   多一次往返，但省掉在 varchar 上分組的成本。這是 API 層量測才發現的方向
5. **Redis 快取** —— 三支都適合，overview 的 TTL 可以最長

每一項都要回來量，沒量到差異的就不要留在程式碼裡。
