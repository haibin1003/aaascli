# lc-otp-gen 命令行 OTP 生成器设计文档

## 变更记录表

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|---------|------|
| v1.0 | 2026-03-21 | 初始版本 | AI Assistant |

---

## 1. 架构设计

### 1.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLI 命令层                                │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ add      │ │ code     │ │ list     │ │ remove   │           │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘           │
│       └────────────┴────────────┴────────────┘                  │
│                              │                                  │
│                         cobra 框架                              │
└──────────────────────────────┼──────────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────────┐
│                         业务逻辑层                              │
│                    ┌─────────┴────────┐                        │
│                    │  OTPManager      │                        │
│                    │  • AddConfig     │                        │
│                    │  • GetCode       │                        │
│                    │  • ListConfigs   │                        │
│                    │  • RemoveConfig  │                        │
│                    └──────────────────┘                        │
└──────────────────────────────┼──────────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────────┐
│                         算法层                                  │
│                    ┌─────────┴────────┐                        │
│                    │  TOTPGenerator   │                        │
│                    │  • GenerateCode  │                        │
│                    │  RFC 6238        │                        │
│                    └──────────────────┘                        │
└──────────────────────────────┼──────────────────────────────────┘
                               │
┌──────────────────────────────┼──────────────────────────────────┐
│                        配置存储层                               │
│              ┌───────────────┴───────────────┐                  │
│              │   ~/.lc-otp-gen/config.json   │                  │
│              │  {                              │                  │
│              │    "otps": [                    │                  │
│              │      {"account": "...",         │                  │
│              │       "secret": "..."}         │                  │
│              │    ]                            │                  │
│              │  }                              │                  │
│              └─────────────────────────────────┘                  │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 核心组件

| 组件 | 文件路径 | 职责 |
|------|---------|------|
| CLI 命令 | `cmd/lc-otp-gen/main.go` | 命令行接口，使用 cobra 框架 |
| 配置管理 | `cmd/lc-otp-gen/main.go` | 配置的增删改查和持久化 |
| TOTP 算法 | `cmd/lc-otp-gen/main.go` | RFC 6238 标准实现 |

## 2. 详细设计

### 2.1 数据结构

```go
// OTPConfig 单个 OTP 配置
type OTPConfig struct {
    Account string `json:"account"`  // 账户名
    Secret  string `json:"secret"`   // Base32 密钥
    Issuer  string `json:"issuer"`   // 发行方（默认为"灵畿CLI"）
}

// Configs 配置集合
type Configs struct {
    OTPs []OTPConfig `json:"otps"`
}
```

### 2.2 TOTP 算法实现

遵循 RFC 6238 标准：

```go
func generateTOTP(secret string) (string, int) {
    // 1. 解码 Base32 密钥
    key, _ := base32.StdEncoding.DecodeString(secret)

    // 2. 计算时间窗口计数器（30秒）
    counter := uint64(math.Floor(float64(time.Now().Unix()) / 30))

    // 3. HMAC-SHA1 计算
    mac := hmac.New(sha1.New, key)
    binary.Write(mac, binary.BigEndian, counter)
    hash := mac.Sum(nil)

    // 4. 动态截断
    offset := hash[len(hash)-1] & 0x0F
    code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF

    // 5. 取模得到 6 位数字
    code = code % 1000000

    // 6. 计算剩余时间
    remaining := 30 - (time.Now().Unix() % 30)

    return fmt.Sprintf("%06d", code), int(remaining)
}
```

### 2.3 配置文件管理

```go
// 获取配置文件路径
func getConfigPath() string {
    if cfgFile != "" {
        return cfgFile  // 自定义路径
    }
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".lc-otp-gen", "config.json")
}

// 加载配置
func loadConfigs() (*Configs, error) {
    data, _ := os.ReadFile(getConfigPath())
    var configs Configs
    json.Unmarshal(data, &configs)
    return &configs, nil
}

// 保存配置（权限 0600）
func saveConfigs(configs *Configs) error {
    data, _ := json.MarshalIndent(configs, "", "  ")
    return os.WriteFile(getConfigPath(), data, 0600)
}
```

