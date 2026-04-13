# 山东能力平台 CLI 助手 - 开发进展记录

> 本文档记录 `sdopen-cli` 项目的开发历程、关键决策和待办事项，供后续开发参考。

---

## 变更记录表

| 序号 | 变更时间 | 变更内容 | 变更人 | 版本 |
|------|----------|----------|--------|------|
| 1 | 2026-04-13 | 项目初始化：基于浏览器采集数据重构 API，实现能力查询、订购、授权全流程，建立 npm 发布架构 | AI Agent | v0.1.0 |

---

## 一、开发背景

### 1.1 目标
为 AI 提供一个可以直接操作山东能力开放平台（https://service.sd.10086.cn/aaas/）的命令行工具。

### 1.2 平台特性
- **架构**：SSR + Vue2，无公开 API
- **认证**：基于 Cookie（`#openPortal#token#`）
- **服务类型**：对外服务（能力广场）、对内服务（数字服务）、网络域

---

## 二、关键开发节点

### 阶段 1：平台调研（浏览器脚本采集）

通过浏览器 Console 脚本采集页面结构：

| 脚本 | 采集内容 | 关键发现 |
|------|----------|----------|
| Script 1 | 数字服务列表 | 200+ 服务，分 Internal/External/Network 三类 |
| Script 2 | 能力详情页 | 发现"立即订购"/"立即退订"按钮 |
| Script 3 | 我的应用页 | 找到"新员工实战应用"，支持"能力授权"操作 |
| Script 4 | 授权弹窗 | 确认授权需填写 BOMC 工单编码，需审批 |

**结论**：
- 订购能力：简单确认弹窗，无需审批
- 授权能力：需要选择能力 + 填写 BOMC 工单编码 + 审批

### 阶段 2：技术架构设计

由于平台无公开 API，采用 **HTTP 抓包 + Cookie 模拟** 的方案：

```
用户浏览器 → 插件提取 Cookie → AI 执行 CLI → HTTP 请求 → 山东能力平台
```

技术栈：
- **Go 1.21+**：CLI 主体（Cobra 框架）
- **Chrome Extension**：提取登录 Cookie
- **Node.js/npm**：跨平台分发（可选）

### 阶段 3：核心功能实现

#### 3.1 清理与重构
- 删除 `detect.go`（功能已整合到 `onboard`）
- 移除 `go.uber.org/zap` 依赖，简化日志系统
- 简化 `config.go`，仅保留 Cookie 配置

#### 3.2 API 层（`internal/api/`）
- `client.go`：统一 HTTP 客户端，Cookie 认证，TLS 跳过验证
- `ability.go`：能力列表、搜索、详情、订购、我的能力
- `app.go`：应用列表、能力授权、授权状态查询

#### 3.3 CLI 命令层（`cmd/sdp/`）
- `ability list/search/view/order/my`：能力管理
- `app list/auth-list/auth-ability/auth-status`：应用授权管理
- `helper extract`：释放浏览器插件
- `login`：Cookie 登录
- `onboard`：AI 入门指南

#### 3.4 npm 发布架构
为支持 `npm install -g @aaas/sd`，建立了完整的 npm 包结构：

```
packages/
├── sdp/              # 主包 (@aaas/sd)
├── sdp-linux-x64/    # Linux x64 平台二进制
├── sdp-linux-arm64/
├── sdp-darwin-x64/
├── sdp-darwin-arm64/
├── sdp-windows-x64/
└── sdp-windows-arm64/
```

主包入口 `bin/sdp.js`：
- 动态加载对应平台包
- 特殊拦截 `helper extract`，直接从 npm 包内复制浏览器插件

### 阶段 4：文档编写

| 文档 | 路径 | 内容 |
|------|------|------|
| 设计文档 | `docs/sdopen-cli/design.md` | 架构设计、API 设计、CLI 设计、安全设计 |
| 安装手册 | `docs/sdopen-cli/install-guide.md` | 本地编译安装、插件安装、使用流程 |
| 开发进展 | `docs/sdopen-cli/development-log.md` | 本文件 |

