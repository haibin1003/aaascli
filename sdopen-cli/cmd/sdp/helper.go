package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var helperOutputDir string

// helperCmd 浏览器插件助手命令
var helperCmd = &cobra.Command{
	Use:   "helper",
	Short: "浏览器插件助手",
	Long:  "安装浏览器插件，一键复制 login 认证命令",
}

// helperExtractCmd 释放插件命令
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

// runHelperExtract 运行释放插件
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

// getDesktopPath 获取桌面路径
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

// createExtensionFiles 创建扩展文件
func createExtensionFiles(dir string) error {
	manifest := `{
  "manifest_version": 3,
  "name": "山东能力平台登录助手",
  "version": "1.0",
  "description": "一键复制山东能力平台 CLI 登录命令",
  "permissions": ["cookies", "activeTab"],
  "host_permissions": ["https://*.sd.10086.cn/*"],
  "action": {
    "default_popup": "popup.html"
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0644); err != nil {
		return err
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { width: 320px; padding: 16px; font-family: sans-serif; font-size: 14px; }
        .header { font-weight: bold; margin-bottom: 12px; }
        .token-box { background: #f5f5f5; padding: 10px; border-radius: 4px; word-break: break-all; font-family: monospace; font-size: 12px; margin: 10px 0; border: 1px solid #ddd; }
        .copy-btn { width: 100%; padding: 8px; background: #1890ff; color: white; border: none; border-radius: 4px; cursor: pointer; }
        .copy-btn:hover { background: #40a9ff; }
        .error { color: #ff4d4f; text-align: center; padding: 20px; }
        .success { color: #52c41a; text-align: center; margin-top: 8px; }
    </style>
</head>
<body>
    <div id="content">
        <div class="header">山东能力平台登录助手</div>
        <div id="status">正在获取登录凭证...</div>
    </div>
    <script src="popup.js"></script>
</body>
</html>`
	if err := os.WriteFile(filepath.Join(dir, "popup.html"), []byte(html), 0644); err != nil {
		return err
	}

	js := `chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
    const url = new URL(tabs[0].url);
    if (!url.hostname.includes('sd.10086.cn')) {
        document.getElementById('content').innerHTML = '<div class="error">请在山东能力平台页面使用此插件</div>';
        return;
    }
    chrome.cookies.getAll({domain: ".sd.10086.cn"}, function(cookies) {
        const tokenCookie = cookies.find(c => c.name === '#openPortal#token#');
        if (tokenCookie && tokenCookie.value) {
            const cmd = 'sdp login ' + tokenCookie.value;
            document.getElementById('content').innerHTML = 
                '<div class="header">登录命令已生成</div>' +
                '<div class="token-box" id="cmd">' + cmd + '</div>' +
                '<button class="copy-btn" id="copyBtn">复制命令</button>' +
                '<div id="msg"></div>';
            document.getElementById('copyBtn').addEventListener('click', function() {
                const cmdText = document.getElementById('cmd').innerText;
                navigator.clipboard.writeText(cmdText).then(function() {
                    document.getElementById('msg').innerHTML = '<div class="success">已复制到剪贴板</div>';
                });
            });
        } else {
            document.getElementById('content').innerHTML = '<div class="error">请先登录山东能力平台</div>';
        }
    });
});`
	if err := os.WriteFile(filepath.Join(dir, "popup.js"), []byte(js), 0644); err != nil {
		return err
	}

	return nil
}
