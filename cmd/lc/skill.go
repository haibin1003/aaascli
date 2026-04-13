package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/embed"
)

var (
	skillOutputDir string
)

var skillCmd = &cobra.Command{
	Use:   "skills",
	Short: "技能管理",
	Long:  `管理 AI 技能，包括安装、列表等操作。`,
}

var skillInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "安装技能到本地",
	Long: `将内置技能安装到本地目录。

默认安装到 ~/.config/joinai-code/skills (Linux/Mac)
或 %APPDATA%/joinai-code/skills (Windows)

安装后可以在 AI 对话中使用这些技能。`,
	Example: `  # 安装到默认目录
  lc skills install

  # 安装到指定目录
  lc skills install --path /path/to/skills`,
	RunE: runSkillInstall,
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出已安装的技能",
	Long:  `列出本地已安装的技能。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 列出已安装的技能
		skillsDir := embed.DefaultSkillsDir()
		if skillsDir == "" {
			fmt.Fprintln(os.Stderr, "无法确定技能目录")
			os.Exit(1)
		}

		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "技能目录不存在，请先运行 lc skills install")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "读取技能目录失败: %v\n", err)
			os.Exit(1)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				fmt.Println(entry.Name())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(skillCmd)
	skillCmd.AddCommand(skillInstallCmd)
	skillCmd.AddCommand(skillListCmd)

	skillInstallCmd.Flags().StringVarP(&skillOutputDir, "path", "p", "", "安装目录（默认：~/.config/joinai-code/skills）")
}

func runSkillInstall(cmd *cobra.Command, args []string) error {
	targetDir := skillOutputDir
	if targetDir == "" && len(args) > 0 {
		targetDir = args[0]
	}

	// 如果未指定，使用默认目录
	if targetDir == "" {
		targetDir = embed.DefaultSkillsDir()
	}

	if targetDir == "" {
		fmt.Fprintln(os.Stderr, "无法确定默认技能目录")
		os.Exit(1)
	}

	// 执行安装
	if err := embed.ExtractSkills(targetDir); err != nil {
		fmt.Fprintf(os.Stderr, "安装失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("技能已安装到: %s\n", targetDir)
	return nil
}