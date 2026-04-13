package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
)

var (
	bugTitle                 string
	bugDescription           string
	bugProjectID             string
	bugHandlerID             string
	bugLevel                 string
	bugPriority              string
	bugDefectType            string
	bugTemplateID            string
	bugWorkspaceKey          string
	bugTemplateSimple        bool
	bugListPage              int
	bugListSize              int
	bugStatusIds             []string
	bugListHandlerID         string // separate var for list command
	bugViewID                string // for view command
	bugStatusSceneID         int    // for status command
	bugUpdateStatusID        string // for update-status command
	bugUpdateStatusNewStatus string // for update-status command
	bugDeleteID              string // for delete command
)

var bugCmd = &cobra.Command{
	Use:   "bug",
	Short: "管理缺陷 (Bug/Defect)",
	Long: `管理测试中心缺陷，包括创建、查询等功能。

自动探测支持:
  list 命令支持自动探测研发空间。
  在 Git 仓库目录下执行时，无需手动指定 --workspace-key。`,
}

var bugCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新的缺陷",
	Long: `创建新的缺陷/bug。

参数说明:
  --title (-t)        缺陷标题（必填）
  --description (-D)  缺陷描述（必填）
  --project-id (-p)   项目 ID（必填，获取方式: lc space project linked -w <spaceCode>）
  --handler-id        处理人 ID
  --level (-l)        缺陷级别: 1(致命), 2(严重), 3(一般), 4(轻微)
  --priority          优先级: 1(高), 2(中), 3(低)
  --type              缺陷类型: 1(功能缺陷), 2(性能缺陷), 3(界面缺陷), 4(兼容性缺陷), 5(安全性缺陷)
  --template-id       模板 ID（自定义模板时使用）
  --template-simple   使用简洁模板（不关联版本、迭代等，默认使用完整模板）

获取项目代码(projectCode):
  使用 'lc space project list' 查看当前研发空间下的项目列表

模板说明:
  完整模板: 关联版本、迭代、系统等信息（字段较多）
  简洁模板: 只包含基本信息，无需填写版本、迭代等（推荐日常使用）

示例:
  # 使用简洁模板创建缺陷（推荐）
  lc bug create -t "登录按钮无法点击" -D "点击登录按钮无响应" -p R24113J3C04 --workspace-key XXJSxiaobaice --template-simple

  # 使用完整模板创建缺陷
  lc bug create -t "API响应慢" -D "接口响应时间超过5秒" -p R24113J3C04 -l 2 --priority 1 --workspace-key XXJSxiaobaice

提示:
  使用 'lc doc workspace-key' 查看如何获取研发空间 key
  使用 'lc space project list' 查看项目列表获取 projectCode`,
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForBugCreate(cmd)
		}, "-w, --workspace-key")
		createBug()
	},
}

var bugListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询缺陷列表",
	Long: `查询当前用户的缺陷列表。

自动探测:
  如果在 Git 仓库目录下执行，且未指定 --workspace-key，
  命令会自动探测当前目录所属的研发空间。

参数说明:
  --handler-id        处理人 ID（可选，默认查询当前用户的缺陷）
  --status-ids        缺陷状态ID列表，逗号分隔（可选）
  -p, --page          页码 (默认: 1)
  -l, --limit         每页数量 (默认: 10)

示例:
  # 自动探测并查询缺陷
  lc bug list

  # 手动指定研发空间
  lc bug list --workspace-key XXJSxiaobaice

  lc bug list --workspace-key XXJSxiaobaice -p 1 -l 20

  lc bug list --workspace-key XXJSxiaobaice --handler-id 1966878514516774920`,
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForBugList(cmd)
		}, "-w, --workspace-key")
		listBugs()
	},
}

var bugViewCmd = &cobra.Command{
	Use:   "view [defect-id]",
	Short: "查看缺陷详情",
	Long: `查看指定缺陷的详细信息。

参数说明:
  defect-id           缺陷 ID（必填，使用 lc bug list 获取）

示例:
  lc bug view 2030506279274856449 --workspace-key XXJSxiaobaice`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForBugView(cmd)
		}, "-w, --workspace-key")
		bugViewID = args[0]
		viewBug()
	},
}

var bugStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "列举可用的缺陷状态",
	Long: `列举系统中可用的缺陷状态列表。

参数说明:
  --scene-id          场景 ID (默认: 6)
  --workspace-key     研发空间 key（可选，支持自动探测）

示例:
  # 自动探测研发空间
  lc bug status

  # 手动指定研发空间
  lc bug status --workspace-key XXJSxiaobaice

  lc bug status --scene-id 6 --workspace-key XXJSxiaobaice`,
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForBugStatus(cmd)
		}, "-w, --workspace-key")
		listBugStatuses()
	},
}

