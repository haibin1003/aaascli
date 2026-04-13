package api

import (
	"encoding/json"
	"fmt"
)

// ProjectListResponse represents the response for listing projects
type ProjectListResponse struct {
	Success   bool          `json:"success"`
	Code      interface{}   `json:"code"`
	Message   interface{}   `json:"message"`
	Data      []ProjectInfo `json:"data"`
	PageNo    int           `json:"pageNo"`
	PageSize  int           `json:"pageSize"`
	Count     int           `json:"count"`
	PageCount int           `json:"pageCount"`
	StartRow  int           `json:"startRow"`
	EndRow    int           `json:"endRow"`
}

// ProjectInfo represents information about a project
type ProjectInfo struct {
	ID                          int           `json:"id"`
	Name                        string        `json:"name"`
	Path                        string        `json:"path"`
	Note                        string        `json:"note"`
	CodeGroupId                 int           `json:"codeGroupId"`
	CodeGroupName               string        `json:"codeGroupName"`
	CodeGroupPath               string        `json:"codeGroupPath"`
	CodeGroupFullPath           string        `json:"codeGroupFullPath"`
	CodeGroupTenantFullPath     string        `json:"codeGroupTenantFullPath"`
	CodeGroupFullName           string        `json:"codeGroupFullName"`
	CodeGroupTenantFullName     string        `json:"codeGroupTenantFullName"`
	GitProjectId                int           `json:"gitProjectId"`
	GitGroupId                  int           `json:"gitGroupId"`
	LastUpdated                 string        `json:"lastUpdated"`
	Language                    interface{}   `json:"language"`
	CreateTime                  string        `json:"createTime"`
	UpdateTime                  string        `json:"updateTime"`
	CreatedBy                   string        `json:"createdBy"`
	CreatorName                 string        `json:"creatorName"`
	UpdatedBy                   interface{}   `json:"updatedBy"`
	TenantId                    string        `json:"tenantId"`
	SystemName                  interface{}   `json:"systemName"`
	SystemCode                  interface{}   `json:"systemCode"`
	PrWorkspaceItemEnabled      interface{}   `json:"prWorkspaceItemEnabled"`
	Apps                        []interface{} `json:"apps"`
	Projects                    []interface{} `json:"projects"`
	SpaceName                   interface{}   `json:"spaceName"`
	SpaceCode                   string        `json:"spaceCode"`
	Visibility                  int           `json:"visibility"`
	EmptyStatus                 int           `json:"emptyStatus"`
	Front2giteeUrl              interface{}   `json:"front2giteeUrl"`
	IsFavorite                  bool          `json:"isFavorite"`
	Archived                    bool          `json:"archived"`
	ProjectTemplateId           interface{}   `json:"projectTemplateId"`
	CardEnable                  interface{}   `json:"cardEnable"`
	PathWithNameSpace           string        `json:"pathWithNameSpace"`
	ForkFromPath                string        `json:"forkFromPath"`
	ForkFromId                  int           `json:"forkFromId"`
	ForkTenantPathWithNameSpace interface{}   `json:"forkTenantPathWithNameSpace"`
	ForkCodeGroupTenantFullPath interface{}   `json:"forkCodeGroupTenantFullPath"`
	NameWithNameSpace           string        `json:"nameWithNameSpace"`
	TenantNameWithNameSpace     string        `json:"tenantNameWithNameSpace"`
	IsOpen                      bool          `json:"isOpen"`
	OpenSourceTag               interface{}   `json:"openSourceTag"`
	HttpPath                    string        `json:"httpPath"`
	SshPath                     string        `json:"sshPath"`
	TenantSshPath               string        `json:"tenantSshPath"`
	TenantHttpPath              string        `json:"tenantHttpPath"`
	Status                      int           `json:"status"`
	ProjectStatistics           interface{}   `json:"projectStatistics"`
}

// ListProjects lists projects with pagination
func (p *ProjectService) ListProjects(pageSize, pageNo int) (*ProjectListResponse, error) {
	// Get workspace key from headers (set by command line)
	workspaceKey := p.Headers["X-Auth-ModuleId"]

	// Construct query parameters
	url := p.BaseURL + "/projects/namespace/" + workspaceKey + "/user/page"

	// Add query parameters
	queryParams := fmt.Sprintf("?pageSize=%d&pageNo=%d&contributedType=all", pageSize, pageNo)

	apiReq := &Request{
		URL:     url + queryParams,
		Method:  "GET",
		Headers: p.Headers,
		Body:    nil,
	}

	resp, err := p.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var projectList ProjectListResponse
	if err := json.NewDecoder(resp.Body).Decode(&projectList); err != nil {
		return nil, err
	}

	return &projectList, nil
}
