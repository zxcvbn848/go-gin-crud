package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Demo7: 連接池管理（Connection Pool）
// 詳細說明請參考 cmd/concurrency/README.md
func demo7() {
	fmt.Println("\n=== Demo 7: 連接池管理 ===")

	fmt.Println("\n--- 場景 1: 無連接池（頻繁建立/關閉連接）---")
	withoutConnectionPool()

	fmt.Println("\n--- 場景 2: 使用連接池（複用連接）---")
	withConnectionPool()

	fmt.Println("\n--- 場景 3: 連接池超時和優雅關閉 ---")
	connectionPoolWithTimeout()

	fmt.Println("\nDemo 7 完成")
}

// ==================== 場景 1: 無連接池 ====================

// withoutConnectionPool 演示無連接池的問題
func withoutConnectionPool() {
	// 模擬數據庫連接
	type DBConnection struct {
		ID        int
		CreatedAt time.Time
	}

	// 模擬建立連接（耗時操作）
	createConnection := func(id int) *DBConnection {
		fmt.Printf("  [無池] 🔴 建立新連接 #%d (耗時操作)\n", id)
		time.Sleep(100 * time.Millisecond) // 模擬建立連接的耗時
		return &DBConnection{
			ID:        id,
			CreatedAt: time.Now(),
		}
	}

	// 模擬關閉連接
	closeConnection := func(conn *DBConnection) {
		fmt.Printf("  [無池] 🔴 關閉連接 #%d\n", conn.ID)
		time.Sleep(50 * time.Millisecond) // 模擬關閉連接的耗時
	}

	// 模擬使用連接執行查詢
	useConnection := func(conn *DBConnection) {
		fmt.Printf("  [無池] 使用連接 #%d 執行查詢\n", conn.ID)
		time.Sleep(50 * time.Millisecond) // 模擬查詢耗時
	}

	// 模擬 10 個請求，每個都需要建立新連接
	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 建立連接
			conn := createConnection(id)
			// 使用連接
			useConnection(conn)
			// 關閉連接
			closeConnection(conn)
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("  [無池] ⚠️  總耗時: %v，建立了 10 個連接並全部關閉\n", elapsed)
}

// ==================== 場景 2: 使用連接池 ====================

// Connection 連接接口
type Connection interface {
	ID() int
	IsHealthy() bool
	Close() error
}

// DBConnection 模擬數據庫連接
type DBConnection struct {
	id        int
	createdAt time.Time
	lastUsed  time.Time
	mu        sync.Mutex
}

func (c *DBConnection) ID() int {
	return c.id
}

func (c *DBConnection) IsHealthy() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 連接超過 1 小時未使用視為不健康
	return time.Since(c.lastUsed) < time.Hour
}

func (c *DBConnection) Close() error {
	fmt.Printf("  [有池] 關閉連接 #%d\n", c.id)
	return nil
}

func (c *DBConnection) Use() {
	c.mu.Lock()
	c.lastUsed = time.Now()
	c.mu.Unlock()
}

