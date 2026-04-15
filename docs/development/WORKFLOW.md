# 研发流程规范（SDP 项目）

> 版本：v1.0  
> 生效日期：2026-04-15  
> 适用范围：所有参与 `sdopen-cli` / `aaascli` 项目的开发者

---

## 一、总体原则

1. **小步快跑**：每个任务必须拆到可独立交付的最小粒度，完成后立即提交并记录。
2. **测试先行**：任何代码变更必须通过测试（单元测试 / 集成测试 / 手工测试）。
3. **文档同步**：代码改了，README / AI 使用手册 / onboard 引导 / Prompt 模板必须同步更新。
4. **GitHub 为唯一信源**：所有进展、讨论、版本均通过 GitHub 管理，禁止本地长时间堆积代码。

---

## 二、需求阶段

### 2.1 需求来源
- 用户反馈（Issue）
- 路线图规划（`docs/development/ROADMAP.md`）
- 紧急 Bug 修复

### 2.2 需求文档格式
每个新功能或重大优化必须先在 `docs/development/requirements/` 下创建需求文档：

```
docs/development/requirements/
├── YYYY-MM-DD-需求简述.md
```

文档必须包含：
1. **背景与目标**：为什么做这件事
2. **验收标准（Acceptance Criteria）**：如何判定完成
3. **影响范围**：哪些命令 / 模块 / 文档会受影响
4. **回退方案**：如果实现不了，Plan B 是什么
5. **接口/API 变更清单**：新增/修改/删除的 CLI 命令或 API 接口

### 2.3 需求评审
- 单人开发：自己写完后静默 10 分钟再读一遍，确保逻辑自洽。
- 多人开发：必须在 GitHub Issue 或 PR 中 @ 相关人评审，至少 1 人 approve。

---

## 三、开发阶段

### 3.1 分支策略
- `main`：唯一长期分支，始终保持可编译、可测试通过。
- 功能开发：从 `main` 切出 `feat/简述` 分支。
- Bug 修复：从 `main` 切出 `fix/简述` 分支。
- 禁止直接向 `main` push（当前单人可例外，但仍需走 PR 流程自我 review）。

### 3.2 代码规范
详见 [`CODING-STANDARDS.md`](./CODING-STANDARDS.md)。核心要点：
- Go 代码必须通过 `go fmt`、`go vet`、`golint`（或 `staticcheck`）。
- 不允许在 `internal/api/*` 中硬编码 URL、Header、成功判断码。
- 所有 `cobra` 命令的 `Run` 函数不得超过 30 行，业务逻辑下沉到 `internal/`。

### 3.3 开发文档
- 涉及新命令：必须更新 `docs/ai-usage-guide.md` 和 `cmd/sdp/onboard.go`。
- 涉及 Prompt 约束：必须更新 `docs/user-prompt-template.md`。
- 涉及知识库：必须更新 `docs/knowledge/` 并验证 embed 同步。

---

## 四、测试阶段

### 4.1 测试类型与要求

| 类型 | 要求 | 输出物 |
|------|------|--------|
| **单元测试** | `internal/api/*` 的每个新增/修改函数必须有对应测试 | `_test.go` 文件 |
| **Mock 测试** | 所有 HTTP 调用使用 `httptest` 模拟，禁止测试时访问真实平台 | `_test.go` 中的 `httptest.Server` |
| **命令行测试** | `cmd/*` 的核心路径通过 `cobra` 的 `ExecuteC` 做集成测试 | `cmd/sdp/*_test.go` |
| **手工验证** | 每次提交前，至少在一个真实环境（Windows/Linux/macOS 任选一个）运行一遍核心命令 | 测试报告截图或日志 |

### 4.2 测试通过标准
以下命令必须全部通过，才算"测试通过"：

```bash
go fmt ./...
go vet ./...
go test ./...
make build      # 或 go build -o sdp.exe .
```

### 4.3 测试报告模板
每次提交前，在 commit message body 中必须包含测试摘要：

```
Test Report:
- go test ./...: PASS (coverage: xx%)
- go vet ./...: PASS
- build: PASS
- manual check: PASS (平台: Windows, 命令: sdp ability list --size 3)
```

---

## 五、提交与发布阶段

### 5.1 Commit 规范
遵循 [Conventional Commits](https://www.conventionalcommits.org/)。

```
<type>(<scope>): <subject>

<body>

Test Report:
- go test ./...: PASS (coverage: xx%)
- go vet ./...: PASS
- build: PASS
- manual check: PASS (平台: xxx, 命令: xxx)
```

常用 type：
- `feat`：新功能
- `fix`：Bug 修复
- `docs`：文档变更
- `test`：测试补充
- `refactor`：重构（无行为变更）
- `chore`：构建/工具/依赖调整

### 5.2 PR / Merge 规范
- PR 标题与 commit 规范一致。
- PR description 必须链接对应的需求文档或 Issue。
- 合并前必须确认：
  1. 代码审查通过（self-review 或他人 review）
  2. 测试报告完整
  3. 文档已同步

### 5.3 版本发布
- 每个里程碑（M1/M2/...）完成后打一个 Git tag，如 `v0.9.1`、`v0.10.0`。
- 二进制发布走 GitHub Releases，不再提交到仓库。

---

## 六、记录与追踪

### 6.1 进展追踪
- 使用 `docs/development/TODO.md` 记录当前 Sprint 的待办、进行中和已完成任务。
- 每完成一个小任务，更新 `TODO.md` 并随代码一起提交。

### 6.2 知识沉淀
- 遇到平台接口变更、踩坑记录、调试技巧，优先补充到 `docs/knowledge/` 或 `docs/development/notes/`。
- 避免"只有某个人知道"的单点知识风险。

---

## 七、违规处理

如果某次提交：
- 没有测试报告 → **必须补充测试后 force-push 修正**
- 文档未同步 → **在下一个 commit 中补文档，禁止连续 3 次 commit 不补文档**
- 直接 push 了二进制 → **revert 并更新 `.gitignore`**
