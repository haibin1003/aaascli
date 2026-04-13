# MCP-001 lc mcp 命令设计文档

## 变更记录表

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|---------|------|
| v1.0 | 2026-03-22 | 初始版本 | AI Assistant |
| v2.0 | 2026-03-22 | 将 `server` 固定子命令改为可变 serverName 位置参数；新增 1 参数模式；新增 dispatcher.GetServerInfo / GetToolInfo 方法；移除 `--server` flag | AI Assistant |
| v3.0 | 2026-03-22 | 配置加载策略改为扫描全部 5 个路径并合并，同名 Server 高优先级覆盖低优先级；输出字段 `configFile` → `configFiles`（[]string） | AI Assistant |
| v4.0 | 2026-03-22 | 新增 Windows 平台搜索路径；新增 `config_paths.go`，通过 `runtime.GOOS` 运行时判断选择平台路径；移除 `expandHome` 函数 | AI Assistant |
| v5.0 | 2026-03-23 | 新增三种传输方式支持（SSE/Streamable HTTP/stdio）；新增 `buildTransport()` 和 `InferTransportType()`；更新 `ServerConfig` 和 `ServerInfo` 结构 | AI Assistant |
| v6.0 | 2026-03-23 | 新增参数类型标注格式 `key:type=value`；工具详情输出改为人类可读格式（`参数格式`、`参数示例`、`调用示例`） | AI Assistant |
| v7.0 | 2026-03-23 | `lc mcp`（0 参数）改为只列出服务器配置，不获取 tools 列表；新增 `ListServersConfig()` 方法 | AI Assistant |
| v8.0 | 2026-03-23 | 工具详情输出字段国际化：`param_format`, `param_example`, `call_example`；新增 `required` 字段显示必填参数 | AI Assistant |
| v7.0 | 2026-03-23 | `lc mcp`（0 参数）改为只列出服务器配置，不获取 tools 列表；新增 `ListServersConfig()` 方法 | AI Assistant |

---

## 1. 架构设计

### 1.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                       CLI 命令层                             │
│                                                             │
│   lc mcp [serverName] [method] [key=val ...]                │
│      ├── [0 args]   → 列出模式（list mode）                  │
│      ├── [1 arg]    → Server 信息模式（server-info mode）    │
│      ├── [2 args]   → 工具详情模式（tool-info mode）         │
│      └── [3+ args]  → 调用模式（call mode）                  │
│                                                             │
│   cobra 框架 / cmd/lc/mcp.go                                │
└─────────────────────────────────┬───────────────────────────┘
                                  │
┌─────────────────────────────────┼───────────────────────────┐
│                      业务逻辑层  │                           │
│                                 │                           │
│   ┌──────────────────┐   ┌──────┴──────────────────┐        │
│   │  ConfigLoader    │   │      MCPDispatcher        │        │
│   │  • SearchPaths() │   │  • ListServersConfig()    │        │
│   │  • LoadConfig()  │   │  • ListAllServers()       │        │
│   │  • LoadConfig()  │   │  • GetServerInfo()        │        │
│   └──────────────────┘   │  • GetToolInfo()          │        │
│   internal/mcp/config.go │  • CallTool()             │        │
│                           └───────────────────────────┘       │
│                           internal/mcp/dispatcher.go         │
└─────────────────────────────────┬───────────────────────────┘
                                  │
┌─────────────────────────────────┼───────────────────────────┐
│                      MCP 客户端层│                           │
│                                 │                           │
│   ┌──────────────────────────────┴──────────────────────┐   │
│   │              serverClient (per server)               │   │
│   │  使用 github.com/modelcontextprotocol/go-sdk          │   │
│   │                                                      │   │
│   │  newServerClient(name, cfg)                          │   │
│   │  client.listTools(ctx) → []*ToolInfo                 │   │
│   │  client.callTool(ctx, method, args) → any            │   │
│   └──────────────────────────────────────────────────────┘   │
│   internal/mcp/client.go                                    │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 文件结构

