package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/cache"
	"github.com/user/lc/internal/config"
)

// WorkspaceService provides workspace-related operations
type WorkspaceService struct {
	client  *api.Client
	config  *config.Config
	headers map[string]string
	baseURL string
	cache   *cache.Manager
}

// NewWorkspaceService creates a new workspace service
func NewWorkspaceService(baseURL string, headers map[string]string, client *api.Client, cfg *config.Config) *WorkspaceService {
	return NewWorkspaceServiceWithCache(baseURL, headers, client, cfg, cache.NewManager())
}

// NewWorkspaceServiceWithCache creates a new workspace service with a cache manager
func NewWorkspaceServiceWithCache(baseURL string, headers map[string]string, client *api.Client, cfg *config.Config, cacheManager *cache.Manager) *WorkspaceService {
	return &WorkspaceService{
		client:  client,
		config:  cfg,
		headers: headers,
		baseURL: baseURL,
		cache:   cacheManager,
	}
}

// Workspace represents a workspace/workspace
type Workspace struct {
	ObjectID    string `json:"objectId"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description,omitempty"`
}

// GetWorkspaceObjectID resolves workspace key to object ID
// Uses caching to avoid repeated API calls
func (s *WorkspaceService) GetWorkspaceObjectID(workspaceKey string) (string, error) {
	// Check cache first
	if objectID, found := s.cache.GetWorkspace(workspaceKey); found {
		return objectID, nil
	}

	// Build request
	url := fmt.Sprintf("%s/space/api/v1/workspaces/%s", s.baseURL, workspaceKey)
	req := &api.Request{
		Method:  http.MethodGet,
		URL:     url,
		Headers: s.headers,
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return "", fmt.Errorf("failed to get workspace info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("workspace not found: %s", workspaceKey)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ObjectID string `json:"objectId"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode workspace response: %w", err)
	}

	if !result.Success || result.Data.ObjectID == "" {
		return "", fmt.Errorf("workspace not found: %s", workspaceKey)
	}

	// Cache the result
	s.cache.SetWorkspace(workspaceKey, result.Data.ObjectID)

	return result.Data.ObjectID, nil
}

// ClearCache clears the workspace object ID cache
func (s *WorkspaceService) ClearCache() {
	s.cache.ClearWorkspaceCache()
}

// ListWorkspaces lists all accessible workspaces
func (s *WorkspaceService) ListWorkspaces() ([]Workspace, error) {
	url := fmt.Sprintf("%s/space/api/v1/workspaces", s.baseURL)
	req := &api.Request{
		Method:  http.MethodGet,
		URL:     url,
		Headers: s.headers,
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list workspaces: status %d", resp.StatusCode)
	}

	var result struct {
		Success bool        `json:"success"`
		Data    []Workspace `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode workspaces response: %w", err)
	}

	return result.Data, nil
}

// WorkspaceResolver provides a simplified interface for workspace resolution
type WorkspaceResolver struct {
	service *WorkspaceService
}

// NewWorkspaceResolver creates a new workspace resolver
func NewWorkspaceResolver(service *WorkspaceService) *WorkspaceResolver {
	return &WorkspaceResolver{service: service}
}

// Resolve resolves workspace information from key and name
// Returns workspace object ID and validated name
func (r *WorkspaceResolver) Resolve(key, name string) (objectID string, workspaceName string, err error) {
	if key == "" {
		return "", "", fmt.Errorf("workspace key is required")
	}

	objectID, err = r.service.GetWorkspaceObjectID(key)
	if err != nil {
		return "", "", err
	}

	// If name not provided, use key as fallback
	if name == "" {
		name = key
	}

	return objectID, name, nil
}
