package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cache := New[string, string](time.Minute)
	assert.NotNil(t, cache)
	assert.Equal(t, 0, cache.Len())
}

func TestCache_SetAndGet(t *testing.T) {
	cache := New[string, string](time.Minute)

	// Test setting and getting a value
	cache.Set("key1", "value1")
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Test getting non-existent key
	value, found = cache.Get("key2")
	assert.False(t, found)
	assert.Equal(t, "", value)
}

func TestCache_SetWithTTL(t *testing.T) {
	cache := New[string, string](time.Minute)

	// Set with short TTL
	cache.SetWithTTL("key1", "value1", 100*time.Millisecond)

	// Should be available immediately
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	value, found = cache.Get("key1")
	assert.False(t, found)
	assert.Equal(t, "", value)
}

func TestCache_Expiration(t *testing.T) {
	cache := New[string, string](100 * time.Millisecond)

	// Set value with default TTL
	cache.Set("key1", "value1")

	// Should be available immediately
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	value, found = cache.Get("key1")
	assert.False(t, found)
}

func TestCache_Delete(t *testing.T) {
	cache := New[string, string](time.Minute)

	cache.Set("key1", "value1")
	cache.Delete("key1")

	value, found := cache.Get("key1")
	assert.False(t, found)
	assert.Equal(t, "", value)
}

func TestCache_Clear(t *testing.T) {
	cache := New[string, string](time.Minute)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	assert.Equal(t, 2, cache.Len())

	cache.Clear()
	assert.Equal(t, 0, cache.Len())

	_, found := cache.Get("key1")
	assert.False(t, found)
}

func TestCache_Len(t *testing.T) {
	cache := New[string, string](time.Minute)

	assert.Equal(t, 0, cache.Len())

	cache.Set("key1", "value1")
	assert.Equal(t, 1, cache.Len())

	cache.Set("key2", "value2")
	assert.Equal(t, 2, cache.Len())
}

func TestCache_Keys(t *testing.T) {
	cache := New[string, string](time.Minute)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	keys := cache.Keys()
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
}

func TestCache_Cleanup(t *testing.T) {
	cache := New[string, string](100 * time.Millisecond)

	cache.Set("key1", "value1")
	cache.SetWithTTL("key2", "value2", time.Minute)

	// Wait for key1 to expire
	time.Sleep(150 * time.Millisecond)

	// Both keys should still be in the map
	assert.Equal(t, 2, cache.Len())

	// Cleanup should remove key1
	cleaned := cache.Cleanup()
	assert.Equal(t, 1, cleaned)
	assert.Equal(t, 1, cache.Len())

	// key2 should still be accessible
	value, found := cache.Get("key2")
	assert.True(t, found)
	assert.Equal(t, "value2", value)
}

func TestCache_GetOrSet(t *testing.T) {
	cache := New[string, string](time.Minute)

	callCount := 0
	factory := func() (string, error) {
		callCount++
		return "computed", nil
	}

	// First call should compute
	value, err := cache.GetOrSet("key1", factory)
	assert.NoError(t, err)
	assert.Equal(t, "computed", value)
	assert.Equal(t, 1, callCount)

	// Second call should use cache
	value, err = cache.GetOrSet("key1", factory)
	assert.NoError(t, err)
	assert.Equal(t, "computed", value)
	assert.Equal(t, 1, callCount) // Factory not called again
}

func TestCache_GetOrSet_Error(t *testing.T) {
	cache := New[string, string](time.Minute)

	factory := func() (string, error) {
		return "", errors.New("factory error")
	}

	value, err := cache.GetOrSet("key1", factory)
	assert.Error(t, err)
	assert.Equal(t, "", value)

	// Key should not be cached on error
	_, found := cache.Get("key1")
	assert.False(t, found)
}

func TestCache_GetOrSetWithTTL(t *testing.T) {
	cache := New[string, string](time.Minute)

	factory := func() (string, error) {
		return "computed", nil
	}

	// Set with short TTL
	value, err := cache.GetOrSetWithTTL("key1", 100*time.Millisecond, factory)
	assert.NoError(t, err)
	assert.Equal(t, "computed", value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found := cache.Get("key1")
	assert.False(t, found)
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache := New[string, int](time.Minute)

	// Concurrent writes
	for i := 0; i < 100; i++ {
		go func(i int) {
			cache.Set("key", i)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		go func() {
			cache.Get("key")
		}()
	}

	// Give goroutines time to complete
	time.Sleep(100 * time.Millisecond)

	// Cache should still be usable
	_, found := cache.Get("key")
	assert.True(t, found)
}

func TestCache_IntKey(t *testing.T) {
	cache := New[int, string](time.Minute)

	cache.Set(1, "value1")
	cache.Set(2, "value2")

	value, found := cache.Get(1)
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	value, found = cache.Get(2)
	assert.True(t, found)
	assert.Equal(t, "value2", value)
}

func TestCache_StructValue(t *testing.T) {
	type User struct {
		Name  string
		Email string
	}

	cache := New[string, User](time.Minute)

	user := User{Name: "John", Email: "john@example.com"}
	cache.Set("user1", user)

	retrieved, found := cache.Get("user1")
	assert.True(t, found)
	assert.Equal(t, user, retrieved)
}

func TestCache_AutoCleanup(t *testing.T) {
	cache := New[string, string](50 * time.Millisecond)

	// Start auto cleanup every 100ms
	stop := cache.AutoCleanup(100 * time.Millisecond)
	defer stop()

	cache.Set("key1", "value1")
	assert.Equal(t, 1, cache.Len())

	// Wait for expiration and cleanup
	time.Sleep(200 * time.Millisecond)

	// key1 should be cleaned up
	assert.Equal(t, 0, cache.Len())
}

func BenchmarkCache_Get(b *testing.B) {
	cache := New[string, string](time.Minute)
	cache.Set("key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

func BenchmarkCache_Set(b *testing.B) {
	cache := New[string, string](time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", "value")
	}
}

func BenchmarkCache_ConcurrentGet(b *testing.B) {
	cache := New[string, string](time.Minute)
	cache.Set("key", "value")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get("key")
		}
	})
}
