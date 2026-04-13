# API 开发规范

本文档规定灵畿 CLI 中 API 客户端的设计规范，包括结构定义、请求构建、错误处理等。

## 目录

1. [文件组织](#文件组织)
2. [结构体定义](#结构体定义)
3. [API 方法实现](#api-方法实现)
4. [错误处理](#错误处理)
5. [响应解析](#响应解析)
6. [分页处理](#分页处理)
7. [最佳实践](#最佳实践)

---

## 文件组织

### 文件命名

- 按业务领域划分文件
- 使用单数名词

```
internal/api/
├── client.go           # HTTP 客户端
├── requirement.go      # 需求管理 API
├── bug.go              # 缺陷管理 API
├── merge_request.go    # 合并请求 API
├── project.go          # 项目管理 API
├── space.go            # 研发空间 API
├── doc.go              # 文档库 API
└── base.go             # 基础 API（用户、认证等）
```

### 包结构

每个文件应包含：

```go
package api

import (
    // 标准库
    "encoding/json"
    "fmt"
    "net/http"

    // 内部包
    "github.com/user/lc/internal/common"
)

// 1. 常量定义
const (
    DefaultPageSize = 20
    MaxPageSize     = 100
)

// 2. 请求/响应结构体
// 3. API 方法实现
```

---

## 结构体定义

### 请求结构体

```go
// CreateRequirementRequest 创建需求请求
type CreateRequirementRequest struct {
    // 必填字段
    Name         string   `json:"name"`
    WorkspaceKey string   `json:"workspaceKey"`
    ProjectCode  string   `json:"projectCode"`

    // 可选字段（omitempty）
    Description  string   `json:"description,omitempty"`
    Assignee     string   `json:"assignee,omitempty"`
    Priority     int      `json:"priority,omitempty"`
    Labels       []string `json:"labels,omitempty"`
}
```

### 响应结构体

```go
// RequirementResponse 需求响应
type RequirementResponse struct {
    // 标识字段
    ObjectID string `json:"objectId"`  // API 返回 camelCase
    Key      string `json:"key"`

    // 基本信息
    Name        string `json:"name"`
    Description string `json:"description"`
    Status      string `json:"status"`

    // 关联信息
    ProjectCode  string `json:"projectCode"`
    WorkspaceKey string `json:"workspaceKey"`

    // 时间戳
    CreatedAt string `json:"createdAt"`
    UpdatedAt string `json:"updatedAt"`

    // 创建人/负责人
    Creator  UserInfo `json:"creator"`
    Assignee UserInfo `json:"assignee,omitempty"`
}

type UserInfo struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Username string `json:"username"`
    Email    string `json:"email"`
}
```

### 列表响应结构体

```go
// ListRequirementsResponse 需求列表响应
type ListRequirementsResponse struct {
    Items      []RequirementResponse `json:"items"`
    Total      int                   `json:"total"`
    Page       int                   `json:"page"`
    PageSize   int                   `json:"pageSize"`
    TotalPages int                   `json:"totalPages"`
}
```

### 通用响应包装

```go
// APIResponse 通用 API 响应包装
type APIResponse struct {
    Success bool        `json:"success"`
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}
```

---

## API 方法实现

### 方法签名规范

```go
// 列表查询：返回列表和总数
func ListXXX(ctx *common.CommandContext, workspaceKey string, options ListOptions) ([]XXXResponse, int, error)

// 单条查询：返回详细信息
func GetXXX(ctx *common.CommandContext, id, workspaceKey string) (*XXXResponse, error)

// 创建：返回创建后的对象
func CreateXXX(ctx *common.CommandContext, req CreateXXXRequest) (*XXXResponse, error)

// 更新：返回更新后的对象
func UpdateXXX(ctx *common.CommandContext, id string, req UpdateXXXRequest) (*XXXResponse, error)

// 删除：只返回错误
func DeleteXXX(ctx *common.CommandContext, id, workspaceKey string) error

// 搜索：返回列表
func SearchXXX(ctx *common.CommandContext, keyword string, options SearchOptions) ([]XXXResponse, error)
```

### 完整示例：列表查询

```go
// ListRequirements 查询需求列表
//
// Parameters:
//   - ctx: 命令上下文，包含配置和 HTTP 客户端
//   - workspaceKey: 研发空间 Key
//   - page: 页码（从 1 开始）
//   - pageSize: 每页数量
//
// Returns:
//   - []RequirementResponse: 需求列表
//   - int: 总数
//   - error: 错误信息
func ListRequirements(ctx *common.CommandContext, workspaceKey string, page, pageSize int) ([]RequirementResponse, int, error) {
    // 1. 参数校验
    if workspaceKey == "" {
        return nil, 0, fmt.Errorf("workspaceKey is required")
    }
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > MaxPageSize {
        pageSize = DefaultPageSize
    }

    // 2. 构建 URL
    url := fmt.Sprintf("%s/requirement/list?page=%d&size=%d",
        ctx.Config.API.BaseReqURL, page, pageSize)

    // 3. 构建请求
    req := &Request{
        URL:     url,
        Method:  http.MethodGet,
        Headers: ctx.GetHeaders(workspaceKey),
    }

    // 4. 发送请求（带日志）
    ctx.Logger.Debug("Listing requirements",
        zap.String("url", url),
        zap.String("workspace", workspaceKey),
        zap.Int("page", page),
        zap.Int("pageSize", pageSize),
    )

    resp, err := ctx.Client.Send(req)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    // 5. 解析响应
    var apiResp struct {
        Success bool                   `json:"success"`
        Code    string                 `json:"code"`
        Message string                 `json:"message"`
        Data    ListRequirementsResponse `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, 0, fmt.Errorf("failed to decode response: %w", err)
    }

    // 6. 处理业务错误
    if !apiResp.Success {
        return nil, 0, &APIError{
            Code:    apiResp.Code,
            Message: apiResp.Message,
        }
    }

    return apiResp.Data.Items, apiResp.Data.Total, nil
}
```

### 完整示例：创建资源

```go
// CreateRequirement 创建需求
func CreateRequirement(ctx *common.CommandContext, req CreateRequirementRequest) (*RequirementResponse, error) {
    // 1. 参数校验
    if req.Name == "" {
        return nil, fmt.Errorf("name is required")
    }
    if req.WorkspaceKey == "" {
        return nil, fmt.Errorf("workspaceKey is required")
    }
    if req.ProjectCode == "" {
        return nil, fmt.Errorf("projectCode is required")
    }

    // 2. 构建 URL
    url := fmt.Sprintf("%s/requirement/create", ctx.Config.API.BaseReqURL)

    // 3. 构建请求
    apiReq := &Request{
        URL:     url,
        Method:  http.MethodPost,
        Headers: ctx.GetHeaders(req.WorkspaceKey),
        Body:    req,
    }

    ctx.Logger.Debug("Creating requirement",
        zap.String("name", req.Name),
        zap.String("projectCode", req.ProjectCode),
    )

    // 4. 发送请求
    resp, err := ctx.Client.Send(apiReq)
    if err != nil {
        return nil, fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    // 5. 检查 HTTP 状态码
    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(body))
    }

    // 6. 解析响应
    var apiResp struct {
        Success bool                `json:"success"`
        Data    RequirementResponse `json:"data"`
        Message string              `json:"message"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    if !apiResp.Success {
        return nil, &APIError{
            Code:    "CREATE_FAILED",
            Message: apiResp.Message,
        }
    }

    return &apiResp.Data, nil
}
```

---

## 错误处理

### 自定义错误类型

```go
// APIError 表示 API 返回的业务错误
type APIError struct {
    Code       string // 错误代码
    Message    string // 错误消息
    StatusCode int    // HTTP 状态码
}

func (e *APIError) Error() string {
    if e.Code != "" {
        return fmt.Sprintf("[%s] %s", e.Code, e.Message)
    }
    return e.Message
}

// IsNotFound 检查是否为资源不存在错误
func IsNotFound(err error) bool {
    if apiErr, ok := err.(*APIError); ok {
        return apiErr.Code == "NOT_FOUND" || apiErr.StatusCode == http.StatusNotFound
    }
    return false
}

// IsUnauthorized 检查是否为未授权错误
func IsUnauthorized(err error) bool {
    if apiErr, ok := err.(*APIError); ok {
        return apiErr.StatusCode == http.StatusUnauthorized
    }
    return false
}
```

### 错误包装

```go
result, err := api.CreateRequirement(ctx, req)
if err != nil {
    // 添加上下文信息
    return nil, fmt.Errorf("创建需求 '%s' 失败: %w", req.Name, err)
}
```

---

## 响应解析

### 解析嵌套 JSON

```go
// 处理复杂的嵌套结构
var response struct {
    Success bool `json:"success"`
    Data    struct {
        Items []struct {
            ObjectID    string                 `json:"objectId"`
            Name        string                 `json:"name"`
            CustomFields map[string]interface{} `json:"customFields"`
        } `json:"items"`
        Pagination struct {
            Page     int `json:"page"`
            PageSize int `json:"pageSize"`
            Total    int `json:"total"`
        } `json:"pagination"`
    } `json:"data"`
}
```

### 动态字段处理

```go
// 处理不确定的字段
type FlexibleResponse struct {
    ID      string                 `json:"id"`
    Name    string                 `json:"name"`
    Extra   map[string]interface{} `json:"extra"`  // 动态字段
}

// 使用
if extra, ok := resp.Extra["custom_field"]; ok {
    // 处理动态字段
}
```

### 时间字段解析

```go
import "time"

type TimeResponse struct {
    // 使用时间指针，允许为空
    CreatedAt *time.Time `json:"createdAt"`
}

// 或使用自定义时间格式
type CustomTime struct {
    time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
    // 处理特定格式的时间字符串
    s := string(b)
    s = strings.Trim(s, `"`)

    parsed, err := time.Parse("2006-01-02 15:04:05", s)
    if err != nil {
        return err
    }
    ct.Time = parsed
    return nil
}
```

---

## 分页处理

### 分页选项

```go
// ListOptions 通用列表选项
type ListOptions struct {
    Page       int
    PageSize   int
    SortBy     string
    SortOrder  string // "asc" 或 "desc"
    Filters    map[string]string
}

// DefaultListOptions 返回默认选项
func DefaultListOptions() ListOptions {
    return ListOptions{
        Page:     1,
        PageSize: 20,
        SortBy:   "createdAt",
        SortOrder: "desc",
    }
}
```

### 分页查询实现

```go
// ListAllRequirements 查询所有需求（自动处理分页）
func ListAllRequirements(ctx *common.CommandContext, workspaceKey string) ([]RequirementResponse, error) {
    var allItems []RequirementResponse
    page := 1
    pageSize := 100

    for {
        items, total, err := ListRequirements(ctx, workspaceKey, page, pageSize)
        if err != nil {
            return nil, err
        }

        allItems = append(allItems, items...)

        // 检查是否已获取所有数据
        if len(allItems) >= total {
            break
        }

        page++

        // 安全限制
        if page > 100 {
            ctx.Logger.Warn("Too many pages, stopping pagination",
                zap.Int("currentPage", page),
                zap.Int("itemsCollected", len(allItems)),
            )
            break
        }
    }

    return allItems, nil
}
```

---

## 最佳实践

### 1. 参数校验

始终在方法开始处校验参数：

```go
func SomeAPI(ctx *common.CommandContext, workspaceKey string) error {
    if ctx == nil {
        return fmt.Errorf("ctx is nil")
    }
    if workspaceKey == "" {
        return fmt.Errorf("workspaceKey is required")
    }
    // ...
}
```

### 2. 超时控制

为长时间运行的操作添加上下文：

```go
func LongRunningOperation(ctx context.Context, ...) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    req := &Request{
        URL:    url,
        Method: http.MethodPost,
        // ...
    }

    // 使用上下文发送请求
    // ...
}
```

### 3. 日志记录

记录关键信息，便于调试：

```go
ctx.Logger.Debug("API request",
    zap.String("method", req.Method),
    zap.String("url", req.URL),
    zap.Int("statusCode", resp.StatusCode),
    zap.Duration("duration", elapsed),
)
```

### 4. 重试机制

对可能失败的请求添加重试：

```go
func SendWithRetry(ctx *common.CommandContext, req *Request, maxRetries int) (*http.Response, error) {
    var resp *http.Response
    var err error

    for i := 0; i < maxRetries; i++ {
        resp, err = ctx.Client.Send(req)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }

        if resp != nil {
            resp.Body.Close()
        }

        // 指数退避
        time.Sleep(time.Second * time.Duration(i+1))
    }

    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}
```

### 5. 响应缓存

对不频繁变化的数据添加缓存：

```go
var (
    spaceCache   map[string]SpaceResponse
    spaceCacheMu sync.RWMutex
    cacheExpiry  = 5 * time.Minute
    lastCacheTime time.Time
)

func GetSpaceWithCache(ctx *common.CommandContext, workspaceKey string) (*SpaceResponse, error) {
    spaceCacheMu.RLock()
    if spaceCache != nil && time.Since(lastCacheTime) < cacheExpiry {
        if space, ok := spaceCache[workspaceKey]; ok {
            spaceCacheMu.RUnlock()
            return &space, nil
        }
    }
    spaceCacheMu.RUnlock()

    // 缓存未命中，从 API 获取
    space, err := GetSpace(ctx, workspaceKey)
    if err != nil {
        return nil, err
    }

    // 更新缓存
    spaceCacheMu.Lock()
    if spaceCache == nil {
        spaceCache = make(map[string]SpaceResponse)
    }
    spaceCache[workspaceKey] = *space
    lastCacheTime = time.Now()
    spaceCacheMu.Unlock()

    return space, nil
}
```

---

## 参考实现

查看以下文件获取完整示例：

- `internal/api/requirement.go` - 完整的 CRUD API
- `internal/api/bug.go` - 带搜索功能的 API
- `internal/api/space.go` - 缓存示例
- `internal/api/merge_request.go` - 复杂业务逻辑
