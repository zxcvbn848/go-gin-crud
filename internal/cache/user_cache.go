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
	userKeyPrefix = "user:"
	userTTL       = 5 * time.Minute
)

// UserCache 單一用戶快取介面
type UserCache interface {
	GetUser(ctx context.Context, id uint) (*dto.UserResponse, error)
	SetUser(ctx context.Context, id uint, user *dto.UserResponse) error
	DeleteUser(ctx context.Context, id uint) error
}

type redisUserCache struct {
	client *redispkg.Client
}

func NewUserCache(client *redispkg.Client) UserCache {
	if client == nil {
		return nil
	}
	return &redisUserCache{client: client}
}

func (c *redisUserCache) key(id uint) string {
	return userKeyPrefix + fmt.Sprintf("%d", id)
}

func (c *redisUserCache) GetUser(ctx context.Context, id uint) (*dto.UserResponse, error) {
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
	var user dto.UserResponse
	if err := json.Unmarshal(val, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *redisUserCache) SetUser(ctx context.Context, id uint, user *dto.UserResponse) error {
	if c.client == nil || user == nil {
		return nil
	}
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(id), val, userTTL).Err()
}

func (c *redisUserCache) DeleteUser(ctx context.Context, id uint) error {
	if c.client == nil {
		return nil
	}
	return c.client.Del(ctx, c.key(id)).Err()
}
