# MCP-001 lc mcp 命令需求文档

## 0. 变更记录表

| 修改人 | 修改时间 | 修改内容 |
|--------|----------|----------|
| AI Assistant | 2026-03-22 | 初始版本 |
| AI Assistant | 2026-03-22 | v2：将 `server` 固定子命令改为可变 serverName 位置参数，新增 1 参数模式（查看指定 Server 的工具列表） |
| AI Assistant | 2026-03-22 | v3：配置加载改为扫描全部 5 个路径并合并，同名 Server 高优先级文件胜出；输出字段 `configFile` → `configFiles`（数组） |
| AI Assistant | 2026-03-22 | v4：新增 Windows 平台搜索路径支持（`%APPDATA%`、`%USERPROFILE%`、`%ProgramData%`），通过 `runtime.GOOS` 运行时判断选择平台路径 |
| AI Assistant | 2026-03-23 | v5：新增三种传输方式支持（SSE/Streamable HTTP/stdio），智能传输类型推断，更新配置格式和输出字段 |
| AI Assistant | 2026-03-23 | v6：新增参数类型标注格式 `key:type=value`；工具详情输出改为人类可读格式（含参数格式说明、参数示例、调用示例） |
| AI Assistant | 2026-03-23 | v7：`lc mcp`（0 参数）不再获取 tools 列表，仅显示服务器配置信息，提升响应速度 |
| AI Assistant | 2026-03-23 | v8：工具详情输出字段国际化（`param_format`, `param_example`, `call_example`），新增 `required` 字段 |

---

## 1. 背景（Why）

灵畿 CLI（lc）目前支持对灵畿平台的各类资源操作（需求、任务、仓库等）。随着 MCP（Model Context Protocol）生态的发展，越来越多的工具和服务以 MCP Server 的形式暴露能力。

为了让开发者可以在 lc 工具链中直接调用任意 MCP Server 提供的工具（Tool），需要为 lc 新增 `mcp` 子命令体系。该命令负责：

1. 发现并列出本机配置的所有 MCP Server 及其可用工具
2. 列出指定 MCP Server 的工具详情
3. 查看某个工具的详细参数说明
4. 直接在命令行调用任意 MCP 工具并以 JSON 返回结果

---

## 2. 目标（What，必须可验证）

- [x] `lc mcp`（0 参数）：列出所有已配置的 MCP Server 信息（名称、传输类型、URL/命令），输出 JSON（不获取工具列表，速度快）
- [x] `lc mcp <serverName>`（1 参数）：列出指定 Server 的工具列表，输出 JSON
- [x] `lc mcp <serverName> <method>`（2 参数）：显示该 Server 上指定工具的名称、描述、参数 schema，输出 JSON
- [x] `lc mcp <serverName> <method> key=val ...`（3+ 参数）：连接对应 MCP Server，调用该工具，返回结果 JSON
- [x] MCP 配置按优先级顺序从多个标准路径中自动搜索加载
- [x] 所有命令默认输出 JSON 格式，行为与现有 lc 命令一致

---

## 3. 非目标（Explicitly Out of Scope）

- 不实现 MCP Server 管理（增删改配置），配置文件由用户手动维护
- 不实现 MCP Server 的认证 / 授权逻辑（使用 SDK 默认行为）
- 不实现流式输出（只返回最终完整结果）
- 不实现 MCP Resources / Prompts 端点（只调用 Tools）

---

## 4. 使用场景 / 用户路径

### 场景 1：查看所有 MCP Server 配置

```bash
$ lc mcp
```

