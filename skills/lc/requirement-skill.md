---
name: lc-requirement
description: |
  管理需求(Requirement)和任务(Task)。当用户提到需求、任务、req、task、requirement、创建需求、查询任务、或需要对需求进行 CRUD 操作时触发。支持自动探测研发空间。
metadata:
  {
    "joinai-code":
      {
        "requires": { "bins": ["lc"] },
      },
  }
---

# 需求和任务管理

## 需求命令 (req)

### 查询需求列表
```bash
lc req list -w <workspace-key>
lc req list -w XXJSLJCLIDEV --pretty
```

### 搜索需求
```bash
lc req search <关键词> -w <workspace-key>
lc req search 用户登录 -w XXJSLJCLIDEV
```

### 查看需求详情
```bash
lc req view <requirement-id> -w <workspace-key>
lc req view nCKhZGcBKu -w XXJSLJCLIDEV
```

### 创建需求
```bash
lc req create "<需求名称>" -w <workspace-key>
lc req create "用户登录功能" -w XXJSLJCLIDEV
```

### 更新/删除需求
```bash
lc req update <requirement-id> -w <workspace-key>
lc req delete <requirement-id> -w <workspace-key>
```

## 任务命令 (task)

### 查询任务列表
```bash
lc task list -w <workspace-key>
# 筛选指定需求下的任务
lc task list -w <workspace-key> -r <requirement-id>
```

### 搜索任务
```bash
lc task search <关键词> -w <workspace-key>
```

### 创建任务

**简单方式**：
```bash
lc task create <requirement-id> "<任务名称>" -w <workspace-key>
lc task create nCKhZGcBKu "实现登录功能" -w XXJSLJCLIDEV
```

**YAML 文件方式**：
```bash
lc task create -f task.yaml -w <workspace-key>
cat task.yaml | lc task create -w <workspace-key>
```

**YAML 格式**：
```yaml
name: 任务名称
requirementId: nCKhZGcBKu
taskType:
  - 开发
taskDescription: |
  任务详细描述
plannedWorkingHours: 8
assignee:
  label: "张三(zhangsan@example.com)"
  value: zhangsan@example.com
```

### 删除任务
```bash
lc task delete <task-id> -w <workspace-key>
```

## 工作流程

```bash
# 1. 创建需求，获取需求 ID
lc req create "用户注册功能" -w XXJSLJCLIDEV
# 返回新需求的 ID，如 abc123

# 2. 在需求下创建任务
lc task create abc123 "设计数据库表" -w XXJSLJCLIDEV
lc task create abc123 "实现 API 接口" -w XXJSLJCLIDEV

# 3. 查看需求详情
lc req view abc123 -w XXJSLJCLIDEV

# 4. 查询需求下的任务
lc task list -w XXJSLJCLIDEV -r abc123
```

## 注意事项

- 需求 ID 是 ObjectId 类型，不是数字
- 删除操作需谨慎，无法恢复
- `lc req list` 和 `lc req search` 支持自动探测研发空间
