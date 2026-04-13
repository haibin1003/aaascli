package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
)

var (
	title          string
	body           string
	sourceBranch   string
	targetBranch   string
	projectId      int
	removeSource   bool
	mrId           int
	reviewType     string
	mergeType      string
	squash         bool
	rebase         bool
	mergeFlag      bool
	deleteBranch   bool
	comment        string
	prWorkspaceKey string
	showComments   bool
	prListState    string
	prListPage     int
	prListSize     int
	commentId      int
	commentState   string
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "管理合并请求 (Merge Request)",
	Long:  `管理合并请求，包括创建、审核、合并 MR 等功能。`,
}

var prCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新的合并请求",
	Long: `创建新的合并请求 (MR)，需要指定标题、源分支和目标分支。

自动探测支持:
  在 Git 仓库目录下执行时，会自动探测以下参数：
  - --workspace-key       从 Git 远程 URL 探测研发空间
  - --git-project-id      从 Git 远程 URL 探测项目 ID
  - --source              自动获取当前 Git 分支
  如果自动探测失败，需要手动指定相应参数

获取 Git 项目 ID:
  1. 使用 'lc repo list --workspace-key <workspace-key>' 查询仓库列表
  2. 从输出中的 'gitProjectId' 字段获取对应值

示例:
  # 在 Git 仓库目录下执行（完全自动探测）
  lc pr create -t "修复登录bug"

  # 指定源分支（在 Git 仓库目录下仍会自动探测其他参数）
  lc pr create -t "修复登录bug" -s feature-branch

  # 手动指定所有参数
  lc pr create -t "修复登录bug" -s feature-branch --target master --git-project-id 44142 --workspace-key XXJSxiaobaice

提示:
  使用 'lc doc git-project-id' 查看如何获取 Git 项目 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Run: func(cmd *cobra.Command, args []string) {
		createPullRequest()
	},
}

var prReviewCmd = &cobra.Command{
	Use:   "review [mr-id]",
	Short: "审核合并请求",
	Long: `审核合并请求，支持批准或拒绝。类似 'gh pr review --approve'。

参数说明:
  [mr-id]          合并请求 ID（MR ID），对应 'lc pr list' 输出中的 'iid' 字段

自动探测支持:
  在 Git 仓库目录下执行时，会自动探测 --workspace-key 和 --git-project-id
  如果自动探测失败，需要手动指定这些参数

获取 Git 项目 ID:
  1. 使用 'lc repo list --workspace-key <workspace-key>' 查询仓库列表
  2. 从输出中的 'gitProjectId' 字段获取对应值

获取 MR ID:
  1. 使用 'lc pr list' 查询 MR 列表（Git 仓库目录下自动探测）
  2. 从输出中的 'iid' 字段获取对应值

示例:
  # 在 Git 仓库目录下执行（自动探测）
  lc pr review 123 --type approve

  # 手动指定参数
  lc pr review 123 --git-project-id 44142 --workspace-key XXJSxiaobaice --type approve
  lc pr review 123 --git-project-id 44142 --workspace-key XXJSxiaobaice --type reject --body "需要修改"

提示:
  使用 'lc doc mr-id' 查看如何获取 MR ID
  使用 'lc doc git-project-id' 查看如何获取 Git 项目 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`, Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
				return nil, fmt.Errorf("mr-id must be a number")
			}, common.ExecuteOptions{DebugMode: debugMode, Insecure: insecureSkipVerify, DryRun: dryRunMode, Logger: &logger})
			return
		}
		mrId = id
		reviewMergeRequest()
	},
}

var prMergeCmd = &cobra.Command{
	Use:   "merge [mr-id]",
	Short: "合并合并请求",
	Long: `合并合并请求。兼容 'gh pr merge' 风格参数。

参数说明:
  [mr-id]          合并请求 ID（MR ID），对应 'lc pr list' 输出中的 'iid' 字段

自动探测支持:
  在 Git 仓库目录下执行时，会自动探测 --workspace-key 和 --git-project-id
  如果自动探测失败，需要手动指定这些参数

获取 Git 项目 ID:
  1. 使用 'lc repo list --workspace-key <workspace-key>' 查询仓库列表
  2. 从输出中的 'gitProjectId' 字段获取对应值

获取 MR ID:
  1. 使用 'lc pr list' 查询 MR 列表（Git 仓库目录下自动探测）
  2. 从输出中的 'iid' 字段获取对应值