输出（JSON，仅显示服务器配置，不获取工具列表）：
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
        "name": "local-stdio-server",
        "transport": "stdio",
        "command": "npx"
      }
    ]
  },
  "meta": {
    "timestamp": "2026-03-22T09:58:13Z",
    "version": "v0.2.8"
  }
}
```

> 注：此模式仅读取配置文件，不连接服务器获取工具列表，响应速度快。如需查看工具列表，请使用 `lc mcp <serverName>`。

### 场景 2：查看指定 Server 的工具列表

```bash
$ lc mcp openDeepWiki
```

输出（JSON）：
```json
{
  "success": true,
  "data": {
    "configFiles": ["/home/user/.config/mcp/config.json"],
    "server": {
      "name": "openDeepWiki",
      "transport": "sse",
      "url": "https://opendeepwiki.k8m.site/mcp/sse",
      "tools": ["get_document_summary", "get_repository", "list_repositories", "read_document", "search_documents"]
    }
  }
}
```

### 场景 3：查看某工具的参数说明

```bash
$ lc mcp openDeepWiki get_repository
```

输出（JSON）：
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
  }
}
```

> 注：工具详情输出将 JSON Schema 转换为人类可读的 `key:type=value` 格式，便于 AI 直接理解和使用。
> - `required`: 必填参数名，多个时用 "OR" 连接
> - `param_format`: 参数格式说明
> - `param_example`: 参数列表示例，格式为 `key:type={value} // 描述`
> - `call_example`: 完整的命令调用示例，包含所有参数

### 场景 4：调用 MCP 工具（含类型标注）

支持两种参数格式：
- `key=value`：默认为字符串类型
- `key:type=value`：显式指定参数类型（string/number/int/float/bool）

```bash
# 字符串类型（默认）
$ lc mcp openDeepWiki search_documents query=go

# 数字类型
$ lc mcp openDeepWiki list_repositories limit:number=3

# 布尔类型
$ lc mcp someServer someTool enabled:bool=true
```

输出（JSON）：
```json
{
  "success": true,
  "data": {
    "server": "openDeepWiki",
    "method": "list_repositories",
    "result": "{ ... }"
  }
}
```

### 场景 5：Server 名称不存在

```bash
$ lc mcp nonExistentServer
```

输出（JSON）：
```json
{
  "success": false,
  "error": {
    "code": "MCP_SERVER_NOT_FOUND",
    "message": "配置中不存在 Server 'nonExistentServer'，请使用 `lc mcp` 查看可用 Server"
  }
}
```

---

## 5. 功能需求清单（Checklist）

- [x] **FR-1** `lc mcp`（0 参数）：列出所有已配置 MCP Server 的基本信息（名称、传输类型、URL/命令），不连接服务器获取工具列表，响应速度快
- [x] **FR-2** `lc mcp <serverName>`（1 参数）：连接指定 Server，列出其工具名、URL，输出 JSON
- [x] **FR-3** `lc mcp <serverName> <method>`（2 参数）：在指定 Server 上查找该工具，返回工具描述和 inputSchema
- [x] **FR-4** `lc mcp <serverName> <method> key=val...`（3+ 参数）：将 key=val 解析为参数，调用工具，返回结果
- [x] **FR-5** 配置文件按以下优先级顺序搜索，**全部路径均扫描**，将所有找到的文件合并为一份配置；同名 Server 以优先级高（序号小）的文件为准（路径因平台不同，见下）：

  **Linux / macOS：**
  1. `~/.config/modelcontextprotocol/mcp.json`
  2. `~/.config/mcp/config.json`
  3. `./mcp.json`（当前目录）
  4. `./.mcp/config.json`（当前目录）
  5. `/etc/mcp/config.json`（系统级）

  **Windows：**
  1. `%APPDATA%\modelcontextprotocol\mcp.json`
  2. `%APPDATA%\mcp\config.json`
  3. `%USERPROFILE%\.mcp\config.json`
  4. `.\mcp.json`（当前目录）
  5. `.\.mcp\config.json`（当前目录）
  6. `%ProgramData%\mcp\config.json`（系统级）
