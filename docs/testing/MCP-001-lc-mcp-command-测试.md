# MCP-001 lc mcp 命令测试文档

## 变更记录表

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|---------|------|
| v1.0 | 2026-03-22 | 初始版本 | AI Assistant |
| v2.0 | 2026-03-22 | 适配新命令结构（server 位置参数化）；新增 TC-3（Server 信息模式）和 TC-4（Server 不存在）；更新所有 TC 命令示例；移除 `--server` flag 相关测试 | AI Assistant |
| v3.0 | 2026-03-22 | 适配配置全扫描合并策略；新增 UT-5（mergesAllFiles）和 UT-6（highPriorityWinsOnConflict）；移除旧 UT-5（firstPathWins）；更新 E2E 断言字段 `configFile` → `configFiles`；更新边界测试 BT-3 | AI Assistant |
| v4.0 | 2026-03-22 | 新增 Windows 路径支持（`runtime.GOOS` 运行时判断）；UT-1/UT-2 在所有平台适用；新增 BT-4（Windows 路径构建） | AI Assistant |
| v5.0 | 2026-03-23 | 新增传输类型推断单元测试（UT-13 ~ UT-18）；新增 stdio/Streamable HTTP 传输 E2E 测试（TC-13 ~ TC-15）；更新测试 Server 配置示例 | AI Assistant |
| v6.0 | 2026-03-23 | 新增参数类型标注单元测试（UT-19 ~ UT-24）；更新工具详情 E2E 测试断言（TC-8）；新增边界测试 BT-5 ~ BT-7 | AI Assistant |
| v7.0 | 2026-03-23 | 更新 TC-4 断言：0 参数模式不再返回 tools 列表；新增 BT-8 边界测试验证快速响应 | AI Assistant |
| v8.0 | 2026-03-23 | 更新 TC-8 和 BT-7 断言：工具详情字段国际化（`param_format`, `param_example`, `call_example`）；新增 `required` 字段验证 | AI Assistant |

---

## 1. 测试概述

本文档定义 `lc mcp` 命令体系的测试用例和验收标准，覆盖：

- 配置加载逻辑（单元测试）
- 参数解析逻辑（单元测试）
- 命令行端到端行为（E2E 测试，依赖真实 MCP Server）

**真实 Server 配置（E2E 测试使用）：**

```
服务名称：openDeepWiki
URL：https://opendeepwiki.k8m.site/mcp/sse
配置文件：~/.config/mcp/config.json
```

---

## 2. 单元测试用例

> 对应文件：`internal/mcp/mcp_test.go`（12 个测试，全部通过）

### UT-1：配置文件搜索路径非空

| 项目 | 内容 |
|------|------|
| 函数 | `TestSearchPaths_notEmpty` |
| 测试目的 | 验证 `SearchPaths()` 返回至少一条路径 |
| 预期结果 | 返回非空 slice |

### UT-2：配置文件搜索路径返回副本

| 项目 | 内容 |
|------|------|
| 函数 | `TestSearchPaths_returnsACopy` |
| 测试目的 | 验证修改返回的 slice 不影响原始路径列表 |
| 预期结果 | 两次调用返回独立副本 |

### UT-3：配置文件加载 - 全部路径不存在

| 项目 | 内容 |
|------|------|
| 函数 | `TestLoadConfig_noneExist` |
| 测试目的 | 验证所有路径均不存在时返回 `MCP_CONFIG_NOT_FOUND` |
| 测试方法 | 替换 configSearchPaths 为不存在的路径 |
| 预期结果 | 返回 `*MCPError{Code: ErrCodeConfigNotFound}`，config 和 path 均为零值 |

### UT-4：配置文件加载 - 合法文件

| 项目 | 内容 |
|------|------|
| 函数 | `TestLoadConfig_validFile` |
| 测试目的 | 验证正确解析 mcpServers 配置 |
| 测试输入 | `{"mcpServers":{"test":{"url":"http://example.com","timeout":5000}}}` |
| 预期结果 | `cfg.MCPServers["test"].URL == "http://example.com"`, `Timeout == 5000`，path 指向该文件 |

### UT-5：配置文件加载 - 合并多个文件

