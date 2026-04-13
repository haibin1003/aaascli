# 山东能力平台 CLI 助手 - 本地安装手册

> 本手册面向 AI 和人类用户，说明如何在本地编译安装 `sdp` 工具。

---

## 一、安装流程概览

```
1. 克隆/获取代码
2. 编译构建
3. 安装到系统 PATH
4. 释放浏览器插件
5. 用户登录平台并复制 token
6. AI 执行登录命令
7. 开始正常使用
```

---

## 二、环境要求

- **Go 1.21 或更高版本**
- **Chrome 浏览器**（用于安装插件提取 cookie）
- **操作系统**：Windows / macOS / Linux

---

## 三、编译安装步骤

### 步骤 1：进入项目目录

项目代码位于：

```bash
# 根据实际路径调整
cd C:\Users\51554\code\lc\lc\sdopen-cli
```

### 步骤 2：编译

```bash
# Windows
go build -o sdp.exe main.go

# macOS / Linux
go build -o sdp main.go
```

编译成功后会生成当前平台的可执行文件：
- Windows: `sdp.exe`
- macOS/Linux: `sdp`

### 步骤 3：安装到 PATH（推荐）

**Windows（PowerShell）：**

```powershell
# 方案 A：复制到已有 PATH 目录（如 C:\Windows\System32，需要管理员权限）
Copy-Item sdp.exe C:\Windows\System32\sdp.exe

# 方案 B：创建目录并加入 PATH
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin" | Out-Null
Copy-Item sdp.exe "$env:USERPROFILE\bin\sdp.exe"
# 然后将 $env:USERPROFILE\bin 加入系统 PATH 环境变量
```

**macOS / Linux：**

```bash
# 安装到 /usr/local/bin
sudo cp sdp /usr/local/bin/sdp
sudo chmod +x /usr/local/bin/sdp
```

### 步骤 4：验证安装

```bash
sdp --help
sdp version
```

如果显示帮助信息和版本号，说明安装成功。

---

## 四、浏览器插件安装（用户操作）

### 4.1 释放插件文件

```bash
# 释放到桌面（默认）
sdp helper extract

# 或释放到指定目录
sdp helper extract /path/to/output
```

执行后会在桌面生成 `sdp-login-helper/` 文件夹，包含：
- `manifest.json`
- `popup.html`
- `popup.js`

### 4.2 加载到 Chrome

1. 打开 Chrome 浏览器，输入地址：`chrome://extensions/`
2. 开启右上角「开发者模式」
3. 点击「加载已解压的扩展程序」
4. 选择桌面上的 `sdp-login-helper` 文件夹
5. 插件图标会出现在浏览器工具栏

### 4.3 获取登录 Token

1. 用户访问 [https://service.sd.10086.cn/aaas/](https://service.sd.10086.cn/aaas/) 并登录
2. 点击浏览器工具栏的插件图标
3. 复制显示的命令：`sdp login xxxxxx`
4. 将该命令交给 AI

---

## 五、AI 登录与验证

### 5.1 执行登录

用户把 token 给 AI 后，AI 执行：

```bash
sdp login <token>
```

登录成功后会显示：
```
登录成功
配置文件: C:\Users\xxx\.sdp\config.json
提示: 使用 'sdp ability list' 查看能力列表
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

如果能正常返回 JSON 数据，说明整个链路已打通。

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

**解决**：
1. 确认用户已在浏览器登录平台
2. 点击插件图标获取 `sdp login xxx` 命令
3. 重新执行登录命令

### Q2: 编译报错 "go: command not found"

**原因**：系统没有安装 Go 或 Go 不在 PATH 中。

**解决**：
- Windows: 使用 `scoop install go` 或从 [https://go.dev/dl/](https://go.dev/dl/) 下载安装
- macOS: `brew install go`
- Linux: `sudo apt install golang-go`

### Q3: 插件点击后显示 "请在山东能力平台页面使用此插件"

**原因**：当前浏览器标签页不是山东能力平台页面。

**解决**：确保用户已经打开了 `https://service.sd.10086.cn/aaas/` 后再点击插件图标。

### Q4: `sdp helper extract` 释放插件失败

**原因**：可能目标目录没有写入权限。

**解决**：尝试指定其他输出目录：
```bash
sdp helper extract D:\temp
```

### Q5: 返回的 JSON 中中文显示为 Unicode 编码

**原因**：这是 JSON 标准编码方式，不影响 AI 解析。

**解决**：使用 `-p`（pretty）参数可以格式化输出，但 Unicode 编码仍会保留。

---

## 八、给 AI 的速查卡

```
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

文件内容示例：
```json
{
  "cookie": "your-token-here"
}
```

---

**文档版本**: v0.1.0  
**更新时间**: 2026-04-13
