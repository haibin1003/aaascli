# MCP 工具静态命令封装 - 实现总结

## 实现范围

本次实现以 `get-repo-wiki` 命令作为试点，验证 MCP 工具静态化方案的可行性。

## 核心实现

### 1. 工具注册中心 (`internal/mcp/registry.go`)

- 定义 `ToolDefinition` 结构体
- 定义 `PredefinedTools` 预定义工具表
- 仅包含一个工具：`get-repo-wiki`

### 2. MCP 配置扩展 (`internal/mcp/config.go`)

- 新增 `HasServer(name string) bool` 方法
- 用于执行时检查 Server 是否配置

### 3. 工具命令实现 (`cmd/lc/tool.go`)

- `init()` 遍历 `PredefinedTools` 动态生成命令
- `runTool()` 统一执行逻辑：
  - 解析 `key=value` 参数
  - 校验必需参数
  - 检查 Server 配置
  - 调用 `Dispatcher.CallTool`

### 4. 命令注册 (`cmd/lc/root.go`)

- `init()` 中调用 `registerToolCommands()`

## 命令使用

```bash
# 查看帮助
lc get-repo-wiki --help

# 正常调用
lc get-repo-wiki owner=torvalds repo=linux

# 缺少参数（报错）
lc get-repo-wiki repo=linux
# Error: MISSING_ARG - 缺少必需参数: owner

# Server 未配置（报错）
lc get-repo-wiki owner=torvalds repo=linux
# Error: MCP_SERVER_NOT_CONFIGURED - 命令 'get-repo-wiki' 需要 MCP Server 'openDeepWiki'
```

## 设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 命令格式 | `lc get-repo-wiki` | 扁平化，简洁直接 |
| 参数格式 | `key=value` | 与 `lc mcp` 保持一致，避免顺序依赖 |
| 校验时机 | 执行时 | 命令始终可见，体验统一 |
| 错误格式 | JSON | 与 `lc` 其他命令一致 |

## 未来扩展

1. **增加预定义工具**：在 `PredefinedTools` 中追加定义
2. **用户自定义**：支持配置文件加载自定义工具映射
3. **参数类型**：支持强类型参数校验（int, bool 等）
