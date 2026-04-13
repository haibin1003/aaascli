# 🎬 灵畿 × Claude Code 全场景演示方案

> **重要说明**：本文档中所有 `lc` 命令和 `git` 命令均为**真实可执行操作**，代码编写部分使用**伪代码示意**。

---

## 一、故事背景：智慧零售平台的诞生

### 场景设定

**团队名称**："极光科技" - 一家专注于智慧零售解决方案的创新公司

**项目背景**：公司决定开发一个「智能库存预警系统」，帮助零售商实时监控库存状态，当商品库存低于安全阈值时自动触发补货提醒。

**核心角色**：
- 👨‍💼 **产品经理 张明** - 负责需求规划
- 👩‍💻 **工程师 李娜** - 负责后端开发（AI 助手扮演）
- 🎨 **设计师 王强** - 负责前端设计（简化演示）
- 🔍 **技术负责人 赵总** - 负责代码审查

---

## 二、演示目标

展示 **Vibe Coding + 灵畿平台** 如何打通软件研发的完整闭环：

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│  创建仓库  │ → │  需求创建  │ → │  任务分解  │ → │  AI开发   │ → │  PR审查   │
│  (lc)   │    │  (lc)   │    │  (lc)   │    │(Claude) │    │  (lc)   │
└─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘
```

**核心优势**：先创建仓库后，后续命令可利用**自动探测**功能，大幅简化操作。

---

## 三、图例说明

| 标记 | 含义 |
|------|------|
| ⚡️ **真实操作** | 必须真实执行的 lc 或 git 命令 |
| 💬 **伪代码示意** | 仅示意代码逻辑，不实际执行 |
| 📺 **展示说明** | 向观众解释的内容 |

---

## 四、详细演示流程

### 🎬 第一幕：创建代码仓库（先准备基础设施）

#### 场景
工程师李娜首先在灵畿平台上创建项目代码仓库，为后续开发做准备。

#### 演示动作

**⚡️ 1.1 查看研发空间和个人代码组**
```bash
# 【真实操作】查看可用的研发空间
lc space list

# 【真实操作】查看个人代码组（获取 group-id）
lc repo group personal
```

📺 **向观众展示**：
- 研发空间列表中有"小白测研发项目"（XXJSxiaobaice）
- 个人代码组 ID 为 617927

---

**⚡️ 1.2 创建演示专用仓库**
```bash
# 【真实操作】创建仓库（使用时间戳确保唯一性）
DEMO_REPO="inventory-demo-$(date +%m%d-%H%M%S)"
echo "创建仓库: $DEMO_REPO"

lc repo create "$DEMO_REPO" \
  --group-id 617927 \
  --workspace-key XXJSxiaobaice
```

📺 **展示输出**：
```json
{
  "success": true,
  "data": {
    "gitProjectId": 45138,
    "name": "inventory-demo-0319-102502",
    "tenantHttpPath": "http://code-xxjs.rdcloud.4c.hq.cmcc/osc/XXJS/weibaohui-hq.cmcc/inventory-demo-0319-102502.git"
  }
}
```

📺 **向观众说明**：
- 记录 `gitProjectId`（如 45138）
- 使用 `tenantHttpPath`（hq.cmcc 域名，外网可访问）

---

**⚡️ 1.3 关闭"提交关联工作项"功能（重要）**

📺 **问题说明**：
> 默认情况下，仓库开启了"提交代码需关联工作项"功能，这会导致 `git push` 失败。我们需要先关闭这个功能。

```bash
# 【真实操作】关闭提交关联工作项功能（在推送代码前执行）
lc repo disable-work-item-link \
  --git-project-id 45138 \
  --workspace-key XXJSxiaobaice
