package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/matoous/go-nanoid/v2"
	"github.com/user/lc/internal/config"
)

// Workspace represents a pointer to Workspace class
type Workspace struct {
	Type      string `json:"__type"`
	ClassName string `json:"className"`
	ObjectID  string `json:"objectId"`
}

// ItemType represents a pointer to ItemType class
type ItemType struct {
	Type      string `json:"__type"`
	ClassName string `json:"className"`
	ObjectID  string `json:"objectId"`
	Key       string `json:"key,omitempty"`
	Name      string `json:"name,omitempty"`
}

// UserValue represents a user value in requirement fields
type UserValue struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Deleted  bool   `json:"deleted"`
	Enabled  bool   `json:"enabled"`
	Email    string `json:"email,omitempty"`
}

// EditorContent represents rich text content
type EditorContent struct {
	Children []struct {
		Text string `json:"text"`
	} `json:"children"`
	Type string `json:"type"`
	ID   string `json:"id"`
}

// RequirementValues represents the values field in requirement request
type RequirementValues struct {
	Source                 [][]string      `json:"source"`
	BelongingSpace         string          `json:"belongingSpace"`
	Proposer               []UserValue     `json:"proposer"`
	AffiliatedUnit         string          `json:"affiliatedunit,omitempty"`
	ContactNumber          string          `json:"contactNumber,omitempty"`
	ContactEmail           string          `json:"contactEmail,omitempty"`
	RequestSubmissionTime  int64           `json:"requestSubmissionTime"`
	YearDemand             int64           `json:"yearDemand"`
	ExpectedCompletionTime int64           `json:"expectedCompletionTime"`
	PlannedStartTime       int64           `json:"plannedStartTime"`
	PlannedEndTime         int64           `json:"plannedEndTime"`
	BusinessBackground     []EditorContent `json:"businessBackground"`
	Requirement            []EditorContent `json:"requirement"`
	AcceptanceCriteria     []EditorContent `json:"acceptanceCriteria"`
	ProjectNo              []interface{}   `json:"projectNo"`
	RequirementType        []string        `json:"requirementType"`
	ProposeSpace           string          `json:"proposeSpace"`
	DevelopmentCompletion  interface{}     `json:"developmentCompletion"`
	Relations              interface{}     `json:"Relations"`
	Assignee               []UserValue     `json:"assignee"`
	Priority               string          `json:"priority"`
	ScreenType             string          `json:"__screen_type"`
}

// FilterItemType represents filter item type in parse context
type FilterItemType struct {
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
	Key       string    `json:"key"`
	Icon      string    `json:"icon"`
	Name      string    `json:"name"`
	UpdatedBy Workspace `json:"updatedBy"`
	ObjectID  string    `json:"objectId"`
}

// EventExtraData represents event extra data in parse context
type EventExtraData struct {
	FilterItemTypeList []FilterItemType `json:"filterItemTypeList"`
}

// ParseContext represents parse context in requirement request
type ParseContext struct {
	EventExtraData EventExtraData `json:"eventExtraData"`
}

// ItemContext represents item context in requirement request
type ItemContext struct{}

// RequirementCreateRequest represents the request body for creating a requirement
type RequirementCreateRequest struct {
	Name         string            `json:"name"`
	Ancestors    []string          `json:"ancestors"`
	Workspace    Workspace         `json:"workspace"`
	ItemType     ItemType          `json:"itemType"`
	Values       RequirementValues `json:"values"`
	Reporter     interface{}       `json:"reporter"`
	ItemContext  ItemContext       `json:"itemContext"`
	ParseContext ParseContext      `json:"parseContext"`
}

// APIErrorResponse represents the error response from API
type APIErrorResponse struct {
	Head struct {
		RequestID  string `json:"requestId"`
		RespStatus string `json:"respStatus"`
		RespCode   string `json:"respCode"`
		RespDesc   string `json:"respDesc"`
	} `json:"head"`
	Data interface{} `json:"data"`
}

