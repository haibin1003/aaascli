package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Request represents an HTTP request
type Request struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    interface{}
}

// Client is an HTTP client for sending requests
type Client struct {
	HTTPClient *http.Client
	Logger     *zap.Logger
}

// DefaultTimeout is the default timeout for HTTP requests
const DefaultTimeout = 30 * time.Second

// defaultTransport is the shared HTTP transport with connection pooling
var defaultTransport = NewEnhancedTransport(false)

// insecureTransport is the shared HTTP transport for insecure connections
var insecureTransport = NewEnhancedTransport(true)

// NewClient creates a new Client with default HTTP client
func NewClient() *Client {
	logger, _ := zap.NewProduction()

	return &Client{
		HTTPClient: &http.Client{
			Timeout:   DefaultTimeout,
			Transport: defaultTransport,
		},
		Logger: logger,
	}
}

// NewClientWithLogger creates a new Client with custom logger
func NewClientWithLogger(logger *zap.Logger) *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout:   DefaultTimeout,
			Transport: defaultTransport,
		},
		Logger: logger,
	}
}

// NewInsecureClient creates a new Client that skips TLS verification
func NewInsecureClient() *Client {
	logger, _ := zap.NewProduction()

	return &Client{
		HTTPClient: &http.Client{
			Transport: insecureTransport,
			Timeout:   DefaultTimeout,
		},
		Logger: logger,
	}
}

// NewInsecureClientWithLogger creates a new Client with custom logger that skips TLS verification
func NewInsecureClientWithLogger(logger *zap.Logger) *Client {
	return &Client{
		HTTPClient: &http.Client{
			Transport: insecureTransport,
			Timeout:   DefaultTimeout,
		},
		Logger: logger,
	}
}

// Send sends an HTTP request and returns the response
func (c *Client) Send(apiReq *Request) (*http.Response, error) {
	var reqBody io.Reader

	// Convert request body to appropriate format
	if apiReq.Body != nil {
		// If body is already an io.Reader, use it directly
		if reader, ok := apiReq.Body.(io.Reader); ok {
			// Read for logging (then we need to recreate it)
			if data, err := io.ReadAll(reader); err == nil {
				c.Logger.Debug("Sending request",
					zap.String("url", apiReq.URL),
					zap.String("method", apiReq.Method),
					zap.String("body", string(data)))
				// Use the read data as body
				reqBody = bytes.NewBuffer(data)
				// Close the original reader if it implements io.Closer
				if closer, ok := reader.(io.Closer); ok {
					closer.Close()
				}
			} else {
				reqBody = reader
			}
		} else {
			// Otherwise, treat as JSON
			jsonData, err := json.Marshal(apiReq.Body)
			if err != nil {
				c.Logger.Error("Error marshaling JSON", zap.Error(err))
				return nil, fmt.Errorf("error marshaling JSON: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonData)

			// Log request details
			c.Logger.Debug("Sending request",
				zap.String("url", apiReq.URL),
				zap.String("method", apiReq.Method),
				zap.String("body", string(jsonData)))
		}
	} else {
		c.Logger.Debug("Sending request",
			zap.String("url", apiReq.URL),
			zap.String("method", apiReq.Method))
	}

	// Create HTTP request
	req, err := http.NewRequest(apiReq.Method, apiReq.URL, reqBody)
	if err != nil {
		c.Logger.Error("Error creating request", zap.Error(err))
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set request headers
	for key, value := range apiReq.Headers {
		req.Header.Set(key, value)
		c.Logger.Debug("Request header", zap.String(key, value))
	}

	// Log request headers
	c.Logger.Debug("Sending request headers", zap.Any("headers", req.Header))

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Logger.Error("Error sending request", zap.Error(err))
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	// Log response status
	c.Logger.Debug("Received response", zap.String("status", resp.Status))

	return resp, nil
}