```

📺 **说明**：
- 关闭后，git commit 无需强制关联需求/任务编号
- 适合演示环境，生产环境可根据需要开启

---

**⚡️ 1.4 克隆仓库并初始化**
```bash
# 【真实操作】使用 tenantHttpPath 克隆仓库
git clone "http://code-xxjs.rdcloud.4c.hq.cmcc/osc/XXJS/weibaohui-hq.cmcc/inventory-demo-0319-102502.git"

# 【真实操作】进入仓库目录
cd inventory-demo-0319-102502

# 【真实操作】初始化 README
echo "# 智能库存预警系统" > README.md
git add README.md
git commit -m "chore: 初始化仓库"

# 【真实操作】推送到远程
# 注意：已关闭提交关联工作项功能，推送成功
git push -u origin master
```

📺 **向观众展示**：
- 关闭提交关联功能后，推送成功
- master 分支已创建

---

**⚡️ 1.4 验证自动探测功能**
```bash
# 【真实操作】在仓库目录下执行，验证自动探测
lc detect
```

📺 **展示输出并解释**：
```json
{
  "success": true,
  "data": {
    "workspaceKey": "XXJSxiaobaice",
    "workspaceName": "小白测研发项目",
    "repository": {
      "gitProjectId": 45138,
      "name": "inventory-demo-0319-102502"
    },
    "matched": true
  }
}
```

**⚡️ 1.5 验证自动探测功能**
```bash
# 【真实操作】在仓库目录下执行，验证自动探测
lc detect
```

📺 **展示输出并解释**：
```json
{
  "success": true,
  "data": {
    "workspaceKey": "XXJSxiaobaice",
    "workspaceName": "小白测研发项目",
    "repository": {
      "gitProjectId": 45138,
      "name": "inventory-demo-0319-102502"
    },
    "matched": true
  }
}
```

📺 **关键说明**：
> 现在我们在仓库目录下，后续 lc 命令可以**自动探测**到：
> - `workspaceKey`: XXJSxiaobaice
> - `gitProjectId`: 45138
> - `sourceBranch`: 当前 Git 分支

---

### 🎬 第二幕：需求诞生（5分钟）

#### 场景
产品经理张明在晨会上提出了智能库存预警系统的构想，需要在灵畿平台上创建正式需求。

#### 演示动作

**⚡️ 2.1 创建产品需求**
```bash
# 【真实操作】在仓库目录外或指定 workspace-key
lc req create "智能库存预警系统 - Phase 1" \
  --workspace-key XXJSxiaobaice \
  --project-code R24113J3C04
```

📺 **展示输出并记录**：
```json
{
  "success": true,
  "data": {
    "key": "XXJSxiaobaice-1203",
    "name": "智能库存预警系统 - Phase 1",
    "objectId": "AxVyCddjGi"
  }
}
```

📺 **向观众说明**：
- 记录 `objectId`（如 AxVyCddjGi），后续创建任务需要用到

---

### 🎬 第三幕：任务规划（5分钟）

#### 场景
李娜（AI 工程师）接收到需求后，开始分解技术任务。

#### 演示动作

**⚡️ 3.1 创建开发任务**
```bash
# 【真实操作】设置需求 ID 变量
REQ_ID="AxVyCddjGi"

# 【真实操作】任务1：数据库设计
lc task create $REQ_ID "设计库存预警数据库表结构" \
  --workspace-key XXJSxiaobaice \
  --project-code R24113J3C04

# 【真实操作】任务2：API 开发
lc task create $REQ_ID "开发库存监控 API" \
  --workspace-key XXJSxiaobaice \
  --project-code R24113J3C04

# 【真实操作】任务3：通知服务
lc task create $REQ_ID "实现多渠道通知服务" \
  --workspace-key XXJSxiaobaice \
  --project-code R24113J3C04
