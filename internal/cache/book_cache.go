package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go-gin-crud/internal/dto"
	redispkg "go-gin-crud/internal/redis"

	"github.com/redis/go-redis/v9"
)

const (
	bookKeyPrefix = "book:"
	bookTTL       = 5 * time.Minute
)

// BookCache 單一書籍快取介面：Get 未命中或錯誤時由呼叫方回 DB 查並 Set
type BookCache interface {
	GetBook(ctx context.Context, id uint) (*dto.BookResponse, error)
	SetBook(ctx context.Context, id uint, book *dto.BookResponse) error
	DeleteBook(ctx context.Context, id uint) error
}

// redisBookCache 使用 Redis 實作的 BookCache
type redisBookCache struct {
	client *redispkg.Client
}

// NewBookCache 建立書籍快取。若 client 為 nil 則回傳 nil（不啟用快取）
func NewBookCache(client *redispkg.Client) BookCache {
	if client == nil {
		return nil
	}
	return &redisBookCache{client: client}
}

func (c *redisBookCache) key(id uint) string {
	return bookKeyPrefix + fmt.Sprintf("%d", id)
}

func (c *redisBookCache) GetBook(ctx context.Context, id uint) (*dto.BookResponse, error) {
	if c.client == nil {
		return nil, nil
	}
	key := c.key(id)
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // 快取未命中
		}
		return nil, err
	}
	var book dto.BookResponse
	if err := json.Unmarshal(val, &book); err != nil {
		return nil, err
	}
	return &book, nil
}

func (c *redisBookCache) SetBook(ctx context.Context, id uint, book *dto.BookResponse) error {
	if c.client == nil || book == nil {
		return nil
	}
	key := c.key(id)
	val, err := json.Marshal(book)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, val, bookTTL).Err()
}

func (c *redisBookCache) DeleteBook(ctx context.Context, id uint) error {
	if c.client == nil {
		return nil
	}
	return c.client.Del(ctx, c.key(id)).Err()
}