| 文件 | 职责 |
|------|------|
| `cmd/lc/mcp.go` | cobra 命令定义，参数解析，调用 dispatcher |
| `internal/mcp/config.go` | MCP 配置文件搜索逻辑与加载（平台无关部分） |
| `internal/mcp/config_paths.go` | 搜索路径构建，运行时通过 `runtime.GOOS` 选择 Linux/macOS 或 Windows 路径 |
| `internal/mcp/client.go` | 单个 MCP Server 的连接与调用封装 |
| `internal/mcp/dispatcher.go` | 多 Server 调度：列出、查找、调用 |
| `internal/mcp/types.go` | 共享数据结构与错误码定义 |

---

## 2. 命令设计

### 2.1 cobra 命令结构

```
rootCmd
└── mcpCmd   (lc mcp [serverName] [method] [key=val ...])
      Args: cobra.ArbitraryArgs
      Run:  runMCP
```

`mcpCmd` 是唯一入口，通过位置参数数量区分四种行为，**不再有 `server` 固定子命令**：

```go
var mcpCmd = &cobra.Command{
    Use:   "mcp [serverName] [method] [key=val ...]",
    Short: "与 MCP (Model Context Protocol) Server 交互",
    Args:  cobra.ArbitraryArgs,
    Run:   runMCP,
}
```

### 2.2 参数解析逻辑

```
args 长度  │ 模式            │ 行为
──────────┼─────────────────┼──────────────────────────────────────────
0          │ list            │ 列出所有 Server 配置信息（不获取工具列表）
1          │ server-info     │ 列出 args[0] Server 的工具列表
2          │ tool-info       │ args[0]=serverName，显示 args[1] 工具详情
3+         │ call            │ args[0]=serverName，args[1]=method，args[2:]=key=val 参数
```

### 2.3 参数解析函数

```go
// ParseKVArgs 将 "key=val" 或 "key:type=value" 格式的参数列表解析为 map
// 若某项不含 "=" 则返回 MCPError（错误码 MCP_PARAM_INVALID）
// 支持类型标注：key:string=val, key:number=123, key:bool=true
func ParseKVArgs(args []string) (map[string]any, error)
```

**参数格式说明：**

| 格式 | 示例 | 解析结果 | 说明 |
|------|------|----------|------|
| `key=value` | `name=tom` | `{"name": "tom"}` | 默认为字符串类型 |
| `key:string=value` | `name:string=tom` | `{"name": "tom"}` | 显式声明字符串类型 |
| `key:number=value` | `age:number=25` | `{"age": 25.0}` | 数字类型（float64） |
| `key:int=value` | `count:int=10` | `{"count": 10.0}` | 整数类型（内部转为 float64） |
| `key:float=value` | `price:float=19.99` | `{"price": 19.99}` | 浮点类型 |
| `key:bool=value` | `enabled:bool=true` | `{"enabled": true}` | 布尔类型（支持 true/false/1/0/yes/no） |

**类型推断优先级：**
1. 若显式指定 `:type`，使用指定类型转换
2. 若省略 `:type`，默认为字符串类型

**错误处理：**
- 格式错误（无 `=`）：返回 `MCPError{Code: ErrCodeParamInvalid}`
- 不支持的类型：返回 `MCPError{Code: ErrCodeParamInvalid}`
- 类型转换失败（如 `age:number=abc`）：返回 `MCPError{Code: ErrCodeParamInvalid}`

---

## 3. 配置加载设计

### 3.1 搜索路径与合并策略

**扫描全部路径**：`LoadConfig()` 遍历所有标准路径，将全部存在的文件解析并合并为一份 `MCPConfig`。找到第一个文件后**不停止**，继续扫描剩余路径。

**平台选择**：`config_paths.go` 中的 `init()` 在程序启动时调用 `buildSearchPaths()`，通过 `runtime.GOOS` 在运行时判断操作系统，选择对应路径构建函数：

```go
func buildSearchPaths() []string {
    if runtime.GOOS == "windows" {
        return buildWindowsSearchPaths()
    }
    return buildUnixSearchPaths()
}
```

**Linux/macOS 搜索路径（优先级由高到低）：**

```
1. ~/.config/modelcontextprotocol/mcp.json
2. ~/.config/mcp/config.json
3. ./mcp.json                    （当前目录）
4. ./.mcp/config.json            （当前目录）
5. /etc/mcp/config.json          （系统级）
```

