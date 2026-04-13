package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockServer 提供 HTTP API 的模拟服务
type MockServer struct {
	Server   *httptest.Server
	Handlers map[string]http.HandlerFunc
}

// NewMockServer 创建新的模拟服务器
func NewMockServer() *MockServer {
	m := &MockServer{
		Handlers: make(map[string]http.HandlerFunc),
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 根据路径和方法查找处理器
		key := r.Method + " " + r.URL.Path
		if handler, ok := m.Handlers[key]; ok {
			handler(w, r)
			return
		}

		// 默认返回 404
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "NOT_FOUND",
				"message": "mock handler not found for " + key,
			},
		})
	}))

	return m
}

// RegisterHandler 注册路径处理器
func (m *MockServer) RegisterHandler(method, path string, handler http.HandlerFunc) {
	key := method + " " + path
	m.Handlers[key] = handler
}

// RegisterJSONHandler 注册返回 JSON 的处理器
func (m *MockServer) RegisterJSONHandler(method, path string, statusCode int, response interface{}) {
	m.RegisterHandler(method, path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	})
}

// Close 关闭模拟服务器
func (m *MockServer) Close() {
	m.Server.Close()
}

// URL 返回服务器 URL
func (m *MockServer) URL() string {
	return m.Server.URL
}

// TestWithMockServer 使用示例
func TestWithMockServer(t *testing.T) {
	mock := NewMockServer()
	defer mock.Close()

	// 注册模拟响应
	mock.RegisterJSONHandler(http.MethodGet, "/api/test", http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]string{
			"message": "hello from mock",
		},
	})

	// 使用模拟服务器进行测试
	client := NewClient()
	req := &Request{
		URL:    mock.URL() + "/api/test",
		Method: http.MethodGet,
	}

	resp, err := client.Send(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "hello from mock")
}

// MockRequirementService 提供需求服务的模拟
type MockRequirementService struct {
	Server *MockServer
}

// NewMockRequirementService 创建需求服务模拟
func NewMockRequirementService() *MockRequirementService {
	m := &MockRequirementService{
		Server: NewMockServer(),
	}

	// 注册常用端点
	m.registerCommonEndpoints()

	return m
}

func (m *MockRequirementService) registerCommonEndpoints() {
	// 列出需求
	m.Server.RegisterJSONHandler(http.MethodPost, "/api/requirements/list", http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"objectId":    "test-req-1",
					"name":        "测试需求1",
					"status":      "新建",
					"key":         "REQ-001",
					"assignee":    "测试用户",
					"createdDate": "2024-01-01T00:00:00Z",
				},
				{
					"objectId":    "test-req-2",
					"name":        "测试需求2",
					"status":      "进行中",
					"key":         "REQ-002",
					"assignee":    "测试用户",
					"createdDate": "2024-01-02T00:00:00Z",
				},
			},
			"totalCount": 2,
		},
	})

	// 创建需求
	m.Server.RegisterHandler(http.MethodPost, "/api/requirements/create", func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"objectId": "new-req-" + generateRandomID(),
				"name":     reqBody["name"],
				"key":      "REQ-NEW-001",
				"status":   "新建",
			},
		})
	})

	// 删除需求
	m.Server.RegisterJSONHandler(http.MethodPost, "/api/requirements/delete", http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"deleted": 1,
		},
	})
}

// Close 关闭模拟服务
func (m *MockRequirementService) Close() {
	m.Server.Close()
}

func generateRandomID() string {
	// 简单实现，实际使用时可导入 nanoid
	return "abc123"
}

// TestMockRequirementList 测试模拟的需求列表
func TestMockRequirementList(t *testing.T) {
	mock := NewMockRequirementService()
	defer mock.Close()

	// 创建客户端并指向模拟服务器
	client := NewClient()
	req := &Request{
		URL:    mock.Server.URL() + "/api/requirements/list",
		Method: http.MethodPost,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"pageNo":   1,
			"pageSize": 10,
		},
	}

	resp, err := client.Send(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, true, result["success"])

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)

	items, ok := data["items"].([]interface{})
	require.True(t, ok)
	assert.Len(t, items, 2)
}
