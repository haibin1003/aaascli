package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/user/lc/internal/config"
)

// DocService handles document library API operations
type DocService struct {
	BaseService
	Config *config.Config
}

// NewDocService creates a new DocService
func NewDocService(baseURL string, headers map[string]string, client HTTPClient, cfg *config.Config) *DocService {
	return &DocService{
		BaseService: NewBaseService(baseURL, headers, client),
		Config:      cfg,
	}
}

// DocLibrary represents a document library
type DocLibrary struct {
	LibID            int64      `json:"libId"`
	ExternalLibID    int64      `json:"externalLibId"`
	LibName          string     `json:"libName"`
	Intro            string     `json:"intro"`
	Icon             string     `json:"icon"`
	CreateTime       string     `json:"createTime"`
	LibType          int        `json:"libType"`
	TenantID         string     `json:"tenantId"`
	RightType        int        `json:"rightType"`
	OwnerIds         []DocOwner `json:"ownerIds"`
	LockStatus       int        `json:"lockStatus"`
	CloudDesktopOnly int        `json:"cloudDesktopOnly"`
	Collect          bool       `json:"collect"`
}

// DocOwner represents a library owner
type DocOwner struct {
	OwnerID   string `json:"ownerId"`
	OwnerName string `json:"ownerName"`
}

// DocLibraryListResponse represents the response from list libraries API
type DocLibraryListResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TotalNumber         int64        `json:"totalNumber"`
		DeputyAccountNumber string       `json:"deputyAccountNumber"`
		QuaRole             bool         `json:"quaRole"`
		List                []DocLibrary `json:"list"`
		TraceInfo           TraceInfo    `json:"traceInfo"`
	} `json:"data"`
}

// TraceInfo represents trace information
type TraceInfo struct {
	TrackID string `json:"trackId"`
	PodName string `json:"podName"`
}

// ListLibraries retrieves the list of document libraries
func (s *DocService) ListLibraries(workspaceKey string) (*DocLibraryListResponse, error) {
	// Get user info
	user := s.Config.GetUser()
	tenantID := s.Config.GetTenantID()

	// Build query parameters
	params := url.Values{}
	params.Set("deputyAccountNumber", user.Username)
	params.Set("tenantId", tenantID)
	params.Set("page", "true")
	params.Set("businessType", "4")
	params.Set("businessId", workspaceKey)

	path := "/moss/web/cmdevops-doc/integration/libraries?" + params.Encode()

	headers := s.Config.GetHeadersWithWorkspace(workspaceKey)
	for k, v := range s.Headers {
		headers[k] = v
	}
	delete(headers, "Content-Type")

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var result DocLibraryListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

// CreateLibraryRequest represents the request to create a document library
type CreateLibraryRequest struct {
	TenantID             string                `json:"tenantId"`
	Description          string                `json:"description"`
	Creator              string                `json:"creator"`
	DeputyAccountNumber  string                `json:"deputyAccountNumber"`
	Icon                 string                `json:"icon"`
	Inherit              bool                  `json:"inherit"`
	ExternalLibID        interface{}           `json:"externalLibId"`
	OwnerIds             []DocOwnerInput       `json:"ownerIds"`
	LibName              string                `json:"libName"`
	LibType              int                   `json:"libType"`
	LibCode              string                `json:"libCode"`
	DeputyAccountNumbers []DeputyAccountNumber `json:"deputyAccountNumbers"`
}

// DocOwnerInput represents owner input
type DocOwnerInput struct {
	OwnerID   string `json:"ownerId"`
	OwnerName string `json:"ownerName"`
}

// DeputyAccountNumber represents deputy account with permissions
type DeputyAccountNumber struct {
	DeputyAccountNumber string       `json:"deputyAccountNumber"`
	UserName            string       `json:"userName"`
	Account             string       `json:"account"`
	OrgName             interface{}  `json:"orgName"`
	OrgID               interface{}  `json:"orgId"`
	OrgCode             interface{}  `json:"orgCode"`
	MemberSource        int          `json:"memberSource"`
	Permissions         []Permission `json:"permissions"`
}

// Permission represents a permission
type Permission struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
}

