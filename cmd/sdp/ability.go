package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/haibin1003/aaascli/internal/api"
	"github.com/haibin1003/aaascli/internal/common"
)

var (
	abilityPage    int
	abilitySize    int
	abilityKeyword string
)

var abilityCmd = &cobra.Command{
	Use:   "ability",
	Short: "能力管理",
	Long:  "查询和管理山东能力开放平台的服务能力",
}

var abilityListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询能力列表",
	Run: func(cmd *cobra.Command, args []string) {
		listAbilities()
	},
}

var abilitySearchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "搜索能力",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		abilityKeyword = args[0]
		searchAbilities()
	},
}

var abilityViewCmd = &cobra.Command{
	Use:   "view [ability-id]",
	Short: "查看能力详情",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viewAbility(args[0])
	},
}

var abilityOrderCmd = &cobra.Command{
	Use:   "order [ability-id]",
	Short: "订购能力",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		orderAbility(args[0])
	},
}

var abilityMyCmd = &cobra.Command{
	Use:   "my",
	Short: "查看我的能力",
	Run: func(cmd *cobra.Command, args []string) {
		listMyAbilities()
	},
}

func init() {
	rootCmd.AddCommand(abilityCmd)
	abilityCmd.AddCommand(abilityListCmd)
	abilityCmd.AddCommand(abilitySearchCmd)
	abilityCmd.AddCommand(abilityViewCmd)
	abilityCmd.AddCommand(abilityOrderCmd)
	abilityCmd.AddCommand(abilityMyCmd)

	abilityListCmd.Flags().IntVar(&abilityPage, "page", 1, "页码")
	abilityListCmd.Flags().IntVarP(&abilitySize, "size", "s", 20, "每页条数")

	abilitySearchCmd.Flags().IntVar(&abilityPage, "page", 1, "页码")
	abilitySearchCmd.Flags().IntVarP(&abilitySize, "size", "s", 20, "每页条数")

	abilityMyCmd.Flags().IntVar(&abilityPage, "page", 1, "页码")
	abilityMyCmd.Flags().IntVarP(&abilitySize, "size", "s", 20, "每页条数")
}

func listAbilities() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		service := api.NewAbilityService(ctx.Client)
		resp, err := service.List(abilityPage, abilitySize)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		return formatAbilityList(resp), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func searchAbilities() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		service := api.NewAbilityService(ctx.Client)
		resp, err := service.Search(abilityKeyword, abilityPage, abilitySize)
		if err != nil {
			return nil, fmt.Errorf("搜索失败: %w", err)
		}
		return formatAbilityList(resp), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func viewAbility(abilityID string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		service := api.NewAbilityService(ctx.Client)
		detail, err := service.GetDetail(abilityID)
		if err != nil {
			return nil, fmt.Errorf("查询详情失败: %w", err)
		}
		return formatAbilityDetail(detail), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func orderAbility(abilityID string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		service := api.NewAbilityService(ctx.Client)
		if err := service.OrderAbility(abilityID); err != nil {
			return nil, fmt.Errorf("订购失败: %w", err)
		}
		return map[string]interface{}{
			"message":   "订购请求已处理（如适用）",
			"abilityId": abilityID,
			"note":      "部分能力需要在网页端完成订购流程",
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func listMyAbilities() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		service := api.NewAbilityService(ctx.Client)
		resp, err := service.List(abilityPage, abilitySize)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		return formatAbilityList(resp), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func formatAbilityList(resp *api.AbilityListResponse) map[string]interface{} {
	if resp.Data.ProductActionList == nil {
		resp.Data.ProductActionList = []api.Ability{}
	}
	var items []map[string]interface{}
	for _, a := range resp.Data.ProductActionList {
		items = append(items, map[string]interface{}{
			"id":       a.ID,
			"name":     a.Name,
			"code":     a.Code,
			"desc":     a.Desc,
			"provider": a.Provider,
			"status":   a.Status,
		})
	}
	return map[string]interface{}{
		"items": items,
		"total": len(items),
	}
}

func formatAbilityDetail(detail *api.AbilityDetail) map[string]interface{} {
	return map[string]interface{}{
		"id":         detail.ID,
		"name":       detail.Name,
		"code":       detail.Code,
		"desc":       detail.Desc,
		"detailDesc": detail.DetailDesc,
		"provider":   detail.Provider,
		"type":       detail.TypeName,
		"callType":   detail.CallType,
		"userId":     detail.UserID,
	}
}