示例:
  # gh 风格命令（推荐）
  lc pr merge 123 --squash --delete-branch
  lc pr merge 123 --rebase --delete-branch
  lc pr merge 123 --merge

  # lc 原有风格命令（向后兼容）
  lc pr merge 123 --type squash --remove-source
  lc pr merge 123 --type rebase -r

  # 手动指定参数
  lc pr merge 123 --git-project-id 44142 --workspace-key XXJSxiaobaice --squash --delete-branch

提示:
  使用 'lc doc mr-id' 查看如何获取 MR ID
  使用 'lc doc git-project-id' 查看如何获取 Git 项目 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
				return nil, fmt.Errorf("mr-id must be a number")
			}, common.ExecuteOptions{DebugMode: debugMode, Insecure: insecureSkipVerify, DryRun: dryRunMode, Logger: &logger})
			return
		}
		mrId = id
		mergeMergeRequest()
	},
}

var prCommentCmd = &cobra.Command{
	Use:   "comment [mr-id]",
	Short: "评论合并请求",
	Long: `在合并请求下发表评论。

参数说明:
  [mr-id]          合并请求 ID（MR ID），对应 'lc pr list' 输出中的 'iid' 字段

自动探测支持:
  在 Git 仓库目录下执行时，会自动探测 --workspace-key 和 --git-project-id
  如果自动探测失败，需要手动指定这些参数

获取 Git 项目 ID:
  1. 使用 'lc repo list --workspace-key <workspace-key>' 查询仓库列表
  2. 从输出中的 'gitProjectId' 字段获取对应值

获取 MR ID:
  1. 使用 'lc pr list' 查询 MR 列表（Git 仓库目录下自动探测）
  2. 从输出中的 'iid' 字段获取对应值

示例:
  # 在 Git 仓库目录下执行（自动探测）
  lc pr comment 123 --body "这条PR代码写得很好"

  # 使用管道输入评论内容
  echo "需要修改第10行" | lc pr comment 123

  # 手动指定参数
  lc pr comment 123 --git-project-id 44142 --workspace-key XXJSxiaobaice --body "这条PR代码写得很好"

提示:
  使用 'lc doc mr-id' 查看如何获取 MR ID
  使用 'lc doc git-project-id' 查看如何获取 Git 项目 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
				return nil, fmt.Errorf("mr-id must be a number")
			}, common.ExecuteOptions{DebugMode: debugMode, Insecure: insecureSkipVerify, DryRun: dryRunMode, Logger: &logger})
			return
		}
		mrId = id
		commentMergeRequest()
	},
}

var prViewCmd = &cobra.Command{
	Use:   "view [mr-id]",
	Short: "查看合并请求详情",
	Long: `查看合并请求的详细信息，包括基本信息和评论。

参数说明:
  [mr-id]          合并请求 ID（MR ID），对应 'lc pr list' 输出中的 'iid' 字段

自动探测支持:
  在 Git 仓库目录下执行时，会自动探测 --workspace-key 和 --git-project-id
  如果自动探测失败，需要手动指定这些参数

获取 Git 项目 ID:
  1. 使用 'lc repo list --workspace-key <workspace-key>' 查询仓库列表
  2. 从输出中的 'gitProjectId' 字段获取对应值

获取 MR ID:
  1. 使用 'lc pr list' 查询 MR 列表（Git 仓库目录下自动探测）
  2. 从输出中的 'iid' 字段获取对应值

示例:
  # 在 Git 仓库目录下执行（自动探测）
  lc pr view 123
  lc pr view 123 --comments

  # 手动指定参数
  lc pr view 123 --git-project-id 44142 --workspace-key XXJSxiaobaice
  lc pr view 123 --git-project-id 44142 --workspace-key XXJSxiaobaice --comments

提示:
  使用 'lc doc mr-id' 查看如何获取 MR ID
  使用 'lc doc git-project-id' 查看如何获取 Git 项目 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
				return nil, fmt.Errorf("mr-id must be a number")
			}, common.ExecuteOptions{DebugMode: debugMode, Insecure: insecureSkipVerify, DryRun: dryRunMode, Logger: &logger})
			return
		}
		mrId = id
		viewMergeRequest()
	},
}

var prListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出合并请求",
	Long: `列出指定项目中的所有合并请求。

自动探测支持:
  在 Git 仓库目录下执行时，会自动探测 --workspace-key 和 --git-project-id
  如果自动探测失败，需要手动指定这些参数

获取 Git 项目 ID:
  1. 使用 'lc repo list --workspace-key <workspace-key>' 查询仓库列表
  2. 从输出中的 'gitProjectId' 字段获取对应值

示例:
  # 在 Git 仓库目录下执行（自动探测）
  lc pr list
  lc pr list --state opened

  # 手动指定参数
  lc pr list --git-project-id 44142 --workspace-key XXJSxiaobaice
  lc pr list --git-project-id 44142 --workspace-key XXJSxiaobaice --state opened
  lc pr list --git-project-id 44142 --workspace-key XXJSxiaobaice -p 1 -l 20

提示:
  使用 'lc doc git-project-id' 查看如何获取 Git 项目 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Run: func(cmd *cobra.Command, args []string) {
		listMergeRequests()
	},
}

var prPatchCommentCmd = &cobra.Command{
	Use:   "patch-comment [mr-id]",
	Short: "修改评论状态",
	Long: `修改合并请求中某条评论的状态，支持标记为已解决、无法修复、已关闭等状态。

参数说明:
  [mr-id]          合并请求 ID（MR ID），对应 'lc pr list' 输出中的 'iid' 字段
  --comment-id     评论 ID，对应 'lc pr view --comments' 输出中 'comments' 数组的 'id' 字段

自动探测支持:
  在 Git 仓库目录下执行时，会自动探测 --workspace-key 和 --git-project-id
  如果自动探测失败，需要手动指定这些参数

获取 Git 项目 ID:
  1. 使用 'lc repo list --workspace-key <workspace-key>' 查询仓库列表
  2. 从输出中的 'gitProjectId' 字段获取对应值

获取 MR ID:
  1. 使用 'lc pr list' 查询 MR 列表（Git 仓库目录下自动探测）
  2. 从输出中的 'iid' 字段获取对应值

获取评论 ID:
  1. 使用 'lc pr view <mr-id> --comments' 查看评论列表（Git 仓库目录下自动探测）
  2. 从输出中 'comments' 数组的 'id' 字段获取对应值

状态说明:
  active   - 活动中
  fixed    - 已解决
  wontFix  - 无法修复
  closed   - 已关闭
  pending  - 正在挂起

示例:
  # 在 Git 仓库目录下执行（自动探测）
  lc pr patch-comment 123 --comment-id 901 --state fixed
  lc pr patch-comment 123 --comment-id 902 --state wontFix

  # 手动指定参数
  lc pr patch-comment 123 --comment-id 901 --state fixed --git-project-id 44142 --workspace-key XXJSxiaobaice

提示:
  使用 'lc doc mr-id' 查看如何获取 MR ID
  使用 'lc doc comment-id' 查看如何获取评论 ID
  使用 'lc doc git-project-id' 查看如何获取 Git 项目 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
				return nil, fmt.Errorf("mr-id must be a number")
			}, common.ExecuteOptions{DebugMode: debugMode, Insecure: insecureSkipVerify, DryRun: dryRunMode, Logger: &logger})
			return
		}
		mrId = id
		patchCommentMergeRequest()
	},
}