// CreateLibraryResponse represents the response from create library API
type CreateLibraryResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		LibID     int64     `json:"libId"`
		TraceInfo TraceInfo `json:"traceInfo"`
	} `json:"data"`
}

// CreateLibrary creates a new document library
func (s *DocService) CreateLibrary(name, workspaceKey string) (*CreateLibraryResponse, error) {
	user := s.Config.GetUser()
	tenantID := s.Config.GetTenantID()

	// LibCode must be the workspaceKey for permission validation
	libCode := workspaceKey

	requestData := &CreateLibraryRequest{
		TenantID:            tenantID,
		Description:         "",
		Creator:             user.Username,
		DeputyAccountNumber: user.Username,
		Icon:                "icon1",
		Inherit:             false,
		ExternalLibID:       nil,
		OwnerIds: []DocOwnerInput{
			{
				OwnerID:   user.Username,
				OwnerName: user.Nickname,
			},
		},
		LibName: name,
		LibType: 4,
		LibCode: libCode,
		DeputyAccountNumbers: []DeputyAccountNumber{
			{
				DeputyAccountNumber: user.Username,
				UserName:            user.Nickname,
				Account:             user.Username,
				OrgName:             nil,
				OrgID:               nil,
				OrgCode:             nil,
				MemberSource:        2,
				Permissions: []Permission{
					{
						ID:   "12",
						Name: "浏览权限",
						Type: 6,
					},
				},
			},
		},
	}

	headers := s.Config.GetHeadersWithWorkspace(workspaceKey)
	for k, v := range s.Headers {
		headers[k] = v
	}

	apiReq := &Request{
		URL:     s.BaseURL + "/moss/web/cmdevops-doc/integration/libraries/create",
		Method:  http.MethodPost,
		Headers: headers,
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var result CreateLibraryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

// DeleteLibraryResponse represents the response from delete library API
type DeleteLibraryResponse struct {
	Code    int           `json:"code"`
	List    []interface{} `json:"list"`
	TrackID string        `json:"trackId"`
	PodName string        `json:"podName"`
}

// DeleteLibrary deletes a document library by external lib ID
func (s *DocService) DeleteLibrary(externalLibID int64) (*DeleteLibraryResponse, error) {
	// First, get client IP
	clientIP, err := s.getClientIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get client IP: %w", err)
	}

	// Then, get OAuth authorization (use externalLibID as folderID for library deletion)
	authToken, err := s.getDocAuthToken(clientIP, externalLibID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	path := fmt.Sprintf("/cmdevops-doc-editor-cloud/folder/deleteLibrary?id=%d", externalLibID)

	headers := make(map[string]string)
	for k, v := range s.Headers {
		headers[k] = v
	}
	delete(headers, "Content-Type")
	headers["Authorization"] = authToken

	// Get the base URL without the path
	baseDocURL := "https://rdcloud.4c.hq.cmcc"

	apiReq := &Request{
		URL:     baseDocURL + path,
		Method:  http.MethodGet,
		Headers: headers,
		Body:    nil,
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var result DeleteLibraryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: code %d, response: %s", result.Code, string(body))
	}

	return &result, nil
}

// CreateFolderResponse represents the response from create folder API
type CreateFolderResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    FolderData `json:"data"`
	TrackID string     `json:"trackId"`
	PodName string     `json:"podName"`
}

// FolderData represents folder creation data
type FolderData struct {
	ID         int64  `json:"id"`
	PrtID      int64  `json:"prtId"`
	Name       string `json:"name"`
	CreateTime string `json:"createTime"`
}

// CreateFolder creates a folder in a document library
// Automatically handles OAuth authentication
func (s *DocService) CreateFolder(prtID int64, name string) (*CreateFolderResponse, error) {
	// Step 1: Get client IP
	clientIP, err := s.getClientIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get client IP: %w", err)
	}

	// Step 2: Get OAuth authorization with client IP and prtID for permission validation
	authToken, err := s.getDocAuthToken(clientIP, prtID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	path := "/cmdevops-doc-editor-cloud/folder/create"

	headers := make(map[string]string)
	for k, v := range s.Headers {
		headers[k] = v
	}
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	headers["Authorization"] = authToken
	headers["x-client-ip"] = clientIP
	// Update Referer to include folderId for permission validation
	headers["Referer"] = fmt.Sprintf("https://rdcloud.4c.hq.cmcc/base/cmdevops-doc-editor-web/document-space/document-library?folderId=%d&type=space&clientIp=%s&source=home&showBack=true", prtID, clientIP)
	// Also update Origin
	headers["Origin"] = "https://rdcloud.4c.hq.cmcc"

	// Form data
	formData := url.Values{}
	formData.Set("prtId", strconv.FormatInt(prtID, 10))
	formData.Set("name", name)

	baseDocURL := "https://rdcloud.4c.hq.cmcc"

	apiReq := &Request{
		URL:     baseDocURL + path,
		Method:  http.MethodPost,
		Headers: headers,
		Body:    strings.NewReader(formData.Encode()),
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var result CreateFolderResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: code %d, message: %s", result.Code, result.Message)
	}

	return &result, nil
}

// TreeItem represents an item in the folder tree
type TreeItem struct {
	ID               int64  `json:"id"`
	Text             string `json:"text"`      // 文件夹/文件名称
	ObjType          int    `json:"objType"`   // 32=文件夹, 33=文件
	Permission       int    `json:"permission"`
	CabPermission    int    `json:"cabPermission"`
	FolderPermission int    `json:"folderPermission"`
	DocPermission    int    `json:"docPermission"`
	AttrPermission   int    `json:"attrPermission"`
	FlowPermission   int    `json:"flowPermission"`
	HasChild         bool   `json:"hasChild"`
	Nodes            []interface{} `json:"nodes"`
}

// TreeListResponse represents the response from treeList API
type TreeListResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	List    []TreeItem `json:"list"`
	TrackID string     `json:"trackId"`
	PodName string     `json:"podName"`
}

