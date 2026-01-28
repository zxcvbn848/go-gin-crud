package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Demo10: 訂閱發布模式（Pub/Sub）
// 詳細說明請參考 cmd/concurrency/README.md
func demo10() {
	fmt.Println("\n=== Demo 10: 訂閱發布模式（Pub/Sub）===")

	fmt.Println("\n--- 場景 1: 基本訂閱發布（單個訂閱者）---")
	basicPubSub()

	fmt.Println("\n--- 場景 2: 多個訂閱者監聽同一事件 ---")
	multipleSubscribers()

	fmt.Println("\n--- 場景 3: 訂閱者離線處理 ---")
	subscriberOfflineHandling()

	fmt.Println("\n--- 場景 4: 消息持久化（保證消息不丟失）---")
	messagePersistence()

	fmt.Println("\nDemo 10 完成")
}

// ==================== 場景 1: 基本訂閱發布 ====================

// basicPubSub 演示基本的訂閱發布模式
func basicPubSub() {
	// 事件類型
	type Event struct {
		Type      string
		Message   string
		Timestamp time.Time
	}

	// 訂閱者通道類型
	type Subscriber chan Event

	// 發布者結構
	type Publisher struct {
		subscribers map[string]Subscriber
		mu          sync.RWMutex
		eventChan   chan Event
		ctx         context.Context
		cancel      context.CancelFunc
	}

	// 創建發布者
	createPublisher := func() *Publisher {
		ctx, cancel := context.WithCancel(context.Background())
		pub := &Publisher{
			subscribers: make(map[string]Subscriber),
			eventChan:   make(chan Event, 100),
			ctx:         ctx,
			cancel:      cancel,
		}

		// 啟動事件分發 goroutine
		go func() {
			for {
				select {
				case <-pub.ctx.Done():
					return
				case event := <-pub.eventChan:
					pub.mu.RLock()
					for id, sub := range pub.subscribers {
						select {
						case sub <- event:
							fmt.Printf("  [基本] 📤 事件已發送給訂閱者: %s\n", id)
						default:
							fmt.Printf("  [基本] ⚠️  訂閱者 %s 通道已滿，跳過\n", id)
						}
					}
					pub.mu.RUnlock()
				}
			}
		}()

		return pub
	}

	// 訂閱
	subscribe := func(pub *Publisher, subscriberID string) Subscriber {
		pub.mu.Lock()
		defer pub.mu.Unlock()

		sub := make(Subscriber, 10)
		pub.subscribers[subscriberID] = sub
		fmt.Printf("  [基本] ✅ 訂閱者 %s 已訂閱\n", subscriberID)
		return sub
	}

	// 取消訂閱
	unsubscribe := func(pub *Publisher, subscriberID string) {
		pub.mu.Lock()
		defer pub.mu.Unlock()

		if sub, exists := pub.subscribers[subscriberID]; exists {
			close(sub)
			delete(pub.subscribers, subscriberID)
			fmt.Printf("  [基本] ❌ 訂閱者 %s 已取消訂閱\n", subscriberID)
		}
	}

	// 發布事件
	publish := func(pub *Publisher, eventType, message string) {
		event := Event{
			Type:      eventType,
			Message:   message,
			Timestamp: time.Now(),
		}
		select {
		case pub.eventChan <- event:
			fmt.Printf("  [基本] 📢 發布事件: %s - %s\n", eventType, message)
		default:
			fmt.Printf("  [基本] ⚠️  事件通道已滿，無法發布\n")
		}
	}

	// 創建發布者
	pub := createPublisher()
	defer pub.cancel()

	// 訂閱者訂閱
	sub1 := subscribe(pub, "訂閱者-1")

	// 啟動訂閱者監聽 goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for event := range sub1 {
			fmt.Printf("  [基本] 📥 訂閱者-1 收到事件: [%s] %s (時間: %v)\n",
				event.Type, event.Message, event.Timestamp.Format("15:04:05"))
		}
		fmt.Printf("  [基本] 訂閱者-1 已停止監聽\n")
	}()

	// 發布幾個事件
	time.Sleep(100 * time.Millisecond)
	publish(pub, "通知", "系統維護將在 10 分鐘後開始")
	time.Sleep(200 * time.Millisecond)
	publish(pub, "警告", "CPU 使用率超過 80%")
	time.Sleep(200 * time.Millisecond)
	publish(pub, "信息", "新用戶註冊成功")

	// 等待事件處理
	time.Sleep(300 * time.Millisecond)

	// 取消訂閱
	unsubscribe(pub, "訂閱者-1")

	// 等待訂閱者 goroutine 結束
	wg.Wait()

	fmt.Printf("  [基本] ✅ 基本訂閱發布演示完成\n")
}

