// Package common provides shared utilities for command execution.
package common

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

// version holds the CLI version, set by cmd package
var version = "dev"

// prettyPrint controls JSON output format
// When true, JSON is indented with 2 spaces; when false, JSON is compact
var prettyPrint = false

// SetVersion sets the version for meta info
func SetVersion(v string) {
	version = v
}

// GetVersion returns the current version
func GetVersion() string {
	return version
}

// SetPrettyPrint enables/disables pretty JSON output
func SetPrettyPrint(pretty bool) {
	prettyPrint = pretty
}

// MetaInfo represents metadata for command execution
type MetaInfo struct {
	RequestID     string `json:"requestId"`
	Timestamp     string `json:"timestamp"`
	ExecutionTime int64  `json:"executionTime,omitempty"`
	Version       string `json:"version"`
}

// ErrorDetail represents detailed error information
type ErrorDetail struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// CommandResult represents the unified result of a command execution
type CommandResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    MetaInfo    `json:"meta"`
}

// CommandFunc is the type of command execution functions
type CommandFunc func(ctx *CommandContext) (interface{}, error)

// CommandFuncWithOutput is the type of command execution functions that handle their own output
type CommandFuncWithOutput func(ctx *CommandContext) error

// ExecuteOptions holds execution options
type ExecuteOptions struct {
	DebugMode   bool
	Insecure    bool
	DryRun      bool
	Cookie      string
	Logger      **zap.Logger
	CommandName string // 命令名称，同时用于只读检查（CommandRegistry.IsWrite）和 OTP 检查（用户配置）
	PrettyPrint bool   // 是否输出格式化 JSON
}

// Execute runs a command function with standard error handling and unified output
func Execute(fn CommandFunc, opts ExecuteOptions) {
	start := time.Now()
	wrapper := func(data interface{}) interface{} {
		return CommandResult{
			Success: true,
			Data:    data,
			Meta: MetaInfo{
				Timestamp:     time.Now().UTC().Format(time.RFC3339),
				ExecutionTime: time.Since(start).Milliseconds(),
				Version:       GetVersion(),
			},
		}
	}
	executeWithWrapper(fn, opts, wrapper)
}

// ExecuteWithOutput runs a command function that handles its own output (no JSON formatting)
func ExecuteWithOutput(fn CommandFuncWithOutput, opts ExecuteOptions) {
	ctx, err := NewCommandContext(opts.DebugMode, opts.Insecure, opts.DryRun, opts.Cookie)
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
	defer ctx.Close()

	if opts.DebugMode && opts.Logger != nil {
		*opts.Logger = ctx.Logger
	}

	if err := fn(ctx); err != nil {
		PrintError(err)
		os.Exit(1)
	}
}

// executeWithWrapper is the common implementation for Execute and ExecuteWithResult
func executeWithWrapper(fn CommandFunc, opts ExecuteOptions, wrapper func(interface{}) interface{}) {
	// 设置 pretty print 模式
	SetPrettyPrint(opts.PrettyPrint)

	ctx, err := NewCommandContext(opts.DebugMode, opts.Insecure, opts.DryRun, opts.Cookie)
	if err != nil {
		if wrapper != nil {
			PrintErrorResult(err)
		} else {
			PrintError(err)
		}
		os.Exit(1)
	}
	defer ctx.Close()

	if opts.DebugMode && opts.Logger != nil {
		*opts.Logger = ctx.Logger
	}

	// 两道安全检查使用同一份 ctx.Config，避免重复读盘及时序不一致
	if opts.CommandName != "" {
		// ① 只读检查：IsWrite=true 的写操作在只读模式下被拦截
		if err := CheckReadonlyWithConfig(ctx.Config, opts.CommandName); err != nil {
			PrintError(err)
			os.Exit(1)
		}
		// ② OTP 检查：保护列表中的命令需要 OTP 二次验证
		if err := RequireOTPOrPrompt(ctx.Config, opts.CommandName); err != nil {
			PrintError(err)
			os.Exit(1)
		}
	}

	result, err := fn(ctx)
	if err != nil {
		if wrapper != nil {
			PrintErrorResult(err)
		} else {
			PrintError(err)
		}
		os.Exit(1)
	}

	if wrapper != nil {
		result = wrapper(result)
	}
	PrintJSON(result)
}

// PrintError prints an error result as JSON to stdout.
// This maintains unified output format - all errors are returned as JSON.
func PrintError(err error) {
	errorDetail := ErrorDetail{
		Code:    "INTERNAL_ERROR",
		Message: err.Error(),
	}

	// If it's an AutoDetectError, preserve the structured information
	if autoErr, ok := err.(*AutoDetectError); ok {
		errorDetail.Code = "AUTO_DETECT_FAILED"
		errorDetail.Message = autoErr.Message
		errorDetail.Details = autoErr.Details
		errorDetail.Suggestion = autoErr.Suggestion
	}

	// If it's a ReadonlyError, preserve the structured information
	if roErr, ok := err.(*ReadonlyError); ok {
		errorDetail.Code = "READONLY_MODE"
		errorDetail.Message = "当前处于只读模式，禁止执行写入操作"
		errorDetail.Details = fmt.Sprintf("命令 '%s' 是写入操作，在只读模式下被禁止", roErr.Command)
		errorDetail.Suggestion = "如需执行写入操作，请先关闭只读模式：\n" +
			"  lc readonly off --duration 5m\n\n" +
			"注意：\n" +
			"  - 关闭只读模式后，所有写入操作将直接生效\n" +
			"  - AI 在获得人类明确授权后可以执行此命令"
	}

	// If it's an OTPGuardError, preserve the structured information
	if otpErr, ok := err.(*OTPGuardError); ok {
		errorDetail.Code = "OTP_REQUIRED"
		errorDetail.Message = otpErr.Message
		errorDetail.Details = fmt.Sprintf("操作 '%s' 需要 OTP 二次验证", otpErr.Description)
		errorDetail.Suggestion = otpErr.Suggestion
	}

	result := CommandResult{
		Success: false,
		Error:   errorDetail,
		Meta: MetaInfo{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   GetVersion(),
		},
	}
	PrintJSON(result)
}

// PrintErrorResult prints an error result as JSON
func PrintErrorResult(err error) {
	wrap := CommandResult{
		Success: false,
		Error: ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		},
		Meta: MetaInfo{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   GetVersion(),
		},
	}
	PrintJSON(wrap)
}

// PrintJSON prints an object as JSON
// When prettyPrint is true, outputs indented JSON with 2 spaces; otherwise outputs compact JSON
func PrintJSON(v interface{}) {
	var output []byte
	var err error
	if prettyPrint {
		output, err = json.MarshalIndent(v, "", "  ")
	} else {
		output, err = json.Marshal(v)
	}
	if err != nil {
		// Fallback: print minimal JSON error when marshaling fails
		// This should not happen in practice as we control the data types
		fmt.Printf(`{
  "success": false,
  "error": {
    "code": "MARSHAL_ERROR",
    "message": "failed to marshal output: %s"
  },
  "meta": {
    "timestamp": "%s",
    "version": "%s"
  }
}
`, err.Error(), time.Now().UTC().Format(time.RFC3339), GetVersion())
		return
	}
	fmt.Println(string(output))
}