// withConnectionPool 演示使用連接池的優勢
func withConnectionPool() {

	// 連接池
	type ConnectionPool struct {
		pool      chan Connection
		maxSize   int
		created   int
		createdMu sync.Mutex
	}

	// 創建連接池
	createPool := func(maxSize int) *ConnectionPool {
		pool := &ConnectionPool{
			pool:    make(chan Connection, maxSize),
			maxSize: maxSize,
		}

		// 預先創建一些連接
		for i := 0; i < maxSize/2; i++ {
			pool.createdMu.Lock()
			pool.created++
			id := pool.created
			pool.createdMu.Unlock()

			conn := &DBConnection{
				id:        id,
				createdAt: time.Now(),
				lastUsed:  time.Now(),
			}
			pool.pool <- conn
			fmt.Printf("  [有池] ✅ 預先創建連接 #%d\n", id)
		}

		return pool
	}

	// 獲取連接
	getConnection := func(pool *ConnectionPool, timeout time.Duration) (Connection, error) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		select {
		case conn := <-pool.pool:
			// 從池中獲取連接
			if dbConn, ok := conn.(*DBConnection); ok {
				dbConn.Use()
			}
			fmt.Printf("  [有池] ✅ 從池中獲取連接 #%d\n", conn.ID())
			return conn, nil

		case <-ctx.Done():
			// 超時，創建新連接
			pool.createdMu.Lock()
			pool.created++
			id := pool.created
			pool.createdMu.Unlock()

			fmt.Printf("  [有池] ⚠️  池中無可用連接，創建新連接 #%d\n", id)
			time.Sleep(100 * time.Millisecond) // 模擬建立連接的耗時
			conn := &DBConnection{
				id:        id,
				createdAt: time.Now(),
				lastUsed:  time.Now(),
			}
			return conn, nil
		}
	}

	// 歸還連接
	returnConnection := func(pool *ConnectionPool, conn Connection) {
		// 檢查連接是否健康
		if !conn.IsHealthy() {
			fmt.Printf("  [有池] ⚠️  連接 #%d 不健康，關閉並丟棄\n", conn.ID())
			_ = conn.Close()
			return
		}

		select {
		case pool.pool <- conn:
			// 成功歸還到池中
			fmt.Printf("  [有池] ✅ 歸還連接 #%d 到池中\n", conn.ID())
		default:
			// 池已滿，關閉連接
			fmt.Printf("  [有池] ⚠️  池已滿，關閉連接 #%d\n", conn.ID())
			_ = conn.Close()
		}
	}

	// 使用連接執行查詢
	useConnection := func(conn Connection) {
		if dbConn, ok := conn.(*DBConnection); ok {
			dbConn.Use()
		}
		fmt.Printf("  [有池] 使用連接 #%d 執行查詢\n", conn.ID())
		time.Sleep(50 * time.Millisecond) // 模擬查詢耗時
	}

	// 創建連接池（最大 5 個連接）
	pool := createPool(5)

	// 模擬 10 個請求
	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 獲取連接
			conn, err := getConnection(pool, 1*time.Second)
			if err != nil {
				fmt.Printf("  [有池] 請求 %d: 獲取連接失敗: %v\n", id, err)
				return
			}

			// 使用連接
			useConnection(conn)

			// 歸還連接
			returnConnection(pool, conn)
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Printf("  [有池] ✅ 總耗時: %v，連接被複用，性能提升明顯\n", elapsed)
}

// ==================== 場景 3: 連接池超時和優雅關閉 ====================

// DBConnectionWithTimeout 帶超時的數據庫連接
type DBConnectionWithTimeout struct {
	id        int
	createdAt time.Time
	lastUsed  time.Time
	mu        sync.Mutex
}

func (c *DBConnectionWithTimeout) ID() int {
	return c.id
}

func (c *DBConnectionWithTimeout) IsHealthy() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Since(c.lastUsed) < time.Hour
}

func (c *DBConnectionWithTimeout) Close() error {
	fmt.Printf("  [超時] 關閉連接 #%d\n", c.id)
	return nil
}

func (c *DBConnectionWithTimeout) Use() {
	c.mu.Lock()
	c.lastUsed = time.Now()
	c.mu.Unlock()
}