// ==================== 場景 2: 多個訂閱者 ====================

// multipleSubscribers 演示多個訂閱者監聽同一事件
func multipleSubscribers() {
	// 事件類型
	type Event struct {
		Type      string
		Message   string
		Timestamp time.Time
	}

	// 訂閱者通道類型
	type Subscriber chan Event

	// 發布者結構
	type Publisher struct {
		subscribers map[string]Subscriber
		mu          sync.RWMutex
		eventChan   chan Event
		ctx         context.Context
		cancel      context.CancelFunc
	}

	// 創建發布者
	createPublisher := func() *Publisher {
		ctx, cancel := context.WithCancel(context.Background())
		pub := &Publisher{
			subscribers: make(map[string]Subscriber),
			eventChan:   make(chan Event, 100),
			ctx:         ctx,
			cancel:      cancel,
		}

		// 啟動事件分發 goroutine
		go func() {
			for {
				select {
				case <-pub.ctx.Done():
					return
				case event := <-pub.eventChan:
					pub.mu.RLock()
					subscriberCount := len(pub.subscribers)
					for id, sub := range pub.subscribers {
						select {
						case sub <- event:
							// 事件已發送
						default:
							fmt.Printf("  [多訂閱者] ⚠️  訂閱者 %s 通道已滿，跳過\n", id)
						}
					}
					pub.mu.RUnlock()
					fmt.Printf("  [多訂閱者] 📢 事件已廣播給 %d 個訂閱者\n", subscriberCount)
				}
			}
		}()

		return pub
	}

	// 訂閱
	subscribe := func(pub *Publisher, subscriberID string) Subscriber {
		pub.mu.Lock()
		defer pub.mu.Unlock()

		sub := make(Subscriber, 10)
		pub.subscribers[subscriberID] = sub
		fmt.Printf("  [多訂閱者] ✅ 訂閱者 %s 已訂閱\n", subscriberID)
		return sub
	}

	// 取消訂閱
	unsubscribe := func(pub *Publisher, subscriberID string) {
		pub.mu.Lock()
		defer pub.mu.Unlock()

		if sub, exists := pub.subscribers[subscriberID]; exists {
			close(sub)
			delete(pub.subscribers, subscriberID)
			fmt.Printf("  [多訂閱者] ❌ 訂閱者 %s 已取消訂閱\n", subscriberID)
		}
	}

	// 發布事件
	publish := func(pub *Publisher, eventType, message string) {
		event := Event{
			Type:      eventType,
			Message:   message,
			Timestamp: time.Now(),
		}
		select {
		case pub.eventChan <- event:
			fmt.Printf("  [多訂閱者] 📢 發布事件: %s - %s\n", eventType, message)
		default:
			fmt.Printf("  [多訂閱者] ⚠️  事件通道已滿，無法發布\n")
		}
	}

	// 創建發布者
	pub := createPublisher()
	defer pub.cancel()

	// 創建多個訂閱者
	subscriberIDs := []string{"用戶-A", "用戶-B", "用戶-C", "用戶-D"}
	subscribers := make(map[string]Subscriber)

	for _, id := range subscriberIDs {
		subscribers[id] = subscribe(pub, id)
	}

	// 啟動所有訂閱者的監聽 goroutine
	var wg sync.WaitGroup
	for id, sub := range subscribers {
		wg.Add(1)
		go func(subID string, subChan Subscriber) {
			defer wg.Done()
			for event := range subChan {
				fmt.Printf("  [多訂閱者] 📥 %s 收到事件: [%s] %s\n",
					subID, event.Type, event.Message)
			}
			fmt.Printf("  [多訂閱者] %s 已停止監聽\n", subID)
		}(id, sub)
	}

	// 發布幾個事件
	time.Sleep(100 * time.Millisecond)
	publish(pub, "系統通知", "新版本已發布，請更新應用")
	time.Sleep(200 * time.Millisecond)
	publish(pub, "促銷活動", "限時優惠：全場商品 8 折")
	time.Sleep(200 * time.Millisecond)
	publish(pub, "維護通知", "系統將在今晚 23:00 進行維護")

	// 等待事件處理
	time.Sleep(300 * time.Millisecond)

	// 取消所有訂閱
	for id := range subscribers {
		unsubscribe(pub, id)
	}

	// 等待所有訂閱者 goroutine 結束
	wg.Wait()

	fmt.Printf("  [多訂閱者] ✅ 多個訂閱者演示完成\n")
}