func init() {
	rootCmd.AddCommand(prCmd)
	prCmd.AddCommand(prCreateCmd)
	prCmd.AddCommand(prReviewCmd)
	prCmd.AddCommand(prMergeCmd)
	prCmd.AddCommand(prCommentCmd)
	prCmd.AddCommand(prViewCmd)
	prCmd.AddCommand(prListCmd)
	prCmd.AddCommand(prPatchCommentCmd)

	// Create command flags
	prCreateCmd.Flags().StringVarP(&title, "title", "t", "", common.GetFlagDesc("title"))
	prCreateCmd.Flags().StringVarP(&body, "body", "b", "", common.GetFlagDesc("description"))
	prCreateCmd.Flags().StringVarP(&sourceBranch, "source", "s", "", common.GetFlagDesc("source-branch"))
	prCreateCmd.Flags().StringVarP(&targetBranch, "target", "", "master", common.GetFlagDesc("target-branch"))
	prCreateCmd.Flags().BoolVarP(&removeSource, "remove-source", "r", false, common.GetFlagDesc("remove-source"))
	prCreateCmd.Flags().IntVar(&projectId, "git-project-id", 0, common.GetFlagDesc("git-project-id")+"（可选，支持自动探测）")
	prCreateCmd.Flags().StringVarP(&prWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	prCreateCmd.MarkFlagRequired("title")

	// Review command flags
	prReviewCmd.Flags().IntVar(&projectId, "git-project-id", 0, common.GetFlagDesc("git-project-id")+"（可选，支持自动探测）")
	prReviewCmd.Flags().StringVar(&reviewType, "type", "approve", common.GetFlagDesc("review-type"))
	prReviewCmd.Flags().StringVar(&comment, "body", "", common.GetFlagDesc("body"))
	prReviewCmd.Flags().StringVarP(&prWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// Merge command flags
	prMergeCmd.Flags().IntVar(&projectId, "git-project-id", 0, common.GetFlagDesc("git-project-id")+"（可选，支持自动探测）")
	prMergeCmd.Flags().StringVar(&mergeType, "type", "merge", common.GetFlagDesc("merge-type"))
	prMergeCmd.Flags().BoolVar(&squash, "squash", false, "使用 squash 方式合并（与 gh pr merge --squash 兼容）")
	prMergeCmd.Flags().BoolVar(&rebase, "rebase", false, "使用 rebase 方式合并（与 gh pr merge --rebase 兼容）")
	prMergeCmd.Flags().BoolVar(&mergeFlag, "merge", false, "使用 merge 方式合并（与 gh pr merge --merge 兼容，默认方式）")
	prMergeCmd.Flags().BoolVarP(&removeSource, "remove-source", "r", false, common.GetFlagDesc("remove-source"))
	prMergeCmd.Flags().BoolVar(&deleteBranch, "delete-branch", false, "合并后删除源分支（与 gh pr merge --delete-branch 兼容）")
	prMergeCmd.Flags().StringVar(&comment, "body", "", common.GetFlagDesc("body"))
	prMergeCmd.Flags().StringVarP(&prWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// Comment command flags
	prCommentCmd.Flags().IntVar(&projectId, "git-project-id", 0, common.GetFlagDesc("git-project-id")+"（可选，支持自动探测）")
	prCommentCmd.Flags().StringVar(&comment, "body", "", common.GetFlagDesc("body")+"（必填，或通过管道输入）")
	prCommentCmd.Flags().StringVarP(&prWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// View command flags
	prViewCmd.Flags().IntVar(&projectId, "git-project-id", 0, common.GetFlagDesc("git-project-id")+"（可选，支持自动探测）")
	prViewCmd.Flags().BoolVar(&showComments, "comments", false, common.GetFlagDesc("comments"))
	prViewCmd.Flags().StringVarP(&prWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// List command flags
	prListCmd.Flags().IntVar(&projectId, "git-project-id", 0, common.GetFlagDesc("git-project-id")+"（可选，支持自动探测）")
	prListCmd.Flags().StringVar(&prListState, "state", "all", common.GetFlagDesc("mr-state"))
	prListCmd.Flags().IntVarP(&prListPage, "page", "p", 1, common.GetFlagDesc("page"))
	prListCmd.Flags().IntVarP(&prListSize, "limit", "l", 10, common.GetFlagDesc("limit"))
	prListCmd.Flags().StringVarP(&prWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// Patch comment command flags
	prPatchCommentCmd.Flags().IntVar(&commentId, "comment-id", 0, common.GetFlagDesc("comment-id"))
	prPatchCommentCmd.Flags().StringVar(&commentState, "state", "", common.GetFlagDesc("comment-state"))
	prPatchCommentCmd.Flags().IntVar(&projectId, "git-project-id", 0, common.GetFlagDesc("git-project-id")+"（可选，支持自动探测）")
	prPatchCommentCmd.MarkFlagRequired("comment-id")
	prPatchCommentCmd.MarkFlagRequired("state")
	prPatchCommentCmd.Flags().StringVarP(&prWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
}

func createPullRequest() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 尝试自动探测参数
		if err := tryAutoDetectForPR(); err != nil {
			return nil, fmt.Errorf("自动探测失败: %w\n\n请手动指定参数:\n  -w, --workspace-key     研发空间 Key\n  --git-project-id        Git 项目 ID", err)
		}

		// 如果未指定源分支，尝试自动获取当前 Git 分支
		if sourceBranch == "" {
			sourceBranch = getCurrentGitBranch()
			if sourceBranch != "" {
				ctx.Debug("Auto-detected source branch", zap.String("branch", sourceBranch))
			}
		}

		// 验证必需参数
		if prWorkspaceKey == "" {
			return nil, fmt.Errorf("缺少必需参数: --workspace-key\n请指定研发空间 Key，或使用自动探测")
		}
		if projectId == 0 {
			return nil, fmt.Errorf("缺少必需参数: --git-project-id\n请指定 Git 项目 ID，或使用自动探测")
		}

		if sourceBranch == "" {
			return nil, fmt.Errorf("缺少必需参数: --source\n请指定源分支，或在 Git 仓库目录下执行以自动获取当前分支")
		}

		headers := ctx.GetHeaders(prWorkspaceKey)
		mrService := api.NewMergeRequestService(ctx.Config.API.BaseRepoURL, headers, ctx.Client)

		// TODO: 根据当前所在路径的 git 配置自动提取仓库地址，反向查找到具体的仓库 ID
		// 1. 获取当前目录的 git remote URL
		// 2. 解析 URL 获取项目路径
		// 3. 调用项目列表 API 或搜索 API 反向查找 projectId

		requestData := &api.MergeRequestCreateRequest{
			SourceBranch:           sourceBranch,
			TargetBranch:           targetBranch,
			SourceProjectId:        nil,
			Title:                  title,
			Description:            body,
			RemoveSourceBranch:     removeSource,
			PrimaryReviewerNum:     0,
			PrimaryReviewerIds:     []int{},
			GeneralReviewerNum:     0,
			GeneralReviewerIds:     []int{},
			PrimaryReviewerUserIds: []int{},
			GeneralReviewerUserIds: []int{},
			WorkItems:              []int{},
			OriginalReviewerIds:    []int{},
			MergeType:              "",
			StateEvent:             "",
			PrAutoMergeEnabled:     false,
		}

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "create",
				"resource":  "merge-request",
				"summary":   fmt.Sprintf("将创建 MR: %s", title),
				"workspace": prWorkspaceKey,
				"projectId": projectId,
				"request":   requestData,
				"simulatedResponse": map[string]interface{}{
					"title":        title,
					"sourceBranch": sourceBranch,
					"targetBranch": targetBranch,
					"status":       "pending",
				},
			}, nil
		}

		resp, err := mrService.Create(projectId, requestData)
		if err != nil {
			return nil, err
		}

		if !resp.Success {
			return nil, fmt.Errorf("create failed: %s", resp.Message)
		}

		return resp.Data, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "pr create",
	})
}

func reviewMergeRequest() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 尝试自动探测参数
		if err := tryAutoDetectForPR(); err != nil {
			return nil, fmt.Errorf("自动探测失败: %w\n\n请手动指定参数:\n  -w, --workspace-key     研发空间 Key\n  --git-project-id        Git 项目 ID", err)
		}

		// 验证必需参数
		if prWorkspaceKey == "" {
			return nil, fmt.Errorf("缺少必需参数: --workspace-key\n请指定研发空间 Key，或使用自动探测")
		}
		if projectId == 0 {
			return nil, fmt.Errorf("缺少必需参数: --git-project-id\n请指定 Git 项目 ID，或使用自动探测")
		}

		headers := ctx.GetHeaders(prWorkspaceKey)
		mrService := api.NewMergeRequestService(ctx.Config.API.BaseRepoURL, headers, ctx.Client)

		requestData := &api.MergeRequestReviewRequest{
			ReviewType: reviewType,
			Comment:    comment,
		}

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    reviewType,
				"resource":  "merge-request",
				"summary":   fmt.Sprintf("将%s MR #%d", reviewType, mrId),
				"workspace": prWorkspaceKey,
				"projectId": projectId,
				"mrId":      mrId,
				"request":   requestData,
				"simulatedResponse": map[string]interface{}{
					"action": reviewType,
					"mrId":   mrId,
					"status": "pending",
				},
			}, nil
		}

		resp, err := mrService.Review(projectId, mrId, requestData)
		if err != nil {
			return nil, err
		}

		if !resp.Success {
			return nil, fmt.Errorf("review failed: %s", resp.Message)
		}

		return map[string]interface{}{
			"success": true,
			"mrId":    mrId,
			"action":  reviewType,
			"message": fmt.Sprintf("Successfully %s merge request #%d", reviewType, mrId),
			"data":    resp.Data,
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "pr review",
	})
}

