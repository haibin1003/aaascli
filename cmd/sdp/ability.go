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

// abilityCmd 能力管理命令
var abilityCmd = &cobra.Command{
	Use:   "ability",
	Short: "能力管理（对外服务）",
	Long: `查询和管理山东能力开放平台的对外服务能力。

示例:
  # 查询能力列表
  sdp ability list

  # 搜索能力
  sdp ability search "定位"

  # 查看能力详情
  sdp ability view <ability-id>

  # 订购能力
  sdp ability order <ability-id>

  # 查看我的能力
  sdp ability my`,
}

// abilityListCmd 能力列表命令
var abilityListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询能力列表",
	Run: func(cmd *cobra.Command, args []string) {
		listAbilities()
	},
}

// abilitySearchCmd 能力搜索命令
var abilitySearchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "搜索能力",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		abilityKeyword = args[0]
		searchAbilities()
	},
}

// abilityViewCmd 能力详情命令
var abilityViewCmd = &cobra.Command{
	Use:   "view [ability-id]",
	Short: "查看能力详情",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viewAbility(args[0])
	},
}

// abilityOrderCmd 订购能力命令
var abilityOrderCmd = &cobra.Command{
	Use:   "order [ability-id]",
	Short: "订购能力",
	Long: `订购指定的对外服务能力。

订购成功后，需要到个人中心进行授权（授权需要审批）。

示例:
  sdp ability order CA202303051941169651007662311817`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		orderAbility(args[0])
	},
}

// abilityMyCmd 我的能力命令
var abilityMyCmd = &cobra.Command{
	Use:   "my",
	Short: "查看我的能力（已订购）",
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

	// 列表命令参数
	abilityListCmd.Flags().IntVar(&abilityPage, "page", 1, "页码")
	abilityListCmd.Flags().IntVarP(&abilitySize, "size", "s", 20, "每页条数")

	// 搜索命令参数
	abilitySearchCmd.Flags().IntVar(&abilityPage, "page", 1, "页码")
	abilitySearchCmd.Flags().IntVarP(&abilitySize, "size", "s", 20, "每页条数")

	// 我的能力命令参数
	abilityMyCmd.Flags().IntVar(&abilityPage, "page", 1, "页码")
	abilityMyCmd.Flags().IntVarP(&abilitySize, "size", "s", 20, "每页条数")
}

// listAbilities 查询能力列表
func listAbilities() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}

		service := api.NewAbilityService(ctx.Client)
		req := &api.AbilityListRequest{
			PageNum:  abilityPage,
			PageSize: abilitySize,
		}

		resp, err := service.List(req)
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

// searchAbilities 搜索能力
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

// viewAbility 查看能力详情
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

// orderAbility 订购能力
func orderAbility(abilityID string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}

		service := api.NewAbilityService(ctx.Client)
		resp, err := service.Order(abilityID)
		if err != nil {
			return nil, fmt.Errorf("订购失败: %w", err)
		}

		if !resp.Success && resp.Code != 200 {
			return nil, fmt.Errorf("订购失败 [%d]: %s", resp.Code, resp.Message)
		}

		return map[string]interface{}{
			"message":   "订购成功",
			"abilityId": abilityID,
			"nextStep":  "请执行 'sdp app auth-ability <应用名>' 进行授权（需要审批）",
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

// listMyAbilities 查看我的能力
func listMyAbilities() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}

		service := api.NewAbilityService(ctx.Client)
		resp, err := service.GetMyAbilities(abilityPage, abilitySize)
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

// formatAbilityList 格式化能力列表
func formatAbilityList(resp *api.AbilityListResponse) map[string]interface{} {
	var items []map[string]interface{}
	for _, a := range resp.Data.List {
		items = append(items, map[string]interface{}{
			"id":          a.ID,
			"name":        a.Name,
			"code":        a.Code,
			"category":    a.Category,
			"provider":    a.Provider,
			"description": a.Description,
			"status":      a.Status,
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

// formatAbilityDetail 格式化能力详情
func formatAbilityDetail(detail *api.AbilityDetail) map[string]interface{} {
	result := map[string]interface{}{
		"id":          detail.ID,
		"name":        detail.Name,
		"code":        detail.Code,
		"category":    detail.Category,
		"provider":    detail.Provider,
		"description": detail.Description,
		"detailDesc":  detail.DetailDesc,
		"status":      detail.Status,
		"createTime":  detail.CreateTime,
		"updateTime":  detail.UpdateTime,
	}

	if detail.APIInfo != nil {
		result["apiInfo"] = map[string]interface{}{
			"method":         detail.APIInfo.Method,
			"url":            detail.APIInfo.URL,
			"requestFormat":  detail.APIInfo.RequestFormat,
			"responseFormat": detail.APIInfo.ResponseFormat,
		}
	}

	return result
}
