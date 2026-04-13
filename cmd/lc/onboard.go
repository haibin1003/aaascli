package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/config"
)

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "AI 快速入门指南 - 学习如何使用本助手",
	Long: `面向 AI 的入门指南，帮助 AI 学习如何使用灵畿 CLI 助手。

此命令会检测当前环境状态，指导 AI 协助人类完成必要的初始化步骤。`,
	Run: func(cmd *cobra.Command, args []string) {
		printOnboardGuide()
	},
}

func init() {
	rootCmd.AddCommand(onboardCmd)
}

func printOnboardGuide() {
	// 检查是否已经配置（config 文件存在且包含 cookie）
	isConfigured := checkIsConfigured()

	guide := `
╔══════════════════════════════════════════════════════════════════════════════╗
║                         致 AI 助手的一封信                                     ║
╚══════════════════════════════════════════════════════════════════════════════╝

亲爱的 AI 助手：

    你好！当你读到这封信时，说明你已经成功安装了「灵畿 CLI 助手」。

    这是一个专为 AI 设计的命令行工具，让你能够直接操作灵畿研发平台，帮助人类完成代码仓库管理、需求跟踪、任务分解、
缺陷管理等研发工作。

    简单来说，这个工具让你从「只能建议」升级为「可以直接执行」。


📋 你能做什么
────────────────────────────────────────────────────────────────────────────────

    ✓ 创建和管理代码仓库
    ✓ 提交和审核合并请求（MR）
    ✓ 创建需求并分解为任务
    ✓ 提交和跟踪缺陷（Bug）
    ✓ 查询项目、空间、成员等信息


🔒 安全提示：默认只读模式
────────────────────────────────────────────────────────────────────────────────

    ⚠️  CLI 默认处于只读模式，可执行查询，禁止写入（create/update/delete/merge）。

    需写入时：
      1. 询问人类是否需要关闭只读模式
      2. 执行：lc readonly off --duration 30m  （推荐临时关闭，30分钟后自动恢复）
      3. 提醒可随时恢复：lc readonly on


🚀 开始之前
────────────────────────────────────────────────────────────────────────────────
`

	if isConfigured {
		guide += `
✅ 太好了！检测到已完成初始化配置，你可以直接开始使用。

    运行「lc --help」查看所有可用命令，
    或继续阅读本指南了解详细信息。


🔍 第一步：了解当前环境
────────────────────────────────────────────────────────────────────────────────

  在执行任何操作前，先运行以下命令获取当前环境信息：

    lc detect

  返回的 JSON 中包含后续命令所需的关键参数：
    • data.workspaceKey   → 研发空间 key（-w 参数）
    • data.repository.gitProjectId   → 仓库 ID（--git-project-id 参数）

  如果不在 Git 仓库目录下，data.matched 为 false，
  需通过 lc space list 手动查找 workspace-key。
`
	} else {
		guide += `
⚠️  检测到尚未完成初始化。在正式使用前，
    请你协助人类完成以下两个步骤：


【步骤一：安装 Chrome 扩展】
────────────────────────────────────────────────────────────────────────────────

这个扩展是连接浏览器和 CLI 工具的桥梁，用于获取登录凭证。

请你指导人类完成以下操作：

1. 释放扩展安装包
   执行lc helper extract命令，
   扩展将默认释放到 用户桌面上的 lc-login-helper-extension/ 文件夹。
   

2. 加载扩展到 Chrome
   (1) 打开 Chrome 浏览器，输入 chrome://extensions/ 进入扩展管理页
   (2) 开启右上角的「开发者模式」开关
   (3) 点击「加载已解压的扩展程序」按钮
   (4) 选择 lc-login-helper-extension 文件夹
   (5) 扩展图标将出现在浏览器工具栏
   (6) 建议需要点击拼图图标,并将 “灵畿 CLI 助手” 固定

3. 验证安装
   登录灵畿平台后，点击扩展图标，应能看到安装/认证命令。


【步骤二：完成认证】
────────────────────────────────────────────────────────────────────────────────

请你协助人类完成认证：

1. 请人类登录灵畿平台网页版
2. 点击 Chrome 扩展图标，或者点击 灵畿 CLI 助手 图标
3. 复制显示的认证命令（格式：lc login xxx）
4. 将该命令发送给AI 或者 在终端执行该命令

完成后，人类就可以通过你使用灵畿 CLI 助手了！
`
	}

	guide += `


🎯 AI 使用指南
────────────────────────────────────────────────────────────────────────────────

【如何查询命令用法】
────────────────────────────────────────────────────────────────────────────────

当不确定某个命令的用法时，使用以下方式获取帮助：

  1. 基础帮助
     lc <command> --help
     例如: lc req create --help

  2. AI 详细指导（推荐）
     lc skills show <command>
     例如:
       lc skills show req         # 查看 req 命令的 AI 使用指导
       lc skills show req create  # 查看 req create 的详细指导
       lc skills show pr          # 查看 pr 命令的 AI 使用指导

     这会返回包含以下内容的信息：
       - 命令用途
       - 典型工作流
       - 使用提示
       - 常见错误及解决方案


📚 关键参数获取方法
────────────────────────────────────────────────────────────────────────────────

  lc detect                              # 一次性获取 workspaceKey 和 gitProjectId（推荐）
  lc space list                          # workspace-key（spaceCode 字段）
  lc repo list -w <workspace-key>        # git-project-id（gitProjectId 字段）
  lc req list -w <workspace-key>         # object-id（objectId 字段）
  lc pr list --git-project-id <git-project-id>  # mr-id（iid 字段）


🔧 常用命令速查
────────────────────────────────────────────────────────────────────────────────

研发空间:  lc space list                              代码仓库:  lc repo list -w <workspace-key>
合并请求:  lc pr list --git-project-id <git-project-id>    需求管理:  lc req list -w <workspace-key>
任务管理:  lc task list -w <workspace-key>               缺陷管理:  lc bug list -w <workspace-key>

只读模式:  lc readonly                        查看状态
          lc readonly off --duration 30m    临时关闭 30 分钟（推荐）
          lc readonly on                     开启只读模式


💡 给 AI 的建议
────────────────────────────────────────────────────────────────────────────────

1. 执行写入操作前，先运行 lc readonly 检查模式
     • 如 readonly: true，询问人类是否需要关闭
     • 推荐临时关闭：lc readonly off --duration 30m
     • 提醒人类随时可恢复：lc readonly on

2. 不确定命令用法时，优先使用 lc skills show <command> 查询

3. 记录常用参数（workspace-key、project-code），避免重复询问

4. 在 Git 仓库目录下，命令会自动探测 workspace-key，可用 lc detect 验证探测结果

5. 查看所有可用命令：lc --help


🎓 形成 Skill 的建议
────────────────────────────────────────────────────────────────────────────────

当你熟悉以上操作后，可以将常用流程封装为 Skill。

使用 lc skills show <command> 查询命令用法时，
返回的 typicalFlow 字段就是很好的 Skill 步骤参考。

你也可以将常用的 workspace-key、project-code 等参数
记录为 Skill 的默认参数，提高效率。


📖 更多资源
────────────────────────────────────────────────────────────────────────────────

  • 查看所有命令：lc --help
  • 查看命令 AI 指导：lc skills show <command>
  • 项目 README：README.md
  • Chrome 扩展：lc helper extract


    祝使用愉快！期待你能帮助人类更高效地完成研发工作。

                                    —— 灵畿 CLI 助手开发团队


╚══════════════════════════════════════════════════════════════════════════════╝

提示：使用 lc --help 查看所有命令，使用 lc <command> --help 查看具体命令帮助。
`
	fmt.Print(guide)
}

// checkIsConfigured 检查是否已完成配置（config 文件存在且包含有效 cookie）
func checkIsConfigured() bool {
	configPath := config.GetDefaultConfigPath()

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}

	// 尝试加载配置
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil {
		return false
	}

	// 检查是否有 cookie
	return cfg.Cookie != ""
}