func mergeMergeRequest() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 尝试自动探测参数
		if err := tryAutoDetectForPR(); err != nil {
			return nil, fmt.Errorf("自动探测失败: %w\n\n请手动指定参数:\n  -w, --workspace-key     研发空间 Key\n  --git-project-id        Git 项目 ID", err)
		}

		// 验证必需参数
		if prWorkspaceKey == "" {
			return nil, fmt.Errorf("缺少必需参数: --workspace-key\n请指定研发空间 Key，或使用自动探测")
		}
		if projectId == 0 {
			return nil, fmt.Errorf("缺少必需参数: --git-project-id\n请指定 Git 项目 ID，或使用自动探测")
		}

		headers := ctx.GetHeaders(prWorkspaceKey)
		mrService := api.NewMergeRequestService(ctx.Config.API.BaseRepoURL, headers, ctx.Client)

		// 处理 gh 风格参数与 lc 原有参数的转换和冲突检测
		// 规则：如果使用了 gh 风格参数（--squash, --rebase, --merge），将其转换为 mergeType
		// 如果同时使用了 --type 和 gh 风格参数，以明确指定的 --type 为准（如果 --type 不是默认值）

		// 统计用户明确指定的合并方式参数数量
		mergeFlagsCount := 0
		if squash {
			mergeFlagsCount++
		}
		if rebase {
			mergeFlagsCount++
		}
		if mergeFlag {
			mergeFlagsCount++
		}

		// 如果同时使用了多个 gh 风格合并参数，报错
		if mergeFlagsCount > 1 {
			return nil, fmt.Errorf("参数冲突：不能同时使用 --squash、--rebase 和 --merge 中的多个参数\n请只选择一种合并方式")
		}

		// 如果使用了 gh 风格参数，检查是否与 --type 冲突
		// 只有当 --type 被显式修改（不是默认值 "merge"）且同时使用了 gh 风格参数时才报错
		if mergeFlagsCount > 0 && mergeType != "merge" {
			// 获取实际使用的 gh 风格参数
			var ghFlag string
			if squash {
				ghFlag = "--squash"
			} else if rebase {
				ghFlag = "--rebase"
			} else if mergeFlag {
				ghFlag = "--merge"
			}
			return nil, fmt.Errorf("参数冲突：%s 与 --type %s 不能同时使用\n请只选择一种方式指定合并类型", ghFlag, mergeType)
		}

		// 将 gh 风格参数转换为 mergeType
		if squash {
			mergeType = "squash"
		} else if rebase {
			mergeType = "rebase"
		} else if mergeFlag {
			mergeType = "merge"
		}
		// 如果没有指定任何 gh 风格参数，保持原有的 mergeType 值（默认 "merge"）

		// 处理 --delete-branch 参数（与 --remove-source 效果相同）
		if deleteBranch {
			removeSource = true
		}

		requestData := &api.MergeRequestMergeRequest{
			ShouldRemoveSourceBranch: removeSource,
			MergeType:                mergeType,
			Squash:                   squash,
			MergeCommitMessage:       comment,
		}

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "merge",
				"resource":  "merge-request",
				"summary":   fmt.Sprintf("将合并 MR #%d (使用 %s 方式)", mrId, mergeType),
				"workspace": prWorkspaceKey,
				"projectId": projectId,
				"mrId":      mrId,
				"request":   requestData,
				"simulatedResponse": map[string]interface{}{
					"merged":       true,
					"mrId":         mrId,
					"mergeType":    mergeType,
					"removeSource": removeSource,
					"status":       "pending",
				},
			}, nil
		}

		resp, err := mrService.Merge(projectId, mrId, requestData)
		if err != nil {
			return nil, err
		}

		if !resp.Success {
			return nil, fmt.Errorf("merge failed: %s", resp.Message)
		}

		return map[string]interface{}{
			"success": true,
			"mrId":    mrId,
			"message": fmt.Sprintf("Successfully merged merge request #%d", mrId),
			"data":    resp.Data,
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "pr merge",
	})
}

