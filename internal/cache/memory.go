package cache

import (
	"sync"
	"time"
)

// MemoryCache implements an in-memory cache with TTL
type MemoryCache struct {
	data  map[string]CacheItem
	mutex sync.RWMutex
	ttl   time.Duration
}

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache with the given TTL
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	cache := &MemoryCache{
		data:  make(map[string]CacheItem),
		mutex: sync.RWMutex{},
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item.Data, true
}

// Set stores a value in the cache
func (c *MemoryCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = CacheItem{
		Data:      value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes a key from the cache
func (c *MemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
}

// cleanup removes expired items from the cache
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, item := range c.data {
			if now.After(item.ExpiresAt) {
				delete(c.data, key)
			}
		}
		c.mutex.Unlock()
	}
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() map[string]int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	total := len(c.data)
	expired := 0
	now := time.Now()

	for _, item := range c.data {
		if now.After(item.ExpiresAt) {
			expired++
		}
	}

	return map[string]int{
		"total":   total,
		"active":  total - expired,
		"expired": expired,
	}
}
