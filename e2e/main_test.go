package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/lc/e2e/framework"
)

var (
	// BinaryPath 是构建的 CLI 二进制文件路径
	BinaryPath string
	// ProjectRoot 是项目根目录
	ProjectRoot string
	// TestWorkspace 是默认测试工作空间（强制使用小白测）
	TestWorkspace = "XXJSxiaobaice"
	// TestWorkspaceName 是默认测试工作空间名称
	TestWorkspaceName = "小白测研发项目"
)

// TestMain 在运行所有测试之前构建 CLI 二进制文件
func TestMain(m *testing.M) {
	// 安全检查1：必须在临时目录或项目目录下运行
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误：无法获取当前工作目录: %v\n", err)
		os.Exit(1)
	}

	// 检查是否在允许的目录下（临时目录或项目目录）
	if !isAllowedDirectory(currentDir) {
		fmt.Fprintf(os.Stderr, "╔══════════════════════════════════════════════════════════════════════════════╗\n")
		fmt.Fprintf(os.Stderr, "║                           安全限制：测试终止                                  ║\n")
		fmt.Fprintf(os.Stderr, "╚══════════════════════════════════════════════════════════════════════════════╝\n")
		fmt.Fprintf(os.Stderr, "\n错误：E2E 测试必须在临时目录或项目目录下运行\n")
		fmt.Fprintf(os.Stderr, "当前目录: %s\n", currentDir)
		fmt.Fprintf(os.Stderr, "\n请先切换到临时目录：\n")
		fmt.Fprintf(os.Stderr, "  cd /tmp\n")
		fmt.Fprintf(os.Stderr, "  go test ./e2e/... -v\n")
		fmt.Fprintf(os.Stderr, "\n或者使用临时目录：\n")
		fmt.Fprintf(os.Stderr, "  cd $(mktemp -d)\n")
		fmt.Fprintf(os.Stderr, "  go test /path/to/lc/e2e/... -v\n")
		os.Exit(1)
	}

	// 安全检查2：强制设置环境变量使用小白测研发空间
	os.Setenv("LC_WORKSPACE_KEY", TestWorkspace)
	os.Setenv("LC_WORKSPACE_NAME", TestWorkspaceName)
	fmt.Fprintf(os.Stdout, "✅ E2E 测试配置检查通过\n")
	fmt.Fprintf(os.Stdout, "   当前目录: %s\n", currentDir)
	fmt.Fprintf(os.Stdout, "   强制使用研发空间: %s (%s)\n", TestWorkspace, TestWorkspaceName)
	fmt.Fprintf(os.Stdout, "\n")

	var err2 error
	ProjectRoot, err2 = framework.GetProjectRoot()
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "Failed to find project root: %v\n", err2)
		os.Exit(1)
	}

	// 构建二进制文件
	BinaryPath = filepath.Join(ProjectRoot, "bin", "lc")
	framework.BinaryPath = BinaryPath
	if err2 := buildBinary(); err2 != nil {
		fmt.Fprintf(os.Stderr, "Failed to build binary: %v\n", err2)
		os.Exit(1)
	}

	// 运行测试
	code := m.Run()

	// 清理二进制文件（可选，保留以便调试）
	// os.Remove(BinaryPath)

	os.Exit(code)
}

// isAllowedDirectory 检查当前目录是否在允许的目录下（临时目录或项目目录）
func isAllowedDirectory(dir string) bool {
	// 标准化路径
	dir = filepath.Clean(dir)

	// 检查常见的临时目录前缀
	tempPrefixes := []string{
		"/tmp",
		"/var/tmp",
		"/private/tmp", // macOS
		os.TempDir(),   // 系统默认临时目录
	}

	for _, prefix := range tempPrefixes {
		prefix = filepath.Clean(prefix)
		if dir == prefix || strings.HasPrefix(dir, prefix+string(filepath.Separator)) {
			return true
		}
	}

	// 检查环境变量定义的临时目录
	if tmpDir := os.Getenv("TMPDIR"); tmpDir != "" {
		tmpDir = filepath.Clean(tmpDir)
		if dir == tmpDir || strings.HasPrefix(dir, tmpDir+string(filepath.Separator)) {
			return true
		}
	}

	// 允许在项目目录（lc/e2e 或其父目录）下运行测试
	if strings.Contains(dir, string(filepath.Separator)+"lc"+string(filepath.Separator)+"e2e") ||
		strings.HasSuffix(dir, string(filepath.Separator)+"lc") {
		return true
	}

	return false
}

// buildBinary 使用 make build 构建 CLI 二进制文件
func buildBinary() error {
	cmd := exec.Command("make", "build")
	cmd.Dir = ProjectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
