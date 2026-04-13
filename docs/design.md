# 山东能力平台 CLI 助手 - 设计文档

## 变更记录表

| 序号 | 变更时间 | 变更内容 | 变更人 | 版本 |
|------|----------|----------|--------|------|
| 1 | 2026-04-13 | 项目初始化：基于浏览器采集数据重构 API，实现能力查询、订购、授权全流程，建立 npm 发布架构 | AI Agent | v0.1.0 |

---

## 1. 项目背景

### 1.1 目标

为 AI 提供一个可以直接操作山东能力开放平台（https://service.sd.10086.cn/aaas/）的命令行工具。

### 1.2 平台特点

- **平台名称**：山东省级能力开放平台
- **URL**：https://service.sd.10086.cn/aaas/
- **架构**：SSR + Vue2，无公开 API
- **认证方式**：基于 Cookie（`#openPortal#token#`）
- **服务类型**：对外服务（能力广场）、对内服务（数字服务）、网络域

### 1.3 范围

本文档涵盖 CLI 工具的整体架构设计、API 模型设计、命令设计以及实现方案。

---

## 2. 架构设计

### 2.1 整体架构

```
用户/AI  ->  sdp CLI  ->  HTTP Client  ->  山东能力平台
                |
                v
         浏览器插件（Cookie 提取）
```

### 2.2 模块划分

| 模块 | 职责 | 对应目录 |
|------|------|----------|
| CLI 命令 | 解析用户输入，调用业务逻辑 | `cmd/sdp/` |
| API 客户端 | 封装 HTTP 请求和响应解析 | `internal/api/` |
| 配置管理 | 管理本地 Cookie 和配置 | `internal/config/` |
| 通用工具 | 命令执行器、格式化输出 | `internal/common/` |
| 浏览器插件 | 从浏览器提取登录凭证 | `sdp-login-helper/` |

---

## 3. API 设计

### 3.1 认证方式

- 通过 Cookie 头部发送：`Cookie: #openPortal#token#=<token>`
- Token 通过浏览器插件提取
- Token 保存在 `~/.sdp/config.json`

### 3.2 基础 URL

```
https://service.sd.10086.cn/aaas
```

### 3.3 能力服务 API（对外服务）

| API | 方法 | 路径 | 说明 |
|-----|------|------|------|
| 能力列表 | POST | `/openProtal/ability/list` | 分页查询能力列表 |
| 能力详情 | GET | `/openProtal/ability/detail/{id}` | 查看能力详情 |
| 订购能力 | POST | `/openProtal/ability/order` | 订购指定能力 |
| 我的能力 | POST | `/openProtal/ability/myList` | 查询已订购能力 |

### 3.4 应用服务 API

| API | 方法 | 路径 | 说明 |
|-----|------|------|------|
| 应用列表 | POST | `/openProtal/app/list` | 查询我的应用 |
| 应用详情 | GET | `/openProtal/app/detail/{id}` | 查看应用详情 |
| 授权能力 | POST | `/openProtal/app/authAbility` | 提交能力授权申请 |
| 已授权列表 | POST | `/openProtal/app/authAbilityList` | 查询已授权能力 |
| 授权状态 | POST | `/openProtal/app/authStatus` | 查询审批状态 |

---

## 4. CLI 命令设计

### 4.1 命令结构

```
sdp
├── login [token]              # 登录
├── ability                    # 能力管理（对外服务）
│   ├── list                   # 查询能力列表
│   ├── search <keyword>       # 搜索能力
│   ├── view <id>              # 查看能力详情
│   ├── order <id>             # 订购能力
│   └── my                     # 我的能力（已订购）
├── app                        # 应用管理
│   ├── list                   # 我的应用列表
│   ├── auth-list <app>        # 已授权能力列表
│   ├── auth-ability <app>     # 授权能力（需审批）
│   └── auth-status <app>      # 授权审批状态
├── helper extract             # 释放浏览器插件
├── onboard                    # AI 入门指南
└── version                    # 版本信息
```

### 4.2 全局参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-d, --debug` | bool | false | 启用调试模式 |
| `-k, --insecure` | bool | true | 跳过 TLS 证书验证 |
| `-p, --pretty` | bool | false | 格式化 JSON 输出 |
| `--dry-run` | bool | false | 模拟执行 |
| `-c, --cookie` | string | "" | 手动指定 cookie |

### 4.3 输出格式

统一使用 JSON 格式输出，便于 AI 解析。

---

## 5. 核心业务流程

### 5.1 能力订购流程

```
sdp ability order <ability-id>
  -> 检查登录状态
  -> POST /openProtal/ability/order
  -> 返回订购成功信息
  -> 提示下一步：执行授权命令
```

### 5.2 能力授权流程

```
sdp app auth-ability <app-name> --ability <id> --bomc <bomc-id>
  -> 检查登录状态
  -> 查询应用列表，匹配应用名称获取应用 ID
  -> POST /openProtal/app/authAbility
  -> 返回提交结果（是否需审批）
```

---

## 6. npm 发布架构

### 6.1 包结构

采用「主包 + 平台包」的模式：

```
packages/
├── sdp/                       # 主包 (@aaas/sd)
├── sdp-linux-x64/            # Linux x64 平台二进制
├── sdp-linux-arm64/
├── sdp-darwin-x64/
├── sdp-darwin-arm64/
├── sdp-windows-x64/
└── sdp-windows-arm64/
```

### 6.2 构建流程

```bash
make build-npm      # 编译所有平台的二进制文件到 packages/*
make npm-publish    # 发布平台包和主包到 npm
make npm-release    # 完整发布流程
```

---

## 7. 浏览器插件设计

### 7.1 功能

- 读取当前浏览器标签页的 Cookie
- 提取 `#openPortal#token#` 的值
- 生成 `sdp login <token>` 命令
- 提供一键复制功能

### 7.2 分发方式

浏览器插件通过 npm 主包一起分发，内嵌在 `sdp-login-helper/` 目录中。

---

## 8. 配置管理

配置文件路径：`~/.sdp/config.json`

---

## 9. 安全设计

- Token 仅保存在本地配置文件
- 支持 TLS 跳过验证（用于内部环境）
- 命令执行前检查登录状态

---

## 10. 扩展计划

### 10.1 近期
- [ ] 用真实 Cookie 验证所有 API 并校准模型
- [ ] 对内服务（数字服务）查询和订购
- [ ] 按分类筛选能力

### 10.2 中期
- [ ] MCP Server 集成
- [ ] 交互式选择
- [ ] 审批状态轮询

### 10.3 长期
- [ ] 日志审计
- [ ] 批量操作
- [ ] SDK 下载
