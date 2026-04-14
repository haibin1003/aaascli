package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/haibin1003/aaascli/internal/api"
	"github.com/haibin1003/aaascli/internal/common"
)

var (
	appPage    int
	appSize    int
	appName    string
	abilityID  string
	bomcID     string
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "应用管理",
	Long:  "管理我的应用及能力授权信息",
}

var appListCmd = &cobra.Command{
	Use:   "list",
	Short: "查看我的应用列表",
	Run: func(cmd *cobra.Command, args []string) {
		listMyApps()
	},
}

var appAuthListCmd = &cobra.Command{
	Use:   "auth-list",
	Short: "查看能力授权列表",
	Run: func(cmd *cobra.Command, args []string) {
		listAppAuths()
	},
}

var appAuthAbilityCmd = &cobra.Command{
	Use:   "auth-ability [app-name]",
	Short: "为应用授权能力（暂未实现）",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
			return nil, fmt.Errorf("授权能力接口尚未实现，请在浏览器中手动完成")
		}, common.ExecuteOptions{PrettyPrint: prettyPrint})
	},
}

var appAuthStatusCmd = &cobra.Command{
	Use:   "auth-status [app-name]",
	Short: "查看授权审批状态（暂未实现）",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
			return nil, fmt.Errorf("授权状态查询接口尚未实现，请在浏览器中查看")
		}, common.ExecuteOptions{PrettyPrint: prettyPrint})
	},
}

func init() {
	rootCmd.AddCommand(appCmd)
	appCmd.AddCommand(appListCmd)
	appCmd.AddCommand(appAuthListCmd)
	appCmd.AddCommand(appAuthAbilityCmd)
	appCmd.AddCommand(appAuthStatusCmd)

	appListCmd.Flags().IntVar(&appPage, "page", 1, "页码")
	appListCmd.Flags().IntVarP(&appSize, "size", "s", 20, "每页条数")
	appListCmd.Flags().StringVarP(&appName, "name", "n", "", "应用名称过滤")

	appAuthAbilityCmd.Flags().StringVarP(&abilityID, "ability", "a", "", "能力 ID")
	appAuthAbilityCmd.Flags().StringVarP(&bomcID, "bomc", "b", "", "BOMC 工单编码")
}

func listMyApps() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		service := api.NewAppService(ctx.Client)
		items, err := service.ListMyApps(appPage, appSize, appName)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		return formatMyAppList(items, appPage, appSize), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func listAppAuths() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		service := api.NewAppService(ctx.Client)
		resp, err := service.List()
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		return formatAppAuthList(resp), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func formatMyAppList(items []api.MyApp, page, size int) map[string]interface{} {
	if items == nil {
		items = []api.MyApp{}
	}
	total := len(items)
	start := (page - 1) * size
	end := start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	pageItems := items[start:end]

	formatted := make([]map[string]interface{}, 0)
	for _, a := range pageItems {
		formatted = append(formatted, map[string]interface{}{
			"appId":       a.AppID,
			"appName":     a.AppName,
			"appCode":     a.AppCode,
			"status":      a.Status,
			"statusName":  a.StatusName,
			"createTime":  a.CreateTime,
			"description": a.Description,
		})
	}

	return map[string]interface{}{
		"items": formatted,
		"pagination": map[string]interface{}{
			"page":  page,
			"size":  size,
			"total": total,
			"pages": (total + size - 1) / size,
		},
	}
}

func formatAppAuthList(resp *api.AppAuthListResponse) map[string]interface{} {
	if resp.Data.AuthorizedList == nil {
		resp.Data.AuthorizedList = []api.AppAuth{}
	}
	items := make([]map[string]interface{}, 0)
	for _, a := range resp.Data.AuthorizedList {
		items = append(items, map[string]interface{}{
			"appName":        a.AppName,
			"abilityName":    a.AbilityName,
			"abilityCode":    a.AbilityCode,
			"authStatus":     a.AuthStatus,
			"authStatusName": a.AuthStatusName,
			"applyTime":      a.ApplyTime,
		})
	}
	return map[string]interface{}{
		"items": items,
		"total": len(items),
	}
}
