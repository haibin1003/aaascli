package api

import (
	"fmt"
	"net/url"
)

// AbilityService 能力服务
type AbilityService struct {
	client *Client
}

// NewAbilityService 创建能力服务
func NewAbilityService(client *Client) *AbilityService {
	return &AbilityService{client: client}
}

// Ability 能力信息
type Ability struct {
	ID          string `json:"capacityId"`
	Name        string `json:"capacityName"`
	Code        string `json:"capacityUniCode"`
	Desc        string `json:"capacityDesc"`
	Provider    string `json:"capacityProviderName"`
	Status      string `json:"authStatus"`
	TypeName    string `json:"capacityTypeName"`
	CallType    string `json:"capacityCallTypeName"`
}

// AbilityDetail 能力详情
type AbilityDetail struct {
	ID          string `json:"capacityId"`
	Name        string `json:"capacityName"`
	Code        string `json:"capacityUniCode"`
	Desc        string `json:"capacityDesc"`
	DetailDesc  string `json:"detailDescription"`
	Provider    string `json:"capacityProviderName"`
	TypeName    string `json:"capacityTypeName"`
	CallType    string `json:"capacityCallTypeName"`
	Img         string `json:"capacityImg"`
	UserID      string `json:"capacityUserId"`
}

// AbilityListResponse 能力列表响应
type AbilityListResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		ProductActionList []Ability `json:"productActionList"`
	} `json:"data"`
}

// AbilityDetailResponse 能力详情响应
type AbilityDetailResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		Product AbilityDetail `json:"product"`
	} `json:"data"`
}

// List 获取能力列表
func (s *AbilityService) List(page, size int) (*AbilityListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	req := map[string]interface{}{
		"pgnum":  page,
		"pgsize": size,
	}

	resp, err := s.client.PostEncrypted("/openportalsrv/rest/portalmain/productMgr/initProductList", req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AbilityListResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != "00000" {
		return nil, fmt.Errorf("API error [%s]: %s", result.Code, result.Msg)
	}

	return &result, nil
}

// Search 搜索能力
func (s *AbilityService) Search(keyword string, page, size int) (*AbilityListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	req := map[string]interface{}{
		"pgnum":   page,
		"pgsize":  size,
		"keyword": keyword,
	}

	resp, err := s.client.PostEncrypted("/openportalsrv/rest/portalmain/productMgr/initProductList", req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AbilityListResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != "00000" {
		return nil, fmt.Errorf("API error [%s]: %s", result.Code, result.Msg)
	}

	return &result, nil
}

// GetDetail 获取能力详情
func (s *AbilityService) GetDetail(abilityID string) (*AbilityDetail, error) {
	body := fmt.Sprintf("capacityId=%s", url.QueryEscape(abilityID))
	resp, err := s.client.Post("/openportalsrv/rest/portalmain/productMgr/qryProduct", body)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AbilityDetailResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != "00000" {
		return nil, fmt.Errorf("API error [%s]: %s", result.Code, result.Msg)
	}

	return &result.Data.Product, nil
}

// OrderAbility 订购能力
func (s *AbilityService) OrderAbility(abilityID string) error {
	// TODO: 需要找到真实的订购提交接口
	return fmt.Errorf("订购接口尚未实现，请在浏览器中手动完成订购")
}
