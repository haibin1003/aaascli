// Package common provides shared utilities for command execution.
package common

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// AutoDetectContext 保存自动探测的上下文信息
type AutoDetectContext struct {
	WorkspaceKey   string                 `json:"workspaceKey"`
	WorkspaceName  string                 `json:"workspaceName"`
	TenantID       string                 `json:"tenantId"`
	GitProjectID   string                 `json:"gitProjectId"`
	Repository     map[string]interface{} `json:"repository"`
	SpaceDetails   map[string]interface{} `json:"spaceDetails"`
	GitInfo        *GitInfo               `json:"gitInfo"`
	Matched        bool                   `json:"matched"`
	MatchReason    string                 `json:"matchReason,omitempty"`
	DetectTime     time.Time              `json:"detectTime"`
}

// GitInfo 保存 Git 仓库信息
type GitInfo struct {
	IsGitRepo   bool   `json:"isGitRepo"`
	RepoName    string `json:"repoName"`
	RemoteURL   string `json:"remoteURL"`
	CurrentPath string `json:"currentPath"`
}

// 缓存相关变量
var (
	cachedDetectResult *AutoDetectContext
	cachedDetectMu     sync.RWMutex
	cacheValidDuration = 5 * time.Minute
)

// ClearAutoDetectCache 清除自动探测缓存
func ClearAutoDetectCache() {
	cachedDetectMu.Lock()
	cachedDetectResult = nil
	cachedDetectMu.Unlock()
}

// GetAutoDetectContext 获取自动探测的上下文（带缓存）
// 如果探测失败返回 nil 和错误信息
func GetAutoDetectContext() (*AutoDetectContext, error) {
	// 先检查缓存
	cachedDetectMu.RLock()
	if cachedDetectResult != nil && time.Since(cachedDetectResult.DetectTime) < cacheValidDuration {
		result := cachedDetectResult
		cachedDetectMu.RUnlock()
		return result, nil
	}
	cachedDetectMu.RUnlock()

	// 执行探测
	ctx, err := performDetect()
	if err != nil {
		return nil, err
	}

	// 更新缓存
	cachedDetectMu.Lock()
	cachedDetectResult = ctx
	cachedDetectMu.Unlock()

	return ctx, nil
}

// performDetect 执行实际的探测逻辑
func performDetect() (*AutoDetectContext, error) {
	ctx := &AutoDetectContext{
		DetectTime: time.Now(),
	}

	// 1. 探测 Git 信息
	gitInfo, err := probeGitInfo(".")
	if err != nil {
		ctx.Matched = false
		ctx.MatchReason = fmt.Sprintf("Git 探测失败: %v", err)
		return ctx, nil
	}
	ctx.GitInfo = gitInfo

	if !gitInfo.IsGitRepo {
		ctx.Matched = false
		ctx.MatchReason = "当前目录不在 Git 仓库中"
		return ctx, nil
	}

	// 2. 尝试执行 lc detect 命令获取完整信息
	detectResult, err := executeLCDetect()
	if err != nil {
		ctx.Matched = false
		ctx.MatchReason = fmt.Sprintf("lc detect 执行失败: %v", err)
		return ctx, nil
	}

	// 3. 解析 detect 结果
	if matched, ok := detectResult["matched"].(bool); !ok || !matched {
		ctx.Matched = false
		ctx.MatchReason = "未能匹配到远程仓库"
		if reason, ok := detectResult["matchReason"].(string); ok && reason != "" {
			ctx.MatchReason = reason
		}
		return ctx, nil
	}

	// 4. 提取关键字段
	ctx.Matched = true

	if workspaceKey, ok := detectResult["workspaceKey"].(string); ok {
		ctx.WorkspaceKey = workspaceKey
	}
	if workspaceName, ok := detectResult["workspaceName"].(string); ok {
		ctx.WorkspaceName = workspaceName
	}
	if tenantId, ok := detectResult["tenantId"].(string); ok {
		ctx.TenantID = tenantId
	}
	if repo, ok := detectResult["repository"].(map[string]interface{}); ok {
		ctx.Repository = repo
		// 提取 gitProjectId
		if gitProjectId, ok := repo["gitProjectId"]; ok {
			switch v := gitProjectId.(type) {
			case float64:
				ctx.GitProjectID = fmt.Sprintf("%.0f", v)
			case string:
				ctx.GitProjectID = v
			}
		}
	}
	if spaceDetails, ok := detectResult["spaceDetails"].(map[string]interface{}); ok {
		ctx.SpaceDetails = spaceDetails
	}

	return ctx, nil
}

