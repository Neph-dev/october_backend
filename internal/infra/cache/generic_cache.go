package cache

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrCacheKeyNotFound is returned when a cache key is not found
	ErrCacheKeyNotFound = errors.New("cache key not found")
)

// GenericCache is a generic in-memory cache implementation for market data
type GenericCache struct {
	data    map[string]CacheItem
	mutex   sync.RWMutex
	stats   CacheStats
}

// CacheItem represents a cached item with TTL
type CacheItem struct {
	Value     []byte
	ExpiresAt time.Time
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64
	Misses      int64
	Sets        int64
	Deletes     int64
	ItemCount   int64
	LastCleared time.Time
}

// NewGenericCache creates a new generic cache instance
func NewGenericCache() *GenericCache {
	cache := &GenericCache{
		data: make(map[string]CacheItem),
		stats: CacheStats{
			LastCleared: time.Now(),
		},
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves a value from the cache
func (c *GenericCache) Get(key string) ([]byte, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	item, exists := c.data[key]
	if !exists {
		c.stats.Misses++
		return nil, ErrCacheKeyNotFound
	}
	
	// Check if item has expired
	if time.Now().After(item.ExpiresAt) {
		// Remove expired item
		delete(c.data, key)
		c.stats.Misses++
		c.stats.ItemCount--
		return nil, ErrCacheKeyNotFound
	}
	
	c.stats.Hits++
	return item.Value, nil
}

// Set stores a value in the cache with TTL
func (c *GenericCache) Set(key string, value []byte, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	expiresAt := time.Now().Add(ttl)
	if ttl <= 0 {
		// No expiration
		expiresAt = time.Now().Add(365 * 24 * time.Hour) // 1 year
	}
	
	// Check if key already exists
	_, exists := c.data[key]
	if !exists {
		c.stats.ItemCount++
	}
	
	c.data[key] = CacheItem{
		Value:     value,
		ExpiresAt: expiresAt,
	}
	
	c.stats.Sets++
	return nil
}

// Delete removes a value from the cache
func (c *GenericCache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if _, exists := c.data[key]; exists {
		delete(c.data, key)
		c.stats.Deletes++
		c.stats.ItemCount--
	}
	
	return nil
}

// Clear removes all values from the cache
func (c *GenericCache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.data = make(map[string]CacheItem)
	c.stats.ItemCount = 0
	c.stats.LastCleared = time.Now()
	
	return nil
}

// GetStats returns cache statistics
func (c *GenericCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	hitRate := float64(0)
	totalRequests := c.stats.Hits + c.stats.Misses
	if totalRequests > 0 {
		hitRate = float64(c.stats.Hits) / float64(totalRequests) * 100
	}
	
	return map[string]interface{}{
		"hits":          c.stats.Hits,
		"misses":        c.stats.Misses,
		"sets":          c.stats.Sets,
		"deletes":       c.stats.Deletes,
		"item_count":    c.stats.ItemCount,
		"hit_rate":      hitRate,
		"last_cleared":  c.stats.LastCleared,
	}
}

// cleanup removes expired items from the cache
func (c *GenericCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()
	
	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		
		for key, item := range c.data {
			if now.After(item.ExpiresAt) {
				delete(c.data, key)
				c.stats.ItemCount--
			}
		}
		
		c.mutex.Unlock()
	}
}