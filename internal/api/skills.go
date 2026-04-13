package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// PluginMarketListRequest represents the query parameters for plugin market list
type PluginMarketListRequest struct {
	PageNum      int    `json:"pageNum"`      // Page number
	PageSize     int    `json:"pageSize"`     // Page size
	Name         string `json:"name"`         // Plugin name for search
	Type         string `json:"type"`         // Plugin type
	LabelID      string `json:"labelId"`      // Label ID for filtering
	Source       string `json:"source"`       // Source filter
	SortField    string `json:"sortField"`    // Sort field
	SortDir      string `json:"sortDir"`      // Sort direction
	HeartMarker  string `json:"heartMarker"`  // Heart marker (1 for favorited)
	Subscribe    string `json:"subscribe"`    // Subscribe filter (1 for subscribed)
}

// PluginListResponse represents the API response for plugin list
type PluginListResponse struct {
	Success  bool         `json:"success"`
	Code     string       `json:"code"`
	Message  string       `json:"message"`
	Data     []PluginInfo `json:"data"`
	PageNo   int          `json:"pageNo"`
	PageSize int          `json:"pageSize"`
	Count    int          `json:"count"`
	PageCount int         `json:"pageCount"`
	StartRow int          `json:"startRow"`
	EndRow   int          `json:"endRow"`
}

// PluginInfo represents a single plugin in the market list
type PluginInfo struct {
	ID                       string  `json:"id"`
	Name                     string  `json:"name"`
	Description              string  `json:"description"`
	Type                     string  `json:"type"`
	IconURL                  string  `json:"iconUrl"`
	CreateBy                 string  `json:"createBy"`
	TenantDept               string  `json:"tenantDept"`
	CreateTime               string  `json:"createTime"`
	FileSize                 string  `json:"fileSize"`
	Version                  string  `json:"version"`
	LabelNames               string  `json:"labelNames"`
	BrowseCount              int     `json:"browseCount"`
	DownloadCount            int     `json:"downloadCount"`
	Status                   *string `json:"status"`
	Provider                 string  `json:"provider"`
	DownloadURL              *string `json:"downloadUrl"`
	PackageID                string  `json:"packageId"`
	DesktopURL               string  `json:"desktopUrl"`
	DesktopType              string  `json:"desktopType"`
	IDEType                  string  `json:"ideType"`
	SubscribeStatusValue     *string `json:"subscribeStatusValue"`
	SubscribeStatusLabel     *string `json:"subscribeStatusLabel"`
	SubscribeAuditResultValue *string `json:"subscribeAuditResultValue"`
	SubscribeAuditResultLabel *string `json:"subscribeAuditResultLabel"`
	UnsubscribeStatusValue   *string `json:"unsubscribeStatusValue"`
	UnsubscribeStatusLabel   *string `json:"unsubscribeStatusLabel"`
	UnsubscribeAuditResultValue *string `json:"unsubscribeAuditResultValue"`
	UnsubscribeAuditResultLabel *string `json:"unsubscribeAuditResultLabel"`
	SubscribeCount           *int    `json:"subscribeCount"`
	HeartMarker              bool    `json:"heartMarker"`
}

// PluginInstallRequest represents the request body for installing a plugin
type PluginInstallRequest struct {
	PluginID    string `json:"pluginId"`
	Version     string `json:"version"`
}

// PluginInstallResponse represents the API response for plugin installation
type PluginInstallResponse struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PluginMarkBO represents plugin basic information in detail response
type PluginMarkBO struct {
	ID                         string  `json:"id"`
	Name                       string  `json:"name"`
	Description                string  `json:"description"`
	Type                       string  `json:"type"`
	IconURL                    string  `json:"iconUrl"`
	CreateBy                   string  `json:"createBy"`
	TenantDept                 *string `json:"tenantDept"`
	CreateTime                 string  `json:"createTime"`
	FileSize                   string  `json:"fileSize"`
	Version                    string  `json:"version"`
	LabelNames                 string  `json:"labelNames"`
	BrowseCount                int     `json:"browseCount"`
	DownloadCount              int     `json:"downloadCount"`
	Status                     *string `json:"status"`
	Provider                   string  `json:"provider"`
	DownloadURL                *string `json:"downloadUrl"`
	PackageID                  string  `json:"packageId"`
	DesktopURL                 string  `json:"desktopUrl"`
	DesktopType                string  `json:"desktopType"`
	IDEType                    string  `json:"ideType"`
	SubscribeStatusValue       *string `json:"subscribeStatusValue"`
	SubscribeStatusLabel       *string `json:"subscribeStatusLabel"`
	SubscribeAuditResultValue  *string `json:"subscribeAuditResultValue"`
	SubscribeAuditResultLabel  *string `json:"subscribeAuditResultLabel"`
	UnsubscribeStatusValue     *string `json:"unsubscribeStatusValue"`
	UnsubscribeStatusLabel     *string `json:"unsubscribeStatusLabel"`
	UnsubscribeAuditResultValue *string `json:"unsubscribeAuditResultValue"`
	UnsubscribeAuditResultLabel *string `json:"unsubscribeAuditResultLabel"`
	SubscribeCount             *int    `json:"subscribeCount"`
	HeartMarker                bool    `json:"heartMarker"`
	ShareCount                 int     `json:"shareCount"`
	HeartMarkerCount           int     `json:"heartMarkerCount"`
	DataScope                  string  `json:"dataScope"`
}

