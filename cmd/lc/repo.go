package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
)

// tryAutoDetectForRepo 尝试为 repo 命令自动探测 workspace-key 和 git-project-id
// 如果用户已经手动指定了参数，则跳过自动检测
func tryAutoDetectForRepo() error {
	// 如果用户已经手动指定了 workspace-key，跳过自动检测
	if repoWorkspaceKey != "" {
		return nil
	}

	autoResult := common.TryAutoDetect(true)
	if autoResult.Success {
		if repoWorkspaceKey == "" {
			repoWorkspaceKey = autoResult.Context.WorkspaceKey
		}
		if repoProjectID == 0 && autoResult.Context.GitProjectID != "" {
			// 尝试将 GitProjectID 转换为 int
			if gitId, err := strconv.Atoi(autoResult.Context.GitProjectID); err == nil {
				repoProjectID = gitId
			}
		}
		return nil
	}
	return autoResult.Error
}

const (
	// DefaultRepoVisibility 默认仓库可见性 (0=公开)
	DefaultRepoVisibility = "0"
	// DefaultCodeGroupID 默认代码组ID
	DefaultCodeGroupID = 615230
	// DefaultListPageSize 默认列表分页大小
	DefaultListPageSize = 10
	// DefaultListPageNo 默认列表页码
	DefaultListPageNo = 1
)

var (
	repoWorkspaceKey string
	repoListLimit    int
	repoListPageNo   int
	repoProjectID    int // 用于存储自动探测或手动指定的 git-project-id

	// repo create flags
	repoCreateGroupID int

	// repo group list flags
	repoGroupListPageNo   int
	repoGroupListPageSize int

	// repo group add flags
	repoGroupAddPath        string
	repoGroupAddDescription string
	repoGroupAddParentID    string
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "管理代码仓库",
	Long:  `管理代码仓库，包括创建仓库、关闭提交代码关联工作项等功能。`,
}

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "新建代码仓库",
	Long: `创建一个代码仓库，需要指定仓库名称和研发空间。

自动探测支持:
  在 Git 仓库目录下执行时，会自动探测 --workspace-key
  如果自动探测失败，需要手动指定该参数

示例:
  # 在 Git 仓库目录下执行（自动探测）
  lc repo create my-repo --group-id 615230

  # 手动指定参数
  lc repo create my-repo -w XXJSLJCLIDEV --group-id 615230`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		createRepository(args[0])
	},
}

var disableWorkItemLinkCmd = &cobra.Command{
	Use:   "disable-work-item-link [git-project-id]",
	Short: "关闭提交代码关联工作项",
	Long: `关闭指定仓库的提交代码关联工作项功能，通过 Git 项目 ID 进行操作。

自动探测支持:
  在 Git 仓库目录下执行时，会自动探测 --workspace-key 和 git-project-id
  如果自动探测失败，需要手动指定这些参数

获取 Git 项目 ID:
  1. 使用 'lc repo list' 查询仓库列表（Git 仓库目录下自动探测）
  2. 从输出中的 'gitProjectId' 字段获取对应值

示例:
  # 在 Git 仓库目录下执行（自动探测）
  lc repo disable-work-item-link

  # 手动指定 git-project-id
  lc repo disable-work-item-link 44618

  # 手动指定所有参数
  lc repo disable-work-item-link 44618 --workspace-key XXJSLJCLIDEV

提示:
  使用 'lc doc git-project-id' 查看如何获取 Git 项目 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 如果提供了参数，使用参数作为 project ID
		if len(args) > 0 {
			if id, err := strconv.Atoi(args[0]); err == nil {
				repoProjectID = id
			}
		}
		disableWorkItemLink()
	},
}

var deleteRepoCmd = &cobra.Command{
	Use:   "delete [git-project-id]",
	Short: "删除代码仓库",
	Long: `删除指定的代码仓库，通过 Git 项目 ID 进行操作。

自动探测支持:
  在 Git 仓库目录下执行时，会自动探测 --workspace-key 和 git-project-id
  如果自动探测失败，需要手动指定这些参数

获取 Git 项目 ID:
  1. 使用 'lc repo list' 查询仓库列表（Git 仓库目录下自动探测）
  2. 从输出中的 'gitProjectId' 字段获取对应值

示例:
  # 在 Git 仓库目录下执行（自动探测）
  lc repo delete

  # 手动指定 git-project-id
  lc repo delete 44618

  # 手动指定所有参数
  lc repo delete 44618 --workspace-key XXJSLJCLIDEV

提示:
  使用 'lc doc git-project-id' 查看如何获取 Git 项目 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 如果提供了参数，使用参数作为 project ID
		if len(args) > 0 {
			if id, err := strconv.Atoi(args[0]); err == nil {
				repoProjectID = id
			}
		}
		deleteRepository()
	},
}

