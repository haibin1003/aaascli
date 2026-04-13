# 自动探测功能开发指南

本文档介绍如何在 `lc` CLI 工具中实现和使用 Git 仓库上下文自动探测功能。

## 概述

自动探测功能允许命令在 Git 仓库目录下自动识别所属的研发空间信息，无需用户手动指定 `--workspace-key` 等参数。

## 核心组件

### 1. 自动探测工具 (`internal/common/autodetect.go`)

```go
// 核心数据结构
type AutoDetectContext struct {
    WorkspaceKey   string                 // 研发空间代码 (如: XXJSLJCLIDEV)
    WorkspaceName  string                 // 研发空间名称
    TenantID       string                 // 租户ID
    GitProjectID   string                 // Git项目ID
    Repository     map[string]interface{} // 仓库详情
    SpaceDetails   map[string]interface{} // 空间详情
    GitInfo        *GitInfo               // Git信息
    Matched        bool                   // 是否匹配成功
    MatchReason    string                 // 匹配失败原因
}

// 探测结果
type AutoDetectResult struct {
    Success bool
    Context *AutoDetectContext
    Error   error
}
```

### 2. 关键函数

| 函数 | 用途 | 缓存 |
|------|------|------|
| `GetAutoDetectContext()` | 获取探测上下文 | 5分钟TTL |
| `TryAutoDetect(requireGitRepo bool)` | 尝试自动探测 | 使用上述缓存 |
| `PrintAutoDetectInfo()` | 打印探测信息 | - |
| `ClearAutoDetectCache()` | 清除缓存 | - |

## 如何为新命令添加自动探测支持

### 步骤1: 修改命令定义

在命令的 `Long` 描述中添加自动探测说明：

```go
var myListCmd = &cobra.Command{
    Use:   "list",
    Short: "查询XXX列表",
    Long: `查询XXX列表。

自动探测:
  如果在 Git 仓库目录下执行，且未指定 --workspace-key，
  命令会自动探测当前目录所属的研发空间。

示例:
  # 自动探测
  lc my list

  # 手动指定
  lc my list -w XXJSLJCLIDEV`,
    Run: func(cmd *cobra.Command, args []string) {
        // 尝试自动探测
        if err := tryAutoDetectForMyList(cmd); err != nil {
            fmt.Fprintf(os.Stderr, "\n[错误] %v\n\n", err)
            fmt.Fprintln(os.Stderr, "请使用以下方式之一解决:")
            fmt.Fprintln(os.Stderr, "  1. 在 Git 仓库目录下执行命令")
            fmt.Fprintln(os.Stderr, "  2. 手动指定参数:")
            fmt.Fprintln(os.Stderr, "     -w, --workspace-key     研发空间 Key")
            fmt.Fprintln(os.Stderr, "\n使用 'lc my list --help' 查看更多参数信息")
            os.Exit(1)
        }
        listMyItems()
    },
}
```

### 步骤2: 修改 Flag 定义

移除必填标记，更新描述：

```go
// 修改前
myListCmd.Flags().StringVarP(&myWorkspaceKey, "workspace-key", "w", "", "研发空间 key（必填）")
myListCmd.MarkFlagRequired("workspace-key")

// 修改后
myListCmd.Flags().StringVarP(&myWorkspaceKey, "workspace-key", "w", "", "研发空间 key（可选，支持自动探测）")
```

### 步骤3: 实现自动探测函数

```go
// tryAutoDetectForMyList 尝试为 my list 命令自动探测参数
func tryAutoDetectForMyList(cmd *cobra.Command) error {
    // 检查用户是否已指定 workspace-key
    workspaceKeySet := cmd.Flags().Changed("workspace-key")

    // 如果参数已指定，无需自动探测
    if workspaceKeySet {
        return nil
    }

    // 尝试自动探测
    result := common.TryAutoDetect(true)
    if !result.Success {
        return result.Error
    }

    ctx := result.Context

    // 填充未指定的参数
    if !workspaceKeySet && ctx.WorkspaceKey != "" {
        myWorkspaceKey = ctx.WorkspaceKey
        fmt.Fprintf(os.Stderr, "[自动探测] 研发空间 Key: %s\n", myWorkspaceKey)
    }

    return nil
}
```

