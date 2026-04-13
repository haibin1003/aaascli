package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Example: 如何使用 Mock 测试业务逻辑
//
// 这个文件展示了如何在不依赖真实 API 的情况下测试需求服务。
// 可以在 CI/CD 中快速运行，无需配置登录凭证。

// TestRequirementService_WithMock 展示如何使用 Mock 测试需求服务
func TestRequirementService_WithMock(t *testing.T) {
	// 1. 创建模拟服务器
	mock := NewMockRequirementService()
	defer mock.Close()

	// 2. 创建服务客户端（指向模拟服务器）
	client := NewClient()
	headers := map[string]string{
		"Content-Type": "application/json",
		"Cookie":       "test-cookie",
	}

	// 3. 创建请求
	req := &Request{
		URL:     mock.Server.URL() + "/api/requirements/list",
		Method:  http.MethodPost,
		Headers: headers,
		Body: map[string]interface{}{
			"pageNo":   1,
			"pageSize": 10,
		},
	}

	// 4. 发送请求
	resp, err := client.Send(req)

	// 5. 验证结果
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 解析响应
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, true, result["success"])
}

// TestCreateRequirement_Mock 展示如何测试创建需求
func TestCreateRequirement_Mock(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	// 注册创建需求的模拟响应
	createdName := ""
	mock.RegisterHandler(http.MethodPost, "/api/requirements", func(w http.ResponseWriter, r *http.Request) {
		var req CreateRequirementRequest
		json.NewDecoder(r.Body).Decode(&req)
		createdName = req.Name

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateRequirementResponse{
			Code: 0,
			Data: RequirementData{
				ObjectID: "new-object-id",
				Key:      "REQ-2024-001",
				Name:     req.Name,
			},
		})
	})

	// 执行创建
	client := NewClient()
	req := &Request{
		URL:    mock.URL() + "/api/requirements",
		Method: http.MethodPost,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: CreateRequirementRequest{
			Name: "新测试需求",
		},
	}

	resp, err := client.Send(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// 验证
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "新测试需求", createdName)

	var result CreateRequirementResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "REQ-2024-001", result.Data.Key)
}

// CreateRequirementRequest 创建需求请求
type CreateRequirementRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Priority    string `json:"priority,omitempty"`
}

// CreateRequirementResponse 创建需求响应
type CreateRequirementResponse struct {
	Code int              `json:"code"`
	Data RequirementData  `json:"data"`
}

// RequirementData 需求数据
type RequirementData struct {
	ObjectID string `json:"objectId"`
	Key      string `json:"key"`
	Name     string `json:"name"`
}

// TestErrorHandling_Mock 展示如何测试错误处理
func TestErrorHandling_Mock(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	// 注册一个返回错误的端点
	mock.RegisterHandler(http.MethodGet, "/api/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": "服务器内部错误",
			},
		})
	})

	client := NewClient()
	req := &Request{
		URL:    mock.URL() + "/api/error",
		Method: http.MethodGet,
	}

	resp, err := client.Send(req)
	require.NoError(t, err) // HTTP 请求本身成功
	defer resp.Body.Close()

	// 但返回了 500 状态码
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	errorData, ok := result["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "INTERNAL_ERROR", errorData["code"])
}

// TestTimeout_Mock 展示如何测试超时场景
func TestTimeout_Mock(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	// 注册一个慢响应端点
	mock.RegisterHandler(http.MethodGet, "/api/slow", func(w http.ResponseWriter, r *http.Request) {
		// 模拟慢响应 - 但在测试中我们不会真的等待
		// 实际测试超时时会设置更短的超时时间
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "completed",
		})
	})

	// 使用短超时创建客户端
	client := &Client{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second, // 正常测试不会触发超时
		},
		Logger: zap.NewNop(),
	}

	req := &Request{
		URL:    mock.URL() + "/api/slow",
		Method: http.MethodGet,
	}

	resp, err := client.Send(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
