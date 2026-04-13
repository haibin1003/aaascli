# OTP 二次验证功能设计文档

## 变更记录表

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|---------|------|
| v1.0 | 2026-03-20 | 初始版本 | AI Assistant |
| v1.1 | 2026-03-20 | 根据代码实现更新，补充具体实现细节 | AI Assistant |
| v1.2 | 2026-03-21 | 添加动态配置保护命令列表设计 | AI Assistant |

---

## 1. 架构设计

### 1.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        用户交互层                                │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ otp      │ │ otp      │ │ otp      │ │ otp      │           │
│  │ setup    │ │ verify   │ │ disable  │ │ status   │           │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘           │
│       └────────────┴────────────┴────────────┘                  │
│                              │                                  │
│  ┌───────────────────────────┼───────────────────────────┐     │
│  │                    OTP 拦截器                         │     │
│  │  ┌─────────────────────┐ │ ┌─────────────────────┐   │     │
│  │  │ RequireOTPOrPrompt  │ │ │ CheckOTPForDangerous│   │     │
│  │  │                     │ │ │ Operation           │   │     │
│  │  └─────────────────────┘ │ └─────────────────────┘   │     │
│  └───────────────────────────┼───────────────────────────┘     │
└──────────────────────────────┼──────────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────────┐
│                         服务层                                 │
│                    ┌─────────┴────────┐                        │
│                    │  OTPService      │                        │
│                    │  • GenerateSecret│                        │
│                    │  • GenerateQRCode│                        │
│                    │  • GenerateCode  │                        │
│                    │  • VerifyCode    │                        │
│                    └──────────────────┘                        │
└─────────────────────────────────────────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────────┐
│                        配置存储层                               │
│              ┌───────────────┴───────────────┐                  │
│              │      ~/.lc/config.json        │                  │
│              │  {                              │                  │
│              │    "otp": {                     │                  │
│              │      "enabled": true,           │                  │
│              │      "secret": "BASE32...",     │                  │
│              │      "verifiedAt": "...",       │                  │
│              │      "sessionExpiryMinutes": 5  │                  │
│              │    }                            │                  │
│              │  }                              │                  │
│              └─────────────────────────────────┘                  │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 核心组件

| 组件 | 文件路径 | 职责 |
|------|---------|------|
| OTP 配置 | `internal/config/config.go` | 定义 OTPConfig 结构体，管理配置存取 |
| OTP 服务 | `internal/service/otp_service.go` | TOTP 算法实现、二维码 URL 生成 |
| OTP 拦截器 | `internal/common/otp_guard.go` | 危险操作前检查 OTP 状态 |
| OTP 命令 | `cmd/lc/otp.go` | 用户交互命令实现 |

## 2. 详细设计

### 2.1 配置设计

```go
// OTPConfig TOTP 配置
type OTPConfig struct {
    Enabled       bool       `json:"enabled,omitempty"`        // 是否启用
    Secret        string     `json:"secret,omitempty"`         // Base32 密钥（20字节）
    VerifiedAt    *time.Time `json:"verifiedAt,omitempty"`     // 最后验证时间
    SessionExpiry int        `json:"sessionExpiryMinutes,omitempty"` // 有效期(分钟)，默认5
    ProtectedCommands []string `json:"protectedCommands,omitempty"` // 自定义保护命令列表
}
```

**设计决策：**
- 密钥存储在本地，不上传服务器，确保安全性
- 使用 `VerifiedAt` 时间戳而非 token，简化实现
- 会话有效期可配置，默认 5 分钟
- OTP 命令的 `CommandName` 为空，绕过只读检查
- `ProtectedCommands` 为空时使用默认保护列表，非空时完全替换默认列表
- 默认只保护不可逆操作：`repo delete`、`readonly off`

### 2.2 TOTP 算法实现

遵循 RFC 6238 标准：

```
TOTP = Truncate(HMAC-SHA1(Secret, TimeCounter))

其中:
- Secret = 20 字节随机数，Base32 编码（32个字符）
- TimeCounter = floor(CurrentUnixTime / 30)
- Truncate = 动态截断，取 31 位后模 1000000
- 输出 = 6 位数字，不足补零
```

**验证窗口：**
- 当前窗口 (t)
- 前一个窗口 (t-1) - 防止刚生成的码过期
- 后一个窗口 (t+1) - 防止时间偏差

### 2.3 二维码 URL 格式

