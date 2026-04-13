# Service Layer

服务层提供业务逻辑封装，位于命令层和 API 客户端层之间。

## 架构

```
cmd/lc/              # 命令层 - 处理 CLI 参数和输出
  └── req.go
  └── task.go
  └── ...

internal/service/    # 服务层 - 业务逻辑
  ├── response_formatter.go   # 响应格式化
  ├── iql_service.go          # IQL 查询构建
  ├── dryrun_service.go       # 试运行模拟
  ├── workspace_service.go    # 工作空间服务
  └── domain/                 # 领域服务
      ├── requirement_service.go
      ├── task_service.go
      └── bug_service.go

internal/api/        # API 客户端层 - HTTP 通信
  └── client.go
```

## 使用示例

### 使用 IQLService 构建查询

```go
iqlService := service.NewIQLService()

// 构建需求列表查询
query := iqlService.BuildRequirementListQuery("MyWorkspace", "search keyword")
// 输出: (所属空间 = 'MyWorkspace' and 类型 in ["需求"] and 名称包含 'search keyword') order by 创建时间 desc

// 使用 QueryBuilder 构建复杂查询
qb := iqlService.NewQuery()
query := qb.WithWorkspace("MyWorkspace").
    WithTypeIn([]string{"需求", "任务"}).
    WithKeyword("bug").
    WithStatus("进行中").
    OrderByCreateTime().
    Build()
```

### 使用 ResponseFormatter 格式化响应

```go
formatter := service.NewResponseFormatter()

// 格式化需求列表
items := []map[string]interface{}{...}  // 来自 API 的原始数据
formatted := formatter.FormatRequirementList(items)

// 格式化单个需求
requirement := map[string]interface{}{...}
formatted := formatter.FormatRequirement(requirement)
```

### 使用 DryRunService 模拟操作

```go
dryRunService := service.NewDryRunService()

// 模拟创建
result := dryRunService.SimulateCreate(
    service.ResRequirement,
    "New Feature",
    map[string]interface{}{
        "priority": "高",
    },
)
// 输出: {"dryRun": true, "action": "create", "resource": "requirement", ...}

// 模拟删除
result := dryRunService.SimulateDelete(
    service.ResTask,
    []string{"task-1", "task-2"},
)
```

### 使用领域服务

```go
// 创建工作空间服务
workspaceService := service.NewWorkspaceService(
    baseURL, headers, client, config,
)

// 创建需求服务
reqService := domain.NewRequirementService(
    baseURL, headers, client, config, workspaceService,
)

// 创建需求
resp, err := reqService.Create(&domain.CreateRequirementRequest{
    Name:         "New Feature",
    Description:  "Feature description",
    WorkspaceKey: "MYSPACE",
    ProjectCode:  "PROJ001",
})
```

## 迁移指南

### 从命令层直接调用 API 迁移到使用 Service 层

**Before:**
```go
// cmd/lc/req.go
func createRequirement(cmd *cobra.Command, args []string) {
    headers := ctx.GetHeaders(workspaceKey)
    body := buildRequestBody(name, desc)
    resp, err := client.Post(url, headers, body)
    // ... 处理响应
}
```

**After:**
```go
// cmd/lc/req.go
func createRequirement(cmd *cobra.Command, args []string) {
    workspaceService := service.NewWorkspaceService(...)
    reqService := domain.NewRequirementService(..., workspaceService)

    if dryRun {
        result := dryRunService.SimulateCreate(ResRequirement, name, details)
        return result
    }

    resp, err := reqService.Create(&domain.CreateRequirementRequest{...})

    formatter := service.NewResponseFormatter()
    formatted := formatter.FormatRequirement(resp)
    return formatted
}
```

## 服务职责

| 服务 | 职责 |
|------|------|
| IQLService | 构建 IQL 查询字符串 |
| ResponseFormatter | 统一格式化 API 响应 |
| DryRunService | 生成试运行模拟结果 |
| WorkspaceService | 工作空间解析和缓存 |
| RequirementService | 需求业务逻辑 |
| TaskService | 任务业务逻辑 |
| BugService | 缺陷业务逻辑 |

## 测试

```bash
# 运行所有服务层测试
go test ./internal/service/... -v

# 运行短测试（跳过集成测试）
go test ./internal/service/... -v -short
```
