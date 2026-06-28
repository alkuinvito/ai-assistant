package cache

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Cache interface {
	Del(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (string, error)
	GetObject(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key, value string, expiration ...time.Duration) error
	SetObject(ctx context.Context, key string, value any, expiration ...time.Duration) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(logger *logrus.Logger) (Cache, func()) {
	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})

	cleanup := func() {
		if err := client.Close(); err != nil {
			logger.WithError(err).Fatalf("failed to close Redis client")
		}
	}

	return &RedisCache{client: client}, cleanup
}

func (c *RedisCache) Del(ctx context.Context, key string) error {
	_, err := c.client.Del(ctx, key).Result()
	return err
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

func (c *RedisCache) GetObject(ctx context.Context, key string, dest any) error {
	val, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return err
	}

	return nil
}

func (c *RedisCache) Set(ctx context.Context, key, value string, expiration ...time.Duration) error {
	var exp time.Duration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	_, err := c.client.Set(ctx, key, value, exp).Result()
	return err
}

func (c *RedisCache) SetObject(ctx context.Context, key string, value any, expiration ...time.Duration) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.Set(ctx, key, string(val), expiration...)
}