```
otpauth://totp/{account}?secret={secret}&issuer={issuer}&algorithm=SHA1&digits=6&period=30

示例:
otpauth://totp/weibaohui@hq.cmcc?secret=HAEXHXIW6QQVFLUPYOVIGQTY7MYPZMKK&issuer=灵畿CLI&algorithm=SHA1&digits=6&period=30
```

**设计决策：**
- Label 只使用账号（不包含 issuer），避免显示 URL 编码的乱码
- Issuer 设置为 "灵畿CLI"（中文）
- 不添加额外的一层 issuer:account 前缀

### 2.4 二维码终端显示

使用 `github.com/skip2/go-qrcode` 库：

```go
qr, _ := qrcode.New(qrURL, qrcode.Low)
fmt.Println(qr.ToSmallString(false))
```

**特点：**
- 使用小尺寸 ASCII 字符（▄▀█ 等半块字符）
- 使用低容错级别（Low），减小二维码尺寸
- 直接输出到终端，无需图片查看器

### 2.5 会话管理

```
验证成功:
  1. 记录 verifiedAt = now()
  2. 保存到配置文件

验证检查:
  1. 读取 verifiedAt
  2. 计算 expiresAt = verifiedAt + sessionExpiry
  3. 如果 now() < expiresAt，验证通过

会话过期:
  无需主动清理，检查时自动判断
```

### 2.6 危险操作拦截流程

```
用户执行危险操作
        │
        ▼
┌───────────────┐
│ OTP 已启用?   │────否────▶ 正常执行
└───────┬───────┘
        │ 是
        ▼
┌───────────────┐
│ 会话有效?     │────是────▶ 正常执行
└───────┬───────┘
        │ 否
        ▼
┌───────────────┐
│ 显示风险警告  │
│ 提示输入 OTP  │
└───────┬───────┘
        │
        ▼
┌───────────────┐
│ 验证 OTP 密码 │
└───────┬───────┘
    │       │
   成功    失败
    │       │
    ▼       ▼
 执行操作  拒绝执行
 更新会话  返回错误
```

## 3. 接口设计

### 3.1 OTP 服务接口

```go
type OTPService struct{}

// 生成新的 TOTP 密钥（20字节，Base32编码）
func (s *OTPService) GenerateSecret() (string, error)

// 生成二维码 URL
// issuer 为 "灵畿CLI"，label 只使用 account
func (s *OTPService) GenerateQRCodeURL(secret, account, issuer string) string

// 生成指定时间的 TOTP 码（6位数字）
func (s *OTPService) GenerateCode(secret string, t time.Time) (string, error)

// 验证 OTP 码，支持时间窗口容错（window=1表示前后各1个窗口）
func (s *OTPService) VerifyCode(secret, code string, window int) (bool, error)

// 生成备用码（10个，8位十六进制）
func (s *OTPService) GenerateBackupCodes() ([]string, error)

// 获取当前时间窗口剩余秒数
func (s *OTPService) GetRemainingSeconds() int

// 格式化密钥显示（4字符一组）
func (s *OTPService) FormatSecretForDisplay(secret string) string
```

### 3.2 OTP 守卫接口

```go
// 默认保护列表（不可逆操作）
var DefaultProtectedCommands = []string{
    "repo delete",   // 删除仓库 - 不可逆
    "readonly off",  // 关闭只读 - 会开放所有写入
}

// 危险操作详细信息（用于提示信息）
var DangerousOperations = map[string]DangerousOpInfo{
    "pr merge":     {Name: "pr merge", Description: "合并代码请求", RiskLevel: "high"},
    "readonly off": {Name: "readonly off", Description: "关闭只读模式", RiskLevel: "medium"},
    "repo delete":  {Name: "repo delete", Description: "删除代码仓库", RiskLevel: "critical"},
    // ...
}

// 获取当前保护列表（配置优先，默认兜底）
func GetProtectedCommands(cfg *config.Config) []string

// 检查命令是否在保护列表中
func IsCommandProtected(cfg *config.Config, operation string) bool

// 检查危险操作是否需要 OTP
func CheckOTPForDangerousOperation(cfg *config.Config, operation string) error

// 提示并验证 OTP（交互式）
func PromptAndVerifyOTP(cfg *config.Config) (bool, error)

// 检查或提示 OTP（统一入口）
func RequireOTPOrPrompt(cfg *config.Config, operation string) error

// 检查会话是否有效
func IsOTPSessionValid(cfg *config.Config) bool

// 检查 OTP 是否已启用
func IsOTPEnabled(cfg *config.Config) bool
```

