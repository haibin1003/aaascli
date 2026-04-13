# lc 登录助手

Chrome 扩展，一键复制 `lc login` 登录命令。

## 功能

- 自动获取灵畿平台的登录 Cookie（支持 HTTP-only）
- 生成 `lc login <token>` 命令
- 点击即可复制，直接粘贴到终端运行

## 安装方法

### 1. 加载未打包扩展

1. 打开 Chrome/Edge 浏览器
2. 地址栏输入: `chrome://extensions/` 或 `edge://extensions/`
3. 开启右上角的"开发者模式"
4. 点击"加载已解压的扩展程序"
5. 选择 `lc-login-helper-extension` 文件夹

### 2. 固定扩展图标（可选）

- 点击浏览器右上角的拼图图标 🧩
- 找到 "lc 登录助手"
- 点击 📌 固定到工具栏

## 使用方法

1. **登录灵畿平台**（如 https://rdcloud.4c.hq.cmcc）
2. **点击扩展图标**，弹出窗口显示 `lc login xxx...` 命令
3. **点击命令**即可复制到剪贴板
4. **粘贴到终端**运行，完成 lc 登录

> 提示：如果显示"请先登录到灵畿平台"，请检查是否已登录。

## 文件说明

| 文件 | 说明 |
|------|------|
| `manifest.json` | 扩展配置 |
| `popup.html/js` | 弹出窗口界面和逻辑 |
| `icon*.png` | 扩展图标（蓝色背景白色 lc 文字） |

## 注意事项

- 需要登录灵畿平台后才能获取 Cookie
- 支持 HTTP-only Cookie（浏览器 JS 无法读取，扩展可以）
