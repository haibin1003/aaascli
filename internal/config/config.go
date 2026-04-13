package config

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// 默认常量定义
const (
	// DefaultTaskItemTypeID 默认任务项类型ID
	DefaultTaskItemTypeID = "Co5QtC7wdQ"
	// DefaultPriorityID 默认优先级ID
	DefaultPriorityID = "8f7912a5-9176-4a79-a269-2269ac42b5a2"
	// DefaultAppID 默认应用ID
	DefaultAppID = "CMDEVOPS"
	// DefaultDomain 默认域
	DefaultDomain = "TENANT"
	// DefaultModule 默认模块
	DefaultModule = "SPACE"
)

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(homeDir, ".lc", "config.json")
}

// EnsureConfigDir ensures the configuration directory exists
func EnsureConfigDir() error {
	configPath := GetDefaultConfigPath()
	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0755)
}

// Config holds the complete configuration
type Config struct {
	Cookie          string         `json:"cookie"`
	Readonly        bool           `json:"readonly"`                  // 只读模式，默认 true
	TempOffExpireAt *time.Time     `json:"tempOffExpireAt,omitempty"` // 临时关闭过期时间
	API             APIConfig      `json:"api"`
	Auth            AuthConfig     `json:"auth"`
	User            UserConfig     `json:"user"`
	Defaults        DefaultsConfig `json:"defaults"`
	OTP             OTPConfig      `json:"otp,omitempty"`             // OTP 二次验证配置
}

// OTPConfig holds OTP (One-Time Password) configuration
// 用于保护危险操作的二次验证
type OTPConfig struct {
	Enabled       bool       `json:"enabled,omitempty"`        // 是否启用 OTP
	Secret        string     `json:"secret,omitempty"`         // TOTP 密钥 (base32 编码)
	VerifiedAt    *time.Time `json:"verifiedAt,omitempty"`     // 最后验证时间
	SessionExpiry int        `json:"sessionExpiryMinutes,omitempty"` // 验证会话有效期(分钟),默认5
	// ProtectedCommands 需要 OTP 验证的命令列表
	// 格式如: ["pr merge", "readonly off", "req delete"]
	// 如果为空，使用默认的受保护命令列表
	ProtectedCommands []string `json:"protectedCommands,omitempty"`
}

