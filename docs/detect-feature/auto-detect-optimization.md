# lc 命令自动探测优化方案

## 优化目标

让支持自动化的命令在未提供必要参数时，自动调用 `lc detect` 获取上下文，减少用户输入。

## 优化范围

### 第一阶段：高优先级命令（仅依赖 workspace-key/git-project-id）

| 命令 | 当前必传参数 | 优化后 |
|------|-------------|--------|
| `repo list` | `--workspace-key` | 可选，自动探测 |
| `repo disable-work-item-link` | `--workspace-key` | 可选，自动探测 |
| `repo delete` | `--workspace-key` | 可选，自动探测 |
| `req list` | `--workspace-key` | workspace-key 可选，name 自动从 API 获取（已实现） |
| `req view` | `--workspace-key` | 可选，自动探测 |
| `req delete` | `--workspace-key` | 可选，自动探测 |
| `req update` | `--workspace-key` | 可选，自动探测 |
| `req search` | `--workspace-key` | 自动探测（已实现） |
| `bug list` | `--workspace-key` | 可选，自动探测 |
| `bug view` | `--workspace-key` | 可选，自动探测 |
| `pr list` | `--workspace-key`, `--git-project-id` | 全部自动探测 |
| `pr view` | `--workspace-key`, `--git-project-id` | 全部自动探测 |

### 第二阶段：中优先级命令

| 命令 | 当前必传参数 | 优化方案 |
|------|-------------|---------|
| `pr create` | `--workspace-key`, `--git-project-id`, `--source`, `--title` | 前两个自动探测，分支自动获取当前 Git 分支 |
| `repo create` | `--workspace-key`, `--group-id` | workspace-key 自动探测 |
| `req create` | `--workspace-key`, `--project-code` | workspace-key 自动探测 |
| `task create` | `--workspace-key`, `--project-code` | workspace-key 自动探测 |

## 技术实现方案

### 方案一：通用参数解析器（推荐）

在 `common` 包中添加 `AutoDetectContext` 函数：

```go
// AutoDetectResult 缓存自动探测结果
type AutoDetectResult struct {
    WorkspaceKey   string
    WorkspaceName  string
    GitProjectId   string
    TenantId       string
    Matched        bool
    cacheTime      time.Time
}

var (
    detectCache     *AutoDetectResult
    detectCacheMu   sync.RWMutex
    cacheValidTime  = 5 * time.Minute
)

// GetAutoDetectContext 获取自动探测的上下文（带缓存）
func GetAutoDetectContext() (*AutoDetectResult, error) {
    detectCacheMu.RLock()
    if detectCache != nil && time.Since(detectCache.cacheTime) < cacheValidTime {
        result := detectCache
        detectCacheMu.RUnlock()
        return result, nil
    }
    detectCacheMu.RUnlock()

    // 执行实际探测逻辑
    result, err := performDetect()
    if err != nil {
        return nil, err
    }

    detectCacheMu.Lock()
    detectCache = result
    detectCacheMu.Unlock()

    return result, nil
}

// ClearDetectCache 清除探测缓存（用于测试或强制刷新）
func ClearDetectCache() {
    detectCacheMu.Lock()
    detectCache = nil
    detectCacheMu.Unlock()
}
```

### 方案二：命令参数预处理器

在每个命令的 `Run` 函数中添加参数预处理：

```go
// 示例：req list 命令
Run: func(cmd *cobra.Command, args []string) {
    // 自动探测缺失的参数
    if err := autoDetectMissingParams(cmd); err != nil {
        fmt.Fprintf(os.Stderr, "自动探测失败: %v\n", err)
        fmt.Fprintf(os.Stderr, "请手动指定 --workspace-key 参数\n")
        os.Exit(1)
    }
    listRequirements()
},

// autoDetectMissingParams 自动探测缺失的参数
func autoDetectMissingParams(cmd *cobra.Command) error {
    ctx, err := GetAutoDetectContext()
    if err != nil {
        return err
    }

    if !ctx.Matched {
        return fmt.Errorf("当前目录不在 Git 仓库中，或未能匹配到远程仓库")
    }

    // 检查并填充 workspace-key
    if !cmd.Flags().Changed("workspace-key") {
        reqWorkspaceKey = ctx.WorkspaceKey
    }

    // 检查并填充 workspace-name
    if !cmd.Flags().Changed("workspace-name") {
        reqWorkspaceName = ctx.WorkspaceName
    }

    return nil
}
```

### 方案三：全局标志控制

添加全局标志 `--no-auto-detect` 禁用自动探测：

```go
// root.go
var noAutoDetect bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&noAutoDetect, "no-auto-detect", false, "禁用自动探测功能")
}
```

## 具体命令改造示例

### 1. `req list` 改造

```go
var reqListCmd = &cobra.Command{
    Use:   "list",
    Short: "查询需求列表",
    Long: `查询当前研发空间下的需求列表。

自动探测:
  如果在 Git 仓库目录下执行，且未指定 --workspace-key，
  命令会自动探测当前目录所属的研发空间。