| 项目 | 内容 |
|------|------|
| 函数 | `TestLoadConfig_mergesAllFiles` |
| 测试目的 | 验证多个配置文件均被加载且 Server 合并 |
| 测试方法 | file1 定义 serverA，file2 定义 serverB，configSearchPaths = [file1, file2] |
| 预期结果 | `cfg.MCPServers` 同时包含 serverA 和 serverB，`paths` 长度为 2 |

### UT-6：配置文件加载 - 同名 Server 高优先级文件胜出

| 项目 | 内容 |
|------|------|
| 函数 | `TestLoadConfig_highPriorityWinsOnConflict` |
| 测试目的 | 验证两个文件定义了同名 Server 时，高优先级文件（序号小）的配置生效 |
| 测试方法 | file1 定义 `shared.url=http://high-priority`，file2 定义 `shared.url=http://low-priority` |
| 预期结果 | `cfg.MCPServers["shared"].URL == "http://high-priority"`，`paths` 长度为 2 |

### UT-7（原 UT-6）：配置文件加载 - mcpServers 为空

| 项目 | 内容 |
|------|------|
| 函数 | `TestLoadConfig_emptymcpServers` |
| 测试目的 | 验证空 Server 列表时不报错 |
| 测试输入 | `{"mcpServers":{}}` |
| 预期结果 | `len(cfg.MCPServers) == 0`，无 error |

### UT-8（原 UT-7）：参数解析 - 合法 key=val 格式

| 项目 | 内容 |
|------|------|
| 函数 | `TestParseKVArgs_valid` |
| 测试输入 | `["id=10086", "name=tom", "age=25"]` |
| 预期结果 | `map{"id":"10086", "name":"tom", "age":"25"}` |

### UT-9（原 UT-8）：参数解析 - 值中含等号

| 项目 | 内容 |
|------|------|
| 函数 | `TestParseKVArgs_valueContainsEquals` |
| 测试目的 | 验证仅按第一个 `=` 分割 |
| 测试输入 | `["token=a=b=c"]` |
| 预期结果 | `map{"token":"a=b=c"}` |

### UT-10（原 UT-9）：参数解析 - 无等号参数

| 项目 | 内容 |
|------|------|
| 函数 | `TestParseKVArgs_noEquals_returnsError` |
| 测试输入 | `["id=10086", "badparam"]` |
| 预期结果 | 返回 `*MCPError{Code: ErrCodeParamInvalid}`，message 含 `"badparam"` |

### UT-11（原 UT-10）：参数解析 - 空参数列表

| 项目 | 内容 |
|------|------|
| 函数 | `TestParseKVArgs_empty` |
| 测试输入 | `[]` |
| 预期结果 | 返回 `map{}`，无 error |

### UT-12（原 UT-11）：参数解析 - 空值（key=）

| 项目 | 内容 |
|------|------|
| 函数 | `TestParseKVArgs_emptyValue` |
| 测试输入 | `["key="]` |
| 预期结果 | `map{"key":""}` |

### UT-13：传输类型推断 - 显式 transport 字段

| 项目 | 内容 |
|------|------|
| 函数 | `TestInferTransportType_explicitTransport` |
| 测试输入 | `ServerConfig{Transport: "stdio", Command: "npx"}` |
| 预期结果 | `"stdio"` |

### UT-14：传输类型推断 - type 字段映射

| 项目 | 内容 |
|------|------|
| 函数 | `TestInferTransportType_typeAlias` |
| 测试输入 | `ServerConfig{Type: "streamable-http"}` |
| 预期结果 | `"streamable"` |

### UT-15：传输类型推断 - URL 含 sse

| 项目 | 内容 |
|------|------|
| 函数 | `TestInferTransportType_urlContainsSSE` |
| 测试输入 | `ServerConfig{URL: "http://host/mcp/sse"}` |
| 预期结果 | `"sse"` |

### UT-16：传输类型推断 - URL 含 stream

| 项目 | 内容 |
|------|------|
| 函数 | `TestInferTransportType_urlContainsStream` |
| 测试输入 | `ServerConfig{URL: "http://host/mcp/streamable"}` |
| 预期结果 | `"streamable"` |

### UT-17：传输类型推断 - 默认 streamable

| 项目 | 内容 |
|------|------|
| 函数 | `TestInferTransportType_defaultStreamable` |
| 测试输入 | `ServerConfig{URL: "http://host/mcp"}` |
| 预期结果 | `"streamable"` |

