package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/haibin1003/aaascli/internal/api"
	"github.com/haibin1003/aaascli/internal/common"
)

var (
	orderPage       int
	orderSize       int
	orderPassStatus bool
)

var orderCmd = &cobra.Command{
	Use:   "order",
	Short: "订购申请管理",
	Long:  "查询我的能力/服务订购申请记录及审批状态",
}

var orderListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询我的申请列表",
	Run: func(cmd *cobra.Command, args []string) {
		listMyApplies()
	},
}

func init() {
	rootCmd.AddCommand(orderCmd)
	orderCmd.AddCommand(orderListCmd)

	orderListCmd.Flags().IntVar(&orderPage, "page", 1, "页码")
	orderListCmd.Flags().IntVarP(&orderSize, "size", "s", 20, "每页条数")
	orderListCmd.Flags().BoolVar(&orderPassStatus, "pass-status", true, "是否只查询已通过/待审批等状态（平台语义）")
}

func listMyApplies() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		svc := api.NewApplyService(ctx.Client)
		resp, err := svc.ListMyApplies(orderPage, orderSize, orderPassStatus)
		if err != nil {
			return nil, fmt.Errorf("查询申请列表失败: %w", err)
		}
		return formatApplyList(resp), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func formatApplyList(resp *api.ApplyListResponse) map[string]interface{} {
	if resp.Data.List == nil {
		resp.Data.List = []api.ApplyItem{}
	}
	items := make([]map[string]interface{}, 0)
	for _, a := range resp.Data.List {
		items = append(items, map[string]interface{}{
			"id":           a.ID,
			"appName":      a.AppName,
			"goodsName":    a.GoodsName,
			"authType":     a.AuthType,
			"authTypeName": a.AuthTypeName,
			"status":       a.Status,
			"statusName":   a.StatusName,
			"applyTime":    a.ApplyTime,
			"passStatus":   a.PassStatus,
		})
	}
	return map[string]interface{}{
		"items": items,
		"pagination": map[string]interface{}{
			"page":  resp.Data.PageNum,
			"size":  resp.Data.PageSize,
			"total": resp.Data.Total,
			"pages": resp.Data.Pages,
		},
	}
}
