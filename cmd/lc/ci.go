package cmd

import (
	"github.com/spf13/cobra"
	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
)

var (
	ciListWorkspaceKey      string
	ciListStatus            string
	ciListLimit             int
	ciListPage              int
	ciHistoryWorkspaceKey   string
	ciHistoryTaskID         string
	ciHistoryLimit          int
	ciHistoryPage           int
	ciHistoryViewSnapshotID string
	ciHistoryLogSnapshotID  string
)

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "管理 CI 构建任务",
	Long: `查询 CI（持续集成）构建任务列表和历史记录。

示例:
  # 查询所有构建任务
  lc ci list -w XXJSLJCLIDEV

  # 按状态筛选构建任务
  lc ci list -w XXJSLJCLIDEV -s 3

  # 查询构建历史
  lc ci history -w XXJSLJCLIDEV -t 205338bc206c4a05a6d6a72f88ab5aa0`,
}

var ciListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询构建任务列表",
	Long: `查询研发空间中的 CI/CD 构建任务列表。

构建状态说明:
  2 - 执行中
  3 - 成功
  4 - 失败
  5 - 已停止

示例:
  # 查询所有构建任务
  lc ci list -w XXJSLJCLIDEV

  # 查询成功的构建
  lc ci list -w XXJSLJCLIDEV -s 3

  # 查询失败的构建
  lc ci list -w XXJSLJCLIDEV -s 4

  # 分页查询，每页10条
  lc ci list -w XXJSLJCLIDEV -p 2 -l 10`,
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForCIList(cmd)
		}, "-w, --workspace-key")
		listCIBuilds()
	},
}

var ciHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "查询构建任务历史记录",
	Long: `查询指定构建任务的历史执行记录。

构建状态说明:
  2 - 执行中
  3 - 成功
  4 - 失败
  5 - 已停止

示例:
  # 查询指定任务的构建历史
  lc ci history -w XXJSLJCLIDEV -t 205338bc206c4a05a6d6a72f88ab5aa0

  # 分页查询历史记录
  lc ci history -w XXJSLJCLIDEV -t 205338bc206c4a05a6d6a72f88ab5aa0 -p 2 -l 10`,
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForCIHistory(cmd)
		}, "-w, --workspace-key")
		getCIBuildHistory()
	},
}

var ciHistoryViewCmd = &cobra.Command{
	Use:   "view [snapshot-id]",
	Short: "查看构建详情",
	Long: `查看指定构建快照的详细信息，包括构建步骤执行情况。

参数:
  snapshot-id - 构建快照ID，可以从 'lc ci history' 命令的输出中获取

示例:
  # 查看构建详情
  lc ci history view 0b6c83659ef64511905713ba212c4f76

  # 使用自动探测的 workspace
  lc ci history view 0b6c83659ef64511905713ba212c4f76 -k`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ciHistoryViewSnapshotID = args[0]
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForCIHistoryView(cmd)
		}, "-w, --workspace-key")
		getCIBuildDetail()
	},
}

var ciHistoryLogCmd = &cobra.Command{
	Use:   "log [snapshot-id]",
	Short: "查看构建日志",
	Long: `查看指定构建快照的完整日志内容。

参数:
  snapshot-id - 构建快照ID，可以从 'lc ci history' 命令的输出中获取

示例:
  # 查看构建日志
  lc ci history log 0b6c83659ef64511905713ba212c4f76

  # 使用自动探测的 workspace
  lc ci history log 0b6c83659ef64511905713ba212c4f76 -k`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ciHistoryLogSnapshotID = args[0]
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForCIHistoryLog(cmd)
		}, "-w, --workspace-key")
		getCIBuildLog()
	},
}