### UT-18：传输类型推断 - command 存在

| 项目 | 内容 |
|------|------|
| 函数 | `TestInferTransportType_stdioByCommand` |
| 测试输入 | `ServerConfig{Command: "npx"}` |
| 预期结果 | `"stdio"` |

### UT-19：参数解析 - 类型标注 number

| 项目 | 内容 |
|------|------|
| 函数 | `TestParseKVArgs_typeNumber` |
| 测试输入 | `["age:number=25"]` |
| 预期结果 | `map{"age": 25.0}`（float64 类型） |

### UT-20：参数解析 - 类型标注 int

| 项目 | 内容 |
|------|------|
| 函数 | `TestParseKVArgs_typeInt` |
| 测试输入 | `["count:int=10"]` |
| 预期结果 | `map{"count": 10.0}`（内部转为 float64） |

### UT-21：参数解析 - 类型标注 float

| 项目 | 内容 |
|------|------|
| 函数 | `TestParseKVArgs_typeFloat` |
| 测试输入 | `["price:float=19.99"]` |
| 预期结果 | `map{"price": 19.99}` |

### UT-22：参数解析 - 类型标注 bool

| 项目 | 内容 |
|------|------|
| 函数 | `TestParseKVArgs_typeBool` |
| 测试输入 | `["enabled:bool=true", "disabled:bool=false"]` |
| 预期结果 | `map{"enabled": true, "disabled": false}` |

### UT-23：参数解析 - 不支持的类型报错

| 项目 | 内容 |
|------|------|
| 函数 | `TestParseKVArgs_invalidType` |
| 测试输入 | `["key:invalid=value"]` |
| 预期结果 | 返回 `*MCPError{Code: ErrCodeParamInvalid}`，message 含 `"invalid"` |

### UT-24：参数解析 - 类型转换失败报错

| 项目 | 内容 |
|------|------|
| 函数 | `TestParseKVArgs_typeConversionError` |
| 测试输入 | `["age:number=abc"]` |
| 预期结果 | 返回 `*MCPError{Code: ErrCodeParamInvalid}`，message 含 `"abc"` 和 `"number"` |

---

## 3. E2E 测试用例

> 对应文件：`e2e/integration/mcp_e2e_test.go`（15 个测试）
>
> **前置条件：**
> - `~/.config/mcp/config.json` 存在，包含以下 Server 配置：
>   - `openDeepWiki`：SSE 传输，`https://opendeepwiki.k8m.site/mcp/sse`
>   - `odk`：Streamable HTTP 传输，`https://opendeepwiki.k8m.site/mcp/streamable`
>   - `mcp-official-stdio`：stdio 传输，`npx -y @modelcontextprotocol/server-everything`
> - 网络可访问 `https://opendeepwiki.k8m.site`
> - 本地已安装 Node.js 和 npx（用于 stdio 测试）
> - 已构建最新 `lc` 二进制：`make build`
> - 短模式（`-short`）下网络测试自动跳过

### TC-1：帮助信息（无需网络）

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPHelp` |
| 测试命令 | `lc mcp --help` |
| 预期结果 | 帮助文本中包含关键词 `mcp`、`MCP`、`serverName`、`配置` |

### TC-2：无配置文件时报错（无需网络）

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPNoConfig` |
| 前置条件 | 不写入 MCP 配置文件 |
| 测试命令 | `lc mcp` |
| 预期结果 | 1. 返回非 0 退出码<br>2. `success=false`<br>3. `error.code == "MCP_CONFIG_NOT_FOUND"`<br>4. `error.details.searchPaths` 存在 |

### TC-3：参数格式非法报错（无需网络）

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPParamInvalid` |
| 测试命令 | `lc mcp openDeepWiki list_repositories badparam` |
| 预期结果 | 1. `success=false`<br>2. `error.code == "MCP_PARAM_INVALID"`<br>3. `error.message` 包含 `"badparam"` |

### TC-4：列出所有 Server（0 参数）

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPList` |
| 测试命令 | `lc mcp` |
| 预期结果 | 1. `success=true`<br>2. `data.configFiles` 为非空数组<br>3. `data.servers` 含 `openDeepWiki`<br>4. `openDeepWiki.name`、`openDeepWiki.transport`、`openDeepWiki.url` 非空<br>5. `openDeepWiki.tools` 字段不存在（或为空） |