// probeGitInfo 探测 Git 仓库信息
func probeGitInfo(path string) (*GitInfo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	info := &GitInfo{
		CurrentPath: absPath,
	}

	// 检查是否在 Git 仓库中
	cmd := exec.Command("git", "-C", absPath, "rev-parse", "--git-dir")
	if _, err := cmd.Output(); err != nil {
		info.IsGitRepo = false
		return info, nil
	}

	info.IsGitRepo = true

	// 获取仓库名称
	cmd = exec.Command("git", "-C", absPath, "rev-parse", "--show-toplevel")
	if output, err := cmd.Output(); err == nil {
		info.RepoName = filepath.Base(strings.TrimSpace(string(output)))
	}

	// 获取 remote URL
	cmd = exec.Command("git", "-C", absPath, "remote", "get-url", "origin")
	if output, err := cmd.Output(); err == nil {
		info.RemoteURL = strings.TrimSpace(string(output))
	}

	return info, nil
}

// executeLCDetect 执行 lc detect 命令获取上下文
func executeLCDetect() (map[string]interface{}, error) {
	// 检查 lc 命令是否存在
	lcPath, err := exec.LookPath("lc")
	if err != nil {
		// 尝试使用 ./bin/lc
		if _, err := os.Stat("./bin/lc"); err == nil {
			lcPath = "./bin/lc"
		} else {
			return nil, fmt.Errorf("lc 命令未找到: %w", err)
		}
	}

	cmd := exec.Command(lcPath, "detect", "-k")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("lc detect 执行失败: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("lc detect 执行失败: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析 lc detect 输出失败: %w", err)
	}

	// 提取 data 字段
	if data, ok := result["data"].(map[string]interface{}); ok {
		return data, nil
	}

	return result, nil
}

// AutoDetectResult 自动探测结果，包含是否成功和上下文信息
type AutoDetectResult struct {
	Success bool
	Context *AutoDetectContext
	Error   error
}

// TryAutoDetect 尝试自动探测，如果失败返回错误
// 用于命令执行前的参数填充
func TryAutoDetect(requireGitRepo bool) *AutoDetectResult {
	ctx, err := GetAutoDetectContext()
	if err != nil {
		return &AutoDetectResult{
			Success: false,
			Error:   err,
		}
	}

	if requireGitRepo && !ctx.GitInfo.IsGitRepo {
		return &AutoDetectResult{
			Success: false,
			Error:   fmt.Errorf("当前目录不在 Git 仓库中"),
			Context: ctx,
		}
	}

	if !ctx.Matched {
		return &AutoDetectResult{
			Success: false,
			Error:   fmt.Errorf("自动探测失败: %s", ctx.MatchReason),
			Context: ctx,
		}
	}

	return &AutoDetectResult{
		Success: true,
		Context: ctx,
	}
}

// PrintAutoDetectInfo 打印自动探测信息到 stderr
func PrintAutoDetectInfo(ctx *AutoDetectContext, logger *zap.Logger) {
	if ctx == nil || !ctx.Matched {
		return
	}

	if logger != nil {
		logger.Debug("Auto detect successful",
			zap.String("workspaceKey", ctx.WorkspaceKey),
			zap.String("workspaceName", ctx.WorkspaceName),
			zap.String("gitProjectId", ctx.GitProjectID),
		)
	}
}

// PrintAutoDetectError 返回自动探测错误（统一JSON格式，不再直接打印到stderr）
// Deprecated: Use NewAutoDetectError instead to return structured errors
func PrintAutoDetectError(err error) *AutoDetectError {
	return NewAutoDetectError("自动探测失败").
		WithDetails(err.Error()).
		WithSuggestion("请使用以下方式之一解决:\n" +
			"  1. 在 Git 仓库目录下执行命令\n" +
			"  2. 手动指定参数:\n" +
			"     -w, --workspace-key     研发空间 Key\n" +
			"     --workspace-name        研发空间名称\n" +
			"     --git-project-id        Git 项目 ID\n" +
			"\n使用 --help 查看更多参数信息")
}

