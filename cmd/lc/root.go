package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/common"
	"go.uber.org/zap"
)

var (
	// 全局标志
	insecureSkipVerify bool
	debugMode          bool
	dryRunMode         bool
	prettyMode         bool
	cookieFlag         string
	logger             *zap.Logger

	// 版本信息，由 main.go 注入
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

var rootCmd = &cobra.Command{
	Use:   "lc",
	Short: "灵畿 CLI 助手 - 让AI可以操作灵畿平台",
	Long: `灵畿 CLI 助手是一个命令行工具，为AI提供管理代码仓库、需求、任务等能力。
支持代码仓库管理、Merge Request 操作、需求管理、任务管理、缺陷管理等功能。
详细文档：https://my.feishu.cn/wiki/VHDtwio84iyWnJkjOwKcC38fnFd
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 添加全局持久化标志
	rootCmd.PersistentFlags().BoolVarP(&insecureSkipVerify, "insecure", "k", true, common.GetFlagDesc("insecure"))
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, common.GetFlagDesc("debug"))
	rootCmd.PersistentFlags().BoolVar(&dryRunMode, "dry-run", false, common.GetFlagDesc("dry-run"))
	rootCmd.PersistentFlags().BoolVar(&prettyMode, "pretty", false, common.GetFlagDesc("pretty"))
	rootCmd.PersistentFlags().StringVarP(&cookieFlag, "cookie", "c", "", common.GetFlagDesc("cookie"))

	// 添加 version 命令
	rootCmd.AddCommand(versionCmd)
}

// versionCmd 显示版本信息
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  `显示版本、Git 提交哈希和构建时间。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("CLI 版本: %s\n", version)
		fmt.Printf("Git 版本:    %s\n", commit)
		fmt.Printf("构建时间:    %s\n", buildTime)
	},
}

// initLogger 根据 debug 模式初始化 logger
// 返回 logger 实例和清理函数（用于 defer 调用）
func initLogger(debug bool) (*zap.Logger, func(), error) {
	if debug {
		l, err := zap.NewDevelopment()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create development logger: %w", err)
		}
		return l, func() { _ = l.Sync() }, nil
	}
	return nil, func() {}, nil
}
