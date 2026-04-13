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

// AppAuth 授权信息
type AppAuth struct {
	AppName        string `json:"appName"`
	AbilityName    string `json:"abilityName"`
	AbilityCode    string `json:"abilityCode"`
	AuthStatus     string `json:"authStatus"`
	AuthStatusName string `json:"authStatusName"`
	ApplyTime      string `json:"applyTime"`
}

// AppAuthListResponse 授权列表响应
type AppAuthListResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		AuthorizedList []AppAuth `json:"authorizedList"`
	} `json:"data"`
}

// List 获取已授权列表（作为应用/能力授权视图）
func (s *AppService) List() (*AppAuthListResponse, error) {
	resp, err := s.client.Post("/openportalsrv/rest/portalmain/capacityMgr/qryAuthorizedList", nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AppAuthListResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != "00000" {
		return nil, fmt.Errorf("API error [%s]: %s", result.Code, result.Msg)
	}

	return &result, nil
}
