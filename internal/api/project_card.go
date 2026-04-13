package api

import (
	"fmt"
)

// ProjectCardService provides methods for managing project card settings
type ProjectCardService struct {
	BaseService
}

// ProjectCardUpdateRequest represents the request body for updating project card settings
type ProjectCardUpdateRequest struct {
	CardEnable bool `json:"cardEnable"`
}

// ProjectCardUpdateResponse represents the response from updating project card settings
type ProjectCardUpdateResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Id         int    `json:"id"`
		Name       string `json:"name"`
		CardEnable bool   `json:"cardEnable"`
	} `json:"data"`
}

// NewProjectCardService creates a new ProjectCardService
func NewProjectCardService(baseURL string, headers map[string]string, client HTTPClient) *ProjectCardService {
	return &ProjectCardService{
		BaseService: NewBaseService(baseURL, headers, client),
	}
}

// DisableCard disables the card (work item) association for a project
func (p *ProjectCardService) DisableCard(projectId int) (*ProjectCardUpdateResponse, error) {
	req := &ProjectCardUpdateRequest{
		CardEnable: false,
	}

	resp, err := p.Put(fmt.Sprintf("/projects/card/%d", projectId), req)
	if err != nil {
		return nil, err
	}

	var result ProjectCardUpdateResponse
	if err := p.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
