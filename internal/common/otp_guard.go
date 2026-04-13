package common

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/user/lc/internal/config"
	"github.com/user/lc/internal/service"
)

// DefaultSessionExpiryMinutes 默认 OTP 会话有效期（分钟）
const DefaultSessionExpiryMinutes = 5

// DefaultProtectedCommands 默认需要 OTP 验证的命令列表
// 用户可通过配置文件覆盖此列表
// 原则: 只保护不可逆的高风险操作
var DefaultProtectedCommands = []string{
	"repo delete",  // 删除仓库 - 不可逆，所有代码和历史记录丢失
	"readonly off", // 关闭只读 - 会开放所有写入操作
}

// GetProtectedCommands 获取需要 OTP 验证的命令列表
// 优先使用配置文件中的列表，如果为空则使用默认列表
func GetProtectedCommands(cfg *config.Config) []string {
	if cfg.OTP.Enabled && len(cfg.OTP.ProtectedCommands) > 0 {
		return cfg.OTP.ProtectedCommands
	}
	return DefaultProtectedCommands
}

// IsCommandProtected 检查指定命令是否在受保护列表中
func IsCommandProtected(cfg *config.Config, operation string) bool {
	protected := GetProtectedCommands(cfg)
	for _, cmd := range protected {
		if cmd == operation {
			return true
		}
	}
	return false
}

// OTPGuardError OTP 验证错误
type OTPGuardError struct {
	Operation   string `json:"operation"`
	Description string `json:"description"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion"`
}

// Error 实现 error 接口
func (e *OTPGuardError) Error() string {
	return fmt.Sprintf("OTP 验证失败 [%s]: %s", e.Operation, e.Message)
}

// IsOTPEnabled 检查 OTP 是否已启用
func IsOTPEnabled(cfg *config.Config) bool {
	return cfg.OTP.Enabled && cfg.OTP.Secret != ""
}

// IsOTPSessionValid 检查 OTP 验证会话是否有效
func IsOTPSessionValid(cfg *config.Config) bool {
	if !IsOTPEnabled(cfg) {
		return false
	}

	if cfg.OTP.VerifiedAt == nil {
		return false
	}

	// Calculate expiry
	expiryMinutes := cfg.OTP.SessionExpiry
	if expiryMinutes == 0 {
		expiryMinutes = DefaultSessionExpiryMinutes
	}

	expiresAt := cfg.OTP.VerifiedAt.Add(time.Duration(expiryMinutes) * time.Minute)
	return time.Now().Before(expiresAt)
}

// getCommandMeta 获取命令元数据，未注册则返回默认值
func getCommandMeta(operation string) CommandMeta {
	if meta, ok := CommandRegistry[operation]; ok {
		return meta
	}
	return CommandMeta{
		Description: operation,
		RiskLevel:   "high",
		Reason:      "该操作需要二次验证",
	}
}

// CheckOTPForDangerousOperation 检查危险操作是否需要 OTP 验证
// 返回 nil 表示验证通过，否则返回错误
func CheckOTPForDangerousOperation(cfg *config.Config, operation string) error {
	// Check if OTP is enabled
	if !IsOTPEnabled(cfg) {
		return nil // OTP not enabled, no check needed
	}

	// Check if this command is in the protected list
	if !IsCommandProtected(cfg, operation) {
		return nil // Not a protected operation
	}

	// Check if session is valid
	if IsOTPSessionValid(cfg) {
		return nil // Session valid, allow operation
	}

	// Get operation info from registry for error message
	meta := getCommandMeta(operation)

	// Need OTP verification
	return &OTPGuardError{
		Operation:   operation,
		Description: meta.Description,
		Message:     fmt.Sprintf("执行 %s 需要 OTP 二次验证", meta.Description),
		Suggestion:  "请先运行 'lc otp verify' 进行验证",
	}
}

