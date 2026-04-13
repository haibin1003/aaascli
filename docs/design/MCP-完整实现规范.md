# MCP (Model Context Protocol) 完整实现规范

本文档包含 `lc mcp` 命令的完整实现细节，可供另一个 AI 根据此文档重新实现全部功能。

---

## 1. 整体架构

### 1.1 分层架构

```
┌─────────────────────────────────────────────────────────────┐
│  CLI 层 (cmd/lc/mcp.go)                                      │
│  - 参数解析 (0/1/2/3+ 参数模式)                              │
│  - 输出格式化 (JSON 统一输出)                                │
└───────────────────────────────┬─────────────────────────────┘
                                │
┌───────────────────────────────┼─────────────────────────────┐
│  调度层 (internal/mcp/dispatcher.go)                        │
│  - ListServersConfig()        │  只读配置，不连接服务器      │
│  - ListAllServers()           │  并发连接获取 tools          │
│  - GetServerInfo()            │  单服务器信息                │
│  - GetToolInfo()              │  工具详情                    │
│  - FindTool()                 │  跨服务器工具搜索            │
│  - CallTool()                 │  工具调用                    │
└───────────────────────────────┼─────────────────────────────┘
                                │
┌───────────────────────────────┼─────────────────────────────┐
│  客户端层 (internal/mcp/client.go)                          │
│  - buildTransport()           │  根据配置创建传输层          │
│  - listTools()                │  获取工具列表                │
│  - callTool()                 │  调用指定工具                │
└───────────────────────────────┼─────────────────────────────┘
                                │
┌───────────────────────────────┼─────────────────────────────┐
│  传输层 (go-sdk)                                            │
│  - SSEClientTransport         │  传统 HTTP/SSE               │
│  - StreamableClientTransport  │  现代流式 HTTP               │
│  - CommandTransport           │  本地子进程 (stdio)          │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 核心文件职责

| 文件 | 职责 |
|------|------|
| `cmd/lc/mcp.go` | cobra 命令定义，参数路由，输出格式化 |
| `internal/mcp/types.go` | 共享结构体、错误码、工具函数 |
| `internal/mcp/config.go` | 配置文件搜索与加载 |
| `internal/mcp/config_paths.go` | 平台相关路径构建 |
| `internal/mcp/client.go` | 单个 MCP 服务器客户端 |
| `internal/mcp/dispatcher.go` | 多服务器调度协调 |

---

## 2. 配置系统

### 2.1 配置文件搜索路径

**Linux/macOS 路径（优先级从高到低）：**
1. `~/.config/modelcontextprotocol/mcp.json`
2. `~/.config/mcp/config.json`
3. `./mcp.json`（当前目录）
4. `./.mcp/config.json`（当前目录）
5. `/etc/mcp/config.json`（系统级）

**Windows 路径：**
1. `%APPDATA%\modelcontextprotocol\mcp.json`
2. `%APPDATA%\mcp\config.json`
3. `%USERPROFILE%\.mcp\config.json`
4. `.\mcp.json`
5. `.\.mcp\config.json`
6. `%ProgramData%\mcp\config.json`

> 重要：需要扫描**所有**存在的配置文件并合并，同名 Server 高优先级覆盖低优先级。

### 2.2 配置文件结构

```json
{
  "mcpServers": {
    "serverName": {
      "transport": "sse|streamable|stdio",
      "type": "streamable-http",
      "url": "https://example.com/mcp",
      "command": "npx",
      "args": ["-y", "@server/mcp"],
      "env": {"KEY": "value"},
      "timeout": 30000
    }
  }
}
```

**字段说明：**
- `transport`: 显式指定传输类型，优先级最高
- `type`: transport 的别名（如 `streamable-http` → `streamable`）
- `url`: HTTP 传输端点（SSE/Streamable）
- `command`: stdio 传输的可执行文件
- `args`: stdio 传输的参数数组
- `env`: stdio 传输的环境变量
- `timeout`: 超时时间（毫秒），默认 30000

### 2.3 传输类型推断（关键逻辑）

```go
func InferTransportType(cfg ServerConfig) string {
    // 1. 显式 transport 字段
    if cfg.Transport != "" {
        return cfg.Transport
    }
    // 2. type 字段别名映射
    if cfg.Type != "" {
        return normalizeTransportType(cfg.Type)
    }
    // 3. 配置了 command → stdio
    if cfg.Command != "" {
        return "stdio"
    }
    // 4. URL 含 "sse"（不区分大小写）
    if strings.Contains(strings.ToLower(cfg.URL), "sse") {
        return "sse"
    }
    // 5. URL 含 "stream" → streamable
    if strings.Contains(strings.ToLower(cfg.URL), "stream") {
        return "streamable"
    }
    // 6. 默认 streamable
    return "streamable"
}
```

**类型别名映射：**
- 含 "stream" → `streamable`
- 含 "sse" → `sse`
- 含 "command" 或 "stdio" → `stdio`

### 2.4 配置合并策略

```go
// 伪代码
for _, path := range searchPaths {
    if fileExists(path) {
        cfg := loadConfigFile(path)
        for name, serverCfg := range cfg.MCPServers {
            if _, exists := result.MCPServers[name]; !exists {
                result.MCPServers[name] = serverCfg  // 高优先级优先
            }
        }
        loadedPaths = append(loadedPaths, path)
    }
}
```

---

## 3. 核心数据结构

### 3.1 MCPConfig

```go
type MCPConfig struct {
    MCPServers map[string]ServerConfig `json:"mcpServers"`
}
```

### 3.2 ServerConfig（配置解析用）

```go
type ServerConfig struct {
    Transport string            `json:"transport,omitempty"`
    Type      string            `json:"type,omitempty"`
    URL       string            `json:"url,omitempty"`
    Command   string            `json:"command,omitempty"`
    Args      []string          `json:"args,omitempty"`
    Env       map[string]string `json:"env,omitempty"`
    Timeout   int               `json:"timeout,omitempty"`
}
```

### 3.3 ServerInfo（输出用）

```go
type ServerInfo struct {
    Name      string   `json:"name"`
    Transport string   `json:"transport"`
    URL       string   `json:"url,omitempty"`     // sse/streamable 时设置
    Command   string   `json:"command,omitempty"` // stdio 时设置
    Tools     []string `json:"tools,omitempty"`   // 获取 tools 时设置，nil 时省略
    Error     string   `json:"error,omitempty"`   // 连接失败时设置
}
```

### 3.4 ToolInfo

```go
type ToolInfo struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    InputSchema any    `json:"inputSchema,omitempty"` // JSON Schema
}
```

### 3.5 ToolMatch

```go
type ToolMatch struct {
    ServerName string
    Tool       *ToolInfo
}
```

### 3.6 ParamInfo（工具详情格式化用）

```go
type ParamInfo struct {
    Name        string `json:"name"`
    Type        string `json:"type"`
    Required    bool   `json:"required"`
    Description string `json:"description,omitempty"`
}
```

### 3.7 MCPError

```go
type MCPError struct {
    Code    string
    Message string
    Details map[string]any
}

