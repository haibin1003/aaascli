package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
	"github.com/user/lc/internal/config"
)

var (
	taskFile             string
	taskWorkspaceKey     string
	taskProjectCode      string
	taskListLimit        int
	taskListOffset       int
	taskListReqID        string
	taskListWorkspace    string
	taskListWorkspaceKey string
)

// TaskYAML represents the YAML structure for creating a task
type TaskYAML struct {
	Name                string   `yaml:"name"`
	RequirementID       string   `yaml:"requirementId,omitempty"`
	TaskType            []string `yaml:"taskType,omitempty"`
	TaskDescription     string   `yaml:"taskDescription,omitempty"`
	PlannedWorkingHours int      `yaml:"plannedWorkingHours,omitempty"`
	PlannedStartTime    int64    `yaml:"plannedStartTime,omitempty"`
	PlannedEndTime      int64    `yaml:"plannedEndTime,omitempty"`
	EgreeOfImportance   []string `yaml:"egreeOfImportance,omitempty"`
	Assignee            UserYAML `yaml:"assignee,omitempty"`
	Priority            string   `yaml:"priority,omitempty"`
	ProjectCode         string   `yaml:"projectCode,omitempty"`
}

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "管理任务",
	Long: `在需求下创建任务，或查询/删除任务。

示例:
  # 在需求下创建简单任务
  lc task create nCKhZGcBKu "实现登录功能" -w workspaceKey

  # 从 YAML 文件创建任务
  lc task create -f task.yaml -w workspaceKey

  # 从标准输入创建任务
  cat task.yaml | lc task create -w workspaceKey

  # 查询任务列表
  lc task list -w workspaceKey

  # 查询指定需求下的任务
  lc task list -w workspaceKey -r nCKhZGcBKu

YAML 格式示例:
  name: 任务名称
  requirementId: nCKhZGcBKu
  taskType:
    - 开发
  taskDescription: |
    任务详细描述
  plannedWorkingHours: 8
  assignee:
    label: "魏宝辉(weibaohui@hq.cmcc)"
    value: weibaohui@hq.cmcc
    username: weibaohui@hq.cmcc
    nickname: 魏宝辉`,
}

var taskCreateCmd = &cobra.Command{
	Use:   "create [requirement-id] [name]",
	Short: "创建任务",
	Long: `在需求下创建任务。

参数获取:
  --project-code 获取方式:
    lc space projects --workspace-key XXJSLJCLIDEV
    # 从输出中的 items[].projectCode 字段获取

示例:
  # 在需求下创建简单任务
  lc task create nCKhZGcBKu "实现登录功能" -w workspaceKey --project-code R24113J3C04

  # 从 YAML 文件创建任务
  lc task create -f task.yaml -w workspaceKey

  # 从标准输入创建任务
  cat task.yaml | lc task create -w workspaceKey`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForTaskCreate(cmd)
		}, "-w, --workspace-key")
		createTask(args)
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询任务列表",
	Long: `查询研发空间中的任务列表。

示例:
  # 查询所有任务（默认20条）
  lc task list -k -w "小白测研发项目"

  # 限制返回数量
  lc task list -k -w "小白测研发项目" -l 10

  # 查询指定需求下的任务
  lc task list -k -w "小白测研发项目" -r nCKhZGcBKu

  # 分页查询
  lc task list -k -w "小白测研发项目" -l 5 -o 10`,
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForTaskList(cmd)
		}, "-w, --workspace-key")
		listTasks()
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete [task-id] [flags]",
	Short: "删除任务",
	Long: `删除单个任务。

获取 task-id:
  1. 使用 'lc task list --workspace-key <key>' 查询任务列表
  2. 从输出中的 'id' 或 'objectId' 字段获取对应值

示例:
  lc task delete nCKhZGcBKu --workspace-key XXJSxiaobaice

提示:
  使用 'lc doc task-id' 查看如何获取任务 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForTaskDelete(cmd)
		}, "-w, --workspace-key")
		deleteTask(args[0])
	},
}

var taskSearchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "搜索任务",
	Long: `根据关键词搜索任务。

支持在任务标题中进行模糊搜索。

