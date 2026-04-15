# SDP - AI 使用手册

> 本手册面向 AI 助手，指导如何在拿到安装包后快速部署并使用「山东能力平台 CLI」工具。

---

## 一、这是什么

这是一个命令行工具（二进制名为 `sdp`），让你（AI）能够直接操作山东能力开放平台，代替用户在 CLI 中完成以下操作：

- 查询平台能力列表
- 搜索具体能力/服务
- 查看能力/服务详情
- 浏览能力下的服务菜单
- 查询用户的应用列表
- **订购能力 / 订购服务**
- **为应用授权能力**
- **查询订购申请进展**

**注意**：涉及订购、授权等写操作现在可以直接通过 CLI 完成，但 BOMC 工单号必须由用户提供。

---

## 二、安装包结构

用户给你的安装包通常包含以下内容：

```
sdp-package/
├── sdp-windows-x64.exe       # Windows x64
├── sdp-windows-arm64.exe     # Windows ARM64
├── sdp-linux-x64             # Linux x64
├── sdp-linux-arm64           # Linux ARM64
├── sdp-darwin-x64            # macOS Intel
├── sdp-darwin-arm64          # macOS Apple Silicon
├── sdp-login-helper/         # Chrome 浏览器扩展（需用户手动导入）
└── docs/
    └── ai-usage-guide.md     # 本手册
```

### 2.1 AI 安装步骤

根据你所在系统的架构，选择对应的二进制文件，重命名为 `sdp`（Windows 为 `sdp.exe`），并放入系统 PATH 中。

**Windows (PowerShell)**：
```powershell
# 假设当前目录是安装包解压目录
Copy-Item .\sdp-windows-x64.exe .\sdp.exe
# 临时加入 PATH（当前会话有效）
$env:PATH += ";$(Get-Location)"
# 验证
.\sdp.exe --help
```

**Linux / macOS (Bash/Zsh)**：
```bash
# 假设你是 x64 架构
chmod +x sdp-linux-x64
sudo mv sdp-linux-x64 /usr/local/bin/sdp
# 验证
sdp --help
```

---

## 三、浏览器插件导入（需用户配合）

由于平台使用 Cookie + Token 认证，且 Token 需要从已登录的浏览器会话中提取，**必须让用户手动完成以下步骤**：

### 步骤 1：释放插件文件

AI 执行：
```bash
sdp helper extract /path/to/sdp-login-helper
```

Windows 示例：
```powershell
.\sdp.exe helper extract C:\Users\<用户名>\Desktop\sdp-login-helper
```

### 步骤 2：用户导入 Chrome

将以下步骤原样告知用户：

1. 打开 Chrome 浏览器，地址栏输入 `chrome://extensions/`
2. 开启右上角「开发者模式」
3. 点击「加载已解压的扩展程序」
4. 选择 AI 刚才释放的 `sdp-login-helper` 文件夹
5. 插件名称应显示为「山东能力平台助手」

### 步骤 3：用户登录并获取 Token

1. 用户访问 https://service.sd.10086.cn/aaas/ 并完成登录
2. 点击浏览器右上角的插件图标
3. 插件会显示一行命令，例如：
   ```
   sdp login a500d28... --verification 5f3004... --service 599483...
   ```
4. 用户将这行命令复制发给你

### 步骤 4：AI 执行登录

收到用户命令后，直接执行：
```bash
sdp login <token> --verification <code> --service <id>
```

登录凭证会保存在 `~/.sdp/config.json` 中，后续命令自动读取。

### 步骤 5：验证登录

```bash
sdp ability list --size 3
```

如果能返回 JSON 数据（包含能力名称、ID 等），说明登录成功。

---

## 四、命令速查表

### 4.1 能力管理（ability）

| 命令 | 作用 |
|------|------|
| `sdp ability list --size 20` | 查询全部能力列表（共 325 个） |
| `sdp ability search "短信" --size 10` | 按关键词搜索能力 |
| `sdp ability view <ability-id>` | 查看能力详情 |
| `sdp ability services <ability-id> --size 10` | 查看该能力下的服务菜单 |

### 4.2 数字服务管理（service）

| 命令 | 作用 |
|------|------|
| `sdp service list --size 20` | 查询全量服务目录（目录+API 混合树） |
| `sdp service search "短信" --size 10` | 搜索具体 API 服务 |
| `sdp service view <service-id>` | 查看服务详情（含请求示例、URL、负责人等） |

### 4.3 应用管理（app）

| 命令 | 作用 |
|------|------|
| `sdp app list --size 10` | 查询用户已创建的应用列表 |
| `sdp ability order <ability-id>` | 提交能力订购申请 |
| `sdp app auth-ability <app-id> --ability <id> --bomc <ticket>` | 为应用授权能力 |
| `sdp service order <service-id> --app-id <id> --bomc <ticket>` | 订购服务（一步完成授权） |
| `sdp order list --size 10` | 查询我的订购/授权申请记录 |

### 4.4 辅助命令

