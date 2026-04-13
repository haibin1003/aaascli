package framework

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// BinaryPath 是测试用的 CLI 二进制文件路径
// 在 TestMain 中设置
var BinaryPath string

// CLI 是 CLI 测试运行器
type CLI struct {
	t    *testing.T
	bin  string
	home string
	env  map[string]string
}

// NewCLI 创建一个新的 CLI 测试运行器
// 自动创建临时 HOME 目录，复制配置文件，测试结束后自动清理
func NewCLI(t *testing.T) *CLI {
	t.Helper()

	// 确保二进制文件已构建
	if err := EnsureBinary(); err != nil {
		t.Fatalf("Failed to ensure binary: %v", err)
	}

	// 创建临时 HOME 目录
	home, err := os.MkdirTemp("", "lc-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}

	// 创建 .lc 配置目录
	lcConfigDir := filepath.Join(home, ".lc")
	if err := os.MkdirAll(lcConfigDir, 0755); err != nil {
		os.RemoveAll(home)
		t.Fatalf("Failed to create .lc config dir: %v", err)
	}

	// 复制用户原始配置文件到临时目录
	copyConfigFile(t, lcConfigDir)

	cli := &CLI{
		t:    t,
		bin:  BinaryPath,
		home: home,
		env:  make(map[string]string),
	}

	// 注册清理函数
	t.Cleanup(func() {
		cli.Cleanup()
	})

	return cli
}

// copyConfigFile 复制用户的 LC 配置文件到临时目录
func copyConfigFile(t *testing.T, destDir string) {
	t.Helper()

	// 尝试从原始 HOME 目录复制配置文件
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}

	if originalHome != "" {
		srcConfig := filepath.Join(originalHome, ".lc", "config.json")
		if _, err := os.Stat(srcConfig); err == nil {
			content, err := os.ReadFile(srcConfig)
			if err == nil {
				destConfig := filepath.Join(destDir, "config.json")
				if err := os.WriteFile(destConfig, content, 0644); err != nil {
					t.Logf("Warning: failed to copy config file: %v", err)
				} else {
					t.Logf("Copied config from %s to %s", srcConfig, destConfig)
				}
			}
		} else {
			t.Logf("No config file found at %s", srcConfig)
		}
	}
}

// WithEnv 设置环境变量
func (c *CLI) WithEnv(key, value string) *CLI {
	c.env[key] = value
	return c
}

// Cleanup 清理临时目录
func (c *CLI) Cleanup() {
	if c.home != "" {
		os.RemoveAll(c.home)
	}
}

// GetHome 返回 CLI 实例的临时 HOME 目录
func (c *CLI) GetHome() string {
	return c.home
}

// Run 执行 CLI 命令
func (c *CLI) Run(args ...string) *Result {
	c.t.Helper()

	cmd := exec.Command(c.bin, args...)

	// 设置 HOME 环境变量以隔离配置
	cmd.Env = c.buildEnv()

	// 捕获输出
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行命令
	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		t:        c.t,
	}
}

// RunWithInput 执行 CLI 命令并通过 stdin 输入内容
func (c *CLI) RunWithInput(input string, args ...string) *Result {
	c.t.Helper()

	cmd := exec.Command(c.bin, args...)
	cmd.Env = c.buildEnv()

	// 设置 stdin
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		t:        c.t,
	}
}

// RunWithRetry 带重试执行命令
func (c *CLI) RunWithRetry(maxRetries int, delaySec int, args ...string) *Result {
	c.t.Helper()

	var lastResult *Result
	for i := 0; i < maxRetries; i++ {
		lastResult = c.Run(args...)
		if lastResult.ExitCode == 0 {
			return lastResult
		}
		if i < maxRetries-1 {
			time.Sleep(time.Duration(delaySec) * time.Second)
		}
	}
	return lastResult
}

// buildEnv 构建环境变量
func (c *CLI) buildEnv() []string {
	// 从当前环境复制
	env := os.Environ()

	// 设置 HOME
	env = setEnv(env, "HOME", c.home)

	// 设置 LC_CONFIG_DIR（如果 CLI 支持）
	env = setEnv(env, "LC_CONFIG_DIR", filepath.Join(c.home, ".lc"))

	// 应用自定义环境变量
	for k, v := range c.env {
		env = setEnv(env, k, v)
	}

	return env
}

// setEnv 设置或替换环境变量
func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
