# LC CLI 只读模式设计方案

## 背景

为了防止 AI 助手或用户误操作导致重要数据被删除或修改，LC CLI 需要引入只读模式（Read-Only Mode）。安装后默认为只读模式，需要显式关闭只读模式后才能执行写入操作。

---

## 命令设计

### 基本命令

```bash
# 查看当前只读模式状态
lc readonly

# 开启只读模式（禁止所有修改操作）
lc readonly on

# 关闭只读模式（允许读写操作）
lc readonly off
```

### 输出格式

#### 查看状态（默认开启）
```json
{
  "success": true,
  "data": {
    "readonly": true,
    "description": "当前处于只读模式，禁止执行创建、更新、删除等操作",
    "writable_commands": ["list", "view", "search", "detect", "status"],
    "readonly_commands": ["create", "update", "delete", "merge"]
  }
}
```

#### 开启/关闭只读模式
```json
{
  "success": true,
  "data": {
    "readonly": false,
    "message": "已关闭只读模式，现在可以执行读写操作"
  }
}
```

---

## 配置文件存储

### 存储位置

只读模式状态存储在 `~/.lc/config.json` 中：

```json
{
  "cookie": "MOSS_SESSION=xxx",
  "readonly": true
}
```

### 默认值

- **新安装/首次使用**：`readonly: true`（默认只读）
- **升级用户**：保持原有设置，如果不存在则默认为 `true`

---

## 权限控制机制

### 命令分类

#### 只读命令（Read-Only Commands）
以下命令不受只读模式限制，始终可执行：

| 模块 | 命令 | 说明 |
|------|------|------|
| 通用 | `lc version` | 查看版本 |
| 通用 | `lc login` | 登录（配置操作，非数据操作） |
| 通用 | `lc detect` | 自动探测 |
| 通用 | `lc helper` | 浏览器插件管理 |
| 通用 | `lc readonly` | 只读模式开关（自身不受限制） |
| 需求 | `lc req list` | 查询需求列表 |
| 需求 | `lc req view` | 查看需求详情 |
| 需求 | `lc req search` | 搜索需求 |
| 任务 | `lc task list` | 查询任务列表 |
| 任务 | `lc task search` | 搜索任务 |
| 缺陷 | `lc bug list` | 查询缺陷列表 |
| 缺陷 | `lc bug view` | 查看缺陷详情 |
| 缺陷 | `lc bug status` | 查看缺陷状态 |
| 仓库 | `lc repo list` | 查询仓库列表 |
| 仓库 | `lc repo search` | 搜索仓库 |
| PR | `lc pr list` | 查询 PR 列表 |
| PR | `lc pr view` | 查看 PR 详情 |
| 空间 | `lc space list` | 查询研发空间 |
| 空间 | `lc space project` | 查询项目列表 |
| 文档库 | `lc lib list` | 查询文档库 |
| 文档库 | `lc lib folder list` | 查询文件夹内容 |

#### 写入命令（Write Commands）
以下命令在只读模式下会被禁止：

| 模块 | 命令 | 说明 |
|------|------|------|
| 需求 | `lc req create` | 创建需求 |
| 需求 | `lc req update` | 更新需求 |
| 需求 | `lc req delete` | 删除需求 |
| 任务 | `lc task create` | 创建任务 |
| 任务 | `lc task delete` | 删除任务 |
| 缺陷 | `lc bug create` | 创建缺陷 |
| 缺陷 | `lc bug update-status` | 更新缺陷状态 |
| 缺陷 | `lc bug delete` | 删除缺陷 |
| 仓库 | `lc repo create` | 创建仓库 |
| 仓库 | `lc repo delete` | 删除仓库 |
| 仓库 | `lc repo disable-work-item-link` | 禁用工作项关联 |
| PR | `lc pr create` | 创建 PR |
| PR | `lc pr comment` | 添加 PR 评论 |
| PR | `lc pr patch-comment` | 解决 PR 评论 |
| PR | `lc pr merge` | 合并 PR |
| 文档库 | `lc lib create` | 创建文档库 |
| 文档库 | `lc lib delete` | 删除文档库 |
| 文档库 | `lc lib folder create` | 创建文件夹 |
| 文档库 | `lc lib upload` | 上传文件 |
| 文档库 | `lc lib file delete` | 删除文件/文件夹 |

### 拦截机制

在只读模式下，当用户执行写入命令时：

1. **命令预处理**：在命令执行前检查 `readonly` 配置
2. **错误返回**：返回统一的 JSON 错误，提示用户关闭只读模式
3. **不执行实际操作**：不会调用任何 API

### 错误提示格式

```json
{
  "success": false,
  "error": {
    "code": "READONLY_MODE",
    "message": "当前处于只读模式，禁止执行写入操作",
    "details": "命令 'lc req create' 是写入操作，在只读模式下被禁止",
    "suggestion": "如需执行写入操作，请先关闭只读模式：\n  lc readonly off\n\n注意：关闭只读模式后，所有写入操作将直接生效，请谨慎操作。"
  },
  "meta": {
    "requestId": "",
    "timestamp": "2026-03-17T05:00:00Z",
    "version": "v0.2.6"
  }
}
```

---

## 实现方案

### 1. 配置文件扩展

在 `internal/config/config.go` 中添加 `Readonly` 字段：

```go
type Config struct {
    Cookie   string            `json:"cookie"`
    Readonly bool              `json:"readonly"`  // 新增：只读模式，默认 true
    API      APIConfig         `json:"api"`
    Auth     AuthConfig        `json:"auth"`
    User     UserConfig        `json:"user"`
    Defaults DefaultsConfig    `json:"defaults"`
}
```

