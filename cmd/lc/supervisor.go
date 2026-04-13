package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
	"github.com/user/lc/internal/config"
)

// supervisorCheckinCmd 检查今日签到人员信息
var supervisorCheckinCmd = &cobra.Command{
	Use:   "checkin",
	Short: "查询今日签到人员列表",
	Long: `查询今天所有需要签到人员的签到信息，显示每个人员所在分组。

示例:
  # 查询今日签到
  lc supervisor checkin

  # 输出格式:
  {
    "total": 42,
    "records": [
      {
        "userId": 249,
        "loginName": "guihaiqing_it",
        "nickName": "桂海清",
        "checkInTime": "2026-03-20 08:25:17",
        "groupInfo": "原型设计1组,设计管理1组,数据建模1组,接口设计1组,架构设计1组,UI设计1组,研发桌面1组"
      },
      ...
    ]
  }`,
  Run: func(cmd *cobra.Command, args []string) {
		common.ExecuteWithOutput(func(ctx *common.CommandContext) error {
			return execSupervisorCheckin(ctx, cmd)
		}, common.ExecuteOptions{
			DebugMode: debugMode,
			Insecure:  insecureSkipVerify,
			Logger:    &logger,
			PrettyPrint: prettyMode,
		})
	},
}

// execSupervisorCheckin 执行签到列表查询
func execSupervisorCheckin(ctx *common.CommandContext, cmd *cobra.Command) error {
	// 加载配置
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if cfg.Cookie == "" {
		return fmt.Errorf("未登录，请先执行: lc login <cookie-value>")
	}

	// 创建HTTP客户端
	var client *api.Client
	if insecureSkipVerify {
		client = api.NewInsecureClient()
	} else {
		client = api.NewClient()
	}

	// 创建监管平台服务并执行认证
	svc := api.NewSupervisorService(cfg.Cookie, client)

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 开始认证流程...")
	}

	_, err = svc.Authenticate()
	if err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 认证成功，开始获取签到列表...")
	}

	// 获取签到列表
	checkin, err := svc.GetCheckinList()
	if err != nil {
		return fmt.Errorf("获取签到列表失败: %w", err)
	}

	// 获取output参数
	outputFile, _ := cmd.Flags().GetString("output")

	// 如果指定了输出CSV文件
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("创建输出文件失败: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		// CSV header
		writer.Write([]string{"userId", "loginName", "nickName", "checkInTime", "groupInfo"})

		// 写入每一行
		for _, u := range checkin.Result {
			writer.Write([]string{
				fmt.Sprintf("%d", u.UserID),
				u.LoginName,
				u.NickName,
				u.CheckInTime,
				u.GroupInfo,
			})
		}

		writer.Flush()
		return nil
	}

	// 默认输出JSON
	output, err := json.MarshalIndent(checkin, "", "  ")
	if err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func init() {
	rootCmd.AddCommand(supervisorCmd)
	supervisorCmd.AddCommand(supervisorOverviewCmd)
	supervisorCmd.AddCommand(supervisorTodoCmd)
	supervisorTodoCmd.AddCommand(supervisorTodoOverviewCmd)
	supervisorCmd.AddCommand(supervisorGroupsCmd)
	supervisorGroupsCmd.AddCommand(supervisorGroupsListCmd)
	supervisorGroupsCmd.AddCommand(supervisorGroupsMembersCmd)
	supervisorCmd.AddCommand(supervisorWorkorderCmd)
	supervisorWorkorderCmd.AddCommand(supervisorWorkorderListCmd)
	supervisorCmd.AddCommand(supervisorWorkordervalueCmd)
	supervisorWorkordervalueCmd.AddCommand(supervisorWorkordervalueListCmd)
	supervisorCmd.AddCommand(supervisorCheckinCmd)

	// 添加分页和筛选参数 - workorder
	supervisorWorkorderListCmd.Flags().IntP("page", "p", 1, "当前页码")
	supervisorWorkorderListCmd.Flags().IntP("size", "s", 10, "每页大小")
	supervisorWorkorderListCmd.Flags().StringP("accept-group", "g", "", "受理组编码筛选")
	supervisorWorkorderListCmd.Flags().StringP("status", "t", "", "工单状态筛选")
	supervisorWorkorderListCmd.Flags().StringP("type", "y", "", "投诉类型编码筛选")
	supervisorWorkorderListCmd.Flags().StringP("output", "o", "", "输出到CSV文件 (默认输出到stdout，打印JSON)")

	// 添加workordervalue output参数
	supervisorWorkordervalueListCmd.Flags().IntP("page", "p", 1, "当前页码")
	supervisorWorkordervalueListCmd.Flags().IntP("size", "s", 10, "每页大小")
	supervisorWorkordervalueListCmd.Flags().StringP("accept-group", "g", "", "受理组编码筛选")
	supervisorWorkordervalueListCmd.Flags().StringP("status", "t", "", "工单状态筛选，多个用逗号分隔")
	supervisorWorkordervalueListCmd.Flags().StringP("type", "y", "", "投诉类型编码筛选")
	supervisorWorkordervalueListCmd.Flags().StringP("start-time", "S", "", "创建开始时间 (格式: 2026-03-08)")
	supervisorWorkordervalueListCmd.Flags().StringP("end-time", "E", "", "创建结束时间 (格式: 2026-03-18)")
	supervisorWorkordervalueListCmd.Flags().StringP("code", "c", "", "工单编号筛选")
	supervisorWorkordervalueListCmd.Flags().StringP("module-code", "m", "", "模块编码筛选")
	supervisorWorkordervalueListCmd.Flags().StringP("menu-code", "M", "", "菜单编码筛选")
	supervisorWorkordervalueListCmd.Flags().StringP("output", "o", "", "输出到CSV文件 (默认输出到stdout，打印JSON)")

	// 添加 output 参数
	supervisorGroupsListCmd.Flags().StringP("output", "o", "", "输出到CSV文件 (默认输出到stdout，打印JSON)")
	supervisorTodoOverviewCmd.Flags().StringP("output", "o", "", "输出到CSV文件 (默认输出到stdout，打印JSON)")
	supervisorCheckinCmd.Flags().StringP("output", "o", "", "输出到CSV文件 (默认输出到stdout，打印JSON)")
	supervisorGroupsMembersCmd.Flags().StringP("output", "o", "", "输出到CSV文件 (默认输出到stdout，打印JSON)")
	
	// 添加工作组列表筛选参数
	supervisorGroupsListCmd.Flags().IntP("page", "p", 1, "当前页码")
	supervisorGroupsListCmd.Flags().IntP("size", "s", 10, "每页大小")
	supervisorGroupsListCmd.Flags().StringP("code", "c", "", "工作组编码筛选")
	supervisorGroupsListCmd.Flags().StringP("name", "n", "", "工作组名称筛选")
	supervisorGroupsListCmd.Flags().StringP("description", "", "", "工作组描述筛选")
	
	// 添加工作组成员参数
	supervisorGroupsMembersCmd.Flags().StringP("group-code", "c", "", "工作组编码 (必填，如果不指定则必须指定 --group-name-filter)")
	supervisorGroupsMembersCmd.Flags().StringP("group-name-filter", "f", "", "工作组名称筛选字符串 (用于自动查找名称包含此字符串的工作组)")
	supervisorOverviewCmd.Flags().StringP("output", "o", "", "输出到CSV文件 (默认输出到stdout，打印JSON)")
}

