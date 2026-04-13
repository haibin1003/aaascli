package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
	"github.com/user/lc/internal/config"
)

var reqName string
var reqFile string
var reqWorkspaceKey string
var reqWorkspaceName string
var reqProjectCode string

// RequirementYAML represents the YAML structure for creating a requirement
type RequirementYAML struct {
	Name               string   `yaml:"name"`
	Workspace          string   `yaml:"workspace,omitempty"`
	ItemType           string   `yaml:"itemType,omitempty"`
	Proposer           UserYAML `yaml:"proposer,omitempty"`
	Assignee           UserYAML `yaml:"assignee,omitempty"`
	Priority           string   `yaml:"priority,omitempty"`
	Source             []string `yaml:"source,omitempty"`
	AffiliatedUnit     string   `yaml:"affiliatedUnit,omitempty"`
	ContactNumber      string   `yaml:"contactNumber,omitempty"`
	ContactEmail       string   `yaml:"contactEmail,omitempty"`
	BusinessBackground string   `yaml:"businessBackground,omitempty"`
	Requirement        string   `yaml:"requirement,omitempty"`
	AcceptanceCriteria string   `yaml:"acceptanceCriteria,omitempty"`
	RequirementType    []string `yaml:"requirementType,omitempty"`
	ProjectCode        string   `yaml:"projectCode,omitempty"`
}

// UserYAML represents user information in YAML
type UserYAML struct {
	Label    string `yaml:"label"`
	Value    string `yaml:"value"`
	Username string `yaml:"username"`
	Nickname string `yaml:"nickname"`
	Email    string `yaml:"email,omitempty"`
}

var reqCmd = &cobra.Command{
	Use:   "req",
	Short: "管理需求",
	Long: `管理需求，包括创建、查询、查看详情、删除需求等功能。

自动探测支持:
  list, search 命令支持自动探测研发空间。
  在 Git 仓库目录下执行时，无需手动指定 --workspace-key。`,
}

var reqCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "创建新需求",
	Long: `创建一个新需求。

参数获取:
  --project-code 获取方式:
    lc space projects --workspace-key XXJSLJCLIDEV
    # 从输出中的 items[].projectCode 字段获取

示例:
  # 使用名称简单创建
  lc req create "需求名称" --workspace-key XXJSxiaobaice --project-code R24113J3C04

  # 从 YAML 文件创建（类似 kubectl apply -f）
  lc req create -f requirement.yaml --workspace-key XXJSxiaobaice

  # 从标准输入创建（管道输入）
  cat requirement.yaml | lc req create --workspace-key XXJSxiaobaice --project-code R24113J3C04

YAML 格式示例:
  name: 需求名称
  projectCode: R24113J3C04
  proposer:
    label: "魏宝辉(weibaohui@hq.cmcc)"
    value: weibaohui@hq.cmcc
    username: weibaohui@hq.cmcc
    nickname: 魏宝辉
  assignee:
    label: "魏宝辉(weibaohui@hq.cmcc)"
    value: weibaohui@hq.cmcc
    username: weibaohui@hq.cmcc
    nickname: 魏宝辉
  businessBackground: |
    这是业务背景内容
    支持多行文本
  requirement: |
    这是需求描述内容
  acceptanceCriteria: |
    这是验收标准内容
  requirementType:
    - 开发域`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			reqName = args[0]
		}
		createRequirement()
	},
}

var reqListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询需求列表",
	Long: `查询研发空间中的需求列表。

自动探测:
  如果在 Git 仓库目录下执行，且未指定 --workspace-key，
  命令会自动探测当前目录所属的研发空间。

示例:
  # 自动探测并列出需求
  lc req list

  # 手动指定研发空间
  lc req list -w XXJSLJCLIDEV`,
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForReqList(cmd)
		}, "-w, --workspace-key")
		listRequirements()
	},
}

