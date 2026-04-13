package framework

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMode 定义测试模式
type TestMode string

const (
	// ManualMode 手动参数模式 - 显式提供所有参数
	ManualMode TestMode = "manual"
	// AutoDetectMode 自动探测模式 - 从 Git 仓库自动探测
	AutoDetectMode TestMode = "autodetect"
)

// TestConfig 测试配置
type TestConfig struct {
	WorkspaceKey  string
	WorkspaceName string
	ProjectCode   string
	GitProjectID  string
}

// GetManualModeConfig 获取手动参数模式测试配置（必须使用小白测空间）
func GetManualModeConfig() *TestConfig {
	return &TestConfig{
		WorkspaceKey:  GetEnvOrDefault("LC_MANUAL_WORKSPACE_KEY", "XXJSxiaobaice"),
		WorkspaceName: GetEnvOrDefault("LC_MANUAL_WORKSPACE_NAME", "小白测研发项目"),
		ProjectCode:   GetEnvOrDefault("LC_MANUAL_PROJECT_CODE", "R24113J3C04"),
		GitProjectID:  GetEnvOrDefault("LC_MANUAL_GIT_PROJECT_ID", "44618"),
	}
}

// GetAutoDetectConfig 获取自动探测模式测试配置（小白测空间，用于验证）
// 自动探测使用 forever 仓库，该仓库属于小白测研发项目
func GetAutoDetectConfig() *TestConfig {
	return &TestConfig{
		// 自动探测模式下这些值仅用于验证，实际通过 Git 仓库自动探测
		WorkspaceKey:  GetEnvOrDefault("LC_AUTODETECT_WORKSPACE_KEY", "XXJSxiaobaice"),
		WorkspaceName: GetEnvOrDefault("LC_AUTODETECT_WORKSPACE_NAME", "小白测研发项目"),
		ProjectCode:   GetEnvOrDefault("LC_AUTODETECT_PROJECT_CODE", "R24113J3C04"),
		GitProjectID:  GetEnvOrDefault("LC_AUTODETECT_GIT_PROJECT_ID", "44650"),
	}
}

// ManualModeTest 手动参数模式测试基类
type ManualModeTest struct {
	T      *testing.T
	CLI    *CLI
	Config *TestConfig
}

// NewManualModeTest 创建手动参数模式测试（使用小白测空间）
func NewManualModeTest(t *testing.T) *ManualModeTest {
	return &ManualModeTest{
		T:      t,
		CLI:    NewCLI(t),
		Config: GetManualModeConfig(),
	}
}

// DisableReadonly 关闭只读模式（用于需要写入操作的测试）
func (m *ManualModeTest) DisableReadonly() {
	m.T.Helper()
	res := m.CLI.Run("readonly", "off", "--duration", "30m")
	if res.IsSuccess() {
		m.T.Logf("已关闭只读模式")
	} else {
		m.T.Logf("关闭只读模式失败（可能已被其他测试关闭）: %s", res.Stderr)
	}
}

// Run 执行命令（自动添加 -k 和 workspace-key）
func (m *ManualModeTest) Run(args ...string) *Result {
	m.T.Helper()

	// 检查是否已包含 -k 或 --workspace-key
	hasWorkspaceKey := false
	hasShortK := false
	for i, arg := range args {
		if arg == "--workspace-key" && i+1 < len(args) {
			hasWorkspaceKey = true
		}
		if arg == "-k" {
			hasShortK = true
		}
	}

	// 构建最终参数
	var finalArgs []string
	if !hasShortK {
		finalArgs = append(finalArgs, "-k")
	}
	finalArgs = append(finalArgs, args...)

	// 如果没有指定 workspace-key，添加默认的
	if !hasWorkspaceKey {
		finalArgs = append(finalArgs, "--workspace-key", m.Config.WorkspaceKey)
	}

	return m.CLI.Run(finalArgs...)
}