// PluginVersion represents version information in detail response
type PluginVersion struct {
	MinVersion        string `json:"minVersion"`
	ManifestContent   string `json:"manifestContent"`
	PluginID          string `json:"pluginId"`
	DesktopURL        string `json:"desktopUrl"`
	Description       string `json:"description"`
	ChangeLog         *string `json:"changeLog"`
	UpdateTime        string `json:"updateTime"`
	DesktopType       string `json:"desktopType"`
	Version           string `json:"version"`
	SuggestScope      *string `json:"suggestScope"`
	ProductFamilyName string `json:"productFamilyName"`
	Size              string `json:"size"`
	CreateTime        string `json:"createTime"`
	IDEPackageName    string `json:"idePackageName"`
	IntellijProductCode string `json:"intellijProductCode"`
	MaxVersion        *string `json:"maxVersion"`
	ID                string `json:"id"`
}

// PluginDetailData represents the data field in plugin detail response
type PluginDetailData struct {
	PluginMarkBO *PluginMarkBO   `json:"pluginMarkBO"`
	VersionList  []PluginVersion `json:"versionList"`
}

// PluginDetailResponse represents the API response for plugin detail
type PluginDetailResponse struct {
	Success bool              `json:"success"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Data    PluginDetailData   `json:"data"`
}

// SkillsService provides methods for skills/plugin management
type SkillsService struct {
	BaseService
}

// NewSkillsService creates a new SkillsService
func NewSkillsService(baseURL string, headers map[string]string, client HTTPClient) *SkillsService {
	return &SkillsService{
		BaseService: NewBaseService(baseURL, headers, client),
	}
}

// List lists available plugins/skills from the plugin market
func (s *SkillsService) List(searchKeyword string, pageNo, pageSize int) (*PluginListResponse, error) {
	// Build URL with query parameters using url.Values for proper encoding
	params := url.Values{}
	params.Set("type", "5")
	params.Set("name", searchKeyword)
	params.Set("pageNum", fmt.Sprintf("%d", pageNo))
	params.Set("pageSize", fmt.Sprintf("%d", pageSize))
	params.Set("sortField", "")
	params.Set("sortDir", "")

	fullURL := fmt.Sprintf("%s/cmdevops-plugin/plugin/market/list?%s", s.BaseURL, params.Encode())

	// Create request with custom headers
	apiReq := &Request{
		URL:     fullURL,
		Method:  http.MethodGet,
		Headers: s.Headers,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PluginListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}

// Install installs a plugin with specified version
func (s *SkillsService) Install(pluginID, version string) (*PluginInstallResponse, error) {
	// Build request
	request := &PluginInstallRequest{
		PluginID:    pluginID,
		Version:     version,
	}

	// Create request with custom headers
	apiReq := &Request{
		URL:     s.BaseURL + "/cmdevops-plugin/api/v1/plugin/install",
		Method:  http.MethodPost,
		Headers: s.Headers,
		Body:    request,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PluginInstallResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}

// Describe retrieves detailed information about a specific plugin
func (s *SkillsService) Describe(pluginType, pluginName, version string) (*PluginDetailResponse, error) {
	var fullURL string
	// Build URL: /cmdevops-plugin/plugin/market/v2/detail/{type}/{name}[/{version}]
	// Use PathEscape for URL path segments
	if version == "" {
		// No version specified
		fullURL = fmt.Sprintf("%s/cmdevops-plugin/plugin/market/v2/detail/%s/%s",
			s.BaseURL,
			url.PathEscape(pluginType),
			url.PathEscape(pluginName),
		)
	} else {
		// Version specified
		fullURL = fmt.Sprintf("%s/cmdevops-plugin/plugin/market/v2/detail/%s/%s/%s",
			s.BaseURL,
			url.PathEscape(pluginType),
			url.PathEscape(pluginName),
			url.PathEscape(version),
		)
	}

	// Create request with custom headers
	apiReq := &Request{
		URL:     fullURL,
		Method:  http.MethodGet,
		Headers: s.Headers,
	}

	resp, err := s.HTTPClient.Send(apiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PluginDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
}
