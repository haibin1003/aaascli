package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	assert.NotNil(t, manager)

	stats := manager.Stats()
	assert.Equal(t, 0, stats["workspace"])
	assert.Equal(t, 0, stats["user"])
	assert.Equal(t, 0, stats["config"])
	assert.Equal(t, 0, stats["api"])
}

func TestManager_WorkspaceCache(t *testing.T) {
	manager := NewManager()

	// Test Set and Get
	manager.SetWorkspace("my-workspace", "object-id-123")

	objectID, found := manager.GetWorkspace("my-workspace")
	assert.True(t, found)
	assert.Equal(t, "object-id-123", objectID)

	// Test non-existent workspace
	_, found = manager.GetWorkspace("non-existent")
	assert.False(t, found)

	// Test Clear
	manager.ClearWorkspaceCache()
	_, found = manager.GetWorkspace("my-workspace")
	assert.False(t, found)
}

func TestManager_WorkspaceCacheWithTTL(t *testing.T) {
	manager := NewManager()

	// Set with short TTL
	manager.SetWorkspaceWithTTL("workspace", "id", 100*time.Millisecond)

	// Should be available
	id, found := manager.GetWorkspace("workspace")
	assert.True(t, found)
	assert.Equal(t, "id", id)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found = manager.GetWorkspace("workspace")
	assert.False(t, found)
}

func TestManager_UserCache(t *testing.T) {
	manager := NewManager()

	user := UserInfo{
		UserID:   "user-123",
		Username: "john",
		Nickname: "John Doe",
		Email:    "john@example.com",
		TenantID: "tenant-1",
	}

	manager.SetUser("user-123", user)

	retrieved, found := manager.GetUser("user-123")
	assert.True(t, found)
	assert.Equal(t, user, retrieved)

	manager.ClearUserCache()
	_, found = manager.GetUser("user-123")
	assert.False(t, found)
}

func TestManager_ConfigCache(t *testing.T) {
	manager := NewManager()

	manager.SetConfig("api.endpoint", "https://api.example.com")

	value, found := manager.GetConfig("api.endpoint")
	assert.True(t, found)
	assert.Equal(t, "https://api.example.com", value)

	manager.ClearConfigCache()
	_, found = manager.GetConfig("api.endpoint")
	assert.False(t, found)
}

func TestManager_APICache(t *testing.T) {
	manager := NewManager()

	data := []byte(`{"success": true, "data": []}`)
	manager.SetAPIResponse("/api/users", data)

	retrieved, found := manager.GetAPIResponse("/api/users")
	assert.True(t, found)
	assert.Equal(t, data, retrieved)

	// Test with custom TTL
	manager.SetAPIResponseWithTTL("/api/temp", []byte("temp"), 100*time.Millisecond)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	_, found = manager.GetAPIResponse("/api/temp")
	assert.False(t, found)

	manager.ClearAPICache()
	_, found = manager.GetAPIResponse("/api/users")
	assert.False(t, found)
}

func TestManager_ClearAll(t *testing.T) {
	manager := NewManager()

	manager.SetWorkspace("ws", "id")
	manager.SetUser("user", UserInfo{})
	manager.SetConfig("key", "value")
	manager.SetAPIResponse("/api", []byte("data"))

	stats := manager.Stats()
	assert.Equal(t, 1, stats["workspace"])
	assert.Equal(t, 1, stats["user"])
	assert.Equal(t, 1, stats["config"])
	assert.Equal(t, 1, stats["api"])

	manager.ClearAll()

	stats = manager.Stats()
	assert.Equal(t, 0, stats["workspace"])
	assert.Equal(t, 0, stats["user"])
	assert.Equal(t, 0, stats["config"])
	assert.Equal(t, 0, stats["api"])
}

func TestManager_Stats(t *testing.T) {
	manager := NewManager()

	manager.SetWorkspace("ws1", "id1")
	manager.SetWorkspace("ws2", "id2")
	manager.SetUser("user1", UserInfo{})

	stats := manager.Stats()
	assert.Equal(t, 2, stats["workspace"])
	assert.Equal(t, 1, stats["user"])
	assert.Equal(t, 0, stats["config"])
	assert.Equal(t, 0, stats["api"])
}

func TestManager_StartAutoCleanup(t *testing.T) {
	manager := NewManager()

	stop := manager.StartAutoCleanup(100 * time.Millisecond)
	defer stop()

	// Set with short TTL
	manager.SetAPIResponseWithTTL("/api/test", []byte("data"), 50*time.Millisecond)
	assert.Equal(t, 1, manager.Stats()["api"])

	// Wait for expiration and cleanup
	time.Sleep(200 * time.Millisecond)

	// Should be cleaned up
	assert.Equal(t, 0, manager.Stats()["api"])
}

func BenchmarkManager_GetWorkspace(b *testing.B) {
	manager := NewManager()
	manager.SetWorkspace("workspace", "object-id")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetWorkspace("workspace")
	}
}

func BenchmarkManager_SetWorkspace(b *testing.B) {
	manager := NewManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.SetWorkspace("workspace", "object-id")
	}
}
