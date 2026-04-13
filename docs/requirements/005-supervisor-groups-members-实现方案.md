# Supervisor 工作组列表和成员功能实现总结

## 变更记录

| 日期 | 版本 | 修改内容 | 修改人 |
|------|------|----------|--------|
| 2026-03-24 | v1.0 | 初始版本，添加工作组列表筛选和成员查询功能 | AI |

## 功能概述

为 `lc supervisor` 命令添加工作组列表筛选和工作组成员查询功能，支持参数化查询和CSV导出。

## 需求对应

- 用户要求：在 `lc supervisor` 下增加子命令获取工作组列表，字段都要提成参数
- 用户要求：获取工作组下的成员详情，字段都要提成参数

## 实现内容

### 1. API 层 (`internal/api/supervisor.go`)

#### 新增结构体
- `GroupMemberByGroupCode`: 工作组成员信息结构（通过groupCode查询）
  - 包含字段：ID, NickName, Phone, LeaderFlag
- `GroupMemberByGroupCodeResponse`: 工作组成员列表响应结构

#### 新增方法
- `GetGroupList(current, pageSize int, code, name, description string)`: 获取业务分组列表，支持分页和筛选
  - 支持参数：`current`, `pageSize`, `code`, `name`, `description`
  - API端点：`/supervision/system-api/business-accept/group/page`

- `GetGroupMembersByCode(groupCode string)`: 获取工作组成员列表（通过工作组编码）
  - 支持参数：`groupCode` (工作组编码)
  - API端点：`/supervision/system-api/business-accept/group/listUserListByGroupCode`

- `GetGroupMembers(deptId string)`: 获取工作组成员列表（兼容函数）
  - 支持参数：`deptId` (工作组编码)
  - 内部调用 `GetGroupMembersByCode`

### 2. 命令层 (`cmd/lc/supervisor.go`)

#### 修改 `groups list` 命令
- 添加参数支持：
  - `--page/-p`: 当前页码 (默认: 1)
  - `--size/-s`: 每页大小 (默认: 10)
  - `--code/-c`: 工作组编码筛选
  - `--name/-n`: 工作组名称筛选
  - `--description`: 工作组描述筛选
  - `--output/-o`: 输出到CSV文件

#### 新增 `groups members` 命令
- 命令路径：`lc supervisor groups members`
- 参数：
  - `--group-code/-c`: 工作组编码 (如果不指定则必须指定 --group-name-filter)
  - `--group-name-filter/-f`: 工作组名称筛选字符串 (用于自动查找名称包含此字符串的工作组)
  - `--output/-o`: 输出到CSV文件
- 功能：
  - 支持查询指定工作组编码的成员列表
  - 支持自动查找名称包含指定字符串的所有工作组，并汇总所有成员

### 3. 文档更新

#### `SUPERVISOR.md`
- 添加工作组列表查询示例
- 添加工作组成员查询示例
- 更新API参考部分，添加新方法说明

## 测试验证

### 编译测试
```bash
go build -o lc-final ./cmd/lc
# 构建成功
```

### 命令帮助测试
```bash
# 查看 groups 命令帮助
go run ./main.go supervisor groups --help
# 输出显示 list 和 members 子命令

# 查看 list 命令帮助
go run ./main.go supervisor groups list --help
# 显示所有筛选参数

# 查看 members 命令帮助
go run ./main.go supervisor groups members --help
# 显示 dept-id 参数
```

## 安全检查

1. **输入验证**: 参数通过cobra解析，API调用时确保参数安全
2. **认证流程**: 使用现有的认证机制，Cookie和认证信息安全存储
3. **权限控制**: 依赖监管平台的认证流程，未引入新的权限问题
4. **错误处理**: 统一错误处理，未暴露敏感信息

## 使用示例

### 查询工作组列表
```bash
# 查询全量工作组
lc supervisor groups list

# 分页查询
lc supervisor groups list --page 2 --size 20

# 按编码筛选
lc supervisor groups list --code it001

# 按名称筛选
lc supervisor groups list --name "知识库"

# 输出到CSV
lc supervisor groups list -o groups.csv
```

### 查询工作组成员
```bash
# 查询指定工作组编码的成员
lc supervisor groups members --group-code it010

# 自动查询名称包含 "1组" 的所有工作组成员
lc supervisor groups members --group-name-filter "1组"

# 输出到CSV
lc supervisor groups members --group-name-filter "1组" -o members.csv
```

## 待改进点

1. 可以添加更多筛选参数，如按标签筛选
2. 可以添加成员详情查看功能
3. 可以优化CSV导出的字段顺序

## 相关文件

- `internal/api/supervisor.go`: API层实现
- `cmd/lc/supervisor.go`: 命令层实现
- `docs/SUPERVISOR.md`: 使用文档
- `docs/requirements/005-supervisor-groups-members-实现总结.md`: 本文档