// RequirementCreateResponse represents the response from create requirement API
// The API returns the created object directly instead of a standard response wrapper
type RequirementCreateResponse struct {
	Name      string          `json:"name"`
	ObjectID  string          `json:"objectId"`
	Key       string          `json:"key"`
	CreatedAt string          `json:"createdAt"`
	Workspace Workspace       `json:"workspace"`
	Status    Status          `json:"status"`
	Values    json.RawMessage `json:"values,omitempty"`
}

// Status represents the status object in response
type Status struct {
	Type       string `json:"__type"`
	ClassName  string `json:"className"`
	ObjectID   string `json:"objectId"`
	Name       string `json:"name,omitempty"`
	StatusType string `json:"type,omitempty"`
}

// RequirementService handles requirement API operations
type RequirementService struct {
	BaseService
	Config *config.Config
}

// NewRequirementService creates a new RequirementService
func NewRequirementService(baseURL string, headers map[string]string, client HTTPClient, cfg *config.Config) *RequirementService {
	return &RequirementService{
		BaseService: NewBaseService(baseURL, headers, client),
		Config:      cfg,
	}
}

// prepareHeaders merges config headers with service headers
func (s *RequirementService) prepareHeaders(workspaceKey string) map[string]string {
	headers := s.Config.GetHeadersWithWorkspace(workspaceKey)
	for k, v := range s.Headers {
		headers[k] = v
	}
	return headers
}