```

📺 **展示输出**：
- 三个任务创建成功，记录各自的 `objectId`

---

**⚡️ 3.2 查看任务列表**
```bash
# 【真实操作】查看任务列表确认创建成功
lc task list --workspace-key XXJSxiaobaice -l 5
```

---

### 🎬 第四幕：AI 开发（15分钟）⭐ 核心环节

#### 场景
李娜开始使用 Claude Code 进行实际的代码开发，同时通过 lc 与灵畿平台保持同步。

#### 演示动作

**⚡️ 4.1 初始化项目**
```bash
# 【真实操作】进入仓库目录（自动探测生效）
cd inventory-demo-0319-102502

# 【真实操作】检查当前仓库上下文
lc detect

# 【真实操作】创建特性分支
git checkout -b feature/inventory-alert-system
```

📺 **向观众展示**：
- 当前在仓库目录下
- 已切换到 feature/inventory-alert-system 分支

---

**💬 4.2 与 Claude Code 协作开发（伪代码示意）**

📺 **用户（对 Claude 说）**：
> Claude，我需要根据灵畿平台的需求 XXJSxiaobaice-1203 开发库存预警系统。首先帮我设计数据库模型。

📺 **Claude 生成代码（展示伪代码）**：
```go
// internal/models/inventory.go - 【伪代码示意】
package models

import "time"

// Inventory 库存表
type Inventory struct {
    ID            uint64    `json:"id"`
    ProductID     string    `json:"product_id"`
    ProductName   string    `json:"product_name"`
    CurrentStock  int       `json:"current_stock"`
    Threshold     int       `json:"threshold"`
    Status        string    `json:"status"`
    LastAlertAt   *time.Time `json:"last_alert_at"`
    CreatedAt     time.Time `json:"created_at"`
}

// AlertRecord 预警记录表
type AlertRecord struct {
    ID              uint64    `json:"id"`
    ProductID       string    `json:"product_id"`
    Severity        string    `json:"severity"`
    Status          string    `json:"status"`
    CreatedAt       time.Time `json:"created_at"`
}
```

📺 **用户（对 Claude 说）**：
> 很好，现在帮我开发核心业务逻辑 - 库存检查服务。

📺 **Claude 生成代码（展示伪代码）**：
```go
// internal/service/inventory_service.go - 【伪代码示意】
package service

// CheckAndAlert 检查库存并触发预警
func (s *InventoryService) CheckAndAlert(productID string) error {
    // 1. 查询库存
    // 2. 判断是否低于阈值
    // 3. 创建预警记录
    // 4. 发送通知
    return nil
}
```

📺 **实际演示中**：
- Claude 实际生成代码并写入文件
- 为节省时间，这里用伪代码示意核心逻辑

---

**⚡️ 4.3 提交代码（真实操作）**
```bash
# 【真实操作】添加代码文件（实际由 Claude 生成）
mkdir -p internal/models internal/service internal/notification

# 【真实操作】添加所有文件
git add .

# 【真实操作】提交代码
git commit -m "feat(inventory): 实现智能库存预警系统 Phase 1

- 添加库存数据模型 (inventory, alert_record)
- 实现库存检查与预警服务
- 集成多渠道通知（邮件/钉钉/短信）
- 支持预警级别自动判定

关联需求: XXJSxiaobaice-1203"

# 【真实操作】推送到远程
git push -u origin feature/inventory-alert-system
```

📺 **向观众展示**：
- 代码已提交到 feature/inventory-alert-system 分支
- 远程仓库已有提交记录

---

### 🎬 第五幕：创建 PR（自动探测生效）

#### 场景
功能开发完成，李娜准备提交代码并创建合并请求。

#### 演示动作

**⚡️ 5.1 创建 PR（自动探测简化参数）**
```bash
# 【真实操作】确保在仓库目录下
cd inventory-demo-0319-102502

# 【真实操作】创建 PR - 在仓库目录下，自动探测生效
# 无需指定 --git-project-id、--workspace-key、--source
lc pr create \
  --title "feat(inventory): 实现智能库存预警系统 Phase 1" \
  --body "## 变更说明

本次 MR 实现了智能库存预警系统的核心功能：

