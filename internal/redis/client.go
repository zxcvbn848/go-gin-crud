package redis

import (
	"context"
	"fmt"
	"time"

	rd "github.com/redis/go-redis/v9"
)

// Client 封裝 go-redis 客戶端，供快取層使用
type Client struct {
	*rd.Client
}

// NewClient 根據位址建立 Redis 連線。若 addr 為空則回傳 nil, nil（不啟用 Redis）
func NewClient(addr string) (*Client, error) {
	if addr == "" {
		return nil, nil
	}

	cli := rd.NewClient(&rd.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Client{Client: cli}, nil
}
