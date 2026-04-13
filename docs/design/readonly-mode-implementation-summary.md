# 只读模式实现总结

## 概述

LC CLI 已添加只读模式（Read-Only Mode）功能，用于防止 AI 助手或用户误操作导致重要数据被删除或修改。

## 核心特性

- **默认安全**：新安装或首次使用时默认开启只读模式
- **显式控制**：通过 `lc readonly [on|off]` 命令显式切换模式
- **命令分类**：自动区分只读命令和写入命令
- **友好提示**：被拦截时返回清晰的错误信息和操作指引

## 实现文件

### 新增文件

| 文件 | 说明 |
|------|------|
| `cmd/lc/readonly.go` | 只读模式管理命令实现 |
| `internal/common/readonly.go` | 只读检查核心逻辑 |
| `docs/design/readonly-mode-design.md` | 设计方案文档 |

### 修改文件

| 文件 | 修改内容 |
|------|----------|
| `internal/config/config.go` | 添加 `Readonly bool` 字段和 `SaveConfig()` 函数 |
| `internal/common/executor.go` | 添加 `CommandName` 字段和只读检查逻辑 |
| `cmd/lc/req.go` | 为 create/update/delete 添加 `CommandName` |
| `cmd/lc/task.go` | 为 create/delete 添加 `CommandName` |
| `cmd/lc/bug.go` | 为 create/update-status/delete 添加 `CommandName` |
| `cmd/lc/repo.go` | 为 create/delete/disable-work-item-link/group add 添加 `CommandName` |
| `cmd/lc/pr.go` | 为 create/review/merge/comment/patch-comment 添加 `CommandName` |
| `cmd/lc/library.go` | 为 create/delete/folder create/upload/file delete 添加 `CommandName` |

## 命令分类

### 只读命令（始终允许）

```
list, view, search, detect, status, readonly, login, helper, version
```

### 写入命令（只读模式下拦截）

```
create, update, delete, merge, comment, patch-comment, upload
```

## 使用示例

```bash
# 查看只读模式状态
lc readonly

# 关闭只读模式（允许写入）
lc readonly off

# 开启只读模式（禁止写入）
lc readonly on
```

## 错误提示示例

当在只读模式下执行写入命令时：

```json
{
  "success": false,
  "error": {
    "code": "READONLY_MODE",
    "message": "当前处于只读模式，禁止执行写入操作",
    "details": "命令 'req create' 是写入操作，在只读模式下被禁止",
    "suggestion": "如需执行写入操作，请先关闭只读模式：\n  lc readonly off\n\n注意：关闭只读模式后，所有写入操作将直接生效，请谨慎操作。"
  }
}
```

## 架构设计

### 检查流程

1. 命令通过 `common.Execute()` 执行
2. 如果 `ExecuteOptions.CommandName` 不为空，调用 `CheckReadonlyForWrite()`
3. 检查配置中的 `readonly` 字段
4. 如果为只读模式，返回 `ReadonlyError`
5. `PrintError()` 将错误转换为统一的 JSON 格式

### 核心代码

```go
// internal/common/executor.go
func Execute(fn CommandFunc, opts ExecuteOptions) {
    // 检查只读模式
    if opts.CommandName != "" {
        if err := CheckReadonlyForWrite(opts.CommandName); err != nil {
            PrintError(err)
            os.Exit(1)
        }
    }
    // ... 原有逻辑
}

// internal/common/readonly.go
func CheckReadonlyForWrite(cmdName string) error {
    if IsReadonly() {
        return NewReadonlyError(cmdName)
    }
    return nil
}
```

## 测试验证

| 测试场景 | 结果 |
|----------|------|
| 默认只读状态 | ✅ 新安装默认为 `readonly: true` |
| 查询命令 | ✅ 只读模式下查询命令正常执行 |
| 写入拦截 | ✅ 只读模式下写入命令被拦截 |
| 状态切换 | ✅ `on`/`off` 切换正常 |
| 配置持久化 | ✅ 重启后配置保持 |

## 后续建议

1. **文档更新**：更新 `README.md` 和 `CLAUDE.md` 添加只读模式说明
2. **版本发布**：建议发布 v0.2.7 版本包含此功能
3. **用户通知**：通知用户新安装将默认开启只读模式
