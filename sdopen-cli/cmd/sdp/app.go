package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourname/sdopen-cli/internal/api"
	"github.com/yourname/sdopen-cli/internal/common"
)

var (
	appPage       int
	appSize       int
	abilityID     string
	bomcID        string
	dailyLimit    int
	rateLimit     int
	ratePeriod    string
)

// appCmd 应用管理命令
var appCmd = &cobra.Command{
	Use:   "app",
	Short: "应用管理",
	Long: `管理我的应用，包括查看应用列表、能力授权等功能。

示例:
  # 查看应用列表
  sdp app list

  # 查看应用已授权的能力
  sdp app auth-list <应用名>

  # 授权能力给应用
  sdp app auth-ability <应用名> --ability <能力ID> --bomc <工单编码>

  # 查看授权审批状态
  sdp app auth-status <应用名>`,
}

// appListCmd 应用列表命令
var appListCmd = &cobra.Command{
	Use:   "list",
	Short: "查看我的应用列表",
	Run: func(cmd *cobra.Command, args []string) {
		listApps()
	},
}

// appAuthListCmd 查看应用已授权能力命令
var appAuthListCmd = &cobra.Command{
	Use:   "auth-list [app-name]",
	Short: "查看应用已授权的能力",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		listAppAuthAbilities(args[0])
	},
}

// appAuthAbilityCmd 能力授权命令
var appAuthAbilityCmd = &cobra.Command{
	Use:   "auth-ability [app-name]",
	Short: "为应用授权能力（需要审批）",
	Long: `为指定应用授权能力，授权申请需要审批。

必填参数:
  --ability  能力 ID
  --bomc     BOMC 工单编码

可选参数:
  --daily-limit   日调用量上限
  --rate-limit    流控限额
  --rate-period   流控周期（秒/分钟/小时/天）

示例:
  sdp app auth-ability 新员工实战应用 --ability CAxxx --bomc WOxxx
  sdp app auth-ability 新员工实战应用 --ability CAxxx --bomc WOxxx --daily-limit 10000`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		authAbilityToApp(args[0])
	},
}

// appAuthStatusCmd 授权状态命令
var appAuthStatusCmd = &cobra.Command{
	Use:   "auth-status [app-name]",
	Short: "查看应用授权审批状态",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viewAuthStatus(args[0])
	},
}

func init() {
	rootCmd.AddCommand(appCmd)
	appCmd.AddCommand(appListCmd)
	appCmd.AddCommand(appAuthListCmd)
	appCmd.AddCommand(appAuthAbilityCmd)
	appCmd.AddCommand(appAuthStatusCmd)

	// 列表命令参数
	appListCmd.Flags().IntVar(&appPage, "page", 1, "页码")
	appListCmd.Flags().IntVarP(&appSize, "size", "s", 20, "每页条数")

	// 授权命令参数
	appAuthAbilityCmd.Flags().StringVarP(&abilityID, "ability", "a", "", "能力 ID（必填）")
	appAuthAbilityCmd.Flags().StringVarP(&bomcID, "bomc", "b", "", "BOMC 工单编码（必填）")
	appAuthAbilityCmd.Flags().IntVar(&dailyLimit, "daily-limit", 0, "日调用量上限")
	appAuthAbilityCmd.Flags().IntVar(&rateLimit, "rate-limit", 0, "流控限额")
	appAuthAbilityCmd.Flags().StringVar(&ratePeriod, "rate-period", "", "流控周期")

	_ = appAuthAbilityCmd.MarkFlagRequired("ability")
	_ = appAuthAbilityCmd.MarkFlagRequired("bomc")
}

// listApps 查看应用列表
func listApps() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}

		service := api.NewAppService(ctx.Client)
		resp, err := service.List(appPage, appSize)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}

		return formatAppList(resp), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

// listAppAuthAbilities 查看应用已授权的能力
func listAppAuthAbilities(appName string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}

		// 先获取应用列表找到应用 ID
		appService := api.NewAppService(ctx.Client)
		appList, err := appService.List(1, 100)
		if err != nil {
			return nil, fmt.Errorf("获取应用列表失败: %w", err)
		}

		var targetApp *api.App
		for _, app := range appList.Data.List {
			if app.Name == appName {
				targetApp = &app
				break
			}
		}

		if targetApp == nil {
			return nil, fmt.Errorf("未找到应用: %s", appName)
		}

		resp, err := appService.ListAuthAbilities(targetApp.ID)
		if err != nil {
			return nil, fmt.Errorf("查询授权能力失败: %w", err)
		}

		return formatAuthAbilityList(resp), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

