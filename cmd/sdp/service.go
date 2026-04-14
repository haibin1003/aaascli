package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/haibin1003/aaascli/internal/api"
	"github.com/haibin1003/aaascli/internal/common"
)

var (
	serviceKeyword string
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "数字服务管理",
	Long:  "查询和管理山东能力开放平台的数字服务（API）目录",
}

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "查询全量服务列表",
	Run: func(cmd *cobra.Command, args []string) {
		listServices()
	},
}

var serviceSearchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "搜索服务",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceKeyword = args[0]
		searchServices()
	},
}

var serviceViewCmd = &cobra.Command{
	Use:   "view [service-id]",
	Short: "查看服务详情",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viewService(args[0])
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceListCmd)
	serviceCmd.AddCommand(serviceSearchCmd)
	serviceCmd.AddCommand(serviceViewCmd)

	serviceListCmd.Flags().IntVar(&abilityPage, "page", 1, "页码")
	serviceListCmd.Flags().IntVarP(&abilitySize, "size", "s", 20, "每页条数")

	serviceSearchCmd.Flags().IntVar(&abilityPage, "page", 1, "页码")
	serviceSearchCmd.Flags().IntVarP(&abilitySize, "size", "s", 20, "每页条数")
}

func listServices() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		svc := api.NewServiceService(ctx.Client)
		resp, err := svc.ListAll()
		if err != nil {
			return nil, fmt.Errorf("查询服务列表失败: %w", err)
		}
		items := flattenCatalogNodes(resp.Data.CataLogList, 0)
		return formatServiceSlice(items, abilityPage, abilitySize), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func searchServices() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		svc := api.NewServiceService(ctx.Client)
		resp, err := svc.ListAll()
		if err != nil {
			return nil, fmt.Errorf("搜索服务失败: %w", err)
		}
		all := flattenCatalogNodes(resp.Data.CataLogList, 0)
		var results []map[string]interface{}
		for _, item := range all {
			name, _ := item["name"].(string)
			code, _ := item["code"].(string)
			if serviceContainsCI(name, serviceKeyword) || serviceContainsCI(code, serviceKeyword) {
				results = append(results, item)
			}
		}
		return formatServiceSlice(results, abilityPage, abilitySize), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func viewService(serviceID string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		if err := ctx.CheckLoggedIn(); err != nil {
			return nil, err
		}
		svc := api.NewServiceService(ctx.Client)
		detail, err := svc.GetDetail(serviceID)
		if err != nil {
			return nil, fmt.Errorf("查询服务详情失败: %w", err)
		}
		return formatServiceDetail(detail), nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func formatServiceDetail(detail *api.ServiceDetail) map[string]interface{} {
	return map[string]interface{}{
		"id":              detail.APIID,
		"name":            detail.Name,
		"version":         detail.APIVersion,
		"requestType":     detail.RequestType,
		"requestTypeText": detail.RequestTypeText,
		"requestUrl":      detail.RequestURL,
		"remark":          detail.Remark,
		"requestExample":  detail.RequestExample,
		"responseExample": detail.ResponseExample,
		"protocol":        detail.Protocol,
		"interfaceId":     detail.InterfaceID,
		"serviceId":       detail.ServiceID,
		"domainName":      detail.DomainName,
		"owner":           detail.Owner,
		"department":      detail.Department,
		"contactNo":       detail.ContactNo,
	}
}

func flattenCatalogNodes(nodes []api.CatalogNode, depth int) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)
	for _, node := range nodes {
		// 目录节点
		nodeType := "catalog"
		if node.IsLeaf == "true" {
			nodeType = "leaf-catalog"
		}
		results = append(results, map[string]interface{}{
			"id":          node.CatalogID,
			"name":        node.CatalogName,
			"code":        "",
			"type":        nodeType,
			"catalogType": node.CatalogType,
			"catalogLevel": node.CatalogLevel,
			"depth":       depth,
		})
		// 递归子目录
		results = append(results, flattenCatalogNodes(node.SmallCatalogList, depth+1)...)
		// 展开 API 列表
		for _, apiSvc := range node.APIList {
			results = append(results, map[string]interface{}{
				"id":           apiSvc.APIID,
				"name":         apiSvc.Name,
				"code":         apiSvc.InterfaceID,
				"type":         "api",
				"requestType":  apiSvc.RequestType,
				"requestUrl":   apiSvc.RequestURL,
				"status":       apiSvc.Status,
				"version":      apiSvc.Version,
				"depth":        depth + 1,
			})
		}
	}
	return results
}

func serviceContainsCI(a, b string) bool {
	if len(a) < len(b) {
		return false
	}
	for i := 0; i <= len(a)-len(b); i++ {
		if a[i:i+len(b)] == b {
			return true
		}
	}
	return false
}

func formatServiceSlice(items []map[string]interface{}, page, size int) map[string]interface{} {
	if items == nil {
		items = []map[string]interface{}{}
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

	return map[string]interface{}{
		"items": pageItems,
		"pagination": map[string]interface{}{
			"page":  page,
			"size":  size,
			"total": total,
			"pages": (total + size - 1) / size,
		},
	}
}
