package cache

import (
	"context"
	"time"

	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
	ttl    time.Duration
	log    *logger.Logger
}

func NewRedisCache(rc *redis.Client, ttl time.Duration) *redisCache {
	return &redisCache{
		client: rc,
		ttl:    ttl,
		log:    logger.New("Redis-Cache"),
	}
}

func (c *redisCache) Get(ctx context.Context, namespace string, id string) ([]byte, error) {
	id = namespace + "." + id
	val, err := c.client.Get(ctx, id).Bytes()
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (c *redisCache) Put(ctx context.Context, namespace string, id string, data []byte, duration time.Duration) error {
	id = namespace + "." + id
	err := c.client.Set(ctx, id, data, duration).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *redisCache) Update(ctx context.Context, namespace string, id string, data []byte, duration time.Duration) error {
	err := c.Delete(ctx, namespace, id)
	if err != nil {
		return err
	}
	err = c.Put(ctx, namespace, id, data, duration)
	if err != nil {
		return err
	}
	return nil
}

func (c *redisCache) Delete(ctx context.Context, namespace string, id string) error {
	id = namespace + "." + id
	err := c.client.Del(ctx, id).Err()
	if err != nil {
		return err
	}
	return nil
}