// supervisorCmd 监管模块根命令
var supervisorCmd = &cobra.Command{
	Use:   "supervisor",
	Short: "监管模块相关命令",
	Long: `监管模块命令集合，提供待办查询等监管平台功能。

认证说明:
  所有supervisor子命令都需要先进行两步认证:
  1. 请求MOSS认证接口获取跳转URL和code
  2. 使用code换取监管平台的Authentication token

  认证过程已封装在 api.SupervisorService.Authenticate() 方法中。`,
}

// supervisorTodoCmd 待办相关命令
var supervisorTodoCmd = &cobra.Command{
	Use:   "todo",
	Short: "待办相关命令",
	Long:  "查询监管平台待办数量等信息",
}

// supervisorTodoOverviewCmd 查询待办概览
var supervisorTodoOverviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "查询待办概览",
	Long: `查询监管平台的待办概览统计。

示例:
  # 查询待办概览
  lc supervisor todo overview

输出格式:
  {
    "组内待认领工单": {
      "问题咨询": "0",
      "普通投诉": "0",
      "升级投诉": "0",
      "意见反馈": "0"
    },
    "我的待办": {
      "问题咨询": "0",
      "普通投诉": "0",
      "升级投诉": "0",
      "意见反馈": "0"
    }
  }`,
	Run: func(cmd *cobra.Command, args []string) {
		common.ExecuteWithOutput(func(ctx *common.CommandContext) error {
			return execSupervisorTodoOverview(ctx, cmd)
		}, common.ExecuteOptions{
			DebugMode: debugMode,
			Insecure:  insecureSkipVerify,
			Logger:    &logger,
			PrettyPrint: prettyMode,
		})
	},
}

