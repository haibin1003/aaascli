# lc 命令参数与 detect 输出对照表

## 各命令必传参数列表

### 1. 代码仓库相关 (repo)

| 命令 | 必传参数 | 从 detect 获取 | 说明 |
|------|---------|---------------|------|
| `repo create` | `--workspace-key`, `--group-id` | ⚠️ partial | workspaceKey ✓, group-id ✗ |
| `repo disable-work-item-link` | `--workspace-key` | ✓ | workspaceKey |
| `repo delete` | `--workspace-key` | ✓ | workspaceKey |
| `repo list` | `--workspace-key` | ✓ | workspaceKey |
| `repo search` | 无 | ✓ | 全局搜索，无需参数 |
| `repo group list` | 无 | ✗ | 全局查询 |
| `repo group add` | 无 | ✗ | 仅需 name 参数 |
| `repo group personal` | 无 | ✗ | 查询个人代码组 |

### 2. PR/MR 相关 (pr)

| 命令 | 必传参数 | 从 detect 获取 | 说明 |
|------|---------|---------------|------|
| `pr create` | `--workspace-key`, `--git-project-id`, `--source`, `--title` | ⚠️ partial | workspaceKey ✓, git-project-id ✓, 分支/标题 ✗ |
| `pr review` | `--workspace-key`, `--git-project-id` | ✓ | workspaceKey, gitProjectId |
| `pr merge` | `--workspace-key`, `--git-project-id` | ✓ | workspaceKey, gitProjectId |
| `pr comment` | `--workspace-key`, `--git-project-id` | ✓ | workspaceKey, gitProjectId |
| `pr view` | `--workspace-key`, `--git-project-id` | ✓ | workspaceKey, gitProjectId |
| `pr list` | `--workspace-key`, `--git-project-id` | ✓ | workspaceKey, gitProjectId |
| `pr patch-comment` | `--workspace-key`, `--git-project-id`, `--comment-id`, `--state` | ⚠️ partial | workspaceKey, gitProjectId ✓, 其他 ✗ |

### 3. 需求相关 (req)

| 命令 | 必传参数 | 从 detect 获取 | 说明 |
|------|---------|---------------|------|
| `req create` | `--workspace-key`, `--project-code` | ⚠️ partial | workspaceKey ✓, project-code ✗ |
| `req list` | `--workspace-key` | ✓ | workspaceKey, workspaceName（自动获取） |
| `req view` | `--workspace-key` | ✓ | workspaceKey |
| `req delete` | `--workspace-key` | ✓ | workspaceKey |
| `req update` | `--workspace-key` | ✓ | workspaceKey |
| `req search` | `--workspace-key` | ✓ | workspaceKey, workspaceName（自动获取） |

### 4. 任务相关 (task)

| 命令 | 必传参数 | 从 detect 获取 | 说明 |
|------|---------|---------------|------|
| `task create` | `--workspace-key`, `--project-code` | ⚠️ partial | workspaceKey ✓, project-code ✗ |
| `task list` | `--workspace-key` | ✓ | workspaceKey, workspaceName（自动获取） |
| `task delete` | `--workspace-key` | ✓ | workspaceKey |
| `task search` | `--workspace-key` | ✓ | workspaceKey, workspaceName（自动获取） |

### 5. 缺陷相关 (bug)

| 命令 | 必传参数 | 从 detect 获取 | 说明 |
|------|---------|---------------|------|
| `bug create` | `--workspace-key`, `--project-id`, `--title`, `--description` | ⚠️ partial | workspaceKey ✓, 其他 ✗ |
| `bug list` | `--workspace-key` | ✓ | workspaceKey |
| `bug view` | `--workspace-key` | ✓ | workspaceKey |
| `bug status` | `--workspace-key` | ✓ | workspaceKey |
| `bug update-status` | `--workspace-key` | ✓ | workspaceKey |
| `bug delete` | `--workspace-key` | ✓ | workspaceKey |

### 6. 文档库相关 (library)

| 命令 | 必传参数 | 从 detect 获取 | 说明 |
|------|---------|---------------|------|
| `lib list` | `--workspace-key` | ✓ | workspaceKey |
| `lib create` | `--workspace-key` | ✓ | workspaceKey |
| `lib delete` | 无 | - | 仅需 name 参数 |
| `lib folder create` | `--prt-id` | ✗ | 父文件夹 ID |
| `lib folder tree` | `--prt-id` | ✗ | 父文件夹 ID |
| `lib folder list` | `--prt-id` | ✗ | 父文件夹 ID |
| `lib file upload` | `--folder-id` | ✗ | 文件夹 ID |
| `lib file delete` | `--folder-id` | ✗ | 文件夹 ID |

