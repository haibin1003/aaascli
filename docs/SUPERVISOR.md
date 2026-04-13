# Supervisor 监管模块开发指南

本文档介绍如何开发 `lc supervisor` 子命令，以及监管平台的认证流程。

## 目录

- [认证流程概述](#认证流程概述)
- [快速开始](#快速开始)
- [开发新子命令](#开发新子命令)
- [API 参考](#api-参考)

---

## 认证流程概述

监管平台(supervision)的所有 API 都需要特殊的认证机制。与灵畿平台其他模块不同，supervisor 模块需要**两步认证**才能获取有效的访问凭证。

### 认证步骤

**重要：所有 supervisor 子命令都必须按以下顺序执行认证，否则无法访问 API。**

```
┌─────────────────────────────────────────────────────────────────┐
│  Step 1: 请求 MOSS 认证接口获取跳转 URL 和 code                    │
│  POST http://4c.hq.cmcc/moss/web/auth/v1/user/oauth/authorize/   │
│           jump-url                                               │
│  Response: {                                                     │
│    "data": "http://4c.hq.cmcc/supervision/#/plugin-auth?code=xxx"│
│  }                                                               │
└────────────────────────────┬────────────────────────────────────┘
                             │ 提取 code
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  Step 2: 使用 code 换取 Authentication token                      │
│  POST http://4c.hq.cmcc/supervision/system-api/business-accept/  │
│           supervision/authorize                                  │
│  Response Header:                                                │
│  Set-Cookie: Authentication=xxx; Max-Age=86400; ...              │
│                                                                  │
│  提取 Authentication 值用于后续请求                               │
└─────────────────────────────────────────────────────────────────┘
```

### 认证封装

认证流程已封装在 `internal/api/supervisor.go` 的 `SupervisorService` 中：

```go
// 创建服务实例
svc := api.NewSupervisorService(cookie, client)

// 执行认证（自动完成两步认证）
authToken, err := svc.Authenticate()
if err != nil {
    return fmt.Errorf("认证失败: %w", err)
}
```

---

## 快速开始

### 查询待办数量

```bash
# 先登录获取 MOSS_SESSION
lc login "your-moss-session-cookie"

# 查询监管平台待办数量
lc supervisor todo list
```

输出示例：
```json
{
  "组内待认领工单": {
    "问题咨询": "0",
    "普通投诉": "0",
    "升级投诉": "0",
    "意见反馈": "0"
  },
  "我的待办": {
    "问题咨询": "0",
    "普通投诉": "0",
    "升级投诉": "0",
    "意见反馈": "0"
  }
}
```

### 查询工作组列表

```bash
# 查询全量工作组列表
lc supervisor groups list

# 查询第2页，每页20条
lc supervisor groups list --page 2 --size 20

# 按编码筛选
lc supervisor groups list --code it001

# 按名称筛选
lc supervisor groups list --name "知识库"

# 组合筛选
lc supervisor groups list -p 1 -s 10 -c it001 -n "知识库"
```

### 查询工作组成员

```bash
# 查询指定工作组的成员（使用工作组编码）
lc supervisor groups members --group-code it010

# 自动查询名称包含 "1组" 的所有工作组成员
lc supervisor groups members --group-name-filter "1组"

# 输出到CSV
lc supervisor groups members --group-name-filter "1组" -o members.csv
```

---

## 开发新子命令

### 步骤 1: 在 internal/api/supervisor.go 中添加 API 方法

```go
// YourAPIResponse API 响应结构
type YourAPIResponse struct {
    Result   interface{} `json:"result"`
    RespCode string      `json:"respCode"`
    RespDesc string      `json:"respDesc"`
}

// YourAPIMethod 你的 API 方法
// 注意：调用前必须先执行 Authenticate() 获取认证
func (s *SupervisorService) YourAPIMethod(param string) (*YourAPIResponse, error) {
    // 1. 获取认证头信息
    headers, err := s.GetAuthHeaders()
    if err != nil {
        return nil, err
    }

    // 2. 构建请求
    req := &Request{
        URL:     SupervisorBaseURL + "/supervision/system-api/your/api/path",
        Method:  "POST",
        Headers: headers,
        Body: map[string]interface{}{
            "param": param,
        },
    }

    // 3. 发送请求
    resp, err := s.HTTPClient.Send(req)
    if err != nil {
        return nil, fmt.Errorf("请求失败: %w", err)
    }
    defer resp.Body.Close()

    // 4. 解析响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("读取响应失败: %w", err)
    }

    var result YourAPIResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("解析响应失败: %w", err)
    }

    if result.RespCode != "0000" {
        return nil, fmt.Errorf("API 调用失败: %s", result.RespDesc)
    }

    return &result, nil
}
```

### 步骤 2: 在 cmd/lc/supervisor.go 中添加子命令

```go
func init() {
    // ... 已有的命令注册 ...
    supervisorCmd.AddCommand(yourNewCmd)
}

// yourNewCmd 新子命令定义
var yourNewCmd = &cobra.Command{
    Use:   "your-cmd [args]",
    Short: "命令简短描述",
    Long: `命令详细描述。

示例:
  lc supervisor your-cmd arg1`,
    Run: func(cmd *cobra.Command, args []string) {
        common.Execute(execYourNewCmd, common.ExecuteOptions{
            DebugMode: debugMode,
            Insecure:  insecureSkipVerify,
            Logger:    &logger,
        })
    },
}

// execYourNewCmd 执行命令逻辑
func execYourNewCmd() error {
    // 1. 加载配置
    configPath := config.GetDefaultConfigPath()
    cfg, err := config.LoadConfigWithDefaults(configPath)
    if err != nil {
        return fmt.Errorf("加载配置失败: %w", err)
    }

    if cfg.Cookie == "" {
        return fmt.Errorf("未登录，请先执行: lc login <cookie-value>")
    }

    // 2. 创建 HTTP 客户端
    var client *api.Client
    if insecureSkipVerify {
        client = api.NewInsecureClient()
    } else {
        client = api.NewClient()
    }

    // 3. 创建服务并执行认证（关键步骤！）
    svc := api.NewSupervisorService(cfg.Cookie, client)

    if debugMode {
        fmt.Fprintln(os.Stderr, "[debug] 开始认证流程...")
    }

    _, err = svc.Authenticate()
    if err != nil {
        return fmt.Errorf("认证失败: %w", err)
    }

    // 4. 调用 API
    result, err := svc.YourAPIMethod("param")
    if err != nil {
        return fmt.Errorf("调用失败: %w", err)
    }

    // 5. 输出结果
    output, err := json.MarshalIndent(result, "", "  ")
    if err != nil {
        return fmt.Errorf("格式化输出失败: %w", err)
    }

    fmt.Println(string(output))
    return nil
}
```

### 关键要点

1. **必须认证**：所有 supervisor API 调用前都必须执行 `svc.Authenticate()`
2. **使用 GetAuthHeaders()**：API 调用时使用 `GetAuthHeaders()` 获取包含认证信息的请求头
3. **错误处理**：统一返回错误，由 `common.Execute` 处理错误输出
4. **调试模式**：使用 `debugMode` 标志控制是否输出调试信息

---

## API 参考

### SupervisorService

监管平台服务的主结构，封装了认证流程和 HTTP 请求。

#### 构造函数

```go
func NewSupervisorService(cookie string, client HTTPClient) *SupervisorService
```

- `cookie`: MOSS_SESSION cookie 值
- `client`: HTTP 客户端（通常是 `api.NewInsecureClient()` 或 `api.NewClient()`）

#### 方法

##### Authenticate

```go
func (s *SupervisorService) Authenticate() (string, error)
```

执行完整的认证流程，返回 Authentication token。

**必须在访问任何 API 前调用。**

##### GetAuthHeaders

```go
func (s *SupervisorService) GetAuthHeaders() (map[string]string, error)
```

获取包含认证信息的请求头。必须在 `Authenticate()` 成功后调用。

##### GetMainOverview

```go
func (s *SupervisorService) GetMainOverview() (*MainOverviewResponse, error)
```

获取主面板概览数据（待办数量统计）。

##### GetGroupList

```go
func (s *SupervisorService) GetGroupList(current, pageSize int, code, name, description string) (*GroupListResponse, error)
```

获取业务分组列表，支持分页和筛选。

参数：
- `current`: 当前页码
- `pageSize`: 每页大小
- `code`: 工作组编码筛选（可选）
- `name`: 工作组名称筛选（可选）
- `description`: 工作组描述筛选（可选）

##### GetGroupMembersByCode

```go
func (s *SupervisorService) GetGroupMembersByCode(groupCode string) (*GroupMemberByGroupCodeResponse, error)
```

获取工作组成员列表（通过工作组编码）。

参数：
- `groupCode`: 工作组编码（必填）

##### GetGroupMembers

```go
func (s *SupervisorService) GetGroupMembers(deptId string) (*GroupMemberByGroupCodeResponse, error)
```

获取工作组成员列表（兼容函数，将 deptId 作为 groupCode 使用）。

参数：
- `deptId`: 工作组编码（必填）

##### GetGroupList

```go
func (s *SupervisorService) GetGroupList(current, pageSize int, code, name, description string) (*GroupListResponse, error)
```

获取业务分组列表，支持分页和筛选。

参数：
- `current`: 当前页码
- `pageSize`: 每页大小
- `code`: 工作组编码筛选（可选）
- `name`: 工作组名称筛选（可选）
- `description`: 工作组描述筛选（可选）

### 常量

```go
const (
    SupervisorBaseURL = "http://4c.hq.cmcc"                                        // 监管平台基础 URL
    MOSSAuthURL       = "http://4c.hq.cmcc/moss/web/auth/v1/user/oauth/authorize/jump-url" // MOSS 认证 URL
    SupervisorAuthURL = "http://4c.hq.cmcc/supervision/system-api/business-accept/supervision/authorize" // 监管平台认证 URL
)
```

---

## 示例：完整的子命令实现

参考 `cmd/lc/supervisor.go` 中的 `supervisorTodoListCmd` 实现，它展示了：

1. 配置加载
2. 客户端创建
3. 认证流程
4. API 调用
5. 结果格式化输出

---

## 注意事项

1. **Cookie 来源**：`MOSS_SESSION` 从 `lc login` 设置的配置中获取
2. **认证有效期**：Authentication token 有效期为 24 小时（Max-Age=86400）
3. **错误处理**：所有错误都应包装后返回，不要直接打印到 stderr
4. **调试信息**：使用 `fmt.Fprintln(os.Stderr, ...)` 输出调试信息，不要干扰正常输出