func commentMergeRequest() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 尝试自动探测参数
		if err := tryAutoDetectForPR(); err != nil {
			return nil, fmt.Errorf("自动探测失败: %w\n\n请手动指定参数:\n  -w, --workspace-key     研发空间 Key\n  --git-project-id        Git 项目 ID", err)
		}

		// 验证必需参数
		if prWorkspaceKey == "" {
			return nil, fmt.Errorf("缺少必需参数: --workspace-key\n请指定研发空间 Key，或使用自动探测")
		}
		if projectId == 0 {
			return nil, fmt.Errorf("缺少必需参数: --git-project-id\n请指定 Git 项目 ID，或使用自动探测")
		}

		headers := ctx.GetHeaders(prWorkspaceKey)
		mrService := api.NewMergeRequestService(ctx.Config.API.BaseRepoURL, headers, ctx.Client)

		// 如果没有通过 -b 提供评论内容，尝试从管道读取
		commentBody := comment
		if commentBody == "" {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				// 有管道输入
				scanner := bufio.NewScanner(os.Stdin)
				var lines []string
				for scanner.Scan() {
					lines = append(lines, scanner.Text())
				}
				if err := scanner.Err(); err != nil {
					return nil, fmt.Errorf("error reading from stdin: %w", err)
				}
				commentBody = strings.Join(lines, "\n")
			}
		}

		if commentBody == "" {
			return nil, fmt.Errorf("评论内容不能为空，请使用 -b 参数提供，或通过管道输入")
		}

		requestData := &api.MergeRequestCommentRequest{
			Note:         commentBody,
			NoteableType: "PullRequestTheme",
		}

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "comment",
				"resource":  "merge-request",
				"summary":   fmt.Sprintf("将评论 MR #%d", mrId),
				"workspace": prWorkspaceKey,
				"projectId": projectId,
				"mrId":      mrId,
				"request": map[string]interface{}{
					"note": commentBody,
				},
			}, nil
		}

		resp, err := mrService.Comment(projectId, mrId, requestData)
		if err != nil {
			return nil, err
		}

		if !resp.Success {
			return nil, fmt.Errorf("comment failed: %s", resp.Message)
		}

		return map[string]interface{}{
			"success": true,
			"mrId":    mrId,
			"noteId":  resp.Data.Id,
			"note":    resp.Data.Note,
			"message": fmt.Sprintf("Successfully added comment to merge request #%d", mrId),
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "pr comment",
	})
}

