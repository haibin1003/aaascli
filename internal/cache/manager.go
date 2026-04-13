// Package cache provides application-specific caching for performance optimization
package cache

import (
	"time"
)

// Default TTL values
const (
	DefaultWorkspaceTTL   = 5 * time.Minute
	DefaultUserInfoTTL    = 10 * time.Minute
	DefaultConfigTTL      = 1 * time.Hour
	DefaultShortTTL       = 30 * time.Second
	DefaultLongTTL        = 1 * time.Hour
)

// Manager provides application-specific caching
// It wraps generic caches for different types of data
type Manager struct {
	// Workspace cache: workspaceKey -> workspaceObjectID
	workspaceCache *Cache[string, string]

	// User info cache: userID -> userInfo
	userCache *Cache[string, UserInfo]

	// Config cache: configKey -> configValue
	configCache *Cache[string, string]

	// API response cache: cacheKey -> responseData
	// Used for caching expensive API calls
	apiCache *Cache[string, []byte]
}

// UserInfo stores cached user information
type UserInfo struct {
	UserID      string
	Username    string
	Nickname    string
	Email       string
	TenantID    string
}

// NewManager creates a new cache manager with default settings
func NewManager() *Manager {
	return &Manager{
		workspaceCache: New[string, string](DefaultWorkspaceTTL),
		userCache:      New[string, UserInfo](DefaultUserInfoTTL),
		configCache:    New[string, string](DefaultConfigTTL),
		apiCache:       New[string, []byte](DefaultShortTTL),
	}
}

// GetWorkspace returns cached workspace object ID
func (m *Manager) GetWorkspace(workspaceKey string) (string, bool) {
	return m.workspaceCache.Get(workspaceKey)
}

// SetWorkspace caches workspace object ID
func (m *Manager) SetWorkspace(workspaceKey, objectID string) {
	m.workspaceCache.Set(workspaceKey, objectID)
}

// SetWorkspaceWithTTL caches workspace object ID with custom TTL
func (m *Manager) SetWorkspaceWithTTL(workspaceKey, objectID string, ttl time.Duration) {
	m.workspaceCache.SetWithTTL(workspaceKey, objectID, ttl)
}

// ClearWorkspaceCache clears all cached workspace data
func (m *Manager) ClearWorkspaceCache() {
	m.workspaceCache.Clear()
}

// GetUser returns cached user info
func (m *Manager) GetUser(userID string) (UserInfo, bool) {
	return m.userCache.Get(userID)
}

// SetUser caches user info
func (m *Manager) SetUser(userID string, info UserInfo) {
	m.userCache.Set(userID, info)
}

// ClearUserCache clears all cached user data
func (m *Manager) ClearUserCache() {
	m.userCache.Clear()
}

// GetConfig returns cached config value
func (m *Manager) GetConfig(key string) (string, bool) {
	return m.configCache.Get(key)
}

// SetConfig caches config value
func (m *Manager) SetConfig(key, value string) {
	m.configCache.Set(key, value)
}

// ClearConfigCache clears all cached config data
func (m *Manager) ClearConfigCache() {
	m.configCache.Clear()
}

// GetAPIResponse returns cached API response
func (m *Manager) GetAPIResponse(key string) ([]byte, bool) {
	return m.apiCache.Get(key)
}

// SetAPIResponse caches API response
func (m *Manager) SetAPIResponse(key string, data []byte) {
	m.apiCache.Set(key, data)
}

// SetAPIResponseWithTTL caches API response with custom TTL
func (m *Manager) SetAPIResponseWithTTL(key string, data []byte, ttl time.Duration) {
	m.apiCache.SetWithTTL(key, data, ttl)
}

// ClearAPICache clears all cached API responses
func (m *Manager) ClearAPICache() {
	m.apiCache.Clear()
}

// ClearAll clears all caches
func (m *Manager) ClearAll() {
	m.workspaceCache.Clear()
	m.userCache.Clear()
	m.configCache.Clear()
	m.apiCache.Clear()
}

// Stats returns cache statistics
func (m *Manager) Stats() map[string]int {
	return map[string]int{
		"workspace": m.workspaceCache.Len(),
		"user":      m.userCache.Len(),
		"config":    m.configCache.Len(),
		"api":       m.apiCache.Len(),
	}
}

// StartAutoCleanup starts background cleanup for all caches
// Returns a function to stop the cleanup
func (m *Manager) StartAutoCleanup(interval time.Duration) func() {
	stoppers := []func(){
		m.workspaceCache.AutoCleanup(interval),
		m.userCache.AutoCleanup(interval),
		m.configCache.AutoCleanup(interval),
		m.apiCache.AutoCleanup(interval),
	}

	return func() {
		for _, stop := range stoppers {
			stop()
		}
	}
}