func (e *MCPError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 错误码常量
const (
    ErrCodeConfigNotFound  = "MCP_CONFIG_NOT_FOUND"
    ErrCodeConnectFailed   = "MCP_CONNECT_FAILED"
    ErrCodeServerNotFound  = "MCP_SERVER_NOT_FOUND"
    ErrCodeMethodNotFound  = "MCP_METHOD_NOT_FOUND"
    ErrCodeMethodAmbiguous = "MCP_METHOD_AMBIGUOUS"
    ErrCodeCallFailed      = "MCP_CALL_FAILED"
    ErrCodeParamInvalid    = "MCP_PARAM_INVALID"
)
```

---

## 4. 传输层实现

### 4.1 传输类型枚举

- `"sse"` - 传统 HTTP/SSE 传输
- `"streamable"` - 现代流式 HTTP 传输（默认）
- `"stdio"` - 本地子进程传输

### 4.2 构建传输层（关键代码）

```go
func buildTransport(cfg ServerConfig, timeout time.Duration) (mcp.ClientTransport, error) {
    transportType := InferTransportType(cfg)

    switch transportType {
    case "stdio":
        if cfg.Command == "" {
            return nil, errors.New("stdio transport requires command")
        }
        return &CommandTransport{
            Command: cfg.Command,
            Args:    cfg.Args,
            Env:     cfg.Env,
            Timeout: timeout,
        }, nil

    case "sse":
        url := cfg.URL
        if url == "" {
            return nil, errors.New("sse transport requires url")
        }
        return &SSEClientTransport{URL: url, Timeout: timeout}, nil

    case "streamable":
        url := cfg.URL
        if url == "" {
            return nil, errors.New("streamable transport requires url")
        }
        return &StreamableClientTransport{URL: url, Timeout: timeout}, nil

    default:
        return nil, fmt.Errorf("unknown transport: %s", transportType)
    }
}
```

---

## 5. 命令行接口

### 5.1 四种操作模式

```
args 数量 │ 模式            │ 处理函数
──────────┼─────────────────┼──────────────────────────────────
0          │ list            │ runMCPList()
1          │ server-info     │ runMCPServerInfo(serverName)
2          │ tool-info       │ runMCPInfo(serverName, toolName)
3+         │ call            │ runMCPCall(serverName, toolName, args...)
```

### 5.2 帮助文本

```
与 MCP (Model Context Protocol) Server 交互

配置文件搜索路径（按优先级）：
  1. ~/.config/modelcontextprotocol/mcp.json
  2. ~/.config/mcp/config.json
  3. ./mcp.json
  4. ./.mcp/config.json
  5. /etc/mcp/config.json

用法示例：

  # 列出所有配置的 Server（仅显示配置，不获取工具列表，速度快）
  lc mcp

  # 列出指定 Server 的工具
  lc mcp openDeepWiki

  # 查看某个工具的详情
  lc mcp openDeepWiki list_repositories

  # 调用工具（参数格式为 key=value 或 key:type=value）
  lc mcp openDeepWiki list_repositories limit=3
```

---

## 6. 各模式详细实现

### 6.1 0 参数模式（list）

**行为：** 只读取配置，不连接服务器，返回服务器基本信息

```go
func runMCPList(ctx context.Context, d *Dispatcher) {
    // 使用 ListServersConfig 而非 ListAllServers
    servers := d.ListServersConfig()
    printMCPSuccess(map[string]any{
        "configFiles": d.ConfigPaths(),
        "servers":     servers,
    })
}

// Dispatcher.ListServersConfig 实现
func (d *Dispatcher) ListServersConfig() []ServerInfo {
    servers := make([]ServerInfo, 0, len(d.config.MCPServers))
    for name, cfg := range d.config.MCPServers {
        servers = append(servers, buildServerInfo(name, cfg))
    }
    return servers
}
```

**输出示例：**
```json
{
  "success": true,
  "data": {
    "configFiles": ["/home/user/.config/mcp/config.json"],
    "servers": [
      {
        "name": "openDeepWiki",
        "transport": "sse",
        "url": "https://opendeepwiki.k8m.site/mcp/sse"
      },
      {
        "name": "local-stdio",
        "transport": "stdio",
        "command": "npx"
      }
    ]
  },
  "meta": {
    "timestamp": "2026-03-23T10:00:00Z",
    "version": "v0.2.8"
  }
}
```

### 6.2 1 参数模式（server-info）

**行为：** 连接指定服务器，获取并返回工具列表

```go
func runMCPServerInfo(ctx context.Context, d *Dispatcher, serverName string) {
    info, err := d.GetServerInfo(ctx, serverName)
    if err != nil {
        printMCPError(...)
        os.Exit(1)
    }
    printMCPSuccess(map[string]any{
        "configFiles": d.ConfigPaths(),
        "server":      info,
    })
}
```

**输出示例：**
```json
{
  "success": true,
  "data": {
    "configFiles": [...],
    "server": {
      "name": "openDeepWiki",
      "transport": "sse",
      "url": "https://...",
      "tools": ["list_repositories", "read_document", "search_documents"]
    }
  }
}
```

### 6.3 2 参数模式（tool-info）- 关键实现

**行为：** 获取指定工具的详细信息，格式化为人类可读格式

```go
func runMCPInfo(ctx context.Context, d *Dispatcher, serverName, toolName string) {
    match, err := d.GetToolInfo(ctx, serverName, toolName)
    if err != nil {
        printMCPError(...)
        os.Exit(1)
    }

    // 格式化工具信息
    formattedParams := FormatInputSchema(match.Tool.InputSchema)
    requiredParams := GetRequiredParams(match.Tool.InputSchema)
    paramInfoList := GetParamInfoList(match.Tool.InputSchema)

    toolData := map[string]any{
        "name":        match.Tool.Name,
        "description": match.Tool.Description,
    }

    // 添加 required 字段（如果有必填参数）
    if len(requiredParams) > 0 {
        toolData["required"] = strings.Join(requiredParams, " OR ")
    }

    if formattedParams != nil {
        toolData["param_format"] = "key:type=value (type: string/number/bool)"
        toolData["param_example"] = formattedParams
        toolData["call_example"] = fmt.Sprintf(
            "lc mcp %s %s %s",
            match.ServerName,
            match.Tool.Name,
            buildCallExample(paramInfoList)
        )
    } else {
        toolData["inputSchema"] = match.Tool.InputSchema
    }

    printMCPSuccess(map[string]any{
        "server": match.ServerName,
        "tool":   toolData,
    })
}
```

#### 6.3.1 FormatInputSchema 实现（关键）

```go
// FormatInputSchema 将 JSON Schema 转换为人类可读格式
// 返回: ["key:type={value} // 描述", ...]
func FormatInputSchema(schema any) []string {
    if schema == nil {
        return nil
    }

    schemaMap, ok := schema.(map[string]any)
    if !ok {
        return nil
    }

    properties, _ := schemaMap["properties"].(map[string]any)
    if properties == nil {
        return nil
    }

    var result []string
    for key, prop := range properties {
        propMap, ok := prop.(map[string]any)
        if !ok {
            continue
        }

        jsonType, _ := propMap["type"].(string)
        typeHint := jsonTypeToTypeHint(jsonType)

        // 格式: key:type={value} // 描述
        line := key + ":" + typeHint + "={value}"
        if desc, ok := propMap["description"].(string); ok && desc != "" {
            line += " // " + desc
        }
        result = append(result, line)
    }

    return result
}

// 类型映射
func jsonTypeToTypeHint(jsonType string) string {
    switch jsonType {
    case "number":
        return "number"
    case "integer":
        return "int"
    case "boolean":
        return "bool"
    case "array":
        return "array"
    case "object":
        return "object"
    case "string":
        return "string"
    default:
        return "string"
    }
}

// GetRequiredParams 提取必填参数列表
func GetRequiredParams(schema any) []string {
    schemaMap, ok := schema.(map[string]any)
    if !ok {
        return nil
    }

    var result []string
    if required, ok := schemaMap["required"].([]any); ok {
        for _, r := range required {
            if s, ok := r.(string); ok {
                result = append(result, s)
            }
        }
    }
    return result
}

// GetParamInfoList 提取结构化参数信息
func GetParamInfoList(schema any) []ParamInfo {
    schemaMap, ok := schema.(map[string]any)
    if !ok {
        return nil
    }

    properties, _ := schemaMap["properties"].(map[string]any)
    if properties == nil {
        return nil
    }

    // 构建必填集合
    requiredSet := make(map[string]bool)
    if required, ok := schemaMap["required"].([]any); ok {
        for _, r := range required {
            if s, ok := r.(string); ok {
                requiredSet[s] = true
            }
        }
    }

    var result []ParamInfo
    for key, prop := range properties {
        propMap, ok := prop.(map[string]any)
        if !ok {
            continue
        }

        jsonType, _ := propMap["type"].(string)
        info := ParamInfo{
            Name:        key,
            Type:        jsonTypeToTypeHint(jsonType),
            Required:    requiredSet[key],
            Description: "",
        }
        if desc, ok := propMap["description"].(string); ok {
            info.Description = desc
        }
        result = append(result, info)
    }

    return result
}

// buildCallExample 构建调用示例
func buildCallExample(params []ParamInfo) string {
    if len(params) == 0 {
        return ""
    }

    var parts []string
    for _, p := range params {
        parts = append(parts, fmt.Sprintf("%s:%s={value}", p.Name, p.Type))
    }
    return strings.Join(parts, " ")
}
```

**输出示例：**
```json
{
  "success": true,
  "data": {
    "server": "openDeepWiki",
    "tool": {
      "name": "get_repository",
      "description": "获取仓库详情...",
      "required": "repo_id OR repo_name",
      "param_format": "key:type=value (type: string/number/bool)",
      "param_example": [
        "repo_id:number={value} // 仓库ID，数字类型，优先使用",
        "repo_name:string={value} // 仓库名称，字符串类型",
        "include_content:bool={value} // 是否返回文档完整内容"
      ],
      "call_example": "lc mcp openDeepWiki get_repository repo_id:number={value} repo_name:string={value} include_content:bool={value}"
    }
  }
}
```

### 6.4 3+ 参数模式（call）

**行为：** 解析参数并调用工具

```go
func runMCPCall(ctx context.Context, d *Dispatcher, serverName, toolName string, kvArgs []string) {
    // 解析 key=value 参数
    params, err := ParseKVArgs(kvArgs)
    if err != nil {
        printMCPError(...)
        os.Exit(1)
    }

    // 调用工具
    actualServer, result, err := d.CallTool(ctx, toolName, serverName, params)
    if err != nil {
        printMCPError(...)
        os.Exit(1)
    }

    printMCPSuccess(map[string]any{
        "server": actualServer,
        "method": toolName,
        "result": result,
    })
}
```

#### 6.4.1 ParseKVArgs 实现（关键）

```go
// ParseKVArgs 解析 "key=value" 或 "key:type=value" 格式
// 支持类型标注: key:string=val, key:number=123, key:bool=true
// 默认无类型标注时为 string
func ParseKVArgs(args []string) (map[string]any, error) {
    result := make(map[string]any, len(args))

    for _, arg := range args {
        // 找第一个 '='
        idx := strings.Index(arg, "=")
        if idx < 0 {
            return nil, &MCPError{
                Code:    ErrCodeParamInvalid,
                Message: fmt.Sprintf("参数格式错误：%q（应为 key=value 或 key:type=value 格式）", arg),
            }
        }

        keyPart := arg[:idx]
        valStr := arg[idx+1:]

        // 解析 key 和可选类型标注
        key, typeHint, err := parseKeyWithType(keyPart)
        if err != nil {
            return nil, err
        }

        // 根据类型转换值
        convertedVal, err := convertValue(valStr, typeHint)
        if err != nil {
            return nil, &MCPError{
                Code:    ErrCodeParamInvalid,
                Message: fmt.Sprintf("参数 %q 的值 %q 无法转换为类型 %q: %v", key, valStr, typeHint, err),
            }
        }

        result[key] = convertedVal
    }

    return result, nil
}

// parseKeyWithType 解析 "key" 或 "key:type"
func parseKeyWithType(keyPart string) (string, string, error) {
    // 找冒号分隔符
    colonIdx := strings.Index(keyPart, ":")
    if colonIdx < 0 {
        return keyPart, "string", nil  // 无类型标注，默认 string
    }

    key := keyPart[:colonIdx]
    typeHint := keyPart[colonIdx+1:]

    if key == "" {
        return "", "", &MCPError{
            Code:    ErrCodeParamInvalid,
            Message: "参数键不能为空",
        }
    }

    // 验证类型
    validTypes := map[string]bool{
        "string": true, "number": true, "int": true,
        "float": true, "bool": true, "boolean": true,
    }
    if !validTypes[typeHint] {
        return "", "", &MCPError{
            Code:    ErrCodeParamInvalid,
            Message: fmt.Sprintf("参数 %q 使用了不支持的类型 %q，支持的类型：string, number, int, float, bool", key, typeHint),
        }
    }

    return key, typeHint, nil
}

// convertValue 根据类型转换字符串值
func convertValue(valStr, typeHint string) (any, error) {
    switch typeHint {
    case "string":
        return valStr, nil

    case "number", "int":
        // 先尝试整数
        if intVal, err := strconv.ParseInt(valStr, 10, 64); err == nil {
            return float64(intVal), nil
        }
        // 再尝试浮点数
        return strconv.ParseFloat(valStr, 64)

    case "float":
        return strconv.ParseFloat(valStr, 64)

    case "bool", "boolean":
        lower := strings.ToLower(valStr)
        if lower == "true" || lower == "1" || lower == "yes" {
            return true, nil
        }
        if lower == "false" || lower == "0" || lower == "no" {
            return false, nil
        }
        return nil, fmt.Errorf("无法将 %q 解析为布尔值", valStr)

    default:
        return valStr, nil
    }
}
```

**输出示例：**
```json
{
  "success": true,
  "data": {
    "server": "openDeepWiki",
    "method": "list_repositories",
    "result": {
      "repositories": [...],
      "total": 58
    }
  }
}
```

---

## 7. 统一输出格式

### 7.1 成功输出

```json
{
  "success": true,
  "data": {...},
  "meta": {
    "timestamp": "2026-03-23T10:00:00Z",
    "version": "v0.2.8"
  }
}
```

### 7.2 错误输出

```json
{
  "success": false,
  "error": {
    "code": "MCP_SERVER_NOT_FOUND",
    "message": "配置中不存在 Server 'xxx'",
    "details": {...}
  },
  "meta": {
    "timestamp": "2026-03-23T10:00:00Z",
    "version": "v0.2.8"
  }
}
```

### 7.3 输出函数实现

```go
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
```

---

## 8. 关键测试用例

### 8.1 单元测试（internal/mcp 包）

| 测试 | 描述 |
|------|------|
| `TestSearchPaths_notEmpty` | 搜索路径非空 |
| `TestLoadConfig_noneExist` | 无配置文件返回 MCP_CONFIG_NOT_FOUND |
| `TestLoadConfig_mergesAllFiles` | 多文件合并 |
| `TestLoadConfig_highPriorityWinsOnConflict` | 同名 Server 高优先级优先 |
| `TestInferTransportType_*` | 传输类型推断（6 个测试） |
| `TestParseKVArgs_valid` | key=value 解析 |
| `TestParseKVArgs_typeNumber` | number 类型标注 |
| `TestParseKVArgs_typeBool` | bool 类型标注 |

### 8.2 E2E 测试（e2e/integration）

| 测试 | 命令 | 断言 |
|------|------|------|
| TestMCPList | `lc mcp` | `servers` 含配置，`tools` 字段不存在 |
| TestMCPServerInfo | `lc mcp openDeepWiki` | `server.tools` 非空数组 |
| TestMCPToolInfo | `lc mcp openDeepWiki list_repositories` | `tool.param_format`, `param_example`, `call_example` 存在 |
| TestMCPCallTool | `lc mcp openDeepWiki list_repositories limit=3` | `result` 非空 |
| TestMCPMethodNotFound | `lc mcp openDeepWiki xxx` | `error.code == MCP_METHOD_NOT_FOUND` |

---

## 9. 实现检查清单

- [ ] 配置文件搜索路径正确（5 个 Unix 路径 / 6 个 Windows 路径）
- [ ] 多配置文件合并，同名 Server 高优先级覆盖
- [ ] 传输类型推断逻辑完整（6 种情况）
- [ ] 支持 SSE/Streamable/stdio 三种传输
- [ ] 0 参数模式使用 ListServersConfig，不连接服务器
- [ ] 1 参数模式连接服务器获取 tools 列表
- [ ] 2 参数模式输出字段国际化（param_format, param_example, call_example, required）
- [ ] 参数示例格式为 `key:type={value} // 描述`
- [ ] 调用示例包含所有参数
- [ ] ParseKVArgs 支持类型标注（:string, :number, :int, :float, :bool）
- [ ] 所有输出统一为 JSON 格式，含 success/data/error/meta
- [ ] 错误码使用常量（MCP_*）
- [ ] 时间戳使用 RFC3339 格式
