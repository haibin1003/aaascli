package framework

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// UniqueName 生成一个唯一的测试名称
func UniqueName(prefix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d", prefix, timestamp)
}

// CreateConfigFile 在 CLI 的临时目录中创建配置文件
func (c *CLI) CreateConfigFile(content string) error {
	configPath := filepath.Join(c.home, ".lc", "config.json")
	return os.WriteFile(configPath, []byte(content), 0644)
}

// CreateConfigFileFromStruct 使用结构体创建配置文件
func (c *CLI) CreateConfigFileFromStruct(cfg interface{}) error {
	// 这里可以实现 JSON/YAML 序列化
	// 简化版本，直接使用字符串
	return fmt.Errorf("not implemented")
}

// SkipIfShort 如果在 short 模式下则跳过
func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

// SkipIfNoConfig 如果没有配置则跳过测试
func SkipIfNoConfig(t *testing.T) {
	t.Helper()
	if os.Getenv("LC_API_TOKEN") == "" {
		t.Skip("Skipping test: LC_API_TOKEN not set")
	}
}

// RequireEnv 要求环境变量必须存在，否则跳过测试
func RequireEnv(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	if value == "" {
		t.Skipf("Skipping test: required environment variable %s not set", key)
	}
	return value
}

// GetEnvOrDefault 获取环境变量或默认值
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Retry 重试执行函数
func Retry(t *testing.T, maxAttempts int, delay time.Duration, fn func() error) error {
	t.Helper()

	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			t.Logf("Attempt %d failed: %v, retrying in %v...", i+1, err, delay)
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

// PollForCondition 轮询直到条件满足或超时
func PollForCondition(t *testing.T, maxRetries int, delay time.Duration, condition func() bool) bool {
	t.Helper()
	for i := 0; i < maxRetries; i++ {
		if condition() {
			return true
		}
		t.Logf("Polling attempt %d/%d failed, waiting...", i+1, maxRetries)
		time.Sleep(delay)
	}
	return false
}

// GenerateTimestamp 生成时间戳字符串
func GenerateTimestamp() string {
	return time.Now().Format("20060102_150405")
}

// GenerateRepoName 生成唯一的仓库名称
func GenerateRepoName() string {
	return fmt.Sprintf("test-e2e-repo-%d", time.Now().Unix())
}

// ReadFile 读取 golden 文件内容
func ReadFile(t *testing.T, projectRoot, filename string) string {
	t.Helper()
	path := filepath.Join(projectRoot, "testdata", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(content)
}

// WriteGolden 写入 golden 文件（用于更新 golden 文件）
func WriteGolden(t *testing.T, projectRoot, filename string, content string) {
	t.Helper()
	testdataDir := filepath.Join(projectRoot, "testdata")
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		t.Fatalf("Failed to create testdata dir: %v", err)
	}
	path := filepath.Join(testdataDir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write golden file %s: %v", path, err)
	}
}

// CreateTempFile 创建临时文件并返回路径
func CreateTempFile(t *testing.T, prefix string, content string) string {
	t.Helper()
	tempFile, err := os.CreateTemp("", prefix+"-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(content); err != nil {
		os.Remove(tempFile.Name())
		t.Fatalf("Failed to write temp file: %v", err)
	}

	// 注册清理
	t.Cleanup(func() {
		os.Remove(tempFile.Name())
	})

	return tempFile.Name()
}

// CreateTempDir 创建临时目录
func CreateTempDir(t *testing.T) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "lc-e2e-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return tempDir
}

// ExecGit 执行 git 命令
func ExecGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\nOutput: %s", strings.Join(args, " "), err, output)
	}
}

// FindGitPath 查找 git 可执行文件路径
func FindGitPath(t *testing.T) string {
	paths := []string{"git", "/usr/bin/git", "/usr/local/bin/git"}
	for _, p := range paths {
		cmd := exec.Command(p, "--version")
		if err := cmd.Run(); err == nil {
			return p
		}
	}
	t.Fatal("git not found")
	return ""
}

// URLEncode 简单的 URL 编码
func URLEncode(s string) string {
	// 简单的编码，生产环境应使用 url.QueryEscape
	s = strings.ReplaceAll(s, "@", "%40")
	s = strings.ReplaceAll(s, "+", "%2B")
	s = strings.ReplaceAll(s, "=", "%3D")
	return s
}

// GetProjectRoot 查找项目根目录（包含 go.mod 的目录）
func GetProjectRoot() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	dir := filepath.Dir(filename)
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found")
}

// EnsureBinary 确保二进制文件已构建，如果没有则自动构建
func EnsureBinary() error {
	if BinaryPath != "" {
		// 检查文件是否存在
		if _, err := os.Stat(BinaryPath); err == nil {
			return nil
		}
	}

	projectRoot, err := GetProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	BinaryPath = filepath.Join(projectRoot, "bin", "lc")

	// 检查文件是否已存在
	if _, err := os.Stat(BinaryPath); err == nil {
		return nil
	}

	// 构建二进制文件
	cmd := exec.Command("make", "build")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build binary: %w", err)
	}

	return nil
}

// FormatFloat64 将 float64 格式化为字符串（去除小数部分）
func FormatFloat64(f float64) string {
	return fmt.Sprintf("%.0f", f)
}

// ExtractTextFromRichText 从富文本格式中提取纯文本
func ExtractTextFromRichText(value interface{}) string {
	if value == nil {
		return ""
	}

	// 处理 []interface{} 类型
	if arr, ok := value.([]interface{}); ok {
		var result string
		for _, item := range arr {
			if block, ok := item.(map[string]interface{}); ok {
				if children, ok := block["children"].([]interface{}); ok {
					for _, child := range children {
						if childMap, ok := child.(map[string]interface{}); ok {
							if text, ok := childMap["text"].(string); ok {
								result += text
							}
						}
					}
				}
			}
			result += "\n"
		}
		return strings.TrimSpace(result)
	}

	// 如果是字符串直接返回
	if s, ok := value.(string); ok {
		return s
	}

	return ""
}