### ✨ 新功能
- 库存实时监控与阈值检查
- 多级预警机制（高/中/低）
- 多渠道通知集成（邮件、钉钉、短信）
- 预警历史记录追踪

### 🔗 关联信息
- 需求: XXJSxiaobaice-1203
- 任务: XXJSxiaobaice-1204, XXJSxiaobaice-1205, XXJSxiaobaice-1206" \
  --target master \
  --remove-source
```

📺 **展示输出**：
```json
{
  "success": true,
  "data": {
    "iid": 1,
    "title": "feat(inventory): 实现智能库存预警系统 Phase 1",
    "sourceBranch": "feature/inventory-alert-system",
    "targetBranch": "master",
    "state": "opened"
  }
}
```

📺 **关键说明**：
> 💡 **自动探测优势**：在仓库目录下执行时，无需指定：
> - `--git-project-id`（自动从 Git remote 探测）
> - `--workspace-key`（自动从仓库关联探测）
> - `--source`（自动从当前分支探测）

---

### 🎬 第六幕：代码审查闭环（完整版）

#### 场景
技术负责人赵总收到审查通知，开始审查代码。演示**完整的审查闭环**：发现问题 → AI修复 → 确认合并。

#### 演示动作

**⚡️ 6.1 查看 PR 列表（自动探测）**
```bash
# 【真实操作】在仓库目录下执行，自动探测生效
lc pr list
```

---

**⚡️ 6.2 查看 PR 详情**
```bash
# 【真实操作】查看 MR 详情（假设 MR iid 为 1）
lc pr view 1
```

---

**⚡️ 6.3 【第六幕-A】审查者发现问题（第一轮评论）**

```bash
# 【真实操作】审查者添加评论，指出需要修复的问题
lc pr comment 1 --body "## 🎬 [第六幕-A] 代码审查 - 发现问题 ❌

### 需要修复的问题

1. **缺少错误处理**：CheckAndAlert 函数没有处理数据库查询失败的情况
2. **缺少日志记录**：关键操作没有日志，不利于排查问题
3. **代码注释不足**：函数缺少说明文档

### 💡 修复建议

请补充：
- [ ] 错误返回和日志记录
- [ ] 关键步骤的注释说明
- [ ] 边界条件处理

@开发者 请修复后回复此评论。"
```

---

**⚡️ 6.4 【第六幕-B】AI读取评论并修复代码**

📺 **AI/开发者读取评论**：
```bash
# 查看评论内容
lc pr view 1 --show-comments
```

📺 **AI分析评论，生成修复代码**（伪代码示意）：
```go
// 修复后的 internal/service/inventory_service.go
func (s *InventoryService) CheckAndAlert(productID string) error {
    // 1. 查询库存（增加错误处理）
    inventory, err := s.db.GetInventory(productID)
    if err != nil {
        logger.Error("查询库存失败", "productID", productID, "error", err)
        return fmt.Errorf("查询库存失败: %w", err)
    }

    // 2. 判断阈值（增加日志）
    if inventory.CurrentStock <= inventory.Threshold {
        logger.Info("库存低于阈值，触发预警",
            "productID", productID,
            "currentStock", inventory.CurrentStock,
            "threshold", inventory.Threshold)
        // 创建预警记录...
    }

    return nil
}
```

⚡️ **提交修复**：
```bash
# 【真实操作】提交修复代码
git add .
git commit -m "fix: 补充错误处理、日志记录和注释

- 添加数据库查询错误处理
- 增加关键操作日志记录
- 补充代码注释说明

Fixes: MR #1 评论反馈 🎬 [第六幕-A]"

git push
```

---

**⚡️ 6.5 【第六幕-C】回复评论（修复完成）**

```bash
# 【真实操作】开发者回复评论，说明修复情况
lc pr comment 1 --body "## 🎬 [第六幕-C] 修复完成 ✅

@审查者 已根据 🎬 [第六幕-A] 的评论完成修复：

