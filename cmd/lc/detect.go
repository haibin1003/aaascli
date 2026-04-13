package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
	"go.uber.org/zap"
)

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "自动探测当前 Git 仓库上下文",
	Long: `检测当前运行路径是否在代码仓库中，自动识别研发空间、租户等信息。

功能说明:
  1. 检测当前目录是否在 Git 仓库中
  2. 获取 Git Remote URL
  3. 搜索匹配的代码仓库
  4. 输出综合上下文信息（研发空间 Key、租户 ID、仓库信息等）

使用场景:
  在任意仓库目录下执行此命令，可自动获取上下文信息，
  后续命令可直接使用这些信息，无需手动指定 --workspace-key。

示例:
  # 探测当前目录的上下文
  lc detect

  # 在特定目录探测
  lc detect --path /path/to/repo

  # 探测并显示调试信息
  lc detect -d

脚本自动化用法:
  # 获取当前目录所属的研发空间 Key
  WORKSPACE=$(lc detect -k | jq -r '.data.workspaceKey')

  # 获取当前目录所属的研发空间名称
  WORKSPACE_NAME=$(lc detect -k | jq -r '.data.workspaceName')

  # 获取当前仓库的 Git Project ID
  PROJECT_ID=$(lc detect -k | jq -r '.data.repository.gitProjectId')

  # 结合使用：在当前仓库创建需求
  lc req create "新功能开发" -w $(lc detect -k | jq -r '.data.workspaceKey')

Shell 集成示例:
  # 在 .bashrc/.zshrc 中添加自动提示当前研发空间
  function lc_prompt() {
    if git rev-parse --git-dir >/dev/null 2>&1; then
      local ws=$(lc detect -k 2>/dev/null | jq -r '.data.workspaceName // empty')
      [[ -n "$ws" ]] && echo "[$ws]"
    fi
  }

输出信息:
  - workspaceKey:  研发空间 Key
  - workspaceName: 研发空间名称
  - tenantId:      租户 ID
  - repository:    仓库信息（含 gitProjectId、spaceCode 等）
  - spaceDetails:  研发空间详细信息
  - gitInfo:       Git 本地信息（仓库名、Remote URL 等）
  - matched:       是否成功匹配到远程仓库`,
	Run: func(cmd *cobra.Command, args []string) {
		detectContext()
	},
}

var (
	detectPath string
)

func init() {
	rootCmd.AddCommand(detectCmd)
	detectCmd.Flags().StringVarP(&detectPath, "path", "p", ".", common.GetFlagDesc("detect-path"))
}

// GitContext 保存 Git 探测信息
type GitContext struct {
	IsGitRepo   bool
	RepoName    string
	RemoteURL   string
	GitPath     string
	CurrentPath string
}

// DetectedContext 保存探测到的完整上下文
type DetectedContext struct {
	WorkspaceKey   string                 `json:"workspaceKey"`
	WorkspaceName  string                 `json:"workspaceName,omitempty"`
	TenantID       string                 `json:"tenantId,omitempty"`
	Repository     map[string]interface{} `json:"repository"`
	SpaceDetails   map[string]interface{} `json:"spaceDetails,omitempty"`
	GitInfo        *GitContext            `json:"gitInfo"`
	Matched        bool                   `json:"matched"`
	MatchReason    string                 `json:"matchReason,omitempty"`
}

func detectContext() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 步骤 1: 探测 Git 上下文
		gitCtx, err := probeGitContext(detectPath)
		if err != nil {
			return nil, fmt.Errorf("探测 Git 上下文失败: %w", err)
		}

		if !gitCtx.IsGitRepo {
			return &DetectedContext{
				Matched:     false,
				MatchReason: "当前目录不在 Git 仓库中",
				GitInfo:     gitCtx,
			}, nil
		}

		ctx.Debug("Git context detected", zap.Any("gitContext", gitCtx))

		// 步骤 2: 搜索匹配的仓库
		repository, err := findMatchingRepository(ctx, gitCtx)
		if err != nil {
			ctx.Debug("Failed to find matching repository", zap.Error(err))
			return &DetectedContext{
				Matched:     false,
				MatchReason: fmt.Sprintf("搜索仓库失败: %v", err),
				GitInfo:     gitCtx,
			}, nil
		}

		if repository == nil {
			return &DetectedContext{
				Matched:     false,
				MatchReason: "未找到匹配的仓库",
				GitInfo:     gitCtx,
			}, nil
		}

		// 步骤 3: 构建完整上下文
		result := buildDetectedContext(ctx, repository, gitCtx)
		return result, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
	})
}

