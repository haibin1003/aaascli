package api

import (
	"fmt"
)

// RepoGroup represents a repository group
type RepoGroup struct {
	ID                    int    `json:"id"`
	Name                  string `json:"name"`
	Path                  string `json:"path"`
	Description           string `json:"description"`
	FullName              string `json:"fullName"`
	FullPath              string `json:"fullPath"`
	TenantFullName        string `json:"tenantFullName"`
	TenantFullPath        string `json:"tenantFullPath"`
	PullRequestCount      *int   `json:"pullRequestCount"`
	IssueCount            *int   `json:"issueCount"`
	ChildExists           bool   `json:"childExists"`
	CreatorName           string `json:"creatorName"`
	CreatedBy             string `json:"createdBy"`
	GroupType             *string `json:"groupType"`
	CreatedAt             string `json:"createdAt"`
	CurrentUserPermission *string `json:"currentUserPermission"`
	PermissionGroup       *string `json:"permissionGroup"`
	PermissionRepo        *string `json:"permissionRepo"`
}

// RepoGroupListResponse represents the API response for repository group list
type RepoGroupListResponse struct {
	Success   bool        `json:"success"`
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      []RepoGroup `json:"data"`
	PageNo    int         `json:"pageNo"`
	PageSize  int         `json:"pageSize"`
	Count     int         `json:"count"`
	PageCount int         `json:"pageCount"`
	StartRow  int         `json:"startRow"`
	EndRow    int         `json:"endRow"`
}

// RepoGroupCreateRequest represents the request body for creating a repository group
type RepoGroupCreateRequest struct {
	Name        string `json:"name"`
	ParentID    string `json:"parentId"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

// RepoGroupCreateResponse represents the API response for creating a repository group
type RepoGroupCreateResponse struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// RepoGroupService provides methods for repository group management
type RepoGroupService struct {
	BaseService
}

// NewRepoGroupService creates a new RepoGroupService
func NewRepoGroupService(baseURL string, headers map[string]string, client HTTPClient) *RepoGroupService {
	return &RepoGroupService{
		BaseService: NewBaseService(baseURL, headers, client),
	}
}

// ListSubGroups retrieves the list of repository groups with pagination
func (s *RepoGroupService) ListSubGroups(pageNo, pageSize int) (*RepoGroupListResponse, error) {
	path := fmt.Sprintf("/code/groups/sub/page?pageNo=%d&pageSize=%d", pageNo, pageSize)

	resp, err := s.Get(path)
	if err != nil {
		return nil, err
	}

	var result RepoGroupListResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Create creates a new repository group
func (s *RepoGroupService) Create(req *RepoGroupCreateRequest) (*RepoGroupCreateResponse, error) {
	resp, err := s.Post("/code/groups", req)
	if err != nil {
		return nil, err
	}

	var result RepoGroupCreateResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PrivateGroupResponse represents the API response for getting private group
type PrivateGroupResponse struct {
	Success bool      `json:"success"`
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Data    RepoGroup `json:"data"`
}

// GetPrivateGroup retrieves the current user's personal (private) repository group
func (s *RepoGroupService) GetPrivateGroup() (*PrivateGroupResponse, error) {
	resp, err := s.Get("/code/groups/private")
	if err != nil {
		return nil, err
	}

	var result PrivateGroupResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