// execSupervisorTodoOverview 执行待办概览查询
func execSupervisorTodoOverview(ctx *common.CommandContext, cmd *cobra.Command) error {
	// 加载配置
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if cfg.Cookie == "" {
		return fmt.Errorf("未登录，请先执行: lc login <cookie-value>")
	}

	// 创建HTTP客户端
	var client *api.Client
	if insecureSkipVerify {
		client = api.NewInsecureClient()
	} else {
		client = api.NewClient()
	}

	// 创建监管平台服务并执行认证
	svc := api.NewSupervisorService(cfg.Cookie, client)

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 开始认证流程...")
	}

	_, err = svc.Authenticate()
	if err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 认证成功，开始获取待办数据...")
	}

	// 获取主面板概览数据
	overview, err := svc.GetMainOverview()
	if err != nil {
		return fmt.Errorf("获取待办数据失败: %w", err)
	}

	// 格式化输出
	result := formatTodoOverview(overview)

	// 获取output参数
	outputFile, _ := cmd.Flags().GetString("output")

	// 如果指定了输出CSV文件
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("创建输出文件失败: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		// CSV header
		writer.Write([]string{"category", "ticketType", "count"})

		// 写入每一行
		for category, tickets := range result {
			for ticketType, count := range tickets {
				writer.Write([]string{category, ticketType, count})
			}
		}

		writer.Flush()
		return nil
	}

	// 默认输出JSON
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// formatTodoOverview 格式化待办概览输出
func formatTodoOverview(overview *api.MainOverviewResponse) map[string]map[string]string {
	result := make(map[string]map[string]string)

	for _, stat := range overview.Result.Stat {
		category := make(map[string]string)
		for _, value := range stat.Values {
			category[value.Label] = value.Value
		}
		// 使用StatItem的Label作为分类名称（如"组内待认领工单"、"我的待办"）
		result[stat.Label] = category
	}

	return result
}

// supervisorOverviewCmd 在途工单概览命令
var supervisorOverviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "输出当前在途工单概览",
	Long: `输出当前在途工单概览统计，显示组内待认领工单和我的待办中的各种工单类型数量。

示例:
  # 查询当前在途工单概览
  lc supervisor overview

输出格式:
  {
    "组内待认领工单": {
      "问题咨询": "0",
      "普通投诉": "1",
      "升级投诉": "0",
      "意见反馈": "0"
    },
    "我的待办": {
      "问题咨询": "28",
      "普通投诉": "2",
      "升级投诉": "0",
      "意见反馈": "1"
    }
  }`,
	Run: func(cmd *cobra.Command, args []string) {
		common.ExecuteWithOutput(func(ctx *common.CommandContext) error {
			return execSupervisorOverview(ctx, cmd)
		}, common.ExecuteOptions{
			DebugMode: debugMode,
			Insecure:  insecureSkipVerify,
			Logger:    &logger,
			PrettyPrint: prettyMode,
		})
	},
}

// execSupervisorOverview 执行在途工单概览查询
func execSupervisorOverview(ctx *common.CommandContext, cmd *cobra.Command) error {
	// 加载配置
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if cfg.Cookie == "" {
		return fmt.Errorf("未登录，请先执行: lc login <cookie-value>")
	}

	// 创建HTTP客户端
	var client *api.Client
	if insecureSkipVerify {
		client = api.NewInsecureClient()
	} else {
		client = api.NewClient()
	}

	// 创建监管平台服务并执行认证
	svc := api.NewSupervisorService(cfg.Cookie, client)

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 开始认证流程...")
	}

	_, err = svc.Authenticate()
	if err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 认证成功，开始获取在途工单概览...")
	}

	// 获取主面板概览数据（在途工单概览）
	overview, err := svc.GetMainOverview()
	if err != nil {
		return fmt.Errorf("获取在途工单概览失败: %w", err)
	}

	// 格式化输出
	result := formatTodoOverview(overview)

	// 获取output参数
	outputFile, _ := cmd.Flags().GetString("output")

	// 如果指定了输出CSV文件
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("创建输出文件失败: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		// CSV header
		writer.Write([]string{"category", "ticketType", "count"})

		// 写入每一行
		for category, tickets := range result {
			for ticketType, count := range tickets {
				writer.Write([]string{category, ticketType, count})
			}
		}

		writer.Flush()
		return nil
	}

	// 默认输出JSON
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// supervisorGroupsCmd 分组相关命令
var supervisorGroupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "工作组相关命令",
	Long:  "查询监管平台业务工作组信息",
}

