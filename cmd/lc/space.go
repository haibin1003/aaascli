package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
)

var (
	spacePageNo           int
	spacePageSize         int
	spaceProjectPage      int
	spaceProjectSize      int
	spaceWorkspaceKey     string
	spaceProjectKeyword   string
	spaceLinkedProjectPage int
	spaceLinkedProjectSize int
)

var spaceCmd = &cobra.Command{
	Use:   "space",
	Short: "管理研发空间",
	Long:  `管理研发空间，包括查询空间列表、项目列表等功能。`,
}

var spaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询研发空间列表",
	Long:  `查询当前用户有权限访问的所有研发空间列表。`,
	Run: func(cmd *cobra.Command, args []string) {
		common.Execute(listSpacesExec, common.ExecuteOptions{
			DebugMode:   debugMode,
			Insecure:    insecureSkipVerify,
			Logger:      &logger,
			PrettyPrint: prettyMode,
		})
	},
}

var spaceProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "管理研发空间项目",
	Long:  `管理研发空间下的项目，包括查询可关联项目、已关联项目等功能。`,
}

var spaceProjectAvailableCmd = &cobra.Command{
	Use:   "available",
	Short: "查询可关联的项目",
	Long: `查询当前租户下所有可被关联到空间的项目列表。

这些项目存在于项目库中，可以通过空间管理界面关联到空间。
项目代码(projectCode)可用于其他命令，如：
  - lc bug create -p <projectCode> ...

示例:
  lc space project available
  lc space project available --page 1 --size 20

提示:
  使用 'lc space project linked' 查看已关联到空间的项目`,
	Run: func(cmd *cobra.Command, args []string) {
		common.Execute(listProjectsExec, common.ExecuteOptions{
			DebugMode:   debugMode,
			Insecure:    insecureSkipVerify,
			Logger:      &logger,
			PrettyPrint: prettyMode,
		})
	},
}

var spaceProjectLinkedCmd = &cobra.Command{
	Use:   "linked",
	Short: "查询已关联的项目",
	Long: `查询指定研发空间已关联的项目列表。

此命令返回的 projectCode 是创建需求和任务时必须关联的参数：
  - lc req create "需求名称" --project-code <projectCode> -w <workspace-key>
  - lc task create <req-objectId> "任务名称" --project-code <projectCode> -w <workspace-key>

自动探测:
  如果在 Git 仓库目录下执行，且未指定 --workspace-key，
  命令会自动探测当前目录所属的研发空间。

参数:
  --workspace-key  研发空间 key（可选，支持自动探测）

示例:
  # 自动探测并查询关联项目
  lc space projects

  # 指定空间查询
  lc space projects --workspace-key XXJSLJCLIDEV

  # 搜索特定项目名称
  lc space projects --workspace-key XXJSLJCLIDEV --keyword "云原生"

提示:
  使用 'lc space list' 查看可用的研发空间列表`,
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForSpaceProjects(cmd)
		}, "-w, --workspace-key")
		common.Execute(listSpaceProjectsExec, common.ExecuteOptions{
			DebugMode:   debugMode,
			Insecure:    insecureSkipVerify,
			Logger:      &logger,
			PrettyPrint: prettyMode,
		})
	},
}

