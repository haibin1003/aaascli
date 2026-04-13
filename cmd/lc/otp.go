package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
	"github.com/user/lc/internal/common"
	"github.com/user/lc/internal/config"
	"github.com/user/lc/internal/service"
)

var otpCmd = &cobra.Command{
	Use:   "otp",
	Short: "OTP 二次验证管理",
	Long: `管理 OTP (One-Time Password) 二次验证，保护危险操作。

子命令:
  setup   - 初始化 OTP，生成密钥并显示二维码
  verify  - 验证 OTP 密码
  disable - 关闭 OTP 验证 (需要 OTP 验证)
  status  - 查看 OTP 状态

示例:
  # 初始化 OTP
  lc otp setup

  # 验证 OTP 密码
  lc otp verify 123456

  # 查看 OTP 状态
  lc otp status`,
}

var otpSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "初始化 OTP 二次验证",
	Long: `生成 OTP 密钥并显示二维码，请使用身份验证器应用扫描。

支持的身份验证器:
  • Google Authenticator
  • Microsoft Authenticator
  • Authy
  • 1Password
  • 其他支持 TOTP 的应用

注意:
  • 请妥善保存密钥，丢失后无法恢复
  • 建议生成并保存备用码
  • 启用后，危险操作将需要输入 OTP 密码`,
	Run: func(cmd *cobra.Command, args []string) {
		runOTPSetup()
	},
}

var otpVerifyCmd = &cobra.Command{
	Use:   "verify [code]",
	Short: "验证 OTP 密码",
	Long: `验证 OTP 密码并创建验证会话。

验证成功后会话将持续 5 分钟（可在配置中修改），
在此期间执行危险操作无需再次输入 OTP。

示例:
  lc otp verify 123456`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		code := ""
		if len(args) > 0 {
			code = args[0]
		}
		runOTPVerify(code)
	},
}

var otpDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "关闭 OTP 二次验证",
	Long: `关闭 OTP 二次验证功能。

⚠️  警告: 此操作需要 OTP 验证，且会立即生效。
关闭后，所有操作将不再需要 OTP 密码。`,
	Run: func(cmd *cobra.Command, args []string) {
		runOTPDisable()
	},
}

var otpStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看 OTP 状态",
	Long:  `查看 OTP 二次验证的当前状态，包括是否启用、会话状态、受保护命令列表等。`,
	Run: func(cmd *cobra.Command, args []string) {
		runOTPStatus()
	},
}

var otpConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "管理 OTP 受保护命令列表",
	Long: `管理需要 OTP 二次验证的命令列表。

子命令:
  list    - 列出当前受保护的命令
  add     - 添加命令到保护列表
  remove  - 从保护列表移除命令
  reset   - 重置为默认列表

示例:
  # 查看当前受保护命令列表
  lc otp config list

  # 添加命令到保护列表
  lc otp config add "req delete"
  lc otp config add "task delete"

  # 从保护列表移除命令
  lc otp config remove "pr create"

  # 重置为默认列表
  lc otp config reset`,
}

var otpConfigListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出受保护的命令",
	Run: func(cmd *cobra.Command, args []string) {
		runOTPConfigList()
	},
}

var otpConfigAddCmd = &cobra.Command{
	Use:   "add [command]",
	Short: "添加命令到保护列表",
	Example: `  lc otp config add "req delete"
  lc otp config add "task delete"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runOTPConfigAdd(args[0])
	},
}

var otpConfigRemoveCmd = &cobra.Command{
	Use:   "remove [command]",
	Short: "从保护列表移除命令",
	Example: `  lc otp config remove "pr create"
  lc otp config remove "req delete"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runOTPConfigRemove(args[0])
	},
}

var otpConfigResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "重置为默认保护列表",
	Long:  `清除自定义保护列表，恢复为系统默认的受保护命令。`,
	Run: func(cmd *cobra.Command, args []string) {
		runOTPConfigReset()
	},
}

