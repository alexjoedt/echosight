package cache

import (
	"context"
	"time"
)

func (c *memoryCache) deleteExpiredEntries(ctx context.Context) {
	go func() {
		t := time.NewTicker(time.Second * 15)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				c.mu.Lock()
				for key, value := range c.cache {
					if time.Since(value.ttl) >= time.Second {
						delete(c.cache, key)
					}
				}
				c.mu.Unlock()
			}
		}
	}()
}
