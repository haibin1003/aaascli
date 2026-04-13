package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
)

var (
	artifactWorkspaceKey       string
	artifactGroupListPage      int
	artifactGroupListLimit     int
	artifactCreateRepoKey      string
	artifactCreateRepoType     string
	artifactCreateRepoEnv      string
	artifactCreateRepoGroupID  string
	artifactListPage           int
	artifactListLimit          int
	artifactListRepoNature     int
)

var artifactCmd = &cobra.Command{
	Use:   "artifact",
	Short: "管理制品库",
	Long: `管理制品库，包括仓库组查询、创建仓库等操作。

支持的仓库类型:
  Maven, Npm, Pypi, Docker, Debian, Composer, Rpm, Go, Conan, Nuget, Generic, Cocoapods, Helm, Cargo

示例:
  # 查询仓库组列表
  lc artifact group list -w XXJSLJCLIDEV

  # 创建 Maven 仓库
  lc artifact create my-repo -w XXJSLJCLIDEV -t Maven -e DEV -g <group-id>`,
}

var artifactGroupCmd = &cobra.Command{
	Use:   "group",
	Short: "仓库组相关操作",
	Long:  `查询和管理仓库组信息。`,
}

var artifactGroupListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询仓库组列表",
	Long: `查询当前租户下的仓库组列表。

仓库组用于组织和管理制品仓库，创建仓库时需要指定所属的仓库组。

示例:
  # 查询所有仓库组
  lc artifact group list -w XXJSLJCLIDEV

  # 分页查询
  lc artifact group list -w XXJSLJCLIDEV -p 1 -l 20`,
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForArtifactGroupList(cmd)
		}, "-w, --workspace-key")
		listArtifactRepoGroups()
	},
}

var artifactCreateCmd = &cobra.Command{
	Use:   "create [repo-key]",
	Short: "创建制品仓库",
	Long: `创建一个新的制品仓库。

参数:
  repo-key - 仓库唯一标识（必需），如 "my-maven-repo"

支持的仓库类型:
  Maven, Npm, Pypi, Docker, Debian, Composer, Rpm, Go, Conan, Nuget, Generic, Cocoapods, Helm, Cargo

环境类型:
  DEV  - 开发环境
  TEST - 测试环境
  PROD - 生产环境

示例:
  # 创建 Generic 仓库（不指定分组）
  lc artifact create my-repo -w XXJSLJCLIDEV -t Generic

  # 创建 Maven 开发仓库
  lc artifact create my-maven-repo -w XXJSLJCLIDEV -t Maven -e DEV

  # 创建仓库并指定分组
  lc artifact create my-repo -w XXJSLJCLIDEV -t Generic -g <group-id>`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		artifactCreateRepoKey = args[0]
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForArtifactCreate(cmd)
		}, "-w, --workspace-key")
		createArtifactRepository()
	},
}

var artifactListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询制品仓库列表",
	Long: `查询当前空间下的制品仓库列表。

仓库类型说明:
  本地仓库(repoNature=1) - 用于存储本地构建的制品
  远程仓库(repoNature=2) - 代理远程仓库
  虚拟仓库(repoNature=3) - 聚合多个仓库

示例:
  # 查询所有制品仓库
  lc artifact list -w XXJSLJCLIDEV

  # 查询本地仓库
  lc artifact list -w XXJSLJCLIDEV -n 1

  # 分页查询
  lc artifact list -w XXJSLJCLIDEV -p 1 -l 20`,
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForArtifactList(cmd)
		}, "-w, --workspace-key")
		listArtifactRepositories()
	},
}

