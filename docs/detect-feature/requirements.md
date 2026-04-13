# lc detect 功能需求文档

## 1. 需求概述

### 1.1 背景
在使用灵畿 CLI (lc) 时，用户需要频繁指定 `--workspace-key` 参数来标识研发空间。对于开发人员来说，当他们在本地 Git 仓库中工作时，往往已经处于某个项目的代码库中。如果能自动识别当前目录所属的灵畿研发空间，将大大提升使用体验。

### 1.2 目标
实现 `lc detect` 命令，自动探测当前 Git 仓库上下文，包括：
- 识别当前目录是否在 Git 仓库中
- 自动获取 Git 远程仓库信息
- 匹配灵畿平台上的对应仓库
- 输出研发空间 Key、租户 ID 等综合上下文信息

## 2. 功能需求

### 2.1 核心功能

| 需求编号 | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-001 | 检测当前目录是否为 Git 仓库 | P0 |
| REQ-002 | 获取 Git 远程仓库 URL (origin) | P0 |
| REQ-003 | 通过仓库名称搜索匹配的灵畿仓库 | P0 |
| REQ-004 | 根据 Git URL 精确匹配远程仓库 | P0 |
| REQ-005 | 输出研发空间 Key (workspaceKey) | P0 |
| REQ-006 | 输出研发空间名称 (workspaceName) | P0 |
| REQ-007 | 输出租户 ID (tenantId) | P0 |
| REQ-008 | 输出仓库详细信息 (repository) | P0 |
| REQ-009 | 支持指定路径探测 `--path` | P1 |
| REQ-010 | 支持调试模式 `-d` | P1 |

### 2.2 输出信息

```json
{
  "success": true,
  "data": {
    "workspaceKey": "XXJSSGKJCXPTYYTG",
    "workspaceName": "灵畿科研平台开发域",
    "tenantId": "ENTP750043923870622608",
    "repository": { ... },
    "spaceDetails": { ... },
    "gitInfo": { ... },
    "matched": true
  }
}
```

## 3. 非功能需求

### 3.1 性能要求
- 命令执行时间 < 3 秒（正常网络环境下）

### 3.2 兼容性要求
- 支持 HTTP 和 SSH 格式的 Git URL
- 支持带端口号的 Git URL (如 :8022)

### 3.3 错误处理
- 非 Git 仓库目录：返回 `matched: false` 及原因
- 未找到匹配的远程仓库：返回 `matched: false`
- API 调用失败：返回清晰的错误信息

## 4. 使用场景

### 4.1 场景一：脚本自动化
```bash
WORKSPACE=$(lc detect -k | jq -r '.data.workspaceKey')
lc req list -w $WORKSPACE
```

### 4.2 场景二：Shell 集成
在命令行提示符中显示当前研发空间名称。

### 4.3 场景三：CI/CD 集成
在自动化流程中自动识别项目上下文。

## 5. 验收标准

- [x] 在 Git 仓库目录下执行，能正确输出 workspaceKey
- [x] 在非 Git 目录执行，返回 matched: false
- [x] 支持 `--path` 参数指定其他目录
- [x] 输出信息包含完整的上下文字段
- [x] 帮助文档包含脚本自动化示例
