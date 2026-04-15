# 工程进展与待办追踪

> 更新日期：2026-04-15  
> 当前 Sprint：Sprint 1 - 夯实基础（M1：可维护基线）

---

## 一、当前 Sprint 目标

**Sprint 1：M1 - 可维护基线**
- 消除最高优先级技术债务
- 建立测试基线和研发规范
- 完成知识库同步与常量抽离

---

## 二、进行中（In Progress）

| # | 任务 | 负责人 | 计划完成 | 状态 |
|---|------|--------|----------|------|
| 1 | 建立研发规范文档（ROADMAP / WORKFLOW / CODING-STANDARDS / TODO / AGENTS） | - | 2026-04-15 | ✅ 已完成 |

---

## 三、待办（Backlog）- 按优先级排序

### P0 - 必须在本 Sprint 完成

| # | 任务 | 依赖 | 预估工时 | 状态 |
|---|------|------|----------|------|
| 1.1 | 为 `internal/api/client.go` 写 Mock 测试，验证 HTTP 请求构建和响应解析 | - | 4h | ✅ 已完成 |
| 1.2 | 为 `internal/api/ability.go` 写单元测试（ListAll / Search / View / Services / My / Order） | 1.1 | 6h | ✅ 已完成 |
| 1.3 | 为 `internal/api/service.go` 写单元测试（ListAll / Search / View / Order） | 1.1 | 5h | ⬜ 待办 |
| 1.4 | 为 `internal/api/app.go` 写单元测试（List / AuthAbility） | 1.1 | 4h | ⬜ 待办 |
| 1.5 | 为 `internal/api/order.go` 写单元测试（ListMyApplies） | 1.1 | 3h | ⬜ 待办 |
| 1.6 | 为 `cmd/sdp` 核心命令写集成测试（至少覆盖 ability list/search、service list/search） | 1.2-1.5 | 4h | ⬜ 待办 |
| 2.1 | 修复知识库 embed 路径：验证 `internal/knowledge/` 与 `docs/knowledge/` 是否同步 | - | 2h | ⬜ 待办 |
| 2.2 | 若不同步，建立构建时同步机制或修正 embed 路径 | 2.1 | 2h | ⬜ 待办 |
| 5.1 | 创建 `internal/api/consts.go`，抽离 `BaseURL`、成功码、通用 Header | - | 2h | ⬜ 待办 |
| 5.2 | 重构 `internal/api/*`，替换所有硬编码字符串为常量引用 | 5.1 | 3h | ⬜ 待办 |

### P1 - 本 Sprint 尽量完成，可延到 Sprint 2

| # | 任务 | 依赖 | 预估工时 | 状态 |
|---|------|------|----------|------|
| 3.1 | 设计并实现 `--format` 全局 flag（json/table/yaml/markdown） | 5.2 | 6h | ⬜ 待办 |
| 3.2 | 为 `ability list/search`、`service list/search`、`app list`、`order list` 实现 table 输出 | 3.1 | 4h | ⬜ 待办 |
| 4.1 | 设计本地缓存结构（`~/.sdp/cache.json`） | - | 2h | ⬜ 待办 |
| 4.2 | 实现缓存读写层和 TTL 过期机制 | 4.1 | 4h | ⬜ 待办 |
| 4.3 | 在 `ability list`、`service list`、`app list` 中接入缓存 | 4.2 | 3h | ⬜ 待办 |
| 4.4 | 实现 `--no-cache` 全局 flag | 4.2 | 1h | ⬜ 待办 |

---

## 四、已完成（Done）

| # | 任务 | 完成日期 | 对应 Commit | 备注 |
|---|------|----------|-------------|------|
| - | 能力订购与授权全流程自动化 | 2026-04-15 | `f474481` | ability order / service order / app auth-ability / order list |
| - | 演示材料与开源平台提案 | 2026-04-15 | `f474481` | presentation-report / proposal |
| - | 移除二进制构建产物跟踪 | 2026-04-15 | `bc8564c` | 新增 `.gitignore`，更新 README |
| - | 保存 Session 快照 | 2026-04-15 | `c9c3725` | `SESSION.md` |
| - | 建立研发规范与项目路线图 | 2026-04-15 | `8aac4dc` | ROADMAP / WORKFLOW / CODING-STANDARDS / TODO |
| - | 建立 AI 助手工作指南 | 2026-04-15 | `5039b13` | `AGENTS.md` + 文档索引 |
| - | client.go Mock 测试 + Client.BaseURL 可注入重构 | 2026-04-15 | `f6e7ad7` | `client_test.go` (10 cases) |
| - | ability.go 单元测试全覆盖 | 2026-04-15 | 当前提交 | `ability_test.go` (7 funcs, 18 cases) |

---

## 五、阻塞项与风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 零测试覆盖 | 高 | Sprint 1 核心目标，优先补齐 API 层测试 |
| 知识库路径不同步 | 中 | 下个任务立即排查，必要时改 embed 或加构建脚本 |
| 平台接口无文档 | 中 | 继续通过浏览器抓包维护接口契约，变化时靠测试第一时间发现 |

---

## 六、更新规则

1. **每次开始新任务**：将任务从"待办"移到"进行中"。
2. **每次完成任务**：
   - 将任务移到"已完成"
   - 填写完成日期和对应 Commit
   - 更新本文件并随代码一起 `git commit`
3. **每周回顾**：检查 Sprint 进度，必要时调整优先级或拆分任务。
