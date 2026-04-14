package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/haibin1003/aaascli/internal/config"
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
================================================================================
                    山东能力平台 CLI 助手 - AI 快速入门指南
================================================================================

亲爱的 AI 助手：

    你好！欢迎安装「山东能力平台 CLI 助手」（命令名：sdp）。

    这个工具是专为 AI 设计的命令行客户端，让你可以直接调用山东能力开放
平台的真实接口，帮助人类完成能力查询、服务检索、接口详情查看等工作。

    你的角色从「只能给建议」升级为「可以动手查、动手整理、甚至生成代码」。


一、这个工具是干什么的
--------------------------------------------------------------------------------

山东能力开放平台（https://service.sd.10086.cn/aaas/）是中国移动山东公司
对外提供数字化能力的统一门户，汇集了 300+ 业务能力（如短信触达、定位、
机器视觉、大数据等）和数千个具体 API 服务。

这个 CLI 工具让你无需打开浏览器，就能通过命令行直接查询平台上的：
- 全部能力目录（325+ 个）
- 能力下的服务菜单
- 全量数字服务/API 目录（3800+ 个）
- 单个服务的接口定义、请求示例、响应示例、负责人信息
- 用户的应用列表

二、核心功能清单
--------------------------------------------------------------------------------

【能力管理】
  sdp ability list                      查询全部能力列表（含分页）
  sdp ability search <keyword>          按关键词搜索能力
  sdp ability view <ability-id>         查看能力详情（描述、提供方、分类等）
  sdp ability services <ability-id>     查看能力下挂载的服务菜单
  sdp ability order <ability-id>        订购能力（未实现，需用户手动在网页完成）
  sdp ability my                        查看我的已订购能力

【数字服务管理】
  sdp service list                      查询全量服务目录（目录+API 混合树）
  sdp service search <keyword>          搜索具体 API 服务
  sdp service view <service-id>         查看服务详情（URL、示例、负责人等）

【应用管理】
  sdp app list                          查看用户已创建的应用列表

【认证与辅助】
  sdp login <token>                     设置登录凭证
  sdp helper extract [dir]              释放 Chrome 登录插件
  sdp onboard                           显示本指南
  sdp --help                            查看所有命令

【知识库】
  sdp knowledge list                    查看内置知识文档列表
  sdp knowledge view <name>             阅读指定知识文档
  sdp knowledge search <keyword>        在知识库中搜索关键词

三、典型使用场景
--------------------------------------------------------------------------------

场景 1：用户想做某个业务，但不知道选什么能力
    → 你查询平台能力列表，根据分类和描述推荐能力组合，输出解决方案。

场景 2：用户需要对接某个 API，想要示例代码
    → 你查询服务详情，根据 requestUrl、requestExample、responseExample
       生成 Java / Python / Go / Shell 等调用代码。

场景 3：用户想整理一份内部能力清单
    → 你批量查询相关能力/服务，整理成 Markdown 表格或接口文档。

场景 4：用户做技术预研，需要能力对比分析
    → 你提取多个能力的描述、提供方、适用场景，输出对比报告。

四、安装与登录流程（需要你与用户配合）
--------------------------------------------------------------------------------

步骤 1：安装 CLI
    根据你的系统架构，从安装包中选择对应的二进制文件：
    - Windows x64     → sdp-windows-x64.exe
    - Windows ARM64   → sdp-windows-arm64.exe
    - Linux x64       → sdp-linux-x64
    - Linux ARM64     → sdp-linux-arm64
    - macOS Intel     → sdp-darwin-x64
    - macOS Apple Si  → sdp-darwin-arm64

    将选中的文件重命名为 sdp（Windows 为 sdp.exe），放入系统 PATH。

步骤 2：释放浏览器插件
    执行：sdp helper extract [输出目录]
    插件会释放到指定目录，默认是桌面上的 sdp-login-helper/ 文件夹。