> 注：此模式仅读取配置不连接服务器，响应速度快，不返回 tools 列表。如需查看 tools，请使用 `lc mcp <serverName>`。 |

### TC-5：输出为合法 JSON

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPListOutputIsValidJSON` |
| 测试命令 | `lc mcp` |
| 预期结果 | stdout 可被 `json.Unmarshal` 解析，无错误 |

### TC-6：查看指定 Server 信息（1 参数）

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPServerInfo` |
| 测试命令 | `lc mcp openDeepWiki` |
| 预期结果 | 1. `success=true`<br>2. `data.configFiles` 为非空数组<br>3. `data.server.name == "openDeepWiki"`<br>4. `data.server.url` 非空<br>5. `data.server.tools` 非空数组 |

### TC-7：Server 不存在返回错误（1 参数）

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPServerNotFound` |
| 测试命令 | `lc mcp nonExistentServer_XYZ_12345` |
| 预期结果 | 1. 返回非 0 退出码<br>2. `success=false`<br>3. `error.code == "MCP_SERVER_NOT_FOUND"` |

### TC-8：查看工具详情（2 参数）

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPToolInfo` |
| 测试命令 | `lc mcp openDeepWiki list_repositories` |
| 预期结果 | 1. `success=true`<br>2. `data.server == "openDeepWiki"`<br>3. `data.tool.name == "list_repositories"`<br>4. `data.tool.description` 非空<br>5. `data.tool.param_format` 存在且含 `key:type=value`<br>6. `data.tool.param_example` 为非空数组，元素格式为 `key:type={value}`<br>7. `data.tool.call_example` 存在且含 `lc mcp openDeepWiki list_repositories`<br>8. `data.tool.required` 存在（如有必填参数） |

### TC-9：工具不存在返回错误（2 参数）

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPMethodNotFound` |
| 测试命令 | `lc mcp openDeepWiki nonExistentMethodXYZ_12345` |
| 预期结果 | 1. 返回非 0 退出码<br>2. `success=false`<br>3. `error.code == "MCP_METHOD_NOT_FOUND"` |

### TC-10：调用工具（3+ 参数）

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPCallTool` |
| 测试命令 | `lc mcp openDeepWiki list_repositories limit=3` |
| 预期结果 | 1. `success=true`<br>2. `data.server == "openDeepWiki"`<br>3. `data.method == "list_repositories"`<br>4. `data.result` 非 nil，可解析为 JSON<br>5. `result.repositories` 为非空数组，长度 ≤ 3 |

### TC-11：成功响应包含 meta 字段

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPOutputHasMeta` |
| 测试命令 | `lc mcp` |
| 预期结果 | 1. `meta.timestamp` 非空<br>2. `meta.version` 非空 |

### TC-12：错误响应也包含 meta 字段

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPErrorOutputHasMeta` |
| 前置条件 | 不写入 MCP 配置文件 |
| 测试命令 | `lc mcp` |
| 预期结果 | `meta.timestamp` 非空（即使 success=false 也输出 meta） |

### TC-13：Streamable HTTP 传输（odk Server）

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPStreamableTransport` |
| 测试命令 | `lc mcp odk` |
| 预期结果 | 1. `success=true`<br>2. `data.server.transport == "streamable"`<br>3. `data.server.tools` 非空数组 |

### TC-14：stdio 传输（本地命令）

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPStdioTransport` |
| 测试命令 | `lc mcp mcp-official-stdio` |
| 预期结果 | 1. `success=true`<br>2. `data.server.transport == "stdio"`<br>3. `data.server.command == "npx"`<br>4. `data.server.tools` 非空数组（如含 `echo`） |

### TC-15：调用 stdio Server 工具

| 项目 | 内容 |
|------|------|
| 函数 | `TestMCPCallStdioTool` |
| 测试命令 | `lc mcp mcp-official-stdio echo message=hello` |
| 预期结果 | 1. `success=true`<br>2. `data.server == "mcp-official-stdio"`<br>3. `data.method == "echo"`<br>4. `data.result` 包含 `"Echo: hello"` |

---

## 4. 边界测试

### BT-1：参数值为空字符串

| 项目 | 内容 |
|------|------|
| 测试目的 | 验证 `key=` 格式（空值）正常解析 |
| 测试方法 | 单元测试 `TestParseKVArgs_emptyValue` |
| 预期结果 | 解析为 `{"key": ""}`，不报错 |

