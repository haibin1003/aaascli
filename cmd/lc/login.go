package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录设置会话 Cookie",
	Long: `登录灵畿平台。

推荐方式（自动从浏览器提取 Cookie）：
  1. lc helper extract    # 从浏览器提取 Cookie 并自动登录  
  2. chrome 浏览器 设置 扩展程序-打开开发者模式-加载已解压扩展程序
  3. 登录灵畿平台，点击灵畿CLI登录助手
  4. 复制 lc login xxx 命令，并执行
	
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			showLoginStatus()
		} else {
			setLoginCookie(args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func setLoginCookie(cookieValue string) {
	// Clean up the cookie value
	cookieValue = strings.TrimSpace(cookieValue)

	// Remove MOSS_SESSION= prefix if present
	cookieValue = strings.TrimPrefix(cookieValue, "MOSS_SESSION=")

	// Remove any surrounding quotes
	cookieValue = strings.Trim(cookieValue, `"'`)

	if cookieValue == "" {
		fmt.Fprintln(os.Stderr, "❌ 错误: cookie 值不能为空")
		os.Exit(1)
	}

	// Load existing config or create new one
	cfg := config.NewConfig()
	configPath := config.GetDefaultConfigPath()

	if loadedCfg, err := config.LoadConfigWithDefaults(configPath); err == nil {
		cfg = loadedCfg
	}

	// Set the cookie
	cfg.Cookie = "MOSS_SESSION=" + cookieValue
	if cfg.API.Headers == nil {
		cfg.API.Headers = make(map[string]string)
	}
	cfg.API.Headers["Cookie"] = cfg.Cookie

	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 错误: 创建配置目录失败: %v\n", err)
		os.Exit(1)
	}

	// Save only cookie to config file
	simpleConfig := map[string]string{
		"cookie": cfg.Cookie,
	}
	data, err := json.MarshalIndent(simpleConfig, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 错误: 序列化配置失败: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 错误: 写入配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 登录成功\n")
	fmt.Printf("配置文件: %s\n", configPath)
}

func showLoginStatus() {
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "未登录\n")
		os.Exit(1)
	}

	if cfg.Cookie == "" {
		fmt.Fprintln(os.Stderr, "未登录")
		os.Exit(1)
	}

	// 登录正常，不输出任何内容
}
