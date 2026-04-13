package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/common"
)

const (
	// DefaultSkillsPageSize 默认技能列表分页大小
	DefaultSkillsPageSize = 25
	// DefaultSkillsPageNo 默认技能列表页码
	DefaultSkillsPageNo = 1
)

var (
	skillsSearchKeyword string
	skillsListLimit     int
	skillsListPageNo    int
	installOutputDir    string
	describeType        = "5" // 默认插件类型
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "管理技能/插件，显示命令用法指导",
	Long: `管理技能/插件，包括插件市场操作和本地命令 AI 使用指导。

子命令:
  hub      插件市场操作（搜索、安装、查看插件）
  show     显示 lc 命令的 AI 使用指导

示例:
  # 插件市场操作
  lc skills hub search git
  lc skills hub install myplugin@1.0.0

  # 本地命令 AI 指导
  lc skills show req        # 显示 req 命令的 AI 使用指导
  lc skills show req list   # 显示 req list 子命令的 AI 使用指导`,
}

var skillsHubCmd = &cobra.Command{
	Use:   "hub",
	Short: "插件市场操作（搜索、安装、查看插件）",
	Long: `插件市场操作，包括搜索、列出、安装、查看插件详情。

子命令:
  search   搜索技能/插件
  list     列出所有技能/插件
  install  安装技能/插件
  describe 查看插件详情

示例:
  lc skills hub search git
  lc skills hub list
  lc skills hub install myplugin@1.0.0
  lc skills hub describe myplugin@1.0.0`,
}

var hubSearchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "搜索技能/插件",
	Long: `搜索可用的技能/插件，支持按关键词搜索。

提示:
  使用关键词搜索插件名称，例如: lc skills hub search "git"`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyword := ""
		if len(args) > 0 {
			keyword = args[0]
		}
		searchSkills(keyword)
	},
}

var hubListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有技能/插件",
	Long: `列出所有可用的技能/插件。

提示:
  使用 --limit 和 --page 参数控制分页`,
	Run: func(cmd *cobra.Command, args []string) {
		listSkills()
	},
}

var hubInstallCmd = &cobra.Command{
	Use:   "install [name@version]",
	Short: "安装技能/插件",
	Long: `安装指定版本的技能/插件。

参数格式:
  name@version     例如: lc skills hub install myplugin@1.0.1

提示:
  使用 'lc skills hub search <keyword>' 查找可用的插件
  使用 'lc skills hub describe name[@version]' 查看插件详情`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		installSkill(args[0])
	},
}

var hubDescribeCmd = &cobra.Command{
	Use:   "describe [name[@version]]",
	Short: "查看插件详情",
	Long: `查看指定插件的详细信息。

参数格式:
  name             列出该插件的所有版本
  name@version     查看指定版本的详细信息，例如: lc skills hub describe myplugin@1.0.0`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			common.PrintError(common.NewAutoDetectError("name is required").
				WithDetails("Usage: lc skills hub describe name[@version]").
				WithSuggestion("请提供插件名称，例如: lc skills hub describe myplugin"))
			os.Exit(1)
		}
		describeSkill(args[0])
	},
}

var skillsShowCmd = &cobra.Command{
	Use:   "show [command]",
	Short: "显示命令的 AI 使用指导",
	Long: `显示指定命令的 AI 使用指导，包含典型工作流、常见陷阱、参数获取方式等。

该命令自动提取命令的 --help 信息，并补充 AI 特定的使用提示。

示例:
  lc skills show req      # 显示 req 命令的 AI 使用指导
  lc skills show pr       # 显示 pr 命令的 AI 使用指导
  lc skills show req list # 显示 req list 子命令的 AI 使用指导`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		skillsShowCommand(args)
	},
}