// authAbilityToApp 为应用授权能力
func authAbilityToApp(appName string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}

		// 获取应用列表找到应用 ID
		appService := api.NewAppService(ctx.Client)
		appList, err := appService.List(1, 100)
		if err != nil {
			return nil, fmt.Errorf("获取应用列表失败: %w", err)
		}

		var targetApp *api.App
		for _, app := range appList.Data.List {
			if app.Name == appName {
				targetApp = &app
				break
			}
		}

		if targetApp == nil {
			return nil, fmt.Errorf("未找到应用: %s", appName)
		}

		// 提交授权申请
		req := &api.AuthAbilityRequest{
			AppID:           targetApp.ID,
			AbilityID:       abilityID,
			BomcID:          bomcID,
			DailyLimit:      dailyLimit,
			RateLimit:       rateLimit,
			RateLimitPeriod: ratePeriod,
		}

		resp, err := appService.AuthAbility(req)
		if err != nil {
			return nil, fmt.Errorf("授权申请失败: %w", err)
		}

		if !resp.Success && resp.Code != 200 {
			return nil, fmt.Errorf("授权申请失败 [%d]: %s", resp.Code, resp.Message)
		}

		result := map[string]interface{}{
			"message":   "授权申请已提交",
			"appName":   appName,
			"abilityId": abilityID,
			"bomcId":    bomcID,
		}

		if resp.Data.NeedVerify {
			result["status"] = "pending"
			result["note"] = "申请需要审批，请等待审批结果"
		} else {
			result["status"] = "approved"
			result["note"] = "授权已生效"
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

// viewAuthStatus 查看授权状态
func viewAuthStatus(appName string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}

		// 获取应用列表找到应用 ID
		appService := api.NewAppService(ctx.Client)
		appList, err := appService.List(1, 100)
		if err != nil {
			return nil, fmt.Errorf("获取应用列表失败: %w", err)
		}

		var targetApp *api.App
		for _, app := range appList.Data.List {
			if app.Name == appName {
				targetApp = &app
				break
			}
		}

		if targetApp == nil {
			return nil, fmt.Errorf("未找到应用: %s", appName)
		}

		resp, err := appService.GetAuthStatus(targetApp.ID)
		if err != nil {
			return nil, fmt.Errorf("查询授权状态失败: %w", err)
		}

		return formatAuthStatus(resp), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

// formatAppList 格式化应用列表
func formatAppList(resp *api.AppListResponse) map[string]interface{} {
	var items []map[string]interface{}
	for _, a := range resp.Data.List {
		items = append(items, map[string]interface{}{
			"id":         a.ID,
			"name":       a.Name,
			"code":       a.Code,
			"status":     a.Status,
			"createTime": a.CreateTime,
		})
	}

	return map[string]interface{}{
		"items": items,
		"pagination": map[string]interface{}{
			"page":    resp.Data.PageNum,
			"size":    resp.Data.PageSize,
			"total":   resp.Data.Total,
			"pages":   (resp.Data.Total + resp.Data.PageSize - 1) / resp.Data.PageSize,
		},
	}
}

// formatAuthAbilityList 格式化授权能力列表
func formatAuthAbilityList(resp *api.AuthAbilityListResponse) map[string]interface{} {
	var items []map[string]interface{}
	for _, a := range resp.Data.List {
		items = append(items, map[string]interface{}{
			"authId":      a.AuthID,
			"abilityId":   a.AbilityID,
			"abilityName": a.AbilityName,
			"abilityCode": a.AbilityCode,
			"status":      a.Status,
			"applyTime":   a.ApplyTime,
		})
	}

	return map[string]interface{}{
		"items": items,
		"total": resp.Data.Total,
	}
}

// formatAuthStatus 格式化授权状态
func formatAuthStatus(resp *api.AuthStatusResponse) map[string]interface{} {
	formatItems := func(items []api.AuthAbility) []map[string]interface{} {
		var result []map[string]interface{}
		for _, a := range items {
			result = append(result, map[string]interface{}{
				"authId":      a.AuthID,
				"abilityId":   a.AbilityID,
				"abilityName": a.AbilityName,
				"abilityCode": a.AbilityCode,
				"applyTime":   a.ApplyTime,
			})
		}
		return result
	}

	return map[string]interface{}{
		"pending":  formatItems(resp.Data.Pending),
		"approved": formatItems(resp.Data.Approved),
		"rejected": formatItems(resp.Data.Rejected),
	}
}