func init() {
	rootCmd.AddCommand(otpCmd)
	otpCmd.AddCommand(otpSetupCmd)
	otpCmd.AddCommand(otpVerifyCmd)
	otpCmd.AddCommand(otpDisableCmd)
	otpCmd.AddCommand(otpStatusCmd)
	otpCmd.AddCommand(otpConfigCmd)

	// Add config subcommands
	otpConfigCmd.AddCommand(otpConfigListCmd)
	otpConfigCmd.AddCommand(otpConfigAddCmd)
	otpConfigCmd.AddCommand(otpConfigRemoveCmd)
	otpConfigCmd.AddCommand(otpConfigResetCmd)
}

func runOTPSetup() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		cfg := ctx.Config

		// Check if OTP is already enabled
		if cfg.OTP.Enabled {
			return nil, &common.AutoDetectError{
				Message:    "OTP 已启用",
				Details:    "当前账户已启用 OTP 二次验证",
				Suggestion: "如需重新设置，请先运行 'lc otp disable' 关闭现有 OTP",
			}
		}

		// Generate new secret
		otpService := service.NewOTPService()
		secret, err := otpService.GenerateSecret()
		if err != nil {
			return nil, fmt.Errorf("生成 OTP 密钥失败: %w", err)
		}

		// Get user info for QR code
		user := cfg.GetUser()
		account := user.Username
		if account == "" {
			account = "lc-user"
		}

		// Generate QR code URL
		qrURL := otpService.GenerateQRCodeURL(secret, account, "灵畿CLI")

		// Generate backup codes
		backupCodes, err := otpService.GenerateBackupCodes()
		if err != nil {
			return nil, fmt.Errorf("生成备用码失败: %w", err)
		}

		// 所有设置向导 UI 写到 stderr，保持 stdout 为纯 JSON
		fmt.Fprintln(os.Stderr, "\n╔════════════════════════════════════════════════════════════════╗")
		fmt.Fprintln(os.Stderr, "║              🔐 OTP 二次验证初始化                              ║")
		fmt.Fprintln(os.Stderr, "╠════════════════════════════════════════════════════════════════╣")
		fmt.Fprintln(os.Stderr, "║                                                                ║")
		fmt.Fprintln(os.Stderr, "║  步骤 1: 使用身份验证器应用扫描二维码                            ║")
		fmt.Fprintln(os.Stderr, "║                                                                ║")
		fmt.Fprintf(os.Stderr, "║  账户: %-55s ║\n", account)
		fmt.Fprintln(os.Stderr, "║                                                                ║")

		// Print QR code URL
		fmt.Fprintln(os.Stderr, "║  [请扫描下方二维码或手动输入密钥]                                ║")
		fmt.Fprintln(os.Stderr, "║                                                                ║")

		// 在终端显示二维码（强制最小版本）
		fmt.Fprintln(os.Stderr, "╠════════════════════════════════════════════════════════════════╣")
		// 尝试强制使用版本 1（最小尺寸），如果失败则使用默认
		qr, err := qrcode.NewWithForcedVersion(qrURL, 1, qrcode.Low)
		if err != nil {
			// 如果版本 1 放不下，使用自动计算的版本
			qr, _ = qrcode.New(qrURL, qrcode.Low)
		}
		fmt.Fprintln(os.Stderr, qr.ToSmallString(false))
		fmt.Fprintln(os.Stderr, "╠════════════════════════════════════════════════════════════════╣")
		fmt.Fprintf(os.Stderr, "║  URL: %s\n", qrURL)
		fmt.Fprintln(os.Stderr, "║                                                                ║")

		// Display manual entry key
		formattedSecret := otpService.FormatSecretForDisplay(secret)
		fmt.Fprintln(os.Stderr, "║  步骤 2: 或手动输入以下密钥                                      ║")
		fmt.Fprintln(os.Stderr, "║                                                                ║")
		fmt.Fprintf(os.Stderr, "║  %-62s ║\n", formattedSecret)
		fmt.Fprintln(os.Stderr, "║                                                                ║")

		// Display backup codes
		fmt.Fprintln(os.Stderr, "║  步骤 3: 请保存以下备用码（用于恢复账户）                        ║")
		fmt.Fprintln(os.Stderr, "║                                                                ║")
		for i := 0; i < len(backupCodes); i += 2 {
			fmt.Fprintf(os.Stderr, "║    %s    %s\n", backupCodes[i], backupCodes[i+1])
		}
		fmt.Fprintln(os.Stderr, "║                                                                ║")
		fmt.Fprintln(os.Stderr, "╚════════════════════════════════════════════════════════════════╝")
		fmt.Fprintln(os.Stderr)

		// Prompt user to verify setup
		fmt.Fprint(os.Stderr, "请输入身份验证器显示的 6 位验证码以确认设置: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("读取输入失败: %w", err)
		}
		verifyCode := strings.TrimSpace(input)

		// Verify the code
		valid, err := otpService.VerifyCode(secret, verifyCode, 1)
		if err != nil || !valid {
			return nil, &common.AutoDetectError{
				Message:    "验证码无效",
				Details:    "请确保正确扫描了二维码或输入了密钥",
				Suggestion: "请重试 'lc otp setup'",
			}
		}

		// Save OTP configuration
		now := time.Now()
		cfg.OTP.Enabled = true
		cfg.OTP.Secret = secret
		cfg.OTP.VerifiedAt = &now
		if cfg.OTP.SessionExpiry == 0 {
			cfg.OTP.SessionExpiry = common.DefaultSessionExpiryMinutes
		}

		if err := config.SaveConfig(cfg); err != nil {
			return nil, fmt.Errorf("保存配置失败: %w", err)
		}

		return map[string]interface{}{
			"enabled":       true,
			"account":       account,
			"backupCodes":   backupCodes,
			"sessionExpiry": cfg.OTP.SessionExpiry,
			"message":       "OTP 二次验证已成功启用",
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "",
	})
}