// supervisorGroupsMembersCmd 查询工作组成员详情
var supervisorGroupsMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "查询工作组成员详情",
	Long: `查询指定工作组的成员列表，支持按工作组编码查询或自动查找名称包含指定字符串的工作组。

示例:
  # 查询指定工作组的成员
  lc supervisor groups members --group-code it010

  # 自动查询名称包含 "1组" 的所有工作组成员
  lc supervisor groups members --group-name-filter "1组"

  # 输出到CSV文件
  lc supervisor groups members --group-name-filter "1组" -o members.csv

  输出为JSON格式，包含成员的详细信息：
  - ID: 用户ID
  - NickName: 昵称
  - Phone: 手机号（加密）
  - LeaderFlag: 是否组长`,
	Run: func(cmd *cobra.Command, args []string) {
		common.ExecuteWithOutput(func(ctx *common.CommandContext) error {
			return execSupervisorGroupsMembers(ctx, cmd)
		}, common.ExecuteOptions{
			DebugMode: debugMode,
			Insecure:  insecureSkipVerify,
			Logger:    &logger,
			PrettyPrint: prettyMode,
		})
	},
}

// supervisorGroupsListCmd 查询全量工作组列表
var supervisorGroupsListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询全量工作组列表",
	Long: `查询监管平台全量业务工作组列表，支持分页和筛选。

示例:
  # 查询全量工作组列表
  lc supervisor groups list

  # 查询第2页，每页20条
  lc supervisor groups list --page 2 --size 20

  # 按编码筛选
  lc supervisor groups list --code it001

  # 按名称筛选
  lc supervisor groups list --name "知识库"

  # 组合筛选
  lc supervisor groups list -p 1 -s 10 -c it001 -n "知识库"

  输出为JSON格式，包含所有工作组的详细信息：
  - ID: 工作组ID
  - Code: 工作组编码
  - Name: 工作组名称
  - Label: 标签（如开发域一线、管理域升级等）
  - 以及其他创建时间、修改时间等信息`,
	Run: func(cmd *cobra.Command, args []string) {
		common.ExecuteWithOutput(func(ctx *common.CommandContext) error {
			return execSupervisorGroupsList(ctx, cmd)
		}, common.ExecuteOptions{
			DebugMode: debugMode,
			Insecure:  insecureSkipVerify,
			Logger:    &logger,
			PrettyPrint: prettyMode,
		})
	},
}

// execSupervisorGroupsList 执行全量工作组列表查询
func execSupervisorGroupsList(ctx *common.CommandContext, cmd *cobra.Command) error {
	// 读取参数
	page, _ := cmd.Flags().GetInt("page")
	pageSize, _ := cmd.Flags().GetInt("size")
	code, _ := cmd.Flags().GetString("code")
	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	outputFile, _ := cmd.Flags().GetString("output")

	// 加载配置
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if cfg.Cookie == "" {
		return fmt.Errorf("未登录，请先执行: lc login <cookie-value>")
	}

	// 创建HTTP客户端
	var client *api.Client
	if insecureSkipVerify {
		client = api.NewInsecureClient()
	} else {
		client = api.NewClient()
	}

	// 创建监管平台服务并执行认证
	svc := api.NewSupervisorService(cfg.Cookie, client)

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 开始认证流程...")
	}

	_, err = svc.Authenticate()
	if err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 认证成功，开始获取工作组列表...")
	}

	// 获取工作组列表（支持筛选）
	groups, err := svc.GetGroupList(page, pageSize, code, name, description)
	if err != nil {
		return fmt.Errorf("获取工作组列表失败: %w", err)
	}

	// 如果指定了输出CSV文件
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("创建输出文件失败: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		// CSV header
		writer.Write([]string{"id", "code", "name", "defaultFlag", "label", "deleteFlag", "tenantId", "tenantName", "createStaffId", "createStaffName", "modifyStaffId", "modifyStaffName", "createTime", "modifyTime"})

		// 写入每一行
		for _, g := range groups.Result.Records {
			writer.Write([]string{
				fmt.Sprintf("%d", g.ID),
				g.Code,
				g.Name,
				g.DefaultFlag,
				g.Label,
				fmt.Sprintf("%d", g.DeleteFlag),
				g.TenantId,
				g.TenantName,
				g.CreateStaffId,
				g.CreateStaffName,
				g.ModifyStaffId,
				g.ModifyStaffName,
				g.CreateTime,
				g.ModifyTime,
			})
		}

		writer.Flush()
		return nil
	}

	// 默认输出JSON
	output, err := json.MarshalIndent(groups.Result.Records, "", "  ")
	if err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// execSupervisorGroupsMembers 执行工作组成员查询
