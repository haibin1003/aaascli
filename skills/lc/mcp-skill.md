---
name: lc-mcp
description: |
  MCP 协议集成。当用户提到 mcp、model context protocol、mcp server、调用外部工具、或需要通过 MCP 协议调用工具服务时触发。
metadata:
  {
    "joinai-code":
      {
        "requires": { "bins": ["lc"] },
      },
  }
---

# MCP 协议集成

## MCP 配置

配置文件按以下优先顺序搜索：
1. `<git-root>/joinai-code.json`
2. `<git-root>/.joinai-code/joinai-code.json`
3. `~/.config/joinai-code/joinai-code.json`
4. `~/.config/modelcontextprotocol/mcp.json`
5. `~/.config/mcp/config.json`
6. `./mcp.json`

## 配置格式

**标准 MCP 格式**：
```json
{
  "mcpServers": {
    "myServer": {
      "url": "https://api.example.com/mcp"
    }
  }
}
```

**joinai-code 格式**：
```json
{
  "mcp": {
    "remote-service": {
      "type": "remote",
      "url": "https://api.example.com/mcp"
    }
  }
}
```

## 常用命令

### 列出所有配置的 Server
```bash
lc mcp
```
仅显示配置信息，不获取工具列表，速度快。

### 列出指定 Server 的工具
```bash
lc mcp <server-name>
# 例如
lc mcp openDeepWiki
```

### 查看工具详情
```bash
lc mcp <server-name> <tool-name>
# 例如
lc mcp openDeepWiki GetWikiContents
```

### 调用工具
```bash
lc mcp <server-name> <tool-name> [key=value ...]
# 例如
lc mcp openDeepWiki GetWikiContents owner=torvalds repo=linux
lc mcp opdk read_document doc_id=615
```

## 输出格式

默认输出紧凑 JSON，使用 `--pretty` 格式化：
```bash
lc mcp opdk read_document doc_id=615 --pretty
```

## 注意事项

- Server 名称区分大小写
- 工具参数格式为 `key=value`
- 远程 Server 需要网络连接
