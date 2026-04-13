---
name: lc-defect
description: |
  管理测试中心缺陷(Bug)。当用户提到 bug、缺陷、defect、bug list、创建缺陷、缺陷管理、或需要操作缺陷时触发。支持自动探测研发空间。
metadata:
  {
    "joinai-code":
      {
        "requires": { "bins": ["lc"] },
      },
  }
---

# 缺陷管理 (Bug)

## 查询缺陷列表
```bash
lc bug list -w <workspace-key>
lc bug list -w XXJSLJCLIDEV --pretty
```

## 查看缺陷详情
```bash
lc bug view <bug-id> -w <workspace-key>
lc bug view 456 -w XXJSLJCLIDEV
```

## 创建缺陷
```bash
lc bug create -w <workspace-key> \
  --title "登录页面崩溃" \
  --severity high \
  --description "复现步骤：..."
```

## 更新缺陷状态
```bash
lc bug update-status <bug-id> -w <workspace-key>
```

## 列举可用状态
```bash
lc bug status -w <workspace-key>
```

## 删除缺陷
```bash
lc bug delete <bug-id> -w <workspace-key>
```

## 工作流程

```bash
# 1. 查询现有缺陷
lc bug list -w XXJSLJCLIDEV

# 2. 创建新缺陷
lc bug create -w XXJSLJCLIDEV \
  --title "用户无法登录" \
  --severity high \
  --description "复现步骤：..."

# 3. 查看缺陷详情
lc bug view <bug-id> -w XXJSLJCLIDEV

# 4. 修复后更新状态
lc bug update-status <bug-id> -w XXJSLJCLIDEV --status resolved

# 5. 删除已关闭的缺陷
lc bug delete <bug-id> -w XXJSLJCLIDEV
```

## 严重程度选项

- `low` - 轻微
- `medium` - 中等
- `high` - 严重
- `critical` - 危急

## 注意事项

- 缺陷 ID 可以是数字或 ObjectId
- `lc bug list` 支持自动探测研发空间
- 删除操作需谨慎
