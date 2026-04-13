package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Space represents a space/workspace item from the API
type Space struct {
	ID                int      `json:"id"`
	TemplateID        string   `json:"templateId"`
	SpaceCode         string   `json:"spaceCode"`
	SpaceName         string   `json:"spaceName"`
	SpaceDesc         string   `json:"spaceDesc"`
	SpaceOwnerID      string   `json:"spaceOwnerId"`
	SpaceOwnerName    string   `json:"spaceOwnerName"`
	SpaceOwnerAccount string   `json:"spaceOwnerAccount"`
	SpaceState        string   `json:"spaceState"`
	CreateName        string   `json:"createName"`
	CreateAccount     string   `json:"createAccount"`
	UpdateName        string   `json:"updateName"`
	UpdateAccount     string   `json:"updateAccount"`
	CreateTime        string   `json:"createTime"`
	UpdateTime        string   `json:"updateTime"`
	TenantID          string   `json:"tenantId"`
	FavouriteID       int      `json:"favouriteId"`
	TemplateName      string   `json:"templateName"`
	BelongOrgID       string   `json:"belongOrgId"`
	BelongOrgName     string   `json:"belongOrgName"`
	SpaceRoleNames    string   `json:"spaceRoleNames"`
	SpaceRoleIDs      string   `json:"spaceRoleIds"`
	JoinTime          string   `json:"joinTime"`
	CanBeDeleted      bool     `json:"canBeDeleted"`
	ProjectNames      []string `json:"projectNames"`
}

// SpaceListResponse represents the API response for space list
type SpaceListResponse struct {
	Success   bool    `json:"success"`
	Code      string  `json:"code"`
	Message   string  `json:"message"`
	Data      []Space `json:"data"`
	PageNo    int     `json:"pageNo"`
	PageSize  int     `json:"pageSize"`
	Count     int     `json:"count"`
	PageCount int     `json:"pageCount"`
}

// SpaceService provides methods for space management
type SpaceService struct {
	BaseService
}

// NewSpaceService creates a new SpaceService
func NewSpaceService(baseURL string, headers map[string]string, client HTTPClient) *SpaceService {
	return &SpaceService{
		BaseService: NewBaseService(baseURL, headers, client),
	}
}

// List retrieves the list of spaces
func (s *SpaceService) List(pageNo, pageSize int) (*http.Response, error) {
	path := fmt.Sprintf("/moss/web/cmdevops-platform/space/api/space/v1/maintain/page?pageNo=%d&pageSize=%d",
		pageNo, pageSize)
	return s.Get(path)
}

// GetSpaceDetail retrieves detailed information for a workspace by its key (spaceCode)
func (s *SpaceService) GetSpaceDetail(workspaceKey string) (*http.Response, error) {
	// 使用空间列表接口，通过 spaceCode 过滤
	path := fmt.Sprintf("/moss/web/cmdevops-platform/space/api/space/v1/maintain/page?pageNo=1&pageSize=100&spaceCode=%s", workspaceKey)
	return s.Get(path)
}

// GetWorkspaceObjectId retrieves the objectId for a workspace by its key (spaceCode)
func (s *SpaceService) GetWorkspaceObjectId(workspaceKey string) (string, error) {
	url := fmt.Sprintf("%s/moss/web/cmdevops-req/api/team/parse/classes/Workspace?where=%%7B%%22key%%22%%3A%%22%s%%22%%7D&limit=1",
		s.BaseURL, workspaceKey)

	resp, err := s.GetWithQuery(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Results []struct {
			ObjectId string `json:"objectId"`
			Key      string `json:"key"`
			Name     string `json:"name"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(result.Results) == 0 {
		return "", fmt.Errorf("workspace not found for key: %s", workspaceKey)
	}

	return result.Results[0].ObjectId, nil
}

// GetSpaceNameByCode retrieves the spaceName for a workspace by its spaceCode (workspaceKey)
func (s *SpaceService) GetSpaceNameByCode(spaceCode string) (string, error) {
	resp, err := s.GetSpaceDetail(spaceCode)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var result struct {
		Success bool `json:"success"`
		Data    []struct {
			SpaceCode string `json:"spaceCode"`
			SpaceName string `json:"spaceName"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("failed to get space details: %s", string(body))
	}

	for _, space := range result.Data {
		if space.SpaceCode == spaceCode {
			return space.SpaceName, nil
		}
	}

	return "", fmt.Errorf("space not found for code: %s", spaceCode)
}

// SpaceProject represents a project associated with a space
type SpaceProject struct {
	ID                   int    `json:"id"`
	ProjectID            *int   `json:"projectId"`
	SpaceID              int    `json:"spaceId"`
	Deleted              bool   `json:"deleted"`
	CreateTime           string `json:"createTime"`
	CreateUid            string `json:"createUid"`
	CreateName           string `json:"createName"`
	CreateAccount        string `json:"createAccount"`
	UpdateTime           string `json:"updateTime"`
	UpdateUid            string `json:"updateUid"`
	UpdateName           string `json:"updateName"`
	UpdateAccount        string `json:"updateAccount"`
	DeleteTime           string `json:"deleteTime"`
	DeleteUid            string `json:"deleteUid"`
	DeleteName           string `json:"deleteName"`
	DeleteAccount        string `json:"deleteAccount"`
	TenantID             string `json:"tenantId"`
	ProjectCode          string `json:"projectCode"`
	ProjectName          string `json:"projectName"`
	ProjectManagerName   string `json:"projectManagerName"`
	ProjectStatus        int    `json:"projectStatus"`
	ProjectSource        string `json:"projectSource"`
	ParentName           string `json:"parentName"`
	DeptName             string `json:"deptName"`
	ProjectCategoryName  string `json:"projectCategoryName"`
}

// SpaceProjectListResponse represents the API response for space project list
type SpaceProjectListResponse struct {
	Success   bool           `json:"success"`
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Data      []SpaceProject `json:"data"`
	PageNo    int            `json:"pageNo"`
	PageSize  int            `json:"pageSize"`
	Count     int            `json:"count"`
	PageCount int            `json:"pageCount"`
}

// ListProjects retrieves the list of projects associated with a space
func (s *SpaceService) ListProjects(spaceID int, projectName string, pageNo, pageSize int) (*SpaceProjectListResponse, error) {
	path := fmt.Sprintf("/moss/web/cmdevops-platform/space/api/space/v1/maintain/space/project/page?spaceId=%d&projectName=%s&pageNo=%d&pageSize=%d",
		spaceID, projectName, pageNo, pageSize)

	resp, err := s.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result SpaceProjectListResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