**Windows 搜索路径（优先级由高到低）：**

```
1. %APPDATA%\modelcontextprotocol\mcp.json
2. %APPDATA%\mcp\config.json
3. %USERPROFILE%\.mcp\config.json
4. .\mcp.json                    （当前目录）
5. .\.mcp\config.json            （当前目录）
6. %ProgramData%\mcp\config.json （系统级）
```

**合并规则**：同名 Server 以优先级高（序号小）的文件为准，低优先级文件的同名 Server 被忽略。

```
文件 1（优先级高）定义 serverA → 使用文件 1 的配置
文件 2（优先级低）也定义 serverA → 忽略
文件 2（优先级低）定义 serverB → 合并进来，因为 serverB 在文件 1 中不存在
```

**返回值**：`LoadConfig()` 返回 `(cfg *MCPConfig, loadedPaths []string, err error)`，其中 `loadedPaths` 为实际加载的所有文件路径列表（按优先级顺序）。路径在 `config_paths.go` 的 `buildSearchPaths()` 中已完全展开，`LoadConfig` 直接使用，无需额外处理。

### 3.2 配置文件结构

```go
// MCPConfig 对应配置文件根节点
type MCPConfig struct {
    MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// ServerConfig 单个 MCP Server 的配置
type ServerConfig struct {
    Transport string            `json:"transport,omitempty"` // sse / streamable / stdio
    Type      string            `json:"type,omitempty"`      // transport 的别名（如 streamable-http）
    URL       string            `json:"url,omitempty"`       // HTTP 传输端点
    Command   string            `json:"command,omitempty"`   // stdio 可执行文件
    Args      []string          `json:"args,omitempty"`      // stdio 参数
    Env       map[string]string `json:"env,omitempty"`       // stdio 环境变量
    Timeout   int               `json:"timeout,omitempty"`   // 毫秒，缺省 30000
}
```

---

## 4. MCP 客户端设计

### 4.1 核心封装（internal/mcp/client.go）

```go
type serverClient struct {
    name   string
    config ServerConfig
}

// listTools 连接 Server，返回其所有工具信息
func (c *serverClient) listTools(ctx context.Context) ([]*ToolInfo, error)

// callTool 调用指定工具，返回结果
func (c *serverClient) callTool(ctx context.Context, method string, args map[string]any) (any, error)
```

传输层支持三种方式，由 `buildTransport()` 根据配置自动选择：

1. **SSE** (`SSEClientTransport`)：传统 HTTP/SSE 传输，对应 `transport: "sse"` 或 URL 含 `sse`
2. **Streamable HTTP** (`StreamableClientTransport`)：现代流式 HTTP 传输，对应 `transport: "streamable"` 或 URL 含 `stream`（默认）
3. **stdio** (`CommandTransport`)：本地子进程传输，对应配置了 `command` 字段

传输类型推断逻辑在 `InferTransportType(cfg ServerConfig)` 中实现，优先级：
1. 显式 `transport` 字段
2. 显式 `type` 字段（如 `streamable-http` → `streamable`）
3. 配置了 `command` → `stdio`
4. URL 含 `sse` → `sse`
5. URL 含 `stream` → `streamable`
6. 默认 → `streamable`

超时由 `context.WithTimeout` 控制。

---

## 5. Dispatcher 设计

### 5.1 列出模式（0 args）

```
for each server in config:
    直接读取配置，构建 ServerInfo（name, transport, url/command）
    不连接服务器，不获取 tools 列表

输出 JSON（configFiles[] + servers[]）
```

对应方法：`Dispatcher.ListServersConfig()`（只读取配置，不发起网络连接）

> 注：此模式仅读取配置文件，不连接服务器获取工具列表，响应速度快。如需查看工具列表，请使用 `lc mcp <serverName>`。

### 5.2 Server 信息模式（1 arg: serverName）

```
if serverName 不在 config 中:
    返回 MCP_SERVER_NOT_FOUND 错误

尝试连接 serverName，调用 listTools()
输出 JSON（configFiles[] + server{name, transport, url/command, tools[]}）
```

对应方法：`Dispatcher.GetServerInfo(ctx, serverName)`

