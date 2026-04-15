# 山东能力平台 CLI 助手 (SDP)

> 专为 AI 助手打造的山东能力开放平台命令行工具。
> 
> 让 AI 从「只能建议」升级为「可以直接执行」——查询能力目录、检索 API 服务、查看接口详情、生成调用代码，一站式完成。

---

## 这是什么

**山东能力开放平台 CLI 助手**（简称 `sdp`）是一款面向 AI 助手的命令行客户端，连接 [山东能力开放平台](https://service.sd.10086.cn/aaas/)。

平台汇集了 **325+ 业务能力** 和 **3800+ 数字服务/API**，但这个门户原本只能通过浏览器访问。本工具通过逆向平台接口协议，让 AI 能够直接在命令行中查询平台数据，并基于返回结果为用户提供方案建议、代码生成等高价值服务。

### 核心设计目标

- **让 AI 能动手**：不再只是「告诉用户去哪里点」，而是直接查询并返回结构化结果
- **提升匹配效率**：秒级从 300+ 能力中找到符合业务需求的选项
- **降低技术门槛**：AI 可根据接口的 `requestExample` / `responseExample` 自动生成调用代码
- **知识沉淀**：内置常用能力的使用指南、场景方案和代码模板

---

## 功能特性

### 能力管理（Ability）
- `list` / `search` — 查询/搜索全部 325+ 业务能力
- `view` — 查看能力详情（描述、提供方、分类、调用类型）
- `services` — 查看能力下挂载的服务菜单
- `my` — 查询我的已订购能力

### 数字服务管理（Service）
- `list` / `search` — 查询/搜索全量 3800+ API 服务目录
- `view` — 查看服务详情，包含：
  - 请求方法、URL、协议
  - 请求/响应示例（JSON）
  - 接口负责人、部门、联系方式

### 知识库（Knowledge）
- `list` / `view` — 阅读内置 Markdown 知识文档
- `search` — 在知识库中全文搜索关键词，自动提取上下文片段
- 涵盖短信触达、机器视觉、定位服务、代码示例、常见问题等主题

### 认证与辅助
- `helper extract` — 释放 Chrome 浏览器扩展，用于提取登录 Token
- `login` — 保存 Cookie/Token，后续命令自动认证
- `onboard` — 面向 AI 的完整入门指南

---

## 安装方式

### 方式一：预编译二进制（发给 AI 时推荐）

访问 [GitHub Releases](https://github.com/haibin1003/aaascli/releases) 下载最新版本的安装包或对应平台的二进制文件：

| 平台 | 文件 |
|------|------|
| Windows x64 | `sdp-windows-x64.exe` |
| Windows ARM64 | `sdp-windows-arm64.exe` |
| Linux x64 | `sdp-linux-x64` |
| Linux ARM64 | `sdp-linux-arm64` |
| macOS Intel | `sdp-darwin-x64` |
| macOS Apple Silicon | `sdp-darwin-arm64` |

将文件重命名为 `sdp`（Windows 为 `sdp.exe`），加入系统 PATH 即可。

### 方式二：从源码编译

```bash
# 克隆仓库
git clone git@github.com:haibin1003/aaascli.git
cd aaascli

# 编译当前平台
go build -o sdp.exe .

# 编译所有平台（参考 Makefile）
make build-all
```

> 本项目所有依赖均已 Vendor 化（包括 `golang.org/x/text`），即使在无法访问外网的离线环境也能正常编译。

---

## 使用流程（AI + 用户配合）

### 第 1 步：AI 安装 CLI
AI 根据运行环境选择对应二进制文件，放入 PATH，验证：
```bash
sdp --help
```

### 第 2 步：AI 释放浏览器插件
```bash
sdp helper extract ~/Desktop/sdp-login-helper
```
AI 将释放路径告知用户。

### 第 3 步：用户导入插件并登录
1. 打开 Chrome，访问 `chrome://extensions/`
2. 开启「开发者模式」
3. 点击「加载已解压的扩展程序」，选择 `sdp-login-helper` 文件夹
4. 访问 [https://service.sd.10086.cn/aaas/](https://service.sd.10086.cn/aaas/) 登录平台
5. 点击插件图标，复制弹出的命令：
   ```bash
   sdp login <token> --verification <code> --service <id>
   ```
6. 将该命令发送给 AI

### 第 4 步：AI 执行登录并验证
```bash
sdp login <token> --verification <code> --service <id>
sdp ability list --size 3
```

如果能返回 JSON 数据，说明连接成功，AI 可以开始为用户服务了。

---

## 命令速查

```bash
# ===== 能力管理 =====
sdp ability list --size 20              # 查询全部能力
sdp ability search "短信定位" --size 10  # 搜索能力
sdp ability view <ability-id>            # 查看能力详情
sdp ability services <ability-id>        # 查看能力下的服务
sdp ability my --size 10                 # 我的已订购能力

# ===== 数字服务管理 =====
sdp service list --size 20               # 查询全量服务目录
sdp service search "短信下发" --size 10  # 搜索 API 服务
sdp service view <service-id>            # 查看服务详情（含示例）

# ===== 知识库 =====
sdp knowledge list                       # 查看内置知识文档
sdp knowledge view <doc-name>            # 阅读指定文档
sdp knowledge search <keyword>           # 搜索知识库

# ===== 应用管理 =====
sdp app list --size 10                   # 查看我的应用

# ===== 辅助命令 =====
sdp helper extract [dir]                 # 释放浏览器插件
sdp onboard                              # AI 入门指南
sdp version                              # 显示版本
```

### 全局参数

| 参数 | 说明 |
|------|------|
| `-d, --debug` | 启用调试模式，输出详细请求信息 |
| `-k, --insecure` | 跳过 TLS 证书验证 |
| `-p, --pretty` | 格式化 JSON 输出 |
| `--dry-run` | 模拟执行 |
| `-c, --cookie` | 手动指定 cookie |

---

## AI 增值建议

安装完成后，AI 不应只做「命令执行器」，而可以主动提供以下高价值输出：

### 1. 解决方案推荐
当用户描述业务场景（如「智慧景区客流预警」）时，AI 可以：
1. 查询平台相关能力
2. 输出「核心能力 + 辅助能力 + 业务流程」的完整方案

### 2. 调用代码生成
当用户需要对接某个 API 时，AI 可以基于 `service view` 返回的：
- `requestTypeText`（HTTP 方法）
- `requestUrl`（接口路径）
- `requestExample` / `responseExample`（JSON 示例）
- `protocol`（协议）

自动生成 Python / Java / Go / Shell 等调用代码。

### 3. 文档整理
帮用户把查询结果整理成：
- Markdown 能力对比表格
- 接口文档草案
- 项目 README 中的能力依赖说明

---

## 项目结构

```
├── cmd/sdp/              # CLI 命令实现
│   ├── ability.go        # 能力管理命令
│   ├── service.go        # 数字服务管理命令
│   ├── knowledge.go      # 知识库命令
│   ├── app.go            # 应用管理命令
│   ├── helper.go         # 浏览器插件释放
│   ├── login.go          # 登录认证
│   ├── onboard.go        # AI 入门指南
│   └── root.go           # 根命令
├── internal/
│   ├── api/              # HTTP 客户端 + 平台 API 封装
│   │   ├── client.go     # Cookie/Token 注入、GBK 转码
│   │   ├── crypto.go     # RSA+AES 混合加密
│   │   ├── ability.go    # 能力服务
│   │   └── service.go    # 数字服务
│   ├── common/           # 通用执行器
│   ├── config/           # 配置文件管理
│   └── knowledge/        # 内置知识库（go:embed）
├── release/
│   ├── bin/              # 6 平台预编译二进制
│   └── sdp-package.zip   # 完整安装包（含插件 + 文档）
├── sdp-login-helper/     # Chrome 浏览器插件源码
├── docs/
│   ├── ai-usage-guide.md         # AI 使用手册
│   ├── user-prompt-template.md   # 给 AI 的提示词模板
│   ├── presentation-report.html  # 项目汇报材料
│   └── knowledge/        # 知识库文档副本（便于阅读）
├── main.go
├── go.mod
└── README.md
```

---

## 配置说明

登录凭证保存在 `~/.sdp/config.json`：

```json
{
  "cookie": "your-token-here",
  "verificationCode": "your-verification-code",
  "serviceId": "your-service-id"
}
```

> 该文件仅存储在本地，不会上传到任何第三方服务器。

---

## 注意事项

1. **Token 有效期** - 平台 Cookie 有一定有效期，过期后需要重新通过浏览器插件获取
2. **订购与授权** - `ability order` 和 `app auth-ability` 目前需要用户手动在网页端完成
3. **服务菜单为空** - `ability services` 返回空数组时，通常表示该能力下暂无服务，或用户尚未订购
4. **安全** - 工具默认跳过 TLS 验证以适配内网环境，如遇到证书问题可加 `--insecure` 参数

---

## 开发

```bash
# 格式化代码
go fmt ./...

# 运行测试
go test ./...

# 构建当前平台
go build -o sdp.exe .

# 构建所有平台
make build-all
```

---

## 相关文档

- [AI 使用手册](docs/ai-usage-guide.md) — 面向 AI 助手的详细操作指南
- [用户提示词模板](docs/user-prompt-template.md) — 可直接复制发给 AI 的提示词
- [项目汇报材料](docs/presentation-report.html) — HTML 格式的项目介绍与成效汇报

---

## 许可证

MIT License