// probeGitContext 探测 Git 上下文信息
func probeGitContext(path string) (*GitContext, error) {
	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("获取绝对路径失败: %w", err)
	}

	ctx := &GitContext{
		CurrentPath: absPath,
	}

	// 检查是否在 Git 仓库中
	cmd := exec.Command("git", "-C", absPath, "rev-parse", "--git-dir")
	output, err := cmd.Output()
	if err != nil {
		// 不在 Git 仓库中
		ctx.IsGitRepo = false
		return ctx, nil
	}

	ctx.IsGitRepo = true
	ctx.GitPath = strings.TrimSpace(string(output))

	// 获取仓库名称（从目录名）
	cmd = exec.Command("git", "-C", absPath, "rev-parse", "--show-toplevel")
	topLevel, err := cmd.Output()
	if err == nil {
		topLevelPath := strings.TrimSpace(string(topLevel))
		ctx.RepoName = filepath.Base(topLevelPath)
	}

	// 获取 remote URL
	cmd = exec.Command("git", "-C", absPath, "remote", "get-url", "origin")
	remoteURL, err := cmd.Output()
	if err == nil {
		ctx.RemoteURL = strings.TrimSpace(string(remoteURL))
	}

	return ctx, nil
}

// findMatchingRepository 根据 Git 上下文搜索匹配的仓库
func findMatchingRepository(ctx *common.CommandContext, gitCtx *GitContext) (map[string]interface{}, error) {
	if gitCtx.RepoName == "" {
		return nil, fmt.Errorf("无法获取仓库名称")
	}

	// 创建项目服务
	headers := ctx.GetHeaders("")
	projectService := api.NewProjectService(ctx.Config.API.BaseRepoURL, headers, ctx.Client, ctx.Config)

	// 搜索仓库
	ctx.Debug("Searching repository", zap.String("name", gitCtx.RepoName))
	resp, err := projectService.SearchAllUserRepos(gitCtx.RepoName, 20, 1)
	if err != nil {
		return nil, fmt.Errorf("搜索仓库失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Success   bool                     `json:"success"`
		Data      []map[string]interface{} `json:"data"`
		Count     int                      `json:"count"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("搜索仓库请求失败")
	}

	ctx.Debug("Search results", zap.Int("count", len(result.Data)))

	// 匹配仓库
	for _, repo := range result.Data {
		if matchRepository(repo, gitCtx) {
			return repo, nil
		}
	}

	// 如果没有精确匹配，返回第一个同名仓库
	for _, repo := range result.Data {
		if name, ok := repo["name"].(string); ok && name == gitCtx.RepoName {
			return repo, nil
		}
	}

	return nil, nil
}

// matchRepository 检查仓库是否匹配 Git 上下文
func matchRepository(repo map[string]interface{}, gitCtx *GitContext) bool {
	if gitCtx.RemoteURL == "" {
		// 没有 remote URL，只能按名称匹配
		if name, ok := repo["name"].(string); ok && name == gitCtx.RepoName {
			return true
		}
		return false
	}

	// 尝试匹配各种 URL 字段
	urlFields := []string{"httpPath", "sshPath", "tenantHttpPath", "tenantSshPath"}
	remoteURL := normalizeGitURL(gitCtx.RemoteURL)

	for _, field := range urlFields {
		if url, ok := repo[field].(string); ok && url != "" {
			normalizedRepoURL := normalizeGitURL(url)
			if normalizedRepoURL == remoteURL {
				return true
			}
		}
	}

	// 按名称匹配
	if name, ok := repo["name"].(string); ok && name == gitCtx.RepoName {
		return true
	}

	return false
}

// normalizeGitURL 标准化 Git URL 用于比较
func normalizeGitURL(url string) string {
	// 移除 .git 后缀
	url = strings.TrimSuffix(url, ".git")

	// 统一协议前缀
	url = strings.ReplaceAll(url, "https://", "http://")
	url = strings.ReplaceAll(url, "ssh://", "")
	url = strings.ReplaceAll(url, "git@", "")

	// 移除端口号 (简单的字符串替换，移除 :8022 这种端口号)
	// 找到 http:// 或 https:// 后的第一个冒号
	if idx := strings.Index(url, "://"); idx != -1 {
		hostPath := url[idx+3:]
		// 在主机路径部分查找第一个冒号（端口号）
		if colonIdx := strings.Index(hostPath, ":"); colonIdx != -1 {
			// 检查冒号后是否有数字
			slashIdx := strings.Index(hostPath[colonIdx:], "/")
			if slashIdx == -1 {
				slashIdx = len(hostPath) - colonIdx
			}
			portPart := hostPath[colonIdx+1 : colonIdx+slashIdx]
			isPort := true
			for _, c := range portPart {
				if c < '0' || c > '9' {
					isPort = false
					break
				}
			}
			if isPort {
				// 移除端口号
				hostPath = hostPath[:colonIdx] + hostPath[colonIdx+slashIdx:]
				url = url[:idx+3] + hostPath
			}
		}
	}

	// 统一路径分隔符（对于 SSH 格式 git@host:path/to/repo）
	url = strings.ReplaceAll(url, ":", "/")

	// 转小写
	return strings.ToLower(url)
}

// buildDetectedContext 构建探测结果
func buildDetectedContext(ctx *common.CommandContext, repo map[string]interface{}, gitCtx *GitContext) *DetectedContext {
	result := &DetectedContext{
		Matched:    true,
		Repository: filterRepoFields(repo),
		GitInfo:    gitCtx,
	}

	// 提取 Workspace Key
	if spaceCode, ok := repo["spaceCode"].(string); ok && spaceCode != "" {
		result.WorkspaceKey = spaceCode
	}

	// 提取 Tenant ID
	if tenantId, ok := repo["tenantId"].(string); ok && tenantId != "" {
		result.TenantID = tenantId
	}

	// 获取研发空间详细信息
	if result.WorkspaceKey != "" {
		spaceDetails := fetchSpaceDetails(ctx, result.WorkspaceKey)
		if spaceDetails != nil {
			result.SpaceDetails = spaceDetails
			// 优先使用 spaceName 字段
			if name, ok := spaceDetails["spaceName"].(string); ok && name != "" {
				result.WorkspaceName = name
			} else if name, ok := spaceDetails["name"].(string); ok && name != "" {
				// 兼容性：尝试 name 字段
				result.WorkspaceName = name
			}
		}
	}

	return result
}

// fetchSpaceDetails 获取研发空间详细信息
func fetchSpaceDetails(ctx *common.CommandContext, workspaceKey string) map[string]interface{} {
	headers := ctx.GetHeaders(workspaceKey)
	spaceService := api.NewSpaceService(ctx.Config.API.BasePlatformURL, headers, ctx.Client)

	resp, err := spaceService.GetSpaceDetail(workspaceKey)
	if err != nil {
		ctx.Debug("Failed to fetch space details", zap.Error(err))
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	// 列表 API 返回的是数组格式
	var listResult struct {
		Success   bool                     `json:"success"`
		Data      []map[string]interface{} `json:"data"`
		PageNo    int                      `json:"pageNo"`
		PageSize  int                      `json:"pageSize"`
		Count     int                      `json:"count"`
		PageCount int                      `json:"pageCount"`
	}

	if err := json.Unmarshal(body, &listResult); err != nil {
		ctx.Debug("Failed to unmarshal space list response", zap.Error(err))
		return nil
	}

	if !listResult.Success {
		return nil
	}

	// 在列表中查找匹配的 workspace
	for _, space := range listResult.Data {
		if spaceCode, ok := space["spaceCode"].(string); ok && spaceCode == workspaceKey {
			return space
		}
	}

	return nil
}

// filterRepoFields 过滤仓库字段
func filterRepoFields(repo map[string]interface{}) map[string]interface{} {
	fields := []string{
		"id", "name", "path", "gitProjectId", "gitGroupId",
		"codeGroupId", "codeGroupName", "spaceCode",
		"httpPath", "sshPath", "tenantHttpPath", "tenantSshPath",
		"createTime", "creatorName",
	}

	filtered := make(map[string]interface{})
	for _, field := range fields {
		if val, ok := repo[field]; ok {
			filtered[field] = val
		}
	}
	return filtered
}