| 命令 | 作用 |
|------|------|
| `sdp helper extract [目录]` | 释放浏览器登录插件 |
| `sdp onboard` | 显示快速入门指南 |
| `sdp --help` | 查看所有命令 |

---

## 五、命令详细说明与示例

### 5.1 ability list — 查询能力列表

```bash
sdp ability list --size 5
```

返回示例：
```json
{
  "items": [
    {
      "id": "CA202309081733156081000045960604",
      "name": "通用场景目标检测能力",
      "code": "",
      "desc": "基于深度学习技术对图像视频中的常见物体进行定位与识别...",
      "provider": "集团公司",
      "status": "",
      "category": "机器视觉"
    }
  ],
  "pagination": {
    "page": 1,
    "size": 5,
    "total": 325,
    "pages": 65
  }
}
```

### 5.2 ability search — 搜索能力

```bash
sdp ability search "短信定位" --size 10
```

说明：在全部 325 个能力中做客户端关键词过滤，支持名称、编码、分类名的模糊匹配。

### 5.3 ability view — 查看能力详情

```bash
sdp ability view CA202309081733156081000045960604
```

返回字段：
- `id` / `name` / `code` / `desc`：基础信息
- `provider`：提供方
- `type` / `callType`：能力类型/调用类型
- `detailDesc`：详细描述
- `userId`：归属用户 ID

注意：平台会对 `id` 做部分脱敏返回（如 `CA2*******...`），不影响使用。

### 5.4 ability services — 查看能力下的服务

```bash
sdp ability services CA202309131409091161012767301509 --size 10
```

说明：查询该能力在平台中挂载的服务菜单。如果返回空数组，说明该能力下暂无服务，或用户尚未订购该能力。

### 5.5 service list — 查询全量服务目录

```bash
sdp service list --size 20
```

说明：返回数字服务目录的展平列表，混合包含两类节点：
- `type: catalog` / `type: leaf-catalog` — 分类目录
- `type: api` — 具体 API 服务

api 节点包含 `id`（serviceId）、`name`、`code`（interfaceId）等。

### 5.6 service search — 搜索服务

```bash
sdp service search "短信" --size 10
```

说明：在全量服务目录中搜索，返回名称或编码匹配的服务节点。

### 5.7 service view — 查看服务详情

```bash
sdp service view SE202604081915467601182242
```

返回字段：
- `id` / `name` / `version`：服务标识与版本
- `requestTypeText`：HTTP 方法（如 POST）
- `requestUrl`：接口路径
- `protocol`：协议（如 HTTP）
- `interfaceId` / `serviceId`：接口/服务编码
- `remark`：功能说明
- `requestExample` / `responseExample`：请求/响应示例（JSON 字符串）
- `domainName`：所属域（如 B域）
- `owner` / `department` / `contactNo`：负责人及联系方式

### 5.8 app list — 查询我的应用

```bash
sdp app list --size 10
```

说明：返回用户已创建的应用列表，用于后续订购、授权等操作。

---

## 六、典型工作流程（AI 参考模板）

### 场景 A：用户想查找并了解某个能力

```bash
# 1. 搜索
sdp ability search "定位" --size 10

# 2. 查看详情
sdp ability view CA202512091059315001040209490536

# 3. 查看该能力下的服务
sdp ability services CA202512091059315001040209490536 --size 10
```

### 场景 B：用户想查找具体 API 接口

```bash
# 1. 搜索服务
sdp service search "短信下发" --size 10

# 2. 查看服务详情（获取接口定义和示例）
sdp service view SE202402071658102321000830
```

### 场景 C：用户想订购能力

```bash
# 1. 先搜索确认能力 ID
sdp ability search "大数据短信" --size 5

# 2. 查看详情确认
sdp ability view CA202507031538155471116699106450

# 3. 提交能力订购
sdp ability order CA202507031538155471116699106450

# 4. 为应用授权该能力（需用户提供 app-id 和 BOMC 工单号）
sdp app auth-ability <app-id> \
  --ability CA202507031538155471116699106450 \
  --bomc <bomc-ticket>
```

### 场景 D：用户想订购服务

```bash
# 1. 搜索确认服务 ID
sdp service search "营业厅信息查询" --size 5

# 2. 查看服务详情
sdp service view SE202604100446026231000121

# 3. 订购并授权给应用（一步完成，需用户提供 app-id 和 BOMC 工单号）
sdp service order SE202604100446026231000121 \
  --app-id <app-id> \
  --bomc <bomc-ticket>
```

### 场景 E：查询订购/授权进展

```bash
# 查询申请记录
sdp order list --size 10 --pass-status true
```

---

## 七、故障排查

### 7.1 提示 "请先登录"

- 说明 `~/.sdp/config.json` 不存在或已过期
- 让用户重新点击浏览器插件，复制 `sdp login ...` 命令给你执行

### 7.2 命令返回空数组