func init() {
	rootCmd.AddCommand(skillsCmd)

	// 添加 hub 子命令
	skillsCmd.AddCommand(skillsHubCmd)
	skillsHubCmd.AddCommand(hubSearchCmd)
	skillsHubCmd.AddCommand(hubListCmd)
	skillsHubCmd.AddCommand(hubInstallCmd)
	skillsHubCmd.AddCommand(hubDescribeCmd)

	// 添加 show 子命令
	skillsCmd.AddCommand(skillsShowCmd)

	// Hub Search command flags
	hubSearchCmd.Flags().StringVar(&skillsSearchKeyword, "keyword", "", common.GetFlagDesc("keyword"))
	hubSearchCmd.Flags().IntVarP(&skillsListLimit, "limit", "l", DefaultSkillsPageSize, common.GetFlagDesc("limit"))
	hubSearchCmd.Flags().IntVarP(&skillsListPageNo, "page", "p", DefaultSkillsPageNo, common.GetFlagDesc("page"))

	// Hub List command flags
	hubListCmd.Flags().IntVarP(&skillsListLimit, "limit", "l", DefaultSkillsPageSize, common.GetFlagDesc("limit"))
	hubListCmd.Flags().IntVarP(&skillsListPageNo, "page", "p", DefaultSkillsPageNo, common.GetFlagDesc("page"))

	// Hub Describe command flags
	hubDescribeCmd.Flags().StringVar(&describeType, "type", "5", common.GetFlagDesc("skill-type"))

	// Hub Install command flags
	hubInstallCmd.Flags().StringVarP(&installOutputDir, "output", "o", "", "下载到"+common.GetFlagDesc("output")+"目录（默认当前目录）")
}

func searchSkills(keyword string) {
	// Use flag keyword if provided, otherwise use argument
	if skillsSearchKeyword != "" {
		keyword = skillsSearchKeyword
	}

	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":       true,
				"action":       "search",
				"keyword":      keyword,
				"pageNo":       skillsListPageNo,
				"limit":        skillsListLimit,
			}, nil
		}

		// Get headers for plugin API
		headers := ctx.Config.GetPlatformHeaders()

		// Create skills service
		skillsService := api.NewSkillsService(ctx.Config.API.BasePlatformURLMoss, headers, ctx.Client)

		// Send search request
		resp, err := skillsService.List(keyword, skillsListPageNo, skillsListLimit)
		if err != nil {
			return nil, err
		}

		ctx.Debug("Search skills response", zap.Any("response", resp))

		// Check if request was successful
		if !resp.Success {
			return nil, fmt.Errorf("搜索失败: %s", resp.Message)
		}

		// Return unified JSON structure
		return map[string]interface{}{
			"items": resp.Data,
			"pagination": map[string]interface{}{
				"page":       resp.PageNo,
				"pageSize":   skillsListLimit,
				"total":      resp.Count,
				"totalPages": resp.PageCount,
			},
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
	})
}

func listSkills() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun": true,
				"action": "list",
				"pageNo": skillsListPageNo,
				"limit":  skillsListLimit,
			}, nil
		}

		// Get headers for plugin API
		headers := ctx.Config.GetPlatformHeaders()

		// Create skills service
		skillsService := api.NewSkillsService(ctx.Config.API.BasePlatformURLMoss, headers, ctx.Client)

		// Send list request (empty keyword for all)
		resp, err := skillsService.List("", skillsListPageNo, skillsListLimit)
		if err != nil {
			return nil, err
		}

		ctx.Debug("List skills response", zap.Any("response", resp))

		// Check if request was successful
		if !resp.Success {
			return nil, fmt.Errorf("列表查询失败: %s", resp.Message)
		}

		// Return unified JSON structure
		return map[string]interface{}{
			"items": resp.Data,
			"pagination": map[string]interface{}{
				"page":       resp.PageNo,
				"pageSize":   skillsListLimit,
				"total":      resp.Count,
				"totalPages": resp.PageCount,
			},
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
	})
}

