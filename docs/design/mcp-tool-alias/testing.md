# MCP 工具静态命令封装 - 测试文档

## 测试范围

针对 `get-repo-wiki` 命令的功能、错误处理和边界情况进行测试。

## 测试环境准备

### 1. 构建安装

```bash
sudo make install
lc version
```

### 2. MCP 配置

测试需要以下两种配置状态：

**状态 A：Server 已配置**

`~/.config/joinai-code/joinai-code.json`:
```json
{
  "mcp": {
    "openDeepWiki": {
      "type": "remote",
      "url": "https://wiki.example.com/mcp"
    }
  }
}
```

**状态 B：Server 未配置**

删除或重命名上述配置文件。

## 测试用例

### TC-1: 命令帮助

**目的**：验证帮助信息正确显示

**步骤**：
```bash
lc get-repo-wiki --help
```

**期望输出**：
- 显示命令描述："获取代码仓库的 Wiki 文档内容"
- 显示用法：包含参数格式说明
- 显示必需参数：owner, repo

### TC-2: Server 未配置报错

**目的**：验证执行时 Server 校验

**前提**：配置状态 B（Server 未配置）

**步骤**：
```bash
lc get-repo-wiki owner=torvalds repo=linux
```

**期望输出**：
```json
{
  "success": false,
  "error": {
    "code": "MCP_SERVER_NOT_CONFIGURED",
    "message": "命令 'get-repo-wiki' 需要 MCP Server 'openDeepWiki'，但未找到相关配置"
  }
}
```

### TC-3: 缺少必需参数

**目的**：验证参数校验

**前提**：配置状态 A（Server 已配置）

**步骤**：
```bash
lc get-repo-wiki repo=linux
```

**期望输出**：
```json
{
  "success": false,
  "error": {
    "code": "MISSING_ARG",
    "message": "缺少必需参数: owner"
  }
}
```

### TC-4: 正常调用（模拟）

**目的**：验证命令正确调用 MCP

**前提**：配置状态 A，且 MCP Server 可用

**步骤**：
```bash
lc get-repo-wiki owner=torvalds repo=linux
```

**期望输出**：
```json
{
  "success": true,
  "data": {
    "server": "openDeepWiki",
    "method": "GetWikiContents",
    "result": "..."
  }
}
```

### TC-5: 参数解析

**目的**：验证 key=value 解析

**步骤**：
```bash
lc get-repo-wiki owner=linux repo=kernel extra=param
```

**验证点**：额外参数应被传递给 MCP Server

### TC-6: 参数格式错误

**目的**：验证无效参数格式处理

**步骤**：
```bash
lc get-repo-wiki owner repo=linux
```

**期望输出**：解析错误提示

## E2E 测试

在 `e2e/` 目录添加测试用例：

```go
func TestMCPToolAlias(t *testing.T) {
    // TC-1: 帮助信息
    // TC-2: Server 未配置
    // TC-3: 缺少参数
}
```

## 回归测试

- [ ] `lc mcp` 原有功能不受影响
- [ ] `lc mcp openDeepWiki GetWikiContents` 仍可正常使用
