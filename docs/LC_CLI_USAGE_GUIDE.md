# LC 命令行工具使用指南

LC 是一款灵畿平台命令行工具，用于管理研发空间中的需求、任务、缺陷、代码仓库、合并请求和文档库等资源。

---

## 目录

1. [快速开始](#快速开始)
2. [研发空间自动探测](#研发空间自动探测)
3. [需求管理](#需求管理)
4. [任务管理](#任务管理)
5. [缺陷管理](#缺陷管理)
6. [代码仓库管理](#代码仓库管理)
7. [合并请求（PR）管理](#合并请求pr管理)
8. [文档库管理](#文档库管理)
9. [常用全局选项](#常用全局选项)
10. [典型工作流示例](#典型工作流示例)

---

## 快速开始

### 安装与登录

```bash
# 安装 lc 工具
sudo make install

# 查看版本
lc version

# 登录（获取访问令牌）
lc login
```

登录信息存储在 `~/.lc/config.json` 中。

### 查看帮助

```bash
# 查看全局帮助
lc --help

# 查看子命令帮助
lc req --help
lc task --help
lc bug --help
lc repo --help
lc pr --help
lc lib --help
```

---

## 研发空间自动探测

### 探测当前 Git 仓库关联的研发空间

**场景**：在 Git 仓库目录下工作，不想手动指定研发空间参数。

```bash
# 自动探测当前目录 Git 仓库关联的研发空间
lc detect

# 指定路径探测
lc detect --path /path/to/git/repo
```

**输出示例**：
```json
{
  "success": true,
  "data": {
    "matched": true,
    "workspaceKey": "XXJSLJCLIDEV",
    "workspaceName": "灵畿cli研发",
    "tenantId": "xxx",
    "repository": {
      "gitProjectId": 12345,
      "name": "my-project",
      "spaceCode": "XXJSLJCLIDEV"
    },
    "gitInfo": {
      "IsGitRepo": true,
      "RepoName": "my-project"
    },
    "spaceDetails": {
      "spaceName": "灵畿cli研发",
      "spaceDesc": "灵畿cli研发空间"
    }
  }
}
```

**使用场景**：
- 脚本自动化：在 CI/CD 流程中自动识别当前仓库所属研发空间
- 简化命令：在 Git 仓库目录下执行命令时，可省略 `--workspace-key` 参数

---

## 需求管理

### 1. 创建需求

#### 简单创建

**场景**：快速创建一个仅包含标题的简单需求。

```bash
lc req create "需求名称" \
  --workspace-key XXJSLJCLIDEV \
  --project-code XXJSLJCLIDEV
```

#### 通过 YAML 文件创建（推荐）

**场景**：创建包含完整字段的详细需求。

创建 `requirement.yaml` 文件：
```yaml
name: 详细需求名称
proposer:
  label: "张三(zhangsan@example.com)"
  value: zhangsan@example.com
  username: zhangsan@example.com
  nickname: 张三
assignee:
  label: "李四(lisi@example.com)"
  value: lisi@example.com
  username: lisi@example.com
  nickname: 李四
businessBackground: |
  业务背景描述
  支持多行文本
requirement: |
  需求描述详情：
  1. 功能点一
  2. 功能点二
acceptanceCriteria: |
  验收标准：
  1. 标准一
  2. 标准二
requirementType:
  - 开发域
```

执行命令：
```bash
lc req create -f requirement.yaml \
  --workspace-key XXJSLJCLIDEV \
  --project-code XXJSLJCLIDEV
```

#### 通过管道输入创建

**场景**：从其他命令或脚本动态生成需求内容。

```bash
cat <<EOF | lc req create -k \
  --workspace-key XXJSLJCLIDEV \
  --project-code XXJSLJCLIDEV
name: 管道输入需求
businessBackground: 通过管道创建
requirement: 需求描述
EOF
```

#### 最小配置创建

**场景**：仅需提供最必要的字段。

```yaml
name: 最小配置需求名称
```

### 2. 查询需求列表

**场景**：查看研发空间中的需求列表，支持分页。

```bash
# 查询前10个需求
lc req list \
  --workspace-key XXJSLJCLIDEV \
  --limit 10

# 分页查询（跳过前20个）
lc req list \
  --workspace-key XXJSLJCLIDEV \
  --limit 10 \
  --offset 20
```

### 3. 搜索需求

**场景**：根据关键词搜索需求。

```bash
lc req search "关键词" \
  --workspace-key XXJSLJCLIDEV \
  --limit 10
```

### 4. 查看需求详情

**场景**：获取需求的完整信息。

```bash
lc req view <需求objectId> \
  --workspace-key XXJSLJCLIDEV
```

### 5. 更新需求

**场景**：修改需求的字段。

```bash
# 更新名称、需求描述、验收标准
lc req update <需求objectId> \
  --workspace-key XXJSLJCLIDEV \
  --name "新名称" \
  --requirement "新的需求描述" \
  --acceptance-criteria "新的验收标准"

# 更新计划完成时间（时间戳，毫秒）
lc req update <需求objectId> \
  --workspace-key XXJSLJCLIDEV \
  --planned-end-time 1704067200000
```

### 6. 删除需求

**场景**：删除不再需要的需求。

```bash
lc req delete <需求objectId> \
  --workspace-key XXJSLJCLIDEV
```

---

## 任务管理

### 1. 创建任务

**场景**：为需求创建开发任务。

#### 简单创建

```bash
lc task create <需求objectId> "任务名称" \
  --workspace-key XXJSLJCLIDEV \
  --project-code XXJSLJCLIDEV
```

#### 通过 YAML 文件创建

创建 `task.yaml` 文件：
```yaml
name: 任务名称
requirementId: <需求objectId>
taskType:
  - 开发
taskDescription: |
  任务描述
  支持多行文本
plannedWorkingHours: 8
assignee:
  label: "张三(zhangsan@example.com)"
  value: zhangsan@example.com
  username: zhangsan@example.com
  nickname: 张三
priority: 8f7912a5-9176-4a79-a269-2269ac42b5a2
```

执行命令：
```bash
lc task create -f task.yaml \
  --workspace-key XXJSLJCLIDEV \
  --project-code XXJSLJCLIDEV
```

### 2. 查询任务列表

**场景**：查看任务列表，支持按需求过滤。

```bash
# 查询所有任务
lc task list \
  --workspace-key XXJSLJCLIDEV \
  --limit 10

# 按需求过滤任务
lc task list \
  --workspace-key XXJSLJCLIDEV \
  --requirement-id <需求objectId> \
  --limit 10

# 分页查询
lc task list \
  --workspace-key XXJSLJCLIDEV \
  --limit 5 \
  --offset 0
```

### 3. 搜索任务

**场景**：根据关键词搜索任务。

```bash
lc task search "关键词" \
  --workspace-key XXJSLJCLIDEV \
  --limit 10
```

### 4. 删除任务

```bash
lc task delete <任务objectId> \
  --workspace-key XXJSLJCLIDEV
```

---

## 缺陷管理

### 1. 创建缺陷

**场景**：记录软件缺陷。

```bash
lc bug create \
  --title "缺陷标题" \
  --description "缺陷描述" \
  --project-id <项目ID> \
  --workspace-key XXJSLJCLIDEV \
  --template-simple
```

### 2. 查询缺陷列表

**场景**：查看缺陷列表。

```bash
lc bug list \
  --workspace-key XXJSLJCLIDEV \
  --page 1 \
  --limit 10
```

### 3. 查看缺陷详情

```bash
lc bug view <缺陷ID> \
  --workspace-key XXJSLJCLIDEV
```

### 4. 获取缺陷状态列表

**场景**：查看可用于缺陷流转的状态。

```bash
lc bug status \
  --workspace-key XXJSLJCLIDEV
```

### 5. 更新缺陷状态

**场景**：将缺陷状态从"待修复"改为"待验证"。

```bash
lc bug update-status <缺陷ID> <状态ID> \
  --workspace-key XXJSLJCLIDEV
```

### 6. 删除缺陷

```bash
lc bug delete <缺陷ID> \
  --workspace-key XXJSLJCLIDEV
```

---

## 代码仓库管理

### 1. 创建仓库

**场景**：在指定代码组下创建新仓库。

```bash
lc repo create "仓库名称" \
  --workspace-key XXJSLJCLIDEV \
  --group-id <代码组ID>
```

**输出**：返回仓库信息，包括 `gitProjectId`、`httpPath`（克隆地址）等。

### 2. 查询仓库列表

**场景**：查看研发空间下的仓库。

```bash
lc repo list \
  --workspace-key XXJSLJCLIDEV \
  --limit 10
```

### 3. 搜索仓库

```bash
lc repo search "关键词" \
  --workspace-key XXJSLJCLIDEV \
  --limit 10
```

### 4. 禁用工作项关联

**场景**：关闭仓库与卡片（需求/任务/缺陷）的自动关联功能。

```bash
lc repo disable-work-item-link <gitProjectId> \
  --workspace-key XXJSLJCLIDEV
```

### 5. 删除仓库

**⚠️ 警告：此操作不可恢复**

```bash
lc repo delete <gitProjectId> \
  --workspace-key XXJSLJCLIDEV
```

---

## 合并请求（PR）管理

### 1. 查询 PR 列表

**场景**：查看仓库的合并请求。

```bash
lc pr list \
  --workspace-key XXJSLJCLIDEV \
  --limit 10
```

### 2. 创建 PR

**场景**：将功能分支合并到主分支。

```bash
lc pr create \
  --title "PR 标题" \
  --body "PR 描述" \
  --source feature-branch \
  --target main \
  --git-project-id <gitProjectId> \
  --workspace-key XXJSLJCLIDEV
```

### 3. 添加 PR 评论

**场景**：在代码审查中添加评论。

```bash
lc pr comment <PR编号> \
  --git-project-id <gitProjectId> \
  --workspace-key XXJSLJCLIDEV \
  --body "评论内容"
```

### 4. 查看 PR 详情（含评论）

**场景**：查看 PR 信息和评论列表。

```bash
lc pr view <PR编号> \
  --git-project-id <gitProjectId> \
  --workspace-key XXJSLJCLIDEV \
  --comments
```

### 5. 解决评论

**场景**：将评论标记为已解决。

```bash
lc pr patch-comment <PR编号> \
  --git-project-id <gitProjectId> \
  --workspace-key XXJSLJCLIDEV \
  --comment-id <评论ID> \
  --state fixed
```

### 6. 合并 PR

**场景**：将 PR 合并到目标分支。

```bash
lc pr merge <PR编号> \
  --git-project-id <gitProjectId> \
  --workspace-key XXJSLJCLIDEV \
  --type merge
```

合并类型：`merge`（合并提交）、`squash`（压缩提交）、`rebase`（变基）。

---

## 文档库管理

### 1. 创建文档库

**场景**：创建知识库或文档存储空间。

```bash
lc lib create "文档库名称" \
  --workspace-key XXJSLJCLIDEV
```

### 2. 查询文档库列表

```bash
lc lib list \
  --workspace-key XXJSLJCLIDEV
```

### 3. 创建文件夹

**场景**：在文档库中创建文件夹。

```bash
lc lib folder create "文件夹名称" \
  --prt-id <文档库externalLibId>
```

### 4. 查询文件夹内容

```bash
lc lib folder list \
  --prt-id <文件夹ID或文档库ID>
```

### 5. 上传文件

**场景**：将本地文件上传到文档库。

```bash
lc lib upload /path/to/local/file.txt \
  --folder-id <文件夹ID> \
  --name "上传后的文件名.txt"
```

### 6. 删除文件/文件夹

```bash
# 删除文件
lc lib file delete <文件ID> \
  --folder-id <父文件夹ID>

# 删除文件夹
lc lib file delete <文件夹ID> \
  --folder-id <父文件夹ID或文档库ID>
```

### 7. 删除文档库

```bash
lc lib delete <externalLibId>
```

---

## 常用全局选项

| 选项 | 简写 | 说明 |
|------|------|------|
| `--workspace-key` | `-w` | 研发空间 Key（如：XXJSLJCLIDEV） |
| `--project-code` | | 项目代码 |
| `--limit` | `-l` | 返回结果数量限制 |
| `--offset` | `-o` | 分页偏移量 |
| `--debug` | `-d` | 启用调试输出 |
| `--insecure` | `-k` | 禁用 TLS 证书验证 |
| `--pretty` | - | 输出格式化 JSON |
| `--cookie` | `-c` | 直接传入认证 Cookie，覆盖本地配置 |

### 自动探测模式

在 Git 仓库目录下执行命令时，可以省略 `--workspace-key` 参数，LC 会自动探测关联的研发空间：

```bash
# 进入 Git 仓库目录
cd /path/to/git/repo

# 自动探测研发空间并执行命令
lc req list -k -l 10
```

---

## 典型工作流示例

### 工作流一：需求驱动开发

```bash
# 1. 创建需求
lc req create "用户登录功能" \
  -k --workspace-key XXJSLJCLIDEV \
  --project-code XXJSLJCLIDEV
# 记录返回的 objectId: req_xxx

# 2. 为需求创建开发任务
lc task create req_xxx "实现登录API" \
  -k --workspace-key XXJSLJCLIDEV \
  --project-code XXJSLJCLIDEV

# 3. 创建代码仓库
lc repo create "login-service" \
  -k --workspace-key XXJSLJCLIDEV \
  --group-id 617927
# 记录返回的 gitProjectId

# 4. 克隆仓库并开始开发
git clone <仓库地址>

# 5. 开发完成后创建 PR
lc pr create \
  -t "实现用户登录功能" \
  -b "关联需求: req_xxx" \
  -s feature-login \
  -m main \
  --git-project-id 12345 \
  -k --workspace-key XXJSLJCLIDEV

# 6. 合并 PR
lc pr merge 1 \
  --git-project-id 12345 \
  -k --workspace-key XXJSLJCLIDEV \
  --type merge

# 7. 更新任务状态（通过其他方式）
# 8. 关闭需求（通过其他方式）
```

### 工作流二：缺陷跟踪

```bash
# 1. 记录缺陷
lc bug create \
  -t "登录按钮无响应" \
  -D "点击登录按钮后页面无任何反应" \
  -p R24113J3C04 \
  -k --workspace-key XXJSLJCLIDEV \
  --template-simple

# 2. 查看缺陷列表
lc bug list -k --workspace-key XXJSLJCLIDEV

# 3. 获取可流转的状态
lc bug status -k --workspace-key XXJSLJCLIDEV

# 4. 修复后更新状态为"待验证"
lc bug update-status <缺陷ID> <待验证状态ID> \
  -k --workspace-key XXJSLJCLIDEV

# 5. 验证通过后关闭缺陷
lc bug update-status <缺陷ID> <已关闭状态ID> \
  -k --workspace-key XXJSLJCLIDEV
```

### 工作流三：文档管理

```bash
# 1. 创建文档库
lc lib create "产品文档" \
  -k --workspace-key XXJSLJCLIDEV

# 2. 查看文档库 ID
lc lib list -k --workspace-key XXJSLJCLIDEV

# 3. 创建文件夹
lc lib folder create "API文档" \
  --prt-id <externalLibId>

# 4. 上传文档
lc lib upload ./api-guide.pdf \
  --folder-id <文件夹ID> \
  --name "API使用指南.pdf"

# 5. 查看文件夹内容
lc lib folder list --prt-id <文件夹ID>
```

### 工作流四：批量操作脚本

```bash
#!/bin/bash
# 批量创建需求的脚本示例

WORKSPACE_KEY="XXJSLJCLIDEV"
PROJECT_CODE="XXJSLJCLIDEV"

requirements=("需求1" "需求2" "需求3")

for req in "${requirements[@]}"; do
    echo "创建需求: $req"
    lc req create "$req" \
        -k --workspace-key $WORKSPACE_KEY \
        --project-code $PROJECT_CODE
done
```

---

## 注意事项

1. **YAML 文件格式**：创建需求/任务时使用 YAML 格式，注意字段缩进和类型。
2. **时间戳格式**：更新时间字段时使用 Unix 毫秒时间戳。
3. **ID 类型区分**：
   - 需求/任务使用 `objectId`（字符串）
   - 缺陷使用 `id`（字符串）
   - 仓库使用 `gitProjectId`（数字）
   - 文档库使用 `externalLibId`（数字）
4. **自动探测限制**：必须在 Git 仓库目录下且远程仓库已关联研发空间才能使用自动探测。
5. **错误处理**：所有命令返回 JSON 格式结果，通过 `success` 字段判断执行状态。

---

## 更多资源

- 查看完整命令帮助：`lc <命令> --help`
- E2E 测试示例：`e2e/integration/` 目录下的测试文件