### 3.3 命令接口

```bash
# 初始化 OTP（不受只读限制）
lc otp setup
输出: 二维码（ASCII）、密钥（分组显示）、备用码

# 验证 OTP（不受只读限制）
lc otp verify [code]
输出: 验证结果、会话过期时间

# 关闭 OTP（不受只读限制，但需要 OTP 验证）
lc otp disable
输出: 关闭结果

# 查看状态（不受只读限制）
lc otp status
输出: 是否启用、会话状态、剩余时间、保护命令列表

# 管理保护命令列表（不受只读限制）
lc otp config list          # 列出当前保护命令
lc otp config add [cmd]     # 添加命令到保护列表
lc otp config remove [cmd]  # 从保护列表移除命令
lc otp config reset         # 重置为默认保护列表
```

## 4. 安全设计

### 4.1 密钥生成

```go
// 使用 crypto/rand 生成 20 字节随机数
secret := make([]byte, 20)
rand.Read(secret)

// Base32 编码（32个字符）
encoded := base32.StdEncoding.EncodeToString(secret)
```

### 4.2 防时序攻击

```go
// 常量时间比较
func constantTimeEquals(a, b string) bool {
    if len(a) != len(b) {
        return false
    }
    var result byte
    for i := 0; i < len(a); i++ {
        result |= a[i] ^ b[i]
    }
    return result == 0
}
```

### 4.3 关闭保护

关闭 OTP 需要：
1. 有效的 OTP 会话，或
2. 输入正确的 OTP 密码

防止恶意脚本关闭 OTP 保护。

### 4.4 只读模式绕过

OTP 命令通过设置 `CommandName = ""` 绕过只读检查：

```go
common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
    // ...
}, common.ExecuteOptions{
    CommandName: "",  // 空表示不检查只读模式
})
```

## 5. 集成点

### 5.1 与只读模式集成

```
执行危险操作前检查:
1. 检查只读模式 ──被阻止──▶ 返回只读错误
2. 检查 OTP 验证 ──被阻止──▶ 返回 OTP 错误
3. 执行操作
```

### 5.2 默认保护的操作

以下操作默认受 OTP 保护（在 `DefaultProtectedCommands` 中定义）：

- `cmd/lc/readonly.go`: `readonly off` - 关闭只读模式
- `cmd/lc/repo.go`: `repo delete` - 删除代码仓库（预留命令）

### 5.3 可动态添加的保护操作

以下操作已定义详细信息（在 `DangerousOperations` 中），用户可通过 `lc otp config add` 动态添加到保护列表：

- `pr merge` - 合并代码请求
- `pr create` - 创建合并请求
- `pr review` - 审核合并请求
- `req delete` - 删除需求
- `task delete` - 删除任务
- `bug delete` - 删除缺陷
- `repo delete` - 删除代码仓库

集成方式：在命令函数开头添加
```go
if err := common.RequireOTPOrPrompt(ctx.Config, "command name"); err != nil {
    return nil, err
}
```

## 6. 依赖库

| 库 | 版本 | 用途 |
|----|------|------|
| github.com/skip2/go-qrcode | latest | 终端二维码显示 |

## 7. 错误处理

| 错误码 | 场景 | 处理方式 |
|--------|------|---------|
| OTP_REQUIRED | 未验证执行危险操作 | 提示先运行 `lc otp verify` |
| OTP_INVALID | 验证码错误 | 提示重新输入 |
| OTP_NOT_ENABLED | OTP 未启用但调用相关命令 | 提示先运行 `lc otp setup` |
| OTP_EXPIRED | 会话过期 | 提示重新验证 |
| READONLY_MODE | 只读模式下执行写入操作 | OTP 命令不触发此错误 |

## 8. 文件变更清单

| 文件路径 | 变更类型 | 说明 |
|---------|---------|------|
| `internal/config/config.go` | 修改 | 添加 OTPConfig 结构体 |
| `internal/service/otp_service.go` | 新增 | TOTP 算法实现 |
| `internal/common/otp_guard.go` | 新增 | OTP 拦截器 |
| `cmd/lc/otp.go` | 新增 | OTP 命令实现 |
| `cmd/lc/readonly.go` | 修改 | 集成 OTP 检查 |
| `cmd/lc/pr.go` | 修改 | 集成 OTP 检查 |
| `internal/common/executor.go` | 修改 | 添加 OTP 错误处理 |
| `go.mod` | 修改 | 添加 go-qrcode 依赖 |
