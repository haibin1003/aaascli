# 山东能力平台 CLI 助手 - 开发进展记录

> 本文档记录项目的开发历程、关键决策和待办事项，供后续开发参考。

---

## 变更记录表

| 序号 | 变更时间 | 变更内容 | 变更人 | 版本 |
|------|----------|----------|--------|------|
| 1 | 2026-04-13 | 项目初始化：基于浏览器采集数据重构 API，实现能力查询、订购、授权全流程 | AI Agent | v0.1.0 |
| 2 | 2026-04-13 | 修复中文乱码（Windows 控制台 UTF-8）、清理无关文件、建立 release 发版包 | AI Agent | v0.1.0 |

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

采用 **HTTP 抓包 + Cookie 模拟** 的方案：

```
用户浏览器 → 插件提取 Cookie → AI 执行 CLI → HTTP 请求 → 山东能力平台
```

技术栈：
- **Go 1.21+**：CLI 主体（Cobra 框架）
- **Chrome Extension**：提取登录 Cookie
- **Node.js/npm**：跨平台分发（可选）

### 阶段 3：核心功能实现

#### 3.1 清理与重构
- 删除 `detect.go`
- 移除 `go.uber.org/zap` 依赖
- 简化 `config.go`

#### 3.2 API 层
- `client.go`：统一 HTTP 客户端
- `ability.go`：能力列表、搜索、详情、订购、我的能力
- `app.go`：应用列表、能力授权、授权状态查询

#### 3.3 CLI 命令层
- `ability list/search/view/order/my`
- `app list/auth-list/auth-ability/auth-status`
- `helper extract`
- `login`
- `onboard`

#### 3.4 npm 发布架构
建立了完整的 npm 包结构（主包 + 6 平台包），包名统一为 `@aaas/sd*` 系列。

### 阶段 4：问题修复

#### 4.1 中文乱码问题
**原因**：Windows 控制台默认代码页为 GBK（936），Go 程序输出 UTF-8 中文时被错误解析。

**解决**：
- 新增 `cmd/sdp/utf8_windows.go`（build tag: `windows`）
- 在 `init()` 中调用 `windows.SetConsoleOutputCP(65001)`
- 将 Windows 控制台输出代码页强制设为 UTF-8

**结果**：`sdp onboard` 等命令的中文输出正常显示。

#### 4.2 仓库清理
**原因**：初始提交包含了大量与 `sdopen-cli` 无关的 `lc` 项目文件。

**解决**：
- 清理根目录，仅保留 `sdopen-cli` 相关内容
- 将 `sdopen-cli/` 内文件提升至仓库根目录
- 更新 `go.mod` 模块路径为 `github.com/haibin1003/aaascli`
- 创建 `release/` 目录作为发版包

---

## 三、发版内容

### release/ 目录结构

```
release/
├── bin/
│   ├── sdp-windows-x64.exe    # Windows x64 二进制
│   ├── sdp-windows-arm64.exe  # Windows arm64 二进制
│   ├── sdp-linux-x64          # Linux x64 二进制
│   ├── sdp-linux-arm64        # Linux arm64 二进制
│   ├── sdp-darwin-x64         # macOS x64 二进制
│   └── sdp-darwin-arm64       # macOS arm64 二进制
├── sdp-login-helper/          # 浏览器插件
│   ├── manifest.json
│   ├── popup.html
│   └── popup.js
├── install-guide.md           # 安装使用手册
└── development-log.md         # 开发进展记录
```

**发给 AI 时，只需提供 `release/` 目录内容即可**，无需提供源码。

---

## 四、完整工作流示例

```bash
# 1. 搜索
sdp ability search "高精度定位" -p

# 2. 查看详情
sdp ability view CA2023xxxx -p

# 3. 订购
sdp ability order CA2023xxxx -p

# 4. 授权（需要 BOMC 工单编码）
sdp app auth-ability "新员工实战应用" --ability CA2023xxxx --bomc WOxxxx -p

# 5. 查状态
sdp app auth-status "新员工实战应用" -p
```

---

## 五、已知问题与限制

1. **API 模型待验证**：所有 API 模型基于浏览器脚本推断，实际运行时可能需要根据报错校准。
2. **Cookie 有效期短**：平台 Cookie 有效期较短，测试过程中需要频繁重新提取。
3. **未实现的功能**：
   - [ ] 对内服务（数字服务）的查询和订购
   - [ ] 网络域服务的查询和订购
   - [ ] 按分类筛选能力
   - [ ] 审批结果主动通知
   - [ ] MCP Server 集成

---

## 六、后续开发建议

### 6.1 短期（1-2 周）
1. 实际环境验证所有命令
2. 补充对内服务命令
3. 增强错误处理

### 6.2 中期（1 个月）
1. MCP Server 集成
2. 交互式选择
3. 审批状态轮询

### 6.3 长期（3 个月）
1. 日志审计
2. 批量操作
3. SDK 下载

---

## 七、重要文件速查

| 文件 | 作用 |
|------|------|
| `main.go` | CLI 入口 |
| `cmd/sdp/ability.go` | 能力命令 |
| `cmd/sdp/app.go` | 应用授权命令 |
| `internal/api/ability.go` | 能力 API |
| `internal/api/app.go` | 应用 API |
| `internal/api/client.go` | HTTP 客户端 |
| `cmd/sdp/utf8_windows.go` | Windows UTF-8 修复 |
| `Makefile` | 构建脚本 |
| `docs/design.md` | 设计文档 |
| `docs/install-guide.md` | 安装手册 |
| `release/` | 发版包 |

---

**记录时间**: 2026-04-13  
**记录版本**: v0.1.0  
**仓库地址**: https://github.com/haibin1003/aaascli