示例:
  # 自动探测并列出需求
  lc req list

  # 手动指定研发空间（只需 workspace-key，name 会自动获取）
  lc req list -w XXJSLJCLIDEV`,
    Run: func(cmd *cobra.Command, args []string) {
        // 尝试自动探测
        if err := tryAutoDetect(cmd); err != nil {
            fmt.Fprintf(os.Stderr, "错误: %v\n", err)
            fmt.Fprintf(os.Stderr, "提示: 请使用 -w 参数指定研发空间，或确保在 Git 仓库目录下执行\n")
            os.Exit(1)
        }
        listRequirements()
    },
}

func tryAutoDetect(cmd *cobra.Command) error {
    // 如果所有必需参数都已提供，跳过探测
    hasWorkspaceKey := cmd.Flags().Changed("workspace-key")
    hasWorkspaceName := cmd.Flags().Changed("workspace-name")

    if hasWorkspaceKey && hasWorkspaceName {
        return nil
    }

    // 执行自动探测
    ctx, err := common.GetAutoDetectContext()
    if err != nil {
        if !hasWorkspaceKey {
            return fmt.Errorf("无法自动探测研发空间: %w", err)
        }
        // 有 workspace-key 但探测失败，使用提供的参数
        return nil
    }

    if !ctx.Matched {
        if !hasWorkspaceKey {
            return fmt.Errorf("当前目录不在 Git 仓库中，无法自动探测")
        }
        return nil
    }

    // 填充缺失的参数
    if !hasWorkspaceKey {
        reqWorkspaceKey = ctx.WorkspaceKey
        fmt.Fprintf(os.Stderr, "已自动探测到研发空间: %s (%s)\n", ctx.WorkspaceName, ctx.WorkspaceKey)
    }
    if !hasWorkspaceName {
        reqWorkspaceName = ctx.WorkspaceName
    }

    return nil
}
```

### 2. `pr list` 改造

```go
var prListCmd = &cobra.Command{
    Use:   "list",
    Short: "查询 Merge Request 列表",
    Long: `查询指定仓库的 Merge Request 列表。

自动探测:
  如果在 Git 仓库目录下执行，会自动探测 git-project-id 和 workspace-key。

示例:
  # 自动探测并列出 MR
  lc pr list

  # 手动指定
  lc pr list --git-project-id 44645 -w XXJSLJCLIDEV`,
    Run: func(cmd *cobra.Command, args []string) {
        if err := tryAutoDetectPR(cmd); err != nil {
            fmt.Fprintf(os.Stderr, "错误: %v\n", err)
            os.Exit(1)
        }
        listPRs()
    },
}

func tryAutoDetectPR(cmd *cobra.Command) error {
    hasGitProjectId := cmd.Flags().Changed("git-project-id")
    hasWorkspaceKey := cmd.Flags().Changed("workspace-key")

    if hasGitProjectId && hasWorkspaceKey {
        return nil
    }

    ctx, err := common.GetAutoDetectContext()
    if err != nil {
        if !hasGitProjectId || !hasWorkspaceKey {
            return fmt.Errorf("无法自动探测仓库信息: %w", err)
        }
        return nil
    }

    if !ctx.Matched {
        return fmt.Errorf("当前目录不在 Git 仓库中")
    }

    if !hasGitProjectId {
        prGitProjectId = ctx.GitProjectId
    }
    if !hasWorkspaceKey {
        prWorkspaceKey = ctx.WorkspaceKey
    }

    return nil
}
```

## 用户体验优化

### 1. 显示自动探测信息

当自动探测成功时，在 stderr 输出提示信息：

```
$ lc req list
已自动探测到研发空间: 灵畿科研平台开发域 (XXJSSGKJCXPTYYTG)
[需求列表...]
```

### 2. 探测失败时友好提示

```
$ lc req list
错误: 当前目录不在 Git 仓库中，无法自动探测
提示: 请使用 -w 参数指定研发空间，或切换到 Git 仓库目录下执行
```

### 3. 添加 `--show-context` 调试标志

```bash
# 显示自动探测的上下文信息
lc req list --show-context
# 输出: 使用研发空间: XXJSSGKJCXPTYYTG, Git项目ID: 42231
```

## 实现优先级

### P0（立即实现）
1. `common` 包中添加 `GetAutoDetectContext()` 函数和缓存机制
2. 改造 `req list/view/delete/update/search`
3. 改造 `bug list/view/status/update-status/delete`

### P1（近期实现）
1. 改造 `pr list/view/comment/merge/review`
2. 改造 `repo list`
3. 添加 `--show-context` 调试标志

### P2（未来实现）
1. 改造 `pr create`（自动获取当前分支）
2. 改造 `req/task create`（交互式选择 project-code）
3. 添加全局 `--no-auto-detect` 标志

## 风险与注意事项

1. **性能影响**: 自动探测会增加一次 API 调用，但缓存机制可以缓解
2. **误判风险**: 确保在明确匹配失败时才报错，避免误判
3. **向后兼容**: 保持现有参数传递方式完全兼容，自动探测仅作为 fallback
4. **并发安全**: 缓存机制需要考虑并发访问的安全性
