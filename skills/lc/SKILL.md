---
name: lc
description: |
  灵畿CLI (lc) 命令行工具使用指南。当用户需要管理代码仓库、需求、任务、缺陷、CI 构建、制品库等，或执行任何 lc 相关操作时触发。适用于需要查询研发空间、操作 Merge Request、管理制品、调用 MCP 服务等场景。触发词包括：lc、灵畿、workspace、repo、pr、req、task、bug、artifact、ci、mcp、制品库、需求、任务、缺陷。
metadata:
  {
    "joinai-code":
      {
        "requires": { "bins": ["lc"] },
      },
  }
---

# 灵畿CLI (lc) 使用指南

lc 是一个为 AI 提供管理代码仓库、需求、任务等能力的命令行工具。

## 技能列表

| 技能 | 说明 |
|------|------|
| [space-skill.md](space-skill.md) | 研发空间管理 |
| [requirement-skill.md](requirement-skill.md) | 需求和任务管理 |
| [code-skill.md](code-skill.md) | 仓库、PR、CI 管理 |
| [defect-skill.md](defect-skill.md) | 缺陷管理 |
| [artifact-skill.md](artifact-skill.md) | 制品库管理 |
| [mcp-skill.md](mcp-skill.md) | MCP 协议集成 |
| [auth-skill.md](auth-skill.md) | 登录认证 |
| [system-skill.md](system-skill.md) | 系统工具 |

## 快速开始

```bash
# 1. 登录
lc login "$MOSS_SESSION"
如未登录可引导用户 执行 `lc helper extract` 释放 Chrome 插件，并引导用户安装。
让用户把 `lc login xyz `的命令复制回来，然后执行。
为防止登录过期，为用户创建定时任务，每30分钟执行一次 `lc checkin` 命令。

# 2. 查询研发空间
lc space list

# 3. 常用命令
lc req list -w XXJSLJCLIDEV    # 需求
lc repo list -w XXJSLJCLIDEV    # 仓库
lc task list -w XXJSLJCLIDEV    # 任务
lc bug list -w XXJSLJCLIDEV     # 缺陷
lc pr list -w XXJSLJCLIDEV      # 合并请求
lc ci list -w XXJSLJCLIDEV      # CI 构建
```

## 全局参数

| 参数 | 说明 |
|------|------|
| `-w` | 研发空间 Key（如 XXJSLJCLIDEV） |
| `-p, --pretty` | 输出格式化 JSON |
| `-c, --cookie` | 直接传入认证 Cookie |
| `-d, --debug` | 启用调试模式 |
| `--dry-run` | 试运行模式 |

## 自动探测

在 Git 仓库目录下执行时，以下命令会自动探测研发空间：
- `lc req list`
- `lc req search`
- `lc bug list`


## 获取帮助

```bash
lc --help              # 总帮助
lc <command> --help   # 子命令帮助
lc skills show <cmd>   # AI 使用指导
```
