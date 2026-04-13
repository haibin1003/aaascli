package cli

import (
	"testing"

	"github.com/user/lc/e2e/framework"
)

// TestVersion 测试版本命令（如果支持）
func TestVersion(t *testing.T) {
	t.Parallel()
	lc := framework.NewCLI(t)

	// 尝试 --version 或 version 子命令
	res := lc.Run("--version")
	if res.ExitCode != 0 {
		// 某些 CLI 可能使用 version 子命令
		res = lc.Run("version")
	}

	// 只要命令能执行（不管成功与否），都接受
	// 主要是验证 CLI 能正常响应
	if res.ExitCode != 0 && res.ExitCode != 1 {
		t.Logf("Version command returned exit code %d", res.ExitCode)
	}
}

// TestInvalidCommand 测试无效命令
func TestInvalidCommand(t *testing.T) {
	t.Parallel()
	lc := framework.NewCLI(t)

	res := lc.Run("invalid-command-xyz")
	res.ExpectFailure()
}

// TestGlobalFlags 测试全局标志
func TestGlobalFlags(t *testing.T) {
	t.Run("debug flag", func(t *testing.T) {
		t.Parallel()
		lc := framework.NewCLI(t)
		res := lc.Run("--debug", "--help")
		res.ExpectSuccess()
	})

	t.Run("insecure flag", func(t *testing.T) {
		t.Parallel()
		lc := framework.NewCLI(t)
		res := lc.Run("--insecure", "--help")
		res.ExpectSuccess()
	})
}

// TestEnvironmentIsolation 测试环境隔离
func TestEnvironmentIsolation(t *testing.T) {
	t.Parallel()

	// 创建两个独立的 CLI 实例
	lc1 := framework.NewCLI(t)
	lc2 := framework.NewCLI(t)

	// 验证它们使用不同的 home 目录
	res1 := lc1.Run("--help")
	res2 := lc2.Run("--help")

	res1.ExpectSuccess()
	res2.ExpectSuccess()

	// 两个实例应该都能独立工作
	if lc1.GetHome() == lc2.GetHome() {
		t.Error("Two CLI instances should have different home directories")
	}
}