`ServerInfo` 结构根据 transport 类型填充不同字段：
- HTTP 传输（sse/streamable）：填充 `url`
- stdio 传输：填充 `command`

### 5.3 工具详情模式（2 args: serverName + method）

```
if serverName 不在 config 中:
    返回 MCP_SERVER_NOT_FOUND 错误

尝试连接 serverName，调用 listTools()
在工具列表中查找 name == method 的项

if 找到:
    使用 FormatInputSchema() 将 inputSchema 转换为人类可读格式
    输出 JSON（server + tool{name, description, 参数格式, 参数示例, 调用示例}）
else:
    返回 MCP_METHOD_NOT_FOUND 错误
```

对应方法：`Dispatcher.GetToolInfo(ctx, serverName, toolName)`

**参数格式化说明：**

`FormatInputSchema(schema any) []string` 将 JSON Schema 转换为 `key:type=<值>` 格式：
- 根据 `properties` 中的 `type` 字段映射为类型标注（`string` → `string`, `number` → `number`, `integer` → `int`, `boolean` → `bool`）
- 根据 `required` 数组标注必填项
- 保留 `description` 作为注释

示例转换：
```json
{
  "type": "object",
  "properties": {
    "limit": { "type": "number", "description": "返回数量限制" },
    "offset": { "type": "number", "description": "分页偏移量" }
  },
  "required": ["limit"]
}
```
↓ 转换为 ↓
```
[
  "limit:number=<值> (必填) // 返回数量限制",
  "offset:number=<值> // 分页偏移量"
]
```

### 5.4 调用模式（3+ args: serverName + method + key=val...）

```
1. 解析 key=val 参数为 map[string]any（ParseKVArgs）
2. 直接通过 serverName 取 config（fast path）
3. 调用 serverClient.callTool(ctx, method, args)
4. 输出 JSON（server + method + result）
```

对应方法：`Dispatcher.CallTool(ctx, method, serverName, args)`（serverName 为 preferServer，走 fast path）

---

## 6. 输出结构设计

### 6.1 列出模式输出（0 args）

```json
{
  "success": true,
  "data": {
    "configFiles": [
      "/home/user/.config/mcp/config.json",
      "/etc/mcp/config.json"
    ],
    "servers": [
      {
        "name": "openDeepWiki",
        "transport": "sse",
        "url": "https://opendeepwiki.k8m.site/mcp/sse"
      }
    ]
  },
  "meta": { "timestamp": "...", "version": "..." }
}
```

> 注：`configFiles` 为数组，列出本次实际加载的所有配置文件路径（按优先级顺序）。
> 此模式仅读取配置文件，不连接服务器获取工具列表，响应速度快。

stdio 类型 Server 输出示例（含 `command` 而非 `url`）：

```json
{
  "name": "local-server",
  "transport": "stdio",
  "command": "npx"
}
```

### 6.2 Server 信息模式输出（1 arg）

```json
{
  "success": true,
  "data": {
    "configFiles": ["/home/user/.config/mcp/config.json"],
    "server": {
      "name": "openDeepWiki",
      "url": "https://opendeepwiki.k8m.site/mcp/sse",
      "tools": ["list_repositories", "read_document", "search_documents", "get_repository", "get_document_summary"]
    }
  },
  "meta": { "timestamp": "...", "version": "..." }
}
```

### 6.3 工具详情模式输出（2 args）

```json
{
  "success": true,
  "data": {
    "server": "openDeepWiki",
    "tool": {
      "name": "get_repository",
      "description": "获取仓库详情，包含该仓库下的所有文档列表。优先使用 repo_id 查询；不知道 repo_id 时可使用 repo_name（二选一）。",
      "required": "repo_id OR repo_name",
      "param_format": "key:type=value (type: string/number/bool)",
      "param_example": [
        "repo_id:number={value} // 仓库ID，数字类型，优先使用",
        "repo_name:string={value} // 仓库名称，字符串类型，无 repo_id 时使用",
        "include_content:bool={value} // 是否返回文档完整内容，默认 false（仅返回文档列表）"
      ],
      "call_example": "lc mcp openDeepWiki get_repository repo_id:number={value} repo_name:string={value} include_content:bool={value}"
    }
  },
  "meta": { "timestamp": "...", "version": "..." }
}
```

