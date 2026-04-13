package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// RetryConfig configures retry behavior for HTTP requests
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (default: 3)
	MaxRetries int

	// InitialDelay is the initial delay between retries (default: 1s)
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries (default: 30s)
	MaxDelay time.Duration

	// RetryableCodes defines HTTP status codes that should trigger a retry
	RetryableCodes []int

	// RetryableMethods defines HTTP methods that can be safely retried
	RetryableMethods []string
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialDelay:   1 * time.Second,
		MaxDelay:       30 * time.Second,
		RetryableCodes: []int{429, 502, 503, 504},
		RetryableMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodOptions,
			http.MethodPut,
			http.MethodDelete,
		},
	}
}

// IsRetryableCode checks if the given status code should trigger a retry
func (rc *RetryConfig) IsRetryableCode(code int) bool {
	for _, retryableCode := range rc.RetryableCodes {
		if code == retryableCode {
			return true
		}
	}
	return false
}

// IsRetryableMethod checks if the given HTTP method can be safely retried
func (rc *RetryConfig) IsRetryableMethod(method string) bool {
	for _, retryableMethod := range rc.RetryableMethods {
		if method == retryableMethod {
			return true
		}
	}
	return false
}

// CalculateBackoff calculates the delay for the next retry attempt using
// exponential backoff with jitter
func (rc *RetryConfig) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	// Exponential backoff: base * 2^(attempt-1)
	delay := float64(rc.InitialDelay) * math.Pow(2, float64(attempt-1))

	// Add jitter (±25% randomization) to prevent thundering herd
	jitter := delay * 0.25 * (2*rand.Float64() - 1)
	delay = delay + jitter

	// Cap at max delay
	if delay > float64(rc.MaxDelay) {
		delay = float64(rc.MaxDelay)
	}

	return time.Duration(delay)
}

// NewEnhancedTransport creates an HTTP transport with production-grade settings
func NewEnhancedTransport(insecure bool) *http.Transport {
	transport := &http.Transport{
		// Connection pool settings
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,

		// Timeout settings
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		// Performance settings
		ForceAttemptHTTP2:  true,
		DisableKeepAlives:  false,
		DisableCompression: false,
	}

	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return transport
}

// RetryableClient wraps an HTTP client with retry functionality
type RetryableClient struct {
	HTTPClient  *http.Client
	RetryConfig *RetryConfig
	Logger      *zap.Logger
}

// Do executes an HTTP request with retry logic
func (rc *RetryableClient) Do(req *http.Request) (*http.Response, error) {
	if rc.RetryConfig == nil {
		rc.RetryConfig = DefaultRetryConfig()
	}

	// Check if method is retryable
	canRetry := rc.RetryConfig.IsRetryableMethod(req.Method)

	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt <= rc.RetryConfig.MaxRetries; attempt++ {
		// Clone request for retry (body can only be read once)
		var reqClone *http.Request
		if attempt > 0 {
			// For retries with body, we need to recreate the request
			// This is a limitation - retry works best for GET/HEAD/DELETE
			reqClone = req.Clone(req.Context())
		} else {
			reqClone = req
		}

		// Execute request
		resp, lastErr = rc.HTTPClient.Do(reqClone)

		// Success case
		if lastErr == nil && !rc.RetryConfig.IsRetryableCode(resp.StatusCode) {
			if attempt > 0 && rc.Logger != nil {
				rc.Logger.Debug("Request succeeded after retry",
					zap.Int("attempt", attempt),
					zap.String("url", req.URL.String()),
				)
			}
			return resp, nil
		}

		// Check if we should retry
		if !canRetry {
			if lastErr != nil {
				return nil, lastErr
			}
			return resp, nil
		}

		// Check if this was the last attempt
		if attempt >= rc.RetryConfig.MaxRetries {
			break
		}

		// Calculate backoff delay
		delay := rc.RetryConfig.CalculateBackoff(attempt + 1)

		// Log retry attempt
		if rc.Logger != nil {
			if lastErr != nil {
				rc.Logger.Warn("Request failed, retrying",
					zap.Int("attempt", attempt+1),
					zap.Int("maxRetries", rc.RetryConfig.MaxRetries),
					zap.Duration("delay", delay),
					zap.Error(lastErr),
					zap.String("url", req.URL.String()),
				)
			} else {
				rc.Logger.Warn("Retryable response received, retrying",
					zap.Int("attempt", attempt+1),
					zap.Int("maxRetries", rc.RetryConfig.MaxRetries),
					zap.Duration("delay", delay),
					zap.Int("statusCode", resp.StatusCode),
					zap.String("url", req.URL.String()),
				)
				// Close response body before retry
				if resp != nil && resp.Body != nil {
					resp.Body.Close()
				}
			}
		}

		// Wait before retry
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", rc.RetryConfig.MaxRetries, lastErr)
	}

	return resp, nil
}

// ResilientClient is an enhanced HTTP client with retry and timeout capabilities
type ResilientClient struct {
	*Client
	retryConfig *RetryConfig
}

// NewResilientClient creates a new client with enhanced resilience
func NewResilientClient() *ResilientClient {
	transport := NewEnhancedTransport(false)

	logger, _ := zap.NewProduction()

	return &ResilientClient{
		Client: &Client{
			HTTPClient: &http.Client{
				Timeout:   DefaultTimeout,
				Transport: transport,
			},
			Logger: logger,
		},
		retryConfig: DefaultRetryConfig(),
	}
}

// NewResilientInsecureClient creates an insecure client with enhanced resilience
func NewResilientInsecureClient() *ResilientClient {
	transport := NewEnhancedTransport(true)

	logger, _ := zap.NewProduction()

	return &ResilientClient{
		Client: &Client{
			HTTPClient: &http.Client{
				Timeout:   DefaultTimeout,
				Transport: transport,
			},
			Logger: logger,
		},
		retryConfig: DefaultRetryConfig(),
	}
}

// SetRetryConfig updates the retry configuration
func (rc *ResilientClient) SetRetryConfig(config *RetryConfig) {
	rc.retryConfig = config
}

// SendWithRetry sends a request with automatic retry
func (rc *ResilientClient) SendWithRetry(apiReq *Request) (*http.Response, error) {
	// Build HTTP request
	req, err := rc.buildHTTPRequest(apiReq)
	if err != nil {
		return nil, err
	}

	// Create retryable client
	retryable := &RetryableClient{
		HTTPClient:  rc.HTTPClient,
		RetryConfig: rc.retryConfig,
		Logger:      rc.Logger,
	}

	return retryable.Do(req)
}

// buildHTTPRequest converts API Request to HTTP Request
func (rc *ResilientClient) buildHTTPRequest(apiReq *Request) (*http.Request, error) {
	var reqBody io.Reader

	// Convert request body
	if apiReq.Body != nil {
		if reader, ok := apiReq.Body.(io.Reader); ok {
			data, err := io.ReadAll(reader)
			if err != nil {
				return nil, fmt.Errorf("error reading body: %w", err)
			}
			reqBody = bytes.NewBuffer(data)
			if closer, ok := reader.(io.Closer); ok {
				closer.Close()
			}
		} else {
			jsonData, err := json.Marshal(apiReq.Body)
			if err != nil {
				return nil, fmt.Errorf("error marshaling JSON: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonData)
		}
	}

	// Create request
	req, err := http.NewRequest(apiReq.Method, apiReq.URL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	for key, value := range apiReq.Headers {
		req.Header.Set(key, value)
	}

	return req, nil
}
