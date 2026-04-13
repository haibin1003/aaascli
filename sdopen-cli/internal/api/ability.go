package api

import (
	"fmt"
)

// AbilityService 能力服务
type AbilityService struct {
	client *Client
}

// NewAbilityService 创建能力服务
func NewAbilityService(client *Client) *AbilityService {
	return &AbilityService{client: client}
}

// Ability 能力信息（对外服务）
type Ability struct {
	ID          string `json:"id"`
	Name        string `json:"abilityName"`        // 能力名称
	Code        string `json:"abilityCode"`        // 能力编码
	Category    string `json:"categoryName"`       // 分类名称
	Provider    string `json:"providerName"`       // 提供方
	Description string `json:"abilityDescription"` // 能力描述
	Status      string `json:"status"`             // 状态
}

// AbilityDetail 能力详情
type AbilityDetail struct {
	ID             string   `json:"id"`
	Name           string   `json:"abilityName"`
	Code           string   `json:"abilityCode"`
	Category       string   `json:"categoryName"`
	CategoryID     string   `json:"categoryId"`
	Provider       string   `json:"providerName"`
	ProviderID     string   `json:"providerId"`
	Description    string   `json:"abilityDescription"`
	DetailDesc     string   `json:"detailDescription"`
	Status         string   `json:"status"`
	APIInfo        *APIInfo `json:"apiInfo,omitempty"`
	CreateTime     string   `json:"createTime"`
	UpdateTime     string   `json:"updateTime"`
}

// APIInfo API 信息
type APIInfo struct {
	Method         string `json:"method"`
	URL            string `json:"url"`
	RequestFormat  string `json:"requestFormat"`
	ResponseFormat string `json:"responseFormat"`
}

// AbilityListRequest 能力列表请求
type AbilityListRequest struct {
	PageNum   int    `json:"pageNum,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
	Keyword   string `json:"keyword,omitempty"`
	Type      string `json:"type,omitempty"`      // 服务类型: internal, external, network
	ServiceID string `json:"serviceId,omitempty"` // 数字服务 ID
}

// AbilityListResponse 能力列表响应
type AbilityListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    struct {
		List     []Ability `json:"list"`
		Total    int       `json:"total"`
		PageNum  int       `json:"pageNum"`
		PageSize int       `json:"pageSize"`
	} `json:"data"`
}

// AbilityDetailResponse 能力详情响应
type AbilityDetailResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Success bool          `json:"success"`
	Data    AbilityDetail `json:"data"`
}

// List 获取能力列表（对外服务 - 能力广场）
func (s *AbilityService) List(req *AbilityListRequest) (*AbilityListResponse, error) {
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	resp, err := s.client.Post("/openProtal/ability/list", req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AbilityListResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &result, nil
}

// Search 搜索能力
func (s *AbilityService) Search(keyword string, page, size int) (*AbilityListResponse, error) {
	return s.List(&AbilityListRequest{
		PageNum:  page,
		PageSize: size,
		Keyword:  keyword,
	})
}

// GetDetail 获取能力详情
func (s *AbilityService) GetDetail(abilityID string) (*AbilityDetail, error) {
	path := fmt.Sprintf("/openProtal/ability/detail/%s", abilityID)
	resp, err := s.client.Get(path)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AbilityDetailResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if !result.Success && result.Code != 200 {
		return nil, fmt.Errorf("API error [%d]: %s", result.Code, result.Message)
	}

	return &result.Data, nil
}

// OrderRequest 订购请求
type OrderRequest struct {
	AbilityID string `json:"abilityId"`
	AppID     string `json:"appId,omitempty"`
}

// OrderResponse 订购响应
type OrderResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    struct {
		OrderID string `json:"orderId"`
		Status  string `json:"status"`
	} `json:"data"`
}

// Order 订购能力
func (s *AbilityService) Order(abilityID string) (*OrderResponse, error) {
	req := &OrderRequest{
		AbilityID: abilityID,
	}

	resp, err := s.client.Post("/openProtal/ability/order", req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result OrderResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &result, nil
}

// GetMyAbilities 获取我的能力列表（已订购）
func (s *AbilityService) GetMyAbilities(page, size int) (*AbilityListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	req := map[string]interface{}{
		"pageNum":  page,
		"pageSize": size,
	}

	resp, err := s.client.Post("/openProtal/ability/myList", req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AbilityListResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &result, nil
}
