package cache

import (
	"sync"
	"time"
)

type item struct {
	value      interface{}
	expiration int64
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]item
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		items: make(map[string]item),
	}
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			c.cleanup()
		}
	}()
	return c
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	it, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		return nil, false
	}

	if it.expiration > 0 && time.Now().UnixNano() > it.expiration {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}

	return it.value, true
}

func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	c.items[key] = item{
		value:      value,
		expiration: exp,
	}
	c.mu.Unlock()
}

func (c *MemoryCache) Flush() {
	c.mu.Lock()
	c.items = make(map[string]item)
	c.mu.Unlock()
}

func (c *MemoryCache) cleanup() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	for k, it := range c.items {
		if it.expiration > 0 && now > it.expiration {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}