- `ability services` 返回空：说明该能力下确实没有挂载服务，或用户尚未订购
- `app list` 返回空：说明该账号没有创建过应用
- `service search` 返回空：关键词未匹配到任何服务，尝试换关键词
- `order list` 返回空：可能是平台查询条件限制，或当前确实没有申请记录

### 7.3 中文显示乱码（Windows）

- 工具已内置 GBK→UTF-8 自动转码，通常无需处理
- 如果终端仍然显示方块字，建议用户换用 Windows Terminal 或 PowerShell 7+

### 7.4 网络连接失败 / TLS 错误

- 平台使用自签名证书，工具默认跳过 TLS 验证
- 如果仍然报错，尝试加 `--insecure` 参数：
  ```bash
  sdp ability list --insecure
  ```

---

## 八、AI 增值建议（如何更好地帮助用户）

你不仅仅是一个命令执行器。拿到平台数据后，你可以主动为用户提供更高价值的输出：

### 8.1 基于能力列表输出解决方案建议

当用户描述业务场景但不知道选什么能力时，你可以：

1. 运行 `sdp ability list/search` 获取相关能力
2. 结合能力的 `category`（分类）、`desc`（描述）、`provider`（提供方）做匹配
3. 输出包含以下内容的能力选型方案：
   - **核心能力**：最匹配用户主诉求的 1-3 个能力
   - **辅助能力**：配合核心能力所需的支撑能力（如短信通知、定位、认证等）
   - **能力组合逻辑**：这些能力如何串联形成完整业务闭环
   - **接入建议**：哪些能力可以直接调用，哪些可能需要用户手动订购

**示例输出结构**：
```markdown
### 智慧景区客流预警方案

| 角色 | 能力名称 | 能力ID | 作用 |
|------|---------|--------|------|
| 核心 | 电子围栏服务能力 | CA2024... | 实时监测景区边界内人口动态 |
| 辅助 | 大数据短信触达能力 | CA2025... | 超员时向管理员发送预警短信 |
| 辅助 | 短信定位能力 | CA2025... | 获取游客实时位置辅助疏散 |

**业务流程建议**：
1. 通过「电子围栏」实时统计景区内人数
2. 当人数超过阈值时，触发「大数据短信触达」通知安保人员
3. 需要精准疏散时，调用「短信定位」获取游客位置分布
```

### 8.2 基于服务详情生成调用代码

当用户拿到 `service view` 的返回后，你可以根据接口元数据生成可直接运行的代码模板。

`service view` 返回的关键字段：
- `requestTypeText`：HTTP 方法（GET / POST）
- `requestUrl`：接口路径
- `protocol`：协议（HTTP / HTTPS）
- `requestExample`：请求体 JSON 示例（字符串）
- `responseExample`：响应体 JSON 示例（字符串）
- `serviceId` / `interfaceId`：接口标识

**你应该生成的内容**：
1. **基础调用代码**：使用用户指定的语言（Java / Python / Go / JavaScript / Shell 等）
2. **必要的常量/配置**：BaseURL、Header（如 `token`）、Content-Type
3. **请求体构造**：根据 `requestExample` 生成结构体 / Dict / Class
4. **响应体解析**：根据 `responseExample` 生成对应的解析逻辑
5. **错误处理**：HTTP 状态码检查、超时设置、异常捕获
6. **使用说明**：如何替换 token、必填参数有哪些

**示例（Python requests）**：
```python
import requests
import json

BASE_URL = "https://service.sd.10086.cn"
TOKEN = "你的token"

def check_line_bind_fw(subs_id, region, priv_id):
    url = f"{BASE_URL}/openapi/GroupServicer?method=checkLineBindFW"
    headers = {
        "Content-Type": "application/json",
        "token": TOKEN
    }
    payload = {
        "subsId": subs_id,
        "region": region,
        "privId": priv_id
    }
    resp = requests.post(url, headers=headers, json=payload, timeout=30)
    resp.raise_for_status()
    return resp.json()

if __name__ == "__main__":
    result = check_line_bind_fw("537*******", "537", "pg.vi.cmnet")
    print(json.dumps(result, indent=2, ensure_ascii=False))
```

### 8.3 其他增值服务建议

- **能力对比表**：当用户在多个相似能力间犹豫时，提取关键字段做成对比表格
- **接口文档草案**：基于 `service view` 输出一份 Markdown 格式的接口文档（含请求参数、返回值、示例）
- **项目 README 能力依赖说明**：帮用户整理 "本项目使用了山东能力平台的以下能力..."
- **订购检查清单**：在执行订购前，帮用户核对需要准备的材料（如 BOMC 工单号、应用名称等）
- **中文乱码/字段解释**：如果返回的字段名不够直观，主动解释每个字段的业务含义

---

## 九、文件位置速查

| 文件 | 路径 |
|------|------|
| 登录凭证 | `~/.sdp/config.json` |
| 浏览器插件 | 由 `sdp helper extract [dir]` 指定 |
| 二进制文件 | 当前目录或 `/usr/local/bin/sdp` |

---

**版本**：dev  
**最后更新**：2026-04-15
