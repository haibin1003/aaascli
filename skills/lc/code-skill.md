---
name: lc-code
description: |
  管理代码仓库(Repo)、合并请求(PR)、CI 构建。当用户提到 repo、repository、仓库、pr、merge request、mr、ci、build、构建、代码仓库、或需要操作 Git 仓库和 MR 时触发。
metadata:
  {
    "joinai-code":
      {
        "requires": { "bins": ["lc"] },
      },
  }
---

# 代码仓库、PR 和 CI 管理

## 仓库命令 (repo)

### 查询仓库列表
```bash
lc repo list -w <workspace-key>
lc repo list -w XXJSLJCLIDEV --pretty
```

### 搜索仓库
```bash
lc repo search <关键词> -w <workspace-key>
lc repo search myproject -w XXJSLJCLIDEV
```

### 创建仓库
```bash
lc repo create <仓库名称> -w <workspace-key>
```

### 删除仓库
```bash
lc repo delete <repo-id> -w <workspace-key>
```

### 管理仓库组
```bash
lc repo group list -w <workspace-key>
lc repo group create <组名称> -w <workspace-key>
```

### 关闭代码关联工作项
```bash
lc repo disable-work-item-link <repo-id> -w <workspace-key>
```

## 合并请求命令 (pr)

### 列出合并请求
```bash
lc pr list -w <workspace-key>
```

### 查看合并请求详情
```bash
lc pr view <mr-id> -w <workspace-key>
lc pr view 123 -w XXJSLJCLIDEV
```

### 创建合并请求
```bash
lc pr create -w <workspace-key> \
  --title "feat: 添加用户登录功能" \
  --source-branch feat/login \
  --target-branch main \
  --repo-id 123
```

### 审核合并请求
```bash
lc pr review <mr-id> -w <workspace-key>
```

### 评论合并请求
```bash
lc pr comment <mr-id> -w <workspace-key>
```

### 合并合并请求
```bash
lc pr merge <mr-id> -w <workspace-key>
```

## CI 命令 (ci)

### 查询构建任务列表
```bash
lc ci list -w <workspace-key>
# 按状态筛选：1-待处理 2-运行中 3-成功 4-失败
lc ci list -w XXJSLJCLIDEV -s 4
```

### 查询构建历史
```bash
lc ci history -w <workspace-key> -t <task-id>
lc ci history -w XXJSLJCLIDEV -t 205338bc206c4a05a6d6a72f88ab5aa0
```

## 工作流程

```bash
# 1. 创建/查询仓库
lc repo create myproject -w XXJSLJCLIDEV
lc repo list -w XXJSLJCLIDEV

# 2. 创建分支并开发
git checkout -b feat/login
# ... 开发代码 ...
git push origin feat/login

# 3. 创建 MR
lc pr create -w XXJSLJCLIDEV --title "feat: 登录功能" \
  --source-branch feat/login --target-branch main --repo-id <id>

# 4. 审核 MR
lc pr review <mr-id> -w XXJSLJCLIDEV --approve

# 5. 合并 MR
lc pr merge <mr-id> -w XXJSLJCLIDEV

# 6. 检查 CI 状态
lc ci list -w XXJSLJCLIDEV -s 4  # 查看失败构建
```

## 注意事项

- MR ID 可以是数字或 ObjectId
- CI 状态码：1-待处理 2-运行中 3-成功 4-失败
