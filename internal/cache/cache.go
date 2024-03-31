package cache

import (
	"context"
	"sync"
	"time"
)

type object struct {
	data []byte
	ttl  time.Time
}

type memoryCache struct {
	cache map[string]object
	mu    sync.RWMutex
}

func NewMemoryCache() *memoryCache {
	cache := &memoryCache{
		cache: make(map[string]object, 0),
		mu:    sync.RWMutex{},
	}

	cache.deleteExpiredEntries(context.Background())

	return cache
}

func (c *memoryCache) Get(ctx context.Context, namespace string, id string) ([]byte, error) {
	id = namespace + "." + id
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.cache[id]; ok {
		return v.data, nil
	}

	return nil, ErrNotExists
}

func (c *memoryCache) Put(ctx context.Context, namespace string, id string, data []byte, duration time.Duration) error {
	id = namespace + "." + id
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.cache[id]; ok {
		return ErrAlreadyExists
	}
	c.cache[id] = object{
		data: data,
		ttl:  time.Now().Add(duration),
	}

	return nil
}

func (c *memoryCache) Update(ctx context.Context, namespace string, id string, data []byte, duration time.Duration) error {
	id = namespace + "." + id
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.cache[id]; ok {
		c.cache[id] = object{
			data: data,
			ttl:  time.Now().Add(duration),
		}
		return nil
	}

	return ErrNotExists
}

func (c *memoryCache) Delete(ctx context.Context, namespace string, id string) error {
	id = namespace + "." + id
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.cache[id]; ok {
		delete(c.cache, id)
		return nil
	}

	return ErrNotExists
}