### 修复内容

- [x] **错误处理**：补充了数据库查询错误处理和返回
- [x] **日志记录**：增加结构化日志，记录关键操作
- [x] **代码注释**：添加函数说明和关键逻辑注释

### 代码变更

查看最新提交记录，请再次审查！"
```

---

**⚡️ 6.6 【第六幕-D】审查者确认合并（第二轮审查）**

```bash
# 【真实操作】审查者确认修复，批准合并
lc pr comment 1 --body "## 🎬 [第六幕-D] 审查通过 ✅

已确认 🎬 [第六幕-C] 的修复内容：

- ✅ 错误处理完善
- ✅ 日志记录充分
- ✅ 代码注释清晰

**修复确认，批准合并！** 👍"

# 【真实操作】执行合并（使用 squash 方式，删除源分支）
# --squash: 使用 squash 合并，保持主分支历史简洁
# --delete-branch: 合并后自动删除源分支（feature/xxx），保持仓库整洁
lc pr merge 1 \
  --squash \
  --delete-branch \
  --body "🎬 [第六幕-D] feat(inventory): 实现智能库存预警系统 Phase 1

完整功能包括库存监控、多级预警、多渠道通知。

演示完整的代码审查闭环：
- 🎬 [第六幕-A] 发现问题
- 🎬 [第六幕-B] AI修复代码
- 🎬 [第六幕-C] 回复修复完成
- 🎬 [第六幕-D] 审查确认合并

关联需求: XXJSxiaobaice-1203"
```

---

### 🎬 第七幕：更新需求状态（2分钟）

#### 场景
代码合并后，更新灵畿平台需求状态为已完成。

#### 演示动作

**⚡️ 7.1 查看合并结果**
```bash
# 【真实操作】查看合并后的提交状态
git checkout master
git pull
git log --oneline -3
```

---

**⚡️ 7.2 更新需求状态**
```bash
# 【真实操作】更新需求为已完成
lc req update XXJSxiaobaice-1203 \
  --workspace-key XXJSxiaobaice \
  --name "🎬 [第七幕-完成] 智能库存预警系统 - Phase 1" \
  --requirement "【演示完成】完整的代码审查闭环演示成功！

演示流程：
✅ [第一幕] 仓库创建
✅ [第二幕] 需求创建
✅ [第三幕] 任务分解
✅ [第四幕] AI开发
✅ [第五幕] PR创建
✅ [第六幕-A] 审查发现问题
✅ [第六幕-B] AI修复代码
✅ [第六幕-C] 回复修复完成
✅ [第六幕-D] 审查确认合并
✅ [第七幕] 需求完成

核心价值: 先建仓库 → 自动探测 → 完整闭环！"
```

---

### 🎬 第八幕：演示总结

#### 完整流程回顾

| 幕次 | 阶段 | 关键操作 | 说明 |
|------|------|----------|------|
| 第一幕 | 创建仓库 | `lc repo create` → `lc repo disable-work-item-link` | 先建仓库，关闭提交关联限制 |
| 第二幕 | 需求创建 | `lc req create` | 带 🎬 [第二幕] 标识 |
| 第三幕 | 任务分解 | `lc task create` ×3 | 带 🎬 [第三幕] 标识 |
| 第四幕 | AI开发 | `git checkout -b` → 代码 → `git push` | 特性分支开发 |
| 第五幕 | PR创建 | `lc pr create` | 自动探测生效 |
| 第六幕-A | 发现问题 | `lc pr comment` | 审查者指出问题 |
| 第六幕-B | AI修复 | 读取评论 → 修复 → `git commit` → `git push` | AI根据评论修复 |
| 第六幕-C | 回复修复 | `lc pr comment` | 开发者回复修复完成 |
| 第六幕-D | 确认合并 | `lc pr comment` → `lc pr merge --delete-branch` | 审查确认，合并并删除分支 |
| 第七幕 | 更新状态 | `lc req update` | 标记为完成 |

#### 自动探测的价值

通过**先创建仓库**的调整，演示中大量命令简化为：

| 场景 | 调整前（需手动指定） | 调整后（自动探测） |
|------|---------------------|-------------------|
| 创建 PR | `--git-project-id 45138 --workspace-key XXJSxiaobaice --source feature/xxx` | **无需这些参数** |
| 查看 PR 列表 | `--git-project-id 45138 --workspace-key XXJSxiaobaice` | **无需参数** |
| PR 评论 | `--git-project-id 45138 --workspace-key XXJSxiaobaice` | **仅需 `--body`** |
| PR 合并 | `--git-project-id 45138 --workspace-key XXJSxiaobaice` | **仅需 `--body`** |

---

## 五、完整演示脚本（可直接执行）

```bash
#!/bin/bash
# 🎬 灵畿 × Claude Code 全场景演示脚本（真实操作版）
# 注意：所有命令均为真实操作，请确保已登录并有相应权限