func viewMergeRequest() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 尝试自动探测参数
		if err := tryAutoDetectForPR(); err != nil {
			return nil, fmt.Errorf("自动探测失败: %w\n\n请手动指定参数:\n  -w, --workspace-key     研发空间 Key\n  --git-project-id        Git 项目 ID", err)
		}

		// 验证必需参数
		if prWorkspaceKey == "" {
			return nil, fmt.Errorf("缺少必需参数: --workspace-key\n请指定研发空间 Key，或使用自动探测")
		}
		if projectId == 0 {
			return nil, fmt.Errorf("缺少必需参数: --git-project-id\n请指定 Git 项目 ID，或使用自动探测")
		}

		headers := ctx.GetHeaders(prWorkspaceKey)
		mrService := api.NewMergeRequestService(ctx.Config.API.BaseRepoURL, headers, ctx.Client)

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":       true,
				"action":       "view",
				"resource":     "merge-request",
				"summary":      fmt.Sprintf("将查看 MR #%d 的详情", mrId),
				"workspace":    prWorkspaceKey,
				"projectId":    projectId,
				"mrId":         mrId,
				"showComments": showComments,
			}, nil
		}

		// If --comments flag is set, only show comments
		if showComments {
			resp, err := mrService.GetComments(projectId, mrId)
			if err != nil {
				return nil, err
			}

			if !resp.Success {
				return nil, fmt.Errorf("get comments failed: %s", resp.Message)
			}

			// Format comments for output
			var comments []map[string]interface{}
			for _, c := range resp.Data {
				// Get the main note from Notes array
				if len(c.Notes) > 0 {
					note := c.Notes[0]
					commentMap := map[string]interface{}{
						"id":              note.Id,
						"note":            note.Note,
						"author":          note.Author.Name,
						"createdAt":       note.CreatedAt,
						"updatedAt":       note.UpdatedAt,
						"resolvedState":   c.ResolvedState,
						"resolvedEnabled": c.ResolvedEnabled,
					}
					comments = append(comments, commentMap)
				}
			}

			return map[string]interface{}{
				"success":  true,
				"mrId":     mrId,
				"count":    len(comments),
				"comments": comments,
			}, nil
		}

		// TODO: Get MR details (not implemented yet)
		return map[string]interface{}{
			"success": true,
			"mrId":    mrId,
			"message": "MR details view not implemented yet. Use --comments to view comments.",
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "pr view",
	})
}