在 `NewConfig()` 中设置默认值：

```go
func NewConfig() *Config {
    return &Config{
        Readonly: true,  // 默认开启只读模式
        // ...
    }
}
```

### 2. 只读检查函数

在 `internal/common` 中创建 `readonly.go`：

```go
package common

import (
    "github.com/user/lc/internal/config"
)

// IsReadonly 检查当前是否处于只读模式
func IsReadonly() bool {
    cfg, err := config.LoadConfigWithDefaults(config.GetDefaultConfigPath())
    if err != nil {
        return true  // 默认安全，出错时返回只读
    }
    return cfg.Readonly
}

// CheckReadonlyForWrite 检查写入操作是否被允许
// 如果是只读模式，返回错误；否则返回 nil
func CheckReadonlyForWrite(cmdName string) error {
    if IsReadonly() {
        return NewReadonlyError(cmdName)
    }
    return nil
}

// ReadonlyError 只读模式错误
type ReadonlyError struct {
    Command string
}

func NewReadonlyError(cmd string) *ReadonlyError {
    return &ReadonlyError{Command: cmd}
}

func (e *ReadonlyError) Error() string {
    return fmt.Sprintf("当前处于只读模式，命令 '%s' 被禁止", e.Command)
}
```

### 3. 命令集成

在每个写入命令的 `Run` 函数开头添加检查：

```go
// 示例：req create 命令
Run: func(cmd *cobra.Command, args []string) {
    // 检查只读模式
    if err := common.CheckReadonlyForWrite("req create"); err != nil {
        common.PrintError(err)
        os.Exit(1)
    }

    // ... 原有逻辑
}
```

### 4. readonly 命令实现

创建 `cmd/lc/readonly.go`：

```go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/user/lc/internal/common"
    "github.com/user/lc/internal/config"
)

var readonlyCmd = &cobra.Command{
    Use:   "readonly [on|off]",
    Short: "只读模式管理",
    Long: `管理系统只读模式，防止误操作删除或修改重要数据。

默认情况下，安装后处于只读模式，只能执行查询操作。
需要执行创建、更新、删除等写入操作时，请先关闭只读模式。

示例:
  # 查看当前只读模式状态
  lc readonly

  # 开启只读模式（禁止写入）
  lc readonly on

  # 关闭只读模式（允许读写）
  lc readonly off`,
    Run: func(cmd *cobra.Command, args []string) {
        common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
            cfg, err := config.LoadConfigWithDefaults(config.GetDefaultConfigPath())
            if err != nil {
                return nil, fmt.Errorf("加载配置失败: %w", err)
            }

            // 没有参数，显示当前状态
            if len(args) == 0 {
                return map[string]interface{}{
                    "readonly":    cfg.Readonly,
                    "description": getReadonlyDescription(cfg.Readonly),
                }, nil
            }

            // 处理 on/off
            switch args[0] {
            case "on":
                cfg.Readonly = true
            case "off":
                cfg.Readonly = false
            default:
                return nil, fmt.Errorf("无效参数: %s，请使用 'on' 或 'off'", args[0])
            }

            // 保存配置
            if err := config.SaveConfig(cfg); err != nil {
                return nil, fmt.Errorf("保存配置失败: %w", err)
            }

            return map[string]interface{}{
                "readonly": cfg.Readonly,
                "message":  getReadonlyMessage(cfg.Readonly),
            }, nil
        }, common.ExecuteOptions{
            DebugMode: debugMode,
            Insecure:  insecureSkipVerify,
            Logger:    &logger,
        })
    },
}

func getReadonlyDescription(readonly bool) string {
    if readonly {
        return "当前处于只读模式，禁止执行创建、更新、删除等操作"
    }
    return "当前处于读写模式，可以执行所有操作"
}

func getReadonlyMessage(readonly bool) string {
    if readonly {
        return "已开启只读模式，写入操作将被禁止"
    }
    return "已关闭只读模式，现在可以执行读写操作"
}

func init() {
    rootCmd.AddCommand(readonlyCmd)
}
```

### 5. 配置保存函数

在 `internal/config/config.go` 中添加保存配置的函数：

```go
// SaveConfig 保存配置到文件
func SaveConfig(cfg *Config) error {
    configPath := GetDefaultConfigPath()

    // 读取现有配置，保留其他字段
    var existing map[string]interface{}
    if data, err := os.ReadFile(configPath); err == nil {
        json.Unmarshal(data, &existing)
    }

    // 更新字段
    existing["cookie"] = cfg.Cookie
    existing["readonly"] = cfg.Readonly

    // 保存
    data, err := json.MarshalIndent(existing, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(configPath, data, 0644)
}
```

---

## 安全考虑

1. **默认安全**：新安装默认只读，防止误操作
2. **显式关闭**：关闭只读模式需要执行明确命令 `lc readonly off`
3. **错误提示友好**：被阻止时给出清晰的错误提示和操作指引
4. **不影响配置操作**：`lc login` 和 `lc readonly` 本身不受只读模式限制

---

## 测试建议

1. **默认状态测试**：新安装后检查默认是否为只读
2. **状态切换测试**：`on`/`off` 切换是否正常
3. **写入阻止测试**：只读模式下执行写入命令是否被阻止
4. **读取允许测试**：只读模式下执行读取命令是否正常
5. **配置持久化测试**：重启后配置是否保持

---

## 文档更新

需要更新以下文档：

1. `docs/LC_CLI_USAGE_GUIDE.md` - 添加只读模式说明章节
2. `README.md` - 在快速开始中添加只读模式说明
3. `CLAUDE.md` - 添加 AI 使用时的注意事项
