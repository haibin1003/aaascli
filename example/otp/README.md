# OTP 二次验证使用示例

本文档提供 OTP 功能的实际使用示例。

---

## 目录

1. [基础配置示例](#1-基础配置示例)
2. [日常开发工作流](#2-日常开发工作流)
3. [自动化脚本示例](#3-自动化脚本示例)
4. [动态配置保护列表](#4-动态配置保护列表)

---

## 1. 基础配置示例

### 1.1 初始化 OTP

```bash
# 设置 OTP（会显示二维码和密钥）
lc otp setup

# 按提示操作：
# 1. 使用手机验证器应用扫描二维码
# 2. 输入手机显示的 6 位验证码
```

### 1.2 查看 OTP 状态

```bash
lc otp status
```

输出示例：
```json
{
  "enabled": true,
  "sessionExpiry": 5,
  "session": {
    "valid": false,
    "message": "无有效验证会话，执行危险操作前需要验证"
  },
  "protectedCommands": {
    "commands": ["repo delete", "readonly off"],
    "count": 2,
    "isCustom": false
  }
}
```

### 1.3 验证 OTP（创建会话）

```bash
# 方式1：命令行参数
lc otp verify 123456

# 方式2：交互式输入
lc otp verify
# 提示: 请输入 OTP 验证码: _
```

成功输出：
```json
{
  "verified": true,
  "verifiedAt": "2026-03-21T10:00:30+08:00",
  "expiresAt": "2026-03-21T10:05:30+08:00",
  "durationMin": 5,
  "message": "验证成功，会话有效期 5 分钟"
}
```

---

## 2. 日常开发工作流

### 场景：合并代码请求

```bash
# 1. 查看当前 OTP 状态
lc otp status
# {
#   "enabled": true,
#   "session": { "valid": false }
# }

# 2. 尝试合并（失败 - 需要 OTP）
lc pr merge 123 --squash
# {
#   "success": false,
#   "error": {
#     "code": "OTP_REQUIRED",
#     "message": "执行 合并代码请求 需要 OTP 二次验证"
#   }
# }

# 3. 验证 OTP（假设验证码为 123456）
lc otp verify 123456
# 验证成功，会话有效期 5 分钟

# 4. 再次执行合并（成功）
lc pr merge 123 --squash
# {
#   "success": true,
#   "data": { ... }
# }

# 5. 5 分钟内再次合并，无需重复验证
lc pr merge 124 --squash
# 直接执行成功
```

---

## 3. 自动化脚本示例

### 3.1 配合 lc-otp-gen 使用

```bash
#!/bin/bash
# example/otp/auto-merge.sh
# 自动化合并脚本

MR_ID=$1

# 检查参数
if [ -z "$MR_ID" ]; then
    echo "用法: $0 <mr-id>"
    exit 1
fi

# 获取当前 OTP 验证码
CODE=$(lc-otp-gen code 2>/dev/null | grep "验证码:" | head -1 | sed 's/.*验证码: //')

if [ ${#CODE} -ne 6 ]; then
    echo "获取 OTP 验证码失败"
    exit 1
fi

echo "当前 OTP 验证码: $CODE"

# 验证 OTP，创建会话
lc otp verify "$CODE" > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "OTP 验证失败"
    exit 1
fi

echo "OTP 验证成功"

# 执行合并
lc pr merge "$MR_ID" --squash --delete-branch
```

使用：
```bash
chmod +x example/otp/auto-merge.sh
./example/otp/auto-merge.sh 123
```

### 3.2 CI/CD 集成示例

```yaml
# .gitlab-ci.yml 示例
deploy:
  stage: deploy
  script:
    # 安装 lc 和 lc-otp-gen
    - make install
    - make install-otp-gen

    # 配置 OTP 账户（只需执行一次，后续复用）
    # lc-otp-gen add ci@example.com $OTP_SECRET

    # 验证 OTP
    - CODE=$(lc-otp-gen code | grep "验证码:" | sed 's/.*验证码: //')
    - lc otp verify "$CODE"

    # 执行部署（危险操作，受 OTP 保护）
    - lc readonly off --duration 30m
    - # ... 部署操作 ...
```

---

## 4. 动态配置保护列表

### 4.1 查看默认保护列表

```bash
lc otp config list
```

输出：
```json
{
  "protectedCommands": ["repo delete", "readonly off"],
  "isCustom": false,
  "message": "使用默认保护列表"
}
```

### 4.2 添加保护命令

```bash
# 添加 PR 合并到保护列表
lc otp config add "pr merge"

# 添加删除操作到保护列表
lc otp config add "req delete"
lc otp config add "task delete"
```

验证：
```bash
lc otp config list
# {
#   "protectedCommands": [
#     "repo delete",
#     "readonly off",
#     "pr merge",
#     "req delete",
#     "task delete"
#   ],
#   "isCustom": true,
#   "message": "使用自定义保护列表"
# }
```

### 4.3 移除保护命令

```bash
# 从自定义列表中移除
lc otp config remove "pr merge"
```

### 4.4 重置为默认

```bash
lc otp config reset
# 确认提示: 确定继续? [y/N]: y
```

### 4.5 直接编辑配置文件

```bash
# 编辑 ~/.lc/config.json
cat ~/.lc/config.json
```

配置示例：
```json
{
  "cookie": "...",
  "readonly": true,
  "otp": {
    "enabled": true,
    "secret": "HAEXHXIW6QQVFLUPYOVIGQTY7MYPZMKK",
    "verifiedAt": "2026-03-21T10:00:30+08:00",
    "sessionExpiryMinutes": 10,
    "protectedCommands": [
      "repo delete",
      "req delete",
      "task delete"
    ]
  }
}
```

---

## 5. 常见问题示例

### Q: 如何完全禁用 OTP？

```bash
# 需要先有有效会话
lc otp verify 123456
lc otp disable
```

### Q: 修改会话有效期？

目前只能通过直接编辑配置文件：

```bash
# 编辑 ~/.lc/config.json，修改 sessionExpiryMinutes
# 默认 5 分钟，建议范围：1-60
```

### Q: 为什么 `pr create` 不需要 OTP？

默认列表只包含**不可逆**操作：
- `repo delete` - 删除后无法恢复
- `readonly off` - 会开放所有写入操作

`pr create` 可以后续关闭，`pr merge` 可以通过 `git revert` 回滚，默认不保护。如需保护，请自行添加：

```bash
lc otp config add "pr merge"
lc otp config add "pr create"
```