func execSupervisorGroupsMembers(ctx *common.CommandContext, cmd *cobra.Command) error {
	// 读取参数
	groupCode, _ := cmd.Flags().GetString("group-code")
	groupNameFilter, _ := cmd.Flags().GetString("group-name-filter")
	outputFile, _ := cmd.Flags().GetString("output")

	// 加载配置
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if cfg.Cookie == "" {
		return fmt.Errorf("未登录，请先执行: lc login <cookie-value>")
	}

	// 创建HTTP客户端
	var client *api.Client
	if insecureSkipVerify {
		client = api.NewInsecureClient()
	} else {
		client = api.NewClient()
	}

	// 创建监管平台服务并执行认证
	svc := api.NewSupervisorService(cfg.Cookie, client)

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 开始认证流程...")
	}

	_, err = svc.Authenticate()
	if err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 认证成功，开始获取工作组成员...")
	}

	var allMembers []api.GroupMemberByGroupCode

	// 如果指定了 groupCode，直接查询该工作组的成员
	if groupCode != "" {
		members, err := svc.GetGroupMembersByCode(groupCode)
		if err != nil {
			return fmt.Errorf("获取工作组成员失败: %w", err)
		}
		allMembers = append(allMembers, members.Result...)
	} else {
		// 如果没有指定 groupCode，则自动查找名称包含 groupNameFilter 的工作组
		if groupNameFilter == "" {
			return fmt.Errorf("请指定 --group-code 参数或 --group-name-filter 参数")
		}

		// 获取所有工作组列表（获取足够多的页数以覆盖所有工作组）
		// 这里假设工作组总数不会超过 10000 个
		groups, err := svc.GetGroupList(1, 10000, "", "", "")
		if err != nil {
			return fmt.Errorf("获取工作组列表失败: %w", err)
		}

		// 筛选名称包含指定字符串的工作组
		var targetGroups []api.Group
		for _, g := range groups.Result.Records {
			if strings.Contains(g.Name, groupNameFilter) {
				targetGroups = append(targetGroups, g)
			}
		}

		if len(targetGroups) == 0 {
			fmt.Printf("未找到名称包含 \"%s\" 的工作组\n", groupNameFilter)
			return nil
		}

		if debugMode {
			fmt.Fprintf(os.Stderr, "[debug] 找到 %d 个匹配的工作组\n", len(targetGroups))
		}

		// 遍历每个目标工作组，获取成员
		for _, g := range targetGroups {
			if debugMode {
				fmt.Fprintf(os.Stderr, "[debug] 正在获取工作组 \"%s\" (Code: %s) 的成员...\n", g.Name, g.Code)
			}

			members, err := svc.GetGroupMembersByCode(g.Code)
			if err != nil {
				// 记录错误但继续处理其他工作组
				fmt.Fprintf(os.Stderr, "获取工作组 \"%s\" 成员失败: %v\n", g.Name, err)
				continue
			}

			allMembers = append(allMembers, members.Result...)
		}
	}

	// 如果指定了输出CSV文件
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("创建输出文件失败: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		// CSV header
		writer.Write([]string{"id", "nickName", "phone", "leaderFlag"})

		// 写入每一行
		for _, m := range allMembers {
			writer.Write([]string{
				fmt.Sprintf("%d", m.ID),
				m.NickName,
				m.Phone,
				m.LeaderFlag,
			})
		}

		writer.Flush()
		return nil
	}

	// 默认输出JSON
	output, err := json.MarshalIndent(allMembers, "", "  ")
	if err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// supervisorWorkorderCmd 工单相关命令
var supervisorWorkorderCmd = &cobra.Command{
	Use:   "workorder",
	Short: "工单相关命令",
	Long:  "查询监管平台工单列表等信息",
}

