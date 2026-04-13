# 004 - PR merge 支持 gh 风格参数

## 变更记录表

| 版本 | 日期 | 变更人 | 变更内容 |
|------|------|--------|----------|
| 1.0 | 2026-03-18 | AI | 初始版本，定义 PR merge gh 风格参数支持需求 |

---

## 1. 需求背景

当前 `lc pr merge` 命令的参数风格与 `gh` CLI 工具不一致，增加了从 `gh` 迁移过来的用户的学习成本。为了降低迁移成本、提升用户体验，需要让 `lc pr merge` 支持 `gh pr merge` 风格的参数。

### 1.1 现状对比

| 功能 | gh 命令 | lc 当前命令 |
|------|---------|-------------|
| Squash 合并并删除分支 | `gh pr merge 18 --squash --delete-branch` | `lc pr merge 123 --type squash --remove-source` |
| 普通合并 | `gh pr merge 18 --merge` | `lc pr merge 123 --type merge` |
| Rebase 合并 | `gh pr merge 18 --rebase` | `lc pr merge 123 --type rebase` |

### 1.2 差异分析

| 参数 | gh | lc 当前 | 问题 |
|------|-----|---------|------|
| Squash 标志 | `--squash` | `--type squash` | lc 使用 `--type` 子参数，不如 gh 直观 |
| 删除分支标志 | `--delete-branch` | `--remove-source` / `-r` | 参数名不同，用户需要记忆两套参数 |
| 普通合并标志 | `--merge` | `--type merge` | gh 有独立的 `--merge` 标志 |
| Rebase 合并标志 | `--rebase` | `--type rebase` | gh 有独立的 `--rebase` 标志 |

---

## 2. 需求目标

### 2.1 功能目标

让 `lc pr merge` 命令兼容 `gh pr merge` 风格的参数，同时保留现有的 `--type` 参数以保持向后兼容。

### 2.2 具体需求

| 需求编号 | 需求描述 | 优先级 |
|----------|----------|--------|
| REQ-001 | 支持 `--squash` 参数，等同于 `--type squash` | 高 |
| REQ-002 | 支持 `--rebase` 参数，等同于 `--type rebase` | 高 |
| REQ-003 | 支持 `--merge` 参数，等同于 `--type merge`（默认行为） | 高 |
| REQ-004 | 支持 `--delete-branch` 参数，等同于 `--remove-source` | 高 |
| REQ-005 | 参数冲突处理：如果同时使用 `--squash` 和 `--type`，以明确指定的 `--type` 为准 | 中 |
| REQ-006 | 向后兼容：保留现有的 `--type` 和 `--remove-source` 参数 | 高 |

---

## 3. 功能边界

### 3.1 适用范围

- 仅影响 `lc pr merge` 命令
- 不影响其他 `lc pr` 子命令（如 `create`, `review`, `comment`, `view`, `list`）

### 3.2 参数互斥规则

- `--squash`、`--rebase`、`--merge` 三者互斥，同一时间只能使用其中一个
- `--delete-branch` 和 `--remove-source` 可以同时使用（效果相同）
- 如果同时使用了 `--squash` 和 `--type rebase`，应报错提示参数冲突

### 3.3 默认值

- 默认合并方式为 `merge`（与 gh 一致）
- 默认不删除源分支（与 gh 一致）

---

## 4. 命令使用示例

### 4.1 gh 风格命令（新增支持）

```bash
# Squash 合并并删除分支（gh 风格）
lc pr merge 123 --squash --delete-branch

# 普通合并
c pr merge 123 --merge

# Rebase 合并并删除分支
lc pr merge 123 --rebase --delete-branch

# 仅删除分支（不指定合并方式时使用默认的 merge）
lc pr merge 123 --delete-branch
```

### 4.2 lc 原有风格命令（保持兼容）

```bash
# 原有的 --type 方式仍然有效
lc pr merge 123 --type squash --remove-source
lc pr merge 123 --type merge
lc pr merge 123 --type rebase -r
```

### 4.3 混合风格命令（不推荐但应支持）

```bash
# gh 风格的合并方式 + lc 风格的分支删除
lc pr merge 123 --squash --remove-source

# lc 风格的合并方式 + gh 风格的分支删除
lc pr merge 123 --type rebase --delete-branch
```

---

## 5. 错误处理

### 5.1 参数冲突错误

当用户同时使用冲突的参数时，应给出清晰的错误提示：

```bash
$ lc pr merge 123 --squash --rebase
错误：不能同时使用 --squash 和 --rebase 参数
请只选择一种合并方式：--squash、--rebase 或 --merge
```

```bash
$ lc pr merge 123 --squash --type rebase
错误：参数冲突
--squash 与 --type rebase 不能同时使用
请只选择一种方式指定合并类型
```

### 5.2 未知参数错误

保持现有的错误处理机制，对未知参数给出提示。

---

## 6. 测试要求

### 6.1 功能测试

| 测试用例 | 输入 | 预期输出 |
|----------|------|----------|
| TC-001 | `lc pr merge 123 --squash` | 使用 squash 方式合并 |
| TC-002 | `lc pr merge 123 --rebase` | 使用 rebase 方式合并 |
| TC-003 | `lc pr merge 123 --merge` | 使用 merge 方式合并 |
| TC-004 | `lc pr merge 123 --delete-branch` | 合并后删除源分支 |
| TC-005 | `lc pr merge 123 --squash --delete-branch` | squash 合并并删除分支 |
| TC-006 | `lc pr merge 123 --type squash` | 向后兼容，正常使用 |
| TC-007 | `lc pr merge 123 -r` | 向后兼容，正常使用 |
| TC-008 | `lc pr merge 123 --squash --rebase` | 报错：参数冲突 |

### 6.2 帮助文档测试

确保 `lc pr merge --help` 显示新增的参数说明。

---

## 7. 影响范围

### 7.1 代码影响

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `cmd/lc/pr.go` | 修改 | 添加新参数定义和冲突检测逻辑 |
| `internal/common/flags.go` | 可能修改 | 如有需要，添加新参数的描述 |

### 7.2 文档影响

- 需要更新命令帮助文档
- 需要更新用户使用指南

### 7.3 兼容性影响

- 向后兼容：完全兼容，原有参数继续使用
- 向前兼容：无影响

---

## 8. 参考文档

- [gh pr merge 官方文档](https://cli.github.com/manual/gh_pr_merge)
- `docs/开发规范/` 目录下的相关规范
- `docs/文档编写规范/README.md` 文档编写规范
