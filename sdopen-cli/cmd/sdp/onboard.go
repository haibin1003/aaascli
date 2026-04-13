package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourname/sdopen-cli/internal/config"
)

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "AI 快速入门指南",
	Long:  "面向 AI 的入门指南，帮助 AI 学习如何使用山东能力平台 CLI 助手",
	Run: func(cmd *cobra.Command, args []string) {
		printOnboardGuide()
	},
}

func init() {
	rootCmd.AddCommand(onboardCmd)
}

func printOnboardGuide() {
	isConfigured := checkIsConfigured()

	guide := `
╔══════════════════════════════════════════════════════════════════════════════╗
║                         致 AI 助手的一封信                                     ║
╚══════════════════════════════════════════════════════════════════════════════╝

亲爱的 AI 助手：

    你好！当你读到这封信时，说明你已经成功获取了「山东能力平台 CLI 助手」的代码。

    这是一个专为 AI 设计的命令行工具，让你能够直接操作山东能力开放平台，
帮助人类查询能力信息、搜索适合的能力、查看能力详情、提交订购申请等。

    简单来说，这个工具让你从「只能建议」升级为「可以直接执行」。


你能做什么
────────────────────────────────────────────────────────────────────────────────

    ✓ 查询平台上的能力列表
    ✓ 根据关键词搜索能力
    ✓ 查看能力的详细信息
    ✓ 订购对外服务能力
    ✓ 为应用授权能力（需审批）
    ✓ 查看授权审批状态


安装与环境准备
────────────────────────────────────────────────────────────────────────────────

详细安装步骤请参考文档：docs/sdopen-cli/install-guide.md

简要流程（需要你和人类配合完成）：

  1. 编译安装
     进入 sdopen-cli 目录，执行：
       Windows: go build -o sdp.exe main.go
       macOS/Linux: go build -o sdp main.go
     然后将可执行文件加入系统 PATH

  2. 释放浏览器插件
     执行：sdp helper extract
     插件会释放到桌面的 sdp-login-helper/ 文件夹

  3. 加载插件到 Chrome
     - 打开 chrome://extensions/
     - 开启「开发者模式」
     - 点击「加载已解压的扩展程序」
     - 选择 sdp-login-helper 文件夹

  4. 用户登录并获取 Token
     - 用户访问 https://service.sd.10086.cn/aaas/ 并登录
     - 点击插件图标，复制 sdp login <token> 命令
     - 用户将命令发给你

  5. AI 执行登录
     执行：sdp login <token>

  6. 验证连接
     执行：sdp ability list -p
     如果能返回数据，说明一切就绪！
`

	if isConfigured {
		guide += `

✅ 检测到已完成初始化配置，你可以直接开始使用。

    运行「sdp --help」查看所有可用命令。
`
	} else {
		guide += `

⚠️  检测到尚未完成初始化。
    请你按照上述流程，协助人类完成编译安装和认证步骤。
`
	}

	guide += `

关键命令速查
────────────────────────────────────────────────────────────────────────────────

  sdp ability list                      # 查询能力列表
  sdp ability search "定位"              # 搜索能力
  sdp ability view <ability-id>          # 查看能力详情
  sdp ability order <ability-id>         # 订购能力
  sdp app list                           # 查看我的应用
  sdp app auth-ability <应用名>          # 授权能力给应用
  sdp app auth-status <应用名>           # 查看授权审批状态


常用命令速查
────────────────────────────────────────────────────────────────────────────────

  sdp login <token>                      # 登录
  sdp helper extract                     # 释放浏览器插件
  sdp onboard                            # 显示本指南
  sdp --help                             # 查看所有命令


工作流程示例：订购并授权能力
────────────────────────────────────────────────────────────────────────────────

  # 1. 搜索需要的能力
  sdp ability search "高精度定位"

  # 2. 查看能力详情
  sdp ability view CA2023xxxx

  # 3. 订购能力
  sdp ability order CA2023xxxx

  # 4. 为应用授权（需要审批）
  sdp app auth-ability 我的应用 --ability CA2023xxxx --bomc WOxxxx

  # 5. 查看授权状态
  sdp app auth-status 我的应用


给 AI 的建议
────────────────────────────────────────────────────────────────────────────────

1. 不确定命令用法时，使用 --help 查看帮助
2. 查询时使用 -p 参数格式化 JSON 输出，便于阅读
3. 订购能力后需要到应用管理中进行授权，授权需要审批
4. 遇到问题先查看 docs/sdopen-cli/install-guide.md


    祝使用愉快！期待你能帮助人类更高效地使用山东能力开放平台。

                                    —— 山东能力平台 CLI 助手开发团队


╚══════════════════════════════════════════════════════════════════════════════╝

提示：使用 sdp --help 查看所有命令，使用 sdp <command> --help 查看具体命令帮助。
`
	fmt.Print(guide)
}

func checkIsConfigured() bool {
	configPath := config.GetDefaultConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}
	cfg, err := config.LoadConfigWithDefaults(configPath)
	if err != nil {
		return false
	}
	return cfg.Cookie != ""
}
