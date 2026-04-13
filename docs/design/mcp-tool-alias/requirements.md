# MCP 工具静态命令封装 - 需求文档

## 背景

当前 `lc mcp` 命令采用动态调用方式：

```bash
lc mcp <serverName> <method> [key=value ...]
```

这种方式虽然灵活，但与 `lc` 其他命令（如 `lc ci list`、`lc req create`）的风格不一致，
用户需要记忆 Server 名称和方法名，体验不够统一。

## 需求目标

将常用的 MCP 动态工具调用封装为**统一的静态命令**，提供与现有 `lc` 命令一致的 CLI 体验。

## 功能需求

### FR-1: 静态命令注册

预定义的 MCP 工具应映射为静态命令：

```bash
# 动态方式（现有）
lc mcp openDeepWiki GetWikiContents owner=torvalds repo=linux

# 静态方式（新增）
lc get-repo-wiki owner=torvalds repo=linux
```

### FR-2: 命令常驻（执行时校验）

命令应**始终注册**，不随 MCP Server 可用性变化。Server 未配置时执行报错：

```json
{
  "success": false,
  "error": {
    "code": "MCP_SERVER_NOT_CONFIGURED",
    "message": "命令 'get-repo-wiki' 需要 MCP Server 'openDeepWiki'，但未找到相关配置"
  }
}
```

### FR-3: 参数校验

命令应校验必需参数，缺失时给出明确提示：

```json
{
  "success": false,
  "error": {
    "code": "MISSING_ARG",
    "message": "缺少必需参数: owner"
  }
}
```

### FR-4: 统一输出格式

保持与 `lc` 其他命令一致的 JSON 输出格式。

## 非功能需求

### NFR-1: 扩展性

预留接口，未来支持用户自定义工具映射（第二阶段）。

### NFR-2: 性能

命令启动时仅注册，不连接 MCP Server；执行时才进行校验和调用。

## 首批实现范围

本次仅实现一个命令作为试点：

| 命令 | Server | Method | 必需参数 |
|------|--------|--------|----------|
| `get-repo-wiki` | openDeepWiki | GetWikiContents | owner, repo |

## 验收标准

- [ ] `lc get-repo-wiki --help` 显示命令帮助
- [ ] `lc get-repo-wiki owner=xxx repo=yyy` 正确调用 MCP Server
- [ ] Server 未配置时返回 `MCP_SERVER_NOT_CONFIGURED` 错误
- [ ] 缺少参数时返回 `MISSING_ARG` 错误
- [ ] 输出格式与其他 `lc` 命令一致