var listRepoCmd = &cobra.Command{
	Use:   "list",
	Short: "查询代码仓库列表",
	Long: `查询指定研发空间中的代码仓库列表。

自动探测支持:
  在 Git 仓库目录下执行时，会自动探测 --workspace-key
  如果自动探测失败，需要手动指定该参数

示例:
  # 在 Git 仓库目录下执行（自动探测）
  lc repo list

  # 手动指定参数
  lc repo list --workspace-key XXJSLJCLIDEV

提示:
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Run: func(cmd *cobra.Command, args []string) {
		listRepositories()
	},
}

var searchRepoCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "搜索代码仓库",
	Long: `根据名称搜索代码仓库（全局搜索，跨所有空间）。

搜索结果包含以下关键字段，可用于其他命令：
  - spaceCode:     研发空间 key，用于 lc pr create -w <spaceCode>
  - gitProjectId:  Git 项目 ID，用于 lc pr create --git-project-id <id>
  - codeGroupId:   代码组 ID，用于 lc repo create -g <id>
  - tenantId:      租户 ID

参数:
  keyword  搜索关键词（仓库名称）

示例:
  # 搜索仓库
  lc repo search ggg

  # 分页搜索
  lc repo search my --page 1 --size 20

完整 workflow 示例:
  # 1. 搜索仓库获取关键信息
  lc repo search ggg
  # 返回: spaceCode=XXJSLJCLIDEV, gitProjectId=44645, codeGroupId=617927

  # 2. 使用获取的信息创建 PR
  lc pr create -t "fix bug" -s feature-branch \
    --git-project-id 44645 -w XXJSLJCLIDEV`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchRepositories(args[0])
	},
}

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "管理仓库组",
	Long:  `管理代码仓库组，包括查询仓库组列表等功能。`,
}

var listGroupCmd = &cobra.Command{
	Use:   "list",
	Short: "查询仓库组列表",
	Long: `查询当前租户下的代码仓库组列表。

示例:
  # 查询仓库组列表（默认第1页，每页10条）
  lc repo group list

  # 查询第2页，每页20条
  lc repo group list --page 2 --size 20

  # 查询第1页，每页50条
  lc repo group list -p 1 -s 50`,
	Run: func(cmd *cobra.Command, args []string) {
		listRepoGroups()
	},
}

var addGroupCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "创建仓库组",
	Long: `创建一个新的代码仓库组。

参数:
  name  仓库组名称（必填）

示例:
  # 创建一个简单的仓库组
  lc repo group add my-group

  # 创建带描述的仓库组
  lc repo group add my-group --description "这是我的仓库组"

  # 指定路径（默认与名称相同）
  lc repo group add "我的组" --path my-group-path

  # 创建子组（指定父组ID）
  lc repo group add my-subgroup --parent-id 615230`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addRepoGroup(args[0])
	},
}

var personalGroupCmd = &cobra.Command{
	Use:   "personal",
	Short: "查询个人代码组",
	Long: `查询当前用户的个人代码组信息。

个人代码组是系统为每个用户自动创建的专属代码组，通常以用户姓名命名。
创建仓库时，可以使用此命令获取的组ID作为 --group-id 参数，将仓库创建在个人名下。

示例:
  # 查询个人代码组
  lc repo group personal

  # 结合使用：创建仓库到个人名下
  lc repo group personal | jq -r '.data.id'
  lc repo create my-repo -w XXJSLJCLIDEV --group-id $(lc repo group personal | jq -r '.data.id')`,
	Run: func(cmd *cobra.Command, args []string) {
		getPersonalGroup()
	},
}

func init() {
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(createCmd)
	repoCmd.AddCommand(disableWorkItemLinkCmd)
	repoCmd.AddCommand(deleteRepoCmd)
	repoCmd.AddCommand(listRepoCmd)
	repoCmd.AddCommand(searchRepoCmd)
	repoCmd.AddCommand(groupCmd)
	groupCmd.AddCommand(listGroupCmd)
	groupCmd.AddCommand(addGroupCmd)
	groupCmd.AddCommand(personalGroupCmd)

	// 添加各命令特有的标志
	createCmd.Flags().StringVarP(&repoWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+" (可选，支持自动探测)")
	createCmd.Flags().IntVarP(&repoCreateGroupID, "group-id", "g", 0, common.GetFlagDesc("group-id")+" (必填，使用 'lc repo group personal' 查询个人代码组 ID)")
	createCmd.MarkFlagRequired("group-id")

	disableWorkItemLinkCmd.Flags().StringVarP(&repoWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+" (可选，支持自动探测)")
	disableWorkItemLinkCmd.Flags().IntVar(&repoProjectID, "git-project-id", 0, common.GetFlagDesc("git-project-id")+" (可选，支持自动探测)")

	deleteRepoCmd.Flags().StringVarP(&repoWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+" (可选，支持自动探测)")
	deleteRepoCmd.Flags().IntVar(&repoProjectID, "git-project-id", 0, common.GetFlagDesc("git-project-id")+" (可选，支持自动探测)")

	// List command flags
	listRepoCmd.Flags().StringVarP(&repoWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+" (可选，支持自动探测)")
	listRepoCmd.Flags().IntVarP(&repoListLimit, "limit", "l", 20, common.GetFlagDesc("limit"))
	listRepoCmd.Flags().IntVarP(&repoListPageNo, "page", "p", 1, common.GetFlagDesc("page"))

	// Search command flags
	searchRepoCmd.Flags().IntVarP(&repoListLimit, "limit", "l", 20, common.GetFlagDesc("limit"))
	searchRepoCmd.Flags().IntVarP(&repoListPageNo, "page", "p", 1, common.GetFlagDesc("page"))

	// Group list command flags
	listGroupCmd.Flags().IntVarP(&repoGroupListPageNo, "page", "p", 1, common.GetFlagDesc("page"))
	listGroupCmd.Flags().IntVarP(&repoGroupListPageSize, "size", "s", 10, common.GetFlagDesc("size"))

	// Group add command flags
	addGroupCmd.Flags().StringVar(&repoGroupAddPath, "path", "", common.GetFlagDesc("path"))
	addGroupCmd.Flags().StringVar(&repoGroupAddDescription, "description", "", common.GetFlagDesc("description"))
	addGroupCmd.Flags().StringVar(&repoGroupAddParentID, "parent-id", "", common.GetFlagDesc("parent-id"))
}

func createRepository(name string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 尝试自动探测参数
		if err := tryAutoDetectForRepo(); err != nil {
			return nil, fmt.Errorf("自动探测失败: %w\n\n请手动指定参数:\n  -w, --workspace-key     研发空间 Key", err)
		}

		// 验证必需参数
		if repoWorkspaceKey == "" {
			return nil, fmt.Errorf("缺少必需参数: --workspace-key\n请指定研发空间 Key，或使用自动探测")
		}

		// Process name to extract path if it contains "/"
		parts := strings.Split(name, "/")
		var repoName, repoPath string
		if len(parts) > 1 {
			// If name contains "/", use the last part as name and full as path
			repoName = parts[len(parts)-1]
			repoPath = name
		} else {
			// Otherwise use the same value for both name and path
			repoName = name
			repoPath = name
		}

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "create",
				"resource":  "repository",
				"summary":   fmt.Sprintf("将创建代码仓库: %s", name),
				"workspace": repoWorkspaceKey,
				"request": map[string]interface{}{
					"name":        repoName,
					"path":        repoPath,
					"spaceCodes":  repoWorkspaceKey,
					"visibility":  DefaultRepoVisibility,
					"codeGroupId": repoCreateGroupID,
				},
				"simulatedResponse": map[string]interface{}{
					"name":   repoName,
					"path":   repoPath,
					"status": "pending",
				},
			}, nil
		}

		// Create project service
		headers := ctx.GetHeaders(repoWorkspaceKey)
		projectService := api.NewProjectService(ctx.Config.API.BaseRepoURL, headers, ctx.Client, ctx.Config)

		// Prepare request data - use command line workspace key
		requestData := &api.ProjectCreateRequest{
			Name:        repoName,
			Path:        repoPath,
			Visibility:  DefaultRepoVisibility,
			CodeGroupId: &repoCreateGroupID,
			SpaceCodes:  repoWorkspaceKey,
			ReadMe:      false,
			Gitignore:   false,
			IsPrivate:   false,
		}

		// Send request
		resp, err := projectService.Create(requestData)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}

		// Parse response to get repository info
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("error parsing response: %w", err)
		}

		ctx.Debug("Create response", zap.Any("response", result))

		// Check if create was successful
		success, ok := result["success"].(bool)
		if !ok || !success {
			ctx.Debug("Create failed", zap.Any("response", result))
			return nil, fmt.Errorf("create failed: %v", result["message"])
		}

		// Always query from list API to get complete repository info
		ctx.Debug("Querying repository list API")
		listResp, err := projectService.ListUserRepos(repoWorkspaceKey, DefaultListPageSize, DefaultListPageNo)
		if err != nil {
			return nil, err
		}
		defer listResp.Body.Close()

		listBody, err := io.ReadAll(listResp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading list response body: %w", err)
		}

		ctx.Debug("List response body", zap.String("body", string(listBody)))

		var listResult map[string]interface{}
		if err := json.Unmarshal(listBody, &listResult); err != nil {
			return nil, fmt.Errorf("error parsing list response: %w", err)
		}

		listData, ok := listResult["data"].([]interface{})
		if !ok || len(listData) == 0 {
			return nil, fmt.Errorf("repository not found in list response")
		}

		var repoInfo map[string]interface{}
		for _, item := range listData {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["name"] == repoName {
					repoInfo = itemMap
					break
				}
			}
		}

		if repoInfo == nil {
			return nil, fmt.Errorf("created repository not found in response")
		}

		// Extract repository URLs
		repoData := map[string]interface{}{
			"id":             repoInfo["id"],
			"name":           repoInfo["name"],
			"path":           repoInfo["path"],
			"codeGroupId":    repoInfo["codeGroupId"],
			"gitProjectId":   repoInfo["gitProjectId"],
			"gitGroupId":     repoInfo["gitGroupId"],
			"httpPath":       repoInfo["httpPath"],
			"sshPath":        repoInfo["sshPath"],
			"tenantHttpPath": repoInfo["tenantHttpPath"],
			"tenantSshPath":  repoInfo["tenantSshPath"],
		}

		// Write to log file (side effect kept for compatibility)
		writeRepoCreationLog(repoData)

		return repoData, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "repo create",
	})
}

// writeRepoCreationLog writes repository creation result to log file
// Note: This function logs errors silently to avoid breaking JSON output format
func writeRepoCreationLog(repoData map[string]interface{}) {
	jsonData, err := json.MarshalIndent(repoData, "", "  ")
	if err != nil {
		// Silently skip log writing on marshal error
		return
	}

	logFile, err := os.OpenFile("repo_creation.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Silently skip log writing on file open error
		return
	}
	defer logFile.Close()
	logFile.Write(jsonData)
	logFile.WriteString("\n")
}

func disableWorkItemLink() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 尝试自动探测参数
		if err := tryAutoDetectForRepo(); err != nil {
			return nil, fmt.Errorf("自动探测失败: %w\n\n请手动指定参数:\n  -w, --workspace-key     研发空间 Key\n  [git-project-id]        Git 项目 ID", err)
		}

		// 验证必需参数
		if repoWorkspaceKey == "" {
			return nil, fmt.Errorf("缺少必需参数: --workspace-key\n请指定研发空间 Key，或使用自动探测")
		}
		if repoProjectID == 0 {
			return nil, fmt.Errorf("缺少必需参数: [git-project-id]\n请指定 Git 项目 ID，或使用自动探测")
		}

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "disable-work-item-link",
				"resource":  "project",
				"summary":   fmt.Sprintf("将关闭项目 (ID: %d) 的代码关联工作项功能", repoProjectID),
				"workspace": repoWorkspaceKey,
				"request": map[string]interface{}{
					"projectId": repoProjectID,
				},
				"simulatedResponse": map[string]interface{}{
					"success":   true,
					"projectId": repoProjectID,
					"status":    "pending",
				},
			}, nil
		}

		// Create project card service
		headers := ctx.GetHeaders(repoWorkspaceKey)
		cardService := api.NewProjectCardService(ctx.Config.API.BaseRepoURL, headers, ctx.Client)

		// Send request to disable card
		resp, err := cardService.DisableCard(repoProjectID)
		if err != nil {
			return nil, err
		}

		ctx.Debug("Disable card response", zap.Any("response", resp))

		// Check if request was successful
		if !resp.Success {
			return nil, fmt.Errorf("failed to disable card: %s", resp.Message)
		}

		return map[string]interface{}{
			"success":   true,
			"projectId": repoProjectID,
			"project":   resp.Data.Name,
			"message":   "Successfully disabled card",
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "repo disable-work-item-link",
	})
}

func deleteRepository() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 尝试自动探测参数
		if err := tryAutoDetectForRepo(); err != nil {
			return nil, fmt.Errorf("自动探测失败: %w\n\n请手动指定参数:\n  -w, --workspace-key     研发空间 Key\n  [git-project-id]        Git 项目 ID", err)
		}

		// 验证必需参数
		if repoWorkspaceKey == "" {
			return nil, fmt.Errorf("缺少必需参数: --workspace-key\n请指定研发空间 Key，或使用自动探测")
		}
		if repoProjectID == 0 {
			return nil, fmt.Errorf("缺少必需参数: [git-project-id]\n请指定 Git 项目 ID，或使用自动探测")
		}


		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "delete",
				"resource":  "repository",
				"summary":   fmt.Sprintf("将删除代码仓库 (ID: %d)", repoProjectID),
				"workspace": repoWorkspaceKey,
				"request": map[string]interface{}{
					"projectId": repoProjectID,
				},
				"simulatedResponse": map[string]interface{}{
					"deleted":   true,
					"projectId": repoProjectID,
					"status":    "pending",
				},
			}, nil
		}

		// Create project service
		headers := ctx.GetHeaders(repoWorkspaceKey)
		projectService := api.NewProjectService(ctx.Config.API.BaseRepoURL, headers, ctx.Client, ctx.Config)

		// Send request to delete repository
		resp, err := projectService.Delete(repoProjectID)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Parse response
		var result struct {
			Success bool   `json:"success"`
			Code    string `json:"code"`
			Message string `json:"message"`
			Data    bool   `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("error parsing response: %w", err)
		}

		// Check if request was successful
		if !result.Success {
			return nil, fmt.Errorf("failed to delete repository: %s", result.Message)
		}

		return map[string]interface{}{
			"success":   true,
			"projectId": repoProjectID,
			"message":   "仓库删除成功",
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "repo delete",
	})
}