// connectionPoolWithTimeout 帶超時和優雅關閉的連接池
func connectionPoolWithTimeout() {

	// 完整的連接池實現
	type ConnectionPool struct {
		pool      chan Connection
		maxSize   int
		created   int
		createdMu sync.Mutex
		ctx       context.Context
		cancel    context.CancelFunc
		closed    bool
		closedMu  sync.Mutex
	}

	// 創建連接池
	createPool := func(maxSize int) *ConnectionPool {
		ctx, cancel := context.WithCancel(context.Background())
		pool := &ConnectionPool{
			pool:    make(chan Connection, maxSize),
			maxSize: maxSize,
			ctx:     ctx,
			cancel:  cancel,
		}

		// 預先創建連接
		for i := 0; i < maxSize/2; i++ {
			pool.createdMu.Lock()
			pool.created++
			id := pool.created
			pool.createdMu.Unlock()

			conn := &DBConnectionWithTimeout{
				id:        id,
				createdAt: time.Now(),
				lastUsed:  time.Now(),
			}
			pool.pool <- conn
		}

		return pool
	}

	// 獲取連接（帶超時）
	getConnection := func(pool *ConnectionPool, timeout time.Duration) (Connection, error) {
		// 檢查連接池是否已關閉
		pool.closedMu.Lock()
		if pool.closed {
			pool.closedMu.Unlock()
			return nil, fmt.Errorf("連接池已關閉")
		}
		pool.closedMu.Unlock()

		// 創建帶超時的 context
		ctx, cancel := context.WithTimeout(pool.ctx, timeout)
		defer cancel()

		select {
		case conn := <-pool.pool:
			// 從池中獲取連接
			if dbConn, ok := conn.(*DBConnectionWithTimeout); ok {
				dbConn.Use()
			}
			return conn, nil

		case <-ctx.Done():
			// 超時或被取消
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("獲取連接超時")
			}
			return nil, fmt.Errorf("連接池已關閉")
		}
	}

	// 歸還連接
	returnConnection := func(pool *ConnectionPool, conn Connection) {
		pool.closedMu.Lock()
		if pool.closed {
			pool.closedMu.Unlock()
			// 連接池已關閉，直接關閉連接
			_ = conn.Close()
			return
		}
		pool.closedMu.Unlock()

		// 檢查連接是否健康
		if !conn.IsHealthy() {
			_ = conn.Close()
			return
		}

		select {
		case pool.pool <- conn:
			// 成功歸還
		default:
			// 池已滿，關閉連接
			_ = conn.Close()
		}
	}

	// 優雅關閉連接池
	shutdownPool := func(pool *ConnectionPool, timeout time.Duration) error {
		pool.closedMu.Lock()
		if pool.closed {
			pool.closedMu.Unlock()
			return fmt.Errorf("連接池已經關閉")
		}
		pool.closed = true
		pool.closedMu.Unlock()

		// 發送關閉信號
		pool.cancel()

		// 創建超時 context
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// 關閉所有連接
		done := make(chan bool, 1)
		go func() {
			closeCount := 0
			for {
				select {
				case conn := <-pool.pool:
					_ = conn.Close()
					closeCount++
				default:
					done <- true
					return
				}
			}
		}()

		// 等待關閉完成或超時
		select {
		case <-done:
			fmt.Printf("  [超時] ✅ 連接池已優雅關閉\n")
			return nil
		case <-ctx.Done():
			return fmt.Errorf("關閉連接池超時")
		}
	}

	// 使用連接執行查詢
	useConnection := func(conn Connection) {
		if dbConn, ok := conn.(*DBConnectionWithTimeout); ok {
			dbConn.Use()
		}
		time.Sleep(50 * time.Millisecond)
	}

	// 創建連接池
	pool := createPool(5)

	// 測試場景 1: 正常使用
	fmt.Println("\n  [超時] 測試 1: 正常使用連接池")
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			conn, err := getConnection(pool, 1*time.Second)
			if err != nil {
				fmt.Printf("  [超時] 請求 %d: 獲取連接失敗: %v\n", id, err)
				return
			}

			useConnection(conn)
			returnConnection(pool, conn)
		}(i)
	}
	wg.Wait()

	// 測試場景 2: 超時處理
	fmt.Println("\n  [超時] 測試 2: 連接獲取超時")
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 設置很短的超時時間，但池中可能沒有連接
		conn, err := getConnection(pool, 10*time.Millisecond)
		if err != nil {
			fmt.Printf("  [超時] ✅ 正確處理超時: %v\n", err)
			return
		}
		returnConnection(pool, conn)
	}()
	wg.Wait()

	// 測試場景 3: 優雅關閉
	fmt.Println("\n  [超時] 測試 3: 優雅關閉連接池")
	err := shutdownPool(pool, 5*time.Second)
	if err != nil {
		fmt.Printf("  [超時] ⚠️  關閉失敗: %v\n", err)
	} else {
		fmt.Printf("  [超時] ✅ 連接池已優雅關閉\n")
	}

	// 測試場景 4: 關閉後嘗試獲取連接
	fmt.Println("\n  [超時] 測試 4: 關閉後嘗試獲取連接")
	conn, err := getConnection(pool, 1*time.Second)
	if err != nil {
		fmt.Printf("  [超時] ✅ 正確處理：連接池已關閉，無法獲取連接: %v\n", err)
	} else {
		returnConnection(pool, conn)
	}
}