func runOTPVerify(code string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		cfg := ctx.Config

		// If code not provided as argument, prompt for it
		if code == "" {
			fmt.Fprint(os.Stderr, "请输入 OTP 验证码: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("读取输入失败: %w", err)
			}
			code = strings.TrimSpace(input)
		}

		// 调用核心验证逻辑
		if err := common.VerifyOTPCode(cfg, code); err != nil {
			// 转换错误类型为 AutoDetectError 以保持向后兼容
			if otpErr, ok := err.(*common.OTPGuardError); ok {
				return nil, &common.AutoDetectError{
					Message:    otpErr.Message,
					Details:    otpErr.Suggestion,
					Suggestion: otpErr.Suggestion,
				}
			}
			return nil, err
		}

		// Calculate expiry time
		now := time.Now()
		expiryMinutes := cfg.OTP.SessionExpiry
		if expiryMinutes == 0 {
			expiryMinutes = common.DefaultSessionExpiryMinutes
		}
		expiresAt := now.Add(time.Duration(expiryMinutes) * time.Minute)

		return map[string]interface{}{
			"verified":    true,
			"verifiedAt":  now.Format(time.RFC3339),
			"expiresAt":   expiresAt.Format(time.RFC3339),
			"durationMin": expiryMinutes,
			"message":     fmt.Sprintf("验证成功，会话有效期 %d 分钟", expiryMinutes),
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "",
	})
}

func runOTPDisable() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		cfg := ctx.Config

		// Check if OTP is enabled
		if !cfg.OTP.Enabled {
			return nil, &common.AutoDetectError{
				Message:    "OTP 未启用",
				Details:    "当前账户未配置 OTP 二次验证",
				Suggestion: "无需禁用，OTP 尚未启用",
			}
		}

		// Require OTP verification to disable.
		// 与其他受保护命令保持一致：无有效会话时自动提示输入验证码，无需手动运行 lc otp verify。
		if !common.IsOTPSessionValid(cfg) {
			if _, err := common.PromptAndVerifyOTP(cfg); err != nil {
				return nil, err
			}
		}

		// 确认提示写到 stderr，保持 stdout 为纯 JSON
		fmt.Fprint(os.Stderr, "⚠️  确定要关闭 OTP 二次验证吗? 这将降低账户安全性。 [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("读取输入失败: %w", err)
		}
		input = strings.TrimSpace(strings.ToLower(input))

		if input != "y" && input != "yes" {
			return nil, fmt.Errorf("操作已取消")
		}

		// Clear OTP configuration
		cfg.OTP.Enabled = false
		cfg.OTP.Secret = ""
		cfg.OTP.VerifiedAt = nil

		if err := config.SaveConfig(cfg); err != nil {
			return nil, fmt.Errorf("保存配置失败: %w", err)
		}

		return map[string]interface{}{
			"enabled": false,
			"message": "OTP 二次验证已关闭",
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "",
	})
}