> 注：工具详情输出使用人类可读的 `key:type=value` 格式替代原始 JSON Schema，便于 AI 直接理解和使用。
> - `required`: 必填参数名，多个时用 "OR" 连接（如 `"repo_id OR repo_name"`）
> - `param_format`: 参数格式说明
> - `param_example`: 参数列表示例，包含类型和描述注释
> - `call_example`: 完整的命令调用示例，包含所有参数

### 6.4 调用模式输出（3+ args）

```json
{
  "success": true,
  "data": {
    "server": "openDeepWiki",
    "method": "list_repositories",
    "result": "{ \"repositories\": [...], \"total\": 58 }"
  },
  "meta": { "timestamp": "...", "version": "..." }
}
```

> 注：`result` 字段为工具返回的原始字符串（通常是 JSON 文本），由调用方按需解析。

### 6.5 错误输出（统一格式）

```json
{
  "success": false,
  "error": {
    "code": "MCP_SERVER_NOT_FOUND",
    "message": "配置中不存在 Server 'badServer'，请使用 `lc mcp` 查看可用 Server"
  },
  "meta": { "timestamp": "...", "version": "..." }
}
```

---

## 7. 错误处理设计

所有错误通过统一的 `MCPError` 结构表达，经 `printMCPError()` 输出 JSON，符合项目规范：

```go
type MCPError struct {
    Code    string
    Message string
    Details map[string]any
}
```

| 错误码 | 触发条件 | details 字段 |
|--------|----------|-------------|
| `MCP_CONFIG_NOT_FOUND` | 五个路径均无配置文件 | `searchPaths: [...]` |
| `MCP_SERVER_NOT_FOUND` | serverName 不在配置中 | 无 |
| `MCP_CONNECT_FAILED` | 连接 MCP Server 失败 | `transport`, `url`/`command`, `hint` |
| `MCP_METHOD_NOT_FOUND` | Server 上无该工具 | 无 |
| `MCP_METHOD_AMBIGUOUS` | 多个 Server 有同名工具（FindTool 路径） | `matches: [...]` |
| `MCP_CALL_FAILED` | 工具调用返回错误 | 无 |
| `MCP_PARAM_INVALID` | 参数非 key=val 格式 | 无 |

连接失败错误消息示例：
```
连接 MCP Server 'xxx' 失败 (使用 streamable 传输): dial tcp: lookup ...

请检查您的配置是否正确。以下是各传输类型的配置示例：

1. SSE (传统协议):
   { "url": "http://host/mcp/sse" }

2. Streamable HTTP (现代协议，默认):
   { "url": "http://host/mcp" } 或 { "type": "streamable-http", ... }

3. stdio (本地命令):
   { "command": "npx", "args": ["-y", "@server/mcp"] }
```

---

## 8. 依赖

| 库 | 用途 | 获取方式 |
|----|------|---------|
| `github.com/modelcontextprotocol/go-sdk` | MCP 客户端（SSE/Streamable/stdio 传输） | `go get github.com/modelcontextprotocol/go-sdk@latest` |
| `github.com/spf13/cobra` | CLI 框架 | 已有依赖 |
| 标准库 `encoding/json`, `os`, `os/exec`, `path/filepath`, `runtime`, `strings` | 配置解析、路径处理、平台检测、子进程管理 | 内置 |

---

## 9. 开发顺序

1. `internal/mcp/types.go` — 定义共享类型与错误码，实现 `InferTransportType()` 和 `normalizeTransportType()`
2. `internal/mcp/config.go` — 配置加载核心逻辑（平台无关）
3. `internal/mcp/config_paths.go` — 搜索路径构建，运行时通过 `runtime.GOOS` 选择平台
4. `internal/mcp/client.go` — 单 Server 客户端，实现 `buildTransport()` 支持三种传输方式
5. `internal/mcp/dispatcher.go` — 多 Server 调度逻辑，实现 `buildServerInfo()` 根据 transport 类型填充不同字段
6. `cmd/lc/mcp.go` — cobra 命令注册，胶水代码
7. 单元测试 + E2E 测试