var reqViewCmd = &cobra.Command{
	Use:   "view [object-id] [flags]",
	Short: "查看需求详情",
	Long: `根据 objectId 查看需求的详细信息。

获取 object-id:
  1. 使用 'lc req list --workspace-key <key>' 查询需求列表
  2. 从输出中的 'objectId' 或 'key' 字段获取对应值

示例:
  lc req view AXi9LpGjsA --workspace-key XXJSxiaobaice

提示:
  使用 'lc doc object-id' 查看如何获取需求对象 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForReqView(cmd)
		}, "-w, --workspace-key")
		viewRequirement(args[0])
	},
}

var reqDeleteCmd = &cobra.Command{
	Use:   "delete [object-id] [flags]",
	Short: "删除需求",
	Long: `删除单个需求。

注意: 必须使用需求的 objectId 进行删除，而非 key。

获取 object-id:
  1. 使用 'lc req list --workspace-key <key>' 查询需求列表
  2. 从输出中的 'objectId' 或 'key' 字段获取对应值

示例:
  # 使用 objectId 删除
  lc req delete AXi9LpGjsA --workspace-key XXJSxiaobaice

提示:
  使用 'lc doc object-id' 查看如何获取需求对象 ID
  使用 'lc doc workspace-key' 查看如何获取研发空间 key`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForReqDelete(cmd)
		}, "-w, --workspace-key")
		deleteRequirement(args[0])
	},
}

var reqSearchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "搜索需求",
	Long: `根据关键词搜索需求。

支持在需求标题中进行模糊搜索。

示例:
  # 搜索包含 "server" 的需求
  lc req search server --workspace-key XXJSxiaobaice

  # 搜索并指定分页
  lc req search mock --workspace-key XXJSxiaobaice --limit 10 --offset 0`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForReqSearch(cmd)
		}, "-w, --workspace-key")
		searchRequirements(args[0])
	},
}

var reqUpdateCmd = &cobra.Command{
	Use:   "update [object-id] [flags]",
	Short: "更新需求",
	Long: `更新需求的指定字段。

可以单独更新一个字段，也可以同时更新多个字段。

支持的字段:
  --name                    需求名称
  --requirement            需求描述（富文本）
  --acceptance-criteria    验收标准（富文本）
  --business-background    业务背景（富文本）
  --planned-end-time       计划完成时间（时间戳，毫秒）
  --planned-start-time     计划开始时间（时间戳，毫秒）
  --priority               优先级ID

获取 object-id:
  1. 使用 'lc req list --workspace-key <key>' 查询需求列表
  2. 从输出中的 'objectId' 或 'key' 字段获取对应值

示例:
  # 更新需求名称
  lc req update 7PBCUE0QPx --workspace-key XXJSxiaobaice --name "新的需求名称"

  # 更新需求描述
  lc req update 7PBCUE0QPx --workspace-key XXJSxiaobaice --requirement "这是新的需求描述"

  # 同时更新多个字段
  lc req update 7PBCUE0QPx --workspace-key XXJSxiaobaice \
    --name "新名称" \
    --requirement "新描述" \
    --acceptance-criteria "新验收标准"

  # 更新计划完成时间（时间戳，毫秒）
  lc req update 7PBCUE0QPx --workspace-key XXJSxiaobaice --planned-end-time 1773763200000`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.HandleAutoDetectWithExit(func() error {
			return tryAutoDetectForReqUpdate(cmd)
		}, "-w, --workspace-key")
		updateRequirement(args[0])
	},
}

var listLimit int
var listOffset int

var (
	updateName               string
	updateRequirementDesc    string
	updateAcceptanceCriteria string
	updateBusinessBackground string
	updatePlannedEndTime     int64
	updatePlannedStartTime   int64
	updatePriority           string
)

func init() {
	rootCmd.AddCommand(reqCmd)
	reqCmd.AddCommand(reqCreateCmd)
	reqCmd.AddCommand(reqListCmd)
	reqCmd.AddCommand(reqViewCmd)
	reqCmd.AddCommand(reqDeleteCmd)
	reqCmd.AddCommand(reqUpdateCmd)
	reqCmd.AddCommand(reqSearchCmd)

	reqCreateCmd.Flags().StringVarP(&reqFile, "filename", "f", "", common.GetFlagDesc("filename"))
	reqCreateCmd.Flags().StringVarP(&reqWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
	reqCreateCmd.Flags().StringVar(&reqProjectCode, "project-code", "", common.GetFlagDesc("project-code")+"（可选，获取方式: lc space project linked -w <spaceCode>）")

	reqListCmd.Flags().IntVarP(&listLimit, "limit", "l", 20, common.GetFlagDesc("limit"))
	reqListCmd.Flags().IntVarP(&listOffset, "offset", "o", 0, common.GetFlagDesc("offset"))
	reqListCmd.Flags().StringVarP(&reqWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	reqViewCmd.Flags().StringVarP(&reqWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	reqDeleteCmd.Flags().StringVarP(&reqWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")

	// Update command flags
	reqUpdateCmd.Flags().StringVarP(&reqWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
	reqUpdateCmd.Flags().StringVar(&updateName, "name", "", common.GetFlagDesc("name"))
	reqUpdateCmd.Flags().StringVar(&updateRequirementDesc, "requirement", "", common.GetFlagDesc("requirement"))
	reqUpdateCmd.Flags().StringVar(&updateAcceptanceCriteria, "acceptance-criteria", "", common.GetFlagDesc("acceptance-criteria"))
	reqUpdateCmd.Flags().StringVar(&updateBusinessBackground, "business-background", "", common.GetFlagDesc("business-background"))
	reqUpdateCmd.Flags().Int64Var(&updatePlannedEndTime, "planned-end-time", 0, common.GetFlagDesc("planned-end-time"))
	reqUpdateCmd.Flags().Int64Var(&updatePlannedStartTime, "planned-start-time", 0, common.GetFlagDesc("planned-start-time"))
	reqUpdateCmd.Flags().StringVar(&updatePriority, "priority", "", common.GetFlagDesc("priority")+"ID")
	// Search command flags
	reqSearchCmd.Flags().StringVarP(&reqWorkspaceKey, "workspace-key", "w", "", common.GetFlagDesc("workspace-key")+"（可选，支持自动探测）")
	reqSearchCmd.Flags().IntVarP(&listLimit, "limit", "l", 20, common.GetFlagDesc("limit"))
	reqSearchCmd.Flags().IntVarP(&listOffset, "offset", "o", 0, common.GetFlagDesc("offset"))
}

func createRequirement() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		var reqYAML *RequirementYAML

		// Try to read from file, stdin, or pipe
		data, err := common.ReadYAMLFromInput(reqFile)
		if err == nil {
			// Parse YAML from input
			reqYAML = &RequirementYAML{}
			if err := common.ParseYAML(data, reqYAML); err != nil {
				return nil, err
			}
			if reqYAML.Name == "" {
				return nil, fmt.Errorf("'name' field is required in YAML")
			}
		} else if reqName != "" {
			// Use simple name argument
			reqYAML = &RequirementYAML{Name: reqName}
		} else {
			return nil, fmt.Errorf("either provide a name argument, use -f flag, or pipe YAML content")
		}

		// Handle dry-run mode early (before API calls)
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "create",
				"resource":  "requirement",
				"summary":   fmt.Sprintf("将创建需求: %s", reqYAML.Name),
				"workspace": reqWorkspaceKey,
				"request": map[string]interface{}{
					"name":      reqYAML.Name,
					"itemType":  "初始需求",
					"workspace": reqWorkspaceKey,
				},
				"simulatedResponse": map[string]interface{}{
					"name":   reqYAML.Name,
					"key":    "DRY-RUN-KEY",
					"status": "pending",
				},
			}, nil
		}

		headers := ctx.GetHeaders(reqWorkspaceKey)
		reqService := api.NewRequirementService(ctx.Config.API.BaseReqURL, headers, ctx.Client, ctx.Config)

		// 获取 workspace objectId
		spaceService := api.NewSpaceService(ctx.Config.API.BasePlatformURL, headers, ctx.Client)
		workspaceObjectId, err := spaceService.GetWorkspaceObjectId(reqWorkspaceKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get workspace objectId: %w", err)
		}

		// 构建请求数据 - 使用查询到的 workspace ObjectID 和原始的 spaceCode
		requestData := buildRequirementCreateRequest(reqYAML, ctx.Config, workspaceObjectId, reqWorkspaceKey)

		resp, err := reqService.Create(requestData, reqWorkspaceKey)
		if err != nil {
			return nil, err
		}

		// 输出创建结果
		result := map[string]interface{}{
			"success":   true,
			"name":      resp.Name,
			"objectId":  resp.ObjectID,
			"key":       resp.Key,
			"createdAt": resp.CreatedAt,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "req create",
	})
}

// buildRequirementCreateRequest 从 YAML 构建创建请求
// workspaceObjectId: Parse Workspace 类的 objectId (如 QF5pqy0yqV)
// spaceCode: 空间的 key/code (如 XXJSxiaobaice)
func buildRequirementCreateRequest(yaml *RequirementYAML, cfg *config.Config, workspaceObjectId, spaceCode string) *api.RequirementCreateRequest {
	// 默认值
	now := time.Now().UnixMilli()
	sevenDaysLater := now + 7*24*60*60*1000
	uniqueID := generateUniqueID()

	// 如果 YAML 中指定了 workspace，则使用 YAML 中的值（优先使用 objectId）
	if yaml.Workspace != "" {
		workspaceObjectId = yaml.Workspace
	}

	// 事项类型从YAML读取，默认为初始需求类型
	itemTypeID := yaml.ItemType
	if itemTypeID == "" {
		itemTypeID = "G5WtJ3fTeo" // 初始需求类型
	}

	// 构建请求
	req := &api.RequirementCreateRequest{
		Name:      yaml.Name,
		Ancestors: []string{},
		Workspace: api.Workspace{
			Type:      "Pointer",
			ClassName: "Workspace",
			ObjectID:  workspaceObjectId,
		},
		ItemType: api.ItemType{
			Type:      "Pointer",
			ClassName: "ItemType",
			ObjectID:  itemTypeID,
		},
		Reporter:    nil,
		ItemContext: api.ItemContext{},
		ParseContext: api.ParseContext{
			EventExtraData: api.EventExtraData{
				FilterItemTypeList: []api.FilterItemType{
					{
						CreatedAt: "2025-09-07T03:13:31.832Z",
						UpdatedAt: "2025-10-30T09:57:13.179Z",
						Key:       "InitialDemand",
						Icon:      "/cmdevops-req/parse/files/ENTP750043923870622608/6b%2Fd5abd4b07f9b9551c694bead1e840272_ItemIR.svg",
						Name:      "初始需求",
						UpdatedBy: api.Workspace{Type: "Pointer", ClassName: "_User", ObjectID: "xieyunhan@hq.cmcc"},
						ObjectID:  "G5WtJ3fTeo",
					},
				},
			},
		},
	}

	// 构建 Values
	// 注意：IPD 模式空间中，belongingSpace 和 proposeSpace 必须是 workspace 的 objectId，不是 spaceCode
	values := api.RequirementValues{
		BelongingSpace:         workspaceObjectId,
		ProposeSpace:           workspaceObjectId,
		RequestSubmissionTime:  int64(now),
		YearDemand:             int64(now),
		ExpectedCompletionTime: int64(sevenDaysLater),
		PlannedStartTime:       int64(sevenDaysLater),
		PlannedEndTime:         int64(sevenDaysLater),
		ProjectNo:              []interface{}{},
		DevelopmentCompletion:  nil,
		Relations:              nil,
		ScreenType:             "create",
	}

	// 设置默认值或 YAML 提供的值
	if len(yaml.Source) >= 2 {
		values.Source = [][]string{{yaml.Source[0], yaml.Source[1]}}
	} else {
		values.Source = [][]string{{"eb553b42-1507-4215-ae67-5476f47d8bf7", "64144d1d-dca3-4cb8-8c2a-983b8d9edb63"}}
	}

	// 这三个字段不传值时不设置默认值
	if yaml.AffiliatedUnit != "" {
		values.AffiliatedUnit = yaml.AffiliatedUnit
	}

	if yaml.ContactNumber != "" {
		values.ContactNumber = yaml.ContactNumber
	}

	if yaml.ContactEmail != "" {
		values.ContactEmail = yaml.ContactEmail
	}

	if len(yaml.RequirementType) > 0 {
		values.RequirementType = yaml.RequirementType
	} else {
		values.RequirementType = []string{"开发域"}
	}

	// 关联项目编号（projectNo）- 优先使用 YAML 中的 projectCode，其次使用命令行参数
	projectCode := reqProjectCode
	if yaml.ProjectCode != "" {
		projectCode = yaml.ProjectCode
	}
	if projectCode != "" {
		values.ProjectNo = []interface{}{projectCode}
	}

	// 优先级从配置或YAML读取
	if yaml.Priority != "" {
		values.Priority = yaml.Priority
	} else {
		values.Priority = cfg.Defaults.PriorityID
	}

	// 处理用户字段
	if yaml.Proposer.Username != "" {
		values.Proposer = []api.UserValue{{
			Label:    yaml.Proposer.Label,
			Value:    yaml.Proposer.Value,
			Username: yaml.Proposer.Username,
			Nickname: yaml.Proposer.Nickname,
			Email:    yaml.Proposer.Email,
			Deleted:  false,
			Enabled:  true,
		}}
	} else {
		user := cfg.GetUser()
		values.Proposer = []api.UserValue{{
			Label:    user.Label(),
			Value:    user.Username,
			Username: user.Username,
			Nickname: user.Nickname,
			Email:    user.Email,
			Deleted:  false,
			Enabled:  true,
		}}
	}

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
		values.Assignee = values.Proposer
	}

	// 处理富文本字段
	if yaml.BusinessBackground != "" {
		values.BusinessBackground = textToEditorContent(yaml.BusinessBackground, "bg"+uniqueID)
	} else {
		values.BusinessBackground = textToEditorContent(yaml.Name, "bg"+uniqueID)
	}

	if yaml.Requirement != "" {
		values.Requirement = textToEditorContent(yaml.Requirement, "req"+uniqueID)
	} else {
		values.Requirement = textToEditorContent(yaml.Name, "req"+uniqueID)
	}

	if yaml.AcceptanceCriteria != "" {
		values.AcceptanceCriteria = textToEditorContent(yaml.AcceptanceCriteria, "ac"+uniqueID)
	} else {
		values.AcceptanceCriteria = textToEditorContent(yaml.Name, "ac"+uniqueID)
	}

	req.Values = values
	return req
}

// textToEditorContent 将纯文本转换为编辑器内容格式
func textToEditorContent(text, id string) []api.EditorContent {
	lines := strings.Split(text, "\n")
	var contents []api.EditorContent
	for i, line := range lines {
		contents = append(contents, api.EditorContent{
			Children: []struct {
				Text string `json:"text"`
			}{{Text: line}},
			Type: "p",
			ID:   fmt.Sprintf("%s_%d", id, i),
		})
	}
	return contents
}

// generateUniqueID 生成简单的唯一ID
func generateUniqueID() string {
	return fmt.Sprintf("%d", os.Getpid())
}

// req 命令的自动探测字段配置
var (
	// req 命令的自动探测字段配置：仅 workspace-key
	reqAutoDetectBase = []common.AutoDetectField{
		{FlagName: "workspace-key", TargetVar: &reqWorkspaceKey, ContextKey: "WorkspaceKey"},
	}
)

// tryAutoDetectForReqList 尝试为 req list 命令自动探测参数
func tryAutoDetectForReqList(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, reqAutoDetectBase)
	return err
}

// tryAutoDetectForReqCreate 尝试为 req create 命令自动探测参数
func tryAutoDetectForReqCreate(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, reqAutoDetectBase)
	return err
}

// tryAutoDetectForReqView 尝试为 req view 命令自动探测参数
func tryAutoDetectForReqView(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, reqAutoDetectBase)
	return err
}

// tryAutoDetectForReqDelete 尝试为 req delete 命令自动探测参数
func tryAutoDetectForReqDelete(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, reqAutoDetectBase)
	return err
}

// tryAutoDetectForReqUpdate 尝试为 req update 命令自动探测参数
func tryAutoDetectForReqUpdate(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, reqAutoDetectBase)
	return err
}

// tryAutoDetectForReqSearch 尝试为 req search 命令自动探测参数
func tryAutoDetectForReqSearch(cmd *cobra.Command) error {
	_, err := common.ApplyAutoDetect(cmd, reqAutoDetectBase)
	return err
}

func listRequirements() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 如果只传了 workspace-key 没传 workspace-name，自动获取
		if reqWorkspaceKey != "" && reqWorkspaceName == "" {
			spaceService := api.NewSpaceService(ctx.Config.API.BasePlatformURL, ctx.GetHeaders(reqWorkspaceKey), ctx.Client)
			if name, err := spaceService.GetSpaceNameByCode(reqWorkspaceKey); err == nil {
				reqWorkspaceName = name
			}
		}

		headers := ctx.GetHeaders(reqWorkspaceKey)
		reqService := api.NewRequirementService(ctx.Config.API.BaseReqURL, headers, ctx.Client, ctx.Config)

		// 构建请求数据 - 使用命令行传入的 workspace 名称和 key
		// 包含初始需求、用户故事、任务等常见需求类型
		requestData := &api.RequirementListRequest{
			IQL:               fmt.Sprintf("((所属空间 = '%s') and ('belongingSpace' in [\"currentWorkspace()\"] or ('类型' in [\"初始需求\",\"用户故事\",\"任务\"] ))) order by 创建时间 desc", reqWorkspaceName),
			Size:              listLimit,
			From:              listOffset,
			IsExpand:          false,
			IsShowAncestors:   true,
			IsShowDescendants: true,
			IsShowLinkItems:   false,
			Extend:            map[string]interface{}{},
			Fields: []string{
				"ancestors", "assignee", "createdAt", "createdBy",
				"earlyWarning", "expectedCompletionTime", "id",
				"itemType", "key", "objectId", "priority", "projectNo",
				"rowId", "status", "workspace",
			},
			RefererInfo: api.RefererInfo{
				WorkspaceKey: reqWorkspaceKey,
			},
		}

		resp, err := reqService.List(requestData, reqWorkspaceKey)
		if err != nil {
			return nil, err
		}

		// Output the requirements list as JSON
		var items []map[string]interface{}
		for _, item := range resp.Payload.Items {
			statusName := "未知"
			if item.Status.Name != "" {
				statusName = item.Status.Name
			}

			assigneeName := "未分配"
			if len(item.Values.Assignee) > 0 {
				assigneeName = item.Values.Assignee[0].Nickname
			}

			items = append(items, map[string]interface{}{
				"name":      item.Name,
				"key":       item.Key,
				"objectId":  item.ObjectID,
				"status":    statusName,
				"assignee":  assigneeName,
				"createdAt": item.CreatedAt,
			})
		}

		result := map[string]interface{}{
			"count":      resp.Payload.Count,
			"totalCount": resp.Payload.TotalCount,
			"items":      items,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "req list",
	})
}

func viewRequirement(objectID string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(reqWorkspaceKey)
		reqService := api.NewRequirementService(ctx.Config.API.BaseReqURL, headers, ctx.Client, ctx.Config)

		resp, err := reqService.View(objectID, reqWorkspaceKey)
		if err != nil {
			return nil, err
		}

		// Output the requirement details as JSON
		item := resp.Item
		result := map[string]interface{}{
			"name":               item.Name,
			"key":                item.Key,
			"objectId":           item.ObjectID,
			"status":             item.Status.Name,
			"type":               item.ItemType.Name,
			"workspace":          item.Workspace.Name,
			"createdAt":          item.CreatedAt,
			"updatedAt":          item.UpdatedAt,
			"createdBy":          item.CreatedBy.Nickname + " (" + item.CreatedBy.Username + ")",
			"version":            item.Version,
			"businessBackground": common.ExtractTextFromRichText(item.Values["businessBackground"]),
			"requirement":        common.ExtractTextFromRichText(item.Values["requirement"]),
			"acceptanceCriteria": common.ExtractTextFromRichText(item.Values["acceptanceCriteria"]),
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "req view",
	})
}

// deleteRequirement deletes a requirement by its object ID
// 注意: 必须使用 objectId，而非 key
func deleteRequirement(objectID string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		headers := ctx.GetHeaders(reqWorkspaceKey)
		reqService := api.NewRequirementService(ctx.Config.API.BaseReqURL, headers, ctx.Client, ctx.Config)

		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":   true,
				"action":   "delete",
				"resource": "requirement",
				"summary":  fmt.Sprintf("将删除需求(objectId): %s", objectID),
				"request": map[string]interface{}{
					"objectId": objectID,
				},
				"simulatedResponse": map[string]interface{}{
					"deleted":  1,
					"objectId": objectID,
					"status":   "pending",
				},
			}, nil
		}

		resp, err := reqService.DeleteRequirements([]string{objectID}, reqWorkspaceKey)
		if err != nil {
			return nil, fmt.Errorf("删除需求失败: %w", err)
		}

		result := map[string]interface{}{
			"success":  resp.Code == 0,
			"code":     resp.Code,
			"deleted":  1,
			"objectId": objectID,
			"message":  resp.Message,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "req delete",
	})
}

// updateRequirement updates a requirement by its object ID
func updateRequirement(objectID string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// Check if at least one field is specified
		if updateName == "" && updateRequirementDesc == "" && updateAcceptanceCriteria == "" &&
			updateBusinessBackground == "" && updatePlannedEndTime == 0 && updatePlannedStartTime == 0 &&
			updatePriority == "" {
			return nil, fmt.Errorf("至少需要一个更新字段，请使用 --name, --requirement, --acceptance-criteria, --business-background, --planned-end-time, --planned-start-time 或 --priority 指定")
		}

		// Handle dry-run mode
		if ctx.DryRun {
			updates := map[string]interface{}{}
			if updateName != "" {
				updates["name"] = updateName
			}
			if updateRequirementDesc != "" {
				updates["requirement"] = updateRequirementDesc
			}
			if updateAcceptanceCriteria != "" {
				updates["acceptanceCriteria"] = updateAcceptanceCriteria
			}
			if updateBusinessBackground != "" {
				updates["businessBackground"] = updateBusinessBackground
			}
			if updatePlannedEndTime != 0 {
				updates["plannedEndTime"] = updatePlannedEndTime
			}
			if updatePlannedStartTime != 0 {
				updates["plannedStartTime"] = updatePlannedStartTime
			}
			if updatePriority != "" {
				updates["priority"] = updatePriority
			}
			return map[string]interface{}{
				"dryRun":   true,
				"action":   "update",
				"resource": "requirement",
				"summary":  fmt.Sprintf("将更新需求(objectId): %s", objectID),
				"request": map[string]interface{}{
					"objectId": objectID,
					"updates":  updates,
				},
				"simulatedResponse": map[string]interface{}{
					"objectId": objectID,
					"status":   "pending",
				},
			}, nil
		}

		headers := ctx.GetHeaders(reqWorkspaceKey)
		reqService := api.NewRequirementService(ctx.Config.API.BaseReqURL, headers, ctx.Client, ctx.Config)

		// Build update request
		updateReq := buildRequirementUpdateRequest()

		resp, err := reqService.Update(objectID, updateReq, reqWorkspaceKey)
		if err != nil {
			return nil, fmt.Errorf("更新需求失败: %w", err)
		}

		// Get name from values.name if available (API stores updated name there)
		name := resp.Item.Name
		if resp.Item.Values.Name != "" {
			name = resp.Item.Values.Name
		}

		result := map[string]interface{}{
			"success":   true,
			"objectId":  resp.Item.ObjectID,
			"key":       resp.Item.Key,
			"name":      name,
			"updatedAt": resp.Item.UpdatedAt,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		DryRun:      dryRunMode,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "req update",
	})
}

// buildRequirementUpdateRequest builds the update request from flags
// Supports partial updates - only sends fields that need to be changed
func buildRequirementUpdateRequest() *api.RequirementUpdateRequest {
	req := &api.RequirementUpdateRequest{}

	values := api.UpdateValues{}

	// Add fields that are specified
	if updateName != "" {
		values.Name = updateName
	}
	if updateRequirementDesc != "" {
		values.Requirement = textToEditorContent(updateRequirementDesc, generateUniqueID())
	}
	if updateAcceptanceCriteria != "" {
		values.AcceptanceCriteria = textToEditorContent(updateAcceptanceCriteria, generateUniqueID())
	}
	if updateBusinessBackground != "" {
		values.BusinessBackground = textToEditorContent(updateBusinessBackground, generateUniqueID())
	}
	if updatePlannedEndTime != 0 {
		values.PlannedEndTime = &updatePlannedEndTime
	}
	if updatePlannedStartTime != 0 {
		values.PlannedStartTime = &updatePlannedStartTime
	}
	if updatePriority != "" {
		values.Priority = updatePriority
	}

	req.Values = values
	return req
}

// searchRequirements searches requirements by keyword
func searchRequirements(keyword string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 如果只传了 workspace-key 没传 workspace-name，自动获取
		if reqWorkspaceKey != "" && reqWorkspaceName == "" {
			spaceService := api.NewSpaceService(ctx.Config.API.BasePlatformURL, ctx.GetHeaders(reqWorkspaceKey), ctx.Client)
			if name, err := spaceService.GetSpaceNameByCode(reqWorkspaceKey); err == nil {
				reqWorkspaceName = name
			}
		}

		headers := ctx.GetHeaders(reqWorkspaceKey)
		reqService := api.NewRequirementService(ctx.Config.API.BaseReqURL, headers, ctx.Client, ctx.Config)

		// 构建 IQL 查询语句，搜索标题包含关键词的需求
		// 使用 ~ 进行模糊匹配
		iql := fmt.Sprintf("((((标题 ~ '%s' or key = '%s')) and (所属空间 = '%s')) and ('belongingSpace' in [\"currentWorkspace()\"] or ('类型' in [\"用户故事\",\"任务\"] ))) order by  创建时间 desc",
			keyword, keyword, reqWorkspaceName)

		requestData := &api.SearchRequest{
			IQL:               iql,
			Size:              listLimit,
			From:              listOffset,
			IsExpand:          false,
			IsShowAncestors:   true,
			IsShowDescendants: true,
			IsShowLinkItems:   false,
			Extend:            map[string]interface{}{},
			Fields: []string{
				"ancestors", "assignee", "createdAt", "createdBy",
				"earlyWarning", "expectedCompletionTime", "id",
				"itemType", "key", "objectId", "priority", "projectNo",
				"rowId", "status", "workspace",
			},
			RefererInfo: api.RefererInfo{
				WorkspaceKey: reqWorkspaceKey,
			},
		}

		resp, err := reqService.Search(requestData, reqWorkspaceKey)
		if err != nil {
			return nil, err
		}

		// Output the search results as JSON
		var items []map[string]interface{}
		for _, item := range resp.Payload.Items {
			statusName := "未知"
			if item.Status.Name != "" {
				statusName = item.Status.Name
			}

			assigneeName := "未分配"
			if len(item.Values.Assignee) > 0 {
				assigneeName = item.Values.Assignee[0].Nickname
			}

			items = append(items, map[string]interface{}{
				"name":      item.Name,
				"key":       item.Key,
				"objectId":  item.ObjectID,
				"status":    statusName,
				"assignee":  assigneeName,
				"createdAt": item.CreatedAt,
			})
		}

		result := map[string]interface{}{
			"keyword":    keyword,
			"count":      resp.Payload.Count,
			"totalCount": resp.Payload.TotalCount,
			"items":      items,
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
		PrettyPrint: prettyMode,
		CommandName: "req search",
	})
}
