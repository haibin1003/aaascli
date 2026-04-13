# 山东能力平台 CLI 助手 - 本地安装手册

> 本手册面向 AI 和人类用户，说明如何在本地安装使用 `sdp` 工具。

---

## 一、安装流程概览

```
1. 获取编译好的二进制文件（release/bin/）
2. 安装到系统 PATH
3. 释放浏览器插件
4. 用户登录平台并复制 token
5. AI 执行登录命令
6. 开始正常使用
```

---

## 二、环境要求

- **Chrome 浏览器**（用于安装插件提取 cookie）
- **操作系统**：Windows / macOS / Linux

**注意**：如果你选择使用预编译二进制，不需要安装 Go。如果要本地编译，需要 Go 1.21+。

---

## 三、安装步骤

### 步骤 1：获取二进制文件

从 `release/bin/` 目录中选择适合你操作系统的二进制文件：

| 平台 | 文件 |
|------|------|
| Windows x64 | `release/bin/sdp-windows-x64.exe` |
| Windows arm64 | `release/bin/sdp-windows-arm64.exe` |
| Linux x64 | `release/bin/sdp-linux-x64` |
| Linux arm64 | `release/bin/sdp-linux-arm64` |
| macOS x64 | `release/bin/sdp-darwin-x64` |
| macOS arm64 | `release/bin/sdp-darwin-arm64` |

### 步骤 2：安装到 PATH

**Windows（PowerShell）：**

```powershell
# 创建 bin 目录并复制
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin" | Out-Null
Copy-Item sdp-windows-x64.exe "$env:USERPROFILE\bin\sdp.exe"

# 将 $env:USERPROFILE\bin 加入系统环境变量 PATH
```

**macOS / Linux：**

```bash
# 安装到 /usr/local/bin
sudo cp sdp-linux-x64 /usr/local/bin/sdp
sudo chmod +x /usr/local/bin/sdp
```

### 步骤 3：验证安装

```bash
sdp --help
sdp version
```

---

## 四、浏览器插件安装（用户操作）

### 4.1 释放插件文件

```bash
# 释放到桌面（默认）
sdp helper extract

# 或释放到指定目录
sdp helper extract /path/to/output
```

执行后会在桌面生成 `sdp-login-helper/` 文件夹。

### 4.2 加载到 Chrome

1. 打开 Chrome 浏览器，输入地址：`chrome://extensions/`
2. 开启右上角「开发者模式」
3. 点击「加载已解压的扩展程序」
4. 选择桌面上的 `sdp-login-helper` 文件夹
5. 插件图标会出现在浏览器工具栏

### 4.3 获取登录 Token

1. 用户访问 https://service.sd.10086.cn/aaas/ 并登录
2. 点击浏览器工具栏的插件图标
3. 复制显示的命令：`sdp login xxxxxx`
4. 将该命令交给 AI

---

## 五、AI 登录与验证

### 5.1 执行登录

收到用户的 `sdp login <token>` 命令后，AI 执行：

```bash
sdp login <token>
```

### 5.2 验证连接

```bash
# 查询能力列表
sdp ability list -p

# 搜索能力
sdp ability search "定位" -p

# 查看应用列表
sdp app list -p
```

---

## 六、完整工作流示例

### 场景：订购并授权一个能力

```bash
# 1. 搜索需要的能力
sdp ability search "高精度定位" -p

# 2. 查看能力详情（替换为真实 ID）
sdp ability view CA2023xxxx -p

# 3. 订购能力
sdp ability order CA2023xxxx -p

# 4. 为应用授权（需要 BOMC 工单编码，替换为真实值）
sdp app auth-ability "新员工实战应用" --ability CA2023xxxx --bomc WOxxxx -p

# 5. 查看授权审批状态
sdp app auth-status "新员工实战应用" -p
```

---

## 七、常见问题

### Q1: 执行命令提示 "未登录，请先执行: sdp login <token>"

**原因**：配置文件中没有有效的 cookie。

**解决**：重新通过浏览器插件获取 token，执行 `sdp login <token>`。

### Q2: 插件点击后显示 "请在山东能力平台页面使用此插件"

**原因**：当前浏览器标签页不是山东能力平台页面。

**解决**：确保用户已打开 `https://service.sd.10086.cn/aaas/` 后再点击插件图标。

### Q3: `sdp helper extract` 释放插件失败

**原因**：目标目录没有写入权限。

**解决**：尝试指定其他输出目录：`sdp helper extract D:\temp`

### Q4: Windows 下中文显示乱码

**原因**：Windows 控制台默认代码页不是 UTF-8。

**解决**：最新版本已内置修复，若仍出现乱码，可手动执行 `chcp 65001` 后再运行命令。

---

## 八、给 AI 的速查卡

```bash
# 安装验证
sdp version
sdp --help

# 登录
sdp login <token>

# 能力查询
sdp ability list -p
sdp ability search "关键词" -p
sdp ability view <id> -p
sdp ability my -p

# 能力订购
sdp ability order <id> -p

# 应用管理
sdp app list -p
sdp app auth-list <应用名> -p
sdp app auth-status <应用名> -p

# 能力授权（需审批）
sdp app auth-ability <应用名> --ability <id> --bomc <工单编码> -p
```

---

## 九、配置文件位置

- **Windows**: `C:\Users\<用户名>\.sdp\config.json`
- **macOS/Linux**: `~/.sdp/config.json`

---

**文档版本**: v0.1.0  
**更新时间**: 2026-04-13
