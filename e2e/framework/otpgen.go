package framework

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

// OTPGenBinaryPath 是 lc-otp-gen 二进制文件路径，由 EnsureOTPGenBinary 设置
var OTPGenBinaryPath string

// OTPGen 封装 lc-otp-gen CLI，与 CLI 实例共享隔离的 HOME 目录。
// lc-otp-gen 将配置写入 ~/.lc-otp-gen/config.json，通过共享同一
// tmpdir HOME 实现完全隔离，不影响真实的 ~/.lc-otp-gen/ 配置。
type OTPGen struct {
	t    *testing.T
	bin  string
	home string // 与 CLI 实例共享的隔离 HOME 目录
}

// NewOTPGen 创建一个与指定 CLI 实例共享隔离 HOME 的 OTPGen。
// lc-otp-gen 的配置（~/.lc-otp-gen/）和 lc 的配置（~/.lc/）
// 都位于同一个 tmpdir 下，互相独立且不污染真实用户配置。
func NewOTPGen(t *testing.T, cli *CLI) *OTPGen {
	t.Helper()
	if err := EnsureOTPGenBinary(); err != nil {
		t.Fatalf("构建 lc-otp-gen 失败: %v", err)
	}
	return &OTPGen{
		t:    t,
		bin:  OTPGenBinaryPath,
		home: cli.GetHome(),
	}
}

// EnsureOTPGenBinary 确保 lc-otp-gen 二进制已构建，若不存在则自动构建。
func EnsureOTPGenBinary() error {
	if OTPGenBinaryPath != "" {
		if _, err := os.Stat(OTPGenBinaryPath); err == nil {
			return nil
		}
	}

	projectRoot, err := GetProjectRoot()
	if err != nil {
		return fmt.Errorf("获取项目根目录失败: %w", err)
	}

	OTPGenBinaryPath = filepath.Join(projectRoot, "bin", "lc-otp-gen")

	if _, err := os.Stat(OTPGenBinaryPath); err == nil {
		return nil // 已存在
	}

	// 构建
	cmd := exec.Command("go", "build", "-o", OTPGenBinaryPath, "./cmd/lc-otp-gen/")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("构建 lc-otp-gen 失败: %w", err)
	}
	return nil
}

// Run 以隔离 HOME 执行 lc-otp-gen 命令，返回原始输出结果。
func (o *OTPGen) Run(args ...string) *Result {
	o.t.Helper()
	cmd := exec.Command(o.bin, args...)
	cmd.Env = append(os.Environ(), "HOME="+o.home)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		t:        o.t,
	}
}

// Add 向隔离的 lc-otp-gen 配置中添加账户和密钥。
func (o *OTPGen) Add(account, secret string) {
	o.t.Helper()
	res := o.Run("add", account, secret)
	if res.ExitCode != 0 {
		o.t.Fatalf("lc-otp-gen add 失败: %s", res.Stdout+res.Stderr)
	}
	o.t.Logf("lc-otp-gen: 已添加账户 %s", account)
}

// Code 从 lc-otp-gen 获取指定账户的当前 TOTP 验证码（6 位数字字符串）。
// lc-otp-gen code 输出格式：
//
//	🔐 账户: xxx
//	🔢 验证码: 123456
//	⏱️  剩余: 25 秒
var codePattern = regexp.MustCompile(`验证码:\s*(\d{6})`)

func (o *OTPGen) Code(account string) string {
	o.t.Helper()
	res := o.Run("code", account)
	if res.ExitCode != 0 {
		o.t.Fatalf("lc-otp-gen code 失败: %s", res.Stdout+res.Stderr)
	}
	matches := codePattern.FindStringSubmatch(res.Stdout)
	if len(matches) < 2 {
		o.t.Fatalf("无法从 lc-otp-gen 输出中解析验证码:\n%s", res.Stdout)
	}
	o.t.Logf("lc-otp-gen: 账户 %s 当前验证码 %s", account, matches[1])
	return matches[1]
}
