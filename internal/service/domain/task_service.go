package domain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/config"
)

// TaskService provides task-related business logic
type TaskService struct {
	client           *api.Client
	config           *config.Config
	headers          map[string]string
	baseURL          string
	workspaceService WorkspaceService
}

// NewTaskService creates a new task service
func NewTaskService(baseURL string, headers map[string]string, client *api.Client, cfg *config.Config, ws WorkspaceService) *TaskService {
	return &TaskService{
		client:           client,
		config:           cfg,
		headers:          headers,
		baseURL:          baseURL,
		workspaceService: ws,
	}
}

// CreateTaskRequest represents a task creation request
type CreateTaskRequest struct {
	Name            string
	RequirementID   string
	WorkspaceKey    string
	WorkspaceName   string
	ProjectCode     string
	TaskType        string
	Assignee        string
	PlannedHours    int
}

// TaskResponse represents a simplified task response
type TaskResponse struct {
	ObjectID   string   `json:"objectId"`
	Name       string   `json:"name"`
	Key        string   `json:"key"`
	Status     string   `json:"status"`
	Assignee   string   `json:"assignee"`
	Ancestors  []string `json:"ancestors,omitempty"`
}

// Create creates a new task
func (s *TaskService) Create(req *CreateTaskRequest) (*TaskResponse, error) {
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
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create task: status %d", resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ObjectID  string `json:"objectId"`
			Name      string `json:"name"`
			Key       string `json:"key"`
			Ancestors []struct {
				ObjectID string `json:"objectId"`
				Name     string `json:"name"`
			} `json:"ancestors"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to create task")
	}

	// Extract ancestor names
	ancestorNames := make([]string, 0, len(result.Data.Ancestors))
	for _, ancestor := range result.Data.Ancestors {
		ancestorNames = append(ancestorNames, ancestor.Name)
	}

	return &TaskResponse{
		ObjectID:  result.Data.ObjectID,
		Name:      result.Data.Name,
		Key:       result.Data.Key,
		Status:    "新建",
		Ancestors: ancestorNames,
	}, nil
}

// buildCreateBody builds the request body for task creation
func (s *TaskService) buildCreateBody(req *CreateTaskRequest, workspaceObjectID string) map[string]interface{} {
	now := time.Now()

	// Default values
	taskType := req.TaskType
	if taskType == "" {
		taskType = "开发"
	}

	assignee := req.Assignee
	if assignee == "" {
		assignee = s.config.User.Username
	}

	plannedHours := req.PlannedHours
	if plannedHours <= 0 {
		plannedHours = 8
	}

	// Calculate planned end date
	plannedEndDate := now.Add(time.Duration(plannedHours) * time.Hour)

	body := map[string]interface{}{
		"name":             req.Name,
		"itemType":         "任务",
		"taskType":         taskType,
		"spaceId":          workspaceObjectID,
		"projectCode":      req.ProjectCode,
		"assignedTo":       assignee,
		"plannedWorkingTime": plannedHours,
		"plannedStartDate": map[string]interface{}{"iso": now.Format(time.RFC3339)},
		"plannedEndDate":   map[string]interface{}{"iso": plannedEndDate.Format(time.RFC3339)},
	}

	// Add parent requirement if specified
	if req.RequirementID != "" {
		body["parent"] = map[string]interface{}{
			"__type":    "Pointer",
			"className": "Item",
			"objectId":  req.RequirementID,
		}
	}

	return body
}

// Delete deletes tasks by their object IDs
func (s *TaskService) Delete(objectIDs []string) error {
	if len(objectIDs) == 0 {
		return fmt.Errorf("no tasks to delete")
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
		return fmt.Errorf("failed to delete tasks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete tasks: status %d", resp.StatusCode)
	}

	return nil
}

// ListTasksRequest represents a list tasks request
type ListTasksRequest struct {
	WorkspaceName string
	Keyword       string
	Status        string
	PageNo        int
	PageSize      int
}

// ListTasksResponse represents a paginated list response
type ListTasksResponse struct {
	Items      []map[string]interface{} `json:"items"`
	TotalCount int                      `json:"totalCount"`
	PageNo     int                      `json:"pageNo"`
	PageSize   int                      `json:"pageSize"`
}

// List lists tasks with filters
func (s *TaskService) List(req *ListTasksRequest) (*ListTasksResponse, error) {
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
		return nil, fmt.Errorf("failed to list tasks: %w", err)
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

	return &ListTasksResponse{
		Items:      result.Data.Items,
		TotalCount: result.Data.TotalCount,
		PageNo:     req.PageNo,
		PageSize:   req.PageSize,
	}, nil
}

// buildListIQL builds IQL query for listing tasks
func (s *TaskService) buildListIQL(req *ListTasksRequest) string {
	// This is a simplified version; use IQLService for complex queries
	iql := fmt.Sprintf("((所属空间 = '%s') and (类型 in [\"任务\"]))", req.WorkspaceName)

	if req.Keyword != "" {
		iql += fmt.Sprintf(" and (名称包含 '%s')", req.Keyword)
	}

	if req.Status != "" {
		iql += fmt.Sprintf(" and (状态 = '%s')", req.Status)
	}

	iql += " order by 创建时间 desc"
	return iql
}
