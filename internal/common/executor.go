package common

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/haibin1003/aaascli/internal/api"
	"github.com/haibin1003/aaascli/internal/config"
)

// version 保存 CLI 版本
var version = "dev"

// SetVersion 设置版本
func SetVersion(v string) {
	version = v
}

// MetaInfo 元信息
type MetaInfo struct {
	Timestamp     string `json:"timestamp"`
	ExecutionTime int64  `json:"executionTime,omitempty"`
	Version       string `json:"version"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// CommandResult 统一命令结果
type CommandResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    MetaInfo    `json:"meta"`
}

// CommandFunc 命令执行函数类型
type CommandFunc func(ctx *CommandContext) (interface{}, error)

// ExecuteOptions 执行选项
type ExecuteOptions struct {
	DebugMode   bool
	Insecure    bool
	DryRun      bool
	Cookie      string
	PrettyPrint bool
}

// Execute 执行命令函数
func Execute(fn CommandFunc, opts ExecuteOptions) {
	start := time.Now()

	ctx, err := NewCommandContext(opts)
	if err != nil {
		PrintError(err, "")
		os.Exit(1)
	}

	result, err := fn(ctx)
	if err != nil {
		suggestion := ""
		if ctx.Config.Cookie == "" {
			suggestion = "请先执行 sdp login <token> 登录"
		}
		PrintError(err, suggestion)
		os.Exit(1)
	}

	cmdResult := CommandResult{
		Success: true,
		Data:    result,
		Meta: MetaInfo{
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			ExecutionTime: time.Since(start).Milliseconds(),
			Version:       version,
		},
	}
	PrintResult(cmdResult, opts.PrettyPrint)
}

// PrintError 打印错误（JSON 格式）
func PrintError(err error, suggestion string) {
	errDetail := ErrorDetail{
		Code:    "ERROR",
		Message: err.Error(),
	}
	if suggestion != "" {
		errDetail.Suggestion = suggestion
	}

	cmdResult := CommandResult{
		Success: false,
		Error:   errDetail,
		Meta: MetaInfo{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   version,
		},
	}
	PrintResult(cmdResult, false)
}

// PrintResult 打印结果
func PrintResult(v interface{}, pretty bool) {
	var output []byte
	var err error
	if pretty {
		output, err = json.MarshalIndent(v, "", "  ")
	} else {
		output, err = json.Marshal(v)
	}
	if err != nil {
		fmt.Printf(`{"success":false,"error":{"code":"MARSHAL_ERROR","message":"%s"}}`, err.Error())
		return
	}
	fmt.Println(string(output))
}

// CommandContext 命令上下文
type CommandContext struct {
	Config   *config.Config
	Client   *api.Client
	Debug    bool
	DryRun   bool
	Insecure bool
}

// NewCommandContext 创建命令上下文
func NewCommandContext(opts ExecuteOptions) (*CommandContext, error) {
	// 加载配置
	cfg := config.NewConfig()
	configPath := config.GetDefaultConfigPath()

	if loadedCfg, err := config.LoadConfigWithDefaults(configPath); err == nil {
		cfg = loadedCfg
	}

	// 如果通过 flag 指定了 cookie，覆盖配置
	if opts.Cookie != "" {
		cfg.Cookie = opts.Cookie
	}

	// 创建 API 客户端
	client := api.NewClient(cfg.Cookie, opts.Insecure)

	return &CommandContext{
		Config:   cfg,
		Client:   client,
		Debug:    opts.DebugMode,
		DryRun:   opts.DryRun,
		Insecure: opts.Insecure,
	}, nil
}

// CheckLoggedIn 检查是否已登录
func (c *CommandContext) CheckLoggedIn() error {
	if c.Config.Cookie == "" {
		return fmt.Errorf("未登录，请先执行: sdp login <token>")
	}
	return nil
}
