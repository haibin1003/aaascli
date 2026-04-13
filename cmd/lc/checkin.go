package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/user/lc/internal/config"
)

var checkinCmd = &cobra.Command{
	Use:   "checkin",
	Short: "签到/保持认证有效",
	Long: `调用 user-info 接口检查认证状态，保持会话有效。

建议 AI 每隔 30 分钟执行一次此命令，以确保认证不会过期。

示例:
  lc checkin`,
	Run: func(cmd *cobra.Command, args []string) {
		checkinExec()
	},
}

func init() {
	rootCmd.AddCommand(checkinCmd)
}

func checkinExec() {
	// FetchCurrentUser 会调用 /v1/self/user-info 接口
	// 如果认证有效，会返回用户信息；如果认证失效，会返回错误
	configPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "❌ 签到失败: 未找到配置文件")
		fmt.Fprintln(os.Stderr, "请先运行: lc login <cookie-value> 进行登录")
		os.Exit(1)
	}

	if err := cfg.FetchCurrentUser(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 签到失败: %v\n", err)
		fmt.Fprintln(os.Stderr, "认证可能已过期，请重新登录: lc login <cookie-value>")
		os.Exit(1)
	}

	fmt.Println("✅ 签到成功，认证有效")
	fmt.Printf("时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
