package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/common"
	lcmcp "github.com/user/lc/internal/mcp"
)

func init() {
	// 在 init 中注册所有预定义的 MCP 工具命令
	registerToolCommands()
}

// registerToolCommands 遍历预定义工具表，为每个工具生成 cobra.Command
func registerToolCommands() {
	for _, tool := range lcmcp.PredefinedTools {
		toolDef := tool // 捕获循环变量
		cmd := &cobra.Command{
			Use:   toolDef.Command,
			Short: toolDef.Description,
			Long:  fmt.Sprintf("%s\n\n参数格式: key=value\n必需参数: %v", toolDef.Description, toolDef.RequiredArgs),
			Args:  cobra.MinimumNArgs(0),
			Run: func(cmd *cobra.Command, args []string) {
				runToolCommand(toolDef, args)
			},
		}
		rootCmd.AddCommand(cmd)
	}
}

// runToolCommand 执行工具命令的统一逻辑
func runToolCommand(toolDef lcmcp.ToolDefinition, args []string) {
	ctx := context.Background()

	// 1. 解析 key=value 参数
	params, err := lcmcp.ParseKVArgs(args)
	if err != nil {
		printToolError("PARAM_INVALID", err.Error(), nil)
		os.Exit(1)
	}

	// 2. 校验必需参数
	var missingArgs []string
	for _, required := range toolDef.RequiredArgs {
		if _, ok := params[required]; !ok {
			missingArgs = append(missingArgs, required)
		}
	}
	if len(missingArgs) > 0 {
		printToolError("MISSING_ARG", fmt.Sprintf("缺少必需参数: %v", missingArgs), map[string]any{
			"command":  toolDef.Command,
			"required": toolDef.RequiredArgs,
			"missing":  missingArgs,
		})
		os.Exit(1)
	}

	// 3. 加载 MCP 配置
	cfg, _, err := lcmcp.LoadConfig()
	if err != nil {
		// 配置加载失败，但命令本身存在，返回 Server 未配置错误
		printToolError("MCP_SERVER_NOT_CONFIGURED", fmt.Sprintf("命令 '%s' 需要 MCP Server '%s'，但未找到相关配置", toolDef.Command, toolDef.Server), map[string]any{
			"server":     toolDef.Server,
			"command":    toolDef.Command,
			"suggestion": fmt.Sprintf("请配置 MCP Server '%s' 后重试，参考: lc mcp --help", toolDef.Server),
		})
		os.Exit(1)
	}

	// 4. 检查 Server 是否配置
	if !cfg.HasServer(toolDef.Server) {
		printToolError("MCP_SERVER_NOT_CONFIGURED", fmt.Sprintf("命令 '%s' 需要 MCP Server '%s'，但未找到相关配置", toolDef.Command, toolDef.Server), map[string]any{
			"server":     toolDef.Server,
			"command":    toolDef.Command,
			"suggestion": fmt.Sprintf("请配置 MCP Server '%s' 后重试，参考: lc mcp --help", toolDef.Server),
		})
		os.Exit(1)
	}

	// 5. 创建 Dispatcher 并执行调用
	dispatcher := lcmcp.NewDispatcher(cfg, nil)
	actualServer, result, err := dispatcher.CallTool(ctx, toolDef.Method, toolDef.Server, params)
	if err != nil {
		var mcpErr *lcmcp.MCPError
		if lcmcpErr, ok := err.(*lcmcp.MCPError); ok {
			mcpErr = lcmcpErr
		}
		if mcpErr != nil {
			printToolError(mcpErr.Code, mcpErr.Message, mcpErr.Details)
		} else {
			printToolError("TOOL_CALL_FAILED", err.Error(), nil)
		}
		os.Exit(1)
	}

	// 6. 输出结果
	printToolSuccess(map[string]any{
		"server": actualServer,
		"method": toolDef.Method,
		"result": result,
	})
}

// printToolSuccess 输出成功的 JSON 结果
func printToolSuccess(data any) {
	common.PrintJSON(map[string]any{
		"success": true,
		"data":    data,
		"meta": map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   common.GetVersion(),
		},
	})
}

// printToolError 输出错误的 JSON 结果
func printToolError(code, message string, details map[string]any) {
	errObj := map[string]any{
		"code":    code,
		"message": message,
	}
	if len(details) > 0 {
		errObj["details"] = details
	}
	common.PrintJSON(map[string]any{
		"success": false,
		"error":   errObj,
		"meta": map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   common.GetVersion(),
		},
	})
}
