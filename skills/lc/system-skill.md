---
name: lc-system
description: |
  系统工具命令。当用户提到 detect、自动探测、update、检查更新、version、版本、help、帮助时触发。
metadata:
  {
    "joinai-code":
      {
        "requires": { "bins": ["lc"] },
      },
  }
---

# 系统工具

## 自动探测 (detect)

### 自动探测
```bash
lc detect
```

在 Git 仓库目录下执行，自动输出探测到的信息。

### 探测并输出详细信息
```bash
lc detect --debug
```

**支持自动探测的命令**（在 Git 仓库目录下执行时，无需手动添加 `-w`）：
- `lc req list`
- `lc req search`
- `lc bug list`

## 检查更新
```bash
lc update
```

检查是否有新版本发布。

## 查看版本
```bash
lc version
```

## 获取帮助
```bash
lc --help              # 总帮助
lc <command> --help   # 子命令帮助
lc skills show <cmd>   # AI 使用指导
```

## 全局参数

| 参数 | 说明 |
|------|------|
| `-c, --cookie` | 直接传入认证 Cookie |
| `-d, --debug` | 启用调试模式 |
| `--dry-run` | 试运行模式 |
| `-k, --insecure` | 跳过 TLS 验证 |
| `-p, --pretty` | 输出格式化 JSON |
| `-w` | 研发空间 Key |