var bugUpdateStatusCmd = &cobra.Command{
	Use:   "update-status [defect-id] [status-id]",
	Short: "更新缺陷状态",
	Long: `更新指定缺陷的状态。

参数说明:
  defect-id           缺陷 ID（必填）
                      使用 'lc bug list' 获取缺陷 ID
  status-id           状态 ID（必填）
                      使用 'lc bug status' 获取可用的状态 ID

如何获取参数:
  1. 缺陷 ID: 运行 'lc bug list --workspace-key XXJSxiaobaice'，从输出中获取 "id" 字段
  2. 状态 ID: 运行 'lc bug status'，从输出中获取 "statusId" 字段

示例:
  # 将缺陷 2030505333202456578 状态更新为"待发布"(1642790711574196231)
  lc bug update-status 2030505333202456578 1642790711574196231 --workspace-key XXJSxiaobaice`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForBugUpdateStatus(cmd)
		}, "-w, --workspace-key")
		bugUpdateStatusID = args[0]
		bugUpdateStatusNewStatus = args[1]
		updateBugStatus()
	},
}

var bugDeleteCmd = &cobra.Command{
	Use:   "delete [defect-id]",
	Short: "删除缺陷",
	Long: `删除指定的缺陷/Bug。

参数说明:
  defect-id           缺陷 ID（必填）
                      使用 'lc bug list' 获取缺陷 ID

如何获取缺陷 ID:
  运行 'lc bug list --workspace-key XXJSxiaobaice'，从输出中获取 "id" 字段

示例:
  # 删除缺陷 2030502335635046402
  lc bug delete 2030502335635046402 --workspace-key XXJSxiaobaice`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForBugDelete(cmd)
		}, "-w, --workspace-key")
		bugDeleteID = args[0]
		deleteBug()
	},
}

const (
	// BugSceneIdDefault 缺陷状态查询默认场景ID
	// TODO: 需要确认sceneId=6的具体含义，目前作为内置常量使用
	BugSceneIdDefault = 6
)

