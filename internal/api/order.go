package api

import (
	"fmt"
)

// ApplyItem 申请记录项
type ApplyItem struct {
	ID           string `json:"id"`
	AppName      string `json:"appName"`
	GoodsName    string `json:"goodsName"`
	AuthType     string `json:"authType"`     // capacity / api
	AuthTypeName string `json:"authTypeName"`
	Status       string `json:"status"`
	StatusName   string `json:"statusName"`
	ApplyTime    string `json:"applyTime"`
	PassStatus   string `json:"passStatus"`   // true / false
}

// ApplyListResponse 申请列表响应
type ApplyListResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		PageNum  int         `json:"pageNum"`
		PageSize int         `json:"pageSize"`
		Total    int         `json:"total"`
		Pages    int         `json:"pages"`
		List     []ApplyItem `json:"list"`
	} `json:"data"`
}

// ApplyService 订购/申请查询服务
type ApplyService struct {
	client *Client
}

// NewApplyService 创建申请查询服务客户端
func NewApplyService(client *Client) *ApplyService {
	return &ApplyService{client: client}
}

// ListMyApplies 查询我的申请列表
func (s *ApplyService) ListMyApplies(page, size int, passStatus bool) (*ApplyListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	statusStr := "false"
	if passStatus {
		statusStr = "true"
	}
	body := fmt.Sprintf("passStatus=%s&pgnum=%d&pgsize=%d", statusStr, page, size)
	contentType := "application/x-www-form-urlencoded"

	resp, err := s.client.PostMultipart("/openportalsrv/rest/portaluser/myApply/queryMyApplyList", contentType, []byte(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result ApplyListResponse
	if err := ParseJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != "00000" {
		return nil, fmt.Errorf("API error [%s]: %s", result.Code, result.Msg)
	}

	return &result, nil
}
