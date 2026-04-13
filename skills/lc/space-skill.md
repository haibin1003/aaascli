---
name: lc-workspace
description: |
  管理研发空间和项目。当用户提到 workspace、研发空间、空间列表、项目列表、space list、project list，或需要查询/指定研发空间时触发。这是使用 lc 的基础，大多数命令需要通过 -w 指定研发空间。
metadata:
  {
    "joinai-code":
      {
        "requires": { "bins": ["lc"] },
      },
  }
---

# 研发空间管理

## 查询研发空间列表

```bash
lc space list
```

返回当前用户可访问的所有研发空间。

## 查询空间下的项目

```bash
lc space project list -w <workspace-key>
# 例如
lc space project list -w XXJSLJCLIDEV
```

## 自动探测

在 Git 仓库目录下执行时，以下命令会自动探测研发空间，无需手动指定 `-w`：
- `lc req list`
- `lc req search`
- `lc bug list`

## 工作流程

```bash
# 1. 查询可用的研发空间
lc space list

# 2. 确定目标空间后，使用空间 Key 执行其他命令
lc req list -w XXJSLJCLIDEV
lc repo list -w XXJSLJCLIDEV
lc task list -w XXJSLJCLIDEV
```

## 核心概念

| 概念 | 说明 |
|------|------|
| 研发空间 Key | 空间的唯一标识，如 `XXJSLJCLIDEV` |
| 项目 Code | 项目编号，与空间 Key 类似 |

## 注意事项

- 空间 Key 是必填参数，通常为全大写字母组合
- 可使用 `-w` 简写代替 `--workspace-key`
- 所有涉及空间操作的命令都需要有效的登录状态
