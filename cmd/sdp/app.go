package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/haibin1003/aaascli/internal/api"
	"github.com/haibin1003/aaascli/internal/common"
)

var (
	appPage          int
	appSize          int
	appName          string
	abilityID        string
	bomcID           string
	quotaLimit       string
	limitCount       string
	policyPeriod     string
	policyTimeUnit   string
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
	Use:   "auth-ability [app-id]",
	Short: "为应用授权能力",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		authAbilityForApp(args[0])
	},
}

var appAuthStatusCmd = &cobra.Command{
	Use:   "auth-status [app-id]",
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
	appAuthAbilityCmd.Flags().StringVar(&quotaLimit, "quota-limit", "500", "流控配额限制")
	appAuthAbilityCmd.Flags().StringVar(&limitCount, "limit-count", "500", "流控次数限制")
	appAuthAbilityCmd.Flags().StringVar(&policyPeriod, "policy-period", "1", "流控周期")
	appAuthAbilityCmd.Flags().StringVar(&policyTimeUnit, "policy-time-unit", "second", "流控周期单位")
}

func authAbilityForApp(appID string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		if abilityID == "" {
			return nil, fmt.Errorf("请使用 --ability 指定要授权的能力 ID")
		}
		if bomcID == "" {
			return nil, fmt.Errorf("请使用 --bomc 指定 BOMC 工单编码")
		}

		// 获取应用名称
		appSvc := api.NewAppService(ctx.Client)
		apps, err := appSvc.ListMyApps(1, 100, "")
		if err != nil {
			return nil, fmt.Errorf("查询应用列表失败: %w", err)
		}
		var appNameFound string
		for _, app := range apps {
			if app.AppID == appID {
				appNameFound = app.AppName
				break
			}
		}
		if appNameFound == "" {
			return nil, fmt.Errorf("未找到应用 ID: %s", appID)
		}

		// 获取能力名称
		abilitySvc := api.NewAbilityService(ctx.Client)
		detail, err := abilitySvc.GetDetail(abilityID)
		if err != nil {
			return nil, fmt.Errorf("获取能力详情失败: %w", err)
		}

		resp, err := appSvc.AuthAbility(&api.AuthAbilityRequest{
			AppID:          appID,
			AbilityID:      abilityID,
			AppName:        appNameFound,
			AuthType:       "capacity",
			Status:         "AppStatusOnline",
			BomcID:         bomcID,
			QuotaLimit:     quotaLimit,
			LimitCount:     limitCount,
			PolicyPeriod:   policyPeriod,
			PolicyTimeUnit: policyTimeUnit,
			GoodsNames:     detail.Name,
		})
		if err != nil {
			return nil, fmt.Errorf("授权失败: %w", err)
		}
		return map[string]interface{}{
			"message":     "授权请求已提交",
			"appId":       appID,
			"appName":     appNameFound,
			"abilityId":   abilityID,
			"abilityName": detail.Name,
			"bomcId":      bomcID,
			"orderId":     resp.Data.OrderID,
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
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
			"appId":          a.AppID,
			"appName":        a.AppName,
			"appLevel":       a.AppLevel,
			"status":         a.Status,
			"showStatusName": a.ShowStatusName,
			"auditStatus":    a.AuditStatus,
			"maxQuotaNum":    a.MaxQuotaNum,
			"appImgPath":     a.AppImgPath,
			"userId":         a.UserID,
			"remark":         a.Remark,
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