// RunWithProject 执行命令（包含 project-code）
func (m *ManualModeTest) RunWithProject(args ...string) *Result {
	// 检查是否已包含 project-code
	hasProjectCode := false
	for i, arg := range args {
		if arg == "--project-code" && i+1 < len(args) {
			hasProjectCode = true
			break
		}
	}

	if !hasProjectCode && m.Config.ProjectCode != "" {
		args = append(args, "--project-code", m.Config.ProjectCode)
	}

	return m.Run(args...)
}

// AutoDetectTest 自动探测模式测试基类
type AutoDetectTest struct {
	T       *testing.T
	CLI     *CLI
	RepoDir string
	Config  *TestConfig
}

// NewAutoDetectTest 创建自动探测模式测试（使用 CLI 研发空间，通过 Git 自动探测）
func NewAutoDetectTest(t *testing.T) *AutoDetectTest {
	repoDir := PrepareTestGitRepo(t)
	return &AutoDetectTest{
		T:       t,
		CLI:     NewCLI(t),
		RepoDir: repoDir,
		Config:  GetAutoDetectConfig(),
	}
}

// RunInRepo 在 Git 仓库目录下执行命令
func (a *AutoDetectTest) RunInRepo(args ...string) *Result {
	a.T.Helper()

	// 保存当前目录
	originalDir, err := os.Getwd()
	if err != nil {
		a.T.Fatalf("获取当前目录失败: %v", err)
	}

	// 切换到仓库目录
	if err := os.Chdir(a.RepoDir); err != nil {
		a.T.Fatalf("切换到仓库目录失败: %v", err)
	}

	// 执行命令
	result := a.CLI.Run(args...)

	// 恢复原始目录
	if err := os.Chdir(originalDir); err != nil {
		a.T.Logf("警告: 恢复目录失败: %v", err)
	}

	return result
}

// GetRepoDir 获取测试仓库目录
func (a *AutoDetectTest) GetRepoDir() string {
	return a.RepoDir
}

// PrepareTestGitRepo 准备测试用的 Git 仓库
func PrepareTestGitRepo(t *testing.T) string {
	t.Helper()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "lc-autodetect-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}

	// 清理函数
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	// 初始化 Git 仓库
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("初始化 Git 仓库失败: %v", err)
	}

	// 设置 Git 配置
	configCmds := [][]string{
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
	}

	for _, cfg := range configCmds {
		cmd := exec.Command(cfg[0], cfg[1:]...)
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			t.Logf("设置 Git 配置失败: %v", err)
		}
	}

	// 设置远程仓库 URL（使用小白测空间的 forever 仓库用于自动探测测试）
	// 自动探测测试使用小白测研发项目空间，通过 Git remote URL 自动探测
	repoURL := os.Getenv("LC_TEST_REPO_URL")
	if repoURL == "" {
		// 默认使用小白测研发项目的 forever 仓库 URL
		// 该仓库关联的研发空间 Key: XXJSxiaobaice, 名称: 小白测研发项目
		repoURL = "http://code-xxjs.rdcloud.4c.hq.cmcc/osc/XXJS/weibaohui-hq.cmcc/forever.git"
	}

	cmd = exec.Command("git", "remote", "add", "origin", repoURL)
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Logf("添加远程仓库失败: %v", err)
	}

	// 创建一个初始提交（某些 Git 操作需要）
	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repo\n"), 0644); err == nil {
		cmd = exec.Command("git", "add", "README.md")
		cmd.Dir = tmpDir
		cmd.Run()

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = tmpDir
		cmd.Run()
	}

	t.Logf("测试 Git 仓库准备完成: %s", tmpDir)
	return tmpDir
}

// SkipIfNotInGitRepo 如果不在 Git 仓库中则跳过测试
func SkipIfNotInGitRepo(t *testing.T) {
	t.Helper()

	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		t.Skip("跳过测试：当前目录不在 Git 仓库中")
	}
}