示例:
  # 搜索包含 "server" 的任务
  lc task search server --workspace-key XXJSxiaobaice

  # 搜索并指定分页
  lc task search mock --workspace-key XXJSxiaobaice --limit 10 --offset 0`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForTaskSearch(cmd)
		}, "-w, --workspace-key")
		searchTasks(args[0])
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskSearchCmd)

	taskCreateCmd.Flags().StringVarP(&taskFile, "filename", "f", "", common.GetFlagDesc("filename"))
	taskCreateCmd.Flags().StringVarP(&taskWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
	taskCreateCmd.Flags().StringVar(&taskProjectCode, "project-code", "", common.GetFlagDesc("project-code")+"（可选，获取方式: lc space project linked -w <spaceCode>）")

	taskListCmd.Flags().IntVarP(&taskListLimit, "limit", "l", 20, common.GetFlagDesc("limit"))
	taskListCmd.Flags().IntVarP(&taskListOffset, "offset", "o", 0, common.GetFlagDesc("offset"))
	taskListCmd.Flags().StringVarP(&taskListReqID, "requirement", "r", "", "按父"+common.GetFlagDesc("requirement-id")+"过滤任务")
	taskListCmd.Flags().StringVarP(&taskListWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	taskDeleteCmd.Flags().StringVarP(&taskWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// Search command flags
	taskSearchCmd.Flags().StringVarP(&taskListWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
	taskSearchCmd.Flags().IntVarP(&taskListLimit, "limit", "l", 20, common.GetFlagDesc("limit"))
	taskSearchCmd.Flags().IntVarP(&taskListOffset, "offset", "o", 0, common.GetFlagDesc("offset"))
}

// createTask creates a task under a requirement
func createTask(args []string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		var taskYAML *TaskYAML
		var requirementID string

		// Try to read from file, stdin, or pipe
		data, err := common.ReadYAMLFromInput(taskFile)
		if err == nil {
			// Parse YAML from input
			taskYAML = &TaskYAML{}
			if err := common.ParseYAML(data, taskYAML); err != nil {
				return nil, err
			}
			if taskYAML.Name == "" {
				return nil, fmt.Errorf("'name' field is required in YAML")
			}
			requirementID = taskYAML.RequirementID
		} else if len(args) >= 2 {
			// 命令行参数: requirement-id 和 name
			requirementID = args[0]
			taskYAML = &TaskYAML{Name: args[1], RequirementID: requirementID}
		} else {
			return nil, fmt.Errorf("either provide [requirement-id] [name] arguments, use -f flag, or pipe YAML content")
		}

		// 验证 requirementID
		if requirementID == "" {
			return nil, fmt.Errorf("'requirementId' is required")
		}

		// Handle dry-run mode early (before API calls)
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":        true,
				"action":        "create",
				"resource":      "task",
				"summary":       fmt.Sprintf("将创建任务: %s", taskYAML.Name),
				"workspace":     taskWorkspaceKey,
				"requirementId": requirementID,
				"request": map[string]interface{}{
					"name":          taskYAML.Name,
					"requirementId": requirementID,
					"taskType":      "开发",
					"workspace":     taskWorkspaceKey,
				},
				"simulatedResponse": map[string]interface{}{
					"name":   taskYAML.Name,
					"key":    "DRY-RUN-TASK",
					"status": "pending",
				},
			}, nil
		}

		headers := ctx.GetHeaders(taskWorkspaceKey)
		reqService := api.NewRequirementService(ctx.Config.API.BaseReqURL, headers, ctx.Client, ctx.Config)

		// 获取 workspace objectId
		spaceService := api.NewSpaceService(ctx.Config.API.BasePlatformURL, headers, ctx.Client)
		workspaceObjectId, err := spaceService.GetWorkspaceObjectId(taskWorkspaceKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get workspace objectId: %w", err)
		}

		// 构建任务请求
		requestData := buildTaskCreateRequest(taskYAML, requirementID, ctx.Config, workspaceObjectId, taskWorkspaceKey, taskProjectCode)

		resp, err := reqService.CreateTask(requestData, taskWorkspaceKey)
		if err != nil {
			return nil, err
		}

		// 输出创建结果
		result := map[string]interface{}{
			"success":   true,
			"name":      resp.Name,
			"objectId":  resp.ObjectID,
			"key":       resp.Key,
			"ancestors": resp.Ancestors,
			"createdAt": resp.CreatedAt,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "task create",
	})
}

// buildTaskCreateRequest 从 YAML 构建任务创建请求
func buildTaskCreateRequest(yaml *TaskYAML, requirementID string, cfg *config.Config, workspaceObjectId, spaceCode, projectCode string) *api.TaskCreateRequest {
	now := time.Now().UnixMilli()
	sevenDaysLater := now + 7*24*60*60*1000
	uniqueID := generateUniqueID()

	// 任务类型 Object ID 从配置读取，Workspace ID 从命令行参数读取
	taskItemTypeID := cfg.Defaults.TaskItemTypeID
	if taskItemTypeID == "" {
		taskItemTypeID = "Co5QtC7wdQ" // 默认任务类型
	}

	req := &api.TaskCreateRequest{
		Name:      yaml.Name,
		Ancestors: []string{requirementID},
		Workspace: api.Workspace{
			Type:      "Pointer",
			ClassName: "Workspace",
			ObjectID:  workspaceObjectId,
		},
		ItemType: api.ItemType{
			Type:      "Pointer",
			ClassName: "ItemType",
			ObjectID:  taskItemTypeID,
		},
		Reporter:    nil,
		ItemContext: api.ItemContext{},
		ParseContext: api.TaskParseContext{
			EventExtraData:            map[string]interface{}{},
			SkipValidateOptionsFields: []string{"relatedAchievements", "associatedVersion"},
		},
	}

	// 构建 Values
	values := api.TaskValues{
		ProjectNo:           []interface{}{},
		Relations:           nil,
		ScreenType:          "create",
		PlannedStartTime:    now,
		PlannedEndTime:      sevenDaysLater,
		PlannedWorkingHours: yaml.PlannedWorkingHours,
	}

	// 设置关联项目编号 - 优先使用 YAML 中的 projectCode，其次使用命令行参数
	if yaml.ProjectCode != "" {
		values.ProjectNo = []interface{}{yaml.ProjectCode}
	} else if projectCode != "" {
		values.ProjectNo = []interface{}{projectCode}
	}

	// 设置 YAML 值或默认值
	if len(yaml.TaskType) > 0 {
		values.TaskType = yaml.TaskType
	} else {
		values.TaskType = []string{"开发"}
	}

	// 计划工时默认为8小时
	if yaml.PlannedWorkingHours == 0 {
		values.PlannedWorkingHours = 8
	}

	if len(yaml.EgreeOfImportance) > 0 {
		values.EgreeOfImportance = yaml.EgreeOfImportance
	} else {
		values.EgreeOfImportance = []string{"不重要不紧急"}
	}

	// 优先级从配置或YAML读取
	if yaml.Priority != "" {
		values.Priority = yaml.Priority
	} else {
		values.Priority = cfg.Defaults.PriorityID
	}

	if yaml.PlannedStartTime > 0 {
		values.PlannedStartTime = yaml.PlannedStartTime
	}
	if yaml.PlannedEndTime > 0 {
		values.PlannedEndTime = yaml.PlannedEndTime
	}

	// 处理负责人
	if yaml.Assignee.Username != "" {
		values.Assignee = []api.UserValue{{
			Label:    yaml.Assignee.Label,
			Value:    yaml.Assignee.Value,
			Username: yaml.Assignee.Username,
			Nickname: yaml.Assignee.Nickname,
			Deleted:  false,
			Enabled:  true,
		}}
	} else {
		user := cfg.GetUser()
		values.Assignee = []api.UserValue{{
			Label:    user.Label(),
			Value:    user.Username,
			Username: user.Username,
			Nickname: user.Nickname,
			Deleted:  false,
			Enabled:  true,
		}}
	}

	// 处理任务描述
	if yaml.TaskDescription != "" {
		values.TaskDescription = textToEditorContent(yaml.TaskDescription, "task"+uniqueID)
	} else {
		values.TaskDescription = textToEditorContent(yaml.Name, "task"+uniqueID)
	}

	req.Values = values
	return req
}

// listTasks lists tasks in the workspace
func listTasks() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 如果只传了 workspace-key 没传 workspace-name，自动获取
		if taskListWorkspaceKey != "" && taskListWorkspace == "" {
			spaceService := api.NewSpaceService(ctx.Config.API.BasePlatformURL, ctx.GetHeaders(taskListWorkspaceKey), ctx.Client)
			if name, err := spaceService.GetSpaceNameByCode(taskListWorkspaceKey); err == nil {
				taskListWorkspace = name
			}
		}

		headers := ctx.GetHeaders(taskListWorkspaceKey)
		reqService := api.NewRequirementService(ctx.Config.API.BaseReqURL, headers, ctx.Client, ctx.Config)

		// Build IQL query
		// Workspace is required for task queries
		if taskListWorkspace == "" {
			return nil, fmt.Errorf("研发空间名称不能为空，请使用 -n 参数指定\n示例: lc task list -k -n \"小白测研发项目\" -w XXJSxiaobaice")
		}

		iql := fmt.Sprintf("((类型 in [\"任务\"]) and (所属空间 = '%s')) order by 创建时间 desc", taskListWorkspace)
		if taskListReqID != "" {
			iql = fmt.Sprintf("((类型 in [\"任务\"]) and (所属空间 = '%s') and (ancestors = '%s')) order by 创建时间 desc", taskListWorkspace, taskListReqID)
		}

		requestData := &api.TaskListRequest{
			IQL:    iql,
			Size:   taskListLimit,
			From:   taskListOffset,
			Extend: map[string]interface{}{},
			Fields: []string{
				"ancestors", "assignee", "createdAt", "createdBy",
				"earlyWarning", "id", "itemType", "key",
				"plannedEndTime", "plannedStartTime", "priority",
				"rowId", "status", "workspace",
			},
			RefererInfo: api.RefererInfo{
				WorkspaceKey: taskListWorkspaceKey,
			},
		}

		resp, err := reqService.ListTasks(requestData, taskListWorkspaceKey)
		if err != nil {
			return nil, err
		}

		// Output results as JSON
		var items []map[string]interface{}
		for _, task := range resp.Payload.Items {
			taskMap := map[string]interface{}{
				"name":      task.Name,
				"key":       task.Key,
				"objectId":  task.ObjectID,
				"status":    task.Status.Name,
				"createdAt": task.CreatedAt,
			}

			// Add assignee if available
			if len(task.Values.Assignee) > 0 {
				taskMap["assignee"] = task.Values.Assignee[0].Nickname
			}

			// Add planned dates if available
			if task.Values.PlannedStartTime != nil {
				startTime := common.FormatTimestamp(task.Values.PlannedStartTime)
				if startTime != "" {
					taskMap["plannedStartTime"] = startTime
				}
			}
			if task.Values.PlannedEndTime != nil {
				endTime := common.FormatTimestamp(task.Values.PlannedEndTime)
				if endTime != "" {
					taskMap["plannedEndTime"] = endTime
				}
			}

			// Add ancestors (parent requirement)
			if len(task.Ancestors) > 0 {
				taskMap["ancestors"] = task.Ancestors
			}

			items = append(items, taskMap)
		}

		result := map[string]interface{}{
			"count":  resp.Payload.Count,
			"offset": taskListOffset,
			"items":  items,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "task list",
	})
}

// deleteTask deletes a task by its object ID
func deleteTask(taskID string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(taskWorkspaceKey)
		reqService := api.NewRequirementService(ctx.Config.API.BaseReqURL, headers, ctx.Client, ctx.Config)

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":   true,
				"action":   "delete",
				"resource": "task",
				"summary":  fmt.Sprintf("将删除任务: %s", taskID),
				"request": map[string]interface{}{
					"taskId": taskID,
				},
				"simulatedResponse": map[string]interface{}{
					"deleted": 1,
					"taskId":  taskID,
					"status":  "pending",
				},
			}, nil
		}

		resp, err := reqService.DeleteTasks([]string{taskID}, taskWorkspaceKey)
		if err != nil {
			return nil, fmt.Errorf("删除任务失败: %w", err)
		}

		result := map[string]interface{}{
			"success": resp.Code == 0,
			"code":    resp.Code,
			"deleted": 1,
			"taskId":  taskID,
			"message": resp.Message,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "task delete",
	})
}

// searchTasks searches tasks by keyword
func searchTasks(keyword string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 如果只传了 workspace-key 没传 workspace-name，自动获取
		if taskListWorkspaceKey != "" && taskListWorkspace == "" {
			spaceService := api.NewSpaceService(ctx.Config.API.BasePlatformURL, ctx.GetHeaders(taskListWorkspaceKey), ctx.Client)
			if name, err := spaceService.GetSpaceNameByCode(taskListWorkspaceKey); err == nil {
				taskListWorkspace = name
			}
		}

		headers := ctx.GetHeaders(taskListWorkspaceKey)
		reqService := api.NewRequirementService(ctx.Config.API.BaseReqURL, headers, ctx.Client, ctx.Config)

		// Build IQL query for task search
		// Search by title or key, filtered by workspace and task type
		iql := fmt.Sprintf("((((标题 ~ '%s' or key = '%s')) and (所属空间 = '%s')) and ('类型' in [\"任务\"])) order by 创建时间 desc",
			keyword, keyword, taskListWorkspace)

		requestData := &api.SearchTasksRequest{
			IQL:    iql,
			Size:   taskListLimit,
			From:   taskListOffset,
			Extend: map[string]interface{}{},
			Fields: []string{
				"ancestors", "assignee", "createdAt", "createdBy",
				"earlyWarning", "id", "itemType", "key",
				"plannedEndTime", "plannedStartTime", "priority",
				"rowId", "status", "workspace",
			},
			RefererInfo: api.RefererInfo{
				WorkspaceKey: taskListWorkspaceKey,
			},
		}

		resp, err := reqService.SearchTasks(requestData, taskListWorkspaceKey)
		if err != nil {
			return nil, err
		}

		// Output search results as JSON
		var items []map[string]interface{}
		for _, task := range resp.Payload.Items {
			taskMap := map[string]interface{}{
				"name":      task.Name,
				"key":       task.Key,
				"objectId":  task.ObjectID,
				"status":    task.Status.Name,
				"createdAt": task.CreatedAt,
			}

			// Add assignee if available
			if len(task.Values.Assignee) > 0 {
				taskMap["assignee"] = task.Values.Assignee[0].Nickname
			}

			// Add planned dates if available
			if task.Values.PlannedStartTime != nil {
				startTime := common.FormatTimestamp(task.Values.PlannedStartTime)
				if startTime != "" {
					taskMap["plannedStartTime"] = startTime
				}
			}
			if task.Values.PlannedEndTime != nil {
				endTime := common.FormatTimestamp(task.Values.PlannedEndTime)
				if endTime != "" {
					taskMap["plannedEndTime"] = endTime
				}
			}

			// Add ancestors (parent requirement)
			if len(task.Ancestors) > 0 {
				taskMap["ancestors"] = task.Ancestors
			}

			items = append(items, taskMap)
		}

		result := map[string]interface{}{
			"keyword":    keyword,
			"count":      resp.Payload.Count,
			"offset":     taskListOffset,
			"items":      items,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "task search",
	})
}

// task 命令的自动探测字段配置
var (
	taskAutoDetectBase = []common.AutoDetectField{
		{FlagName: "workspace-key", TargetVar: &taskWorkspaceKey, ContextKey: "WorkspaceKey"},
	}
	taskListAutoDetectFull = []common.AutoDetectField{
		{FlagName: "workspace-key", TargetVar: &taskListWorkspaceKey, ContextKey: "WorkspaceKey"},
	}
)

// tryAutoDetectForTaskCreate 尝试为 task create 命令自动探测参数
func tryAutoDetectForTaskCreate(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, taskAutoDetectBase)
	return err
}

// tryAutoDetectForTaskList 尝试为 task list 命令自动探测参数
func tryAutoDetectForTaskList(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, taskListAutoDetectFull)
	return err
}

// tryAutoDetectForTaskDelete 尝试为 task delete 命令自动探测参数
func tryAutoDetectForTaskDelete(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, taskAutoDetectBase)
	return err
}

// tryAutoDetectForTaskSearch 尝试为 task search 命令自动探测参数
func tryAutoDetectForTaskSearch(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, taskListAutoDetectFull)
	return err
}