// ==================== 場景 3: 訂閱者離線處理 ====================

// subscriberOfflineHandling 演示訂閱者離線處理
func subscriberOfflineHandling() {
	// 事件類型
	type Event struct {
		Type      string
		Message   string
		Timestamp time.Time
	}

	// 訂閱者信息
	type SubscriberInfo struct {
		ID        string
		Channel   chan Event
		LastSeen  time.Time
		IsOnline  bool
		Context   context.Context
		Cancel    context.CancelFunc
	}

	// 發布者結構
	type Publisher struct {
		subscribers map[string]*SubscriberInfo
		mu          sync.RWMutex
		eventChan   chan Event
		ctx         context.Context
		cancel      context.CancelFunc
	}

	// 創建發布者
	createPublisher := func() *Publisher {
		ctx, cancel := context.WithCancel(context.Background())
		pub := &Publisher{
			subscribers: make(map[string]*SubscriberInfo),
			eventChan:   make(chan Event, 100),
			ctx:         ctx,
			cancel:      cancel,
		}

		// 啟動事件分發 goroutine
		go func() {
			for {
				select {
				case <-pub.ctx.Done():
					return
				case event := <-pub.eventChan:
					pub.mu.Lock()
					// 檢查並清理離線訂閱者
					for id, subInfo := range pub.subscribers {
						if !subInfo.IsOnline {
							fmt.Printf("  [離線處理] 🗑️  清理離線訂閱者: %s\n", id)
							close(subInfo.Channel)
							delete(pub.subscribers, id)
							continue
						}

						// 嘗試發送事件
						select {
						case subInfo.Channel <- event:
							subInfo.LastSeen = time.Now()
						default:
							fmt.Printf("  [離線處理] ⚠️  訂閱者 %s 通道已滿，標記為離線\n", id)
							subInfo.IsOnline = false
						}
					}
					pub.mu.Unlock()
				}
			}
		}()

		// 啟動健康檢查 goroutine
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-pub.ctx.Done():
					return
				case <-ticker.C:
					pub.mu.Lock()
					now := time.Now()
					for id, subInfo := range pub.subscribers {
						// 如果訂閱者超過 3 秒沒有活動，標記為離線
						if subInfo.IsOnline && now.Sub(subInfo.LastSeen) > 3*time.Second {
							fmt.Printf("  [離線處理] ⏰ 訂閱者 %s 超時未響應，標記為離線\n", id)
							subInfo.IsOnline = false
						}
					}
					pub.mu.Unlock()
				}
			}
		}()

		return pub
	}

	// 訂閱
	subscribe := func(pub *Publisher, subscriberID string) *SubscriberInfo {
		pub.mu.Lock()
		defer pub.mu.Unlock()

		ctx, cancel := context.WithCancel(context.Background())
		subInfo := &SubscriberInfo{
			ID:       subscriberID,
			Channel:  make(chan Event, 10),
			LastSeen: time.Now(),
			IsOnline: true,
			Context:  ctx,
			Cancel:   cancel,
		}
		pub.subscribers[subscriberID] = subInfo
		fmt.Printf("  [離線處理] ✅ 訂閱者 %s 已訂閱\n", subscriberID)
		return subInfo
	}

	// 取消訂閱
	unsubscribe := func(pub *Publisher, subscriberID string) {
		pub.mu.Lock()
		defer pub.mu.Unlock()

		if subInfo, exists := pub.subscribers[subscriberID]; exists {
			subInfo.Cancel()
			close(subInfo.Channel)
			delete(pub.subscribers, subscriberID)
			fmt.Printf("  [離線處理] ❌ 訂閱者 %s 已取消訂閱\n", subscriberID)
		}
	}

	// 發布事件
	publish := func(pub *Publisher, eventType, message string) {
		event := Event{
			Type:      eventType,
			Message:   message,
			Timestamp: time.Now(),
		}
		select {
		case pub.eventChan <- event:
			fmt.Printf("  [離線處理] 📢 發布事件: %s - %s\n", eventType, message)
		default:
			fmt.Printf("  [離線處理] ⚠️  事件通道已滿，無法發布\n")
		}
	}

	// 創建發布者
	pub := createPublisher()
	defer pub.cancel()

	// 創建訂閱者
	sub1 := subscribe(pub, "訂閱者-在線")
	sub2 := subscribe(pub, "訂閱者-將離線")

	// 啟動訂閱者監聽 goroutine
	var wg sync.WaitGroup

	// 訂閱者 1：正常在線
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-sub1.Context.Done():
				return
			case event, ok := <-sub1.Channel:
				if !ok {
					return
				}
				fmt.Printf("  [離線處理] 📥 %s 收到事件: [%s] %s\n",
					sub1.ID, event.Type, event.Message)
				// 更新最後活動時間
				sub1.LastSeen = time.Now()
			}
		}
	}()

	// 訂閱者 2：模擬離線（不讀取通道）
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 只讀取前兩個事件，然後停止讀取（模擬離線）
		count := 0
		for {
			select {
			case <-sub2.Context.Done():
				return
			case event, ok := <-sub2.Channel:
				if !ok {
					return
				}
				if count < 2 {
					fmt.Printf("  [離線處理] 📥 %s 收到事件: [%s] %s\n",
						sub2.ID, event.Type, event.Message)
					sub2.LastSeen = time.Now()
					count++
				} else {
					// 停止讀取，模擬離線
					fmt.Printf("  [離線處理] 💤 %s 停止讀取事件（模擬離線）\n", sub2.ID)
					// 不更新 LastSeen，讓健康檢查檢測到超時
					time.Sleep(5 * time.Second) // 等待健康檢查檢測
					return
				}
			}
		}
	}()

	// 發布幾個事件
	time.Sleep(100 * time.Millisecond)
	publish(pub, "事件-1", "第一個事件")
	time.Sleep(200 * time.Millisecond)
	publish(pub, "事件-2", "第二個事件")
	time.Sleep(200 * time.Millisecond)
	publish(pub, "事件-3", "第三個事件")
	time.Sleep(200 * time.Millisecond)
	publish(pub, "事件-4", "第四個事件（訂閱者-將離線應該已離線）")

	// 等待健康檢查檢測離線
	time.Sleep(4 * time.Second)

	// 發布更多事件（離線訂閱者應該已被清理）
	publish(pub, "事件-5", "第五個事件（只有在線訂閱者會收到）")

	// 等待事件處理
	time.Sleep(500 * time.Millisecond)

	// 取消所有訂閱
	unsubscribe(pub, "訂閱者-在線")
	unsubscribe(pub, "訂閱者-將離線")

	// 等待所有訂閱者 goroutine 結束
	wg.Wait()

	fmt.Printf("  [離線處理] ✅ 訂閱者離線處理演示完成\n")
}

