package api

import (
	"fmt"
	"net/http"
)

// ProjectCreateRequest represents the request body for creating a project
type ProjectCreateRequest struct {
	SystemCodes      string `json:"systemCodes"`
	ApplicationCodes string `json:"applicationCodes"`
	Name             string `json:"name"`
	Path             string `json:"path"`
	Language         string `json:"language"`
	ReadMe           bool   `json:"readMe"`
	Gitignore        bool   `json:"gitignore"`
	Note             string `json:"note"`
	Visibility       string `json:"visibility"`
	CodeGroupId      *int   `json:"codeGroupId,omitempty"`
	SpaceCodes       string `json:"spaceCodes"`
	ProjectCodes     string `json:"projectCodes"`
	IsPrivate        bool   `json:"isPrivate"`
}

// ProjectService provides methods for project management
type ProjectService struct {
	BaseService
	Config interface{}
}

// NewProjectService creates a new ProjectService
func NewProjectService(baseURL string, headers map[string]string, client HTTPClient, cfg interface{}) *ProjectService {
	return &ProjectService{
		BaseService: NewBaseService(baseURL, headers, client),
		Config:      cfg,
	}
}

// Create creates a new project
func (p *ProjectService) Create(req *ProjectCreateRequest) (*http.Response, error) {
	return p.Post("/projects/create", req)
}

// Delete deletes a project by its ID
func (p *ProjectService) Delete(projectID int) (*http.Response, error) {
	return p.BaseService.Delete(fmt.Sprintf("/projects/%d", projectID))
}

// ListUserRepos lists repositories for a user in a namespace/space
func (p *ProjectService) ListUserRepos(spaceCode string, pageSize, pageNo int) (*http.Response, error) {
	path := fmt.Sprintf("/projects/namespace/%s/user/page?pageSize=%d&pageNo=%d&contributedType=all", spaceCode, pageSize, pageNo)
	return p.Get(path)
}

// SearchUserRepos searches repositories by name in a namespace/space
func (p *ProjectService) SearchUserRepos(spaceCode, keyword string, pageSize, pageNo int) (*http.Response, error) {
	path := fmt.Sprintf("/projects/namespace/%s/user/page?pageSize=%d&pageNo=%d&contributedType=all&name=%s", spaceCode, pageSize, pageNo, keyword)
	return p.Get(path)
}

// SearchAllUserRepos searches repositories globally without workspace restriction
func (p *ProjectService) SearchAllUserRepos(keyword string, pageSize, pageNo int) (*http.Response, error) {
	path := fmt.Sprintf("/projects/user/page?pageSize=%d&pageNo=%d&contributedType=all&name=%s", pageSize, pageNo, keyword)
	return p.Get(path)
}

// ============================================================================
// Project Management API (from cmdevops-project)
// ============================================================================

// ProjectBaseInfo represents a project in the project management system
type ProjectBaseInfo struct {
	ID                     int    `json:"id"`
	ProjectCode            string `json:"projectCode"`
	ParentProjectCode      string `json:"parentProjectCode"`
	ProjectName            string `json:"projectName"`
	ProjectSource          string `json:"projectSource"`
	ProjectCategory        int    `json:"projectCategory"`
	ProjectStatus          int    `json:"projectStatus"`
	ProjectManagerName     string `json:"projectManagerName"`
	DeptName               string `json:"deptName"`
	ActualApprovalTime     string `json:"actualApprovalTime"`
	PlanFinishTime         string `json:"planFinishTime"`
	CreateName             string `json:"createName"`
	CompanyName            string `json:"companyName"`
	Deleted                bool   `json:"deleted"`
	CreateTime             string `json:"createTime"`
	ProjectManagerUid      string `json:"projectManagerUid"`
	ProjectMemberFlag      bool   `json:"projectMemberFlag"`
	SelfProjectType        int    `json:"selfProjectType"`
	ProjectExpenditureType string `json:"projectExpenditureType"`
	ManualModified         bool   `json:"manualModified"`
}

// ProjectBaseListResponse represents the API response for project list
type ProjectBaseListResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Data      []ProjectBaseInfo `json:"data"`
	PageNo    int               `json:"pageNo"`
	PageSize  int               `json:"pageSize"`
	Count     int               `json:"count"`
	PageCount int               `json:"pageCount"`
}

// ProjectBaseService handles project management operations
type ProjectBaseService struct {
	BaseService
}

// NewProjectBaseService creates a new ProjectBaseService
func NewProjectBaseService(baseURL string, headers map[string]string, client HTTPClient) *ProjectBaseService {
	return &ProjectBaseService{
		BaseService: NewBaseService(baseURL, headers, client),
	}
}

// ListProjects retrieves the project list with pagination
func (p *ProjectBaseService) ListProjects(pageNo, pageSize int) (*ProjectBaseListResponse, error) {
	path := fmt.Sprintf("/projectBaseInfo/page?pageNo=%d&pageSize=%d", pageNo, pageSize)

	resp, err := p.Get(path)
	if err != nil {
		return nil, err
	}

	var result ProjectBaseListResponse
	if err := p.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
