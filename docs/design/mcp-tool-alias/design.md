# MCP 工具静态命令封装 - 设计文档

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                         cmd/lc/tool.go                       │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  init()                                              │  │
│  │   └── 遍历 PredefinedTools 生成 cobra.Command       │  │
│  └──────────────────────────────────────────────────────┘  │
│                            │                                │
│                            ▼                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  runTool(cmd, args, toolDef)                         │  │
│  │   ├── 解析 key=value 参数                           │  │
│  │   ├── 校验必需参数                                   │  │
│  │   ├── 检查 Server 配置（LoadConfig + HasServer）    │  │
│  │   └── 调用 Dispatcher.CallTool                      │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    internal/mcp/registry.go                  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  ToolDefinition 结构体                               │  │
│  │   - Command: 命令名（如 "get-repo-wiki"）            │  │
│  │   - Server: 依赖的 MCP Server                        │  │
│  │   - Method: MCP 方法名                               │  │
│  │   - RequiredArgs: 必需参数列表                       │  │
│  └──────────────────────────────────────────────────────┘  │
│                            │                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  PredefinedTools: []ToolDefinition                   │  │
│  │   预定义工具表（代码写死）                           │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 数据结构

### ToolDefinition

```go
type ToolDefinition struct {
    Command       string   // 命令名，如 "get-repo-wiki"
    Server        string   // 依赖的 MCP Server 名
    Method        string   // MCP 方法名
    Description   string   // 命令描述
    RequiredArgs  []string // 必需参数
    OptionalArgs  []string // 可选参数
}
```

### 预定义工具表

```go
var PredefinedTools = []ToolDefinition{
    {
        Command:      "get-repo-wiki",
        Server:       "openDeepWiki",
        Method:       "GetWikiContents",
        Description:  "获取代码仓库的 Wiki 文档内容",
        RequiredArgs: []string{"owner", "repo"},
    },
}
```

## 执行流程

```
用户执行: lc get-repo-wiki owner=torvalds repo=linux
    │
    ▼
┌────────────────┐
│ 解析参数       │───▶ 解析 key=value 格式
└────────────────┘
    │
    ▼
┌────────────────┐
│ 校验必需参数   │───▶ 检查 owner, repo 是否提供
└────────────────┘
    │
    ▼
┌────────────────┐
│ 加载 MCP 配置  │───▶ mcp.LoadConfig()
└────────────────┘
    │
    ▼
┌────────────────┐
│ 检查 Server    │───▶ cfg.HasServer("openDeepWiki")
│ 是否配置       │
└────────────────┘
    │
    ├─ 未配置 ──▶ 返回 MCP_SERVER_NOT_CONFIGURED 错误
    │
    ▼ 已配置
┌────────────────┐
│ 调用 MCP 工具  │───▶ dispatcher.CallTool(...)
└────────────────┘
    │
    ▼
┌────────────────┐
│ 返回 JSON 结果 │
└────────────────┘
```

## 错误处理

### MCP_SERVER_NOT_CONFIGURED

```json
{
  "success": false,
  "error": {
    "code": "MCP_SERVER_NOT_CONFIGURED",
    "message": "命令 'get-repo-wiki' 需要 MCP Server 'openDeepWiki'，但未找到相关配置",
    "details": {
      "server": "openDeepWiki",
      "command": "get-repo-wiki",
      "suggestion": "请配置 MCP Server 'openDeepWiki' 后重试，参考: lc mcp --help"
    }
  }
}
```

### MISSING_ARG

```json
{
  "success": false,
  "error": {
    "code": "MISSING_ARG",
    "message": "缺少必需参数: owner",
    "details": {
      "command": "get-repo-wiki",
      "required": ["owner", "repo"],
      "provided": ["repo"]
    }
  }
}
```

## 接口扩展（第二阶段）

预留用户自定义配置结构：

```go
// CustomToolConfig 用户自定义工具配置
type CustomToolConfig struct {
    Tools []ToolDefinition `json:"customTools"`
}

// 加载顺序：PredefinedTools → CustomTools（自定义覆盖预定义）
```

## 文件变更

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `internal/mcp/registry.go` | 新增 | 工具定义结构和预定义表 |
| `internal/mcp/config.go` | 修改 | 新增 `HasServer()` 方法 |
| `cmd/lc/tool.go` | 新增 | 工具命令实现 |
| `cmd/lc/root.go` | 修改 | 注册工具命令 |