func init() {
	rootCmd.AddCommand(spaceCmd)
	spaceCmd.AddCommand(spaceListCmd)
	spaceCmd.AddCommand(spaceProjectCmd)
	spaceProjectCmd.AddCommand(spaceProjectAvailableCmd)
	spaceProjectCmd.AddCommand(spaceProjectLinkedCmd)

	// List command flags
	spaceListCmd.Flags().IntVarP(&spacePageNo, "page", "p", 1, common.GetFlagDesc("page"))
	spaceListCmd.Flags().IntVarP(&spacePageSize, "page-size", "s", 1000, common.GetFlagDesc("page-size"))

	// Project available command flags
	spaceProjectAvailableCmd.Flags().IntVar(&spaceProjectPage, "page", 1, common.GetFlagDesc("page"))
	spaceProjectAvailableCmd.Flags().IntVar(&spaceProjectSize, "size", 10, common.GetFlagDesc("limit"))

	// Project linked command flags
	spaceProjectLinkedCmd.Flags().StringVarP(&spaceWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
	spaceProjectLinkedCmd.Flags().StringVar(&spaceProjectKeyword, "keyword", "", common.GetFlagDesc("keyword"))
	spaceProjectLinkedCmd.Flags().IntVar(&spaceLinkedProjectPage, "page", 1, common.GetFlagDesc("page"))
	spaceProjectLinkedCmd.Flags().IntVar(&spaceLinkedProjectSize, "size", 10, common.GetFlagDesc("limit"))
}

func listSpacesExec(ctx *common.CommandContext) (interface{}, error) {
	// Use platform headers for space API
	headers := ctx.Config.GetPlatformHeaders()
	spaceService := api.NewSpaceService(ctx.Config.API.BasePlatformURL, headers, ctx.Client)

	resp, err := spaceService.List(spacePageNo, spacePageSize)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parse response
	var spaceResp api.SpaceListResponse
	if err := json.Unmarshal(body, &spaceResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if !spaceResp.Success {
		// Check if error is in head field (alternative error format)
		var errResp api.APIErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Head.RespCode != "" && errResp.Head.RespCode != "00" {
			return nil, fmt.Errorf("%s", api.FormatAIError(errResp.Head.RespDesc))
		}
		// Use AI-friendly error formatting
		errMsg := spaceResp.Message
		if errMsg == "" {
			errMsg = "space list request failed"
		}
		return nil, fmt.Errorf("%s", api.FormatAIError(errMsg))
	}

	// Filter to only required fields
	var resultSpaces []map[string]interface{}
	for _, space := range spaceResp.Data {
		spaceMap := map[string]interface{}{
			"id":             space.ID,
			"belongOrgName":  space.BelongOrgName,
			"createTime":     space.CreateTime,
			"spaceCode":      space.SpaceCode,
			"spaceDesc":      space.SpaceDesc,
			"spaceName":      space.SpaceName,
			"spaceOwnerName": space.SpaceOwnerName,
			"spaceRoleNames": space.SpaceRoleNames,
			"spaceState":     space.SpaceState,
			"templateName":   space.TemplateName,
			"updateTime":     space.UpdateTime,
		}
		resultSpaces = append(resultSpaces, spaceMap)
	}

	return resultSpaces, nil
}

func listProjectsExec(ctx *common.CommandContext) (interface{}, error) {
	// Use platform headers for project API
	headers := ctx.Config.GetPlatformHeaders()
	// Use project management base URL
	baseURL := ctx.Config.API.BaseProjectURL
	if baseURL == "" {
		// Fallback
		baseURL = ctx.Config.API.BasePlatformURL + "/moss/web/cmdevops-project/server/api/v1"
	}
	projectService := api.NewProjectBaseService(baseURL, headers, ctx.Client)

	resp, err := projectService.ListProjects(spaceProjectPage, spaceProjectSize)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		errMsg := resp.Message
		if errMsg == "" {
			errMsg = "project list request failed"
		}
		return nil, fmt.Errorf("%s", api.FormatAIError(errMsg))
	}

	// Filter to only required fields
	var resultProjects []map[string]interface{}
	for _, proj := range resp.Data {
		projMap := map[string]interface{}{
			"projectCode":        proj.ProjectCode,
			"projectName":        proj.ProjectName,
			"projectManagerName": proj.ProjectManagerName,
			"deptName":           proj.DeptName,
			"projectStatus":      proj.ProjectStatus,
			"createTime":         proj.CreateTime,
			"planFinishTime":     proj.PlanFinishTime,
		}
		resultProjects = append(resultProjects, projMap)
	}

	return map[string]interface{}{
		"success":   true,
		"total":     resp.Count,
		"page":      resp.PageNo,
		"pageSize":  resp.PageSize,
		"pageCount": resp.PageCount,
		"items":     resultProjects,
	}, nil
}

func listSpaceProjectsExec(ctx *common.CommandContext) (interface{}, error) {
	// Step 1: Get space details to find spaceId
	headers := ctx.Config.GetPlatformHeaders()
	spaceService := api.NewSpaceService(ctx.Config.API.BasePlatformURL, headers, ctx.Client)

	// Get space list with filter by spaceCode
	resp, err := spaceService.List(1, 1000)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var spaceResp api.SpaceListResponse
	if err := json.Unmarshal(body, &spaceResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if !spaceResp.Success {
		return nil, fmt.Errorf("获取空间列表失败: %s", spaceResp.Message)
	}

	// Find space by spaceCode
	var spaceID int
	for _, space := range spaceResp.Data {
		if space.SpaceCode == spaceWorkspaceKey {
			spaceID = space.ID
			break
		}
	}

	if spaceID == 0 {
		return nil, fmt.Errorf("未找到研发空间: %s", spaceWorkspaceKey)
	}

	// Step 2: Query linked projects using spaceId
	projectResp, err := spaceService.ListProjects(spaceID, spaceProjectKeyword, spaceLinkedProjectPage, spaceLinkedProjectSize)
	if err != nil {
		return nil, err
	}

	if !projectResp.Success {
		return nil, fmt.Errorf("获取关联项目失败: %s", projectResp.Message)
	}

	// Format output
	var resultProjects []map[string]interface{}
	for _, proj := range projectResp.Data {
		projMap := map[string]interface{}{
			"id":                  proj.ID,
			"projectCode":         proj.ProjectCode,
			"projectName":         proj.ProjectName,
			"projectManagerName":  proj.ProjectManagerName,
			"projectStatus":       proj.ProjectStatus,
			"projectSource":       proj.ProjectSource,
			"deptName":            proj.DeptName,
			"projectCategoryName": proj.ProjectCategoryName,
			"createName":          proj.CreateName,
			"createTime":          proj.CreateTime,
		}
		resultProjects = append(resultProjects, projMap)
	}

	return map[string]interface{}{
		"success":   true,
		"total":     projectResp.Count,
		"page":      projectResp.PageNo,
		"pageSize":  projectResp.PageSize,
		"pageCount": projectResp.PageCount,
		"items":     resultProjects,
	}, nil
}

// tryAutoDetectForSpaceProjects 尝试为 space projects 命令自动探测参数
// 如果用户已指定参数，则使用用户指定的值
// 如果未指定，则尝试自动探测
func tryAutoDetectForSpaceProjects(cmd *cobra.Command) error {
	// 检查用户是否已指定 workspace-key
	workspaceKeySet := cmd.Flags().Changed("workspace-key")

	// 如果已指定，无需自动探测
	if workspaceKeySet {
		return nil
	}

	// 尝试自动探测
	result := common.TryAutoDetect(true)
	if !result.Success {
		// 探测失败，返回错误
		return result.Error
	}

	ctx := result.Context

	// 填充未指定的参数（静默填充，不打印信息以保持 JSON 输出纯净）
	if !workspaceKeySet && ctx.WorkspaceKey != "" {
		spaceWorkspaceKey = ctx.WorkspaceKey
	}

	return nil
}