步骤 3：用户导入 Chrome 插件
    请把以下步骤原样告诉用户：
    1. 打开 Chrome，地址栏输入 chrome://extensions/
    2. 开启右上角「开发者模式」
    3. 点击「加载已解压的扩展程序」
    4. 选择你释放的 sdp-login-helper 文件夹
    5. 插件名称应显示为「山东能力平台助手」

步骤 4：用户登录并获取 Token
    1. 用户访问 https://service.sd.10086.cn/aaas/ 并完成登录
    2. 点击浏览器右上角的插件图标
    3. 插件会显示一行命令，例如：
       sdp login a500d28... --verification 5f3004... --service 599483...
    4. 用户将这行命令复制发给你

步骤 5：你执行登录
    收到命令后直接运行即可。凭证会保存在 ~/.sdp/config.json 中。

步骤 6：验证连接
    运行：sdp ability list --size 3
    如果能返回 JSON 数据（包含能力名称、ID 等），说明一切就绪。
`

	if isConfigured {
		guide += `
[OK] 检测到已完成初始化配置，你可以直接开始使用。

    运行「sdp --help」查看所有可用命令。
`
	} else {
		guide += `
[!] 检测到尚未完成初始化。
    请你按照上述流程，先协助人类完成插件导入和认证登录。
`
	}

	guide += `
五、关键命令速查
--------------------------------------------------------------------------------

  sdp ability list --size 20              查询能力列表
  sdp ability search "短信" --size 10      搜索能力
  sdp ability view <ability-id>            查看能力详情
  sdp ability services <ability-id>        查看能力下的服务
  sdp service list --size 20               查询全量服务目录
  sdp service search <keyword> --size 10   搜索 API 服务
  sdp service view <service-id>            查看服务详情
  sdp app list --size 10                   查看我的应用
  sdp knowledge list                       查看知识文档
  sdp knowledge search <keyword>           搜索知识库
  sdp login <token>                        登录
  sdp helper extract [dir]                 释放浏览器插件

六、给 AI 的增值建议
--------------------------------------------------------------------------------

你拿到平台数据后，不要只做「复读机」，可以主动提供更高价值的输出：

1. 输出解决方案建议
   当用户描述业务场景时，查询相关能力列表，结合分类、描述、提供方，
   输出「核心能力 + 辅助能力 + 业务流程」的完整方案。

2. 生成调用代码
   当用户需要对接 API 时，根据 service view 返回的 requestTypeText、
   requestUrl、requestExample、responseExample，生成可直接运行的代码
   （Java / Python / Go / JavaScript / Shell 等），并补充错误处理。

3. 整理对比表格与文档
   帮用户把查询结果整理成 Markdown 表格、接口文档草案、项目 README
   中的能力依赖说明等。

4. 解释与翻译
   如果返回的字段名不够直观，主动解释其业务含义；如果用户有技术背景
   差异，用通俗易懂的语言翻译平台返回的内容。

七、故障排查
--------------------------------------------------------------------------------

  提示 "请先登录"
    → ~/.sdp/config.json 不存在或已过期，让用户重新点击插件复制 login 命令。

  ability services 返回空数组
    → 该能力下确实没有挂载服务，或用户尚未在网页端订购该能力。

  app list 返回空数组
    → 该账号尚未创建任何应用。

  中文显示乱码（Windows）
    → 工具已内置 GBK→UTF-8 自动转码。如仍异常，建议换用 Windows Terminal。

  网络连接失败 / TLS 错误
    → 平台使用自签名证书，工具默认跳过验证。如仍报错，加 --insecure 参数。

八、文件位置速查
--------------------------------------------------------------------------------

  登录凭证    ~/.sdp/config.json
  浏览器插件   由 sdp helper extract [dir] 指定
  二进制文件   当前目录或 /usr/local/bin/sdp

================================================================================

提示：使用 sdp --help 查看所有命令，使用 sdp <command> --help 查看具体帮助。

                            -- 山东能力平台 CLI 助手开发团队
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
