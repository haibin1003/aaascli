// Package cache provides a generic caching layer for performance optimization
package cache

import (
	"sync"
	"time"
)

// Cache is a generic thread-safe cache with TTL support
type Cache[K comparable, V any] struct {
	mu       sync.RWMutex
	items    map[K]cacheItem[V]
	defaultTTL time.Duration
}

// cacheItem stores a cached value with its expiration time
type cacheItem[V any] struct {
	value      V
	expiresAt  time.Time
}

// New creates a new cache with the specified default TTL
func New[K comparable, V any](defaultTTL time.Duration) *Cache[K, V] {
	return &Cache[K, V]{
		items:      make(map[K]cacheItem[V]),
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value from the cache
// Returns the value and true if found and not expired, zero value and false otherwise
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		var zero V
		return zero, false
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		c.Delete(key)
		var zero V
		return zero, false
	}

	return item.value, true
}

// Set stores a value in the cache with the default TTL
func (c *Cache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a specific TTL
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem[V]{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a key from the cache
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]cacheItem[V])
}

// Len returns the number of items in the cache (including expired items)
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Cleanup removes all expired items from the cache
func (c *Cache[K, V]) Cleanup() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	count := 0
	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
			count++
		}
	}
	return count
}

// Keys returns all keys in the cache (including expired items)
func (c *Cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// GetOrSet retrieves a value from the cache, or computes and stores it if not present
func (c *Cache[K, V]) GetOrSet(key K, factory func() (V, error)) (V, error) {
	// Try to get from cache
	if value, found := c.Get(key); found {
		return value, nil
	}

	// Not found, compute the value
	value, err := factory()
	if err != nil {
		var zero V
		return zero, err
	}

	// Store in cache
	c.Set(key, value)
	return value, nil
}

// GetOrSetWithTTL retrieves a value from the cache, or computes and stores it with specific TTL if not present
func (c *Cache[K, V]) GetOrSetWithTTL(key K, ttl time.Duration, factory func() (V, error)) (V, error) {
	// Try to get from cache
	if value, found := c.Get(key); found {
		return value, nil
	}

	// Not found, compute the value
	value, err := factory()
	if err != nil {
		var zero V
		return zero, err
	}

	// Store in cache with custom TTL
	c.SetWithTTL(key, value, ttl)
	return value, nil
}

// AutoCleanup starts a background goroutine that periodically cleans up expired items
// Returns a function that can be called to stop the cleanup goroutine
func (c *Cache[K, V]) AutoCleanup(interval time.Duration) func() {
	ticker := time.NewTicker(interval)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				c.Cleanup()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() {
		close(done)
	}
}
