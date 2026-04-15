# Session Snapshot - 2026-04-15

## 本次会话完成的工作

### 1. 功能开发（已提交并推送到 GitHub）
- 完成能力订购与授权全流程自动化：
  - `sdp ability order <ability-id>` — 提交能力订购申请
  - `sdp service order <service-id>` — 订购服务（一步完成订购+授权）
  - `sdp app auth-ability <app-id>` — 为应用授权能力（支持 BOMC 工单关联）
  - `sdp order list` — 查询订购/授权申请及审批状态
- 新增 `internal/api/order.go`、`internal/api/payloads.go` 对接真实平台接口

### 2. 文档与知识库（已提交并推送）
- 新增 `docs/knowledge/06-能力与服务订购流程指南.md`
- 更新 `docs/ai-usage-guide.md` 和 `docs/user-prompt-template.md`，补充订购/授权场景
- 更新 `cmd/sdp/onboard.go` 引导内容，补充新命令和场景 5/6

### 3. 演示材料（已提交并推送）
- 更新 `docs/presentation-report.html`：
  - 核心功能区增加「工具 × 知识 × 约束」融合示意
  - 新增「项目结构与交付物」章节（目录树 + 说明卡片）
  - 新增「给 AI 助手的 onboarding 信」章节
  - 增加打印分页控制 CSS，解决 PDF 截断问题
- 重新生成 `docs/presentation-report.pdf`
- 新增 `docs/open-source-platform-proposal.html` 和 `.pdf`

### 4. 仓库清理（已提交并推送）
- 移除二进制构建产物跟踪：`release/bin/*`、`release/sdp-package.zip`、`sdp.exe`
- 新增 `.gitignore`，防止编译产物再次误提交
- 更新 `README.md` 安装方式，改为指向 GitHub Releases

## 仓库状态
- **分支**：`main`
- **远程**：`github.com:haibin1003/aaascli.git`
- **状态**：working tree clean，所有变更已 push 到 origin/main

## 未提交到 Git 的本地文件（构建产物）
- `release/bin/sdp-*`（6 个平台二进制，约 75MB）
- `release/sdp-package.zip`（约 41MB）
- `sdp.exe`、`release/bin/sdp.exe`、`release/bin/sdp-test.exe`
- 以上文件因 `.gitignore` 被忽略，仅在本地存在

## 关键概念
- **三位一体**：工具（CLI）+ 知识（内置知识库）+ 约束（Prompt 模板/决策树）
- **零侵入**：纯客户端实现，无需平台后端改代码、加接口、动数据库