### 步骤4: 更新父命令描述

在父命令中添加自动探测支持说明：

```go
var myCmd = &cobra.Command{
    Use:   "my",
    Short: "管理XXX",
    Long: `管理XXX功能。

自动探测支持:
  list 命令支持自动探测研发空间。
  在 Git 仓库目录下执行时，无需手动指定 --workspace-key。`,
}
```

## 设计原则

### 1. 向后兼容
- 自动探测是**可选**功能，不是强制要求
- 用户手动指定的参数优先级最高
- 探测失败时提供清晰的错误提示和手动指定方式

### 2. 缓存机制
- 探测结果缓存5分钟，避免重复调用 API
- 使用 `sync.RWMutex` 保证线程安全
- 提供 `ClearAutoDetectCache()` 用于测试和调试

### 3. 错误处理
- 探测失败时返回错误，不静默处理
- 错误信息包含解决建议
- 提供分级错误提示（命令级 + 父命令级）

## 最佳实践

### 1. 参数检查顺序
```go
// 正确：先检查用户是否已指定
if workspaceKeySet {
    return nil  // 用户已指定，跳过探测
}

// 尝试自动探测
result := common.TryAutoDetect(true)
```

### 2. 部分参数支持
某些命令可能只需要 `workspace-key`，不需要 `workspace-name`：

```go
// 只检查 workspace-key
workspaceKeySet := cmd.Flags().Changed("workspace-key")

// 探测后只填充需要的参数
if !workspaceKeySet && ctx.WorkspaceKey != "" {
    myWorkspaceKey = ctx.WorkspaceKey
}
```

### 3. 调试支持
```go
// 使用调试模式查看探测过程
lc my list -d

// 手动清除缓存（在代码中）
common.ClearAutoDetectCache()
```

## 完整的实现示例

参考以下已实现命令：

### `lc req list` (`cmd/lc/req.go`)
- 需要 `workspace-key` 和 `workspace-name`
- 完整实现了所有步骤

### `lc bug list` (`cmd/lc/bug.go`)
- 只需要 `workspace-key`
- 简化的探测函数示例

## 常见问题

### Q: 自动探测和手动参数的关系？
A: 手动参数优先级最高。如果用户指定了参数，自动探测会被跳过。

### Q: 不在 Git 仓库目录会怎样？
A: 会返回错误，提示用户手动指定参数或在 Git 仓库目录下执行。

### Q: 如何测试自动探测功能？
A:
1. 在 Git 仓库目录下直接执行命令
2. 在非 Git 目录下执行，验证错误提示
3. 使用 `--workspace-key` 手动指定，验证参数优先级

### Q: 缓存何时失效？
A:
- 5分钟后自动失效
- 进程重启后失效
- 调用 `ClearAutoDetectCache()` 立即失效

## 扩展建议

### 1. 支持更多探测源
可以扩展 `performDetect()` 函数，支持从环境变量、配置文件等获取上下文：

```go
// 检查环境变量
if workspaceKey := os.Getenv("LC_WORKSPACE_KEY"); workspaceKey != "" {
    ctx.WorkspaceKey = workspaceKey
    ctx.Matched = true
    return ctx, nil
}
```

### 2. 交互式选择
当探测到多个可能的研发空间时，可以提示用户选择：

```go
if len(possibleSpaces) > 1 {
    // 交互式选择或返回错误提示
}
```

### 3. 保存默认空间
可以添加 `lc config set-default-workspace` 命令，将探测结果保存为默认值。

## 相关文档

- [需求文档](./requirements.md)
- [设计文档](./design.md)
- [实现总结](./implementation.md)
- [参数映射对照表](./parameter-mapping.md)