// TreeList retrieves the folder/file tree list for a given parent folder
func (s *DocService) TreeList(prtID int64) (*TreeListResponse, error) {
	// Get client IP
	clientIP, err := s.getClientIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get client IP: %w", err)
	}

	// Get OAuth authorization
	authToken, err := s.getDocAuthToken(clientIP, prtID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Build query parameters
	params := url.Values{}
	params.Set("prtId", strconv.FormatInt(prtID, 10))

	path := "/cmdevops-doc-editor-cloud/dms/treeList?" + params.Encode()

	headers := make(map[string]string)
	for k, v := range s.Headers {
		headers[k] = v
	}
	delete(headers, "Content-Type")
	headers["Authorization"] = authToken
	headers["x-client-ip"] = clientIP
	headers["Referer"] = fmt.Sprintf("https://rdcloud.4c.hq.cmcc/base/cmdevops-doc-editor-web/document-space/document-library?folderId=%d&type=space&clientIp=%s&source=home&showBack=true", prtID, clientIP)

	baseDocURL := "https://rdcloud.4c.hq.cmcc"

	apiReq := &Request{
		URL:     baseDocURL + path,
		Method:  http.MethodGet,
		Headers: headers,
		Body:    nil,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, fmt.Errorf("treeList request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var result TreeListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: code %d, message: %s", result.Code, result.Message)
	}

	return &result, nil
}