// ==================== 場景 4: 消息持久化 ====================

// messagePersistence 演示消息持久化（保證消息不丟失）
func messagePersistence() {
	// 事件類型
	type Event struct {
		ID        int
		Type      string
		Message   string
		Timestamp time.Time
	}

	// 訂閱者信息
	type SubscriberInfo struct {
		ID          string
		Channel     chan Event
		LastEventID int // 最後處理的事件 ID
		IsOnline    bool
		Context     context.Context
		Cancel      context.CancelFunc
	}

	// 發布者結構（帶消息持久化）
	type Publisher struct {
		subscribers   map[string]*SubscriberInfo
		mu            sync.RWMutex
		eventChan     chan Event
		eventHistory  []Event // 消息歷史（持久化）
		maxHistory    int     // 最大歷史記錄數
		nextEventID   int     // 下一個事件 ID
		ctx           context.Context
		cancel        context.CancelFunc
	}

	// 創建發布者
	createPublisher := func(maxHistory int) *Publisher {
		ctx, cancel := context.WithCancel(context.Background())
		pub := &Publisher{
			subscribers:  make(map[string]*SubscriberInfo),
			eventChan:    make(chan Event, 100),
			eventHistory: make([]Event, 0, maxHistory),
			maxHistory:   maxHistory,
			nextEventID:  1,
			ctx:          ctx,
			cancel:       cancel,
		}

		// 啟動事件分發 goroutine
		go func() {
			for {
				select {
				case <-pub.ctx.Done():
					return
				case event := <-pub.eventChan:
					// 持久化消息
					pub.mu.Lock()
					pub.eventHistory = append(pub.eventHistory, event)
					// 保持歷史記錄不超過最大值
					if len(pub.eventHistory) > pub.maxHistory {
						pub.eventHistory = pub.eventHistory[1:]
					}

					// 發送給所有在線訂閱者
					for id, subInfo := range pub.subscribers {
						if !subInfo.IsOnline {
							continue
						}

						select {
						case subInfo.Channel <- event:
							subInfo.LastEventID = event.ID
						default:
							fmt.Printf("  [持久化] ⚠️  訂閱者 %s 通道已滿\n", id)
						}
					}
					pub.mu.Unlock()

					fmt.Printf("  [持久化] 📢 事件 #%d 已發布並持久化: %s - %s\n",
						event.ID, event.Type, event.Message)
				}
			}
		}()

		return pub
	}

	// 訂閱（支持從指定事件 ID 開始）
	subscribe := func(pub *Publisher, subscriberID string, fromEventID int) *SubscriberInfo {
		pub.mu.Lock()
		defer pub.mu.Unlock()

		ctx, cancel := context.WithCancel(context.Background())
		subInfo := &SubscriberInfo{
			ID:          subscriberID,
			Channel:     make(chan Event, 10),
			LastEventID: fromEventID - 1,
			IsOnline:    true,
			Context:     ctx,
			Cancel:      cancel,
		}

		// 如果指定了起始事件 ID，發送歷史消息
		if fromEventID > 0 && fromEventID <= len(pub.eventHistory) {
			startIdx := fromEventID - 1
			if startIdx < 0 {
				startIdx = 0
			}
			for i := startIdx; i < len(pub.eventHistory); i++ {
				event := pub.eventHistory[i]
				select {
				case subInfo.Channel <- event:
					subInfo.LastEventID = event.ID
					fmt.Printf("  [持久化] 📤 向 %s 發送歷史事件 #%d: %s\n",
						subscriberID, event.ID, event.Message)
				default:
					fmt.Printf("  [持久化] ⚠️  無法向 %s 發送歷史事件 #%d\n", subscriberID, event.ID)
				}
			}
		}

		pub.subscribers[subscriberID] = subInfo
		fmt.Printf("  [持久化] ✅ 訂閱者 %s 已訂閱（從事件 #%d 開始）\n", subscriberID, subInfo.LastEventID+1)
		return subInfo
	}

	// 取消訂閱
	unsubscribe := func(pub *Publisher, subscriberID string) {
		pub.mu.Lock()
		defer pub.mu.Unlock()

		if subInfo, exists := pub.subscribers[subscriberID]; exists {
			subInfo.Cancel()
			close(subInfo.Channel)
			delete(pub.subscribers, subscriberID)
			fmt.Printf("  [持久化] ❌ 訂閱者 %s 已取消訂閱（最後處理事件 #%d）\n",
				subscriberID, subInfo.LastEventID)
		}
	}

	// 發布事件
	publish := func(pub *Publisher, eventType, message string) {
		pub.mu.Lock()
		eventID := pub.nextEventID
		pub.nextEventID++
		pub.mu.Unlock()

		event := Event{
			ID:        eventID,
			Type:      eventType,
			Message:   message,
			Timestamp: time.Now(),
		}
		select {
		case pub.eventChan <- event:
		default:
			fmt.Printf("  [持久化] ⚠️  事件通道已滿，無法發布\n")
		}
	}

	// 創建發布者（保留最近 10 條消息）
	pub := createPublisher(10)
	defer pub.cancel()

	// 發布一些初始事件
	time.Sleep(100 * time.Millisecond)
	publish(pub, "通知", "系統啟動")
	time.Sleep(100 * time.Millisecond)
	publish(pub, "通知", "數據庫連接成功")
	time.Sleep(100 * time.Millisecond)
	publish(pub, "通知", "緩存服務已就緒")

	// 等待事件處理
	time.Sleep(200 * time.Millisecond)

	// 訂閱者 1：從頭開始訂閱（會收到所有歷史消息）
	sub1 := subscribe(pub, "訂閱者-新用戶", 1)

	// 訂閱者 2：從事件 #2 開始訂閱（只收到部分歷史消息）
	sub2 := subscribe(pub, "訂閱者-斷線重連", 2)

	// 啟動訂閱者監聽 goroutine
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-sub1.Context.Done():
				return
			case event, ok := <-sub1.Channel:
				if !ok {
					return
				}
				fmt.Printf("  [持久化] 📥 %s 收到事件 #%d: [%s] %s\n",
					sub1.ID, event.ID, event.Type, event.Message)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-sub2.Context.Done():
				return
			case event, ok := <-sub2.Channel:
				if !ok {
					return
				}
				fmt.Printf("  [持久化] 📥 %s 收到事件 #%d: [%s] %s\n",
					sub2.ID, event.ID, event.Type, event.Message)
			}
		}
	}()

	// 等待歷史消息處理
	time.Sleep(300 * time.Millisecond)

	// 發布新事件
	fmt.Println("\n  [持久化] --- 發布新事件 ---")
	publish(pub, "警告", "CPU 使用率過高")
	time.Sleep(200 * time.Millisecond)
	publish(pub, "信息", "新訂單已創建")

	// 等待事件處理
	time.Sleep(300 * time.Millisecond)

	// 模擬訂閱者 3：斷線後重新連接，從上次處理的事件繼續
	fmt.Println("\n  [持久化] --- 模擬斷線重連 ---")
	sub3 := subscribe(pub, "訂閱者-斷線重連-2", 4) // 從事件 #4 開始

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-sub3.Context.Done():
				return
			case event, ok := <-sub3.Channel:
				if !ok {
					return
				}
				fmt.Printf("  [持久化] 📥 %s 收到事件 #%d: [%s] %s\n",
					sub3.ID, event.ID, event.Type, event.Message)
			}
		}
	}()

	// 等待歷史消息處理
	time.Sleep(300 * time.Millisecond)

	// 發布更多新事件
	publish(pub, "通知", "系統維護完成")

	// 等待事件處理
	time.Sleep(300 * time.Millisecond)

	// 取消所有訂閱
	unsubscribe(pub, "訂閱者-新用戶")
	unsubscribe(pub, "訂閱者-斷線重連")
	unsubscribe(pub, "訂閱者-斷線重連-2")

	// 等待所有訂閱者 goroutine 結束
	wg.Wait()

	fmt.Printf("  [持久化] ✅ 消息持久化演示完成\n")
	fmt.Printf("  [持久化] 💡 關鍵點：\n")
	fmt.Printf("    - 消息歷史記錄確保訂閱者不會丟失消息\n")
	fmt.Printf("    - 訂閱者可以從指定事件 ID 開始接收（支持斷線重連）\n")
	fmt.Printf("    - 歷史記錄有大小限制，避免內存無限增長\n")
}