func addRepoGroup(name string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// Use name as path if path not specified
		path := repoGroupAddPath
		if path == "" {
			path = name
		}

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":   true,
				"action":   "create",
				"resource": "repository-group",
				"summary":  fmt.Sprintf("将创建代码仓库组: %s", name),
				"request": map[string]interface{}{
					"name":        name,
					"path":        path,
					"description": repoGroupAddDescription,
					"parentId":    repoGroupAddParentID,
				},
				"simulatedResponse": map[string]interface{}{
					"success": true,
					"name":    name,
					"path":    path,
					"status":  "pending",
				},
			}, nil
		}

		// Create repo group service
		headers := ctx.GetHeaders("")
		groupService := api.NewRepoGroupService(ctx.Config.API.BaseRepoURL, headers, ctx.Client)

		// Prepare request
		requestData := &api.RepoGroupCreateRequest{
			Name:        name,
			Path:        path,
			Description: repoGroupAddDescription,
			ParentID:    repoGroupAddParentID,
		}

		ctx.Debug("Creating repository group", zap.Any("request", requestData))

		// Send request
		resp, err := groupService.Create(requestData)
		if err != nil {
			return nil, err
		}

		// Check if request was successful
		if !resp.Success {
			return nil, fmt.Errorf("创建仓库组失败: %s", resp.Message)
		}

		return map[string]interface{}{
			"success":     true,
			"name":        name,
			"path":        path,
			"description": repoGroupAddDescription,
			"message":     "仓库组创建成功",
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "repo group add",
	})
}

