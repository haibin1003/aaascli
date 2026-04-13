package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultTimeout 默认请求超时
	DefaultTimeout = 30 * time.Second
	// BaseURL 平台基础 URL
	BaseURL = "https://service.sd.10086.cn/aaas"
)

// Client HTTP 客户端
type Client struct {
	HTTPClient *http.Client
	Headers    map[string]string
	Cookie     string
}

// NewClient 创建新客户端
func NewClient(cookie string, insecure bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	return &Client{
		HTTPClient: &http.Client{
			Timeout:   DefaultTimeout,
			Transport: transport,
		},
		Headers: map[string]string{
			"Accept":          "application/json, text/plain, */*",
			"Accept-Language": "zh-CN,zh;q=0.9",
			"Content-Type":    "application/json",
			"Origin":          "https://service.sd.10086.cn",
			"Referer":         "https://service.sd.10086.cn/aaas/",
			"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		},
		Cookie: cookie,
	}
}

// SetCookie 设置认证 Cookie
func (c *Client) SetCookie(cookie string) {
	c.Cookie = cookie
}

// GetFullCookie 构建完整 Cookie 字符串
func (c *Client) GetFullCookie() string {
	if strings.HasPrefix(c.Cookie, "#openPortal#token#=") {
		return c.Cookie
	}
	return fmt.Sprintf("#openPortal#token#=%s", c.Cookie)
}

// Request 发送 HTTP 请求
func (c *Client) Request(method, path string, body interface{}) (*http.Response, error) {
	url := BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body failed: %w", err)
		}
		bodyReader = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	// 设置 Cookie
	if c.Cookie != "" {
		req.Header.Set("Cookie", c.GetFullCookie())
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}

	return resp, nil
}

// Get 发送 GET 请求
func (c *Client) Get(path string) (*http.Response, error) {
	return c.Request(http.MethodGet, path, nil)
}

// Post 发送 POST 请求
func (c *Client) Post(path string, body interface{}) (*http.Response, error) {
	return c.Request(http.MethodPost, path, body)
}

// ParseJSON 解析 JSON 响应
func ParseJSON(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("parse JSON failed: %w, body: %s", err, string(body))
	}

	return nil
}

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// IsSuccess 判断响应是否成功
func (r *Response) IsSuccess() bool {
	return r.Code == 200 || r.Success
}

// Error 返回错误信息
func (r *Response) Error() error {
	if r.IsSuccess() {
		return nil
	}
	return fmt.Errorf("API error [%d]: %s", r.Code, r.Message)
}