// PageListItem represents an item in the page list (file or folder)
type PageListItem struct {
	ID             string `json:"id"`
	ObjType        int    `json:"objType"` // 32=文件夹, 33=文件
	Name           string `json:"name"`
	CreatedDt      string `json:"createdDt"`
	UpdatedBy      int64  `json:"updatedBy"`
	UpdatedByName  string `json:"updatedByName"`
	UpdatedDt      string `json:"updatedDt"`
	PrtID          int64  `json:"prtId"`
	OwnerID        int64  `json:"ownerId"`
	OwnerName      string `json:"ownerName"`
	Permission     int    `json:"permission"`
	CurrentRev     int64  `json:"currentRev"`
	RevNum         string `json:"revNum"`
	Ext            string `json:"ext"`
	Size           int64  `json:"size"`
	SizeStr        string `json:"sizeStr"`
	LockState      int    `json:"lockState"`
	CanView        bool   `json:"canView"`
	CanEdit        bool   `json:"canEdit"`
	CanDownload    bool   `json:"canDownload"`
}

// PageListData represents the data section of page list response
type PageListData struct {
	TotalNumber  int64  `json:"totalNumber"`
	PageNo       int    `json:"pageNo"`
	PageSize     int    `json:"pageSize"`
	TotalPage    int    `json:"totalPage"`
	PrtID        int64  `json:"prtId"`
	PrtName      string `json:"prtName"`
	LastModifyTime string `json:"lastModifyTime"`
	Permission   int    `json:"permission"`
}

// PageListResponse represents the response from pageList API
type PageListResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    PageListData   `json:"data"`
	List    []PageListItem `json:"list"`
	TrackID string         `json:"trackId"`
	PodName string         `json:"podName"`
}

// PageList retrieves the paginated file/folder list for a given folder
func (s *DocService) PageList(prtID int64, pageNo, pageSize int) (*PageListResponse, error) {
	// Get client IP
	clientIP, err := s.getClientIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get client IP: %w", err)
	}

	// Get OAuth authorization
	authToken, err := s.getDocAuthToken(clientIP, prtID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Build query parameters
	params := url.Values{}
	params.Set("prtId", strconv.FormatInt(prtID, 10))
	params.Set("pageNo", strconv.Itoa(pageNo))
	params.Set("pageSize", strconv.Itoa(pageSize))
	params.Set("name", "") // empty name for all items

	path := "/cmdevops-doc-editor-cloud/dms/pageList?" + params.Encode()

	headers := make(map[string]string)
	for k, v := range s.Headers {
		headers[k] = v
	}
	delete(headers, "Content-Type")
	headers["Authorization"] = authToken
	headers["x-client-ip"] = clientIP
	headers["Referer"] = fmt.Sprintf("https://rdcloud.4c.hq.cmcc/base/cmdevops-doc-editor-web/document-space/document-library?folderId=%d&type=space&clientIp=%s&source=home&showBack=true", prtID, clientIP)

	baseDocURL := "https://rdcloud.4c.hq.cmcc"

	apiReq := &Request{
		URL:     baseDocURL + path,
		Method:  http.MethodGet,
		Headers: headers,
		Body:    nil,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, fmt.Errorf("pageList request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var result PageListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: code %d, message: %s", result.Code, result.Message)
	}

	return &result, nil
}

// OAuthAuthorizeResponse represents the OAuth authorize response
type OAuthAuthorizeResponse struct {
	Head struct {
		RequestID  string `json:"requestId"`
		RespStatus string `json:"respStatus"`
		RespCode   string `json:"respCode"`
		RespDesc   string `json:"respDesc"`
	} `json:"head"`
	Data struct {
		Code string `json:"code"`
	} `json:"data"`
}

// PzLoginResponse represents the pzLogin response
type PzLoginResponse struct {
	Code int `json:"code"`
	Data struct {
		SessionID string `json:"sessionId"`
		ID        string `json:"id"`
		Name      string `json:"name"`
		LoginName string `json:"loginName"`
	} `json:"data"`
	TrackID string `json:"trackId"`
	PodName string `json:"podName"`
}

