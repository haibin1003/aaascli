# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
按照 [AGENTS.md](AGENTS.md) 中的约定开展工作

---

## 工作原则

### 写代码前必须确认设计方案（强制）

**在动手写任何代码之前，必须先向人类说明设计方案并得到明确确认，方可开始实现。**

#### 要求

1. **描述方案**：说明打算做什么、改哪些文件、核心逻辑是什么
2. **等待确认**：收到人类明确同意（如"好"、"可以"、"proceed"）后才开始写代码
3. **禁止自作主张**：不得在未经确认的情况下直接输出代码或修改文件

#### 适用场景

- 新增功能或命令
- 重构现有代码结构
- 涉及多个文件的改动
- 任何架构层面的决策

#### 例外

以下情况可直接执行，无需额外确认：
- 人类已在指令中描述了具体实现细节
- 修复明确的单行 bug（如语法错误、typo）

---

### 新功能必须使用分支开发（强制）

**任何新功能、新命令、新模块的开发，必须在独立分支上进行，严禁直接在 main / master 等主分支上提交功能代码。**

#### 要求

1. **开发前建分支**：开始编写新功能代码前，先从主分支切出一条功能分支
   ```bash
   git checkout -b feat/xxx
   ```
2. **主分支只接受合并**：主分支（main / master）上只允许出现 merge commit 或紧急 hotfix，不得直接 push 功能代码
3. **分支命名规范**：
   - 新功能：`feat/<简短描述>`，如 `feat/otp-guard`
   - 缺陷修复：`fix/<简短描述>`，如 `fix/readonly-check`
   - 重构：`refactor/<简短描述>`

#### 适用场景

- 新增任何命令或子命令
- 新增业务功能模块
- 较大范围的重构（涉及 3 个以上文件）

#### 例外

以下情况可在主分支直接操作：
- 文档、注释的修改（如 CLAUDE.md、README.md）
- 单行 typo / 配置值的紧急修正
- 人类明确指示在主分支操作

---

## 项目信息

### 研发空间配置

| 配置项 | 值 | 说明 |
|--------|-----|------|
| 研发空间 Key | `XXJSLJCLIDEV` | 灵畿cli研发空间标识 |
| 研发空间名称 | `灵畿cli研发` | 完整空间名称 |
| 项目代码 | `XXJSLJCLIDEV` | 关联项目编号 |
| 默认分支 | `master` | 主分支名称 |

### 研发空间用途说明

| 研发空间 | Key | 用途 |
|----------|-----|------|
| **灵畿cli研发** | `XXJSLJCLIDEV` | **项目正式开发空间** - 所有功能开发、需求管理、任务跟踪都在此空间进行 |
| **小白测研发项目** | `XXJSxiaobaice` | **测试专用空间** - 仅用于端到端测试、功能验证等测试活动 |

### 常用命令参考

```bash
# 列出灵畿cli研发空间下的需求
lc req list --workspace-key XXJSLJCLIDEV

# 列出该空间下的仓库
lc repo list --workspace-key XXJSLJCLIDEV

# 创建需求（在灵畿cli研发空间）
lc req create "需求名称" \
  --workspace-key XXJSLJCLIDEV \
  --project-code XXJSLJCLIDEV

# 创建任务
lc task create <需求objectId> "任务名称" \
  --workspace-key XXJSLJCLIDEV \
  --project-code XXJSLJCLIDEV
```

---

## 测试说明

### 端到端测试 (E2E)

运行端到端测试前，必须先构建并安装最新的 `lc` 二进制文件：

```bash
# 构建并安装最新版本（需要 sudo 权限）
sudo make install

# 验证安装版本
lc version

# 运行端到端测试
go test ./e2e/... -v
```

**注意**: E2E 测试会调用真实的 API，请确保已配置有效的登录凭证（`~/.lc/config.json`）。

---

## 开发规范

### 错误输出规范（强制）

**所有错误输出必须统一为 JSON 格式，禁止直接打印到 stderr。**

#### 要求

1. **禁止使用 `fmt.Fprintf(os.Stderr, ...)` 或 `fmt.Fprintln(os.Stderr, ...)` 直接输出错误信息**
   - 错误：破坏命令行工具的 JSON 输出统一性
   - 所有错误必须通过 `common.PrintError()` 输出

2. **使用统一错误处理函数**
   ```go
   // 正确
   if err := tryAutoDetectForXXX(cmd); err != nil {
       common.PrintError(common.HandleAutoDetectError(err, "-w, --workspace-key"))
       os.Exit(1)
   }
   ```

3. **自动探测函数中禁止打印日志**
   ```go
   // 错误做法 - 不要这样做
   fmt.Fprintf(os.Stderr, "[自动探测] 研发空间 Key: %s\n", workspaceKey)

   // 正确做法 - 静默设置变量即可
   workspaceKey = ctx.WorkspaceKey
   ```

4. **错误类型**
   - 使用 `common.AutoDetectError` 表示自动探测失败
   - 使用 `common.NewAutoDetectError()` 创建错误
   - 通过 `WithDetails()`, `WithSuggestion()`, `WithMissing()` 添加详细信息

#### 相关文件

- `internal/common/errors.go`: 错误类型定义
- `internal/common/executor.go`: `PrintError()` 实现
- `internal/common/autodetect.go`: `HandleAutoDetectError()` 实现

#### 检查清单

新增或修改命令时，确保：
- [ ] 没有直接调用 `fmt.Fprintf(os.Stderr, ...)`
- [ ] 自动探测函数中没有直接打印语句
- [ ] 错误通过 `common.PrintError()` 输出
- [ ] 命令输出保持 JSON 统一格式

---