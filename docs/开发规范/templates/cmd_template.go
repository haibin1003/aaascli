// 命令文件模板
// 使用说明：
// 1. 复制此文件到 cmd/lc/xxx.go
// 2. 将所有 "xxx" 替换为你的命令名
// 3. 所有 "Xxx" 替换为导出名（首字母大写）
// 4. 填写实际的 API 调用逻辑
// 5. 删除本注释

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
)

// 包级变量 - 用于存储标志值
var (
	xxxWorkspaceKey string
	xxxLimit        int
	// 在此添加更多标志变量
)

// xxxCmd 父命令定义
var xxxCmd = &cobra.Command{
	Use:   "xxx",
	Short: "管理xxx",
	Long:  `创建、查询、管理xxx。`,
}

// xxxListCmd 子命令：列表查询
var xxxListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出xxx",
	Long:  `列出指定研发空间的所有xxx。`,
	Run:   runXxxList,
}

// xxxCreateCmd 子命令：创建
var xxxCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "创建xxx",
	Long:  `创建新的xxx。`,
	Args:  cobra.ExactArgs(1), // 需要1个位置参数
	Run:   runXxxCreate,
}

// xxxViewCmd 子命令：查看详情
var xxxViewCmd = &cobra.Command{
	Use:   "view \u003cid\u003e",
	Short: "查看xxx详情",
	Long:  `查看指定xxx的详细信息。`,
	Args:  cobra.ExactArgs(1),
	Run:   runXxxView,
}

// xxxDeleteCmd 子命令：删除
var xxxDeleteCmd = &cobra.Command{
	Use:   "delete \u003cid\u003e",
	Short: "删除xxx",
	Long:  `删除指定的xxx。`,
	Args:  cobra.ExactArgs(1),
	Run:   runXxxDelete,
}

// init 函数 - 注册命令和标志
func init() {
	// 添加子命令
	xxxCmd.AddCommand(xxxListCmd)
	xxxCmd.AddCommand(xxxCreateCmd)
	xxxCmd.AddCommand(xxxViewCmd)
	xxxCmd.AddCommand(xxxDeleteCmd)

	// list 命令的标志
	xxxListCmd.Flags().StringVarP(&xxxWorkspaceKey, "workspace-key", "w", "", "研发空间 Key")
	xxxListCmd.Flags().IntVarP(&xxxLimit, "limit", "l", 10, "返回数量限制")

	// create 命令的标志
	xxxCreateCmd.Flags().StringVarP(&xxxWorkspaceKey, "workspace-key", "w", "", "研发空间 Key")
	xxxCreateCmd.Flags().String("description", "", "描述")
	// xxxCreateCmd.Flags().StringP("project-code", "p", "", "项目代码") // 如需要

	// view 命令的标志
	xxxViewCmd.Flags().StringVarP(&xxxWorkspaceKey, "workspace-key", "w", "", "研发空间 Key")

	// delete 命令的标志
	xxxDeleteCmd.Flags().StringVarP(&xxxWorkspaceKey, "workspace-key", "w", "", "研发空间 Key")

	// 注册到根命令
	rootCmd.AddCommand(xxxCmd)
}

// runXxxList 执行列表查询
func runXxxList(cmd *cobra.Command, args []string) {
	// 1. 获取显式参数
	workspaceKey := xxxWorkspaceKey
	limit := xxxLimit

	// 2. 尝试自动探测
	if workspaceKey == "" {
		autoResult := common.TryAutoDetect(true)
		if autoResult.Success {
			workspaceKey = autoResult.Context.WorkspaceKey
			if debugMode {
				common.PrintAutoDetectInfo(autoResult.Context, logger)
			}
		} else {
			common.PrintAutoDetectError(autoResult.Error)
			os.Exit(1)
		}
	}

	// 3. 执行命令
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 调用 API
		items, total, err := api.ListXxx(ctx, workspaceKey, limit)
		if err != nil {
			return nil, fmt.Errorf("查询列表失败: %w", err)
		}

		return map[string]interface{}{
			"items": items,
			"count": len(items),
			"total": total,
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		Logger:    &logger,
	})
}

// runXxxCreate 执行创建
func runXxxCreate(cmd *cobra.Command, args []string) {
	// 1. 获取位置参数
	name := args[0]

	// 2. 获取标志参数
	workspaceKey := xxxWorkspaceKey
	description, _ := cmd.Flags().GetString("description")

	// 3. 尝试自动探测
	if workspaceKey == "" {
		autoResult := common.TryAutoDetect(true)
		if autoResult.Success {
			workspaceKey = autoResult.Context.WorkspaceKey
		} else {
			common.PrintAutoDetectError(autoResult.Error)
			os.Exit(1)
		}
	}

	// 4. 执行命令
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 构建请求
		req := api.CreateXxxRequest{
			Name:         name,
			Description:  description,
			WorkspaceKey: workspaceKey,
		}

		// 调用 API
		result, err := api.CreateXxx(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("创建失败: %w", err)
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		Logger:    &logger,
	})
}

// runXxxView 执行查看详情
func runXxxView(cmd *cobra.Command, args []string) {
	// 1. 获取位置参数
	id := args[0]

	// 2. 获取标志参数
	workspaceKey := xxxWorkspaceKey

	// 3. 尝试自动探测
	if workspaceKey == "" {
		autoResult := common.TryAutoDetect(true)
		if autoResult.Success {
			workspaceKey = autoResult.Context.WorkspaceKey
		} else {
			common.PrintAutoDetectError(autoResult.Error)
			os.Exit(1)
		}
	}

	// 4. 执行命令
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 调用 API
		result, err := api.GetXxx(ctx, id, workspaceKey)
		if err != nil {
			return nil, fmt.Errorf("查询详情失败: %w", err)
		}

		return result, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		Logger:    &logger,
	})
}

// runXxxDelete 执行删除
func runXxxDelete(cmd *cobra.Command, args []string) {
	// 1. 获取位置参数
	id := args[0]

	// 2. 获取标志参数
	workspaceKey := xxxWorkspaceKey

	// 3. 尝试自动探测
	if workspaceKey == "" {
		autoResult := common.TryAutoDetect(true)
		if autoResult.Success {
			workspaceKey = autoResult.Context.WorkspaceKey
		} else {
			common.PrintAutoDetectError(autoResult.Error)
			os.Exit(1)
		}
	}

	// 4. 执行命令
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 调用 API
		err := api.DeleteXxx(ctx, id, workspaceKey)
		if err != nil {
			return nil, fmt.Errorf("删除失败: %w", err)
		}

		return map[string]string{
			"id":      id,
			"message": "删除成功",
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		Logger:    &logger,
	})
}