func installSkill(pluginSpec string) {
	// Parse name@version format
	parts := strings.Split(pluginSpec, "@")
	if len(parts) != 2 {
		common.PrintError(common.NewAutoDetectError("invalid format").
			WithDetails(fmt.Sprintf("Expected 'name@version', got '%s'", pluginSpec)).
			WithSuggestion("请使用正确的格式: lc skills hub install pluginName@version"))
		os.Exit(1)
	}

	name := parts[0]
	version := parts[1]

	if name == "" || version == "" {
		common.PrintError(common.NewAutoDetectError("name and version are required").
			WithSuggestion("请提供插件名称和版本，例如: lc skills hub install myplugin@1.0.0"))
		os.Exit(1)
	}

	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":    true,
				"action":    "install",
				"type":      describeType,
				"name":      name,
				"version":   version,
				"outputDir": installOutputDir,
			}, nil
		}

		// Get headers for plugin API
		headers := ctx.Config.GetPlatformHeaders()

		// Create skills service
		skillsService := api.NewSkillsService(ctx.Config.API.BasePlatformURLMoss, headers, ctx.Client)

		// Get plugin detail to find download URL
		resp, err := skillsService.Describe(describeType, name, version)
		if err != nil {
			return nil, err
		}

		ctx.Debug("Describe skill response", zap.Any("response", resp))

		// Check if request was successful
		if !resp.Success {
			return nil, fmt.Errorf("查询插件信息失败: %s", resp.Message)
		}

		// Check if plugin exists
		if resp.Data.PluginMarkBO == nil {
			return nil, fmt.Errorf("插件不存在")
		}

		plugin := resp.Data.PluginMarkBO
		if plugin.PackageID == "" {
			return nil, fmt.Errorf("插件没有 PackageID")
		}

		// Build download URL: /cmdevops-plugin/plugin/market/download/{packageId}
		downloadURL := fmt.Sprintf("%s/cmdevops-plugin/plugin/market/download/%s", ctx.Config.API.BasePlatformURLMoss, plugin.PackageID)

		// Download the file
		zipData, err := downloadFile(downloadURL, headers)
		if err != nil {
			return nil, fmt.Errorf("下载失败: %w", err)
		}

		// Determine output directory
		outputDir := installOutputDir
		if outputDir == "" {
			outputDir, err = os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("获取当前目录失败: %w", err)
			}
		}

		// Ensure output directory exists
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return nil, fmt.Errorf("创建输出目录失败: %w", err)
		}

		// Extract zip to current directory
		if err := extractZipToCurrentDir(zipData, outputDir); err != nil {
			return nil, fmt.Errorf("解压失败: %w", err)
		}

		// Return unified JSON structure
		return map[string]interface{}{
			"name":        plugin.Name,
			"version":     plugin.Version,
			"description": plugin.Description,
			"packageId":   plugin.PackageID,
			"fileSize":    plugin.FileSize,
			"outputDir":   outputDir,
		}, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
	})
}

// downloadFile downloads a file from URL with headers and returns its content
func downloadFile(fileURL string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// extractZipToCurrentDir extracts a zip file to the specified directory
func extractZipToCurrentDir(zipData []byte, destDir string) error {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		// Construct the full destination path
		destPath := filepath.Join(destDir, file.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("不安全的文件路径: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				return err
			}
			continue
		}

		// Create file's directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Create and write file
		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			destFile.Close()
			return err
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func describeSkill(pluginSpec string) {
	// Parse name@version or name format
	parts := strings.Split(pluginSpec, "@")
	name := parts[0]
	version := "" // Default to empty if not specified

	if len(parts) == 2 {
		version = parts[1]
	}

	if name == "" {
		common.PrintError(common.NewAutoDetectError("name is required").
			WithSuggestion("请提供插件名称，例如: lc skills hub describe myplugin"))
		os.Exit(1)
	}

	// Always use detail API
	describeSkillDetail(name, version)
}

// describeSkillDetail shows detailed information for a plugin
func describeSkillDetail(name, version string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		// Handle dry-run mode
		if ctx.DryRun {
			return map[string]interface{}{
				"dryRun":  true,
				"action":  "describe",
				"type":    describeType,
				"name":    name,
				"version": version,
			}, nil
		}

		// Get headers for plugin API
		headers := ctx.Config.GetPlatformHeaders()

		// Create skills service
		skillsService := api.NewSkillsService(ctx.Config.API.BasePlatformURLMoss, headers, ctx.Client)

		// Send describe request
		resp, err := skillsService.Describe(describeType, name, version)
		if err != nil {
			return nil, err
		}

		ctx.Debug("Describe skill response", zap.Any("response", resp))

		// Check if request was successful
		if !resp.Success {
			return nil, fmt.Errorf("查询失败: %s", resp.Message)
		}

		// Check if plugin exists
		if resp.Data.PluginMarkBO == nil {
			return map[string]interface{}{
				"exists": false,
				"name":   name,
				"version": version,
			}, nil
		}

		// Return unified JSON structure
		plugin := resp.Data.PluginMarkBO
		result := map[string]interface{}{
			"exists": true,
			"plugin": plugin,
		}
		if len(resp.Data.VersionList) > 0 {
			result["versions"] = resp.Data.VersionList
		}
		return result, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
	})
}