func init() {
	rootCmd.AddCommand(ciCmd)
	ciCmd.AddCommand(ciListCmd)
	ciCmd.AddCommand(ciHistoryCmd)
	ciHistoryCmd.AddCommand(ciHistoryViewCmd)

	ciListCmd.Flags().StringVarP(&ciListWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
	ciListCmd.Flags().StringVarP(&ciListStatus, "status", "s", "", "构建状态过滤: 2=执行中, 3=成功, 4=失败, 5=已停止")
	ciListCmd.Flags().IntVarP(&ciListLimit, "limit", "l", 10, common.GetFlagDesc("limit"))
	ciListCmd.Flags().IntVarP(&ciListPage, "page", "p", 1, common.GetFlagDesc("page"))

	ciHistoryCmd.Flags().StringVarP(&ciHistoryWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
	ciHistoryCmd.Flags().StringVarP(&ciHistoryTaskID, "task-id", "t", "", "构建任务ID（必需）")
	ciHistoryCmd.Flags().IntVarP(&ciHistoryLimit, "limit", "l", 10, common.GetFlagDesc("limit"))
	ciHistoryCmd.Flags().IntVarP(&ciHistoryPage, "page", "p", 1, common.GetFlagDesc("page"))
	_ = ciHistoryCmd.MarkFlagRequired("task-id")

	ciHistoryViewCmd.Flags().StringVarP(&ciHistoryWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	ciHistoryCmd.AddCommand(ciHistoryLogCmd)
	ciHistoryLogCmd.Flags().StringVarP(&ciHistoryWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
}

// listCIBuilds lists CI builds in the workspace
func listCIBuilds() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(ciListWorkspaceKey)
		ciService := api.NewCIService(ctx.Config.API.BaseCIURL, headers, ctx.Client)

		resp, err := ciService.ListBuilds(ciListWorkspaceKey, ciListStatus, ciListPage, ciListLimit)
		if err != nil {
			return nil, err
		}

		// Output results as JSON
		var items []map[string]interface{}
		for _, build := range resp.Data {
			buildMap := map[string]interface{}{
				"id":             build.ID,
				"taskName":       build.TaskName,
				"buildStatus":    api.GetBuildStatusText(build.BuildStatus),
				"buildStatusCode": build.BuildStatus,
				"buildNumber":    build.BuildNumber,
				"buildType":      build.BuildType,
				"lastBuildTime":  build.LastBuildTime,
				"createTime":     build.CreateTime,
				"createUserName": build.CreateUserName,
			}

			// Add code snapshot info if available
			if build.BuildCodeSnapshot != nil {
				buildMap["vcsName"] = build.BuildCodeSnapshot.VcsName
				buildMap["vcsBranch"] = build.BuildCodeSnapshot.VcsBranch
				buildMap["commitId"] = build.BuildCodeSnapshot.CommitID
				if build.BuildCodeSnapshot.CommitMsg != nil {
					buildMap["commitMessage"] = build.BuildCodeSnapshot.CommitMsg.Title
					buildMap["commitAuthor"] = build.BuildCodeSnapshot.CommitMsg.AuthorName
				}
			}

			items = append(items, buildMap)
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
		CommandName: "ci list",
	})
}

// ci 命令的自动探测字段配置
var (
	ciListAutoDetectBase    = []common.AutoDetectField{
		{FlagName: "workspace-key", TargetVar: &ciListWorkspaceKey, ContextKey: "WorkspaceKey"},
	}
	ciHistoryAutoDetectBase = []common.AutoDetectField{
		{FlagName: "workspace-key", TargetVar: &ciHistoryWorkspaceKey, ContextKey: "WorkspaceKey"},
	}
	ciHistoryViewAutoDetectBase = []common.AutoDetectField{
		{FlagName: "workspace-key", TargetVar: &ciHistoryWorkspaceKey, ContextKey: "WorkspaceKey"},
	}
	ciHistoryLogAutoDetectBase = []common.AutoDetectField{
		{FlagName: "workspace-key", TargetVar: &ciHistoryWorkspaceKey, ContextKey: "WorkspaceKey"},
	}
)

// tryAutoDetectForCIList 尝试为 ci list 命令自动探测参数
func tryAutoDetectForCIList(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, ciListAutoDetectBase)
	return err
}

// tryAutoDetectForCIHistory 尝试为 ci history 命令自动探测参数
func tryAutoDetectForCIHistory(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, ciHistoryAutoDetectBase)
	return err
}

// tryAutoDetectForCIHistoryView 尝试为 ci history view 命令自动探测参数
func tryAutoDetectForCIHistoryView(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, ciHistoryViewAutoDetectBase)
	return err
}

// tryAutoDetectForCIHistoryLog 尝试为 ci history log 命令自动探测参数
func tryAutoDetectForCIHistoryLog(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, ciHistoryLogAutoDetectBase)
	return err
}

// getCIBuildLog gets the build log for a specific build snapshot
func getCIBuildLog() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(ciHistoryWorkspaceKey)
		ciService := api.NewCIService(ctx.Config.API.BaseCIURL, headers, ctx.Client)

		resp, err := ciService.GetBuildLog(ciHistoryLogSnapshotID)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"snapshotId": ciHistoryLogSnapshotID,
			"log":        resp.Data,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "ci history log",
	})
}

