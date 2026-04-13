# OTP 动态配置保护命令

> 版本: v1.1 | 日期: 2026-03-21

---

## 功能概述

OTP 二次验证现在支持**动态配置**需要保护的命令列表。用户可以通过 `lc otp config` 命令管理受保护的命令，而无需修改代码。

---

## 默认保护列表

系统预定义的默认受保护命令（**不可逆操作**）：

```json
["repo delete", "readonly off"]
```

| 命令 | 风险说明 |
|------|---------|
| `repo delete` | **不可逆** - 删除后所有代码和历史记录永久丢失 |
| `readonly off` | 会开放所有写入操作，增加误操作风险 |

> **原则**: 默认只保护**不可逆**的高风险操作。可逆操作（如 `pr merge` 可通过 git revert 回滚）不列入默认保护，用户可自行添加。

---

## 配置方式

### 1. 查看当前保护列表

```bash
lc otp config list
```

输出示例：
```json
{
  "protectedCommands": ["repo delete", "readonly off"],
  "isCustom": false,
  "message": "使用默认保护列表",
  "defaultCommands": ["repo delete", "readonly off"],
  "otpEnabled": true
}
```

### 2. 添加命令到保护列表

```bash
lc otp config add "req delete"
lc otp config add "task delete"
lc otp config add "pr create"
```

添加成功后：
```json
{
  "added": "req delete",
  "protectedCommands": ["repo delete", "readonly off", "req delete"],
  "message": "已将 'req delete' 添加到 OTP 保护列表"
}
```

### 3. 从保护列表移除命令

```bash
lc otp config remove "pr create"
```

> **注意**: 如果当前使用的是默认列表（未自定义），则无法单独移除命令。需要先添加至少一个命令，创建自定义列表后才能移除。

### 4. 重置为默认列表

```bash
lc otp config reset
```

会清除所有自定义设置，恢复为系统默认的 `["repo delete", "readonly off"]`。

---

## 配置文件结构

配置保存在 `~/.lc/config.json` 中：

```json
{
  "cookie": "...",
  "readonly": false,
  "otp": {
    "enabled": true,
    "secret": "HAEXHXIW6QQVFLUPYOVIGQTY7MYPZMKK",
    "verifiedAt": "2026-03-21T10:00:30+08:00",
    "sessionExpiryMinutes": 5,
    "protectedCommands": [
      "repo delete",
      "readonly off",
      "req delete",
      "task delete"
    ]
  }
}
```

- `protectedCommands` 为空数组或未设置时，使用默认列表
- `protectedCommands` 有内容时，使用自定义列表（完全替换默认列表）

---

## 工作原理

### 优先级规则

1. **自定义列表优先**: 如果配置中存在 `protectedCommands` 且非空，则完全使用该列表
2. **默认列表兜底**: 如果 `protectedCommands` 为空或未设置，使用 `DefaultProtectedCommands`

### 代码实现

```go
// internal/common/otp_guard.go

// 默认保护列表
var DefaultProtectedCommands = []string{
    "pr merge",
    "readonly off",
}

// 获取当前保护列表（配置优先）
func GetProtectedCommands(cfg *config.Config) []string {
    if cfg.OTP.Enabled && len(cfg.OTP.ProtectedCommands) > 0 {
        return cfg.OTP.ProtectedCommands
    }
    return DefaultProtectedCommands
}

// 检查命令是否受保护
func IsCommandProtected(cfg *config.Config, operation string) bool {
    protected := GetProtectedCommands(cfg)
    for _, cmd := range protected {
        if cmd == operation {
            return true
        }
    }
    return false
}
```

---

## 使用场景

### 场景一：扩展保护范围

默认只保护 `repo delete` 和 `readonly off`。如果需要保护更多操作：

```bash
# 保护删除类操作
lc otp config add "req delete"
lc otp config add "task delete"
lc otp config add "bug delete"

# 保护合并操作（虽然可回滚，但可增加确认）
lc otp config add "pr merge"
```

现在执行这些命令都会触发 OTP 验证。

### 场景二：缩小保护范围

认为 `readonly off` 不需要 OTP（只保留 `repo delete` 即可）：

```bash
# 创建只包含 repo delete 的自定义列表
lc otp config add "repo delete"
# 现在只有 repo delete 需要 OTP，readonly off 不需要了
```

### 场景三：缩小保护范围

如果认为 `readonly off` 不需要 OTP（只保留默认只读机制即可）：

```bash
# 创建自定义列表，只保留 repo delete
lc otp config add "repo delete"        # 先添加要保留的
lc otp config remove "readonly off"    # 再移除不需要的
```

或者完全自定义：

```bash
lc otp config reset                    # 重置为默认
lc otp config add "pr merge"           # 额外保护 merge
# 现在 ["repo delete", "readonly off", "pr merge"] 都需要 OTP
```

---

## 向后兼容

- 未设置 `protectedCommands` 的用户继续使用默认行为
- 现有 OTP 配置不受此变更影响
- 禁用 OTP 时，保护列表配置会被清除（与现有行为一致）

---

## 限制与注意事项

1. **命令匹配是精确匹配**: 必须完全匹配命令的 `CommandName`（如 `"pr merge"` 不能写成 `"prmerge"`）

2. **子命令粒度**: 支持子命令级别的保护，例如可以单独保护 `pr merge` 而不保护 `pr create`

3. **提示信息**: 对于自定义命令，如果不在 `DangerousOperations` map 中定义的，会显示通用提示信息

4. **配置生效**: 修改保护列表后立即生效，无需重启

---

## 预定义的危险操作信息

以下命令有详细的提示信息（名称、描述、风险等级、原因）：

| 命令 | 描述 | 风险等级 |
|------|------|----------|
| `pr merge` | 合并代码请求 | high |
| `pr create` | 创建合并请求 | medium |
| `pr review` | 审核合并请求 | medium |
| `req delete` | 删除需求 | high |
| `task delete` | 删除任务 | medium |
| `bug delete` | 删除缺陷 | medium |
| `repo delete` | 删除代码仓库 | critical |
| `readonly off` | 关闭只读模式 | medium |

对于自定义添加但不在上表中的命令，会显示通用提示：
- 名称: 命令本身
- 描述: 命令本身
- 风险等级: high
- 原因: "该操作需要二次验证"

---

## 相关命令

```bash
# 查看 OTP 状态（含保护列表）
lc otp status

# 查看保护列表详情
lc otp config list

# 管理保护列表
lc otp config add <command>
lc otp config remove <command>
lc otp config reset
```