set -e  # 出错时停止

echo "================================"
echo "  灵畿 × Claude Code 演示脚本  "
echo "================================"

# ====== 第一幕：创建代码仓库 ======
echo ""
echo "【第一幕】创建代码仓库"
echo "================================"

# 查看个人代码组
echo "→ 查看个人代码组..."
lc repo group personal

# 创建仓库（请替换 group-id）
echo "→ 创建演示仓库..."
DEMO_REPO="inventory-demo-$(date +%m%d-%H%M%S)"
lc repo create "$DEMO_REPO" \
  --group-id 617927 \
  --workspace-key XXJSxiaobaice

# 记录返回的 gitProjectId
echo ""
echo "⚠️ 请记录上面返回的 gitProjectId 和 tenantHttpPath"
echo "按回车继续..."
read

# 关闭提交关联工作项功能（重要！否则 git push 会失败）
echo "→ 关闭'提交关联工作项'功能..."
echo "请输入 gitProjectId:"
read GIT_PROJECT_ID
lc repo disable-work-item-link \
  --git-project-id $GIT_PROJECT_ID \
  --workspace-key XXJSxiaobaice

# 克隆并初始化仓库
echo "→ 克隆仓库..."
cd /tmp
git clone "http://code-xxjs.rdcloud.4c.hq.cmcc/osc/XXJS/weibaohui-hq.cmcc/${DEMO_REPO}.git"
cd "$DEMO_REPO"

echo "→ 初始化仓库..."
echo "# 智能库存预警系统" > README.md
git add README.md
git commit -m "chore: 初始化仓库"
git push -u origin master

# 验证自动探测
echo "→ 验证自动探测..."
lc detect

echo "按回车继续..."
read

# ====== 第二幕：创建需求 ======
echo ""
echo "【第二幕】创建需求"
echo "================================"

lc req create "智能库存预警系统 - Phase 1" \
  --workspace-key XXJSxiaobaice \
  --project-code R24113J3C04

echo ""
echo "⚠️ 请记录返回的 objectId"
echo "按回车继续..."
read

# ====== 第三幕：创建任务 ======
echo ""
echo "【第三幕】创建任务"
echo "================================"

echo "请输入需求 objectId:"
read REQ_ID

lc task create $REQ_ID "设计库存预警数据库表结构" \
  --workspace-key XXJSxiaobaice \
  --project-code R24113J3C04

lc task create $REQ_ID "开发库存监控 API" \
  --workspace-key XXJSxiaobaice \
  --project-code R24113J3C04

lc task create $REQ_ID "实现多渠道通知服务" \
  --workspace-key XXJSxiaobaice \
  --project-code R24113J3C04

echo "按回车继续..."
read

