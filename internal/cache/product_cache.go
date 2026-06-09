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
	productKeyPrefix = "product:"
	productTTL       = 5 * time.Minute
)

// ProductCache 單一產品快取介面
type ProductCache interface {
	GetProduct(ctx context.Context, id uint) (*dto.ProductResponse, error)
	SetProduct(ctx context.Context, id uint, product *dto.ProductResponse) error
	DeleteProduct(ctx context.Context, id uint) error
}

type redisProductCache struct {
	client *redispkg.Client
}

func NewProductCache(client *redispkg.Client) ProductCache {
	if client == nil {
		return nil
	}
	return &redisProductCache{client: client}
}

func (c *redisProductCache) key(id uint) string {
	return productKeyPrefix + fmt.Sprintf("%d", id)
}

func (c *redisProductCache) GetProduct(ctx context.Context, id uint) (*dto.ProductResponse, error) {
	if c.client == nil {
		return nil, nil
	}
	val, err := c.client.Get(ctx, c.key(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	var product dto.ProductResponse
	if err := json.Unmarshal(val, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (c *redisProductCache) SetProduct(ctx context.Context, id uint, product *dto.ProductResponse) error {
	if c.client == nil || product == nil {
		return nil
	}
	val, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(id), val, productTTL).Err()
}

func (c *redisProductCache) DeleteProduct(ctx context.Context, id uint) error {
	if c.client == nil {
		return nil
	}
	return c.client.Del(ctx, c.key(id)).Err()
}
