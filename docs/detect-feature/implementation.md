# lc detect 功能实现总结

## 1. 实现概述

`lc detect` 命令已成功实现，用于自动探测当前 Git 仓库上下文信息。

## 2. 实现详情

### 2.1 新增文件

| 文件 | 说明 |
|------|------|
| `cmd/lc/detect.go` | detect 命令主实现（约 400 行） |
| `docs/detect-feature/requirements.md` | 需求文档 |
| `docs/detect-feature/design.md` | 设计文档 |
| `docs/detect-feature/implementation.md` | 本文档 |

### 2.2 修改文件

| 文件 | 修改内容 |
|------|---------|
| `internal/api/space.go` | 新增 `GetSpaceDetail()` 方法 |

### 2.3 核心实现

#### Git 探测 (`probeGitContext`)
```go
func probeGitContext(path string) (*GitContext, error) {
    // 1. 检查 git rev-parse --git-dir
    // 2. 获取 git rev-parse --show-toplevel
    // 3. 获取 git remote get-url origin
}
```

#### 仓库匹配 (`findMatchingRepository`)
- 调用 `ProjectService.SearchAllUserRepos()` 全局搜索
- 使用 `normalizeGitURL()` 标准化 URL 后比较
- 优先 URL 精确匹配，其次名称匹配

#### 空间详情获取 (`fetchSpaceDetails`)
- 调用 `SpaceService.GetSpaceDetail()`
- 从列表响应中筛选匹配的空间
- 提取 `spaceName` 字段

### 2.4 命令行接口

```bash
lc detect [flags]

Flags:
  -p, --path string   指定探测路径（默认为当前目录）
  -d, --debug         启用调试模式
  -k, --insecure      跳过 SSL 证书验证
      --dry-run       模拟执行
```

## 3. 功能验证

### 3.1 测试场景

| 场景 | 结果 | 备注 |
|------|------|------|
| 在 Git 仓库目录执行 | ✓ 通过 | 正确输出 workspaceKey 等信息 |
| 在非 Git 目录执行 | ✓ 通过 | 返回 matched=false |
| 指定 `--path` 参数 | ✓ 通过 | 可探测其他目录 |
| URL 匹配（HTTP） | ✓ 通过 | 正确匹配 tenantHttpPath |
| URL 匹配（SSH） | ✓ 通过 | 标准化后正确匹配 |
| 带端口号 URL | ✓ 通过 | :8022 端口处理正确 |

### 3.2 示例输出

```json
{
  "success": true,
  "data": {
    "workspaceKey": "XXJSSGKJCXPTYYTG",
    "workspaceName": "灵畿科研平台开发域",
    "tenantId": "ENTP750043923870622608",
    "repository": {
      "gitProjectId": 42231,
      "spaceCode": "XXJSSGKJCXPTYYTG",
      ...
    },
    "spaceDetails": {
      "spaceName": "灵畿科研平台开发域",
      "spaceDesc": "...",
      ...
    },
    "gitInfo": {
      "IsGitRepo": true,
      "RepoName": "lc",
      "RemoteURL": "http://code-xxjs.rdcloud.4c.hq.cmcc/osc/XXJS/weibaohui-hq.cmcc/lc.git"
    },
    "matched": true
  }
}
```

## 4. 使用示例

### 4.1 基础用法
```bash
# 探测当前目录
lc detect

# 探测指定目录
lc detect --path /path/to/repo

# 调试模式
lc detect -d
```

### 4.2 脚本自动化
```bash
# 获取 workspace key
WORKSPACE=$(lc detect -k | jq -r '.data.workspaceKey')

# 获取 git project id
PROJECT_ID=$(lc detect -k | jq -r '.data.repository.gitProjectId')

# 创建需求
lc req create "新功能" -w $(lc detect -k | jq -r '.data.workspaceKey')
```

## 5. 后续优化方向

1. **缓存机制**：对空间详情添加本地缓存，减少 API 调用
2. **配置自动识别**：读取 `.lc/config` 中的默认工作空间
3. **自动 workspace 推断**：其他命令在无 `-w` 参数时自动调用 detect
4. **多 remote 支持**：支持指定 remote 名称（不仅 origin）

## 6. 提交信息

```
feat: 新增 lc detect 命令自动探测 Git 仓库上下文

- 实现自动检测当前目录 Git 仓库信息
- 通过 Git URL 匹配灵畿平台远程仓库
- 输出研发空间 Key、名称、租户 ID 等上下文
- 支持脚本自动化，方便其他命令使用

新增文件:
- cmd/lc/detect.go
- docs/detect-feature/requirements.md
- docs/detect-feature/design.md
- docs/detect-feature/implementation.md
- e2e/integration/detect_test.go

修改文件:
- internal/api/space.go (添加 GetSpaceDetail 方法)
```
