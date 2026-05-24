package cache

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func NewRedisClient(connStr string) (*redis.Client, error) {
	if strings.HasPrefix(connStr, "redis://") || strings.HasPrefix(connStr, "rediss://") {
		opts, err := redis.ParseURL(connStr)
		if err != nil {
			return nil, err
		}
		return redis.NewClient(opts), nil
	}

	return redis.NewClient(&redis.Options{Addr: connStr}), nil
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}