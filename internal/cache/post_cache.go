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
	postKeyPrefix = "post:"
	postTTL       = 5 * time.Minute
)

// PostCache 單一文章快取介面
type PostCache interface {
	GetPost(ctx context.Context, id uint) (*dto.PostResponse, error)
	SetPost(ctx context.Context, id uint, post *dto.PostResponse) error
	DeletePost(ctx context.Context, id uint) error
}

type redisPostCache struct {
	client *redispkg.Client
}

func NewPostCache(client *redispkg.Client) PostCache {
	if client == nil {
		return nil
	}
	return &redisPostCache{client: client}
}

func (c *redisPostCache) key(id uint) string {
	return postKeyPrefix + fmt.Sprintf("%d", id)
}

func (c *redisPostCache) GetPost(ctx context.Context, id uint) (*dto.PostResponse, error) {
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
	var post dto.PostResponse
	if err := json.Unmarshal(val, &post); err != nil {
		return nil, err
	}
	return &post, nil
}

func (c *redisPostCache) SetPost(ctx context.Context, id uint, post *dto.PostResponse) error {
	if c.client == nil || post == nil {
		return nil
	}
	val, err := json.Marshal(post)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(id), val, postTTL).Err()
}

func (c *redisPostCache) DeletePost(ctx context.Context, id uint) error {
	if c.client == nil {
		return nil
	}
	return c.client.Del(ctx, c.key(id)).Err()
}
