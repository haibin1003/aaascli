# 山东能力平台 CLI 助手 - 产品路线图

> 版本：v1.0  
> 更新日期：2026-04-15  
> 维护者：项目开发团队

---

## 一、当前状态（Baseline）

**已完成（v0.9）**
- [x] 能力查询：`ability list/search/view/services/my`
- [x] 服务查询：`service list/search/view`
- [x] 应用管理：`app list/auth-ability`
- [x] 订购授权：`ability order/service order/order list`
- [x] 知识库：`knowledge list/view/search`
- [x] 认证辅助：`login/helper extract/onboard`
- [x] 演示材料与开源平台提案
- [x] 跨平台构建（6 平台二进制）
- [x] Chrome 插件 Token 提取

**已知问题**
- [ ] 零单元测试覆盖（技术债务最高）
- [ ] 硬编码 URL/Header/成功码，扩展性差
- [ ] 知识库 embed 路径可能不同步
- [ ] 输出格式仅有 JSON，对人不够友好
- [ ] 无本地缓存，每次都走网络
- [ ] 写操作无幂等性，AI 决策负担重

---

## 二、近期目标（Q2 2026）- 夯实基础

| # | 任务 | 目标 | 验收标准 |
|---|------|------|----------|
| 1 | **单元测试与 Mock 测试** | 给 `internal/api/*` 和 `cmd/*` 补齐测试 | 核心包测试覆盖率 ≥ 60%，`go test ./...` 全绿 |
| 2 | **修复知识库同步机制** | 确保 `sdp knowledge list` 读取的与 `docs/knowledge/` 一致 | `make test` 中包含知识库同步校验测试 |
| 3 | **输出格式增强** | 支持 `--format table/json/yaml/markdown` | `list/search` 默认 table，AI 可显式指定 json |
| 4 | **本地缓存层** | 缓存能力/服务/应用列表到 `~/.sdp/cache.json` | 命中缓存时搜索秒级响应，支持 `--no-cache` 强制刷新 |
| 5 | **常量与错误码统一** | 提取 `BaseURL`、成功码、`User-Agent` 等 | 无硬编码字符串分散在 api 层，统一在 `internal/api/consts.go` |

### 近期里程碑定义
> **M1：可维护基线** — 完成 1、2、5，消除最高优先级技术债务。  
> **M2：体验升级** — 完成 3、4，CLI 从"能用"升级为"好用"。

---

## 三、中期目标（Q3 2026）- AI 体验与产品壁垒

| # | 任务 | 目标 | 验收标准 |
|---|------|------|----------|
| 6 | **幂等授权命令 `ensure-auth`** | `sdp app ensure-auth <app-id> <ability-id>` | 已授权时返回跳过，未授权时执行授权，支持 `--dry-run` |
| 7 | **交互式订购向导** | `sdp wizard order` 分步引导 | 减少 AI 对话轮次，用户只需回答 3-5 个问题 |
| 8 | **审批状态监听** | `sdp order watch --interval 60` | 轮询到状态变更时输出高亮提示，支持 `--once` |
| 9 | **能力对比与推荐** | `sdp ability compare <id1> <id2>` / `sdp ability recommend <keyword>` | 对比输出 Markdown 表格，推荐按语义相似度排序 |

### 中期里程碑定义
> **M3：AI 操作系统** — 完成 6、7，AI 从"执行单步命令"升级为"管理目标状态"。  
> **M4：智能化增强** — 完成 8、9，补齐主动服务与智能推荐能力。

---

## 四、远期目标（Q4 2026+）- 生态与平台化

| # | 任务 | 目标 | 验收标准 |
|---|------|------|----------|
| 10 | **MCP Server / Go SDK** | 封装 MCP 协议，让 Claude/Cursor 原生接入 | 提供 `mcp-sdp` 子命令或独立二进制，通过 MCP Inspector 验证 |
| 11 | **文档自动生成** | `sdp doc generate --format openapi/markdown` | 基于服务详情生成符合 OpenAPI 3.0 的接口文档 |
| 12 | **配置文件驱动 (IaC)** | `sdp sync --file sdp.yaml` | 支持 YAML 声明式批量授权，执行后输出变更清单 |

---

## 五、变更记录

| 日期 | 版本 | 变更说明 |
|------|------|----------|
| 2026-04-15 | v1.0 | 初始版本，基于当前技术债务和业务需求制定 |