func listRepoGroups() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// Create repo group service
		headers := ctx.GetHeaders("")
		groupService := api.NewRepoGroupService(ctx.Config.API.BaseRepoURL, headers, ctx.Client)

		// Send request to list repository groups
		resp, err := groupService.ListSubGroups(repoGroupListPageNo, repoGroupListPageSize)
		if err != nil {
			return nil, err
		}

		// Check if request was successful
		if !resp.Success {
			return nil, fmt.Errorf("查询仓库组列表失败: %s", resp.Message)
		}

		// Filter repository group fields
		var groupList []map[string]interface{}
		for _, item := range resp.Data {
			group := map[string]interface{}{
				"id":             item.ID,
				"name":           item.Name,
				"path":           item.Path,
				"description":    item.Description,
				"fullName":       item.FullName,
				"fullPath":       item.FullPath,
				"tenantFullName": item.TenantFullName,
				"tenantFullPath": item.TenantFullPath,
				"childExists":    item.ChildExists,
				"creatorName":    item.CreatorName,
				"createdBy":      item.CreatedBy,
				"createdAt":      item.CreatedAt,
			}
			groupList = append(groupList, group)
		}

		// Build output with pagination info
		return map[string]interface{}{
			"groups": groupList,
			"pagination": map[string]interface{}{
				"total":     resp.Count,
				"pageNo":    resp.PageNo,
				"pageSize":  resp.PageSize,
				"totalPage": resp.PageCount,
				"pageCount": resp.PageCount,
			},
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "repo group list",
	})
}

