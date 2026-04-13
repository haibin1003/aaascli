# 山东能力平台 CLI 助手 (SDP)

为 AI 设计的山东能力开放平台命令行工具，支持查询能力、订购能力、授权管理等功能。

## 功能特性

- 🔍 **能力查询** - 查询平台上的对外服务能力列表
- 🔎 **能力搜索** - 根据关键词搜索能力
- 📋 **能力详情** - 查看能力的详细信息
- 🛒 **能力订购** - 订购对外服务能力
- 🔐 **应用授权** - 为应用授权能力（需审批）
- 📊 **授权状态** - 查看授权审批状态
- 🧩 **浏览器插件** - 一键获取登录凭证

## 快速开始

### 方式 1：使用预编译二进制（推荐，发给 AI 时使用）

从 `release/bin/` 目录获取对应平台的二进制文件：

- Windows x64: `release/bin/sdp-windows-x64.exe`
- Linux x64: `release/bin/sdp-linux-x64`
- macOS x64: `release/bin/sdp-darwin-x64`
- macOS arm64: `release/bin/sdp-darwin-arm64`

将二进制文件重命名为 `sdp`（或 `sdp.exe`），加入系统 PATH 即可使用。

### 方式 2：本地编译

```bash
# 进入项目根目录
cd aaascli

# 编译
go build -o sdp.exe main.go
```

### 安装浏览器插件

```bash
# 释放浏览器插件到桌面
sdp helper extract
```

然后：
1. 打开 Chrome 浏览器，输入 `chrome://extensions/`
2. 开启右上角的「开发者模式」
3. 点击「加载已解压的扩展程序」
4. 选择桌面上的 `sdp-login-helper` 文件夹

### 登录

1. 访问 [山东能力开放平台](https://service.sd.10086.cn/aaas/) 并登录
2. 点击浏览器工具栏的插件图标
3. 复制显示的 `sdp login <token>` 命令
4. 在终端执行该命令

### 开始使用

```bash
# 查询能力列表
sdp ability list

# 搜索能力
sdp ability search "定位"

# 查看能力详情
sdp ability view <ability-id>

# 订购能力
sdp ability order <ability-id>

# 查看我的应用
sdp app list

# 为应用授权能力（需要审批）
sdp app auth-ability <应用名> --ability <能力ID> --bomc <工单编码>

# 查看授权审批状态
sdp app auth-status <应用名>
```

## 命令列表

### 核心命令

| 命令 | 说明 |
|------|------|
| `sdp login <token>` | 使用 token 登录 |
| `sdp ability list` | 查询能力列表 |
| `sdp ability search <keyword>` | 搜索能力 |
| `sdp ability view <id>` | 查看能力详情 |
| `sdp ability order <id>` | 订购能力 |
| `sdp ability my` | 查看我的能力（已订购） |
| `sdp app list` | 查看我的应用列表 |
| `sdp app auth-list <app>` | 查看应用已授权的能力 |
| `sdp app auth-ability <app>` | 为应用授权能力 |
| `sdp app auth-status <app>` | 查看授权审批状态 |
| `sdp helper extract` | 释放浏览器插件 |
| `sdp onboard` | AI 入门指南 |
| `sdp version` | 显示版本信息 |

### 全局参数

| 参数 | 说明 |
|------|------|
| `-d, --debug` | 启用调试模式 |
| `-k, --insecure` | 跳过 TLS 证书验证 |
| `-p, --pretty` | 格式化 JSON 输出 |
| `--dry-run` | 模拟执行 |
| `-c, --cookie` | 手动指定 cookie |

## 项目结构

```
├── cmd/sdp/              # CLI 命令实现
├── internal/             # API 客户端和通用工具
├── release/              # 发版内容（二进制 + 部署文档）
├── sdp-login-helper/     # 浏览器插件源码
├── docs/                 # 设计文档和安装手册
├── main.go               # 入口
├── go.mod                # Go 模块定义
└── README.md             # 说明文档
```

## 配置

配置文件保存在 `~/.sdp/config.json`：

```json
{
  "cookie": "your-token-here"
}
```

## 注意事项

1. **Cookie 有效期** - Token 有一定的有效期，过期后需要重新登录
2. **授权需审批** - 能力授权申请需要审批，请耐心等待
3. **工单编码** - 授权时需要提供 BOMC 工单编码
4. **安全** - Token 只保存在本地，不会上传到任何服务器

## 开发

### 构建

```bash
# 构建当前平台
go build -o sdp.exe main.go

# 运行测试
go test ./...

# 格式化代码
go fmt ./...
```

## 许可证

MIT License
