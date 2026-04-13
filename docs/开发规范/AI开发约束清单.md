# AI开发约束清单

**本文档面向AI助手，规定了开发前、开发中、开发后必须遵守的强制性约束。**

---

## 开发前必读（强制）

### 第一步：阅读顺序

**必须按以下顺序阅读文档：**

1. ✅ [README.md](./README.md) - 了解文档结构
2. ✅ [01-总体架构.md](./01-总体架构.md) - 理解架构设计
3. ✅ [02-命令命名规范.md](./02-命令命名规范.md) - 掌握命名规则
4. ✅ [04-Go代码规范.md](./04-Go代码规范.md) - 了解编码规范
5. ✅ [本清单](#) - 了解约束条件

**在开始编码前，必须确认已阅读以上所有文档。**

### 第二步：理解现有代码

**在添加新功能前，必须：**

- [ ] 查看至少2个现有命令的实现（如 `cmd/lc/req.go` 和 `cmd/lc/bug.go`）
- [ ] 理解 `common.Execute()` 的使用方式
- [ ] 理解 `internal/api/` 的结构

**禁止行为：**
- ❌ 不阅读现有代码就开始编写
- ❌ 凭空创造新的代码结构
- ❌ 忽略现有的工具函数，自己重新实现

---

## 强制性架构约束

### 1. 命令实现约束

**必须遵守：**

```go
// ✅ 正确的结构
var xxxCmd = &cobra.Command{
    Use:   "create",
    Short: "简短描述",
    Long:  `详细描述`,
    Run: func(cmd *cobra.Command, args []string) {
        // 1. 获取显式参数
        workspaceKey, _ := cmd.Flags().GetString("workspace-key")

        // 2. 自动探测（如果参数为空）
        if workspaceKey == "" {
            autoResult := common.TryAutoDetect(true)
            if autoResult.Success {
                workspaceKey = autoResult.Context.WorkspaceKey
            } else {
                common.PrintAutoDetectError(autoResult.Error)
                os.Exit(1)
            }
        }

        // 3. 使用 common.Execute 执行
        common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
            // 调用 API
            return api.DoSomething(ctx, ...)
        }, common.ExecuteOptions{
            DebugMode: debugMode,
            Insecure:  insecureSkipVerify,
            Logger:    &logger,
        })
    },
}
```

**禁止行为：**
- ❌ 不使用 `common.Execute()`，直接输出结果
- ❌ 不实现自动探测功能
- ❌ 自行处理配置加载和HTTP客户端创建
- ❌ 不使用统一的JSON输出格式

### 2. API客户端约束

**必须遵守：**

```go
// ✅ 正确的API方法签名
func DoSomething(ctx *common.CommandContext, params ...) (ResultType, error) {
    // 1. 参数校验
    if workspaceKey == "" {
        return nil, fmt.Errorf("workspaceKey is required")
    }

    // 2. 构建请求
    url := fmt.Sprintf("%s/...", ctx.Config.API.BaseXXXURL)
    req := &Request{...}

    // 3. 发送请求
    resp, err := ctx.Client.Send(req)
    if err != nil {
        return nil, fmt.Errorf("...: %w", err)
    }
    defer resp.Body.Close()

    // 4. 解析响应
    // ...

    return result, nil
}
```

**禁止行为：**
- ❌ 不接收 `*common.CommandContext` 参数
- ❌ 自行创建 HTTP 客户端
- ❌ 不处理错误，直接 panic
- ❌ 不 defer resp.Body.Close()

### 3. 输出格式约束

**必须遵守的统一输出格式：**

```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "timestamp": "2026-03-15T10:30:00Z",
    "executionTime": 123,
    "version": "v1.2.3"
  }
}
```

**禁止行为：**
- ❌ 使用 `fmt.Println()` 直接输出
- ❌ 使用 `log.Println()` 输出结果
- ❌ 不包装直接返回 API 原始响应
- ❌ 添加自定义的输出格式

---

## 命名约束（强制）

### 命令命名

| 类型 | 规则 | 示例 |
|------|------|------|
| 父命令 | 名词，小写 | `req`, `bug`, `repo` |
| 子命令 | 动词 | `list`, `create`, `view` |
| 三级命令 | 名词+动词 | `space project list` |

**必须使用已有的命名：**
- `workspace-key` / `-w` - 研发空间Key
- `workspace-name` - 研发空间名称
- `project-code` / `-p` - 项目代码
- `git-project-id` - Git项目ID

**禁止：**
- ❌ 创建新的参数名表示相同概念
- ❌ 使用复数形式（`reqs`）
- ❌ 使用下划线（`workspace_key`）

### 变量命名

```go
// ✅ 正确
workspaceKey := "XXJSxiaobaice"
objectID := "xxx"
gitProjectID := 123

// ❌ 错误
wk := "XXJSxiaobaice"           // 过于简短
workspace_key := ""              // 下划线
workspaceKeyStr := ""            // 冗余后缀
```

---

## 代码复用约束

### 必须使用现有工具函数

| 功能 | 必须使用 | 位置 |
|------|----------|------|
| 命令执行 | `common.Execute()` | `internal/common/executor.go` |
| 自动探测 | `common.TryAutoDetect()` | `internal/common/autodetect.go` |
| JSON输出 | `common.PrintJSON()` | `internal/common/executor.go` |
| 错误输出 | `common.PrintError()` | `internal/common/executor.go` |
| YAML解析 | `common.LoadFromYAMLFile()` | `internal/common/yaml.go` |
| 富文本提取 | `common.ExtractTextFromRichText()` | `internal/common/util.go` |

**禁止行为：**
- ❌ 重新实现已有功能的函数
- ❌ 复制粘贴其他命令的工具函数到本地
- ❌ 引入新的依赖做已有功能

### 可接受的复制范围

**可以复制并修改的代码：**
- 从现有命令复制整体结构框架
- 从现有API文件复制方法结构

**禁止复制并保留的内容：**
- 业务特定的ID、名称、URL
- 测试数据
- 特定项目的配置

---

## 安全检查清单

### 提交前必须检查

- [ ] 代码使用 `go vet ./...` 无错误
- [ ] 代码使用 `gofmt -w .` 格式化
- [ ] 所有新命令都有 `--help` 输出
- [ ] 自动探测功能正常工作（如适用）
- [ ] 错误时使用 `common.PrintError()` 输出统一格式
- [ ] API方法有错误处理，不 panic
- [ ] 使用了 `defer resp.Body.Close()`
- [ ] 没有引入不必要的新依赖
- [ ] 命令已注册到 `rootCmd`

### 禁止提交的代码

- ❌ `panic()` - 必须使用错误返回
- ❌ `fmt.Println()` - 必须使用统一输出
- ❌ `os.Exit()` 在 `common.Execute` 之外
- ❌ 硬编码的敏感信息（密码、密钥）
- ❌ 注释掉的调试代码
- ❌ `TODO` 没有后续说明

---

## 扩展约束

### 何时可以新增

✅ **可以新增：**
- 新的子命令（遵循命名规范）
- 新的 API 方法（遵循 API 规范）
- 新的标志（使用已有命名约定）

### 何时不可以新增

❌ **不可以新增：**
- 新的输出格式（必须使用统一JSON）
- 新的配置管理方式（必须使用 `CommandContext`）
- 新的 HTTP 客户端创建方式
- 新的日志库（必须使用 zap）
- 新的命令行框架（必须使用 Cobra）

---

## 问题处理流程

**当不确定如何实施时：**

1. **查看现有代码** - 找到相似功能的命令，复制其模式
2. **查阅文档** - 仔细阅读 [05-新增命令指南.md](./05-新增命令指南.md)
3. **保守处理** - 不确定时，选择最简单、最符合现有模式的方式
4. **不要创造** - 不要创造新的设计模式或架构

**禁止行为：**
- ❌ "我觉得这样可以改进" - 除非有明确需求
- ❌ "这样写更简洁" - 违反统一性原则
- ❌ "未来可能需要" - 只做当前需求

---

## AI开发流程模板

### 添加新命令的标准流程

```
1. 阅读文档（15分钟）
   ├── README.md
   ├── 01-总体架构.md
   ├── 02-命令命名规范.md
   ├── 05-新增命令指南.md
   └── 本清单

2. 查看参考代码（10分钟）
   ├── cmd/lc/req.go（复制整体结构）
   └── internal/api/requirement.go（复制API模式）

3. 实现命令（30分钟）
   ├── 创建 cmd/lc/xxx.go
   ├── 实现子命令（使用 Execute）
   ├── 实现 internal/api/xxx.go
   └── 在 init() 中注册

4. 验证（15分钟）
   ├── go vet ./...
   ├── go build ./...
   ├── ./bin/lc xxx --help
   └── 测试基本功能

5. 提交
   └── 确保通过所有安全检查
```

---

**违反本清单的后果：**
- 代码需要重写
- 破坏项目一致性
- 增加维护成本

**原则：先理解，后复制，再修改。**