---

## 三、已确认的流程

### 对外服务能力全流程

```
1. 搜索能力
   sdp ability search "高精度定位"

2. 查看详情
   sdp ability view <能力ID>

3. 订购能力
   sdp ability order <能力ID>
   → 弹窗确认 → 订购成功

4. 授权给应用
   sdp app auth-ability <应用名> --ability <能力ID> --bomc <工单编码>
   → 提交申请 → 等待审批

5. 查看状态
   sdp app auth-status <应用名>
```

### 页面与 API 映射

| 功能 | 页面 URL | API 路径 |
|------|----------|----------|
| 能力列表 | `#/sdOpenPortal/abilityList` | `POST /openProtal/ability/list` |
| 能力详情 | `#/sdOpenPortal/capacityDetail?capacityId=xxx` | `GET /openProtal/ability/detail/{id}` |
| 订购 | 弹窗确认 | `POST /openProtal/ability/order` |
| 我的应用 | `#/openProtal/userIndex` | `POST /openProtal/app/list` |
| 能力授权 | 应用授权弹窗 | `POST /openProtal/app/authAbility` |
| 授权状态 | - | `POST /openProtal/app/authStatus` |

---

## 四、已知问题与限制

### 4.1 API 模型待验证
由于平台无公开 API 文档，所有 API 模型基于浏览器脚本推断。如果实际返回结构与推断不一致，可能出现 JSON 解析错误。

**解决方法**：运行命令时加上 `-d`（debug）参数，将响应体贴出即可校准模型。

### 4.2 Cookie 有效期短
平台 Cookie（`#openPortal#token#`）有效期较短，开发测试过程中需要频繁重新提取。

### 4.3 未实现的功能
- [ ] 对内服务（数字服务）的查询和订购
- [ ] 网络域服务的查询和订购
- [ ] 按分类筛选能力
- [ ] 审批结果主动通知
- [ ] MCP Server 集成（供 AI 直接调用）

---

## 五、后续开发建议

### 5.1 短期（1-2 周）
1. **实际环境验证**：用真实 Cookie 测试所有命令，根据报错校准 API 模型
2. **补充对内服务**：实现 `sdp service list/view/order` 命令（数字服务流程）
3. **错误处理增强**：针对平台特有的错误码提供更友好的提示

### 5.2 中期（1 个月）
1. **MCP Server 集成**：将 CLI 能力封装为 MCP Tools，让 AI 可以直接调用
2. **交互式选择**：对于应用选择、能力选择等场景，支持交互式 TUI
3. **审批状态轮询**：支持 `sdp app auth-status --watch` 轮询审批结果

### 5.3 长期（3 个月）
1. **日志审计**：记录所有 AI 操作日志，便于回溯和审计
2. **批量操作**：支持批量订购、批量授权
3. **SDK 下载**：集成能力的 SDK 下载功能

---

## 六、重要文件速查

| 文件 | 作用 |
|------|------|
| `sdopen-cli/main.go` | CLI 入口 |
| `sdopen-cli/cmd/sdp/ability.go` | 能力命令 |
| `sdopen-cli/cmd/sdp/app.go` | 应用授权命令 |
| `sdopen-cli/internal/api/ability.go` | 能力 API |
| `sdopen-cli/internal/api/app.go` | 应用 API |
| `sdopen-cli/internal/api/client.go` | HTTP 客户端 |
| `sdopen-cli/packages/sdp/bin/sdp.js` | npm 包入口 |
| `sdopen-cli/Makefile` | 构建脚本 |
| `docs/sdopen-cli/design.md` | 设计文档 |
| `docs/sdopen-cli/install-guide.md` | 安装手册 |

---

**记录时间**: 2026-04-13  
**记录版本**: v0.1.0  
**仓库地址**: https://github.com/haibin1003/aaascli
