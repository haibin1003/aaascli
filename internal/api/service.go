package api

import (
	"fmt"
)

// CatalogNode 服务目录树节点（对应 qryApiCatalogListData 返回的 cataLogList）
type CatalogNode struct {
	CatalogID        string        `json:"catalogId"`
	CatalogName      string        `json:"catalogName"`
	CatalogType      string        `json:"catalogType"`
	CatalogLevel     string        `json:"catalogLevel"`
	IsLeaf           string        `json:"isLeaf"`
	SmallCatalogList []CatalogNode `json:"smallCatalogList"`
	APIList          []APIService  `json:"apiList"`
}

// APIService 目录中的具体 API 服务项
type APIService struct {
	APIID       string `json:"apiID"`
	Name        string `json:"name"`
	InterfaceID string `json:"interfaceId"`
	RequestType string `json:"requestType"`
	RequestURL  string `json:"requestUrl"`
	Status      string `json:"status"`
	Version     string `json:"version"`
}

// ServiceCatalogResponse 服务目录响应
type ServiceCatalogResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		CataLogList []CatalogNode `json:"cataLogList"`
	} `json:"data"`
}

// ServiceDetail 服务详情（对应 queryServiceInfo 的 serviceInfo）
type ServiceDetail struct {
	APIID           string `json:"apiId"`
	Name            string `json:"name"`
	APIVersion      string `json:"apiVersion"`
	RequestType     string `json:"requestType"`
	RequestTypeText string `json:"requestTypeText"`
	RequestURL      string `json:"requestUrl"`
	Remark          string `json:"remark"`
	RequestExample  string `json:"requestExample"`
	ResponseExample string `json:"responseExample"`
	Protocol        string `json:"protocol"`
	InterfaceID     string `json:"interfaceId"`
	ServiceID       string `json:"serviceId"`
	DomainName      string `json:"domainName"`
	Owner           string `json:"owner"`
	Department      string `json:"department"`
	ContactNo       string `json:"contactNo"`
}

// ServiceDetailResponse 服务详情响应
type ServiceDetailResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		ServiceInfo ServiceDetail `json:"serviceInfo"`
	} `json:"data"`
}

// ServiceService 数字服务
type ServiceService struct {
	client *Client
}

// NewServiceService 创建数字服务客户端
func NewServiceService(client *Client) *ServiceService {
	return &ServiceService{client: client}
}

// ListAll 查询全量服务目录列表
func (s *ServiceService) ListAll() (*ServiceCatalogResponse, error) {
	body := "parentId=APISHOWROOT&catalogType=APISHOW"
	contentType := "application/x-www-form-urlencoded"

	resp, err := s.client.PostMultipart("/openportalsrv/rest/portalmain/capacityMgr/qryApiCatalogListData", contentType, []byte(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result ServiceCatalogResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != "00000" {
		return nil, fmt.Errorf("API error [%s]: %s", result.Code, result.Msg)
	}

	return &result, nil
}

// GetDetail 查询服务详情
func (s *ServiceService) GetDetail(serviceID string) (*ServiceDetail, error) {
	body := fmt.Sprintf("serviceId=%s&belongType=apiGw&domainId=", serviceID)
	contentType := "application/x-www-form-urlencoded"

	resp, err := s.client.PostMultipart("/openportalsrv/rest/portalmain/capacityMgr/queryServiceInfo", contentType, []byte(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result ServiceDetailResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != "00000" {
		return nil, fmt.Errorf("API error [%s]: %s", result.Code, result.Msg)
	}

	return &result.Data.ServiceInfo, nil
}
