# lc detect 功能设计文档

## 1. 架构设计

### 1.1 整体流程

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  1. Git 探测     │────▶│  2. 仓库搜索匹配  │────▶│  3. 空间详情获取 │
└─────────────────┘     └──────────────────┘     └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
   - 检查 git 目录          - 按名称搜索            - 查询空间详情
   - 获取 remote URL        - URL 匹配              - 提取空间名称
   - 提取仓库名称           - 返回仓库信息          - 构建上下文
```

### 1.2 模块划分

| 模块 | 文件 | 职责 |
|-----|------|------|
| Git 探测 | `detect.go:probeGitContext()` | 检测 Git 仓库并提取基本信息 |
| 仓库匹配 | `detect.go:findMatchingRepository()` | 搜索并匹配远程仓库 |
| URL 标准化 | `detect.go:normalizeGitURL()` | 统一 Git URL 格式用于比较 |
| 上下文构建 | `detect.go:buildDetectedContext()` | 整合信息输出结果 |
| 空间服务 | `space.go:GetSpaceDetail()` | 获取研发空间详细信息 |

## 2. 详细设计

### 2.1 Git 探测模块

```go
type GitContext struct {
    IsGitRepo   bool   // 是否为 Git 仓库
    RepoName    string // 仓库名称（目录名）
    RemoteURL   string // Git remote origin URL
    GitPath     string // .git 目录路径
    CurrentPath string // 当前工作目录
}
```

**算法**：
1. 执行 `git rev-parse --git-dir` 检查是否为 Git 仓库
2. 执行 `git rev-parse --show-toplevel` 获取仓库根目录
3. 执行 `git remote get-url origin` 获取远程 URL

### 2.2 仓库匹配算法

**步骤**：
1. 使用 `ProjectService.SearchAllUserRepos()` 全局搜索仓库名称
2. 对每个返回的仓库，调用 `matchRepository()` 进行匹配
3. 优先使用 URL 精确匹配，其次使用名称匹配

**URL 匹配逻辑**：
```go
// 标准化后比较
normalizeGitURL(repo["tenantHttpPath"]) == normalizeGitURL(gitCtx.RemoteURL)
```

**标准化规则**：
- 移除 `.git` 后缀
- HTTP/HTTPS 统一为 HTTP
- 移除 `ssh://` 和 `git@` 前缀
- 移除端口号（如 `:8022`）
- 统一路径分隔符为 `/`
- 转小写

### 2.3 数据结构

```go
type DetectedContext struct {
    WorkspaceKey   string                 `json:"workspaceKey"`
    WorkspaceName  string                 `json:"workspaceName"`
    TenantID       string                 `json:"tenantId"`
    Repository     map[string]interface{} `json:"repository"`
    SpaceDetails   map[string]interface{} `json:"spaceDetails"`
    GitInfo        *GitContext            `json:"gitInfo"`
    Matched        bool                   `json:"matched"`
}
```

## 3. API 调用设计

### 3.1 搜索仓库
```
GET /moss/web/cmdevops-code/server/projects/user/page?name={repoName}
```

### 3.2 获取空间详情
```
GET /moss/web/cmdevops-platform/space/api/space/v1/maintain/page?spaceCode={workspaceKey}
```

## 4. 错误处理策略

| 场景 | 处理方式 | 返回值 |
|-----|---------|--------|
| 非 Git 目录 | 返回成功，matched=false | 原因：不在 Git 仓库中 |
| 无 remote URL | 按名称匹配 | - |
| 搜索失败 | 返回成功，matched=false | 原因：搜索失败 |
| 未找到仓库 | 返回成功，matched=false | 原因：未找到匹配仓库 |
| 获取空间详情失败 | 仅影响 spaceDetails 字段 | 其他字段正常返回 |

## 5. 性能优化

1. **并行查询**：Git 探测和 API 调用串行，无法并行
2. **缓存考虑**：暂不做缓存，每次实时查询确保准确性
3. **超时控制**：依赖 HTTP 客户端默认超时

## 6. 安全考虑

1. 仅读取 Git 配置，不修改任何文件
2. 使用已登录用户的 Cookie/Token 进行 API 调用
3. 不输出敏感信息（如完整 Cookie）