func listRepositories() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 尝试自动探测参数
		if err := tryAutoDetectForRepo(); err != nil {
			return nil, fmt.Errorf("自动探测失败: %w\n\n请手动指定参数:\n  -w, --workspace-key     研发空间 Key", err)
		}

		// 验证必需参数
		if repoWorkspaceKey == "" {
			return nil, fmt.Errorf("缺少必需参数: --workspace-key\n请指定研发空间 Key，或使用自动探测")
		}

		// Create project service
		headers := ctx.GetHeaders(repoWorkspaceKey)
		projectService := api.NewProjectService(ctx.Config.API.BaseRepoURL, headers, ctx.Client, ctx.Config)

		// Send request to list repositories
		resp, err := projectService.ListUserRepos(repoWorkspaceKey, repoListLimit, repoListPageNo)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Parse response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}

		if debugMode {
			ctx.Debug("List repositories response", zap.String("body", string(body)))
		}

		var result struct {
			Success   bool                     `json:"success"`
			Code      interface{}              `json:"code"`
			Message   interface{}              `json:"message"`
			Data      []map[string]interface{} `json:"data"`
			PageNo    int                      `json:"pageNo"`
			PageSize  int                      `json:"pageSize"`
			Count     int                      `json:"count"`
			PageCount int                      `json:"pageCount"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("error parsing response: %w", err)
		}

		// Check if request was successful
		if !result.Success {
			msg := "unknown error"
			if result.Message != nil {
				msg = fmt.Sprintf("%v", result.Message)
			}
			return nil, fmt.Errorf("查询仓库列表失败: %s", msg)
		}

		// Filter repository fields
		var repoList []map[string]interface{}
		for _, item := range result.Data {
			repo := map[string]interface{}{
				"id":             item["id"],
				"name":           item["name"],
				"path":           item["path"],
				"codeGroupId":    item["codeGroupId"],
				"codeGroupName":  item["codeGroupName"],
				"gitProjectId":   item["gitProjectId"],
				"gitGroupId":     item["gitGroupId"],
				"spaceCode":      item["spaceCode"],
				"httpPath":       item["httpPath"],
				"sshPath":        item["sshPath"],
				"tenantHttpPath": item["tenantHttpPath"],
				"tenantSshPath":  item["tenantSshPath"],
				"createTime":     item["createTime"],
				"creatorName":    item["creatorName"],
			}
			repoList = append(repoList, repo)
		}

		// Build output with pagination info
		return map[string]interface{}{
			"repositories": repoList,
			"pagination": map[string]interface{}{
				"total":     result.Count,
				"pageNo":    result.PageNo,
				"pageSize":  result.PageSize,
				"totalPage": result.PageCount,
				"pageCount": result.PageCount,
			},
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "repo list",
	})
}

func getPersonalGroup() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// Create repo group service
		headers := ctx.GetHeaders("")
		groupService := api.NewRepoGroupService(ctx.Config.API.BaseRepoURL, headers, ctx.Client)

		// Send request to get private group
		resp, err := groupService.GetPrivateGroup()
		if err != nil {
			return nil, err
		}

		// Check if request was successful
		if !resp.Success {
			return nil, fmt.Errorf("查询个人代码组失败: %s", resp.Message)
		}

		// Build output
		return map[string]interface{}{
			"id":             resp.Data.ID,
			"name":           resp.Data.Name,
			"path":           resp.Data.Path,
			"description":    resp.Data.Description,
			"fullName":       resp.Data.FullName,
			"fullPath":       resp.Data.FullPath,
			"tenantFullName": resp.Data.TenantFullName,
			"tenantFullPath": resp.Data.TenantFullPath,
			"permissionGroup": resp.Data.PermissionGroup,
			"permissionRepo":  resp.Data.PermissionRepo,
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "repo personal-group",
	})
}

func searchRepositories(keyword string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// Create project service - use global search (no workspace restriction)
		headers := ctx.GetHeaders("")
		projectService := api.NewProjectService(ctx.Config.API.BaseRepoURL, headers, ctx.Client, ctx.Config)

		// Global search across all workspaces
		resp, err := projectService.SearchAllUserRepos(keyword, repoListLimit, repoListPageNo)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Parse response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}

		if debugMode {
			ctx.Debug("Search repositories response", zap.String("body", string(body)))
		}

		var result struct {
			Success   bool                     `json:"success"`
			Code      interface{}              `json:"code"`
			Message   interface{}              `json:"message"`
			Data      []map[string]interface{} `json:"data"`
			PageNo    int                      `json:"pageNo"`
			PageSize  int                      `json:"pageSize"`
			Count     int                      `json:"count"`
			PageCount int                      `json:"pageCount"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("error parsing response: %w", err)
		}

		// Check if request was successful
		if !result.Success {
			msg := "unknown error"
			if result.Message != nil {
				msg = fmt.Sprintf("%v", result.Message)
			}
			return nil, fmt.Errorf("搜索仓库失败: %s", msg)
		}

		// Filter repository fields
		var repoList []map[string]interface{}
		for _, item := range result.Data {
			repo := map[string]interface{}{
				"id":             item["id"],
				"name":           item["name"],
				"path":           item["path"],
				"codeGroupId":    item["codeGroupId"],
				"codeGroupName":  item["codeGroupName"],
				"gitProjectId":   item["gitProjectId"],
				"gitGroupId":     item["gitGroupId"],
				"spaceCode":      item["spaceCode"],
				"tenantId":       item["tenantId"],
				"httpPath":       item["httpPath"],
				"sshPath":        item["sshPath"],
				"tenantHttpPath": item["tenantHttpPath"],
				"tenantSshPath":  item["tenantSshPath"],
				"createTime":     item["createTime"],
				"creatorName":    item["creatorName"],
			}
			repoList = append(repoList, repo)
		}

		// Build output with pagination info
		return map[string]interface{}{
			"repositories": repoList,
			"pagination": map[string]interface{}{
				"total":     result.Count,
				"pageNo":    result.PageNo,
				"pageSize":  result.PageSize,
				"totalPage": result.PageCount,
				"pageCount": result.PageCount,
			},
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "repo search",
	})
}
