package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	// DefaultTimeout 默认请求超时
	DefaultTimeout = 60 * time.Second
	// BaseURL 平台基础 URL
	BaseURL = "https://service.sd.10086.cn"
)

// Client HTTP 客户端
type Client struct {
	HTTPClient       *http.Client
	Headers          map[string]string
	Cookie           string
	VerificationCode string
	ServiceID        string
	Crypto           *CryptoContext
}

// NewClient 创建新客户端
func NewClient(cookie string, insecure bool) *Client {
	return NewClientWithExtra(cookie, "", "", insecure)
}

// NewClientWithExtra 创建带额外 cookie 的新客户端
func NewClientWithExtra(cookie, verificationCode, serviceID string, insecure bool) *Client {
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
			"Content-Type":    "application/x-www-form-urlencoded",
			"Origin":          "https://service.sd.10086.cn",
			"Referer":         "https://service.sd.10086.cn/aaas/",
			"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		},
		Cookie:           cookie,
		VerificationCode: verificationCode,
		ServiceID:        serviceID,
		Crypto:           &CryptoContext{},
	}
}

// ensureCrypto 确保加密上下文就绪
func (c *Client) ensureCrypto() error {
	if c.Crypto.LocalKeyPair == nil {
		pair, err := GenerateRSAKeyPair()
		if err != nil {
			return fmt.Errorf("generate local RSA key pair failed: %w", err)
		}
		c.Crypto.LocalKeyPair = pair
	}

	if c.Crypto.PlatformPubKey == "" || time.Now().After(c.Crypto.PubKeyExpireAt) {
		pubKey, err := RefreshPlatformPublicKey(c.GetFullCookie())
		if err != nil {
			return fmt.Errorf("refresh platform public key failed: %w", err)
		}
		c.Crypto.PlatformPubKey = pubKey
		c.Crypto.PubKeyExpireAt = time.Now().Add(5 * time.Minute)
	}

	return nil
}

// SetCookie 设置认证 Cookie
func (c *Client) SetCookie(cookie string) {
	c.Cookie = cookie
}

// GetFullCookie 构建完整 Cookie 字符串
func (c *Client) GetFullCookie() string {
	var parts []string
	if c.Cookie != "" {
		if strings.HasPrefix(c.Cookie, "#openPortal#token#=") {
			parts = append(parts, c.Cookie)
		} else {
			parts = append(parts, fmt.Sprintf("#openPortal#token#=%s", c.Cookie))
		}
	}
	if c.VerificationCode != "" {
		parts = append(parts, fmt.Sprintf("openPortalVerificationCode=%s", c.VerificationCode))
	}
	if c.ServiceID != "" {
		parts = append(parts, fmt.Sprintf("openPortalServiceID=%s", c.ServiceID))
	}
	return strings.Join(parts, "; ")
}

// Request 发送 HTTP 请求
func (c *Client) Request(method, path string, body interface{}) (*http.Response, error) {
	url := BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		default:
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshal request body failed: %w", err)
			}
			bodyReader = bytes.NewReader(jsonData)
		}
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

	// 设置 token header（平台部分接口需要）
	if c.Cookie != "" {
		tokenValue := c.Cookie
		if strings.HasPrefix(tokenValue, "#openPortal#token#=") {
			tokenValue = strings.TrimPrefix(tokenValue, "#openPortal#token#=")
		}
		req.Header.Set("token", tokenValue)
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

// PostMultipart 发送 multipart/form-data POST 请求
func (c *Client) PostMultipart(path string, contentType string, body []byte) (*http.Response, error) {
	url := BaseURL + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}
	// 覆盖 Content-Type
	req.Header.Set("Content-Type", contentType)

	// 设置 Cookie
	if c.Cookie != "" {
		req.Header.Set("Cookie", c.GetFullCookie())
	}

	// 设置 token header
	if c.Cookie != "" {
		tokenValue := c.Cookie
		if strings.HasPrefix(tokenValue, "#openPortal#token#=") {
			tokenValue = strings.TrimPrefix(tokenValue, "#openPortal#token#=")
		}
		req.Header.Set("token", tokenValue)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}

	return resp, nil
}

// PostEncrypted 发送加密的 POST 请求，并自动解密响应
func (c *Client) PostEncrypted(path string, body interface{}) (*http.Response, error) {
	if err := c.ensureCrypto(); err != nil {
		return nil, err
	}

	// 将 body 转为 JSON 字符串作为明文
	var plainBody string
	switch v := body.(type) {
	case string:
		plainBody = v
	case []byte:
		plainBody = string(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal request body failed: %w", err)
		}
		plainBody = string(b)
	}

	encryptedBody, err := c.Crypto.EncryptRequest(plainBody, c.Headers["Content-Type"])
	if err != nil {
		return nil, fmt.Errorf("encrypt request failed: %w", err)
	}

	resp, err := c.Request(http.MethodPost, path, encryptedBody)
	if err != nil {
		return nil, err
	}

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// 尝试解密响应
	decryptedBody, err := c.Crypto.DecryptResponse(string(respBody))
	if err != nil {
		// 解密失败，可能是未加密的响应，回退原文
		decryptedBody = string(respBody)
	}

	// 重新构造 Response
	resp.Body = io.NopCloser(strings.NewReader(decryptedBody))
	resp.ContentLength = int64(len(decryptedBody))
	resp.Header.Del("Content-Length")

	return resp, nil
}

// ParseJSON 解析 JSON 响应，自动检测并转换 GBK 编码
func ParseJSON(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if err := tryUnmarshalJSON(body, v); err != nil {
		return fmt.Errorf("parse JSON failed: %w, body: %s", err, string(body))
	}

	return nil
}

// tryUnmarshalJSON 尝试解析 JSON，如果检测到 GBK 编码则自动转换
func tryUnmarshalJSON(body []byte, v interface{}) error {
	if err := json.Unmarshal(body, v); err != nil {
		return err
	}

	// 检测是否包含 Unicode replacement character（说明原始 body 是 GBK 编码）
	tmp, _ := json.Marshal(v)
	if strings.Contains(string(tmp), "\uFFFD") || strings.Contains(string(tmp), "�") {
		decoder := simplifiedchinese.GBK.NewDecoder()
		utf8Bytes, _, err := transform.Bytes(decoder, body)
		if err == nil {
			if err := json.Unmarshal(utf8Bytes, v); err == nil {
				return nil
			}
		}
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
