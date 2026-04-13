package domain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/config"
)

// RequirementService provides requirement-related business logic
type RequirementService struct {
	client     *api.Client
	config     *config.Config
	headers    map[string]string
	baseURL    string
	workspaceService WorkspaceService
}

// WorkspaceService interface for workspace operations
type WorkspaceService interface {
	GetWorkspaceObjectID(workspaceKey string) (string, error)
}

// NewRequirementService creates a new requirement service
func NewRequirementService(baseURL string, headers map[string]string, client *api.Client, cfg *config.Config, ws WorkspaceService) *RequirementService {
	return &RequirementService{
		client:           client,
		config:           cfg,
		headers:          headers,
		baseURL:          baseURL,
		workspaceService: ws,
	}
}

// CreateRequirementRequest represents a requirement creation request
type CreateRequirementRequest struct {
	Name            string
	Description     string
	WorkspaceKey    string
	WorkspaceName   string
	ProjectCode     string
	Priority        string
	Assignee        string
	DueDate         *time.Time
}

// RequirementResponse represents a simplified requirement response
type RequirementResponse struct {
	ObjectID   string `json:"objectId"`
	Name       string `json:"name"`
	Key        string `json:"key"`
	Status     string `json:"status"`
	Assignee   string `json:"assignee"`
	CreateTime string `json:"createTime"`
}

// Create creates a new requirement
func (s *RequirementService) Create(req *CreateRequirementRequest) (*RequirementResponse, error) {
	// Resolve workspace
	workspaceObjectID, err := s.workspaceService.GetWorkspaceObjectID(req.WorkspaceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace: %w", err)
	}

	// Build request body
	body := s.buildCreateBody(req, workspaceObjectID)

	// Send request
	url := fmt.Sprintf("%s/parse/classes/Item", s.baseURL)
	apiReq := &api.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: s.headers,
		Body:    body,
	}

	resp, err := s.client.Send(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create requirement: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create requirement: status %d", resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ObjectID string `json:"objectId"`
			Name     string `json:"name"`
			Key      string `json:"key"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to create requirement")
	}

	return &RequirementResponse{
		ObjectID: result.Data.ObjectID,
		Name:     result.Data.Name,
		Key:      result.Data.Key,
		Status:   "新建",
	}, nil
}

// buildCreateBody builds the request body for requirement creation
func (s *RequirementService) buildCreateBody(req *CreateRequirementRequest, workspaceObjectID string) map[string]interface{} {
	now := time.Now()
	dueDate := now.AddDate(0, 0, 7) // Default: 7 days from now
	if req.DueDate != nil {
		dueDate = *req.DueDate
	}

	// Default assignee
	assignee := req.Assignee
	if assignee == "" {
		assignee = s.config.User.Username
	}

	// Default priority
	priority := req.Priority
	if priority == "" {
		priority = "中"
	}

	body := map[string]interface{}{
		"name":             req.Name,
		"itemType":         "需求",
		"spaceId":          workspaceObjectID,
		"projectCode":      req.ProjectCode,
		"priority":         priority,
		"assignedTo":       assignee,
		"plannedEndDate":   map[string]interface{}{"iso": dueDate.Format(time.RFC3339)},
	}

	// Add description if provided
	if req.Description != "" {
		body["description"] = s.textToEditorContent(req.Description)
	}

	return body
}

// textToEditorContent converts plain text to editor content format
func (s *RequirementService) textToEditorContent(text string) map[string]interface{} {
	return map[string]interface{}{
		"type": "doc",
		"content": []map[string]interface{}{
			{
				"type": "paragraph",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": text,
					},
				},
			},
		},
	}
}

// Delete deletes requirements by their object IDs
func (s *RequirementService) Delete(objectIDs []string) error {
	if len(objectIDs) == 0 {
		return fmt.Errorf("no requirements to delete")
	}

	// Build batch delete request
	requests := make([]map[string]interface{}, 0, len(objectIDs))
	for _, id := range objectIDs {
		requests = append(requests, map[string]interface{}{
			"method": "DELETE",
			"path":   fmt.Sprintf("/parse/classes/Item/%s", id),
		})
	}

	body := map[string]interface{}{
		"requests": requests,
	}

	url := fmt.Sprintf("%s/parse/batch", s.baseURL)
	apiReq := &api.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: s.headers,
		Body:    body,
	}

	resp, err := s.client.Send(apiReq)
	if err != nil {
		return fmt.Errorf("failed to delete requirements: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete requirements: status %d", resp.StatusCode)
	}

	return nil
}

// ListRequirementsRequest represents a list requirements request
type ListRequirementsRequest struct {
	WorkspaceName string
	Keyword       string
	Status        string
	Assignee      string
	PageNo        int
	PageSize      int
}

// ListRequirementsResponse represents a paginated list response
type ListRequirementsResponse struct {
	Items      []map[string]interface{} `json:"items"`
	TotalCount int                      `json:"totalCount"`
	PageNo     int                      `json:"pageNo"`
	PageSize   int                      `json:"pageSize"`
}

// List lists requirements with filters
func (s *RequirementService) List(req *ListRequirementsRequest) (*ListRequirementsResponse, error) {
	// Build IQL query
	iql := s.buildListIQL(req)

	body := map[string]interface{}{
		"iql":      iql,
		"pageNo":   req.PageNo,
		"pageSize": req.PageSize,
	}

	url := fmt.Sprintf("%s/parse/functions/queryItemList", s.baseURL)
	apiReq := &api.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: s.headers,
		Body:    body,
	}

	resp, err := s.client.Send(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list requirements: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Items      []map[string]interface{} `json:"items"`
			TotalCount int                      `json:"totalCount"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ListRequirementsResponse{
		Items:      result.Data.Items,
		TotalCount: result.Data.TotalCount,
		PageNo:     req.PageNo,
		PageSize:   req.PageSize,
	}, nil
}

// buildListIQL builds IQL query for listing requirements
func (s *RequirementService) buildListIQL(req *ListRequirementsRequest) string {
	// This is a simplified version; use IQLService for complex queries
	iql := fmt.Sprintf("((所属空间 = '%s') and (类型 in [\"需求\"]))", req.WorkspaceName)

	if req.Keyword != "" {
		iql += fmt.Sprintf(" and (名称包含 '%s')", req.Keyword)
	}

	if req.Status != "" {
		iql += fmt.Sprintf(" and (状态 = '%s')", req.Status)
	}

	iql += " order by 创建时间 desc"
	return iql
}