func runOTPStatus() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		cfg := ctx.Config

		data := map[string]interface{}{
			"enabled": cfg.OTP.Enabled,
		}

		if cfg.OTP.Enabled {
			data["sessionExpiry"] = cfg.OTP.SessionExpiry
			if cfg.OTP.SessionExpiry == 0 {
				data["sessionExpiry"] = common.DefaultSessionExpiryMinutes
			}

			// Check session validity
			if common.IsOTPSessionValid(cfg) {
				verifiedAt := *cfg.OTP.VerifiedAt
				expiryMinutes := cfg.OTP.SessionExpiry
				if expiryMinutes == 0 {
					expiryMinutes = common.DefaultSessionExpiryMinutes
				}
				expiresAt := verifiedAt.Add(time.Duration(expiryMinutes) * time.Minute)
				remaining := time.Until(expiresAt)

				data["session"] = map[string]interface{}{
					"valid":        true,
					"verifiedAt":   verifiedAt.Format(time.RFC3339),
					"expiresAt":    expiresAt.Format(time.RFC3339),
					"remainingMin": int(remaining.Minutes()),
					"remainingSec": int(remaining.Seconds()) % 60,
				}
			} else {
				data["session"] = map[string]interface{}{
					"valid":   false,
					"message": "无有效验证会话，执行危险操作前需要验证",
				}
			}

			// Add protected commands info
			protected := common.GetProtectedCommands(cfg)
			data["protectedCommands"] = map[string]interface{}{
				"commands": protected,
				"count":    len(protected),
				"isCustom": len(cfg.OTP.ProtectedCommands) > 0,
			}
		}

		return data, nil
	}, common.ExecuteOptions{
		DebugMode: debugMode,
		Insecure:  insecureSkipVerify,
		Logger:    &logger,
		// CommandName is empty for read-only query
	})
}

func runOTPConfigList() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		cfg := ctx.Config

		// Get protected commands
		protected := common.GetProtectedCommands(cfg)
		isCustom := len(cfg.OTP.ProtectedCommands) > 0

		data := map[string]interface{}{
			"protectedCommands": protected,
			"isCustom":          isCustom,
		}

		if !isCustom {
			data["message"] = "使用默认保护列表"
			data["defaultCommands"] = common.DefaultProtectedCommands
		} else {
			data["message"] = "使用自定义保护列表"
		}

		if cfg.OTP.Enabled {
			data["otpEnabled"] = true
		} else {
			data["otpEnabled"] = false
			// OTP 未启用时明确提示：保护列表已配置但当前未生效
			data["warning"] = "OTP 未启用，以下列表仅供参考，当前不会触发任何拦截"
		}

		return data, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "",
	})
}

func runOTPConfigAdd(command string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		cfg := ctx.Config

		// Check if OTP is enabled
		if !cfg.OTP.Enabled {
			return nil, &common.AutoDetectError{
				Message:    "OTP 未启用",
				Details:    "需要先启用 OTP 才能配置保护列表",
				Suggestion: "运行 'lc otp setup' 启用 OTP",
			}
		}

		// Validate command format (simple check)
		if command == "" {
			return nil, fmt.Errorf("命令不能为空")
		}

		// 对有效列表（自定义列表或默认列表）做重复检查，防止静默覆盖默认保护命令
		effectiveList := common.GetProtectedCommands(cfg)
		for _, cmd := range effectiveList {
			if cmd == command {
				return nil, &common.AutoDetectError{
					Message:    "命令已在保护列表中",
					Details:    fmt.Sprintf("'%s' 已经是受保护命令", command),
					Suggestion: "使用 'lc otp config list' 查看当前列表",
				}
			}
		}

		// 若当前使用默认列表（自定义列表为空），先以默认列表为基础初始化自定义列表，
		// 避免切换到自定义列表后丢失其他默认保护命令。
		if len(cfg.OTP.ProtectedCommands) == 0 {
			cfg.OTP.ProtectedCommands = make([]string, len(effectiveList))
			copy(cfg.OTP.ProtectedCommands, effectiveList)
		}

		// Add to protected commands
		cfg.OTP.ProtectedCommands = append(cfg.OTP.ProtectedCommands, command)

		if err := config.SaveConfig(cfg); err != nil {
			return nil, fmt.Errorf("保存配置失败: %w", err)
		}

		return map[string]interface{}{
			"added":             command,
			"protectedCommands": cfg.OTP.ProtectedCommands,
			"message":           fmt.Sprintf("已将 '%s' 添加到 OTP 保护列表", command),
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "",
	})
}