// Create creates a new requirement
func (s *RequirementService) Create(requestData *RequirementCreateRequest, workspaceKey string) (*RequirementCreateResponse, error) {
	// Create custom request with merged headers
	apiReq := &Request{
		URL:     s.BaseURL + "/api/v2/items",
		Method:  http.MethodPost,
		Headers: s.prepareHeaders(workspaceKey),
		Body:    requestData,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Try to parse error response with "code" and "error" fields
		var parseErr struct {
			Code  int    `json:"code"`
			Error string `json:"error"`
		}
		if err := json.Unmarshal(body, &parseErr); err == nil && parseErr.Error != "" {
			return nil, FormatAIError(parseErr.Error)
		}
		// Try to parse standard API error response
		var errResp APIErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Head.RespCode != "" && errResp.Head.RespCode != "00" {
			return nil, FormatAIError(errResp.Head.RespDesc)
		}
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Check if it's an error response
	var errResp APIErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Head.RespCode != "" && errResp.Head.RespCode != "00" {
		return nil, FormatAIError(errResp.Head.RespDesc)
	}

	var result RequirementCreateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}

// TaskCreateRequest represents the request body for creating a task
type TaskCreateRequest struct {
	Name         string           `json:"name"`
	Ancestors    []string         `json:"ancestors"`
	Workspace    Workspace        `json:"workspace"`
	ItemType     ItemType         `json:"itemType"`
	Values       TaskValues       `json:"values"`
	Reporter     interface{}      `json:"reporter"`
	ItemContext  ItemContext      `json:"itemContext"`
	ParseContext TaskParseContext `json:"parseContext"`
}

// TaskValues represents the values field in task request
type TaskValues struct {
	TaskDescription     []EditorContent `json:"taskDescription"`
	TaskType            []string        `json:"taskType"`
	PlannedStartTime    int64           `json:"plannedStartTime"`
	PlannedEndTime      int64           `json:"plannedEndTime"`
	PlannedWorkingHours int             `json:"plannedWorkingHours,omitempty"`
	EgreeOfImportance   []string        `json:"egreeOfImportance,omitempty"`
	ProjectNo           []interface{}   `json:"projectNo"`
	Relations           interface{}     `json:"Relations"`
	Assignee            []UserValue     `json:"assignee"`
	Priority            string          `json:"priority"`
	ScreenType          string          `json:"__screen_type"`
}

// TaskParseContext represents parse context in task request
type TaskParseContext struct {
	EventExtraData            map[string]interface{} `json:"eventExtraData,omitempty"`
	SkipValidateOptionsFields []string               `json:"skipValidateOptionsFields,omitempty"`
}

// TaskCreateResponse represents the response from create task API
type TaskCreateResponse struct {
	Code      int             `json:"code"`
	Error     string          `json:"error"`
	Name      string          `json:"name"`
	ObjectID  string          `json:"objectId"`
	Key       string          `json:"key"`
	CreatedAt string          `json:"createdAt"`
	Workspace Workspace       `json:"workspace"`
	Status    Status          `json:"status"`
	Values    json.RawMessage `json:"values,omitempty"`
	Ancestors []string        `json:"ancestors"`
}

// CreateTask creates a new task under a requirement
func (s *RequirementService) CreateTask(requestData *TaskCreateRequest, workspaceKey string) (*TaskCreateResponse, error) {
	// Get headers from config and merge with service headers (which include Cookie)
	// Create custom request with merged headers
	apiReq := &Request{
		URL:     s.BaseURL + "/api/v2/items",
		Method:  http.MethodPost,
		Headers: s.prepareHeaders(workspaceKey),
		Body:    requestData,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}

	var result TaskCreateResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	// Check for API error
	if result.Code != 0 {
		return nil, fmt.Errorf("API error (code %d): %s", result.Code, result.Error)
	}

	return &result, nil
}

// RequirementListRequest represents the request body for listing requirements
type RequirementListRequest struct {
	IQL               string                 `json:"iql"`
	Size              int                    `json:"size"`
	From              int                    `json:"from"`
	IsExpand          bool                   `json:"isExpand"`
	IsShowAncestors   bool                   `json:"isShowAncestors"`
	IsShowDescendants bool                   `json:"isShowDescendants"`
	IsShowLinkItems   bool                   `json:"isShowLinkItems"`
	Extend            map[string]interface{} `json:"extend"`
	Fields            []string               `json:"fields"`
	RefererInfo       RefererInfo            `json:"refererInfo"`
}

// RefererInfo represents referer information
type RefererInfo struct {
	WorkspaceKey string `json:"workspaceKey"`
}

// RequirementItem represents a single requirement item in the list
type RequirementItem struct {
	Name       string                 `json:"name"`
	ObjectID   string                 `json:"objectId"`
	Key        string                 `json:"key"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt,omitempty"`
	Workspace  Workspace              `json:"workspace"`
	Status     Status                 `json:"status"`
	ItemType   ItemType               `json:"itemType"`
	Assignee   []UserValue            `json:"assignee,omitempty"`
	Ancestors  []RequirementItem      `json:"ancestors,omitempty"`
	ID         string                 `json:"id,omitempty"`
	RowID      string                 `json:"rowId,omitempty"`
	Values     ItemValues             `json:"values,omitempty"`
	DataQuotes map[string]interface{} `json:"dataQuotes,omitempty"`
	CreatedBy  interface{}            `json:"createdBy,omitempty"`
	Hit        bool                   `json:"hit,omitempty"`
}

// ItemValues represents the values field in requirement item
type ItemValues struct {
	ExpectedCompletionTime interface{} `json:"expectedCompletionTime,omitempty"`
	Priority               interface{} `json:"priority,omitempty"`
	ProjectNo              interface{} `json:"projectNo,omitempty"`
	Assignee               []UserValue `json:"assignee,omitempty"`
	EarlyWarning           interface{} `json:"earlyWarning,omitempty"`
}

// UserPointer represents a pointer to _User class
type UserPointer struct {
	Type      string `json:"__type"`
	ClassName string `json:"className"`
	ObjectID  string `json:"objectId"`
}

// RequirementListPayload represents the payload field in list response
type RequirementListPayload struct {
	TotalCount int               `json:"totalCount"`
	Count      int               `json:"count"`
	Items      []RequirementItem `json:"items"`
}

// RequirementListResponse represents the response from list requirement API
type RequirementListResponse struct {
	Code    int                    `json:"code"`
	Payload RequirementListPayload `json:"payload"`
}

// List retrieves a list of requirements
func (s *RequirementService) List(requestData *RequirementListRequest, workspaceKey string) (*RequirementListResponse, error) {
	// Get merged headers

	// Create custom request with modified headers
	apiReq := &Request{
		URL:     s.BaseURL + "/api/search/hierarchy",
		Method:  http.MethodPost,
		Headers: s.prepareHeaders(workspaceKey),
		Body:    requestData,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}

	var result RequirementListResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RequirementViewResponse represents the response from view requirement API
type RequirementViewResponse struct {
	Fields          []Field             `json:"fields"`
	CustomFieldKeys []string            `json:"customFieldKeys"`
	Item            RequirementViewItem `json:"item"`
	ApprovalInfo    ApprovalInfo        `json:"approvalInfo"`
	CheckInInfo     CheckInInfo         `json:"checkInInfo"`
	Watchers        []interface{}       `json:"watchers"`
	HasComment      bool                `json:"hasComment"`
	Screen          Screen              `json:"screen"`
	Revoke          bool                `json:"revoke"`
}

// Field represents a field definition
type Field struct {
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	FieldType FieldType `json:"fieldType"`
	Required  bool      `json:"required"`
	Hidden    bool      `json:"hidden"`
	Readonly  bool      `json:"readonly"`
	ObjectID  string    `json:"objectId"`
}

// FieldType represents field type information
type FieldType struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	DataType  string `json:"dataType"`
	Component string `json:"component"`
	ObjectID  string `json:"objectId"`
}

// RequirementViewItem represents the detailed requirement item
type RequirementViewItem struct {
	Name      string                 `json:"name"`
	ObjectID  string                 `json:"objectId"`
	Key       string                 `json:"key"`
	CreatedAt string                 `json:"createdAt"`
	UpdatedAt string                 `json:"updatedAt"`
	Status    StatusDetail           `json:"status"`
	ItemType  ItemTypeDetail         `json:"itemType"`
	Workspace WorkspaceDetail        `json:"workspace"`
	CreatedBy UserDetail             `json:"createdBy"`
	UpdatedBy UserDetail             `json:"updatedBy"`
	Values    map[string]interface{} `json:"values"`
	Ancestors []interface{}          `json:"ancestors"`
	Version   int                    `json:"version"`
}

// StatusDetail represents detailed status information
type StatusDetail struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	ObjectID string `json:"objectId"`
}

