package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/common"
	lcmcp "github.com/user/lc/internal/mcp"
)

// mcpCmd is the top-level "lc mcp" command.
// It handles all four operating modes based on the number of positional arguments:
//
//	0 args                             – list all configured servers and their tools
//	1 arg  <serverName>                – list tools exposed by that specific server
//	2 args <serverName> <method>       – display description and input schema for the tool
//	3+ args <serverName> <method> kv… – call the tool with key=value parameters
var mcpCmd = &cobra.Command{
	Use:   "mcp [serverName] [method] [key=val ...]",
	Short: "与 MCP (Model Context Protocol) Server 交互",
	Long: `通过 MCP 协议调用外部工具服务。

MCP 配置文件按以下优先顺序自动搜索：
  1. <git-root>/joinai-code.json                        (joinai-code 项目级)
  2. <git-root>/.joinai-code/joinai-code.json           (joinai-code 项目级)
  3. ~/.config/joinai-code/joinai-code.json             (joinai-code 全局级)
  4. ~/.config/modelcontextprotocol/mcp.json            (MCP 标准路径)
  5. ~/.config/mcp/config.json
  6. ./mcp.json                                         (当前目录)
  7. ./.mcp/config.json
  8. /etc/mcp/config.json                               (系统级)

支持两种配置格式：

1. 标准 MCP 格式 (根键: mcpServers):
   {
     "mcpServers": {
       "myServer": {
         "url": "https://api.example.com/mcp"
       }
     }
   }

2. joinai-code 格式 (根键: mcp):
   {
     "mcp": {
       "local-server": {
         "type": "local",
         "command": ["npx", "-y", "@modelcontextprotocol/server-filesystem"],
         "environment": { "KEY": "value" }
       },
       "remote-service": {
         "type": "remote",
         "url": "https://api.example.com/mcp",
         "headers": {
           "Authorization": "Bearer {env:API_KEY}"
         }
       }
     }
   }

配置说明:
   • type: "local" 映射为 stdio, "remote" 映射为 streamable/sse
   • enabled: false 可禁用服务器（默认启用）
   • command: 支持字符串或数组格式
   • headers: 支持 {env:VAR_NAME} 占位符，从环境变量读取
   • oauth: {} 标记使用 OAuth 认证

用法示例：

  # 列出所有配置的 Server（仅显示配置信息，不获取工具列表，速度快）
  lc mcp

  # 列出指定 Server 的工具
  lc mcp openDeepWiki

  # 查看某个工具的详情（描述 + 参数 schema）
  lc mcp openDeepWiki GetWikiContents

  # 调用工具（参数格式为 key=value）
  lc mcp openDeepWiki GetWikiContents owner=torvalds repo=linux`,
	Args: cobra.ArbitraryArgs,
	Run:  runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

// runMCP is the entry point for all operating modes of "lc mcp".
func runMCP(_ *cobra.Command, args []string) {
	ctx := context.Background()

	// Load MCP config file (scans all paths and merges results).
	cfg, configPaths, err := lcmcp.LoadConfig()
	if err != nil {
		var mcpErr *lcmcp.MCPError
		if errors.As(err, &mcpErr) {
			printMCPError(mcpErr.Code, mcpErr.Message, mcpErr.Details)
		} else {
			printMCPError(lcmcp.ErrCodeConfigNotFound, err.Error(), nil)
		}
		os.Exit(1)
	}

	dispatcher := lcmcp.NewDispatcher(cfg, configPaths)

	switch len(args) {
	case 0:
		runMCPList(ctx, dispatcher)
	case 1:
		runMCPServerInfo(ctx, dispatcher, args[0])
	case 2:
		runMCPInfo(ctx, dispatcher, args[0], args[1])
	default:
		runMCPCall(ctx, dispatcher, args[0], args[1], args[2:])
	}
}

// runMCPList prints server configurations without connecting to fetch tools.
// This is fast and only shows server metadata. Use 'lc mcp <serverName>' to see tools.
func runMCPList(_ context.Context, d *lcmcp.Dispatcher) {
	servers := d.ListServersConfig()
	printMCPSuccess(map[string]any{
		"configFiles": d.ConfigPaths(),
		"servers":     servers,
	})
}

// runMCPServerInfo connects to the named server and prints its tool list.
func runMCPServerInfo(ctx context.Context, d *lcmcp.Dispatcher, serverName string) {
	info, err := d.GetServerInfo(ctx, serverName)
	if err != nil {
		var mcpErr *lcmcp.MCPError
		if errors.As(err, &mcpErr) {
			printMCPError(mcpErr.Code, mcpErr.Message, mcpErr.Details)
		} else {
			printMCPError(lcmcp.ErrCodeConnectFailed, err.Error(), nil)
		}
		os.Exit(1)
	}
	printMCPSuccess(map[string]any{
		"configFiles": d.ConfigPaths(),
		"server":      info,
	})
}

// runMCPInfo finds the named tool on the specified server and prints its description and input schema.
func runMCPInfo(ctx context.Context, d *lcmcp.Dispatcher, serverName, method string) {
	match, err := d.GetToolInfo(ctx, serverName, method)
	if err != nil {
		var mcpErr *lcmcp.MCPError
		if errors.As(err, &mcpErr) {
			printMCPError(mcpErr.Code, mcpErr.Message, mcpErr.Details)
		} else {
			printMCPError(lcmcp.ErrCodeMethodNotFound, err.Error(), nil)
		}
		os.Exit(1)
	}

	// Format input schema as key:type=value for easy AI usage
	formattedParams := lcmcp.FormatInputSchema(match.Tool.InputSchema)
	requiredParams := lcmcp.GetRequiredParams(match.Tool.InputSchema)
	paramInfoList := lcmcp.GetParamInfoList(match.Tool.InputSchema)

	toolData := map[string]any{
		"name":        match.Tool.Name,
		"description": match.Tool.Description,
	}

	// Add required field if there are required parameters
	if len(requiredParams) > 0 {
		toolData["required"] = strings.Join(requiredParams, " OR ")
	}

	if formattedParams != nil {
		toolData["param_format"] = "key:type=value (type: string/number/bool)"
		toolData["param_example"] = formattedParams
		toolData["call_example"] = fmt.Sprintf("lc mcp %s %s %s", match.ServerName, match.Tool.Name, buildCallExample(paramInfoList))
	} else {
		toolData["inputSchema"] = match.Tool.InputSchema
	}

	printMCPSuccess(map[string]any{
		"server": match.ServerName,
		"tool":   toolData,
	})
}

// buildCallExample builds a complete command example from parameter info list.
// It includes all parameters with their types in the format: key:type={value}
func buildCallExample(params []lcmcp.ParamInfo) string {
	if len(params) == 0 {
		return ""
	}

	var parts []string
	for _, p := range params {
		parts = append(parts, fmt.Sprintf("%s:%s={value}", p.Name, p.Type))
	}
	return strings.Join(parts, " ")
}

// runMCPCall parses key=value arguments and calls the named tool on the specified server.
func runMCPCall(ctx context.Context, d *lcmcp.Dispatcher, serverName, method string, kvArgs []string) {
	// Parse key=value argument list.
	params, err := lcmcp.ParseKVArgs(kvArgs)
	if err != nil {
		var mcpErr *lcmcp.MCPError
		if errors.As(err, &mcpErr) {
			printMCPError(mcpErr.Code, mcpErr.Message, mcpErr.Details)
		} else {
			printMCPError(lcmcp.ErrCodeParamInvalid, err.Error(), nil)
		}
		os.Exit(1)
	}

	actualServer, result, callErr := d.CallTool(ctx, method, serverName, params)
	if callErr != nil {
		var mcpErr *lcmcp.MCPError
		if errors.As(callErr, &mcpErr) {
			printMCPError(mcpErr.Code, mcpErr.Message, mcpErr.Details)
		} else {
			printMCPError(lcmcp.ErrCodeCallFailed, callErr.Error(), nil)
		}
		os.Exit(1)
	}

	printMCPSuccess(map[string]any{
		"server": actualServer,
		"method": method,
		"result": result,
	})
}

// printMCPSuccess writes a successful JSON result to stdout.
func printMCPSuccess(data any) {
	common.PrintJSON(map[string]any{
		"success": true,
		"data":    data,
		"meta": map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   common.GetVersion(),
		},
	})
}

// printMCPError writes a failure JSON result to stdout.
// details may be nil if there is no additional structured context.
func printMCPError(code, message string, details map[string]any) {
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
