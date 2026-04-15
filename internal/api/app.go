package api

import (
	"encoding/json"
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

// MyApp 我的应用信息
type MyApp struct {
	AppID           string `json:"appId"`
	AppName         string `json:"appName"`
	AppLevel        string `json:"appLevel"`
	Status          string `json:"status"`
	ShowStatusName  string `json:"showStatusName"`
	AuditStatus     string `json:"auditStatus"`
	MaxQuotaNum     int    `json:"maxQuotaNum"`
	AppImgPath      string `json:"appImgPath"`
	UserID          string `json:"userId"`
	Remark          string `json:"remark"`
}

// MyAppPage 应用分页对象
type MyAppPage struct {
	PageNum   int     `json:"pageNum"`
	PageSize  int     `json:"pageSize"`
	Total     int     `json:"total"`
	Pages     int     `json:"pages"`
	List      []MyApp `json:"list"`
}

// MyAppListResponse 我的应用列表响应
type MyAppListResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		AppList MyAppPage `json:"appList"`
	} `json:"data"`
}

// ListMyApps 查询我的应用列表
func (s *AppService) ListMyApps(page, size int, appName string) ([]MyApp, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	body := fmt.Sprintf("pgnum=%d&pgsize=%d&appName=%s", page, size, appName)
	contentType := "application/x-www-form-urlencoded"

	resp, err := s.client.PostMultipart("/openportalsrv/rest/portaluser/appManager/qryMyAppList", contentType, []byte(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result MyAppListResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != "00000" {
		return nil, fmt.Errorf("API error [%s]: %s", result.Code, result.Msg)
	}

	return result.Data.AppList.List, nil
}

// AuthAbilityRequest 应用授权能力请求
type AuthAbilityRequest struct {
	AppID          string `json:"appId"`
	AbilityID      string `json:"-"`
	AppName        string `json:"appName"`
	AuthType       string `json:"authType"`        // 固定 "capacity"
	Status         string `json:"status"`          // 固定 "AppStatusOnline"
	BomcID         string `json:"bomcId"`
	QuotaLimit     string `json:"-"`
	LimitCount     string `json:"-"`
	PolicyPeriod   string `json:"-"`
	PolicyTimeUnit string `json:"-"`
	GoodsNames     string `json:"goodsNames"`
}

// AuthAbilityResponse 授权响应
type AuthAbilityResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		OrderID string `json:"orderId"`
	} `json:"data"`
}

// AuthAbility 为应用授权能力
func (s *AppService) AuthAbility(req *AuthAbilityRequest) (*AuthAbilityResponse, error) {
	body := appAuthAbilityPayload{
		AppID:              req.AppID,
		OrderedGoodList:    []string{},
		NewOrderedGoodList: []string{req.AbilityID},
		CrmOrderList:       []string{},
		Status:             req.Status,
		AuthType:           req.AuthType,
		BomcID:             req.BomcID,
		LimitAndQuotaData: []limitAndQuotaItem{
			{
				ID:             req.AbilityID,
				Name:           req.GoodsNames,
				AppID:          req.AppID,
				AppName:        req.AppName,
				Type:           "capacity",
				QuotaLimit:     req.QuotaLimit,
				LimitCount:     req.LimitCount,
				PolicyPeriod:   req.PolicyPeriod,
				PolicyTimeUnit: req.PolicyTimeUnit,
			},
		},
		AppNames:   req.AppName,
		GoodsNames: req.GoodsNames,
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	resp, err := s.client.PostMultipart("/openportalsrv/rest/portaluser/appManager/doAppOeder", "application/json", jsonBytes)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result AuthAbilityResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != "00000" {
		return nil, fmt.Errorf("API error [%s]: %s", result.Code, result.Msg)
	}

	return &result, nil
}