// PromptAndVerifyOTP 提示用户输入 OTP 并进行验证
// 用于交互式 OTP 验证，调用核心逻辑 VerifyOTPCode
func PromptAndVerifyOTP(cfg *config.Config) (bool, error) {
	if !IsOTPEnabled(cfg) {
		return true, nil // OTP not enabled, consider as passed
	}

	otpService := service.NewOTPService()

	// 所有交互提示写到 stderr，保持 stdout 为纯 JSON
	remaining := otpService.GetRemainingSeconds()
	if remaining < 10 {
		fmt.Fprintf(os.Stderr, "⏱️  当前验证码将在 %d 秒后过期，建议等待新验证码\n", remaining)
	}

	fmt.Fprint(os.Stderr, "🔐 请输入 OTP 验证码: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("读取输入失败: %w", err)
	}

	code := strings.TrimSpace(input)

	// 调用核心验证逻辑
	if err := VerifyOTPCode(cfg, code); err != nil {
		return false, err
	}

	fmt.Fprintf(os.Stderr, "✅ OTP 验证成功，会话有效期 %d 分钟\n", getSessionExpiry(cfg))
	return true, nil
}

// RequireOTPOrPrompt 检查 OTP 或提示用户输入
// 用于命令中集成 OTP 检查
func RequireOTPOrPrompt(cfg *config.Config, operation string) error {
	// First check if OTP check is needed
	if err := CheckOTPForDangerousOperation(cfg, operation); err == nil {
		return nil // No OTP needed or session valid
	}

	// Need OTP verification - show warning using CommandRegistry
	meta := getCommandMeta(operation)
	fmt.Fprintf(os.Stderr, "\n⚠️  警告: 即将执行操作 [%s]\n", meta.Description)
	fmt.Fprintf(os.Stderr, "   风险等级: %s\n", meta.RiskLevel)
	fmt.Fprintf(os.Stderr, "   原因: %s\n\n", meta.Reason)

	// Prompt and verify
	verified, err := PromptAndVerifyOTP(cfg)
	if err != nil {
		return err
	}

	if !verified {
		return &OTPGuardError{
			Operation:  operation,
			Message:    "OTP 验证失败",
			Suggestion: "请运行 'lc otp verify' 先进行验证",
		}
	}

	return nil
}

// GetOTPStatus 获取 OTP 状态信息
func GetOTPStatus(cfg *config.Config) map[string]interface{} {
	status := map[string]interface{}{
		"enabled": IsOTPEnabled(cfg),
	}

	if !IsOTPEnabled(cfg) {
		status["message"] = "OTP 二次验证未启用"
		return status
	}

	status["sessionExpiry"] = getSessionExpiry(cfg)

	if IsOTPSessionValid(cfg) {
		verifiedAt := *cfg.OTP.VerifiedAt
		expiresAt := verifiedAt.Add(time.Duration(getSessionExpiry(cfg)) * time.Minute)
		remaining := time.Until(expiresAt)

		status["session"] = map[string]interface{}{
			"valid":        true,
			"verifiedAt":   verifiedAt.Format(time.RFC3339),
			"expiresAt":    expiresAt.Format(time.RFC3339),
			"remainingMin": int(remaining.Minutes()),
			"remainingSec": int(remaining.Seconds()) % 60,
		}
	} else {
		status["session"] = map[string]interface{}{
			"valid":   false,
			"message": "无有效验证会话",
		}
	}

	return status
}

// getSessionExpiry 获取会话有效期（分钟）
func getSessionExpiry(cfg *config.Config) int {
	if cfg.OTP.SessionExpiry > 0 {
		return cfg.OTP.SessionExpiry
	}
	return DefaultSessionExpiryMinutes
}

// VerifyOTPCode 验证 OTP 码并保存会话（核心逻辑）
// 检查 OTP 是否启用、验证格式、验证码、保存会话
// 返回成功/失败和错误信息
func VerifyOTPCode(cfg *config.Config, code string) error {
	// Check if OTP is enabled
	if !IsOTPEnabled(cfg) {
		return &OTPGuardError{
			Message:    "OTP 未启用",
			Suggestion: "运行 'lc otp setup' 启用 OTP",
		}
	}

	// Validate code format
	if len(code) != 6 {
		return &OTPGuardError{
			Message:    "验证码必须是 6 位数字",
			Suggestion: "请输入身份验证器显示的 6 位数字",
		}
	}

	// Verify code
	otpService := service.NewOTPService()
	valid, err := otpService.VerifyCode(cfg.OTP.Secret, code, 1)
	if err != nil {
		return fmt.Errorf("验证失败: %w", err)
	}

	if !valid {
		return &OTPGuardError{
			Message:    "验证码无效",
			Suggestion: "请检查验证码是否正确，或等待新验证码生成后重试",
		}
	}

	// Update verified timestamp
	now := time.Now()
	cfg.OTP.VerifiedAt = &now
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("保存验证状态失败: %w", err)
	}

	return nil
}
