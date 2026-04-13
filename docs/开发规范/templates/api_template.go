// API 客户端模板
// 使用说明：
// 1. 复制此文件到 internal/api/xxx.go
// 2. 将所有 "xxx" 替换为你的业务名
// 3. 填写实际的 API 端点和业务逻辑
// 4. 删除本注释

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/user/lc/internal/common"
	"go.uber.org/zap"
)

// 常量定义
const (
	DefaultXxxPageSize = 20
	MaxXxxPageSize     = 100
)

// ==================== 请求/响应结构体 ====================

// CreateXxxRequest 创建请求
type CreateXxxRequest struct {
	// 必填字段
	Name         string `json:"name"`
	WorkspaceKey string `json:"workspaceKey"`

	// 可选字段（使用 omitempty）
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateXxxRequest 更新请求
type UpdateXxxRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Status      string   `json:"status,omitempty"`
}

// XxxResponse 响应结构体
type XxxResponse struct {
	// 标识字段
	ObjectID string `json:"objectId"`
	Key      string `json:"key,omitempty"`

	// 基本信息
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`

	// 关联信息
	WorkspaceKey string `json:"workspaceKey"`
	ProjectCode  string `json:"projectCode,omitempty"`

	// 时间戳
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt,omitempty"`

	// 人员信息
	Creator  string `json:"creator,omitempty"`
	Assignee string `json:"assignee,omitempty"`
}

// ListXxxResponse 列表响应
type ListXxxResponse struct {
	Items      []XxxResponse `json:"items"`
	Total      int           `json:"total"`
	Page       int           `json:"page,omitempty"`
	PageSize   int           `json:"pageSize,omitempty"`
	TotalPages int           `json:"totalPages,omitempty"`
}

// ==================== API 方法 ====================

