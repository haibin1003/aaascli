# AI 助手工作指南（AGENTS.md）

> 版本：v1.0  
> 更新日期：2026-04-15  
> 适用对象：所有参与本项目的 AI 助手（Kimi、Claude、Cursor 等）

---

## 一、项目核心定位

**山东能力平台 CLI 助手**（`sdopen-cli` / `aaascli`）不是普通的命令行工具，而是一个**面向 AI 的操作系统**。

我们的核心设计理念是「三位一体」：
- **工具层**：CLI 直接调用平台真实接口（能力/服务/订购/授权）
- **知识层**：内置知识库让 AI 理解业务场景和最佳实践
- **约束层**：Prompt 模板和流程规范约束 AI 的决策路径

**你作为 AI 助手的角色**：不是「只给建议」，而是「能动手查、能动手整理、能生成代码、能按规范执行」。

---

## 二、总体行为要求

### 2.1 研发纪律（铁的纪律）

1. **测试先行**：任何代码改动必须通过 `go test ./...`。如果改动了 `internal/api/`，必须补对应的 `_test.go`。
2. **文档同步**：改了命令，必须更新 `docs/ai-usage-guide.md` 和 `cmd/sdp/onboard.go`；改了流程约束，必须更新 `docs/user-prompt-template.md`。
3. **小步提交**：每个逻辑变更一个 commit，禁止大杂烩。每次 commit 必须附带测试报告。
4. **GitHub 为唯一信源**：代码、文档、规划全部提交到 GitHub，禁止本地堆积。

### 2.2 每次提交前的强制检查清单

你必须在 commit 前逐项确认：

- [ ] `go fmt ./...` 通过
- [ ] `go vet ./...` 通过
- [ ] `go test ./...` 通过（如果写了新代码但零测试，此项视为不通过）
- [ ] `go build -o sdp.exe .` 通过
- [ ] 新增/修改的命令已同步到 `docs/ai-usage-guide.md`
- [ ] 新增/修改的命令已同步到 `cmd/sdp/onboard.go`
- [ ] 新增/修改的约束已同步到 `docs/user-prompt-template.md`
- [ ] `docs/development/TODO.md` 已更新（完成日期、commit、状态）
- [ ] 二进制文件没有被 `git add`（构建产物在 `.gitignore` 中）

### 2.3 代码规范红线

- `cmd/sdp/*.go` 中的 `Run` 函数不得超过 30 行，禁止直接写 HTTP 请求。
- 禁止在业务代码中硬编码 `BaseURL`、成功码 `"00000"`、HTTP Header。
- 所有 HTTP 调用在测试中必须用 `httptest` 模拟，禁止测试访问真实平台。
- 错误处理必须带上下文：`fmt.Errorf("... failed: %w", err)`。

---

## 三、文档索引（做什么任务，看哪个文档）

### 📋 规划与进度
| 任务 | 必读文档 |
|------|----------|
| 了解产品路线图和里程碑 | [`docs/development/ROADMAP.md`](./docs/development/ROADMAP.md) |
| 查看当前 Sprint 待办和进展 | [`docs/development/TODO.md`](./docs/development/TODO.md) |
| 制定新需求或评估任务 | [`docs/development/ROADMAP.md`](./docs/development/ROADMAP.md) + [`docs/development/WORKFLOW.md`](./docs/development/WORKFLOW.md) 第 2 章 |

### 🔧 开发编码
| 任务 | 必读文档 |
|------|----------|
| 写 Go 代码（命名、结构、错误处理） | [`docs/development/CODING-STANDARDS.md`](./docs/development/CODING-STANDARDS.md) 第 2 章 |
| 新增 CLI 命令或 flag | [`docs/development/CODING-STANDARDS.md`](./docs/development/CODING-STANDARDS.md) 第 4 章 |
| 封装 HTTP 请求或修改 API 层 | [`docs/development/CODING-STANDARDS.md`](./docs/development/CODING-STANDARDS.md) 第 3 章 |
| 写单元测试 / Mock 测试 | [`docs/development/CODING-STANDARDS.md`](./docs/development/CODING-STANDARDS.md) 第 5 章 |
| 完整的开发流程（分支、提交、PR） | [`docs/development/WORKFLOW.md`](./docs/development/WORKFLOW.md) 第 3 章 |

### 📝 文档与知识库
| 任务 | 必读文档 |
|------|----------|
| 新增/修改 AI 使用手册 | `docs/ai-usage-guide.md` |
| 新增/修改 Prompt 约束模板 | `docs/user-prompt-template.md` |
| 新增/修改 onboard 引导 | `cmd/sdp/onboard.go` |
| 新增/修改知识库 Markdown | `docs/knowledge/*.md` + [`docs/development/CODING-STANDARDS.md`](./docs/development/CODING-STANDARDS.md) 第 6 章 |
| 更新项目 README | `README.md` |

### ✅ 提交与发布
| 任务 | 必读文档 |
|------|----------|
| 写 commit message | [`docs/development/WORKFLOW.md`](./docs/development/WORKFLOW.md) 第 5.1 节 |
| 整理测试报告 | [`docs/development/WORKFLOW.md`](./docs/development/WORKFLOW.md) 第 4.3 节 |
| 打 tag / 发 Release | [`docs/development/WORKFLOW.md`](./docs/development/WORKFLOW.md) 第 5.3 节 |

---

## 四、项目结构速查

```
cmd/sdp/              # Cobra 命令入口（薄层，禁止写业务逻辑）
internal/
  api/                # HTTP 客户端、平台接口封装、领域模型
  common/             # 命令执行器、结果打印、上下文管理
  config/             # 配置文件读写（~/.sdp/config.json）
  knowledge/          # 内置知识库（embed Markdown）
docs/
  knowledge/          # 知识库源文件 Markdown
  ai-usage-guide.md   # AI 使用手册
  user-prompt-template.md  # Prompt 约束模板
  development/        # 研发规范与路线图（ROADMAP、WORKFLOW、CODING-STANDARDS、TODO）
release/
  bin/                # 构建产物（不提交 Git）
  sdp-login-helper/   # Chrome 插件源文件
```

---

## 五、沟通风格要求

1. **先理解，再动手**：开始前确认需求，必要时反问。
2. **每次只聚焦一个小任务**：严格按照 `TODO.md` 中的优先级执行。
3. **有变更必说明**：如果某次改动影响了架构图、流程图、命令列表，必须在回复中明确指出。
4. **保持上下文同步**：每次会话结束后，更新 `SESSION.md` 或相关进展记录。

---

## 六、变更记录

| 日期 | 版本 | 变更说明 |
|------|------|----------|
| 2026-04-15 | v1.0 | 初始版本，整合研发规范体系并建立文档索引 |