func listMergeRequests() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 尝试自动探测参数
		if err := tryAutoDetectForPR(); err != nil {
			return nil, fmt.Errorf("自动探测失败: %w\n\n请手动指定参数:\n  -w, --workspace-key     研发空间 Key\n  --git-project-id        Git 项目 ID", err)
		}

		// 验证必需参数
		if prWorkspaceKey == "" {
			return nil, fmt.Errorf("缺少必需参数: --workspace-key\n请指定研发空间 Key，或使用自动探测")
		}
		if projectId == 0 {
			return nil, fmt.Errorf("缺少必需参数: --git-project-id\n请指定 Git 项目 ID，或使用自动探测")
		}

		headers := ctx.GetHeaders(prWorkspaceKey)
		mrService := api.NewMergeRequestService(ctx.Config.API.BaseRepoURL, headers, ctx.Client)

		requestData := &api.MergeRequestListRequest{
			PageNo:   prListPage,
			PageSize: prListSize,
			State:    prListState,
		}

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "list",
				"resource":  "merge-requests",
				"summary":   fmt.Sprintf("将列出项目 %d 的 MR", projectId),
				"workspace": prWorkspaceKey,
				"projectId": projectId,
				"state":     prListState,
				"page":      prListPage,
				"size":      prListSize,
			}, nil
		}

		resp, err := mrService.List(projectId, requestData)
		if err != nil {
			return nil, err
		}

		if !resp.Success {
			return nil, fmt.Errorf("list failed: %s", resp.Message)
		}

		// Format MR list for output
		var items []map[string]interface{}
		for _, mr := range resp.Data {
			item := map[string]interface{}{
				"iid":            mr.Iid,
				"title":          mr.Title,
				"state":          mr.State,
				"sourceBranch":   mr.SourceBranch,
				"targetBranch":   mr.TargetBranch,
				"mergeStatus":    mr.MergeStatus,
				"userNotesCount": mr.UserNotesCount,
				"author":         mr.Author.Name,
				"createdAt":      mr.CreatedAt,
				"updatedAt":      mr.UpdatedAt,
			}
			items = append(items, item)
		}

		return map[string]interface{}{
			"success":   true,
			"count":     resp.Count,
			"pageNo":    resp.PageNo,
			"pageSize":  resp.PageSize,
			"pageCount": resp.PageCount,
			"items":     items,
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "pr list",
	})
}

// tryAutoDetectForPR 尝试为PR命令自动探测参数
func tryAutoDetectForPR() error {
	autoResult := common.TryAutoDetect(true)
	if autoResult.Success {
		if prWorkspaceKey == "" {
			prWorkspaceKey = autoResult.Context.WorkspaceKey
		}
		if projectId == 0 && autoResult.Context.GitProjectID != "" {
			// 尝试将GitProjectID转换为int
			if gitId, err := strconv.Atoi(autoResult.Context.GitProjectID); err == nil {
				projectId = gitId
			}
		}
		return nil
	}
	return autoResult.Error
}

// getCurrentGitBranch 获取当前 Git 分支名称
// 如果不是在 Git 仓库目录下，返回空字符串
func getCurrentGitBranch() string {
	// 使用兼容的方式获取当前分支（支持旧版 git）
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func patchCommentMergeRequest() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 尝试自动探测参数
		if err := tryAutoDetectForPR(); err != nil {
			return nil, fmt.Errorf("自动探测失败: %w\n\n请手动指定参数:\n  -w, --workspace-key     研发空间 Key\n  --git-project-id        Git 项目 ID", err)
		}

		// 验证必需参数
		if prWorkspaceKey == "" {
			return nil, fmt.Errorf("缺少必需参数: --workspace-key\n请指定研发空间 Key，或使用自动探测")
		}
		if projectId == 0 {
			return nil, fmt.Errorf("缺少必需参数: --git-project-id\n请指定 Git 项目 ID，或使用自动探测")
		}

		headers := ctx.GetHeaders(prWorkspaceKey)
		mrService := api.NewMergeRequestService(ctx.Config.API.BaseRepoURL, headers, ctx.Client)

		// Validate state value
		validStates := map[string]string{
			"active":  "活动中",
			"fixed":   "已解决",
			"wontFix": "无法修复",
			"closed":  "已关闭",
			"pending": "正在挂起",
		}
		stateDesc, valid := validStates[commentState]
		if !valid {
			return nil, fmt.Errorf("invalid state '%s', must be one of: active, fixed, wontFix, closed, pending", commentState)
		}

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "patch-comment",
				"resource":  "merge-request-comment",
				"summary":   fmt.Sprintf("将修改 MR #%d 的评论 #%d 状态为 %s(%s)", mrId, commentId, commentState, stateDesc),
				"workspace": prWorkspaceKey,
				"projectId": projectId,
				"mrId":      mrId,
				"commentId": commentId,
				"state":     commentState,
			}, nil
		}

		resp, err := mrService.UpdateCommentState(projectId, mrId, commentId, commentState)
		if err != nil {
			return nil, err
		}

		if !resp.Success {
			return nil, fmt.Errorf("update comment state failed: %s", resp.Message)
		}

		return map[string]interface{}{
			"success":   true,
			"mrId":      mrId,
			"commentId": commentId,
			"state":     commentState,
			"stateDesc": stateDesc,
			"message":   fmt.Sprintf("Successfully updated comment #%d state to %s", commentId, commentState),
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "pr patch-comment",
	})
}
