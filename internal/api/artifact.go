package api

import (
	"fmt"
	"net/url"
)

// ArtifactRepoGroup represents an artifact repository group from the API
type ArtifactRepoGroup struct {
	ID            int    `json:"id"`            // 原始ID字段
	RepoGroupID   string `json:"repoGroupId"`   // 可能为空的组ID
	RepoGroupCode string `json:"repoGroupCode"`
	RepoGroupName string `json:"repoGroupName"`
}

// ArtifactRepoGroupListResponse represents the API response for artifact repository group list
type ArtifactRepoGroupListResponse struct {
	Success bool                `json:"success"`
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Data    []ArtifactRepoGroup `json:"data"`
}

// RepositoryInfoDTO represents the repository info for creation
type RepositoryInfoDTO struct {
	RepoKey             string   `json:"repoKey"`
	RepoNature          int      `json:"repoNature"`
	RepoType            string   `json:"repoType"`
	RepoEnvironments    string   `json:"repoEnvironments,omitempty"`
	RepoIncludesPattern string   `json:"repoIncludesPattern,omitempty"`
	RepoExcludesPattern string   `json:"repoExcludesPattern,omitempty"`
	AnonymousAccess     int      `json:"anonymousAccess,omitempty"`
	DescInfo            string   `json:"descInfo,omitempty"`
	TenantID            string   `json:"tenantId"`
	SystemCode          string   `json:"systemCode,omitempty"`
	SpaceCode           string   `json:"spaceCode"`
	DeputyAccountNumber string   `json:"deputyAccountNumber"`
	RelatedProjects     []string `json:"relatedProjects,omitempty"`
	CreateUid           string   `json:"createUid,omitempty"`
	UserName            string   `json:"userName,omitempty"`
	CreateUsername      string   `json:"createUsername,omitempty"`
	SnapshotType        string   `json:"snapshotType,omitempty"`
	MaxSnapshotCount    int      `json:"maxSnapshotCount,omitempty"`
	RepoGroupID         string   `json:"repoGroupId,omitempty"`
	RepoGroupCode       string   `json:"repoGroupCode,omitempty"`
	RepoGroupName       string   `json:"repoGroupName,omitempty"`
}

// CreateRepoRequest represents the request to create a repository
type CreateRepoRequest struct {
	RepositoryInfoDTO RepositoryInfoDTO `json:"repositoryInfoDTO"`
}

// CreateRepoResponse represents the API response for creating a repository
type CreateRepoResponse struct {
	Success bool                   `json:"success"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// ArtifactService provides methods for artifact repository management
type ArtifactService struct {
	BaseService
}

// NewArtifactService creates a new ArtifactService
func NewArtifactService(baseURL string, headers map[string]string, client HTTPClient) *ArtifactService {
	return &ArtifactService{
		BaseService: NewBaseService(baseURL, headers, client),
	}
}

// ListRepoGroups retrieves the list of repository groups
func (s *ArtifactService) ListRepoGroups(tenantID string) (*ArtifactRepoGroupListResponse, error) {
	path := fmt.Sprintf("/cmdevops-repo/repository/group/v1/list?tenantId=%s", tenantID)

	resp, err := s.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ArtifactRepoGroupListResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateRepository creates a new artifact repository
func (s *ArtifactService) CreateRepository(req *CreateRepoRequest) (*CreateRepoResponse, error) {
	path := "/cmdevops-repo/repository/createRepositoryInfo"

	resp, err := s.Post(path, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CreateRepoResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Repository represents a single artifact repository
type Repository struct {
	ID                           int         `json:"id"`
	RepoKey                      string      `json:"repoKey"`
	ShowRepoKey                  string      `json:"showRepoKey"`
	RepoType                     string      `json:"repoType"`
	RepoNature                   int         `json:"repoNature"`
	RepoEnvironment              interface{} `json:"repoEnvironment"`
	RepoEnvironments             string      `json:"repoEnvironments"`
	RepoGroupID                  int         `json:"repoGroupId"`
	RepoGroupCode                string      `json:"repoGroupCode"`
	RepoGroupName                string      `json:"repoGroupName"`
	RemoteRepoURL                interface{} `json:"remoteRepoUrl"`
	IsPreset                     int         `json:"isPreset"`
	ReadOnly                     bool        `json:"readOnly"`
	ShareWithAll                 bool        `json:"shareWithAll"`
	TenantID                     string      `json:"tenantId"`
	SpaceCode                    string      `json:"spaceCode"`
	CreateTime                   string      `json:"createTime"`
	CreateUid                    string      `json:"createUid"`
	CreateUsername               string      `json:"createUsername"`
	UpdateTime                   string      `json:"updateTime"`
	UpdateUid                    interface{} `json:"updateUid"`
	UpdateUsername               interface{} `json:"updateUsername"`
	Deleted                      bool        `json:"deleted"`
	UsedSize                     string      `json:"usedSize"`
}

// RepositoryListResponse represents the API response for repository list
type RepositoryListResponse struct {
	Success   bool         `json:"success"`
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Data      []Repository `json:"data"`
	PageNo    int          `json:"pageNo"`
	PageSize  int          `json:"pageSize"`
	Count     int          `json:"count"`
	PageCount int          `json:"pageCount"`
	StartRow  int          `json:"startRow"`
	EndRow    int          `json:"endRow"`
}
// GetRepositoryList retrieves the list of repositories
func (s *ArtifactService) GetRepositoryList(spaceCode, tenantID, deputyAccountNumber string, repoNature, pageNo, pageSize int) (*RepositoryListResponse, error) {
	path := fmt.Sprintf("/cmdevops-repo/repository/list/getRepositoryList?pageNo=%d&pageSize=%d&deputyAccountNumber=%s&tenantId=%s&repoNature=%d&spaceCode=%s",
		pageNo, pageSize, url.QueryEscape(deputyAccountNumber), tenantID, repoNature, spaceCode)

	resp, err := s.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result RepositoryListResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRepoTypeDescription returns the human-readable description for repository type
func GetRepoTypeDescription(repoType string) string {
	descriptions := map[string]string{
		"Maven":      "Maven 仓库",
		"Npm":        "NPM 仓库",
		"Pypi":       "PyPI 仓库",
		"Docker":     "Docker 仓库",
		"Debian":     "Debian 仓库",
		"Composer":   "Composer 仓库",
		"Rpm":        "RPM 仓库",
		"Go":         "Go 模块仓库",
		"Conan":      "Conan 仓库",
		"Nuget":      "NuGet 仓库",
		"Generic":    "通用仓库",
		"Cocoapods":  "CocoaPods 仓库",
		"Helm":       "Helm 仓库",
		"Cargo":      "Cargo 仓库",
	}

	if desc, ok := descriptions[repoType]; ok {
		return desc
	}
	return repoType
}

// GetRepoEnvironmentDescription returns the human-readable description for repository environment
func GetRepoEnvironmentDescription(env string) string {
	descriptions := map[string]string{
		"DEV":   "开发环境",
		"TEST":  "测试环境",
		"PROD":  "生产环境",
	}

	if desc, ok := descriptions[env]; ok {
		return desc
	}
	return env
}
