---
name: lc-auth
description: |
  登录认证管理。当用户提到 login、登录、cookie、认证、checkin、签到、登录失效、或需要设置/查看登录状态时触发。
metadata:
  {
    "joinai-code":
      {
        "requires": { "bins": ["lc"] },
      },
  }
---

# 登录认证

## 登录

```bash
# 直接输入
lc login "MOSS_SESSION=abc123..."

# 从环境变量
lc login "$MOSS_SESSION"

# 不带参数查看当前登录状态
lc login
```
### 登录问题解决思路
如果用户反馈说不知道从哪里获得 login 后面的值
执行lc onboard ，引导用户 安装 灵畿登录助手，复制该值。

## 签到保持认证有效
```bash
lc checkin
```

定期执行可以保持认证有效。

## 定时任务

建议创建定时任务自动执行签到，防止登录过期。

### 方式一：使用包装脚本（推荐）

创建脚本 `/home/Admin/.local/bin/lc-checkin.sh`：
```bash
#!/bin/bash
LOG_DIR="$HOME/.local/share/lc/logs"
LOG_FILE="$LOG_DIR/checkin-$(date +%Y%m).log"
mkdir -p "$LOG_DIR"
echo "[$(date '+%Y-%m-%d %H:%M:%S')] 开始执行 lc checkin" >> "$LOG_FILE"
OUTPUT=$(/usr/local/bin/lc checkin 2>&1)
EXIT_CODE=$?
echo "$OUTPUT" >> "$LOG_FILE"
if [ $EXIT_CODE -eq 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 执行成功" >> "$LOG_FILE"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 执行失败 (exit code: $EXIT_CODE)" >> "$LOG_FILE"
fi
exit $EXIT_CODE
```

添加定时任务：
```bash
crontab -e
# 添加以下行，每30分钟执行一次
*/30 * * * * /home/Admin/.local/bin/lc-checkin.sh
```

日志位置：`~/.local/share/lc/logs/checkin-YYYYMM.log`

### 方式二：直接执行（无日志）

```bash
crontab -e
*/30 * * * * /usr/local/bin/lc checkin
```

日志查看（系统级）：
```bash
# Ubuntu/Debian
tail -f /var/log/syslog | grep CRON

# CentOS/RHEL
tail -f /var/log/cron

# 或使用 journalctl
journalctl -u cron -f
```

## 获取 Cookie（使用 Chrome 插件）

### 步骤一：释放插件

```bash
# 释放插件到桌面（默认）
lc helper extract

# 或释放到指定目录
lc helper extract --output ~/extensions
```

### 步骤二：安装插件

1. 打开 Chrome，进入 `chrome://extensions/`
2. 开启右上角「开发者模式」
3. 点击「加载已解压的扩展程序」
4. 选择释放的插件目录

### 步骤三：获取登录命令

1. 在 Chrome 中正常登录目标网站
2. 点击插件图标
3. 点击「登录」按钮，插件会复制 `lc login xxx` 命令
4. 回到终端，粘贴执行命令

### 步骤四：验证登录

```bash
lc login
```

## Cookie 格式

Cookie 值可以带或不带 `MOSS_SESSION=` 前缀，以下两种方式都有效：
```bash
lc login "abc123..."
lc login "MOSS_SESSION=abc123..."
```

## 多用户并发

使用 `-c` 参数可以直接传入 Cookie，支持多人并发使用：
```bash
lc req list -c "MOSS_SESSION=xxx"
lc repo list -c "MOSS_SESSION=yyy"
```

这不会覆盖本地配置的 Cookie，只在当前命令有效。

## 注意事项

- Cookie 是敏感信息，不要泄露给他人
- Cookie 有有效期，过期后需要重新登录