### 2.4 命令设计

#### add 命令
```go
Usage: lc-otp-gen add [account] [secret]

流程:
1. 验证参数数量
2. 清理密钥（转大写、去空格）
3. 检查账户是否已存在
4. 如存在则更新，否则添加
5. 保存配置
6. 输出成功信息
```

#### code 命令
```go
Usage: lc-otp-gen code [account]

流程:
1. 加载配置
2. 如果没有指定账户且只有一个配置，使用该配置
3. 如果有多个配置但未指定账户，列出所有配置
4. 查找指定账户
5. 调用 generateTOTP 生成验证码
6. 输出账户名、验证码、剩余时间
```

#### list 命令
```go
Usage: lc-otp-gen list

流程:
1. 加载所有配置
2. 遍历每个配置
3. 生成当前验证码
4. 格式化输出（账户名、验证码、剩余时间、格式化密钥）
```

#### remove 命令
```go
Usage: lc-otp-gen remove [account]

流程:
1. 验证参数
2. 加载配置
3. 查找并删除指定账户
4. 保存配置
```

## 3. 安全设计

### 3.1 文件权限
- 配置文件权限：`0600`（仅所有者可读写）
- 配置目录权限：`0755`

### 3.2 密钥处理
- 自动转大写，去除空格，便于复制粘贴
- 存储原始密钥，不加密（本地使用）

### 3.3 错误处理
- 密钥格式错误时返回 "ERROR"
- 配置不存在时给出友好提示

## 4. 输出设计

### 4.1 颜色与图标
使用 Unicode 图标增强可读性：
- 🔐 账户
- 🔢 验证码
- ⏱️  剩余时间
- 📋 列表
- ✅ 成功
- ❌ 错误

### 4.2 密钥格式化显示
```go
func formatSecret(secret string) string {
    // 每 4 个字符一组
    for i := 0; i < len(secret); i += 4 {
        result.WriteRune(secret[i])
        if i > 0 && i%4 == 0 {
            result.WriteByte(' ')
        }
    }
}
// 输出: JXDW 7TXQ 2ZX2 C676 YWE7 MK7E F5J5 ZH2N
```

## 5. 自动化测试集成

### 5.1 脚本示例
```bash
#!/bin/bash

# 获取验证码
CODE=$(lc-otp-gen code 2>/dev/null | grep "验证码:" | sed 's/.*验证码: //')

# 验证格式
if [ ${#CODE} -eq 6 ]; then
    # 使用验证码
    lc otp verify "$CODE"
    lc pr merge 123 --squash
fi
```

### 5.2 多账户选择
```bash
# 指定账户
CODE=$(lc-otp-gen code work@company.com | grep "验证码:" | awk '{print $2}')
```

## 6. 构建与安装

### 6.1 独立构建
```bash
cd cmd/lc-otp-gen
go build -o lc-otp-gen .
```

### 6.2 系统安装
```bash
sudo cp lc-otp-gen /usr/local/bin/
```

### 6.3 Makefile 集成
```makefile
## build-otp-gen: 构建 OTP 生成器工具
build-otp-gen:
    go build -o $(BUILD_DIR)/lc-otp-gen ./cmd/lc-otp-gen

## install-otp-gen: 安装 OTP 生成器到系统
install-otp-gen: build-otp-gen
    install -d /usr/local/bin && cp $(BUILD_DIR)/lc-otp-gen /usr/local/bin/

## test-otp: 运行 OTP 功能自动化测试
test-otp: install-otp-gen
    ./scripts/test-otp.sh
```

## 7. 依赖

| 库 | 用途 | 版本 |
|----|------|------|
| github.com/spf13/cobra | CLI 框架 | v1.10.2 |
| 标准库 | TOTP 算法、文件操作 | 内置 |
