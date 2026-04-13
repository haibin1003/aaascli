# OTP 二次验证与只读模式 使用说明

> 版本: v1.0 | 日期: 2026-03-21

---

## 目录

1. [功能概述](#1-功能概述)
2. [只读模式（Readonly）](#2-只读模式readonly)
3. [OTP 二次验证](#3-otp-二次验证)
4. [两者的交互关系](#4-两者的交互关系)
5. [受影响的命令范围（完整清单）](#5-受影响的命令范围完整清单)
6. [执行流程详解](#6-执行流程详解)
7. [典型使用场景](#7-典型使用场景)
8. [配置说明](#8-配置说明)
9. [常见问题](#9-常见问题)

---

## 1. 功能概述

`lc` 提供两层操作保护机制，相互独立、依次生效：

```
用户执行命令
     │
     ▼
[第一层] 只读模式检查 (Readonly)
     │  ── 拦截所有写入命令（命令粒度）
     │
     ▼
[第二层] OTP 二次验证 (OTP Guard)
     │  ── 对高风险命令额外要求身份验证
     │
     ▼
执行命令
```

| 保护层 | 粒度 | 目的 | 配置位置 |
|--------|------|------|----------|
| 只读模式 | 命令级 | 防止所有写操作误执行 | `~/.lc/config.json` → `readonly` |
| OTP 验证 | 操作级 | 对高危操作要求二次身份确认 | `~/.lc/config.json` → `otp` |

---

## 2. 只读模式（Readonly）

### 2.1 什么是只读模式

只读模式是一个全局开关，**开启时禁止所有写操作**。默认安装后处于只读模式。

```bash
# 查看当前状态
lc readonly

# 开启只读模式
lc readonly on

# 永久关闭只读模式
lc readonly off

# 临时关闭（30分钟后自动恢复）
lc readonly off --duration 30m
```

### 2.2 只读模式的工作原理

只读检查在 `internal/common/executor.go` 的 `Execute()` 函数中最先执行，通过 `CommandName` 字段判断：

```go
// 只要 CommandName 不为空，就会检查只读模式
if opts.CommandName != "" {
    if err := CheckReadonlyForWrite(opts.CommandName); err != nil {
        PrintError(err)
        os.Exit(1)
    }
}
```

**凡是 `CommandName` 设置为非空字符串的命令，在只读模式下均会被拦截。**

### 2.3 只读模式拦截的命令

以下命令在只读模式下会被拦截（返回 `READONLY_MODE` 错误）：

| 命令 | CommandName |
|------|-------------|
| `lc pr create` | `pr create` |
| `lc pr review` | `pr review` |
| `lc pr merge` | `pr merge` |
| `lc pr comment` | `pr comment` |
| `lc pr list` | `pr patch-comment` |
| 其他写命令 | 各自对应名称 |

> 注意：`lc pr view` 和 `lc pr list` 当前实现中 `CommandName` 为空，因此**不受**只读模式限制。

### 2.4 只读模式不影响的命令

以下命令无论只读状态如何，**始终可以执行**：

- 所有 `lc otp` 子命令（`setup` / `verify` / `disable` / `status`）
- `lc readonly`（查看状态）
- `lc version`、`lc help` 等

---

## 3. OTP 二次验证

### 3.1 什么是 OTP

OTP（One-Time Password，一次性密码）是基于 RFC 6238 的 TOTP 算法，每 30 秒生成一个 6 位数字验证码，需配合手机验证器应用（Google Authenticator、Microsoft Authenticator 等）使用。

OTP 功能是**可选的**，未启用时对所有命令无影响。

### 3.2 OTP 命令

```bash
# 初始化 OTP（扫描二维码绑定手机验证器）
lc otp setup

# 验证 OTP（创建 5 分钟会话）
lc otp verify 123456
lc otp verify          # 不带参数，交互式输入

# 查看 OTP 状态
lc otp status

# 关闭 OTP（需要有效会话）
lc otp disable
```

### 3.3 OTP 命令的特殊权限

**所有 `lc otp` 子命令均不受只读模式限制**，可以在只读模式下直接执行。

这是通过将 `CommandName` 设置为空字符串 `""` 实现的：

```go
common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
    // OTP 命令逻辑...
}, common.ExecuteOptions{
    CommandName: "",  // 空字符串 = 跳过只读检查
})
```

### 3.4 会话机制

OTP 验证通过后会创建一个本地会话：

- **默认有效期**：5 分钟（可通过配置修改）
- **存储位置**：`~/.lc/config.json` 中的 `otp.verifiedAt` 字段
- **有效期内**：执行危险操作无需重复输入 OTP
- **自动失效**：超时后自动需要重新验证，无需手动退出

```bash
# 查看会话状态
lc otp status
# 输出示例：
# {
#   "enabled": true,
#   "session": {
#     "valid": true,
#     "remainingMin": 4,
#     "remainingSec": 23,
#     "expiresAt": "2026-03-21T10:05:30Z"
#   }
# }
```

---

## 4. 两者的交互关系

### 4.1 优先级规则

**只读模式检查优先于 OTP 检查。** 只有通过只读检查后，才会进行 OTP 检查。

```
命令到达
  │
  ▼
只读模式开启？
  ├── 是 → 直接返回 READONLY_MODE 错误（不进行 OTP 检查）
  └── 否 ↓
        OTP 已启用且为危险操作？
          ├── 否 → 直接执行
          └── 是 → 会话有效？
                    ├── 是 → 直接执行
                    └── 否 → 提示输入 OTP → 验证 → 执行/拒绝
```

### 4.2 `readonly off` 与 OTP 的特殊关系

`readonly off` 本身是一个危险操作，被纳入 OTP 保护列表（风险等级：medium）。

执行 `lc readonly off` 时的完整流程：

```
lc readonly off
  │
  ├── [注意] 此时尚未进入 executor，不走只读检查逻辑
  │
  ▼
OTP 检查（在 readonly.go 中直接调用）
  ├── OTP 未启用 → 直接关闭只读
  └── OTP 已启用 →
        会话有效？
          ├── 是 → 直接关闭只读
          └── 否 → 显示警告，提示输入 OTP
                    验证成功 → 关闭只读
                    验证失败 → 拒绝操作
```

> **关键点**：`readonly off` 的 OTP 检查在执行器（executor）外部完成，因此不存在"只读模式拦截 readonly off"的问题。这是设计上的特殊处理，保证了 OTP 对 `readonly off` 的有效保护。

### 4.3 场景矩阵

| 只读模式 | OTP状态 | 执行危险命令 | 结果 |
|----------|---------|-------------|------|
| OFF | 未启用 | `lc pr merge 123` | 直接执行 |
| OFF | 已启用，会话有效 | `lc pr merge 123` | 直接执行 |
| OFF | 已启用，会话无效 | `lc pr merge 123` | 提示输入 OTP |
| ON | 未启用 | `lc pr merge 123` | READONLY_MODE 错误 |
| ON | 已启用，会话有效 | `lc pr merge 123` | READONLY_MODE 错误 |
| ON | 已启用，会话有效 | `lc readonly off` | OTP通过，关闭只读 |
| ON | 已启用，会话无效 | `lc readonly off` | 提示输入 OTP |

---

## 5. 受影响的命令范围（完整清单）

### 5.1 已集成 OTP 保护的命令

经代码审查，以下命令已集成 OTP 检查（`RequireOTPOrPrompt` 调用）：

#### `pr` 命令组（全部子命令）

> ⚠️ 注意：所有 `pr` 子命令均使用 `"pr merge"` 作为操作键进行 OTP 检查，包括查询类命令。

| 子命令 | 危险程度 | OTP检查键 | 说明 |
|--------|---------|-----------|------|
| `lc pr merge` | 高 | `pr merge` | 代码合并，不可逆 |
| `lc pr create` | 中 | `pr merge` | 创建合并请求 |
| `lc pr review` | 中 | `pr merge` | 审核合并请求 |
| `lc pr comment` | 低 | `pr merge` | 评论合并请求 |
| `lc pr view` | 低 | `pr merge` | 查看合并请求（只读）|
| `lc pr list` | 低 | `pr merge` | 列出合并请求（只读）|
| `lc pr patch-comment` | 低 | `pr merge` | 修改评论状态 |

#### `readonly` 命令

| 子命令 | 危险程度 | OTP检查键 | 说明 |
|--------|---------|-----------|------|
| `lc readonly off` | 中 | `readonly off` | 关闭只读保护 |

### 5.2 已定义但未集成 OTP 的操作（预留扩展）

以下危险操作已在 `DangerousOperations` 中定义，但尚未在对应命令中调用：

| 操作键 | 描述 | 风险等级 | 对应命令（待集成）|
|--------|------|---------|-----------------|
| `req delete` | 删除需求 | high | `lc req delete` |
| `task delete` | 删除任务 | medium | `lc task delete` |
| `bug delete` | 删除缺陷 | medium | `lc bug delete` |
| `repo delete` | 删除代码仓库 | critical | `lc repo delete` |

### 5.3 不受 OTP 限制的命令

以下命令无论 OTP 状态如何，**均无需验证**：

- 所有查询类命令（除 `pr` 命令组外）
- `lc otp setup / verify / disable / status`（OTP 管理命令本身）
- `lc readonly on / off`（OTP 检查内置，无需额外验证）

---

## 6. 执行流程详解

### 6.1 普通危险操作（以 `pr merge` 为例）

```
lc pr merge 123 --squash
       │
       ▼
executor.Execute()
       │
       ├── 检查 CommandName = "pr merge"
       │
       ▼
只读模式检查 (CheckReadonlyForWrite)
       │
       ├── 只读开启 → 输出 READONLY_MODE 错误，exit 1
       │
       └── 只读关闭 ↓
              进入命令函数 mergeMergeRequest()
                     │
                     ▼
              OTP 检查 (RequireOTPOrPrompt "pr merge")
                     │
                     ├── OTP 未启用 → 跳过
                     ├── 会话有效 → 跳过
                     └── 需要验证 ↓
                            显示警告信息
                            提示输入 OTP
                            验证成功 → 更新会话时间戳 → 继续执行
                            验证失败 → 返回 OTP_REQUIRED 错误
```

### 6.2 `readonly off` 的特殊流程

```
lc readonly off
       │
       ▼
readonlyCmd.Run()（在 executor 外部）
       │
       ▼
OTP 检查 (RequireOTPOrPrompt "readonly off")
       │
       ├── OTP 未启用 → 跳过
       ├── 会话有效 → 跳过
       └── 需要验证 ↓
              显示警告：即将关闭只读模式
              提示输入 OTP
              验证成功 → 继续
              验证失败 → 输出错误，return
       │
       ▼
修改 cfg.Readonly = false
       │
       ▼
SaveConfig()
       │
       ▼
输出结果（通过 executor 的只读查询路径）
```

### 6.3 OTP 命令的流程（以 `otp verify` 为例）

```
lc otp verify 123456
       │
       ▼
executor.Execute()
       │
       ├── CommandName = ""（空）
       └── 跳过只读检查 ↓
              进入命令函数 runOTPVerify()
                     │
                     ▼
              验证 6 位验证码
                     │
                     ├── 格式错误 → 报错
                     └── 格式正确 ↓
                            TOTP 算法验证（含±1窗口容错）
                            成功 → 更新 VerifiedAt → 保存配置 → 输出结果
                            失败 → 报错
```

---

## 7. 典型使用场景

### 场景一：日常开发（推荐工作流）

```bash
# 早上开始工作，关闭只读（OTP 保护下需要验证一次）
lc otp verify 123456          # 创建 5 分钟会话
lc readonly off --duration 8h # 工作时间内保持只读关闭

# 此后 5 分钟内，危险操作无需重复输入 OTP
lc pr merge 123 --squash      # 直接执行
lc pr merge 456 --squash      # 直接执行

# 5 分钟后会话过期，下次危险操作会弹出 OTP 提示
lc pr merge 789 --squash
# ⚠️  警告: 即将执行危险操作 [合并代码请求]
# 🔐 请输入 OTP 验证码: ______
```

### 场景二：AI 助手使用 lc（只读保护 AI 误操作）

```
默认状态：只读模式 ON
              │
AI 执行查询命令 ✅（无限制）
              │
AI 尝试执行 lc pr merge → READONLY_MODE 错误 ✅（被拦截）
              │
人类操作：lc readonly off
              │
  OTP 已启用？
  └── 是：需要人类输入 OTP 验证码（AI 无法完成）✅
              │
  只读关闭后，AI 可执行写命令
              │
  AI 执行 lc pr merge 123 → 触发 OTP 检查
  └── 会话有效（刚由人类验证）→ 直接执行 ✅
```

### 场景三：自动化脚本（配合 lc-otp-gen）

```bash
#!/bin/bash
# 自动化 CI 脚本

# 获取当前 OTP 验证码
CODE=$(lc-otp-gen code 2>/dev/null | grep "验证码:" | sed 's/.*验证码: //')

# 验证 OTP（创建会话）
lc otp verify "$CODE"

# 执行操作（5 分钟内有效）
lc pr merge $MR_ID --squash --delete-branch
```

### 场景四：快速一次性操作

```bash
# 不想每次都保持 readonly off，只需一次操作时
lc otp verify 123456          # 验证（创建 5 分钟会话）
lc pr merge 123 --squash      # 在会话内执行
# 操作完成，不需要手动关闭什么
# 5 分钟后会话自动过期，只读模式如果之前是 on，仍然是 on
```

### 场景五：`otp disable` 的正确流程

```bash
# 错误做法（会被拦截）
lc otp disable
# 错误: 需要先通过 OTP 验证

# 正确做法
lc otp verify 123456   # 先验证，建立会话
lc otp disable         # 然后关闭
# ⚠️  确定要关闭 OTP 二次验证吗? [y/N]: y
```

---

## 8. 配置说明

### 8.1 配置文件位置

```
~/.lc/config.json
```

### 8.2 OTP 相关配置项

```json
{
  "readonly": false,
  "otp": {
    "enabled": true,
    "secret": "HAEXHXIW6QQVFLUPYOVIGQTY7MYPZMKK",
    "verifiedAt": "2026-03-21T10:00:30+08:00",
    "sessionExpiryMinutes": 5
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `otp.enabled` | bool | 是否启用 OTP |
| `otp.secret` | string | TOTP 密钥（Base32 编码，20字节）|
| `otp.verifiedAt` | time | 最后验证时间，用于会话有效性判断 |
| `otp.sessionExpiryMinutes` | int | 会话有效期（分钟），默认 5 |

### 8.3 修改会话有效期

目前只能直接编辑配置文件修改 `sessionExpiryMinutes`：

```bash
# 将会话有效期改为 30 分钟
# 手动编辑 ~/.lc/config.json，修改 sessionExpiryMinutes 的值
```

---

## 9. 常见问题

### Q1: 我启用了 OTP，但执行普通查询命令（比如 lc req list）也被拦截了？

**不会**。OTP 只对定义在 `DangerousOperations` 中的操作生效，当前包括：
- 所有 `lc pr` 子命令（含列表和查看）
- `lc readonly off`

普通查询命令（req list、task list 等）不受 OTP 影响。

### Q2: 所有 pr 命令都需要 OTP，包括 `pr list`？

**是的，这是当前实现**。代码中 `pr list`、`pr view` 也调用了 OTP 检查，使用的操作键是 `"pr merge"`。这属于过度保护，未来版本可能会调整为只对 `pr merge` 启用 OTP。

### Q3: 只读模式 ON 时，我想关闭它但没有 OTP 会话怎么办？

```bash
# 1. 先验证 OTP（otp verify 不受只读限制）
lc otp verify 123456

# 2. 再关闭只读（此时有有效会话）
lc readonly off
```

### Q4: 我想先关闭只读再验证 OTP 可以吗？

不行。`readonly off` 本身受 OTP 保护，必须先有有效的 OTP 会话才能关闭只读模式。正确顺序是：**先验证 OTP → 再关闭只读**。

### Q5: OTP 验证码还剩几秒就过期了，验证会失败吗？

不会。系统支持 ±1 个时间窗口（±30秒）的容错，当前验证码和前后各一个窗口的码都能通过验证。当剩余时间 < 10 秒时，系统会提示建议等待新验证码。

### Q6: 忘记手机了，没有 OTP 怎么执行危险操作？

当前没有备用码验证机制（已生成备用码，但验证逻辑待实现）。临时方案：
1. 直接编辑 `~/.lc/config.json`，将 `otp.enabled` 改为 `false`
2. 执行操作后记得重新开启 OTP

### Q7: lc-otp-gen 和 lc 的 OTP 是什么关系？

`lc-otp-gen` 是一个独立的命令行工具，相当于手机上的 Google Authenticator，用于在自动化场景中生成 OTP 验证码。它与 `lc` 共享同一个 TOTP 密钥，两者生成的验证码完全一致。

配置文件分离：
- `lc` 的 OTP 配置：`~/.lc/config.json`（存储密钥）
- `lc-otp-gen` 的配置：`~/.lc-otp-gen/config.json`（独立存储账户和密钥）

---

## 附录：错误码说明

| 错误码 | 触发条件 | 解决方案 |
|--------|---------|---------|
| `READONLY_MODE` | 只读模式下执行写命令 | `lc readonly off` |
| `OTP_REQUIRED` | 危险操作未通过 OTP 验证 | `lc otp verify <code>` |
| `AUTO_DETECT_FAILED` | OTP 未启用时调用相关命令 | `lc otp setup` |

---

*文档基于源码（commit `b841bc1`）审查生成，如代码逻辑有更新请同步更新本文档。*
