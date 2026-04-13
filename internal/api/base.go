package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// BaseService provides common HTTP operations for all API services.
// It should be embedded in specific service implementations.
type BaseService struct {
	BaseURL    string
	Headers    map[string]string
	HTTPClient HTTPClient
}

// HTTPClient is an interface for HTTP client operations
type HTTPClient interface {
	Send(*Request) (*http.Response, error)
}

// NewBaseService creates a new BaseService
func NewBaseService(baseURL string, headers map[string]string, client HTTPClient) BaseService {
	return BaseService{
		BaseURL:    baseURL,
		Headers:    headers,
		HTTPClient: client,
	}
}

// Get performs an HTTP GET request
func (b *BaseService) Get(path string) (*http.Response, error) {
	apiReq := &Request{
		URL:     b.BaseURL + path,
		Method:  "GET",
		Headers: b.Headers,
		Body:    nil,
	}
	return b.HTTPClient.Send(apiReq)
}

// GetWithQuery performs an HTTP GET request with query parameters
func (b *BaseService) GetWithQuery(path string) (*http.Response, error) {
	apiReq := &Request{
		URL:     path,
		Method:  "GET",
		Headers: b.Headers,
		Body:    nil,
	}
	return b.HTTPClient.Send(apiReq)
}

// Post performs an HTTP POST request with a JSON body
func (b *BaseService) Post(path string, body interface{}) (*http.Response, error) {
	apiReq := &Request{
		URL:     b.BaseURL + path,
		Method:  "POST",
		Headers: b.Headers,
		Body:    body,
	}
	return b.HTTPClient.Send(apiReq)
}

// Put performs an HTTP PUT request with a JSON body
func (b *BaseService) Put(path string, body interface{}) (*http.Response, error) {
	apiReq := &Request{
		URL:     b.BaseURL + path,
		Method:  "PUT",
		Headers: b.Headers,
		Body:    body,
	}
	return b.HTTPClient.Send(apiReq)
}

// Delete performs an HTTP DELETE request
func (b *BaseService) Delete(path string) (*http.Response, error) {
	apiReq := &Request{
		URL:     b.BaseURL + path,
		Method:  "DELETE",
		Headers: b.Headers,
		Body:    nil,
	}
	return b.HTTPClient.Send(apiReq)
}

// Do performs a custom HTTP request
func (b *BaseService) Do(method, path string, body interface{}) (*http.Response, error) {
	apiReq := &Request{
		URL:     b.BaseURL + path,
		Method:  method,
		Headers: b.Headers,
		Body:    body,
	}
	return b.HTTPClient.Send(apiReq)
}

// ParseJSON parses the response body into the target object
func (b *BaseService) ParseJSON(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w, body: %s", err, string(body))
	}
	return nil
}

// ParseJSONWithoutClose parses the response body without closing it
func (b *BaseService) ParseJSONWithoutClose(resp *http.Response, target interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w, body: %s", err, string(body))
	}
	return nil
}