### BT-2：配置文件存在但 mcpServers 为空

| 项目 | 内容 |
|------|------|
| 测试目的 | 验证空 Server 列表时 `lc mcp` 正常返回 |
| 测试输入 | `{"mcpServers": {}}` |
| 预期结果 | `data.servers = []`，`success=true` |

### BT-3：配置文件合并与优先级验证

| 项目 | 内容 |
|------|------|
| 测试目的 | 验证多个文件均被加载，同名 Server 以高优先级文件配置为准 |
| 测试方法 | 单元测试 `TestLoadConfig_mergesAllFiles` 和 `TestLoadConfig_highPriorityWinsOnConflict` |
| 预期结果 | 不同名 Server 均合并进配置；同名 Server 使用路径索引更小（优先级更高）文件的配置 |

### BT-4：Windows 路径正确构建

| 项目 | 内容 |
|------|------|
| 测试目的 | 验证 Windows 平台路径通过环境变量正确拼接（非 `%xxx%` 原始字符串） |
| 测试方法 | 在 Windows 环境下运行 `SearchPaths()`，检查返回路径不含 `%` 字符且 `filepath.IsAbs()` 对非当前目录路径返回 true |
| 说明 | 此测试仅在 Windows CI 环境下验证；Linux/macOS 构建通过 `GOOS=windows go build` 交叉编译验证 |

### BT-5：参数类型标注 - number 类型转换

| 项目 | 内容 |
|------|------|
| 测试目的 | 验证 `key:number=value` 格式正确解析为 float64 |
| 测试方法 | 单元测试 `TestParseKVArgs_typeNumber` |
| 预期结果 | `"123"` → `123.0`，`"3.14"` → `3.14` |

### BT-6：参数类型标注 - bool 类型转换

| 项目 | 内容 |
|------|------|
| 测试目的 | 验证 `key:bool=value` 支持多种布尔值写法 |
| 测试方法 | 单元测试 `TestParseKVArgs_typeBool` |
| 预期结果 | `true`/`1`/`yes` → `true`；`false`/`0`/`no` → `false` |

### BT-7：工具详情输出格式验证

| 项目 | 内容 |
|------|------|
| 测试目的 | 验证工具详情输出包含人类可读的参数格式（字段国际化） |
| 测试方法 | E2E 测试 `TestMCPToolInfo` |
| 预期结果 | 输出包含 `param_format`、`param_example`、`call_example`、`required` 字段，参数示例格式为 `key:type={value} // 描述` |

### BT-8：0 参数模式快速响应验证

| 项目 | 内容 |
|------|------|
| 测试目的 | 验证 `lc mcp`（0 参数）不连接服务器，响应速度快 |
| 测试方法 | E2E 测试 `TestMCPList`，测量执行时间 |
| 预期结果 | 命令执行时间 < 1s（不依赖网络连接），输出不包含 `tools` 字段 |

---

## 5. 验收检查清单

- [x] UT-1 ～ UT-24 全部通过（`go test ./internal/mcp/... -v`）
- [x] TC-1 ～ TC-15 全部通过（`go test ./e2e/integration/... -v -run TestMCP`）
- [x] 参数类型标注测试：number/int/float/bool 类型正确转换
- [x] 工具详情输出测试：包含 `param_format`、`param_example`、`call_example`、`required`
- [x] `go vet ./internal/mcp/... ./cmd/...` 无报错
- [x] `go build ./internal/mcp/... ./cmd/...` 成功
- [x] `lc mcp --help` 输出命令说明，含 `serverName`、`method`、`key=val` 关键词
- [x] `lc mcp` 输出合法 JSON，可被 `jq` 解析，含 `transport` 字段
- [x] `lc mcp openDeepWiki` 输出合法 JSON，含工具列表，transport 为 `sse`
- [x] `lc mcp odk` 输出合法 JSON，transport 为 `streamable`
- [x] `lc mcp mcp-official-stdio` 输出合法 JSON，transport 为 `stdio`，含 `command` 字段
- [x] `lc mcp openDeepWiki list_repositories limit=3` 成功调用并返回仓库列表
- [x] `lc mcp mcp-official-stdio echo message=hello` 成功调用 stdio Server 工具