- [x] **FR-6** 若未找到任何配置文件，返回 JSON 格式错误提示（错误码 `MCP_CONFIG_NOT_FOUND`），并列出已搜索的路径
- [x] **FR-7** 若 serverName 不存在于配置中，返回 `MCP_SERVER_NOT_FOUND` 错误
- [x] **FR-8** 若工具名不存在于指定 Server，返回 `MCP_METHOD_NOT_FOUND` 错误
- [x] **FR-9** 所有命令输出统一 JSON 格式（`success`, `data` / `error`, `meta`），与现有 lc 命令一致
- [x] **FR-10** 参数调用支持类型标注格式 `key:type=value`，支持类型：`string`（默认）、`number`、`int`、`float`、`bool`
- [x] **FR-11** 工具详情输出包含：
  - `参数格式`：说明 `key:type=value` 格式及支持的类型
  - `参数示例`：将 JSON Schema 转换为 `key:type=<值>` 格式列表，标注必填项和描述
  - `调用示例`：可直接复制使用的完整命令示例

---

## 6. 约束条件

### 技术约束
- 使用官方 MCP Go SDK：`github.com/modelcontextprotocol/go-sdk`
- 传输方式：HTTP/SSE（`SSEClientTransport`），对应配置中有 `url` 字段的 Server
- CLI 框架：保持与现有 lc 命令一致，使用 cobra

### 架构约束
- 命令文件放在 `cmd/lc/mcp.go`
- MCP 客户端逻辑放在 `internal/mcp/` 包下
- 不得修改现有命令的行为

### 安全约束
- 不在日志中打印 MCP Server 的响应内容（可能含敏感数据），除非开启 debug 模式
- 参数传递使用 `map[string]any`，支持显式类型标注（`key:type=value`）

### 性能约束
- 每个 `lc mcp` 调用列出模式下并发连接所有 Server（goroutine + WaitGroup），减少总等待时间
- 列出模式中若某个 Server 连接失败，在结果中标记 `error` 字段，不中断整体输出

---

## 7. 可修改 / 不可修改项

- ❌ 不可修改：输出格式（必须 JSON），配置文件搜索路径优先级顺序
- ✅ 可调整：MCP 连接超时时间（默认使用配置文件中的 timeout 字段，缺省 30s）
- ✅ 可调整：列出模式下各 Server 是否并发连接（建议并发，但串行也可接受）

---

## 8. 接口与数据约定

### 8.1 MCP 配置文件格式

