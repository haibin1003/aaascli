package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourname/sdopen-cli/internal/config"
)

// loginCmd 登录命令
var loginCmd = &cobra.Command{
	Use:   "login [token]",
	Short: "登录设置会话 Cookie",
	Long: `登录山东能力开放平台。

推荐方式（自动从浏览器提取 Cookie）:
  1. sdp helper extract    # 释放浏览器插件到桌面
  2. Chrome 浏览器 -> 设置 -> 扩展程序 -> 开启开发者模式 -> 加载已解压扩展程序
  3. 登录平台后点击插件图标
  4. 复制 sdp login xxx 命令并执行

手动方式（不推荐）:
  sdp login <token>
  注意：token 是 #openPortal#token# 的值`,
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

// setLoginCookie 设置登录 Cookie
func setLoginCookie(cookieValue string) {
	// 清理 cookie 值
	cookieValue = strings.TrimSpace(cookieValue)
	cookieValue = strings.TrimPrefix(cookieValue, "#openPortal#token#=")
	cookieValue = strings.TrimPrefix(cookieValue, "token=")
	cookieValue = strings.Trim(cookieValue, `"'`)

	if cookieValue == "" {
		fmt.Fprintln(os.Stderr, "错误: token 值不能为空")
		os.Exit(1)
	}

	// 确保配置目录存在
	if err := config.EnsureConfigDir(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 创建配置目录失败: %v\n", err)
		os.Exit(1)
	}

	// 保存配置
	cfg := &config.Config{
		Cookie: cookieValue,
	}

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 保存配置失败: %v\n", err)
		os.Exit(1)
	}

	configPath := config.GetDefaultConfigPath()
	fmt.Println("登录成功")
	fmt.Printf("配置文件: %s\n", configPath)
	fmt.Println("提示: 使用 'sdp ability list' 查看能力列表")
}

// showLoginStatus 显示登录状态
func showLoginStatus() {
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil || cfg.Cookie == "" {
		fmt.Println("未登录")
		fmt.Println("请使用浏览器插件获取 token，然后执行: sdp login <token>")
		os.Exit(1)
	}

	fmt.Println("已登录")
	fmt.Printf("配置文件: %s\n", configPath)
}
