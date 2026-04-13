package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/common"
	"github.com/user/lc/internal/embed"
)

var (
	helperOutputDir string
	helperListOnly  bool
)

var helperCmd = &cobra.Command{
	Use:   "helper",
	Short: "安装浏览器(Chrome)插件，一键复制login认证命令",
	Long: `将浏览器登录辅助插件释放，安装浏览器(Chrome)插件，一键复制login认证命令。

默认释放到桌面，可用于浏览器开发者模式加载。`,
}

var helperExtractCmd = &cobra.Command{
	Use:   "extract [output-dir]",
	Short: "释放浏览器扩展到指定目录",
	Long: `将嵌入的浏览器登录辅助扩展释放到指定目录。

如果未指定输出目录，默认释放到桌面。
释放后的扩展可以通过浏览器的"开发者模式"-"加载已解压的扩展"来使用。`,
	Example: `  # 释放到桌面（默认）
  lc helper extract

  # 释放到指定目录
  lc helper extract /path/to/output

  # 使用 --output 指定目录
  lc helper extract --output /path/to/output`,
	RunE: runHelperExtract,
}

var helperStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "检查扩展打包状态",
	Long:  `显示嵌入的浏览器扩展是否已打包以及文件大小。`,
	Run: func(cmd *cobra.Command, args []string) {
		common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
			size := embed.GetHelperExtensionSize()
			status := "packed"
			if size < 1024 {
				status = "placeholder"
			}
			return map[string]interface{}{
				"status": status,
				"size":   size,
				"sizeKB": float64(size) / 1024,
			}, nil
		}, common.ExecuteOptions{
			DebugMode: debugMode,
			Insecure:  insecureSkipVerify,
			Logger:    &logger,
		})
	},
}

func init() {
	// 添加子命令
	helperCmd.AddCommand(helperExtractCmd)
	helperCmd.AddCommand(helperStatusCmd)

	// 添加标志
	helperExtractCmd.Flags().StringVarP(&helperOutputDir, "output", "o", "", common.GetFlagDesc("output")+"目录（默认：桌面）")

	// 添加到 root
	rootCmd.AddCommand(helperCmd)
}

func runHelperExtract(cmd *cobra.Command, args []string) error {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// 确定输出目录
		outputDir := helperOutputDir
		if outputDir == "" && len(args) > 0 {
			outputDir = args[0]
		}

		// 如果未指定，使用桌面
		if outputDir == "" {
			outputDir = getDesktopPath()
		}

		// 确保是绝对路径
		absPath, err := filepath.Abs(outputDir)
		if err != nil {
			return nil, fmt.Errorf("无法解析路径: %w", err)
		}
		extensionDir := filepath.Join(absPath, "lc-login-helper-extension")

		// 检查目录是否已存在，如果存在则删除
		if _, err := os.Stat(extensionDir); err == nil {
			// 目录已存在，删除旧目录
			if err := os.RemoveAll(extensionDir); err != nil {
				return nil, fmt.Errorf("删除旧目录失败: %w", err)
			}
		}

		// 执行解压
		if err := embed.ExtractHelperExtension(outputDir); err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"extensionDir": extensionDir,
			"message":      "Extension extracted successfully. Load it in Chrome/Edge developer mode.",
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		Logger:    &logger,
	})
	return nil
}

// getDesktopPath 获取桌面路径
func getDesktopPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	// 尝试常见的桌面路径
	candidates := []string{
		filepath.Join(homeDir, "Desktop"),
		filepath.Join(homeDir, "桌面"),
		filepath.Join(homeDir, "Escritorio"), // Spanish
		filepath.Join(homeDir, "Bureau"),     // French
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 默认返回 home 目录
	return homeDir
}
