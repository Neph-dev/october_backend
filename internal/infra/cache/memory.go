package cache

import (
	"sync"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/ai"
)

// MemoryCache implements an in-memory cache for article summaries
type MemoryCache struct {
	mu    sync.RWMutex
	cache map[string]*ai.CachedSummary
}

// NewMemoryCache creates a new in-memory cache instance
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		cache: make(map[string]*ai.CachedSummary),
	}
	
	// Start a cleanup goroutine to remove expired entries
	go cache.cleanup()
	
	return cache
}

// Get retrieves a cached summary by article ID
func (m *MemoryCache) Get(articleID string) (*ai.CachedSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	cached, exists := m.cache[articleID]
	if !exists {
		return nil, nil // Cache miss
	}
	
	// Check if the cache entry has expired
	if time.Now().After(cached.ExpiresAt) {
		// Remove expired entry
		m.mu.RUnlock()
		m.mu.Lock()
		delete(m.cache, articleID)
		m.mu.Unlock()
		m.mu.RLock()
		return nil, nil // Cache miss due to expiration
	}
	
	return cached, nil
}

// Set stores a summary in the cache with the specified TTL
func (m *MemoryCache) Set(articleID string, summary *ai.ArticleSummaryResponse, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	cached := &ai.CachedSummary{
		ArticleID:     summary.ArticleID,
		OriginalTitle: summary.OriginalTitle,
		Summary:       summary.Summary,
		SourceURL:     summary.SourceURL,
		CachedAt:      now,
		ExpiresAt:     now.Add(ttl),
	}
	
	m.cache[articleID] = cached
	return nil
}

// Delete removes a cached summary by article ID
func (m *MemoryCache) Delete(articleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.cache, articleID)
	return nil
}

// Clear removes all cached summaries
func (m *MemoryCache) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.cache = make(map[string]*ai.CachedSummary)
	return nil
}

// cleanup runs periodically to remove expired cache entries
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.cleanupExpired()
		}
	}
}

// cleanupExpired removes expired entries from the cache
func (m *MemoryCache) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	for articleID, cached := range m.cache {
		if now.After(cached.ExpiresAt) {
			delete(m.cache, articleID)
		}
	}
}

// GetCacheStats returns statistics about the cache
func (m *MemoryCache) GetCacheStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	totalEntries := len(m.cache)
	expiredEntries := 0
	now := time.Now()
	
	for _, cached := range m.cache {
		if now.After(cached.ExpiresAt) {
			expiredEntries++
		}
	}
	
	return map[string]interface{}{
		"total_entries":   totalEntries,
		"expired_entries": expiredEntries,
		"active_entries":  totalEntries - expiredEntries,
	}
}