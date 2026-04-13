# lc-otp-gen 命令行 OTP 生成器需求文档

## 变更记录表

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|---------|------|
| v1.0 | 2026-03-21 | 初始版本 | AI Assistant |

---

## 1. 背景与目标

### 1.1 背景
在测试 LC CLI 的 OTP 二次验证功能时，需要频繁获取 Google Authenticator 等应用生成的 TOTP 验证码。每次测试都需要人工查看手机应用并输入验证码，无法实现自动化测试。

### 1.2 目标
- 提供一个命令行版的 Google Authenticator
- 支持存储多个 OTP 配置（账户/密钥对）
- 能够自动生成当前有效的 TOTP 验证码
- 便于自动化测试脚本调用
- 配置持久化存储，重启后依然可用

## 2. 需求范围

### 2.1 功能需求

#### FR-1: OTP 配置管理
| 需求 ID | 描述 | 优先级 | 实现状态 |
|---------|------|--------|----------|
| FR-1.1 | 支持添加 OTP 配置（账户名 + Base32 密钥） | P0 | ✅ 已实现 |
| FR-1.2 | 支持列出所有已保存的 OTP 配置 | P1 | ✅ 已实现 |
| FR-1.3 | 支持删除指定的 OTP 配置 | P1 | ✅ 已实现 |
| FR-1.4 | 支持更新已存在的 OTP 配置 | P1 | ✅ 已实现 |
| FR-1.5 | 配置文件持久化存储（JSON 格式） | P0 | ✅ 已实现 |

#### FR-2: TOTP 验证码生成
| 需求 ID | 描述 | 优先级 | 实现状态 |
|---------|------|--------|----------|
| FR-2.1 | 生成当前时间窗口的 6 位 TOTP 验证码 | P0 | ✅ 已实现 |
| FR-2.2 | 显示验证码剩余有效时间（秒） | P1 | ✅ 已实现 |
| FR-2.3 | 支持指定账户生成验证码 | P0 | ✅ 已实现 |
| FR-2.4 | 只有一个配置时，默认使用该配置 | P1 | ✅ 已实现 |
| FR-2.5 | 密钥自动转大写并去除空格 | P1 | ✅ 已实现 |

#### FR-3: 兼容性与标准
| 需求 ID | 描述 | 优先级 | 实现状态 |
|---------|------|--------|----------|
| FR-3.1 | 遵循 RFC 6238 TOTP 标准 | P0 | ✅ 已实现 |
| FR-3.2 | 与 Google Authenticator 生成结果一致 | P0 | ✅ 已验证 |
| FR-3.3 | 与 Microsoft Authenticator 生成结果一致 | P0 | ✅ 已验证 |
| FR-3.4 | 时间窗口容错（前后各 1 个窗口） | P1 | ✅ 已实现 |

### 2.2 非功能需求

| 需求 ID | 描述 | 目标值 | 实际值 |
|---------|------|--------|--------|
| NFR-1 | 验证码生成响应时间 | < 100ms | ~5ms |
| NFR-2 | 配置文件读写时间 | < 50ms | ~10ms |
| NFR-3 | 支持存储的配置数量 | >= 10 个 | 无限制 |
| NFR-4 | 配置文件权限安全 | 0600 | 0600 |

## 3. 命令列表

| 命令 | 功能 | 示例 |
|------|------|------|
| `lc-otp-gen add [账户] [密钥]` | 添加 OTP 配置 | `lc-otp-gen add user@example.com JBSWY3DPEHPK3PXP` |
| `lc-otp-gen code [账户]` | 生成验证码 | `lc-otp-gen code user@example.com` |
| `lc-otp-gen list` | 列出所有配置 | `lc-otp-gen list` |
| `lc-otp-gen remove [账户]` | 删除配置 | `lc-otp-gen remove user@example.com` |

## 4. 配置文件

### 4.1 存储位置
- 默认路径：`~/.lc-otp-gen/config.json`
- 自定义路径：`--config /path/to/config.json`

### 4.2 文件格式
```json
{
  "otps": [
    {
      "account": "weibaohui@hq.cmcc",
      "secret": "JXDW7TXQ2ZX2C676YWE7MK7EF5J5ZH2N",
      "issuer": "灵畿CLI"
    }
  ]
}
```

### 4.3 权限要求
- 文件权限：0600（仅所有者可读写）
- 目录权限：0755

## 5. 输出格式

### 5.1 生成验证码输出
```
🔐 账户: weibaohui@hq.cmcc
🔢 验证码: 596065
⏱️  剩余: 24 秒
```

### 5.2 列表输出
```
📋 OTP 配置列表:
────────────────────────────────────────
1. weibaohui@hq.cmcc
   当前验证码: 890803 (剩余 24 秒)
   密钥: JXDW 7TXQ 2ZX2 C676 YWE7 MK7E F5J5 ZH2N
```

## 6. 使用场景

### 场景 1: 自动化测试脚本
```bash
# 生成验证码并直接用于 lc 命令
CODE=$(lc-otp-gen code 2>/dev/null | grep "验证码:" | sed 's/.*验证码: //')
lc otp verify "$CODE"
lc pr merge 123 --squash
```

### 场景 2: 多账户管理
```bash
# 管理工作账户和个人账户
lc-otp-gen add work@company.com SECRET1
lc-otp-gen add personal@gmail.com SECRET2
lc-otp-gen list
```

### 场景 3: CI/CD 集成
```bash
# 在自动化流程中使用
lc-otp-gen code work@company.com | grep "验证码:" | awk '{print $2}'
```

## 7. 依赖库

| 库 | 用途 | 版本 |
|----|------|------|
| github.com/spf13/cobra | CLI 框架 | latest |
| 标准库 (crypto/hmac, crypto/sha1, encoding/base32) | TOTP 算法 | 内置 |

## 8. 术语表

| 术语 | 说明 |
|------|------|
| TOTP | Time-based One-Time Password，基于时间的一次性密码 |
| Base32 | 一种编码方式，使用 A-Z 和 2-7 共 32 个字符 |
| 时间窗口 | TOTP 的时间间隔，默认 30 秒 |
| 密钥 | 20 字节随机数，Base32 编码后 32 个字符 |