// getClientIP gets the client IP from the server
func (s *DocService) getClientIP() (string, error) {
	path := "/moss/web/cmdevops-doc/integration/libraries/client-ip"

	// 使用已有的 headers，从 Config 获取基础 headers
	headers := s.Config.GetHeadersWithWorkspace("")
	for k, v := range s.Headers {
		headers[k] = v
	}
	delete(headers, "Content-Type")

	apiReq := &Request{
		URL:     s.BaseURL + path,
		Method:  http.MethodGet,
		Headers: headers,
		Body:    nil,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return "", fmt.Errorf("get client ip failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ClientIP string `json:"clientIp"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing response: %w, body: %s", err, string(body))
	}

	if !result.Success {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	return result.Data.ClientIP, nil
}

// getDocAuthToken gets the authentication token for document editor APIs
func (s *DocService) getDocAuthToken(clientIP string, folderID int64) (string, error) {
	// Step 1: Get sourceid from env.json
	envURL := "https://rdcloud.4c.hq.cmcc/cmdevops-doc-editor-web/env.json"

	envHeaders := make(map[string]string)
	for k, v := range s.Headers {
		envHeaders[k] = v
	}
	delete(envHeaders, "Content-Type")
	envHeaders["Referer"] = fmt.Sprintf("https://rdcloud.4c.hq.cmcc/base/cmdevops-doc-editor-web/auth?folderId=%d&type=space&clientIp=%s&source=home&showBack=true", folderID, clientIP)

	envReq := &Request{
		URL:     envURL,
		Method:  http.MethodGet,
		Headers: envHeaders,
		Body:    nil,
	}

	envResp, err := s.HTTPClient.Send(envReq)
	if err != nil {
		return "", fmt.Errorf("get env.json failed: %w", err)
	}
	defer envResp.Body.Close()

	envBody, err := io.ReadAll(envResp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading env response: %w", err)
	}

	var envResult struct {
		SourceID string `json:"sourceid"`
	}
	if err := json.Unmarshal(envBody, &envResult); err != nil {
		return "", fmt.Errorf("error parsing env response: %w", err)
	}

	sourceID := envResult.SourceID
	if sourceID == "" {
		sourceID = "4298157360" // 默认 sourceid
	}

	// Step 2: OAuth authorize to get session
	authorizeURL := "https://rdcloud.4c.hq.cmcc/moss/web/auth/v1/user/oauth/authorize"

	// Generate a unique session ID
	sessionID := generateUUID()
	redirectURI := fmt.Sprintf("https://rdcloud.4c.hq.cmcc/cmdevops-doc-editor-cloud/user/pzLogin?sessionId=%s", sessionID)

	authReqBody := map[string]interface{}{
		"head": map[string]interface{}{
			"requestId": generateRequestID(),
			"sourceid":  sourceID,
		},
		"data": map[string]interface{}{
			"redirectUri":  redirectURI,
			"responseType": "code",
		},
	}

	headers := make(map[string]string)
	for k, v := range s.Headers {
		headers[k] = v
	}
	headers["Content-Type"] = "application/json"
	headers["Referer"] = fmt.Sprintf("https://rdcloud.4c.hq.cmcc/base/cmdevops-doc-editor-web/auth?folderId=%d&type=space&clientIp=%s&source=home&showBack=true", folderID, clientIP)
	headers["Origin"] = "https://rdcloud.4c.hq.cmcc"

	apiReq := &Request{
		URL:     authorizeURL,
		Method:  http.MethodPost,
		Headers: headers,
		Body:    authReqBody,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return "", fmt.Errorf("oauth authorize failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading authorize response: %w", err)
	}

	// OAuth authorize directly returns user info with sessionId
	var userResp struct {
		Code int `json:"code"`
		Data struct {
			SessionID string `json:"sessionId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &userResp); err != nil {
		return "", fmt.Errorf("error parsing user response: %w, body: %s", err, string(body))
	}

	if userResp.Code != 200 {
		return "", fmt.Errorf("oauth failed: code %d, body: %s", userResp.Code, string(body))
	}

	return userResp.Data.SessionID, nil
}

// generateRequestID generates a request ID (timestamp in milliseconds)
func generateRequestID() int64 {
	return time.Now().UnixNano() / 1e6
}

// generateTraceID generates a trace ID for traceparent header
func generateTraceID() string {
	return fmt.Sprintf("%016x%016x", time.Now().UnixNano(), time.Now().UnixNano()>>32)
}

// generateSpanID generates a span ID for traceparent header
func generateSpanID() string {
	return fmt.Sprintf("%016x", time.Now().UnixNano())
}

// DeleteDmsResponse represents the delete file/folder response
type DeleteDmsResponse struct {
	Code    int           `json:"code"`
	Data    DeleteData    `json:"data"`
	List    []interface{} `json:"list"`
	TrackID string        `json:"trackId"`
	PodName string        `json:"podName"`
}

// DeleteData contains delete operation results
type DeleteData struct {
	WithOutPermissionCount int           `json:"withOutPermissionCount"`
	WithOutPermissionList  interface{}   `json:"withOutPermissionList"`
	FilePath               interface{}   `json:"filePath"`
	SuccessList            []DeletedItem `json:"successList"`
	FailureList            []interface{} `json:"failureList"`
}

// DeletedItem represents a successfully deleted item
type DeletedItem struct {
	ID         string `json:"id"`
	ObjType    int    `json:"objType"`
	Name       string `json:"name"`
	PrtID      int64  `json:"prtId"`
	PrtName    string `json:"prtName"`
	PrtPath    string `json:"prtPath"`
	Ext        string `json:"ext"`
	Size       int64  `json:"size"`
	SizeStr    string `json:"sizeStr"`
	CurrentRev int64  `json:"currentRev"`
	RevNum     string `json:"revNum"`
}

// DeleteFileOrFolder deletes a file or folder by object ID
// 完全复制创建文件夹的认证流程
func (s *DocService) DeleteFileOrFolder(objID int64, folderID int64) (*DeleteDmsResponse, error) {
	// Step 1: Get client IP (与CreateFolder完全相同)
	clientIP, err := s.getClientIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get client IP: %w", err)
	}

	// Step 2: Get OAuth authorization with client IP (与CreateFolder完全相同)
	authToken, err := s.getDocAuthToken(clientIP, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Step 3: 执行删除 (与CreateFolder的请求构造方式相同)
	path := "/cmdevops-doc-editor-cloud/dms/delete"

	headers := make(map[string]string)
	for k, v := range s.Headers {
		headers[k] = v
	}
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	headers["Authorization"] = authToken
	headers["x-client-ip"] = clientIP
	// Update Referer to include folderId for permission validation (与CreateFolder相同)
	headers["Referer"] = fmt.Sprintf("https://rdcloud.4c.hq.cmcc/base/cmdevops-doc-editor-web/document-space/document-library?folderId=%d&type=space&clientIp=%s&source=home&showBack=true", folderID, clientIP)
	// Also update Origin (与CreateFolder相同)
	headers["Origin"] = "https://rdcloud.4c.hq.cmcc"

	// Form data
	formData := url.Values{}
	formData.Set("objIds", strconv.FormatInt(objID, 10))

	baseDocURL := "https://rdcloud.4c.hq.cmcc"

	apiReq := &Request{
		URL:     baseDocURL + path,
		Method:  http.MethodPost,
		Headers: headers,
		Body:    strings.NewReader(formData.Encode()),
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var result DeleteDmsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w, body: %s", err, string(body))
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: code %d, body: %s", result.Code, string(body))
	}

	return &result, nil
}

// PreUploadResponse represents the pre-upload response
type PreUploadResponse struct {
	Code    int           `json:"code"`
	Data    PreUploadData `json:"data"`
	List    []interface{} `json:"list"`
	TrackID string        `json:"trackId"`
	PodName string        `json:"podName"`
}

// PreUploadData contains upload URLs and key
type PreUploadData struct {
	SingleLimitSize     int64    `json:"singleLimitSize"`
	OnceLimitSize       int64    `json:"onceLimitSize"`
	UploadKey           string   `json:"uploadKey"`
	UploadURL           string   `json:"uploadUrl"`
	LengthURL           string   `json:"lengthUrl"`
	BefUploadURL        string   `json:"befUploadUrl"`
	SplitCheckURL       string   `json:"splitCheckUrl"`
	SplitUploadURL      string   `json:"splitUploadUrl"`
	SplitMergeURL       string   `json:"splitMergeUrl"`
	SplitMergeFormURL   string   `json:"splitMergeFormUrl"`
	SplitAsyncMergeURL  string   `json:"splitAsyncMergeUrl"`
	SplitMergeStatusURL string   `json:"splitMergeStatusUrl"`
	SplitMergeUploadURL string   `json:"splitMergeUploadUrl"`
	WhiteExtList        []string `json:"whiteExtList"`
	BlackExtList        []string `json:"blackExtList"`
}

// PreUpload gets upload key and URLs for file upload
func (s *DocService) PreUpload(folderID int64) (*PreUploadResponse, error) {
	// Get client IP first
	clientIP, err := s.getClientIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get client IP: %w", err)
	}

	// Get OAuth authorization (use folderID for permission validation)
	authToken, err := s.getDocAuthToken(clientIP, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	path := "/cmdevops-doc-editor-cloud/document/preUpload"

	headers := make(map[string]string)
	for k, v := range s.Headers {
		headers[k] = v
	}
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	headers["Authorization"] = authToken
	headers["x-client-ip"] = clientIP
	headers["Referer"] = fmt.Sprintf("https://rdcloud.4c.hq.cmcc/base/cmdevops-doc-editor-web/document-space/document-library?folderId=%d&type=space&clientIp=%s&source=home&showBack=true", folderID, clientIP)
	headers["Origin"] = "https://rdcloud.4c.hq.cmcc"

	// Form data
	formData := url.Values{}
	formData.Set("id", strconv.FormatInt(folderID, 10))
	formData.Set("uploadObjType", "1")

	baseDocURL := "https://rdcloud.4c.hq.cmcc"

	apiReq := &Request{
		URL:     baseDocURL + path,
		Method:  http.MethodPost,
		Headers: headers,
		Body:    strings.NewReader(formData.Encode()),
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var result PreUploadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: code %d", result.Code)
	}

	return &result, nil
}

// UploadFileResponse represents the file upload response
type UploadFileResponse struct {
	Code    int            `json:"code"`
	List    []UploadedFile `json:"list"`
	TrackID string         `json:"trackId"`
	PodName string         `json:"podName"`
}

// UploadedFile represents uploaded file info
type UploadedFile struct {
	PrtID       int64  `json:"prtId"`
	DocID       int64  `json:"docId"`
	RevID       int64  `json:"revId"`
	RevCode     int    `json:"revCode"`
	FileName    string `json:"fileName"`
	Status      int    `json:"status"`
	DirPath     string `json:"dirPath"`
	IntFileName string `json:"intFileName"`
	SizeStr     string `json:"sizeStr"`
}

// UploadFile uploads a file to the document library
func (s *DocService) UploadFile(uploadURL string, filePath string, fileName string) (*UploadFileResponse, error) {
	// Read file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Create multipart form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Create form file field
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Write file content
	if _, err := part.Write(fileContent); err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Prepare headers
	headers := map[string]string{
		"Content-Type":    writer.FormDataContentType(),
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh-TW;q=0.9,zh;q=0.8,en-US;q=0.7,en;q=0.6",
		"Cache-Control":   "no-cache",
		"Origin":          "https://rdcloud.4c.hq.cmcc",
		"Pragma":          "no-cache",
		"Referer":         "https://rdcloud.4c.hq.cmcc/",
		"Sec-Fetch-Dest":  "empty",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Site":  "same-site",
		"User-Agent":      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.6478.251 Safari/537.36 UOS Professional",
	}

	// Send request using HTTP client
	resp, err := s.HTTPClient.Send(&Request{
		URL:     uploadURL,
		Method:  http.MethodPost,
		Headers: headers,
		Body:    &body,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(respBody))
	}

	var result UploadFileResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API error: code %d", result.Code)
	}

	return &result, nil
}