func init() {
	rootCmd.AddCommand(bugCmd)
	bugCmd.AddCommand(bugCreateCmd)
	bugCmd.AddCommand(bugListCmd)
	bugCmd.AddCommand(bugViewCmd)
	bugCmd.AddCommand(bugStatusCmd)
	bugCmd.AddCommand(bugUpdateStatusCmd)
	bugCmd.AddCommand(bugDeleteCmd)

	// Create command flags
	bugCreateCmd.Flags().StringVarP(&bugTitle, "title", "t", "", common.GetFlagDesc("title"))
	bugCreateCmd.Flags().StringVarP(&bugDescription, "description", "D", "", common.GetFlagDesc("description"))
	bugCreateCmd.Flags().StringVarP(&bugProjectID, "project-id", "p", "", common.GetFlagDesc("project-id"))
	bugCreateCmd.Flags().StringVar(&bugHandlerID, "handler-id", "", common.GetFlagDesc("handler-id"))
	bugCreateCmd.Flags().StringVarP(&bugLevel, "level", "l", "3", common.GetFlagDesc("level"))
	bugCreateCmd.Flags().StringVar(&bugPriority, "priority", "2", common.GetFlagDesc("priority")+": 1(高), 2(中), 3(低)")
	bugCreateCmd.Flags().StringVar(&bugDefectType, "type", "1", common.GetFlagDesc("defect-type"))
	bugCreateCmd.Flags().StringVar(&bugTemplateID, "template-id", "", common.GetFlagDesc("template-id"))
	bugCreateCmd.Flags().BoolVar(&bugTemplateSimple, "template-simple", false, common.GetFlagDesc("template-simple"))
	bugCreateCmd.Flags().StringVarP(&bugWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	bugCreateCmd.MarkFlagRequired("title")
	bugCreateCmd.MarkFlagRequired("description")
	bugCreateCmd.MarkFlagRequired("project-id")

	// List command flags
	bugListCmd.Flags().StringVar(&bugListHandlerID, "handler-id", "", common.GetFlagDesc("handler-id")+"（可选，默认查询当前用户）")
	bugListCmd.Flags().StringSliceVar(&bugStatusIds, "status-ids", nil, common.GetFlagDesc("bug-status"))
	bugListCmd.Flags().IntVarP(&bugListPage, "page", "p", 1, common.GetFlagDesc("page"))
	bugListCmd.Flags().IntVarP(&bugListSize, "limit", "l", 10, common.GetFlagDesc("limit"))
	bugListCmd.Flags().StringVarP(&bugWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// View command flags
	bugViewCmd.Flags().StringVarP(&bugWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// Status command flags
	bugStatusCmd.Flags().IntVar(&bugStatusSceneID, "scene-id", BugSceneIdDefault, common.GetFlagDesc("scene-id")+" (默认: 6)")
	bugStatusCmd.Flags().StringVarP(&bugWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// Update-status command flags
	bugUpdateStatusCmd.Flags().StringVarP(&bugWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// Delete command flags
	bugDeleteCmd.Flags().StringVarP(&bugWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
}

const (
	// BugTemplateFull 完整模板（关联版本、迭代等）
	// TODO: 模板ID应该从配置或API动态获取，不应硬编码
	BugTemplateFull = "1969947017807343618"
	// BugTemplateSimple 简洁模板（只包含基本信息）
	// TODO: 模板ID应该从配置或API动态获取，不应硬编码
	BugTemplateSimple = "1969947017924784130"
)

func createBug() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(bugWorkspaceKey)
		// Use test center base URL
		baseURL := ctx.Config.API.BaseTestCenterURL
		if baseURL == "" {
			// Fallback to constructing from platform URL
			baseURL = ctx.Config.API.BasePlatformURL + "/moss/web/cmdevops-ct/testcenter"
		}
		bugService := api.NewBugService(baseURL, headers, ctx.Client)

		// Determine template ID
		templateId := bugTemplateID
		if templateId == "" {
			if bugTemplateSimple {
				templateId = BugTemplateSimple
			} else {
				templateId = BugTemplateFull
			}
		}

		requestData := &api.BugCreateRequest{
			DefectName:  bugTitle,
			Remark:      bugDescription,
			ProjectId:   bugProjectID,
			HandlerId:   bugHandlerID,
			DefectLevel: bugLevel,
			Priority:    bugPriority,
			DefectType:  bugDefectType,
			TemplateId:  templateId,
			DefectFrom:  1,
			ColumnJson:  "{}",
			DemandId:    []string{},
			PlanId:      []string{},
			CaseId:      []string{},
			LabelList:   []string{},
			FileIds:     []string{},
		}

		// Handle dry-run mode
		if ctx.DryRun {
			templateType := "完整模板"
			if templateId == BugTemplateSimple {
				templateType = "简洁模板"
			}
			return map[string]interface{}{
				"dryRun":       true,
				"action":       "create",
				"resource":     "bug",
				"summary":      fmt.Sprintf("将创建缺陷: %s", bugTitle),
				"templateType": templateType,
				"templateId":   templateId,
				"workspace":    bugWorkspaceKey,
				"projectId":    bugProjectID,
				"request":      requestData,
			}, nil
		}

		resp, err := bugService.Create(requestData)
		if err != nil {
			return nil, err
		}

		if resp.Code != 0 {
			return nil, fmt.Errorf("create bug failed: %s (code: %d)", resp.Msg, resp.Code)
		}

		return map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Successfully created bug: %s", bugTitle),
			"data":    resp.Data,
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "bug create",
	})
}

// getPriorityName returns the display name for priority
func getPriorityName(priority string) string {
	switch priority {
	case "1":
		return "高"
	case "2":
		return "中"
	case "3":
		return "低"
	default:
		return priority
	}
}

// getLevelName returns the display name for defect level
func getLevelName(level string) string {
	switch level {
	case "1":
		return "致命"
	case "2":
		return "严重"
	case "3":
		return "一般"
	case "4":
		return "轻微"
	default:
		return level
	}
}

// bug 命令的自动探测字段配置
var bugAutoDetectBase = []common.AutoDetectField{
	{FlagName: "workspace-key", TargetVar: &bugWorkspaceKey, ContextKey: "WorkspaceKey"},
}

// tryAutoDetectForBugList 尝试为 bug list 命令自动探测参数
func tryAutoDetectForBugList(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, bugAutoDetectBase)
	return err
}

// tryAutoDetectForBugCreate 尝试为 bug create 命令自动探测参数
func tryAutoDetectForBugCreate(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, bugAutoDetectBase)
	return err
}

// tryAutoDetectForBugView 尝试为 bug view 命令自动探测参数
func tryAutoDetectForBugView(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, bugAutoDetectBase)
	return err
}

