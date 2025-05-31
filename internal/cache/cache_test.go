package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)

	// Test setting and getting a value
	cache.Set("key1", "value1")

	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Test getting non-existent key
	_, found = cache.Get("nonexistent")
	assert.False(t, found)
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache(100 * time.Millisecond)

	// Set a value
	cache.Set("key1", "value1")

	// Should be available immediately
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	_, found = cache.Get("key1")
	assert.False(t, found)
}

func TestMemoryCache_Stats(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)

	// Initially empty
	stats := cache.Stats()
	assert.Equal(t, 0, stats["active"])
	assert.Equal(t, 0, stats["expired"])
	assert.Equal(t, 0, stats["total"])

	// Add some items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	stats = cache.Stats()
	assert.Equal(t, 2, stats["active"])
	assert.Equal(t, 0, stats["expired"])
	assert.Equal(t, 2, stats["total"])
}

func TestMemoryCache_CleanupExpired(t *testing.T) {
	cache := NewMemoryCache(50 * time.Millisecond)

	// Add items
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Verify items are there initially
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	assert.True(t, found1)
	assert.True(t, found2)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Items should be expired now
	_, found1 = cache.Get("key1")
	_, found2 = cache.Get("key2")
	assert.False(t, found1)
	assert.False(t, found2)

	// Add new item
	cache.Set("key3", "value3")

	// Manual cleanup would happen in background, but for testing just verify behavior
	stats := cache.Stats()

	// Should have 1 active item (key3), but expired items might still be counted in total
	// since automatic cleanup runs in background
	assert.Equal(t, 1, stats["active"])
	assert.True(t, stats["total"] >= 1) // At least the active one
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache(1 * time.Hour)

	// Test concurrent reads and writes
	done := make(chan bool, 2)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Set("key", i)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Get("key")
		}
		done <- true
	}()

	// Wait for completion
	<-done
	<-done

	// Should not panic and should have some value
	value, found := cache.Get("key")
	assert.True(t, found)
	assert.NotNil(t, value)
}
