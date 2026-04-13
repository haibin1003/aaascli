package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type APIRequest struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    interface{}
}

// DefaultAPITimeout is the default timeout for API requests
const DefaultAPITimeout = 30 * time.Second

func SendAPIRequest(apiReq *APIRequest) (*http.Response, error) {
	// 将请求体转换为JSON
	jsonData, err := json.Marshal(apiReq.Body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest(apiReq.Method, apiReq.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// 设置请求头
	for key, value := range apiReq.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	client := &http.Client{Timeout: DefaultAPITimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	return resp, nil
}
