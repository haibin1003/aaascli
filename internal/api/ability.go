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

// Catalog 分类信息
type Catalog struct {
	CatalogID   string    `json:"catalogId"`
	CatalogName string    `json:"catalogName"`
	ChildList   []Catalog `json:"childList"`
	CapacityList []Ability `json:"capacityList"`
}

// Ability 能力信息
type Ability struct {
	ID           string `json:"capacityId"`
	Name         string `json:"capacityName"`
	Code         string `json:"capacityUniCode"`
	Desc         string `json:"capacityDesc"`
	Provider     string `json:"capacityProviderName"`
	Status       string `json:"authStatus"`
	TypeName     string `json:"capacityTypeName"`
	CallType     string `json:"capacityCallTypeName"`
	CatalogName  string `json:"catalogName"`
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

// CatalogListResponse 分类列表响应
type CatalogListResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		CapacityCatalogList []Catalog `json:"capacityCatalogList"`
	} `json:"data"`
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
		Product         AbilityDetail `json:"product"`
		CapacityDefine  AbilityDetail `json:"capacityDefineBean"`
	} `json:"data"`
}

// ListAll 从分类树中获取全部能力
func (s *AbilityService) ListAll() ([]Ability, error) {
	resp, err := s.client.Post("/openportalsrv/rest/portalmain/gwCapacityMgr/qryGwCapacityCatalogList", nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result CatalogListResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != "00000" {
		return nil, fmt.Errorf("API error [%s]: %s", result.Code, result.Msg)
	}

	var abilities []Ability
	for _, cat := range result.Data.CapacityCatalogList {
		for _, sub := range cat.ChildList {
			for _, ab := range sub.CapacityList {
				if ab.CatalogName == "" {
					ab.CatalogName = sub.CatalogName
				}
				abilities = append(abilities, ab)
			}
		}
	}

	return abilities, nil
}

// List 获取产品能力列表（默认走加密接口，返回产品）
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

// Search 搜索能力（基于全部能力做客户端过滤）
func (s *AbilityService) Search(keyword string) ([]Ability, error) {
	all, err := s.ListAll()
	if err != nil {
		return nil, err
	}

	var results []Ability
	for _, ab := range all {
		if containsAny(ab.Name, keyword) || containsAny(ab.Code, keyword) || containsAny(ab.CatalogName, keyword) {
			results = append(results, ab)
		}
	}
	return results, nil
}

func containsAny(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (containsCI(s, substr) || containsCI(substr, s))
}

func containsCI(a, b string) bool {
	if len(a) < len(b) {
		return false
	}
	for i := 0; i <= len(a)-len(b); i++ {
		if a[i:i+len(b)] == b {
			return true
		}
	}
	return false
}

// GetDetail 获取能力详情
func (s *AbilityService) GetDetail(abilityID string) (*AbilityDetail, error) {
	boundary := "----WebKitFormBoundaryoWjo5w3HbKA3wKEa"
	body := fmt.Sprintf(
		"------%s\r\nContent-Disposition: form-data; name=\"capacityId\"\r\n\r\n%s\r\n------%s--\r\n",
		boundary, abilityID, boundary,
	)
	contentType := fmt.Sprintf("multipart/form-data; boundary=----%s", boundary)

	resp, err := s.client.PostMultipart("/openportalsrv/rest/portalmain/capacityMgr/initCapacityInfo", contentType, []byte(body))
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

	// 优先使用 capacityDefineBean，回退到 product
	detail := result.Data.CapacityDefine
	if detail.ID == "" && detail.Name == "" {
		detail = result.Data.Product
	}
	return &detail, nil
}

// AbilityServiceItem 能力下的服务项
type AbilityServiceItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	ServiceType string `json:"serviceType"` // inner / outer
	CatalogName string `json:"catalogName"`
}

// AbilityServiceMenuResponse 能力下服务列表响应
type AbilityServiceMenuResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		ServiceMenus []AbilityServiceItem `json:"serviceMenus"`
	} `json:"data"`
}

// ListServices 查询能力下的服务列表
func (s *AbilityService) ListServices(abilityID string) ([]AbilityServiceItem, error) {
	body := fmt.Sprintf("capacityId=%s", abilityID)
	contentType := "application/x-www-form-urlencoded"

	resp, err := s.client.PostMultipart("/openportalsrv/rest/portalmain/capacityMgr/queryServiceMenuList", contentType, []byte(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AbilityServiceMenuResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != "00000" {
		return nil, fmt.Errorf("API error [%s]: %s", result.Code, result.Msg)
	}

	return result.Data.ServiceMenus, nil
}

// OrderAbility 订购能力
func (s *AbilityService) OrderAbility(abilityID string) error {
	return fmt.Errorf("订购接口尚未实现，请在浏览器中手动完成订购")
}