```json
{
  "mcpServers": {
    "<serverName>": {
      "transport": "streamable",
      "url": "https://host/mcp",
      "timeout": 30000
    },
    "<stdioServerName>": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-everything"],
      "env": { "API_KEY": "xxx" },
      "timeout": 30000
    }
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `mcpServers` | object | 是 | 各 server 名称到配置的映射 |
| `mcpServers.<name>.transport` | string | 否 | 传输类型：`sse`、`streamable`、`stdio`；未指定时自动推断 |
| `mcpServers.<name>.type` | string | 否 | transport 的别名，用于兼容 Claude Desktop 等客户端（如 `streamable-http`） |
| `mcpServers.<name>.url` | string | HTTP 传输必填 | MCP Server 的 HTTP/SSE 端点 URL |
| `mcpServers.<name>.command` | string | stdio 传输必填 | 可执行文件路径（如 `npx`、`python`） |
| `mcpServers.<name>.args` | string[] | 否 | 传递给 command 的参数列表 |
| `mcpServers.<name>.env` | object | 否 | 额外的环境变量（key-value 形式） |
| `mcpServers.<name>.timeout` | number | 否 | 超时毫秒数，缺省 30000 |

#### 8.1.1 传输类型自动推断规则

当 `transport` 和 `type` 均未指定时，按以下优先级自动推断：

1. 配置了 `command` 字段 → `stdio`
2. URL 包含 `sse`（不区分大小写）→ `sse`
3. URL 包含 `stream`（不区分大小写）→ `streamable`
4. 默认 → `streamable`（现代协议优先）

#### 8.1.2 配置示例

**SSE（传统协议）：**
```json
{
  "mcpServers": {
    "legacy-server": { "url": "http://host/mcp/sse" }
  }
}
```

**Streamable HTTP（现代协议，默认）：**
```json
{
  "mcpServers": {
    "modern-server": { "url": "http://host/mcp" },
    "explicit-server": { "type": "streamable-http", "url": "http://host/mcp" }
  }
}
```

**stdio（本地命令）：**
```json
{
  "mcpServers": {
    "local-server": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-everything"],
      "env": { "DEBUG": "true" }
    }
  }
}
```

### 8.2 成功输出结构

```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "timestamp": "2026-03-22T09:58:13Z",
    "version": "v0.2.8"
  }
}
```

### 8.3 错误输出结构

```json
{
  "success": false,
  "error": {
    "code": "MCP_CONFIG_NOT_FOUND",
    "message": "未找到 MCP 配置文件，请在以下路径之一创建配置：...",
    "details": { "searchPaths": ["..."] }
  },
  "meta": {
    "timestamp": "2026-03-22T09:58:13Z",
    "version": "v0.2.8"
  }
}
```

### 8.4 错误码定义

| 错误码 | 说明 |
|--------|------|
| `MCP_CONFIG_NOT_FOUND` | 未找到任何 MCP 配置文件 |
| `MCP_SERVER_NOT_FOUND` | 配置中不存在指定 Server 名称 |
| `MCP_CONNECT_FAILED` | 连接 MCP Server 失败（包含传输类型信息和配置示例） |
| `MCP_METHOD_NOT_FOUND` | 指定 Server 上未找到该工具 |
| `MCP_METHOD_AMBIGUOUS` | 同名方法存在于多个 Server 中（仅内部 FindTool 路径使用） |
| `MCP_CALL_FAILED` | 工具调用失败 |
| `MCP_PARAM_INVALID` | 参数格式非法（非 key=val 格式） |

连接失败时，错误消息会明确告知当前使用的传输类型（如 `sse`、`streamable`、`stdio`），并提供各传输类型的配置示例，帮助用户排查配置问题。

---

## 9. 验收标准（Acceptance Criteria）

- 执行 `lc mcp` 且配置文件存在：返回包含所有 Server 配置信息（不含 tools 列表）的 JSON，`success=true`
- 执行 `lc mcp` 且未找到配置文件：返回 `success=false`，错误码 `MCP_CONFIG_NOT_FOUND`，并列出搜索路径
- 执行 `lc mcp <serverName>`（Server 存在）：返回该 Server 的工具列表，`success=true`
- 执行 `lc mcp <serverName>`（Server 不存在）：返回 `success=false`，错误码 `MCP_SERVER_NOT_FOUND`
- 执行 `lc mcp <serverName> <method>`（工具存在）：返回工具描述、参数格式说明、参数示例和调用示例，`success=true`
- 执行 `lc mcp <serverName> <method>`（工具不存在）：返回 `success=false`，错误码 `MCP_METHOD_NOT_FOUND`
- 执行 `lc mcp <serverName> <method> k=v...`：成功调用并返回结果，`success=true`
- 执行 `lc mcp <serverName> <method> key:type=value...`：根据类型标注正确转换参数类型后调用，`success=true`
- key=val 格式非法（如无等号）：返回 `success=false`，错误码 `MCP_PARAM_INVALID`，错误信息包含非法参数名
- 类型标注非法（如不支持的类型）：返回 `success=false`，错误码 `MCP_PARAM_INVALID`，错误信息包含非法类型
- 所有输出均为合法 JSON，可被 `jq` 解析

---

## 10. 风险与已知不确定点

| 风险 | 影响 | 处理方式 |
|------|------|----------|
| MCP Go SDK API 可能与示例代码有出入 | 编码时需核查实际 API | 开发前运行 `go get` 并查阅 SDK 源码 |
| SSE 连接在某些网络环境下超时 | 命令卡住 | 使用配置中的 timeout 字段控制超时 |
| inputSchema 格式各 Server 可能不一致 | 参数类型推断困难 | 所有参数统一以 `string` 类型传入 |

---

## 11. 非目标（再次明确）

- 不实现 MCP 配置文件的 CRUD 命令
- 不实现 MCP Resources / Prompts
- 不实现流式（streaming）结果输出
- 不再支持 `--server` 参数消歧（server 名称已作为位置参数强制指定）