// APIConfig holds API configuration
type APIConfig struct {
	BaseRepoURL         string            `json:"base_url"`
	BaseReqURL          string            `json:"req_base_url"`
	BasePlatformURL     string            `json:"base_platform_url"`
	BasePlatformURLMoss string            `json:"base_platform_url_moss"`
	BaseTestCenterURL   string            `json:"test_center_base_url"`
	BaseProjectURL      string            `json:"base_project_url"`
	BaseDocURL          string            `json:"doc_base_url"`
	BaseCIURL           string            `json:"ci_base_url"`
	BaseArtifactURL     string            `json:"artifact_base_url"`
	Headers             map[string]string `json:"headers"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	AppID    string `json:"app_id"`
	Domain   string `json:"domain"`
	Module   string `json:"module"`
	TenantID string `json:"tenant_id"`
}

// UserConfig holds default user configuration
type UserConfig struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

// Label returns the formatted label like "Nickname(Username)"
func (u *UserConfig) Label() string {
	return u.Nickname + "(" + u.Username + ")"
}

// DefaultsConfig holds default values for items
type DefaultsConfig struct {
	TaskItemTypeID string `json:"task_item_type_id"`
	PriorityID     string `json:"priority_id"`
}

// UserAPIResponse represents the API response for user query (old API)
type UserAPIResponse struct {
	Success bool       `json:"success"`
	Code    string     `json:"code"`
	Message string     `json:"message"`
	Data    []UserData `json:"data"`
}

// UserData represents a single user entry from the old API
type UserData struct {
	EuserID             string `json:"euserId"`
	TenantID            string `json:"tenantId"`
	UserID              string `json:"userId"`
	Name                string `json:"name"`
	Email               string `json:"email"`
	DeputyAccountNumber string `json:"deputyAccountNumber"`
	StaffCode           string `json:"staffCode"`
	OrgName             string `json:"orgName"`
}

// SelfUserInfoResponse represents the API response from /v1/self/user-info
type SelfUserInfoResponse struct {
	Head struct {
		RequestID  string `json:"requestId"`
		RespStatus string `json:"respStatus"`
		RespCode   string `json:"respCode"`
		RespDesc   string `json:"respDesc"`
	} `json:"head"`
	Data struct {
		UserID               string `json:"userId"`
		UserName             string `json:"userName"`
		EnterpriseID         string `json:"enterpriseId"`
		EnterpriseName       string `json:"enterpriseName"`
		EuserID              string `json:"euserId"`
		ExternalEnterpriseID string `json:"externalEnterpriseId"`
		DeputyAccountNumber  string `json:"deputyAccountNumber"`
		StaffCode            string `json:"staffCode"`
		EnterpriseCode       string `json:"enterpriseCode"`
		Insider              int    `json:"insider"`
	} `json:"data"`
}

// CachedUserInfo holds all user information fetched from API (in-memory cache)
type CachedUserInfo struct {
	UserID               string
	UserName             string
	EnterpriseID         string // TenantID
	EnterpriseName       string
	EuserID              string
	ExternalEnterpriseID string
	DeputyAccountNumber  string // Username/Email
	StaffCode            string
	EnterpriseCode       string
	Insider              int
}

// Label returns the formatted label like "UserName(DeputyAccountNumber)"
func (c *CachedUserInfo) Label() string {
	return c.UserName + "(" + c.DeputyAccountNumber + ")"
}

// ToUserConfig converts CachedUserInfo to UserConfig for compatibility
func (c *CachedUserInfo) ToUserConfig() *UserConfig {
	return &UserConfig{
		Username: c.DeputyAccountNumber,
		Nickname: c.UserName,
		Email:    c.DeputyAccountNumber,
	}
}

// cachedUserInfo holds the user information fetched from API (in-memory cache)
var (
	cachedUserInfo   *CachedUserInfo
	cachedUserInfoMu sync.RWMutex
)

// NewConfig creates a new Config with default values
func NewConfig() *Config {
	return &Config{
		Readonly: true, // 默认开启只读模式
		API: APIConfig{
			BaseRepoURL:         "https://rdcloud.4c.hq.cmcc/moss/web/cmdevops-code/server",
			BaseReqURL:          "https://rdcloud.4c.hq.cmcc/moss/web/cmdevops-req/api/team/parse",
			BasePlatformURL:     "https://rdcloud.4c.hq.cmcc",
			BasePlatformURLMoss: "https://rdcloud.4c.hq.cmcc/moss/web/",
			//BasePlatformURLMoss: "https://rdcloud.4c-uat3.hq.cmcc:20019/moss/web",
			BaseTestCenterURL: "https://rdcloud.4c.hq.cmcc/moss/web/cmdevops-ct/testcenter",
			BaseProjectURL:    "https://rdcloud.4c.hq.cmcc/moss/web/cmdevops-project/server/api/v1",
			BaseDocURL:        "https://rdcloud.4c.hq.cmcc",
			BaseCIURL:         "https://rdcloud.4c.hq.cmcc/moss/web",
			BaseArtifactURL:   "https://rdcloud.4c.hq.cmcc/moss/web",
			Headers: map[string]string{
				"Accept":       "application/json, text/plain, */*",
				"Content-Type": "application/json",
				"Cookie":       "xyz",
				"Origin":       "https://rdcloud.4c.hq.cmcc",
				"Referer":      "https://rdcloud.4c.hq.cmcc",
				"User-Agent":   "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.6478.251 Safari/537.36 UOS Professional",
			},
		},
		Auth: AuthConfig{
			AppID:  DefaultAppID,
			Domain: DefaultDomain,
			Module: DefaultModule,
		},
		Defaults: DefaultsConfig{
			TaskItemTypeID: DefaultTaskItemTypeID,
			PriorityID:     DefaultPriorityID,
		},
		OTP: OTPConfig{
			SessionExpiry: 5, // 默认5分钟会话有效期
		},
	}
}

// LoadConfig loads configuration from a JSON file
// Deprecated: Use LoadConfigWithDefaults instead
func LoadConfig(filename string) (*Config, error) {
	return LoadConfigWithDefaults(filename)
}

// LoadConfigWithDefaults loads configuration from a JSON file and applies defaults for empty values
// The config file only needs to contain the cookie, all other values use defaults
func LoadConfigWithDefaults(filename string) (*Config, error) {
	// Start with default config
	cfg := NewConfig()

	// Try to load cookie from config file
	if file, err := os.Open(filename); err == nil {
		defer file.Close()

		// Try to parse as simple config (only cookie) first
		var simpleConfig map[string]string
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&simpleConfig); err == nil {
			if cookie, ok := simpleConfig["cookie"]; ok {
				cfg.Cookie = cookie
			}
		} else {
			// If not a map, try full Config structure (backward compatibility)
			file.Seek(0, 0)
			var loadedCfg Config
			if err := json.NewDecoder(file).Decode(&loadedCfg); err == nil {
				cfg.Cookie = loadedCfg.Cookie
				cfg.Readonly = loadedCfg.Readonly
				cfg.TempOffExpireAt = loadedCfg.TempOffExpireAt
				// Load OTP config if present
				if loadedCfg.OTP.Enabled {
					cfg.OTP = loadedCfg.OTP
				}
			}
		}
	}

	// Set cookie in headers if present
	if cfg.Cookie != "" {
		cfg.API.Headers["Cookie"] = cfg.Cookie
	}

	// Try to fetch current user from API if not configured
	if cfg.User.Username == "" {
		if err := cfg.FetchCurrentUser(); err == nil {
			cachedUserInfoMu.RLock()
			info := cachedUserInfo
			cachedUserInfoMu.RUnlock()
			if info != nil {
				cfg.User = *info.ToUserConfig()
				// Also update Auth.TenantID from API response
				if info.EnterpriseID != "" {
					cfg.Auth.TenantID = info.EnterpriseID
				}
			}
		}
	}

	return cfg, nil
}

// GetHeaders returns the complete headers map including auth headers
func (c *Config) GetHeaders(workspaceKey string) map[string]string {
	return c.GetHeadersWithWorkspace(workspaceKey)
}

// GetPlatformHeaders returns headers for platform API calls
func (c *Config) GetPlatformHeaders() map[string]string {
	// For platform calls, we don't need workspace key
	return c.GetHeadersWithWorkspace("")
}

// GetHeadersWithWorkspace returns headers with the specified workspace key
func (c *Config) GetHeadersWithWorkspace(workspaceKey string) map[string]string {
	headers := make(map[string]string)

	// Copy API headers
	for k, v := range c.API.Headers {
		headers[k] = v
	}

	// Add auth headers
	headers["X-Auth-AppId"] = c.Auth.AppID
	headers["X-Auth-Domain"] = c.Auth.Domain
	headers["X-Auth-Module"] = c.Auth.Module
	headers["X-Auth-ModuleId"] = workspaceKey
	headers["X-Auth-Moudle"] = workspaceKey
	headers["X-Auth-Moudleid"] = workspaceKey
	headers["X-Auth-TenantId"] = c.Auth.TenantID
	headers["X-Parse-Application-Id"] = c.Auth.TenantID
	headers["X-Auth-Spaceid"] = workspaceKey

	return headers
}

// FetchCurrentUser fetches the current user info from the API and caches it
func (c *Config) FetchCurrentUser() error {
	cachedUserInfoMu.RLock()
	if cachedUserInfo != nil {
		cachedUserInfoMu.RUnlock()
		return nil
	}
	cachedUserInfoMu.RUnlock()

	url := "https://rdcloud.4c.hq.cmcc/moss/web/manage/v1/self/user-info"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers - use basic headers without TenantId for bootstrap
	for k, v := range c.API.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("X-Auth-AppId", c.Auth.AppID)
	req.Header.Set("X-Auth-Domain", c.Auth.Domain)
	req.Header.Set("X-Auth-Module", c.Auth.Module)

	// Use insecure client to skip SSL certificate verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var userResp SelfUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if userResp.Head.RespCode != "00" {
		return fmt.Errorf("failed to get user info: %s", userResp.Head.RespDesc)
	}

	// Cache all user info
	cachedUserInfoMu.Lock()
	cachedUserInfo = &CachedUserInfo{
		UserID:               userResp.Data.UserID,
		UserName:             userResp.Data.UserName,
		EnterpriseID:         userResp.Data.EnterpriseID,
		EnterpriseName:       userResp.Data.EnterpriseName,
		EuserID:              userResp.Data.EuserID,
		ExternalEnterpriseID: userResp.Data.ExternalEnterpriseID,
		DeputyAccountNumber:  userResp.Data.DeputyAccountNumber,
		StaffCode:            userResp.Data.StaffCode,
		EnterpriseCode:       userResp.Data.EnterpriseCode,
		Insider:              userResp.Data.Insider,
	}
	cachedUserInfoMu.Unlock()

	// Also update config's Auth.TenantID from API response
	if userResp.Data.EnterpriseID != "" {
		c.Auth.TenantID = userResp.Data.EnterpriseID
	}

	return nil
}

// GetCachedUser returns the cached full user info, fetching from API if not cached
func (c *Config) GetCachedUser() (*CachedUserInfo, error) {
	cachedUserInfoMu.RLock()
	info := cachedUserInfo
	cachedUserInfoMu.RUnlock()

	if info == nil {
		if err := c.FetchCurrentUser(); err != nil {
			return nil, err
		}
		cachedUserInfoMu.RLock()
		info = cachedUserInfo
		cachedUserInfoMu.RUnlock()
	}
	return info, nil
}

// GetUser returns the user config (from config file or API cached)
func (c *Config) GetUser() *UserConfig {
	// If config has user info, use it
	if c.User.Username != "" {
		return &c.User
	}
	// Otherwise return cached user converted to UserConfig
	cachedUserInfoMu.RLock()
	info := cachedUserInfo
	cachedUserInfoMu.RUnlock()
	if info != nil {
		return info.ToUserConfig()
	}
	// Return default as fallback
	return &c.User
}

// GetTenantID returns the tenant ID (from config or API)
func (c *Config) GetTenantID() string {
	// First check API cached info
	cachedUserInfoMu.RLock()
	info := cachedUserInfo
	cachedUserInfoMu.RUnlock()

	if info != nil && info.EnterpriseID != "" {
		return info.EnterpriseID
	}
	// Then check config
	if c.Auth.TenantID != "" {
		return c.Auth.TenantID
	}
	// Try to fetch from API
	if err := c.FetchCurrentUser(); err == nil {
		cachedUserInfoMu.RLock()
		info = cachedUserInfo
		cachedUserInfoMu.RUnlock()
		if info != nil {
			return info.EnterpriseID
		}
	}
	// Return empty if all fail
	return ""
}

// SaveConfig saves configuration to the default config file
func SaveConfig(cfg *Config) error {
	configPath := GetDefaultConfigPath()

	// Ensure config directory exists
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to ensure config directory: %w", err)
	}

	// Read existing config to preserve other fields
	var existing map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, &existing)
	}
	if existing == nil {
		existing = make(map[string]interface{})
	}

	// Update fields
	existing["cookie"] = cfg.Cookie
	existing["readonly"] = cfg.Readonly
	if cfg.TempOffExpireAt != nil {
		existing["tempOffExpireAt"] = cfg.TempOffExpireAt.Format(time.RFC3339)
	} else {
		delete(existing, "tempOffExpireAt")
	}

	// Update OTP config if enabled
	if cfg.OTP.Enabled {
		existing["otp"] = cfg.OTP
	} else {
		delete(existing, "otp")
	}

	// Save back to file
	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