// ItemTypeDetail represents detailed item type information
type ItemTypeDetail struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	ObjectID string `json:"objectId"`
}

// WorkspaceDetail represents detailed workspace information
type WorkspaceDetail struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	ObjectID string `json:"objectId"`
}

// UserDetail represents detailed user information
type UserDetail struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	ObjectID string `json:"objectId"`
}

// ApprovalInfo represents approval information
type ApprovalInfo struct {
	ApprovalActive bool `json:"approvalActive"`
	ApprovalExist  bool `json:"approvalExist"`
}

// CheckInInfo represents check-in information
type CheckInInfo struct {
	Active       bool `json:"active"`
	CheckInExist bool `json:"checkInExist"`
}

// Screen represents screen configuration
type Screen struct {
	Name     string `json:"name"`
	ObjectID string `json:"objectId"`
}

// View retrieves a requirement detail by objectId
func (s *RequirementService) View(objectID string, workspaceKey string) (*RequirementViewResponse, error) {
	path := fmt.Sprintf("/api/items/%s?screen=view&currentPageWorkspace=%s", objectID, workspaceKey)

	// Get merged headers and remove Content-Type for GET requests without body
	headers := s.prepareHeaders(workspaceKey)
	delete(headers, "Content-Type")

	// Create custom request with modified headers
	apiReq := &Request{
		URL:     s.BaseURL + path,
		Method:  http.MethodGet,
		Headers: headers,
		Body:    nil,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result RequirementViewResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}

// TaskListRequest represents the request body for listing tasks
type TaskListRequest struct {
	IQL         string                 `json:"iql"`
	Size        int                    `json:"size"`
	From        int                    `json:"from"`
	Extend      map[string]interface{} `json:"extend"`
	Fields      []string               `json:"fields"`
	RefererInfo RefererInfo            `json:"refererInfo"`
}

// TaskListItem represents a single task item in the list
type TaskListItem struct {
	Name       string                 `json:"name"`
	ObjectID   string                 `json:"objectId"`
	Key        string                 `json:"key"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt,omitempty"`
	Workspace  Workspace              `json:"workspace"`
	Status     Status                 `json:"status"`
	ItemType   ItemType               `json:"itemType"`
	Assignee   []UserValue            `json:"assignee,omitempty"`
	Ancestors  []string               `json:"ancestors,omitempty"`
	ID         string                 `json:"id,omitempty"`
	RowID      string                 `json:"rowId,omitempty"`
	Values     TaskItemValues         `json:"values,omitempty"`
	DataQuotes map[string]interface{} `json:"dataQuotes,omitempty"`
	CreatedBy  interface{}            `json:"createdBy,omitempty"`
}

// TaskItemValues represents the values field in task item
type TaskItemValues struct {
	PlannedEndTime   interface{} `json:"plannedEndTime,omitempty"`
	PlannedStartTime interface{} `json:"plannedStartTime,omitempty"`
	Priority         interface{} `json:"priority,omitempty"`
	Assignee         []UserValue `json:"assignee,omitempty"`
	EarlyWarning     interface{} `json:"earlyWarning,omitempty"`
}

// TaskListPayload represents the payload field in task list response
type TaskListPayload struct {
	Count int            `json:"count"`
	Items []TaskListItem `json:"items"`
}

// TaskListResponse represents the response from list task API
type TaskListResponse struct {
	Code    int             `json:"code"`
	Payload TaskListPayload `json:"payload"`
}

// ListTasks retrieves a list of tasks
func (s *RequirementService) ListTasks(requestData *TaskListRequest, workspaceKey string) (*TaskListResponse, error) {
	// Get merged headers

	// Create custom request with modified headers
	apiReq := &Request{
		URL:     s.BaseURL + "/api/search",
		Method:  http.MethodPost,
		Headers: s.prepareHeaders(workspaceKey),
		Body:    requestData,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}

	var result TaskListResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// TaskDeleteRequest represents the request body for deleting tasks
type TaskDeleteRequest struct {
	ItemIds           string                 `json:"itemIds"`
	ProcessBarKeysMap map[string]string      `json:"processBarKeysMap"`
	Extend            map[string]interface{} `json:"extend"`
	ApplicationID     string                 `json:"_ApplicationId"`
}

// TaskDeleteResponse represents the response from delete task API
type TaskDeleteResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// DeleteTasks deletes tasks by their object IDs
func (s *RequirementService) DeleteTasks(itemIds []string, workspaceKey string) (*TaskDeleteResponse, error) {
	// Generate process bar keys map
	processBarKeysMap := make(map[string]string)
	for _, id := range itemIds {
		processBarKeysMap[id] = fmt.Sprintf("delete-items-%s", generateUUID())
	}

	// Join item IDs with comma
	itemIdsStr := strings.Join(itemIds, ",")

	// Get applicationID from config
	applicationID := s.Config.Auth.TenantID

	// Build extend data with workspace info
	extendData := map[string]interface{}{
		"extend": map[string]interface{}{
			"board": map[string]interface{}{
				"iql":           "'类型' in [\"任务\"]",
				"key":           "e19d6cdf-c971-4822-a6cc-cd4b8cf65fee",
				"routeKey":      "task",
				"name":          "任务",
				"sort":          4,
				"filterSource":  "inWorkspace",
				"icon":          "Panel1",
				"itemTypes":     []interface{}{},
				"itemTypeLimit": true,
				"objectId":      "Cf5v4BqCrP",
				"workspace": map[string]interface{}{
					"key": workspaceKey,
				},
			},
		},
	}

	requestData := &TaskDeleteRequest{
		ItemIds:           itemIdsStr,
		ProcessBarKeysMap: processBarKeysMap,
		Extend:            extendData,
		ApplicationID:     applicationID,
	}

	// Get merged headers and set custom Content-Type
	headers := s.prepareHeaders(workspaceKey)
	headers["Content-Type"] = "text/plain"

	// Create custom request with modified headers
	apiReq := &Request{
		URL:     s.BaseURL + "/functions/deleteItems",
		Method:  http.MethodPost,
		Headers: headers,
		Body:    requestData,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}

	// The API returns empty object {} on success
	var result TaskDeleteResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		// Empty response is also considered success
		return &TaskDeleteResponse{Code: 0}, nil
	}

	return &result, nil
}

// generateUUID generates a NanoID for process bar keys
func generateUUID() string {
	id, err := gonanoid.New()
	if err != nil {
		// Fallback to a timestamp-based ID if NanoID generation fails
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return id
}

// SearchRequest represents the request body for searching requirements
type SearchRequest struct {
	IQL               string                 `json:"iql"`
	Size              int                    `json:"size"`
	From              int                    `json:"from"`
	IsExpand          bool                   `json:"isExpand"`
	IsShowAncestors   bool                   `json:"isShowAncestors"`
	IsShowDescendants bool                   `json:"isShowDescendants"`
	IsShowLinkItems   bool                   `json:"isShowLinkItems"`
	Extend            map[string]interface{} `json:"extend"`
	Fields            []string               `json:"fields"`
	RefererInfo       RefererInfo            `json:"refererInfo"`
}

// Search searches requirements with IQL query
func (s *RequirementService) Search(requestData *SearchRequest, workspaceKey string) (*RequirementListResponse, error) {
	// Get merged headers

	// Create custom request with modified headers
	apiReq := &Request{
		URL:     s.BaseURL + "/api/search/hierarchy",
		Method:  http.MethodPost,
		Headers: s.prepareHeaders(workspaceKey),
		Body:    requestData,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}

	var result RequirementListResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SearchTasksRequest represents the request body for searching tasks
type SearchTasksRequest struct {
	IQL         string                 `json:"iql"`
	Size        int                    `json:"size"`
	From        int                    `json:"from"`
	Extend      map[string]interface{} `json:"extend"`
	Fields      []string               `json:"fields"`
	RefererInfo RefererInfo            `json:"refererInfo"`
}

// SearchTasks searches tasks with IQL query
func (s *RequirementService) SearchTasks(requestData *SearchTasksRequest, workspaceKey string) (*TaskListResponse, error) {
	// Get merged headers

	// Create custom request with modified headers
	apiReq := &Request{
		URL:     s.BaseURL + "/api/search",
		Method:  http.MethodPost,
		Headers: s.prepareHeaders(workspaceKey),
		Body:    requestData,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}

	var result TaskListResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteRequirements deletes requirements by their object IDs
func (s *RequirementService) DeleteRequirements(itemIds []string, workspaceKey string) (*TaskDeleteResponse, error) {
	// Generate process bar keys map
	processBarKeysMap := make(map[string]string)
	for _, id := range itemIds {
		processBarKeysMap[id] = fmt.Sprintf("delete-items-%s", generateUUID())
	}

	// Join item IDs with comma
	itemIdsStr := strings.Join(itemIds, ",")

	// Get applicationID from config
	applicationID := s.Config.Auth.TenantID

	// Build extend data with workspace info for requirements
	extendData := map[string]interface{}{
		"extend": map[string]interface{}{
			"board": map[string]interface{}{
				"iql":           "'belongingSpace' in [\"currentWorkspace()\"] or ('类型' in [\"用户故事\",\"任务\"] )",
				"key":           "bf94a87a-f995-4f40-bea6-a51cc846100a",
				"routeKey":      "initial",
				"name":          "需求",
				"sort":          4,
				"filterSource":  "inWorkspace",
				"icon":          "Panel1",
				"itemTypes":     []interface{}{},
				"itemTypeLimit": false,
				"objectId":      "tbV8vdL2l3",
				"workspace": map[string]interface{}{
					"key": workspaceKey,
				},
			},
		},
	}

	requestData := &TaskDeleteRequest{
		ItemIds:           itemIdsStr,
		ProcessBarKeysMap: processBarKeysMap,
		Extend:            extendData,
		ApplicationID:     applicationID,
	}

	// Get merged headers and set custom Content-Type
	headers := s.prepareHeaders(workspaceKey)
	headers["Content-Type"] = "text/plain"

	// Create custom request with modified headers
	apiReq := &Request{
		URL:     s.BaseURL + "/functions/deleteItems",
		Method:  http.MethodPost,
		Headers: headers,
		Body:    requestData,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}

	// The API returns empty object {} on success
	var result TaskDeleteResponse
	if err := s.ParseJSON(resp, &result); err != nil {
		// Empty response is also considered success
		return &TaskDeleteResponse{Code: 0}, nil
	}

	return &result, nil
}

// RequirementUpdateRequest represents the request body for updating a requirement
// Supports partial updates - only fields that need to be changed should be included
type RequirementUpdateRequest struct {
	ParseContext *UpdateParseContext `json:"parseContext,omitempty"`
	Values       UpdateValues        `json:"values"`
}

// UpdateParseContext represents parse context in update request
type UpdateParseContext struct {
	Extend UpdateExtend `json:"extend"`
}

// UpdateExtend represents extend data in update request
type UpdateExtend struct {
	Board UpdateBoard `json:"board"`
}

// UpdateBoard represents board data in update request
type UpdateBoard struct {
	IQL          string                 `json:"iql"`
	Key          string                 `json:"key"`
	RouteKey     string                 `json:"routeKey"`
	Name         string                 `json:"name"`
	Sort         int                    `json:"sort"`
	FilterSource string                 `json:"filterSource"`
	Icon         string                 `json:"icon"`
	Workspace    map[string]interface{} `json:"workspace,omitempty"`
}

// UpdateValues represents the values field in update request
type UpdateValues struct {
	ScreenType            string          `json:"__screen_type,omitempty"`
	Requirement           []EditorContent `json:"requirement,omitempty"`
	AcceptanceCriteria    []EditorContent `json:"acceptanceCriteria,omitempty"`
	BusinessBackground    []EditorContent `json:"businessBackground,omitempty"`
	FeasibilityAssessment string          `json:"feasibilityAssessment,omitempty"`
	PlannedEndTime        *int64          `json:"plannedEndTime,omitempty"`
	PlannedStartTime      *int64          `json:"plannedStartTime,omitempty"`
	Name                  string          `json:"name,omitempty"`
	Priority              string          `json:"priority,omitempty"`
}

// RequirementUpdateResponse represents the response from update requirement API
type RequirementUpdateResponse struct {
	Item RequirementUpdateItem `json:"item"`
}

// RequirementUpdateItem represents the item in update response
type RequirementUpdateItem struct {
	ObjectID  string                 `json:"objectId"`
	Key       string                 `json:"key"`
	Name      string                 `json:"name"`
	CreatedAt string                 `json:"createdAt"`
	UpdatedAt string                 `json:"updatedAt"`
	Status    Status                 `json:"status"`
	ItemType  ItemType               `json:"itemType"`
	Workspace Workspace              `json:"workspace"`
	Values    UpdateResponseValues   `json:"values"`
	Version   int                    `json:"version"`
}

// UpdateResponseValues represents values in update response
type UpdateResponseValues struct {
	Name               string      `json:"name,omitempty"`
	Requirement        interface{} `json:"requirement,omitempty"`
	AcceptanceCriteria interface{} `json:"acceptanceCriteria,omitempty"`
	BusinessBackground interface{} `json:"businessBackground,omitempty"`
}

// Update updates a requirement by its object ID
func (s *RequirementService) Update(objectID string, requestData *RequirementUpdateRequest, workspaceKey string) (*RequirementUpdateResponse, error) {
	path := fmt.Sprintf("/api/v2/items/%s?detail=true", objectID)

	// Get merged headers

	// Create custom request with modified headers
	apiReq := &Request{
		URL:     s.BaseURL + path,
		Method:  http.MethodPut,
		Headers: s.prepareHeaders(workspaceKey),
		Body:    requestData,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errResp APIErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Head.RespCode != "" && errResp.Head.RespCode != "00" {
			return nil, FormatAIError(errResp.Head.RespDesc)
		}
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Check if it's an error response
	var errResp APIErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Head.RespCode != "" && errResp.Head.RespCode != "00" {
		return nil, FormatAIError(errResp.Head.RespDesc)
	}

	var result RequirementUpdateResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result, nil
}