// ListXxx 查询列表
//
// Parameters:
//   - ctx: 命令上下文
//   - workspaceKey: 研发空间 Key
//   - page: 页码（从1开始）
//   - pageSize: 每页数量
//
// Returns:
//   - []XxxResponse: 列表数据
//   - int: 总数
//   - error: 错误信息
func ListXxx(ctx *common.CommandContext, workspaceKey string, page, pageSize int) ([]XxxResponse, int, error) {
	// 参数校验
	if workspaceKey == "" {
		return nil, 0, fmt.Errorf("workspaceKey is required")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > MaxXxxPageSize {
		pageSize = DefaultXxxPageSize
	}

	// 构建 URL
	url := fmt.Sprintf("%s/xxx/list?page=%d&size=%d",
		ctx.Config.API.BaseReqURL, page, pageSize)

	// 构建请求
	req := &Request{
		URL:     url,
		Method:  http.MethodGet,
		Headers: ctx.GetHeaders(workspaceKey),
	}

	// 记录日志
	ctx.Logger.Debug("Listing xxx",
		zap.String("workspace", workspaceKey),
		zap.Int("page", page),
		zap.Int("pageSize", pageSize),
	)

	// 发送请求
	resp, err := ctx.Client.Send(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var apiResp struct {
		Success bool            `json:"success"`
		Code    string          `json:"code"`
		Message string          `json:"message"`
		Data    ListXxxResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	// 处理业务错误
	if !apiResp.Success {
		return nil, 0, &APIError{
			Code:    apiResp.Code,
			Message: apiResp.Message,
		}
	}

	return apiResp.Data.Items, apiResp.Data.Total, nil
}

// GetXxx 查询详情
func GetXxx(ctx *common.CommandContext, id, workspaceKey string) (*XxxResponse, error) {
	if workspaceKey == "" {
		return nil, fmt.Errorf("workspaceKey is required")
	}
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	url := fmt.Sprintf("%s/xxx/detail?id=%s", ctx.Config.API.BaseReqURL, id)

	req := &Request{
		URL:     url,
		Method:  http.MethodGet,
		Headers: ctx.GetHeaders(workspaceKey),
	}

	ctx.Logger.Debug("Getting xxx", zap.String("id", id))

	resp, err := ctx.Client.Send(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Success bool        `json:"success"`
		Data    XxxResponse `json:"data"`
		Message string      `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, &APIError{
			Code:    apiResp.Code,
			Message: apiResp.Message,
		}
	}

	return &apiResp.Data, nil
}

// CreateXxx 创建
func CreateXxx(ctx *common.CommandContext, req CreateXxxRequest) (*XxxResponse, error) {
	// 参数校验
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.WorkspaceKey == "" {
		return nil, fmt.Errorf("workspaceKey is required")
	}

	// 构建 URL
	url := fmt.Sprintf("%s/xxx/create", ctx.Config.API.BaseReqURL)

	// 构建请求
	apiReq := &Request{
		URL:     url,
		Method:  http.MethodPost,
		Headers: ctx.GetHeaders(req.WorkspaceKey),
		Body:    req,
	}

	ctx.Logger.Debug("Creating xxx", zap.String("name", req.Name))

	// 发送请求
	resp, err := ctx.Client.Send(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var apiResp struct {
		Success bool        `json:"success"`
		Data    XxxResponse `json:"data"`
		Message string      `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, &APIError{
			Code:    apiResp.Code,
			Message: apiResp.Message,
		}
	}

	return &apiResp.Data, nil
}

// UpdateXxx 更新
func UpdateXxx(ctx *common.CommandContext, id string, req UpdateXxxRequest, workspaceKey string) (*XxxResponse, error) {
	if workspaceKey == "" {
		return nil, fmt.Errorf("workspaceKey is required")
	}
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	url := fmt.Sprintf("%s/xxx/update?id=%s", ctx.Config.API.BaseReqURL, id)

	apiReq := &Request{
		URL:     url,
		Method:  http.MethodPut, // 或 http.MethodPost
		Headers: ctx.GetHeaders(workspaceKey),
		Body:    req,
	}

	ctx.Logger.Debug("Updating xxx", zap.String("id", id))

	resp, err := ctx.Client.Send(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Success bool        `json:"success"`
		Data    XxxResponse `json:"data"`
		Message string      `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, &APIError{
			Code:    apiResp.Code,
			Message: apiResp.Message,
		}
	}

	return &apiResp.Data, nil
}

// DeleteXxx 删除
func DeleteXxx(ctx *common.CommandContext, id, workspaceKey string) error {
	if workspaceKey == "" {
		return fmt.Errorf("workspaceKey is required")
	}
	if id == "" {
		return fmt.Errorf("id is required")
	}

	url := fmt.Sprintf("%s/xxx/delete?id=%s", ctx.Config.API.BaseReqURL, id)

	req := &Request{
		URL:     url,
		Method:  http.MethodDelete,
		Headers: ctx.GetHeaders(workspaceKey),
	}

	ctx.Logger.Debug("Deleting xxx", zap.String("id", id))

	resp, err := ctx.Client.Send(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return &APIError{
			Code:    "DELETE_FAILED",
			Message: apiResp.Message,
		}
	}

	return nil
}

// SearchXxx 搜索
func SearchXxx(ctx *common.CommandContext, keyword, workspaceKey string, limit int) ([]XxxResponse, error) {
	if workspaceKey == "" {
		return nil, fmt.Errorf("workspaceKey is required")
	}
	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}
	if limit < 1 {
		limit = 10
	}

	url := fmt.Sprintf("%s/xxx/search?keyword=%s&limit=%d",
		ctx.Config.API.BaseReqURL, keyword, limit)

	req := &Request{
		URL:     url,
		Method:  http.MethodGet,
		Headers: ctx.GetHeaders(workspaceKey),
	}

	ctx.Logger.Debug("Searching xxx",
		zap.String("keyword", keyword),
		zap.Int("limit", limit),
	)

	resp, err := ctx.Client.Send(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Success bool            `json:"success"`
		Data    ListXxxResponse `json:"data"`
		Message string          `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return nil, &APIError{
			Code:    apiResp.Code,
			Message: apiResp.Message,
		}
	}

	return apiResp.Data.Items, nil
}