// skillsShowCommand 实现 skills show 功能
func skillsShowCommand(args []string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		commandPath := strings.Join(args, " ")

		// 查找对应的 cobra 命令
		targetCmd, targetArgs := findCommandForShow(rootCmd, args)
		if targetCmd == nil {
			return nil, fmt.Errorf("未找到命令: %s", commandPath)
		}

		// 构建 show 输出
		result := buildSkillsShowOutput(targetCmd, targetArgs, commandPath)
		return result, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		DryRun:    dryRunMode,
		Logger:    &logger,
	})
}

// findCommandForShow 根据参数路径查找 cobra 命令
func findCommandForShow(root *cobra.Command, args []string) (*cobra.Command, []string) {
	cmdArgs := make([]string, len(args))
	copy(cmdArgs, args)

	cmd, foundArgs, err := root.Find(cmdArgs)
	if err != nil || cmd == root {
		return nil, args
	}

	consumed := len(args) - len(foundArgs)
	remaining := []string{}
	if consumed < len(args) {
		remaining = args[consumed:]
	}

	return cmd, remaining
}

// buildSkillsShowOutput 构建 show 输出
func buildSkillsShowOutput(cmd *cobra.Command, args []string, path string) map[string]interface{} {
	helpText := getCommandHelpForShow(cmd)
	aiHints := getAIHintsForShow(path)

	return map[string]interface{}{
		"command":     path,
		"help":        helpText,
		"aiGuidance":  aiHints,
		"description": cmd.Long,
	}
}

// getCommandHelpForShow 获取命令的 help 文本
func getCommandHelpForShow(cmd *cobra.Command) string {
	if cmd == rootCmd {
		return ""
	}

	var parts []string

	if cmd.Use != "" {
		parts = append(parts, fmt.Sprintf("用法: %s", cmd.Use))
	}

	if cmd.Short != "" {
		parts = append(parts, fmt.Sprintf("简介: %s", cmd.Short))
	}

	longDesc := strings.TrimSpace(cmd.Long)
	if longDesc != "" && longDesc != cmd.Short {
		parts = append(parts, fmt.Sprintf("详细说明:\n%s", longDesc))
	}

	if cmd.HasAvailableFlags() {
		parts = append(parts, fmt.Sprintf("可用参数:\n%s", getFlagsHelpForShow(cmd)))
	}

	if cmd.HasAvailableSubCommands() {
		var subcmds []string
		for _, sub := range cmd.Commands() {
			if !sub.Hidden {
				subcmds = append(subcmds, fmt.Sprintf("  %s - %s", sub.Name(), sub.Short))
			}
		}
		if len(subcmds) > 0 {
			parts = append(parts, fmt.Sprintf("子命令:\n%s", strings.Join(subcmds, "\n")))
		}
	}

	return strings.Join(parts, "\n\n")
}

// getFlagsHelpForShow 获取命令的参数帮助
func getFlagsHelpForShow(cmd *cobra.Command) string {
	var flags []string

	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		flagLine := fmt.Sprintf("  --%s", f.Name)
		if f.Shorthand != "" {
			flagLine = fmt.Sprintf("  -%s, %s", f.Shorthand, flagLine)
		}
		flagLine += fmt.Sprintf(": %s", f.Usage)
		if f.DefValue != "" {
			flagLine += fmt.Sprintf(" (默认: %s)", f.DefValue)
		}
		flags = append(flags, flagLine)
	})

	return strings.Join(flags, "\n")
}