func init() {
	rootCmd.AddCommand(artifactCmd)
	artifactCmd.AddCommand(artifactGroupCmd)
	artifactCmd.AddCommand(artifactCreateCmd)
	artifactCmd.AddCommand(artifactListCmd)
	artifactGroupCmd.AddCommand(artifactGroupListCmd)

	// artifact group list flags
	artifactGroupListCmd.Flags().StringVarP(&artifactWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
	artifactGroupListCmd.Flags().IntVarP(&artifactGroupListPage, "page", "p", 1, common.GetFlagDesc("page"))
	artifactGroupListCmd.Flags().IntVarP(&artifactGroupListLimit, "limit", "l", 100, common.GetFlagDesc("limit"))

	// artifact create flags
	artifactCreateCmd.Flags().StringVarP(&artifactWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
	artifactCreateCmd.Flags().StringVarP(&artifactCreateRepoType, "type", "t", "", "仓库类型（必需）: Maven, Npm, Pypi, Docker, Debian, Composer, Rpm, Go, Conan, Nuget, Generic, Cocoapods, Helm, Cargo")
	artifactCreateCmd.Flags().StringVarP(&artifactCreateRepoEnv, "env", "e", "", "环境类型（可选）: DEV, TEST, PROD")
	artifactCreateCmd.Flags().StringVarP(&artifactCreateRepoGroupID, "group-id", "g", "", "仓库组ID（可选），可通过 'lc artifact group list' 获取")
	_ = artifactCreateCmd.MarkFlagRequired("type")

	// artifact list flags
	artifactListCmd.Flags().StringVarP(&artifactWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
	artifactListCmd.Flags().IntVarP(&artifactListPage, "page", "p", 1, common.GetFlagDesc("page"))
	artifactListCmd.Flags().IntVarP(&artifactListLimit, "limit", "l", 10, common.GetFlagDesc("limit"))
	artifactListCmd.Flags().IntVarP(&artifactListRepoNature, "nature", "n", 1, "仓库类型: 1=本地仓库, 2=远程仓库, 3=虚拟仓库")
}

// artifact 命令的自动探测字段配置
var (
	artifactGroupListAutoDetectBase = []common.AutoDetectField{
		{FlagName: "workspace-key", TargetVar: &artifactWorkspaceKey, ContextKey: "WorkspaceKey"},
	}
	artifactCreateAutoDetectBase = []common.AutoDetectField{
		{FlagName: "workspace-key", TargetVar: &artifactWorkspaceKey, ContextKey: "WorkspaceKey"},
	}
	artifactListAutoDetectBase = []common.AutoDetectField{
		{FlagName: "workspace-key", TargetVar: &artifactWorkspaceKey, ContextKey: "WorkspaceKey"},
	}
)

// tryAutoDetectForArtifactGroupList 尝试为 artifact group list 命令自动探测参数
func tryAutoDetectForArtifactGroupList(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, artifactGroupListAutoDetectBase)
	return err
}

// tryAutoDetectForArtifactCreate 尝试为 artifact create 命令自动探测参数
func tryAutoDetectForArtifactCreate(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, artifactCreateAutoDetectBase)
	return err
}

// tryAutoDetectForArtifactList 尝试为 artifact list 命令自动探测参数
func tryAutoDetectForArtifactList(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, artifactListAutoDetectBase)
	return err
}

// listArtifactRepoGroups lists artifact repository groups
func listArtifactRepoGroups() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(artifactWorkspaceKey)
		artifactService := api.NewArtifactService(ctx.Config.API.BaseArtifactURL, headers, ctx.Client)

		tenantID := ctx.Config.GetTenantID()
		if tenantID == "" {
			return nil, fmt.Errorf("无法获取租户ID，请检查配置")
		}

		resp, err := artifactService.ListRepoGroups(tenantID)
		if err != nil {
			return nil, err
		}

		// Format output
		var items []map[string]interface{}
		for _, group := range resp.Data {
			// 如果 repoGroupId 为空，用 id 填充
			repoGroupID := group.RepoGroupID
			if repoGroupID == "" && group.ID != 0 {
				repoGroupID = fmt.Sprintf("%d", group.ID)
			}
			items = append(items, map[string]interface{}{
				"repoGroupId":   repoGroupID,
				"repoGroupCode": group.RepoGroupCode,
				"repoGroupName": group.RepoGroupName,
			})
		}

		result := map[string]interface{}{
			"count": len(items),
			"items": items,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "artifact group list",
	})
}

