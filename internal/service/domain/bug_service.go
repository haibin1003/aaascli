package domain

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/config"
)

// BugTemplate constants
type BugTemplate string

const (
	BugTemplateFull   BugTemplate = "1969947017807343618"
	BugTemplateSimple BugTemplate = "1969947017924784130"
)

// BugService provides bug/defect-related business logic
type BugService struct {
	client     *api.Client
	config     *config.Config
	headers    map[string]string
	baseURL    string
}

// NewBugService creates a new bug service
func NewBugService(baseURL string, headers map[string]string, client *api.Client, cfg *config.Config) *BugService {
	return &BugService{
		client:  client,
		config:  cfg,
		headers: headers,
		baseURL: baseURL,
	}
}

// CreateBugRequest represents a bug creation request
type CreateBugRequest struct {
	Title         string
	Description   string
	WorkspaceKey  string
	ProjectID     string
	Template      BugTemplate
	Priority      int
	Level         int
	AssigneeID    string
}

// BugResponse represents a simplified bug response
type BugResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Priority   string `json:"priority"`
	Level      string `json:"level"`
	Assignee   string `json:"assignee"`
}

// Create creates a new bug
func (s *BugService) Create(req *CreateBugRequest) (*BugResponse, error) {
	// Select template
	template := req.Template
	if template == "" {
		template = BugTemplateSimple
	}

	body := map[string]interface{}{
		"title":       req.Title,
		"description": req.Description,
		"templateId":  string(template),
		"projectId":   req.ProjectID,
		"priority":    req.Priority,
		"level":       req.Level,
	}

	if req.AssigneeID != "" {
		body["assignUserId"] = req.AssigneeID
	}

	url := fmt.Sprintf("%s/defect/defectInfo/insertDefectInfo", s.baseURL)
	apiReq := &api.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: s.headers,
		Body:    body,
	}

	resp, err := s.client.Send(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create bug: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create bug: status %d", resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to create bug")
	}

	return &BugResponse{
		ID:       fmt.Sprintf("%d", result.Data.ID),
		Title:    req.Title,
		Status:   "新建",
		Priority: GetPriorityName(req.Priority),
		Level:    GetLevelName(req.Level),
	}, nil
}

// ListBugsRequest represents a list bugs request
type ListBugsRequest struct {
	ProjectID  string
	Status     string
	Keyword    string
	PageNo     int
	PageSize   int
}

// ListBugsResponse represents a paginated list response
type ListBugsResponse struct {
	Items      []map[string]interface{} `json:"items"`
	TotalCount int                      `json:"totalCount"`
	PageNo     int                      `json:"pageNo"`
	PageSize   int                      `json:"pageSize"`
}

// List lists bugs with filters
func (s *BugService) List(req *ListBugsRequest) (*ListBugsResponse, error) {
	body := map[string]interface{}{
		"projectId": req.ProjectID,
		"pageNo":    req.PageNo,
		"pageSize":  req.PageSize,
	}

	if req.Status != "" {
		body["status"] = req.Status
	}

	if req.Keyword != "" {
		body["keyword"] = req.Keyword
	}

	url := fmt.Sprintf("%s/defect/defectInfo/getDefectInfoPage", s.baseURL)
	apiReq := &api.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: s.headers,
		Body:    body,
	}

	resp, err := s.client.Send(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list bugs: %w", err)
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

	return &ListBugsResponse{
		Items:      result.Data.Items,
		TotalCount: result.Data.TotalCount,
		PageNo:     req.PageNo,
		PageSize:   req.PageSize,
	}, nil
}

// UpdateStatus updates the status of a bug
func (s *BugService) UpdateStatus(bugID int64, status string) error {
	body := map[string]interface{}{
		"id":     bugID,
		"status": status,
	}

	url := fmt.Sprintf("%s/defect/defectInfo/updateDefectStatus", s.baseURL)
	apiReq := &api.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: s.headers,
		Body:    body,
	}

	resp, err := s.client.Send(apiReq)
	if err != nil {
		return fmt.Errorf("failed to update bug status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update bug status: status %d", resp.StatusCode)
	}

	return nil
}

// Delete deletes a bug
func (s *BugService) Delete(bugID int64) error {
	url := fmt.Sprintf("%s/defect/defectInfo/deleteDefectInfo?id=%d", s.baseURL, bugID)
	apiReq := &api.Request{
		Method:  http.MethodDelete,
		URL:     url,
		Headers: s.headers,
	}

	resp, err := s.client.Send(apiReq)
	if err != nil {
		return fmt.Errorf("failed to delete bug: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete bug: status %d", resp.StatusCode)
	}

	return nil
}

// GetPriorityName returns priority name from code
func GetPriorityName(code int) string {
	names := map[int]string{
		0: "提示",
		1: "次要",
		2: "主要",
		3: "严重",
		4: "致命",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return "未知"
}

// GetLevelName returns level name from code
func GetLevelName(code int) string {
	names := map[int]string{
		1: "低",
		2: "中",
		3: "高",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return "未知"
}