### 7. 研发空间相关 (space)

| 命令 | 必传参数 | 从 detect 获取 | 说明 |
|------|---------|---------------|------|
| `space list` | 无 | ✗ | 全局查询 |
| `space project list` | 无 | ✗ | 从 space list 中选择 |

## 参数获取总结

### ✅ 可完全从 detect 获取的参数

| 参数 | detect 输出路径 | 说明 |
|------|----------------|------|
| `workspace-key` | `.data.workspaceKey` | 研发空间 Key |
| `workspace-name` | `.data.workspaceName` | 研发空间名称 |
| `git-project-id` | `.data.repository.gitProjectId` | Git 项目 ID |
| `tenant-id` | `.data.tenantId` | 租户 ID |

### ⚠️ 部分可获取的参数

| 参数 | detect 输出 | 说明 |
|------|------------|------|
| `spaceCode` | `.data.repository.spaceCode` | 与 workspace-key 一致 |
| `codeGroupId` | `.data.repository.codeGroupId` | 代码组 ID |
| `codeGroupName` | `.data.repository.codeGroupName` | 代码组名称 |

### ❌ 无法从 detect 获取的参数

| 参数 | 说明 | 获取方式 |
|------|------|---------|
| `project-code` | 项目代码 | `lc space project list` |
| `project-id` | 项目 ID | `lc space project list` |
| `group-id` | 代码组 ID | `lc repo group personal` |
| `prt-id` | 父文件夹 ID | `lc lib folder list` |
| `folder-id` | 文件夹 ID | `lc lib folder list` |
| `comment-id` | 评论 ID | `lc pr view` |
| `source` | 源分支 | Git 本地获取 |
| `target` | 目标分支 | 通常默认 master/main |
| `title` | 标题 | 用户输入 |
| `description` | 描述 | 用户输入 |

## 自动化脚本示例

### 获取 workspace-key 执行命令
```bash
# 基础用法
WORKSPACE=$(lc detect -k | jq -r '.data.workspaceKey')
lc req list -w $WORKSPACE

# 完整用法（包含 workspace-name，仍然有效）
DETECT=$(lc detect -k)
WORKSPACE=$(echo $DETECT | jq -r '.data.workspaceKey')
WORKSPACE_NAME=$(echo $DETECT | jq -r '.data.workspaceName')
GIT_PROJECT_ID=$(echo $DETECT | jq -r '.data.repository.gitProjectId')

# 列需求（workspace-name 已自动获取，可省略）
lc req list -w $WORKSPACE

# 列 PR
lc pr list --git-project-id $GIT_PROJECT_ID -w $WORKSPACE
```

### 创建 PR 的完整流程
```bash
# 获取 detect 信息
DETECT=$(lc detect -k)
WORKSPACE=$(echo $DETECT | jq -r '.data.workspaceKey')
GIT_PROJECT_ID=$(echo $DETECT | jq -r '.data.repository.gitProjectId')

# 获取当前分支
SOURCE_BRANCH=$(git branch --show-current)

# 创建 PR（title 和 target 仍需手动指定）
lc pr create \
  -w $WORKSPACE \
  --git-project-id $GIT_PROJECT_ID \
  -s $SOURCE_BRANCH \
  --target "main" \
  --title "功能开发"
```

### 创建需求/任务的完整流程
```bash
# 获取 detect 信息
DETECT=$(lc detect -k)
WORKSPACE=$(echo $DETECT | jq -r '.data.workspaceKey')
WORKSPACE_NAME=$(echo $DETECT | jq -r '.data.workspaceName')

# 获取项目代码（需要额外查询）
PROJECT_CODE=$(lc space project list | jq -r '.data[0].projectCode')

# 创建需求
lc req create "新需求" -w $WORKSPACE --project-code $PROJECT_CODE

# 创建任务
lc task create "新任务" -w $WORKSPACE --project-code $PROJECT_CODE
```

## 改进建议

1. **增强 detect 输出**：考虑将 `projectCode` 也加入 detect 输出（如果在仓库信息中可获取）

2. **智能默认值**：对于 `pr create`，可以自动获取当前 Git 分支作为 `source-branch`

3. **缓存机制**：detect 结果可以本地缓存，避免频繁调用 API

4. **交互式选择**：当需要 `project-code` 等额外参数时，提供交互式选择界面
