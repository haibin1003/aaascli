package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, []int{429, 502, 503, 504}, config.RetryableCodes)
	assert.Contains(t, config.RetryableMethods, http.MethodGet)
	assert.Contains(t, config.RetryableMethods, http.MethodPut)
	assert.Contains(t, config.RetryableMethods, http.MethodDelete)
}

func TestRetryConfig_IsRetryableCode(t *testing.T) {
	config := DefaultRetryConfig()

	tests := []struct {
		code     int
		expected bool
	}{
		{429, true},
		{502, true},
		{503, true},
		{504, true},
		{200, false},
		{404, false},
		{500, false},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.code), func(t *testing.T) {
			result := config.IsRetryableCode(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetryConfig_IsRetryableMethod(t *testing.T) {
	config := DefaultRetryConfig()

	tests := []struct {
		method   string
		expected bool
	}{
		{http.MethodGet, true},
		{http.MethodPost, false}, // POST is not idempotent
		{http.MethodPut, true},
		{http.MethodDelete, true},
		{http.MethodHead, true},
		{http.MethodPatch, false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := config.IsRetryableMethod(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetryConfig_CalculateBackoff(t *testing.T) {
	config := &RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
	}

	tests := []struct {
		attempt  int
		minDelay time.Duration
		maxDelay time.Duration
	}{
		{0, 0, 0},
		{1, 750 * time.Millisecond, 1250 * time.Millisecond},   // 1s ± 25%
		{2, 1500 * time.Millisecond, 2500 * time.Millisecond},  // 2s ± 25%
		{3, 3000 * time.Millisecond, 5000 * time.Millisecond},  // 4s ± 25%
		{10, 29000 * time.Millisecond, 30000 * time.Millisecond}, // capped at 30s
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.attempt)), func(t *testing.T) {
			delay := config.CalculateBackoff(tt.attempt)
			if tt.attempt == 0 {
				assert.Equal(t, time.Duration(0), delay)
			} else {
				assert.GreaterOrEqual(t, delay, tt.minDelay)
				assert.LessOrEqual(t, delay, tt.maxDelay)
			}
		})
	}
}

func TestRetryableClient_Do_Success(t *testing.T) {
	// Create a test server that always succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := &RetryableClient{
		HTTPClient:  http.DefaultClient,
		RetryConfig: DefaultRetryConfig(),
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRetryableClient_Do_RetryOnServerError(t *testing.T) {
	// Count requests
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable) // 503
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &RetryConfig{
		MaxRetries:     3,
		InitialDelay:   10 * time.Millisecond,
		MaxDelay:       100 * time.Millisecond,
		RetryableCodes: []int{503},
		RetryableMethods: []string{http.MethodGet},
	}

	client := &RetryableClient{
		HTTPClient:  http.DefaultClient,
		RetryConfig: config,
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, requestCount)
}

func TestRetryableClient_Do_NoRetryForNonRetryableMethod(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	config := &RetryConfig{
		MaxRetries:       3,
		RetryableMethods: []string{http.MethodGet}, // POST not included
		RetryableCodes:   []int{503},
		InitialDelay:     10 * time.Millisecond,
	}

	client := &RetryableClient{
		HTTPClient:  http.DefaultClient,
		RetryConfig: config,
	}

	req, _ := http.NewRequest(http.MethodPost, server.URL, nil)
	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, 1, requestCount) // No retries
}

func TestRetryableClient_Do_MaxRetriesExceeded(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	config := &RetryConfig{
		MaxRetries:       2,
		InitialDelay:     1 * time.Millisecond,
		RetryableCodes:   []int{503},
		RetryableMethods: []string{http.MethodGet},
	}

	client := &RetryableClient{
		HTTPClient:  http.DefaultClient,
		RetryConfig: config,
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)

	assert.NoError(t, err) // Returns the last response
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, 3, requestCount) // Initial + 2 retries
}

func TestNewEnhancedTransport(t *testing.T) {
	transport := NewEnhancedTransport(false)

	assert.Equal(t, 100, transport.MaxIdleConns)
	assert.Equal(t, 10, transport.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, transport.IdleConnTimeout)
	assert.Equal(t, 10*time.Second, transport.TLSHandshakeTimeout)
	assert.Equal(t, 10*time.Second, transport.ResponseHeaderTimeout)
	assert.True(t, transport.ForceAttemptHTTP2)
}

func TestNewEnhancedTransport_Insecure(t *testing.T) {
	transport := NewEnhancedTransport(true)

	assert.NotNil(t, transport.TLSClientConfig)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestResilientClient_SendWithRetry(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	client := NewResilientClient()
	client.retryConfig.InitialDelay = 1 * time.Millisecond

	req := &Request{
		Method: http.MethodGet,
		URL:    server.URL,
	}

	resp, err := client.SendWithRetry(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, requestCount)
}
