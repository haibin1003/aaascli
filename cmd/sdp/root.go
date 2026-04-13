package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/haibin1003/aaascli/internal/common"
)

var (
	// 全局标志
	debugMode   bool
	insecure    bool
	dryRun      bool
	prettyPrint bool
	cookieFlag  string

	// 版本信息
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

// SetVersionInfo 设置版本信息
func SetVersionInfo(v, c, bt string) {
	version = v
	commit = c
	buildTime = bt
	common.SetVersion(v)
}

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "sdp",
	Short: "山东能力平台 CLI 助手",
	Long: `山东能力平台 CLI 助手 - 让 AI 可以操作山东能力开放平台

支持能力查询、订购、授权等功能。

快速开始:
  1. 安装浏览器插件: sdp helper extract
  2. 登录平台后点击插件图标，复制 sdp login <token> 命令
  3. 执行登录命令完成认证
  4. 开始使用: sdp ability list
`,
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "启用调试模式")
	rootCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "k", true, "跳过 TLS 证书验证")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "模拟执行，不实际提交")
	rootCmd.PersistentFlags().BoolVarP(&prettyPrint, "pretty", "p", false, "格式化 JSON 输出")
	rootCmd.PersistentFlags().StringVarP(&cookieFlag, "cookie", "c", "", "手动指定认证 cookie")

	// 添加子命令
	rootCmd.AddCommand(versionCmd)
}

// versionCmd 版本命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("版本: %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("构建时间: %s\n", buildTime)
	},
}