func runOTPConfigRemove(command string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		cfg := ctx.Config

		// Check if OTP is enabled
		if !cfg.OTP.Enabled {
			return nil, &common.AutoDetectError{
				Message:    "OTP 未启用",
				Details:    "需要先启用 OTP 才能配置保护列表",
				Suggestion: "运行 'lc otp setup' 启用 OTP",
			}
		}

		// Check if using custom list
		if len(cfg.OTP.ProtectedCommands) == 0 {
			return nil, &common.AutoDetectError{
				Message:    "当前使用默认保护列表",
				Details:    fmt.Sprintf("'%s' 在默认列表中，无法单独移除", command),
				Suggestion: "先运行 'lc otp config add <其他命令>' 创建自定义列表，然后移除不需要的命令",
			}
		}

		// Find and remove command
		found := false
		newList := make([]string, 0, len(cfg.OTP.ProtectedCommands))
		for _, cmd := range cfg.OTP.ProtectedCommands {
			if cmd == command {
				found = true
				continue
			}
			newList = append(newList, cmd)
		}

		if !found {
			return nil, &common.AutoDetectError{
				Message:    "命令不在保护列表中",
				Details:    fmt.Sprintf("'%s' 不是受保护命令", command),
				Suggestion: "使用 'lc otp config list' 查看当前列表",
			}
		}

		cfg.OTP.ProtectedCommands = newList

		if err := config.SaveConfig(cfg); err != nil {
			return nil, fmt.Errorf("保存配置失败: %w", err)
		}

		return map[string]interface{}{
			"removed":           command,
			"protectedCommands": cfg.OTP.ProtectedCommands,
			"message":           fmt.Sprintf("已将 '%s' 从 OTP 保护列表移除", command),
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "",
	})
}

func runOTPConfigReset() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		cfg := ctx.Config

		// Check if OTP is enabled
		if !cfg.OTP.Enabled {
			return nil, &common.AutoDetectError{
				Message:    "OTP 未启用",
				Details:    "需要先启用 OTP 才能配置保护列表",
				Suggestion: "运行 'lc otp setup' 启用 OTP",
			}
		}

		// Check if already using default
		if len(cfg.OTP.ProtectedCommands) == 0 {
			return map[string]interface{}{
				"message":           "当前已经是默认保护列表",
				"protectedCommands": common.DefaultProtectedCommands,
			}, nil
		}

		// 确认提示写到 stderr，保持 stdout 为纯 JSON
		fmt.Fprintln(os.Stderr, "即将重置 OTP 保护列表为默认值:")
		for _, cmd := range common.DefaultProtectedCommands {
			fmt.Fprintf(os.Stderr, "  - %s\n", cmd)
		}
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "当前自定义列表 (%d 个命令) 将被清除。确定继续? [y/N]: ", len(cfg.OTP.ProtectedCommands))

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("读取输入失败: %w", err)
		}
		input = strings.TrimSpace(strings.ToLower(input))

		if input != "y" && input != "yes" {
			return nil, fmt.Errorf("操作已取消")
		}

		// Clear custom list
		cfg.OTP.ProtectedCommands = nil

		if err := config.SaveConfig(cfg); err != nil {
			return nil, fmt.Errorf("保存配置失败: %w", err)
		}

		return map[string]interface{}{
			"protectedCommands": common.DefaultProtectedCommands,
			"message":           "已重置为默认 OTP 保护列表",
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecureSkipVerify,
		Logger:      &logger,
		PrettyPrint: prettyMode,
		CommandName: "",
	})
}