# ====== 第四幕：AI 开发 ======
echo ""
echo "【第四幕】AI 开发（Claude Code）"
echo "================================"
echo ""
echo "💡 请与 Claude 交互生成代码..."
echo "   - 数据库模型（internal/models/inventory.go）"
echo "   - 业务服务（internal/service/inventory_service.go）"
echo "   - 通知服务（internal/notification/notifier.go）"
echo ""
echo "按回车继续（假设代码已生成）..."
read

# 创建特性分支并提交
echo "→ 创建特性分支..."
cd /tmp/$DEMO_REPO
git checkout -b feature/inventory-alert-system

echo "→ 添加并提交代码..."
git add .
git commit -m "feat(inventory): 实现智能库存预警系统 Phase 1

- 添加库存数据模型
- 实现库存检查与预警服务
- 集成多渠道通知

关联需求: $REQ_ID"

git push -u origin feature/inventory-alert-system

echo "按回车继续..."
read

# ====== 第五幕：创建 PR ======
echo ""
echo "【第五幕】创建 PR（自动探测生效）"
echo "================================"

lc pr create \
  --title "feat(inventory): 实现智能库存预警系统 Phase 1" \
  --body "## 变更说明

本次 MR 实现了智能库存预警系统的核心功能。

### 🔗 关联信息
- 需求: $REQ_ID" \
  --target master \
  --remove-source

echo ""
echo "⚠️ 请记录返回的 MR iid"
echo "按回车继续..."
read

# ====== 第六幕：代码审查 ======
echo ""
echo "【第六幕】代码审查"
echo "================================"

echo "请输入 MR iid:"
read MR_ID

echo "→ 添加审查评论..."
lc pr comment $MR_ID --body "代码审查通过，质量优秀！"

echo "→ 批准 MR..."
lc pr review $MR_ID --type approve --body "LGTM!"

echo "→ 合并 MR（自动删除源分支）..."
# --squash: squash 合并，保持主分支历史整洁
# --delete-branch: 合并后自动删除 feature 分支，保持仓库整洁
lc pr merge $MR_ID \
  --squash \
  --delete-branch \
  --body "feat: 智能库存预警系统 Phase 1"

echo "按回车继续..."
read

# ====== 第七幕：更新需求 ======
echo ""
echo "【第七幕】更新需求状态"
echo "================================"

lc req update $REQ_ID \
  --workspace-key XXJSxiaobaice \
  --name "【已完成】智能库存预警系统 - Phase 1" \
  --requirement "开发完成，代码已合并至 master 分支。"

echo ""
echo "================================"
echo "  🎉 演示完成！"
echo "================================"
```

---

## 六、关键要点

### 6.1 操作类型说明

| 类型 | 说明 | 示例 |
|------|------|------|
| ⚡️ **真实操作** | 实际执行的 lc/git 命令 | `lc repo create`、`git commit` |
| 💬 **伪代码示意** | 仅展示逻辑，不实际执行 | `func CheckAndAlert(...) error` |
| 📺 **展示说明** | 向观众解释的内容 | 讲解自动探测原理 |

### 6.2 仓库地址

- 使用 `tenantHttpPath`：`http://code-xxjs.rdcloud.4c.hq.cmcc/...`（hq.cmcc 域名，外网可访问）
- 避免使用 `httpPath`：`code-repo.rdcloud.indn`（内网域名）

### 6.3 自动探测触发条件

在 Git 仓库目录下执行 lc 命令时，自动探测：
- `workspaceKey` → 研发空间 Key
- `gitProjectId` → Git 项目 ID
- `sourceBranch` → 当前 Git 分支

---

## 七、核心价值

> **"先建仓库，后提需求，自动探测贯穿全程"**

这个演示方案展示了一个**更顺畅的 AI 辅助研发闭环**：

1. **真实操作**：所有 lc 和 git 命令均为真实执行
2. **伪代码示意**：代码编写部分用伪代码节省演示时间
3. **自动探测**：大幅减少命令参数，演示更流畅
4. **完整闭环**：从仓库创建到代码合并，全程可追溯
