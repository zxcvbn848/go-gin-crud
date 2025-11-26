# go-gin-crud

## Docker MySQL

在專案根目錄可使用下列指令啟動 MySQL：

```bash
docker compose up -d
```

預設使用的資料庫資訊：

- host: `127.0.0.1` (容器 `go_gin_crud_mysql`)
- port: `3307`
- database: `goGinCRUD`
- user: `gogin`
- password: `a3935522`

需要關閉時執行：

```bash
docker compose down
```