// supervisorWorkorderListCmd 查询工单列表（支持分页筛选）
var supervisorWorkorderListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询工单列表",
	Long: `查询监管平台工单列表，支持分页和筛选。

示例:
  # 查询第一页，每页10条
  lc supervisor workorder list

  # 查询第2页，每页20条
  lc supervisor workorder list --page 2 --size 20

  # 按受理组筛选
  lc supervisor workorder list --accept-group it070

  # 组合筛选
  lc supervisor workorder list -p 1 -s 20 -g it070 -t JD03

  # 输出为JSON格式，包含所有工单详细信息：
  # - 工单ID、编号、标题、状态等基本信息
  # - 投诉人、受理人、处理人信息
  # - 创建时间、受理时间、处理时间
  # - 所属模块、菜单、平台信息
  # - 满意度等处理结果信息`,
 	Run: func(cmd *cobra.Command, args []string) {
		common.ExecuteWithOutput(func(ctx *common.CommandContext) error {
			return execSupervisorWorkorderList(ctx, cmd)
		}, common.ExecuteOptions{
			DebugMode: debugMode,
			Insecure:  insecureSkipVerify,
			Logger:    &logger,
			PrettyPrint: prettyMode,
		})
	},
}

// execSupervisorWorkorderList 执行工单列表查询
func execSupervisorWorkorderList(ctx *common.CommandContext, cmd *cobra.Command) error {
	// 读取分页参数
	page, _ := cmd.Flags().GetInt("page")
	pageSize, _ := cmd.Flags().GetInt("size")
	acceptGroup, _ := cmd.Flags().GetString("accept-group")
	status, _ := cmd.Flags().GetString("status")
	typ, _ := cmd.Flags().GetString("type")
	outputFile, _ := cmd.Flags().GetString("output")

	// 加载配置
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if cfg.Cookie == "" {
		return fmt.Errorf("未登录，请先执行: lc login <cookie-value>")
	}

	// 创建HTTP客户端
	var client *api.Client
	if insecureSkipVerify {
		client = api.NewInsecureClient()
	} else {
		client = api.NewClient()
	}

	// 创建监管平台服务并执行认证
	svc := api.NewSupervisorService(cfg.Cookie, client)

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 开始认证流程...")
	}

	_, err = svc.Authenticate()
	if err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 认证成功，开始获取工单列表...")
	}

	// 构建筛选参数
	params := make(map[string]string)
	if acceptGroup != "" {
		params["acceptGroupCode"] = acceptGroup
	}
	if status != "" {
		params["status"] = status
	}
	if typ != "" {
		params["complaintTypeCode"] = typ
	}

	// 获取工单列表
	workorders, err := svc.GetWorkOrderList(page, pageSize, params)
	if err != nil {
		return fmt.Errorf("获取工单列表失败: %w", err)
	}

	// 如果指定了输出CSV文件
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("创建输出文件失败: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		// CSV header
		writer.Write([]string{
			"id", "code", "complaintStaffName", "provinceName", "provinceCode", "complaintTypeName",
			"complaintTypeCode", "createTime", "modifyTime", "acceptTime", "handleTime", "status",
			"menuName", "moduleName", "moduleCode", "workorderContent", "platformName", "platformCode",
			"acceptGroupName", "acceptGroupCode", "acceptStaffId", "acceptStaffName", "handleStaffId", "handleStaffName",
			"historyHandleStaffId", "handleDescription", "satisfactionValue", "satisfactionResultId", "moreTimeInfo", "moreTimeStatus",
			"menuUrl", "childModuleName", "childModuleCode", "menuCode", "initTenantId", "initTenantName",
			"tenantId", "tenantName", "complaintStaffMobile", "complaintStaffEmail", "satStatus", "createStaffId", "createStaffName",
			"kcbStatus", "errorMsg",
		})

		// 写入每一行
		for _, wo := range workorders.Result.Records {
			writer.Write([]string{
				fmt.Sprintf("%d", wo.ID),
				wo.Code,
				wo.ComplaintStaffName,
				wo.ProvinceName,
				wo.ProvinceCode,
				wo.ComplaintTypeName,
				wo.ComplaintTypeCode,
				wo.CreateTime,
				wo.ModifyTime,
				wo.AcceptTime,
				wo.HandleTime,
				wo.Status,
				wo.MenuName,
				wo.ModuleName,
				wo.ModuleCode,
				wo.WorkorderContent,
				wo.PlatformName,
				wo.PlatformCode,
				wo.AcceptGroupName,
				wo.AcceptGroupCode,
				wo.AcceptStaffId,
				wo.AcceptStaffName,
				wo.HandleStaffId,
				wo.HandleStaffName,
				wo.HistoryHandleStaffId,
				wo.HandleDescription,
				wo.SatisfactionValue,
				wo.SatisfactionResultId,
				wo.MoreTimeInfo,
				wo.MoreTimeStatus,
				wo.MenuUrl,
				wo.ChildModuleName,
				wo.ChildModuleCode,
				wo.MenuCode,
				wo.InitTenantId,
				wo.InitTenantName,
				wo.TenantId,
				wo.TenantName,
				wo.ComplaintStaffMobile,
				wo.ComplaintStaffEmail,
				wo.SatStatus,
				wo.CreateStaffId,
				wo.CreateStaffName,
				wo.KcbStatus,
				wo.ErrorMsg,
			})
		}

		writer.Flush()
		return nil
	}

	// 格式化输出 - 输出完整结果包括分页信息
	result := struct {
		Page    int             `json:"page"`
		Size    int             `json:"size"`
		Total   int             `json:"total"`
		Records []api.WorkOrder `json:"records"`
	}{
		Page:    workorders.Result.Current,
		Size:    workorders.Result.Size,
		Total:   workorders.Result.Total,
		Records: workorders.Result.Records,
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// supervisorWorkordervalueCmd 工单价值模块根命令
var supervisorWorkordervalueCmd = &cobra.Command{
	Use:   "workordervalue",
	Short: "工单价值模块",
	Long:  "工单价值相关查询，支持创建时间范围等更多筛选条件",
}

// supervisorWorkordervalueListCmd 查询工单价值列表
var supervisorWorkordervalueListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询工单价值列表",
	Long: `查询监管平台工单价值列表，支持分页和多种筛选条件，包括创建时间范围。

示例:
  # 查询第一页，默认每页10条
  lc supervisor workordervalue list

  # 按创建时间范围筛选
  lc supervisor workordervalue list --start-time 2026-03-08 --end-time 2026-03-18

  # 组合筛选 - 查询指定状态和时间范围内的工单
  lc supervisor workordervalue list -p 1 -s 20 -t "JD04,JD05,JD31,JD32" -y ptts --start-time 2026-03-08 --end-time 2026-03-18

  # 按受理组和工单编号筛选
  lc supervisor workordervalue list -g it122 -c PTTS2026031809260023

  支持的筛选参数:
  - 分页: --page/-p 当前页码，--size/-s 每页大小
  - 受理组: --accept-group/-g 受理组编码
  - 状态: --status/-t 工单状态，多个用逗号分隔
  - 类型: --type/-y 投诉类型编码
  - 时间范围: --start-time/-S 创建开始时间，--end-time/-E 创建结束时间
  - 工单编号: --code/-c 工单编号
  - 模块编码: --module-code/-m 模块编码
  - 菜单编码: --menu-code/-M 菜单编码`,
 	Run: func(cmd *cobra.Command, args []string) {
		common.ExecuteWithOutput(func(ctx *common.CommandContext) error {
			return execSupervisorWorkordervalueList(ctx, cmd)
		}, common.ExecuteOptions{
			DebugMode: debugMode,
			Insecure:  insecureSkipVerify,
			Logger:    &logger,
			PrettyPrint: prettyMode,
		})
	},
}

