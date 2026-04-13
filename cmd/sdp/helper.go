package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var helperOutputDir string

var helperCmd = &cobra.Command{
	Use:   "helper",
	Short: "浏览器插件助手",
	Long:  "安装浏览器插件，一键复制 login 认证命令",
}

var helperExtractCmd = &cobra.Command{
	Use:   "extract [output-dir]",
	Short: "释放浏览器扩展到指定目录",
	Long: `将浏览器登录辅助扩展释放到指定目录。

如果未指定输出目录，默认释放到桌面。`,
	Example: `  # 释放到桌面（默认）
  sdp helper extract

  # 释放到指定目录
  sdp helper extract /path/to/output`,
	RunE: runHelperExtract,
}

func init() {
	helperCmd.AddCommand(helperExtractCmd)
	helperExtractCmd.Flags().StringVarP(&helperOutputDir, "output", "o", "", "输出目录（默认：桌面）")
	rootCmd.AddCommand(helperCmd)
}

func runHelperExtract(cmd *cobra.Command, args []string) error {
	outputDir := helperOutputDir
	if outputDir == "" && len(args) > 0 {
		outputDir = args[0]
	}
	if outputDir == "" {
		outputDir = getDesktopPath()
	}

	absPath, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("无法解析路径: %w", err)
	}
	extensionDir := filepath.Join(absPath, "sdp-login-helper")

	if _, err := os.Stat(extensionDir); err == nil {
		if err := os.RemoveAll(extensionDir); err != nil {
			return fmt.Errorf("删除旧目录失败: %w", err)
		}
	}

	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := createExtensionFiles(extensionDir); err != nil {
		return fmt.Errorf("创建扩展文件失败: %w", err)
	}

	fmt.Printf("浏览器扩展已释放到: %s\n", extensionDir)
	fmt.Println("\n安装步骤:")
	fmt.Println("1. 打开 Chrome 浏览器，输入 chrome://extensions/")
	fmt.Println("2. 开启右上角的「开发者模式」")
	fmt.Println("3. 点击「加载已解压的扩展程序」")
	fmt.Println("4. 选择上述目录")
	fmt.Println("5. 登录平台后点击插件图标，复制 sdp login 命令")

	return nil
}

func getDesktopPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	candidates := []string{
		filepath.Join(homeDir, "Desktop"),
		filepath.Join(homeDir, "桌面"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return homeDir
}

func createExtensionFiles(dir string) error {
	manifest := `{
  "manifest_version": 3,
  "name": "山东能力平台助手",
  "version": "1.0",
  "description": "一键复制山东能力平台 CLI 登录命令",
  "permissions": [
    "cookies",
    "activeTab"
  ],
  "host_permissions": [
    "<all_urls>"
  ],
  "action": {
    "default_popup": "popup.html"
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0644); err != nil {
		return err
	}

	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>山东能力平台助手</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { width: 380px; padding: 16px; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif; background: #f5f7fa; }
        .header { text-align: center; margin-bottom: 16px; }
        .header h1 { font-size: 18px; color: #1a1a1a; margin-bottom: 4px; }
        .header p { font-size: 12px; color: #666; }
        .content { background: #fff; border-radius: 8px; padding: 12px; }
        .hint { font-size: 12px; color: #666; margin-bottom: 8px; }
        .command-box { background: #f0f2f5; border-radius: 6px; padding: 12px; font-family: 'Consolas', 'Monaco', monospace; font-size: 13px; color: #333; cursor: pointer; position: relative; word-break: break-all; line-height: 1.5; transition: background 0.2s; }
        .command-box:hover { background: #e4e7eb; }
        .copy-icon { position: absolute; right: 8px; top: 50%; transform: translateY(-50%); font-size: 14px; }
        .success-tip { text-align: center; color: #52c41a; font-size: 12px; margin-top: 8px; }
        .error { text-align: center; color: #ff4d4f; padding: 20px; font-size: 14px; }
        .promo-section { margin-top: 16px; padding-top: 12px; border-top: 1px dashed #d9d9d9; }
        .promo-title { font-size: 12px; color: #1890ff; margin-bottom: 8px; font-weight: 500; }
    </style>
</head>
<body>
    <div class="header">
        <h1>山东能力平台助手</h1>
        <p>快速获取 CLI 登录凭证</p>
    </div>
    <div id="content" class="content"></div>
    <script src="popup.js"></script>
</body>
</html>`
	if err := os.WriteFile(filepath.Join(dir, "popup.html"), []byte(html), 0644); err != nil {
		return err
	}

	js := `document.addEventListener('DOMContentLoaded', async () => {
    const content = document.getElementById('content');

    try {
        const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });

        if (!tab || !tab.url) {
            content.innerHTML = '<div class="error">请在山东能力平台页面点击本图标</div>';
            return;
        }

        const url = new URL(tab.url);

        if (!url.hostname.includes('10086.cn') && !url.hostname.includes('sd.10086.cn')) {
            content.innerHTML = '<div class="error">请在山东能力平台页面点击本图标</div>';
            return;
        }

        let allCookies = [];
        const currentCookies = await chrome.cookies.getAll({ domain: url.hostname });
        allCookies.push(...currentCookies);

        const parts = url.hostname.split('.');
        for (let i = 1; i < parts.length - 1; i++) {
            const parentDomain = parts.slice(i).join('.');
            const parentCookies = await chrome.cookies.getAll({ domain: parentDomain });
            allCookies.push(...parentCookies);
        }

        const seen = new Set();
        allCookies = allCookies.filter(c => {
            const key = c.domain + c.name;
            if (seen.has(key)) return false;
            seen.add(key);
            return true;
        });

        const tokenCookie = allCookies.find(c => c.name === '#openPortal#token#');

        if (tokenCookie && tokenCookie.value) {
            const token = tokenCookie.value;
            const loginCommand = 'sdp login ' + token;
            const aiInstallPrompt = '请帮我配置山东能力平台 CLI 助手（sdp），安装完成后执行 sdp onboard 命令';
            const copyIcon = '\uD83D\uDCCB';

            content.innerHTML =
                '<div class="hint">1. AI 配置指令（推荐）</div>' +
                '<div class="command-box" id="aiInstallBox" title="点击复制">' +
                    escapeHtml(aiInstallPrompt) +
                    '<span class="copy-icon">' + copyIcon + '</span>' +
                '</div>' +
                '<div class="success-tip" id="aiInstallTip" style="display:none;">\u2713 已复制，请发送给 AI</div>' +

                '<div class="hint" style="margin-top: 12px;">2. 登录命令（点击复制）</div>' +
                '<div class="command-box" id="authBox" title="点击复制">' +
                    escapeHtml(loginCommand) +
                    '<span class="copy-icon">' + copyIcon + '</span>' +
                '</div>' +
                '<div class="success-tip" id="authTip" style="display:none;">\u2713 已复制，请在终端执行</div>' +

                '<div class="promo-section">' +
                    '<div class="promo-title">\uD83D\uDE80 推荐给其他同学</div>' +
                    '<div class="command-box" id="promoBox" title="点击复制">' +
                        escapeHtml(aiInstallPrompt) +
                        '<span class="copy-icon">' + copyIcon + '</span>' +
                    '</div>' +
                    '<div class="success-tip" id="promoTip" style="display:none;">\u2713 已复制</div>' +
                '</div>';

            document.getElementById('aiInstallBox').addEventListener('click', () => {
                navigator.clipboard.writeText(aiInstallPrompt).then(() => {
                    const tip = document.getElementById('aiInstallTip');
                    tip.style.display = 'block';
                    setTimeout(() => tip.style.display = 'none', 2000);
                });
            });

            document.getElementById('authBox').addEventListener('click', () => {
                navigator.clipboard.writeText(loginCommand).then(() => {
                    const tip = document.getElementById('authTip');
                    tip.style.display = 'block';
                    setTimeout(() => tip.style.display = 'none', 2000);
                });
            });

            document.getElementById('promoBox').addEventListener('click', () => {
                navigator.clipboard.writeText(aiInstallPrompt).then(() => {
                    const tip = document.getElementById('promoTip');
                    tip.style.display = 'block';
                    setTimeout(() => tip.style.display = 'none', 2000);
                });
            });

        } else {
            content.innerHTML = '<div class="error">未找到登录凭证<br>请先登录山东能力平台</div>';
        }

    } catch (error) {
        console.error(error);
        content.innerHTML = '<div class="error">获取登录凭证失败<br>请刷新页面后重试</div>';
    }
});

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}`
	if err := os.WriteFile(filepath.Join(dir, "popup.js"), []byte(js), 0644); err != nil {
		return err
	}

	return nil
}
