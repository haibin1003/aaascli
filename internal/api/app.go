package api

import (
	"fmt"
)

// AppService 应用服务
type AppService struct {
	client *Client
}

// NewAppService 创建应用服务
func NewAppService(client *Client) *AppService {
	return &AppService{client: client}
}

// App 应用信息
type App struct {
	ID          string `json:"id"`
	Name        string `json:"appName"`        // 应用名称
	Code        string `json:"appCode"`        // 应用编码
	Status      string `json:"status"`         // 状态：通过、待审批等
	Description string `json:"description"`    // 描述
	CreateTime  string `json:"createTime"`
	UpdateTime  string `json:"updateTime"`
}

// AppListResponse 应用列表响应
type AppListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    struct {
		List     []App `json:"list"`
		Total    int   `json:"total"`
		PageNum  int   `json:"pageNum"`
		PageSize int   `json:"pageSize"`
	} `json:"data"`
}

// List 获取应用列表
func (s *AppService) List(page, size int) (*AppListResponse, error) {
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

	resp, err := s.client.Post("/openProtal/app/list", req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AppListResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &result, nil
}

// GetDetail 获取应用详情
func (s *AppService) GetDetail(appID string) (*App, error) {
	path := fmt.Sprintf("/openProtal/app/detail/%s", appID)
	resp, err := s.client.Get(path)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    App    `json:"data"`
	}
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if !result.Success && result.Code != 200 {
		return nil, fmt.Errorf("API error [%d]: %s", result.Code, result.Message)
	}

	return &result.Data, nil
}

// AuthAbilityRequest 能力授权请求
type AuthAbilityRequest struct {
	AppID          string `json:"appId"`
	AbilityID      string `json:"abilityId"`
	BomcID         string `json:"bomcId"`              // BOMC 工单编码（必填）
	DailyLimit     int    `json:"dailyLimit,omitempty"` // 日调用量上限
	RateLimit      int    `json:"rateLimit,omitempty"`  // 流控限额
	RateLimitPeriod string `json:"rateLimitPeriod,omitempty"` // 流控周期
}

// AuthAbilityResponse 能力授权响应
type AuthAbilityResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    struct {
		AuthID     string `json:"authId"`
		Status     string `json:"status"`     // pending, approved, rejected
		NeedVerify bool   `json:"needVerify"` // 是否需要审批
	} `json:"data"`
}

// AuthAbility 能力授权
func (s *AppService) AuthAbility(req *AuthAbilityRequest) (*AuthAbilityResponse, error) {
	resp, err := s.client.Post("/openProtal/app/authAbility", req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AuthAbilityResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &result, nil
}

// GetAuthAbilities 获取应用已授权的能力列表
type AuthAbility struct {
	AuthID     string `json:"authId"`
	AbilityID  string `json:"abilityId"`
	AbilityName string `json:"abilityName"`
	AbilityCode string `json:"abilityCode"`
	Status     string `json:"status"`      // 授权状态
	ApplyTime  string `json:"applyTime"`
	ApproveTime string `json:"approveTime,omitempty"`
}

// AuthAbilityListResponse 授权能力列表响应
type AuthAbilityListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    struct {
		List     []AuthAbility `json:"list"`
		Total    int           `json:"total"`
	} `json:"data"`
}

// ListAuthAbilities 获取应用已授权的能力列表
func (s *AppService) ListAuthAbilities(appID string) (*AuthAbilityListResponse, error) {
	req := map[string]interface{}{
		"appId": appID,
	}

	resp, err := s.client.Post("/openProtal/app/authAbilityList", req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AuthAbilityListResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &result, nil
}

// GetAvailableAbilities 获取可授权的能力列表（已订购但未授权）
func (s *AppService) GetAvailableAbilities(appID string) (*AbilityListResponse, error) {
	req := map[string]interface{}{
		"appId": appID,
	}

	resp, err := s.client.Post("/openProtal/app/availableAbilities", req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AbilityListResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &result, nil
}

// GetAuthStatus 获取授权审批状态
type AuthStatusResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    struct {
		Pending  []AuthAbility `json:"pending"`  // 待审批
		Approved []AuthAbility `json:"approved"` // 已通过
		Rejected []AuthAbility `json:"rejected"` // 已拒绝
	} `json:"data"`
}

// GetAuthStatus 获取授权状态
func (s *AppService) GetAuthStatus(appID string) (*AuthStatusResponse, error) {
	req := map[string]interface{}{
		"appId": appID,
	}

	resp, err := s.client.Post("/openProtal/app/authStatus", req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AuthStatusResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &result, nil
}