// execSupervisorWorkordervalueList 执行工单价值列表查询
func execSupervisorWorkordervalueList(ctx *common.CommandContext, cmd *cobra.Command) error {
	// 读取分页参数
	page, _ := cmd.Flags().GetInt("page")
	pageSize, _ := cmd.Flags().GetInt("size")
	acceptGroup, _ := cmd.Flags().GetString("accept-group")
	status, _ := cmd.Flags().GetString("status")
	typ, _ := cmd.Flags().GetString("type")
	startTime, _ := cmd.Flags().GetString("start-time")
	endTime, _ := cmd.Flags().GetString("end-time")
	code, _ := cmd.Flags().GetString("code")
	moduleCode, _ := cmd.Flags().GetString("module-code")
	menuCode, _ := cmd.Flags().GetString("menu-code")
	outputFile, _ := cmd.Flags().GetString("output")

	// 加载配置
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if cfg.Cookie == "" {
		return fmt.Errorf("未登录，请先执行: lc login <cookie-value>")
	}

	// 创建HTTP客户端
	var client *api.Client
	if insecureSkipVerify {
		client = api.NewInsecureClient()
	} else {
		client = api.NewClient()
	}

	// 创建监管平台服务并执行认证
	svc := api.NewSupervisorService(cfg.Cookie, client)

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 开始认证流程...")
	}

	_, err = svc.Authenticate()
	if err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	if debugMode {
		fmt.Fprintln(os.Stderr, "[debug] 认证成功，开始获取工单列表...")
	}

	// 构建筛选参数
	params := make(map[string]string)
	if acceptGroup != "" {
		params["acceptGroupCode"] = acceptGroup
	}
	if status != "" {
		params["status"] = status
	}
	if typ != "" {
		params["complaintTypeCode"] = typ
	}
	if startTime != "" {
		params["createStartTime"] = startTime
	}
	if endTime != "" {
		params["createEndTime"] = endTime
	}
	if code != "" {
		params["code"] = code
	}
	if moduleCode != "" {
		params["moduleCode"] = moduleCode
	}
	if menuCode != "" {
		params["menuCode"] = menuCode
	}

	// 获取工单列表
	workorders, err := svc.GetWorkOrderList(page, pageSize, params)
	if err != nil {
		return fmt.Errorf("获取工单列表失败: %w", err)
	}

	// 如果指定了输出CSV文件
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("创建输出文件失败: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		// CSV header
		writer.Write([]string{
			"id", "code", "complaintStaffName", "provinceName", "provinceCode", "complaintTypeName",
			"complaintTypeCode", "createTime", "modifyTime", "acceptTime", "handleTime", "status",
			"menuName", "moduleName", "moduleCode", "workorderContent", "platformName", "platformCode",
			"acceptGroupName", "acceptGroupCode", "acceptStaffId", "acceptStaffName", "handleStaffId", "handleStaffName",
			"historyHandleStaffId", "handleDescription", "satisfactionValue", "satisfactionResultId", "moreTimeInfo", "moreTimeStatus",
			"menuUrl", "childModuleName", "childModuleCode", "menuCode", "initTenantId", "initTenantName",
			"tenantId", "tenantName", "complaintStaffMobile", "complaintStaffEmail", "satStatus", "createStaffId", "createStaffName",
			"kcbStatus", "errorMsg",
		})

		// 写入每一行
		for _, wo := range workorders.Result.Records {
			writer.Write([]string{
				fmt.Sprintf("%d", wo.ID),
				wo.Code,
				wo.ComplaintStaffName,
				wo.ProvinceName,
				wo.ProvinceCode,
				wo.ComplaintTypeName,
				wo.ComplaintTypeCode,
				wo.CreateTime,
				wo.ModifyTime,
				wo.AcceptTime,
				wo.HandleTime,
				wo.Status,
				wo.MenuName,
				wo.ModuleName,
				wo.ModuleCode,
				wo.WorkorderContent,
				wo.PlatformName,
				wo.PlatformCode,
				wo.AcceptGroupName,
				wo.AcceptGroupCode,
				wo.AcceptStaffId,
				wo.AcceptStaffName,
				wo.HandleStaffId,
				wo.HandleStaffName,
				wo.HistoryHandleStaffId,
				wo.HandleDescription,
				wo.SatisfactionValue,
				wo.SatisfactionResultId,
				wo.MoreTimeInfo,
				wo.MoreTimeStatus,
				wo.MenuUrl,
				wo.ChildModuleName,
				wo.ChildModuleCode,
				wo.MenuCode,
				wo.InitTenantId,
				wo.InitTenantName,
				wo.TenantId,
				wo.TenantName,
				wo.ComplaintStaffMobile,
				wo.ComplaintStaffEmail,
				wo.SatStatus,
				wo.CreateStaffId,
				wo.CreateStaffName,
				wo.KcbStatus,
				wo.ErrorMsg,
			})
		}

		writer.Flush()
		return nil
	}

	// 格式化输出 - 输出完整结果包括分页信息
	result := struct {
		Page    int             `json:"page"`
		Size    int             `json:"size"`
		Total   int             `json:"total"`
		Records []api.WorkOrder `json:"records"`
	}{
		Page:    workorders.Result.Current,
		Size:    workorders.Result.Size,
		Total:   workorders.Result.Total,
		Records: workorders.Result.Records,
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("格式化输出失败: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