// createArtifactRepository creates a new artifact repository
func createArtifactRepository() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(artifactWorkspaceKey)
		artifactService := api.NewArtifactService(ctx.Config.API.BaseArtifactURL, headers, ctx.Client)

		tenantID := ctx.Config.GetTenantID()
		if tenantID == "" {
			return nil, fmt.Errorf("无法获取租户ID，请检查配置")
		}

		user := ctx.Config.GetUser()
		deputyAccountNumber := user.Username
		if deputyAccountNumber == "" {
			return nil, fmt.Errorf("无法获取用户信息，请检查配置")
		}

		// Build repository info DTO
		repoInfo := api.RepositoryInfoDTO{
			RepoKey:             artifactCreateRepoKey,
			RepoNature:          1, // 本地仓库
			RepoType:            artifactCreateRepoType,
			RepoEnvironments:    artifactCreateRepoEnv,
			AnonymousAccess:     1,
			TenantID:            tenantID,
			SpaceCode:           artifactWorkspaceKey,
			DeputyAccountNumber: deputyAccountNumber,
			SnapshotType:        "unique",
		}

		// Set user info if available
		if user.Nickname != "" {
			repoInfo.UserName = user.Nickname
			repoInfo.CreateUsername = user.Nickname
		}

		// If group-id is provided, fetch group info and populate group fields
		var groupInfo map[string]string
		if artifactCreateRepoGroupID != "" {
			groupResp, err := artifactService.ListRepoGroups(tenantID)
			if err != nil {
				return nil, fmt.Errorf("获取仓库组信息失败: %w", err)
			}

			var selectedGroup *api.ArtifactRepoGroup
			for _, group := range groupResp.Data {
				// 支持用 id 或 repoGroupId 匹配
				groupID := group.RepoGroupID
				if groupID == "" && group.ID != 0 {
					groupID = fmt.Sprintf("%d", group.ID)
				}
				if groupID == artifactCreateRepoGroupID {
					selectedGroup = &group
					break
				}
			}

			if selectedGroup == nil {
				return nil, fmt.Errorf("未找到指定的仓库组ID: %s", artifactCreateRepoGroupID)
			}

			// 使用 id 作为 repoGroupID
			repoGroupID := fmt.Sprintf("%d", selectedGroup.ID)
			repoInfo.RepoGroupID = repoGroupID
			repoInfo.RepoGroupCode = selectedGroup.RepoGroupCode
			repoInfo.RepoGroupName = selectedGroup.RepoGroupName

			groupInfo = map[string]string{
				"repoGroupId":   repoGroupID,
				"repoGroupCode": selectedGroup.RepoGroupCode,
				"repoGroupName": selectedGroup.RepoGroupName,
			}
		}

		req := &api.CreateRepoRequest{
			RepositoryInfoDTO: repoInfo,
		}

		resp, err := artifactService.CreateRepository(req)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"repoKey":         artifactCreateRepoKey,
			"repoType":        artifactCreateRepoType,
			"repoTypeDesc":    api.GetRepoTypeDescription(artifactCreateRepoType),
			"repoEnvironment": artifactCreateRepoEnv,
			"repoEnvDesc":     api.GetRepoEnvironmentDescription(artifactCreateRepoEnv),
			"success":         resp.Success,
			"code":            resp.Code,
			"message":         resp.Message,
		}

		// Add group info if provided
		if groupInfo != nil {
			result["repoGroupId"] = groupInfo["repoGroupId"]
			result["repoGroupCode"] = groupInfo["repoGroupCode"]
			result["repoGroupName"] = groupInfo["repoGroupName"]
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "artifact create",
	})
}

// listArtifactRepositories lists artifact repositories
func listArtifactRepositories() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(artifactWorkspaceKey)
		artifactService := api.NewArtifactService(ctx.Config.API.BaseArtifactURL, headers, ctx.Client)

		tenantID := ctx.Config.GetTenantID()
		if tenantID == "" {
			return nil, fmt.Errorf("无法获取租户ID，请检查配置")
		}

		user := ctx.Config.GetUser()
		deputyAccountNumber := user.Username
		if deputyAccountNumber == "" {
			return nil, fmt.Errorf("无法获取用户信息，请检查配置")
		}

		resp, err := artifactService.GetRepositoryList(
			artifactWorkspaceKey,
			tenantID,
			deputyAccountNumber,
			artifactListRepoNature,
			artifactListPage,
			artifactListLimit,
		)
		if err != nil {
			return nil, err
		}

		// Format output
		var items []map[string]interface{}
		for _, repo := range resp.Data {
			repoMap := map[string]interface{}{
				"id":            repo.ID,
				"repoKey":       repo.RepoKey,
				"repoType":      repo.RepoType,
				"repoTypeDesc":  api.GetRepoTypeDescription(repo.RepoType),
				"repoNature":    repo.RepoNature,
				"environment":   repo.RepoEnvironments,
				"envDesc":       api.GetRepoEnvironmentDescription(repo.RepoEnvironments),
				"repoGroupName": repo.RepoGroupName,
				"readOnly":      repo.ReadOnly,
				"shareWithAll":  repo.ShareWithAll,
				"usedSize":      repo.UsedSize,
				"createTime":    repo.CreateTime,
				"createUser":    repo.CreateUsername,
			}
			items = append(items, repoMap)
		}

		result := map[string]interface{}{
			"count":     resp.Count,
			"pageNo":    resp.PageNo,
			"pageSize":  resp.PageSize,
			"pageCount": resp.PageCount,
			"items":     items,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "artifact list",
	})
}