// getAIHintsForShow 获取 AI 使用提示
func getAIHintsForShow(path string) map[string]interface{} {
	if hints, ok := skillsShowData[path]; ok {
		return map[string]interface{}{
			"purpose":      hints.Purpose,
			"typicalFlow":  hints.TypicalFlow,
			"tips":         hints.Tips,
			"commonErrors": hints.CommonErrors,
		}
	}

	parts := strings.Split(path, " ")
	if len(parts) > 1 {
		parentPath := strings.Join(parts[:len(parts)-1], " ")
		if hints, ok := skillsShowData[parentPath]; ok {
			return map[string]interface{}{
				"purpose":    hints.Purpose,
				"parentTips": hints.Tips,
				"note":       "该子命令暂无特定提示，请参考父命令说明",
			}
		}
	}

	return map[string]interface{}{
		"note": "该命令暂无详细 AI 提示，请使用 --help 查看基本用法",
	}
}

// skillsShowData 存储 AI 使用提示
type skillsShowDataEntry struct {
	Purpose      string
	TypicalFlow  []string
	Tips         []string
	CommonErrors []string
}

var skillsShowData = map[string]skillsShowDataEntry{
	"req": {
		Purpose: "管理研发需求，包括创建、查询、查看、删除、更新需求",
		TypicalFlow: []string{
			"1. 列出需求: lc req list (在 Git 仓库目录下自动探测空间)",
			"2. 查看详情: lc req view <objectId> -w <workspace-key>",
			"3. 创建任务: lc task create <req-objectId> '任务名称' -w <workspace-key>",
			"4. 更新需求: lc req update <objectId> -w <workspace-key> --name '新名称'",
		},
		Tips: []string{
			"objectId 和 key 的区别: objectId 是内部唯一标识(如 AXi9LpGjsA)，key 是显示编号(如 XXJS-123)",
			"创建需求必须指定 --project-code，可通过 'lc space project linked -w <key>' 获取",
			"list 和 search 命令支持自动探测，在 Git 仓库目录下无需 -w 参数",
			"workspace-name 会自动从 workspace-key 获取，无需手动指定",
		},
		CommonErrors: []string{
			"错误: '关联项目字段必填' → 未指定 --project-code 或指定的项目代码无效",
			"错误: 'workspace-key is required' → 不在 Git 仓库目录且未指定 -w 参数",
		},
	},
	"req list": {
		Purpose: "查询研发空间中的需求列表",
		TypicalFlow: []string{
			"1. 自动探测: cd /path/to/repo && lc req list",
			"2. 手动指定: lc req list -w XXJSxiaobaice（workspace-name 自动获取）",
			"3. 分页查询: lc req list -l 50 -o 0 (限制50条，从第0条开始)",
		},
		Tips: []string{
			"在 Git 仓库目录下执行可自动探测 --workspace-key",
			"workspace-name 会自动从 workspace-key 获取，无需手动指定",
			"返回结果中的 objectId 用于后续操作（view/delete/update）",
			"返回结果中的 key 是需求编号（如 XXJSxiaobaice-123），可用于展示",
			"使用 -l 和 -o 参数控制分页，默认返回20条",
		},
		CommonErrors: []string{
			"错误: 自动探测失败 → 当前目录不在 Git 仓库中，需手动指定 -w",
		},
	},
	"req create": {
		Purpose: "创建新需求",
		TypicalFlow: []string{
			"1. 获取项目代码: lc space project linked -w <workspace-key>",
			"2. 创建需求: lc req create '需求名称' -w <key> --project-code <code>",
			"3. 或使用 YAML: lc req create -f requirement.yaml -w <key>",
		},
		Tips: []string{
			"简单创建只需名称和 project-code，系统会自动填充默认值",
			"YAML 创建更灵活，可指定负责人、优先级、业务背景等",
			"创建后会返回 objectId 和 key，保存这些值用于后续操作",
			"project-code 是项目代码（如 R24113J3C04），不是项目名称",
		},
		CommonErrors: []string{
			"错误: '关联项目字段必填' → 未指定 --project-code 或项目代码不正确",
			"错误: project-code 格式错误 → 需使用项目代码（如 R24113J3C04），不是数字ID",
		},
	},
	"pr": {
		Purpose: "管理代码合并请求（Merge Request），包括创建、审核、合并",
		TypicalFlow: []string{
			"1. 创建 MR: lc pr create -t '标题' -s feature -m master --git-project-id <id> -w <key>",
			"2. 查看 MR: lc pr list --git-project-id <id> -w <key>",
			"3. 审核 MR: lc pr review <iid> --git-project-id <id> -w <key> --type approve",
			"4. 合并 MR: lc pr merge <iid> --git-project-id <id> -w <key>",
		},
		Tips: []string{
			"git-project-id 是仓库的数字ID，不是项目名称",
			"mr-id 是 MR 的 iid（如 123），不是全局ID",
			"审核类型: approve(通过), reject(拒绝), request_changes(要求修改)",
			"合并前确保 MR 已审核通过，否则可能失败",
		},
		CommonErrors: []string{
			"错误: 'Not Found' → git-project-id 错误或该仓库不存在",
			"错误: 'Conflict' → MR 有冲突，需要先解决冲突",
		},
	},
	"task": {
		Purpose: "管理任务，包括创建、查询、删除",
		TypicalFlow: []string{
			"1. 在需求下创建任务: lc task create <req-objectId> '任务名称' -w <key>",
			"2. 查询任务列表: lc task list -w <key>（workspace-name 自动获取）",
			"3. 删除任务: lc task delete <task-id> -w <key>",
		},
		Tips: []string{
			"创建任务需要父需求的 objectId，不是 key",
			"workspace-name 会自动从 workspace-key 获取，无需手动指定",
		},
		CommonErrors: []string{
			"错误: 'req-objectId is required' → 创建任务时必须指定父需求ID",
		},
	},
	"bug": {
		Purpose: "管理缺陷（Bug），包括创建、查询、更新状态",
		TypicalFlow: []string{
			"1. 创建缺陷: lc bug create -t '标题' -D '描述' -p <project-code> -w <key> --template-simple",
			"2. 查询缺陷: lc bug list -w <key> (支持自动探测)",
			"3. 更新状态: lc bug update-status <id> <status-id> -w <key>",
		},
		Tips: []string{
			"推荐使用 --template-simple 使用简洁模板，避免填写大量字段",
			"list 命令支持自动探测，在 Git 仓库目录下无需 -w 参数",
			"project-id 是 -p 参数，不是 --project-code（注意命令差异）",
		},
		CommonErrors: []string{
			"错误: 创建失败 → 检查是否使用了正确的模板参数",
		},
	},
	"space": {
		Purpose: "管理研发空间，查询空间列表和项目关联",
		TypicalFlow: []string{
			"1. 查询空间列表: lc space list",
			"2. 查询可关联项目: lc space project list -w <key>",
			"3. 查询已关联项目: lc space project linked -w <key>",
		},
		Tips: []string{
			"space list 不需要参数，返回所有有权限的空间",
			"project list 支持自动探测，在 Git 仓库目录下无需 -w",
			"project linked 用于获取创建需求所需的 project-code",
		},
		CommonErrors: []string{},
	},
	"repo": {
		Purpose: "管理代码仓库，包括创建、查询、删除",
		TypicalFlow: []string{
			"1. 查询仓库: lc repo list -w <key>",
			"2. 创建仓库: lc repo create '名称' -w <key> --description '描述'",
			"3. 关闭关联工作项: lc repo disable-work-item-link <git-project-id> -w <key>",
		},
		Tips: []string{
			"repo list 不显示 git-project-id，需要详细信息请使用 -d 参数",
			"创建仓库后需要到网页端配置代码提交关联工作项",
		},
		CommonErrors: []string{
			"错误: '仓库名已存在' → 该空间下已有同名仓库",
		},
	},
}
