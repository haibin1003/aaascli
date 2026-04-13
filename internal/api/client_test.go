package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	require.NotNil(t, client)
	require.NotNil(t, client.HTTPClient)
	require.NotNil(t, client.Logger)

	// 验证超时设置
	assert.Equal(t, DefaultTimeout, client.HTTPClient.Timeout)

	// 验证使用了共享transport
	assert.Equal(t, defaultTransport, client.HTTPClient.Transport)
}

func TestNewInsecureClient(t *testing.T) {
	client := NewInsecureClient()
	require.NotNil(t, client)
	require.NotNil(t, client.HTTPClient)

	// 验证使用了共享的insecure transport
	assert.Equal(t, insecureTransport, client.HTTPClient.Transport)
}

func TestClient_Send(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		assert.Equal(t, http.MethodPost, r.Method)

		// 验证请求头
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-value", r.Header.Get("X-Test-Header"))

		// 返回响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewClient()
	req := &Request{
		URL:    server.URL,
		Method: http.MethodPost,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Test-Header": "test-value",
		},
		Body: map[string]string{"key": "value"},
	}

	resp, err := client.Send(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_Send_GetRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	req := &Request{
		URL:    server.URL,
		Method: http.MethodGet,
	}

	resp, err := client.Send(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_Send_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	client := NewClient()
	req := &Request{
		URL:    server.URL,
		Method: http.MethodGet,
	}

	resp, err := client.Send(req)
	require.NoError(t, err) // Send不检查HTTP状态码
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestClient_Send_InvalidURL(t *testing.T) {
	client := NewClient()
	req := &Request{
		URL:    "://invalid-url",
		Method: http.MethodGet,
	}

	_, err := client.Send(req)
	assert.Error(t, err)
}

func TestDefaultTimeout(t *testing.T) {
	assert.Equal(t, 30*time.Second, DefaultTimeout)
}

func TestSharedTransport(t *testing.T) {
	// 验证共享transport实例
	client1 := NewClient()
	client2 := NewClient()

	assert.Equal(t, client1.HTTPClient.Transport, client2.HTTPClient.Transport)
	assert.Equal(t, defaultTransport, client1.HTTPClient.Transport)
}

func TestInsecureSharedTransport(t *testing.T) {
	// 验证共享insecure transport实例
	client1 := NewInsecureClient()
	client2 := NewInsecureClient()

	assert.Equal(t, client1.HTTPClient.Transport, client2.HTTPClient.Transport)
	assert.Equal(t, insecureTransport, client1.HTTPClient.Transport)
}