// tryAutoDetectForBugStatus 尝试为 bug status 命令自动探测参数
func tryAutoDetectForBugStatus(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, bugAutoDetectBase)
	return err
}

// tryAutoDetectForBugUpdateStatus 尝试为 bug update-status 命令自动探测参数
func tryAutoDetectForBugUpdateStatus(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, bugAutoDetectBase)
	return err
}

// tryAutoDetectForBugDelete 尝试为 bug delete 命令自动探测参数
func tryAutoDetectForBugDelete(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, bugAutoDetectBase)
	return err
}

func listBugs() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(bugWorkspaceKey)
		// Use test center base URL
		baseURL := ctx.Config.API.BaseTestCenterURL
		if baseURL == "" {
			// Fallback to constructing from platform URL
			baseURL = ctx.Config.API.BasePlatformURL + "/moss/web/cmdevops-ct/testcenter"
		}
		bugService := api.NewBugService(baseURL, headers, ctx.Client)

		// Build request - only use handlerIds if explicitly provided
		req := &api.BugListRequest{
			Page:  bugListPage,
			Limit: bugListSize,
		}

		// Add status IDs if provided
		if len(bugStatusIds) > 0 {
			req.StatusIds = bugStatusIds
		}

		// Add handler ID only if explicitly provided
		if bugListHandlerID != "" {
			req.HandlerIds = []string{bugListHandlerID}
		}

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "list",
				"resource":  "bugs",
				"summary":   "将查询缺陷列表",
				"workspace": bugWorkspaceKey,
				"page":      bugListPage,
				"size":      bugListSize,
				"request":   req,
			}, nil
		}

		resp, err := bugService.ListBugs(req)
		if err != nil {
			return nil, err
		}

		if resp.Code != 0 {
			return nil, fmt.Errorf("list bugs failed: %s (code: %d)", resp.Msg, resp.Code)
		}

		// Build simplified JSON output for AI
		var items []map[string]interface{}
		for _, bug := range resp.Data {
			item := map[string]interface{}{
				"defectName":     bug.DefectName,
				"defectCode":     bug.DefectCode,
				"priority":       bug.Priority,
				"priorityDesc":   getPriorityName(fmt.Sprintf("%d", bug.Priority)),
				"defectLevel":    bug.DefectLevel,
				"defectLevelDes": getLevelName(fmt.Sprintf("%d", bug.DefectLevel)),
				"status":         bug.DefectStatusName.Name,
				"handlerId":      bug.HandlerId,
				"handlerName":    bug.HandlerName,
				"project":        bug.ProjectName,
				"createDate":     bug.CreateDate,
				"id":             bug.ID,
			}
			items = append(items, item)
		}

		return map[string]interface{}{
			"success": true,
			"total":   resp.TotalRecCount,
			"page":    bugListPage,
			"items":   items,
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "bug list",
	})
}

func viewBug() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(bugWorkspaceKey)
		baseURL := ctx.Config.API.BaseTestCenterURL
		if baseURL == "" {
			baseURL = ctx.Config.API.BasePlatformURL + "/moss/web/cmdevops-ct/testcenter"
		}
		bugService := api.NewBugService(baseURL, headers, ctx.Client)

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "view",
				"resource":  "bug",
				"summary":   fmt.Sprintf("将查看缺陷详情: %s", bugViewID),
				"workspace": bugWorkspaceKey,
				"bugId":     bugViewID,
			}, nil
		}

		resp, err := bugService.GetBugDetail(bugViewID)
		if err != nil {
			return nil, err
		}

		if resp.Code != 0 {
			return nil, fmt.Errorf("get bug detail failed: %s (code: %d)", resp.Msg, resp.Code)
		}

		// Build simplified JSON output for AI
		bug := resp.Data
		result := map[string]interface{}{
			"id":             bug.ID,
			"defectCode":     bug.DefectCode,
			"defectName":     bug.DefectName,
			"remark":         bug.Remark,
			"priority":       bug.Priority,
			"priorityDesc":   getPriorityName(fmt.Sprintf("%d", bug.Priority)),
			"defectLevel":    bug.DefectLevel,
			"defectLevelDes": bug.DefectLevelDes,
			"status":         bug.DefectStatusName.Name,
			"statusId":       bug.DefectStatus,
			"handlerId":      bug.HandlerId,
			"handlerName":    bug.HandlerName,
			"creatorId":      bug.Creator,
			"creatorName":    bug.CreatorName,
			"projectId":      bug.ProjectId,
			"projectName":    bug.ProjectName,
			"createDate":     bug.CreateDate,
			"updateDate":     bug.UpdateDate,
		}

		// Add optional fields if present
		if bug.DefectType > 0 {
			result["defectType"] = bug.DefectType
			result["defectTypeDes"] = bug.DefectTypeDes
		}
		if bug.DefectFrom > 0 {
			result["defectFrom"] = bug.DefectFrom
			result["defectFromDesc"] = bug.DefectFromDesc
		}
		if bug.FixedVersion != "" {
			result["fixedVersion"] = bug.FixedVersion
		}
		if bug.ReleaseVersion != "" {
			result["releaseVersion"] = bug.ReleaseVersion
		}
		if bug.IterationId != "" {
			result["iterationId"] = bug.IterationId
			result["iterationName"] = bug.IterationName
		}
		if bug.SystemId != "" {
			result["systemId"] = bug.SystemId
			result["systemName"] = bug.SystemName
		}
		if bug.WorkspaceId != "" {
			result["workspaceId"] = bug.WorkspaceId
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "bug view",
	})
}