// HandleAutoDetectError 处理自动探测错误，返回结构化的错误信息
// 用于替换直接打印到 stderr 的错误处理方式
func HandleAutoDetectError(err error, missingParams ...string) error {
	autoErr := NewAutoDetectError("自动探测失败").
		WithDetails(err.Error()).
		WithMissing(missingParams...)

	if len(missingParams) > 0 {
		suggestion := "请使用以下方式之一解决:\n" +
			"  1. 在 Git 仓库目录下执行命令\n" +
			"  2. 手动指定参数:\n"
		for _, param := range missingParams {
			suggestion += "     " + param + "\n"
		}
		autoErr.WithSuggestion(suggestion)
	} else {
		autoErr.WithSuggestion("请使用以下方式之一解决:\n" +
			"  1. 在 Git 仓库目录下执行命令\n" +
			"  2. 手动指定参数:\n" +
			"     -w, --workspace-key     研发空间 Key\n" +
			"     --workspace-name        研发空间名称\n" +
			"     --git-project-id        Git 项目 ID\n" +
			"\n使用 --help 查看更多参数信息")
	}

	return autoErr
}

// ExecuteWithAutoDetect runs auto-detect and then executes the command function.
// If auto-detect fails, it outputs a JSON error and exits without printing to stderr.
// This maintains unified output format.
func ExecuteWithAutoDetect(detectFunc func() error, executeFunc CommandFunc, opts ExecuteOptions) {
	if err := detectFunc(); err != nil {
		// Create a minimal context just for output
		PrintError(err)
		os.Exit(1)
	}
	Execute(executeFunc, opts)
}

// AutoDetectField 定义要自动探测的字段配置
type AutoDetectField struct {
	FlagName    string // flag 名称，如 "workspace-key"
	TargetVar   *string // 目标变量指针
	ContextKey  string // 从 AutoDetectContext 取值的字段名: "WorkspaceKey" | "WorkspaceName" | "GitProjectID"
}

// MakeAutoDetectFunc 创建一个通用的自动探测函数
// 简化命令文件中的重复代码
// 示例：tryAutoDetect = common.MakeAutoDetectFunc(cmd, []common.AutoDetectField{...})
func MakeAutoDetectFunc(fields []AutoDetectField) func(*cobra.Command) error {
	return func(cmd *cobra.Command) error {
		_, err := ApplyAutoDetect(cmd, fields)
		return err
	}
}

// ApplyAutoDetect 通用自动探测应用函数
// 根据 fields 配置自动探测并填充未指定的参数
// 返回是否使用了自动探测的值
func ApplyAutoDetect(cmd *cobra.Command, fields []AutoDetectField) (bool, error) {
	// 检查是否所有字段都已指定
	allSpecified := true
	for _, field := range fields {
		if !cmd.Flags().Changed(field.FlagName) {
			allSpecified = false
			break
		}
	}
	if allSpecified {
		return false, nil
	}

	// 尝试自动探测
	result := TryAutoDetect(true)
	if !result.Success {
		return false, result.Error
	}

	ctx := result.Context
	applied := false

	// 填充未指定的参数
	for _, field := range fields {
		if cmd.Flags().Changed(field.FlagName) {
			continue // 用户已指定，跳过
		}

		var value string
		switch field.ContextKey {
		case "WorkspaceKey":
			value = ctx.WorkspaceKey
		case "WorkspaceName":
			value = ctx.WorkspaceName
		case "GitProjectID":
			value = ctx.GitProjectID
		}

		if value != "" && field.TargetVar != nil {
			*field.TargetVar = value
			applied = true
		}
	}

	return applied, nil
}

// HandleAutoDetectWithExit runs the auto-detect function and handles errors uniformly.
// If auto-detect fails, it prints a JSON error and exits the program.
// This eliminates repetitive error handling code in command files.
//
// Usage:
//
//	common.HandleAutoDetectWithExit(func() error {
//		return tryAutoDetectForXXX(cmd)
//	}, "-w, --workspace-key")
func HandleAutoDetectWithExit(detectFunc func() error, missingParams ...string) {
	if err := detectFunc(); err != nil {
		PrintError(HandleAutoDetectError(err, missingParams...))
		os.Exit(1)
	}
}