// getCIBuildHistory gets the build history for a specific task
func getCIBuildHistory() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(ciHistoryWorkspaceKey)
		ciService := api.NewCIService(ctx.Config.API.BaseCIURL, headers, ctx.Client)

		resp, err := ciService.GetBuildHistory(ciHistoryWorkspaceKey, ciHistoryTaskID, ciHistoryPage, ciHistoryLimit)
		if err != nil {
			return nil, err
		}

		// Output results as JSON
		var items []map[string]interface{}
		for _, build := range resp.Data {
			buildMap := map[string]interface{}{
				"id":             build.ID,
				"buildNumber":    build.BuildNumber,
				"buildStatus":    api.GetBuildStatusText(build.BuildStatus),
				"buildStatusCode": build.BuildStatus,
				"duration":       api.FormatDuration(build.Duration),
				"durationMs":     build.Duration,
				"startTime":      build.StartTime,
				"endTime":        build.EndTime,
				"source":         build.Source,
				"createTime":     build.CreateTime,
				"createUserName": build.CreateUserName,
			}

			// Add code snapshot info if available
			if build.BuildCodeSnapshot != nil {
				buildMap["vcsName"] = build.BuildCodeSnapshot.VcsName
				buildMap["vcsBranch"] = build.BuildCodeSnapshot.VcsBranch
				buildMap["commitId"] = build.BuildCodeSnapshot.CommitID
				if build.BuildCodeSnapshot.CommitMsg != nil {
					buildMap["commitMessage"] = build.BuildCodeSnapshot.CommitMsg.Title
					buildMap["commitAuthor"] = build.BuildCodeSnapshot.CommitMsg.AuthorName
				}
			}

			items = append(items, buildMap)
		}

		result := map[string]interface{}{
			"count":     resp.Count,
			"pageNo":    resp.PageNo,
			"pageSize":  resp.PageSize,
			"pageCount": resp.PageCount,
			"taskId":    ciHistoryTaskID,
			"items":     items,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "ci history",
	})
}

// getCIBuildDetail gets the detailed information for a specific build snapshot
func getCIBuildDetail() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(ciHistoryWorkspaceKey)
		ciService := api.NewCIService(ctx.Config.API.BaseCIURL, headers, ctx.Client)

		resp, err := ciService.GetBuildDetail(ciHistoryViewSnapshotID)
		if err != nil {
			return nil, err
		}

		build := resp.Data

		// Format build steps
		var steps []map[string]interface{}
		for _, step := range build.BuildStepSnapshots {
			stepMap := map[string]interface{}{
				"id":       step.ID,
				"name":     step.Name,
				"type":     step.Type,
				"serial":   step.Serial,
				"stepState": api.GetStepStateText(step.StepState),
				"stepStateCode": step.StepState,
				"stepResult": api.GetStepResultText(step.StepResult),
				"stepResultCode": step.StepResult,
				"hasOpen":  step.HasOpen,
			}

			// Add optional fields if present
			if step.StartTime != nil {
				stepMap["startTime"] = *step.StartTime
			}
			if step.EndTime != nil {
				stepMap["endTime"] = *step.EndTime
			}
			if step.Duration != nil {
				stepMap["duration"] = api.FormatDuration(*step.Duration)
				stepMap["durationMs"] = *step.Duration
			}
			if step.StepLogPath != nil {
				stepMap["stepLogPath"] = *step.StepLogPath
			}

			steps = append(steps, stepMap)
		}

		// Format the main result
		result := map[string]interface{}{
			"id":          build.ID,
			"taskId":      build.TaskID,
			"buildNumber": build.BuildNumber,
			"buildStatus": api.GetBuildStatusText(build.BuildStatus),
			"buildStatusCode": build.BuildStatus,
			"duration":    api.FormatDuration(build.Duration),
			"durationMs":  build.Duration,
			"startTime":   build.StartTime,
			"endTime":     build.EndTime,
			"source":      build.Source,
			"createTime":  build.CreateTime,
			"steps":       steps,
			"stepCount":   len(steps),
		}

		// Add code snapshot info if available
		if build.BuildCodeSnapshot != nil {
			result["vcsName"] = build.BuildCodeSnapshot.VcsName
			result["vcsBranch"] = build.BuildCodeSnapshot.VcsBranch
			result["commitId"] = build.BuildCodeSnapshot.CommitID
			if build.BuildCodeSnapshot.CommitMsg != nil {
				result["commitMessage"] = build.BuildCodeSnapshot.CommitMsg.Title
				result["commitAuthor"] = build.BuildCodeSnapshot.CommitMsg.AuthorName
			}
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "ci history view",
	})
}