func listBugStatuses() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// Use workspace key if provided, otherwise use platform headers
		var headers map[string]string
		if bugWorkspaceKey != "" {
			headers = ctx.GetHeaders(bugWorkspaceKey)
		} else {
			headers = ctx.Config.GetPlatformHeaders()
		}
		baseURL := ctx.Config.API.BaseTestCenterURL
		if baseURL == "" {
			baseURL = ctx.Config.API.BasePlatformURL + "/moss/web/cmdevops-ct/testcenter"
		}
		bugService := api.NewBugService(baseURL, headers, ctx.Client)

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "status",
				"resource":  "bug",
				"summary":   "将查询缺陷状态列表",
				"sceneId":   bugStatusSceneID,
				"workspace": bugWorkspaceKey,
			}, nil
		}

		resp, err := bugService.GetBugStatuses(bugStatusSceneID)
		if err != nil {
			return nil, err
		}

		if resp.Code != 0 {
			return nil, fmt.Errorf("get bug statuses failed: %s (code: %d)", resp.Msg, resp.Code)
		}

		// Build simplified output with only statusId and statusName
		var items []map[string]string
		for _, status := range resp.Data {
			item := map[string]string{
				"statusId":   status.StatusId,
				"statusName": status.StatusName,
			}
			items = append(items, item)
		}

		return items, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "bug list-statuses",
	})
}

func updateBugStatus() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(bugWorkspaceKey)
		baseURL := ctx.Config.API.BaseTestCenterURL
		if baseURL == "" {
			baseURL = ctx.Config.API.BasePlatformURL + "/moss/web/cmdevops-ct/testcenter"
		}
		bugService := api.NewBugService(baseURL, headers, ctx.Client)

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "update-status",
				"resource":  "bug",
				"summary":   fmt.Sprintf("将更新缺陷 %s 状态为 %s", bugUpdateStatusID, bugUpdateStatusNewStatus),
				"workspace": bugWorkspaceKey,
				"bugId":     bugUpdateStatusID,
				"statusId":  bugUpdateStatusNewStatus,
			}, nil
		}

		resp, err := bugService.UpdateBugStatus(bugUpdateStatusID, bugUpdateStatusNewStatus)
		if err != nil {
			return nil, err
		}

		if resp.Code != 0 {
			return nil, fmt.Errorf("update bug status failed: %s (code: %d)", resp.Msg, resp.Code)
		}

		return map[string]interface{}{
			"success": true,
			"message": resp.Data,
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "bug update-status",
	})
}

func deleteBug() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(bugWorkspaceKey)
		baseURL := ctx.Config.API.BaseTestCenterURL
		if baseURL == "" {
			baseURL = ctx.Config.API.BasePlatformURL + "/moss/web/cmdevops-ct/testcenter"
		}
		bugService := api.NewBugService(baseURL, headers, ctx.Client)

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "delete",
				"resource":  "bug",
				"summary":   fmt.Sprintf("将删除缺陷: %s", bugDeleteID),
				"workspace": bugWorkspaceKey,
				"bugId":     bugDeleteID,
			}, nil
		}

		resp, err := bugService.DeleteBugs([]string{bugDeleteID})
		if err != nil {
			return nil, err
		}

		if resp.Code != 0 {
			return nil, fmt.Errorf("delete bug failed: %s (code: %d)", resp.Msg, resp.Code)
		}

		return map[string]interface{}{
			"success": true,
			"data":    resp.Data,
			"message": fmt.Sprintf("Successfully deleted bug: %s", bugDeleteID),
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "bug delete",
	})
}
