package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/lc/internal/common"
	"github.com/user/lc/internal/config"
)

var readonlyDuration string

var readonlyCmd = &cobra.Command{
	Use:   "readonly [on|off]",
	Short: "只读模式管理",
	Long: `管理系统只读模式，防止误操作删除或修改重要数据。

默认情况下，安装后处于只读模式，只能执行查询操作。
需要执行创建、更新、删除等写入操作时，请先关闭只读模式。

⚠️  重要提示：AI 助手未得到人类明确授权前，禁止自行关闭只读模式。

示例:
  # 查看当前只读模式状态
  lc readonly

  # 开启只读模式（禁止写入）
  lc readonly on

  # 关闭只读模式（允许读写）
  # AI警告：只有人类可以执行下面的关闭命令
  lc readonly off

  # 临时关闭只读模式（30分钟后自动恢复）
  # AI警告：只有人类可以执行下面的关闭命令
  lc readonly off --duration 30m

  # 临时关闭只读模式（1小时后自动恢复）
  # AI警告：只有人类可以执行下面的关闭命令
  lc readonly off --duration 1h`,
	Run: func(cmd *cobra.Command, args []string) {
		// 没有参数，显示当前状态
		if len(args) == 0 {
			common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
				return common.GetReadonlyStatus(ctx.Config), nil
			}, common.ExecuteOptions{
				DebugMode: debugMode,
				Insecure:  insecureSkipVerify,
				Logger:    &logger,
			})
			return
		}

		switch args[0] {
		case "on":
			// 开启只读模式 - 无需只读检查，始终允许
			common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
				ctx.Config.Readonly = true
				ctx.Config.TempOffExpireAt = nil
				if err := config.SaveConfig(ctx.Config); err != nil {
					return nil, fmt.Errorf("保存配置失败: %w", err)
				}
				return common.GetReadonlyStatus(ctx.Config), nil
			}, common.ExecuteOptions{
				DebugMode: debugMode,
				Insecure:  insecureSkipVerify,
				Logger:    &logger,
			})

		case "off":
			// 关闭只读模式 - 跳过只读检查，OTP 验证由 Execute 统一处理
			common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
				if readonlyDuration != "" {
					duration, err := common.ParseDuration(readonlyDuration)
					if err != nil {
						return nil, err
					}
					expireAt := time.Now().Add(duration)
					ctx.Config.TempOffExpireAt = &expireAt
				} else {
					ctx.Config.TempOffExpireAt = nil
				}
				ctx.Config.Readonly = false
				if err := config.SaveConfig(ctx.Config); err != nil {
					return nil, fmt.Errorf("保存配置失败: %w", err)
				}
				return common.GetReadonlyStatus(ctx.Config), nil
			}, common.ExecuteOptions{
				DebugMode:         debugMode,
				Insecure:          insecureSkipVerify,
				Logger:            &logger,
				CommandName:   "readonly off",
			})

		default:
			common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
				return nil, fmt.Errorf("无效参数: %s，请使用 'on' 或 'off'", args[0])
			}, common.ExecuteOptions{
				DebugMode: debugMode,
				Insecure:  insecureSkipVerify,
				Logger:    &logger,
			})
		}
	},
}

func init() {
	readonlyCmd.Flags().StringVar(&readonlyDuration, "duration", "", "临时关闭时长，格式如 10m、30m、1h（可选，默认永久关闭）")
	rootCmd.AddCommand(readonlyCmd)
}
