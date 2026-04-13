package framework

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Result 存储命令执行结果
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	t        *testing.T
}

// ExpectExit 断言退出码
func (r *Result) ExpectExit(code int) *Result {
	r.t.Helper()
	if r.ExitCode != code {
		r.t.Errorf("Expected exit code %d, got %d\nStdout: %s\nStderr: %s",
			code, r.ExitCode, r.Stdout, r.Stderr)
	}
	return r
}

// ExpectSuccess 断言成功退出（退出码为 0）
func (r *Result) ExpectSuccess() *Result {
	return r.ExpectExit(0)
}

// MustSucceed 确保结果成功，否则 Fatal
func (r *Result) MustSucceed() *Result {
	r.t.Helper()
	if r.ExitCode != 0 {
		r.t.Fatalf("Expected success, got exit code %d\nStdout: %s\nStderr: %s",
			r.ExitCode, r.Stdout, r.Stderr)
	}
	return r
}

// ExpectFailure 断言失败退出（退出码非 0）
func (r *Result) ExpectFailure() *Result {
	r.t.Helper()
	if r.ExitCode == 0 {
		r.t.Errorf("Expected non-zero exit code, got 0\nStdout: %s\nStderr: %s",
			r.Stdout, r.Stderr)
	}
	return r
}

// ExpectContains 断言输出包含指定文本
func (r *Result) ExpectContains(text string) *Result {
	r.t.Helper()
	combined := r.Stdout + r.Stderr
	if !strings.Contains(combined, text) {
		r.t.Errorf("Expected output to contain %q, but it didn't\nStdout: %s\nStderr: %s",
			text, r.Stdout, r.Stderr)
	}
	return r
}

// ExpectNotContains 断言输出不包含指定文本
func (r *Result) ExpectNotContains(text string) *Result {
	r.t.Helper()
	combined := r.Stdout + r.Stderr
	if strings.Contains(combined, text) {
		r.t.Errorf("Expected output NOT to contain %q, but it did\nStdout: %s\nStderr: %s",
			text, r.Stdout, r.Stderr)
	}
	return r
}

// ExpectStdoutContains 断言 stdout 包含指定文本
func (r *Result) ExpectStdoutContains(text string) *Result {
	r.t.Helper()
	if !strings.Contains(r.Stdout, text) {
		r.t.Errorf("Expected stdout to contain %q, but it didn't\nStdout: %s",
			text, r.Stdout)
	}
	return r
}

// ExpectStderrContains 断言 stderr 包含指定文本
func (r *Result) ExpectStderrContains(text string) *Result {
	r.t.Helper()
	if !strings.Contains(r.Stderr, text) {
		r.t.Errorf("Expected stderr to contain %q, but it didn't\nStderr: %s",
			text, r.Stderr)
	}
	return r
}

// AssertOutputContains 断言输出包含所有指定的文本
func (r *Result) AssertOutputContains(texts ...string) *Result {
	r.t.Helper()
	combined := r.Stdout + r.Stderr
	for _, text := range texts {
		if !contains(combined, text) {
			r.t.Errorf("Expected output to contain %q, but it didn't\nOutput: %s", text, combined)
		}
	}
	return r
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && strings.Contains(s, substr)
}

// ExpectJSON 断言输出是有效的 JSON 并返回解析后的数据
func (r *Result) ExpectJSON() map[string]interface{} {
	r.t.Helper()
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(r.Stdout), &data); err != nil {
		r.t.Errorf("Expected valid JSON output, but parsing failed: %v\nStdout: %s",
			err, r.Stdout)
	}
	return data
}

// ExpectJSONPath 断言 JSON 路径的值等于预期值
func (r *Result) ExpectJSONPath(path string, expected interface{}) *Result {
	r.t.Helper()
	data := r.ExpectJSON()
	actual := getJSONValue(data, path)
	if actual != expected {
		r.t.Errorf("Expected JSON path %q to be %v, got %v\nStdout: %s",
			path, expected, actual, r.Stdout)
	}
	return r
}

// getJSONValue 从 JSON 数据中获取路径值（支持简单的点号分隔路径）
func getJSONValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data
	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	return nil
}

// ExtractObjectID 从 JSON 输出中提取 objectId
func (r *Result) ExtractObjectID() string {
	data := r.ExpectJSON()
	if objId, ok := data["objectId"].(string); ok {
		return objId
	}
	return ""
}

// ExtractJSONField 从 JSON 输出中提取指定字段
func (r *Result) ExtractJSONField(field string) interface{} {
	data := r.ExpectJSON()
	return data[field]
}

// IsSuccess 检查 JSON 响应是否 success=true
func (r *Result) IsSuccess() bool {
	data := r.ExpectJSON()
	if success, ok := data["success"].(bool); ok {
		return success
	}
	return false
}

// PrintOutput 打印命令输出（用于调试）
func (r *Result) PrintOutput() *Result {
	fmt.Printf("=== Exit Code: %d ===\n", r.ExitCode)
	fmt.Printf("=== Stdout ===\n%s\n", r.Stdout)
	fmt.Printf("=== Stderr ===\n%s\n", r.Stderr)
	return r
}

// GetStdout 返回标准输出
func (r *Result) GetStdout() string {
	return r.Stdout
}

// GetStderr 返回标准错误
func (r *Result) GetStderr() string {
	return r.Stderr
}

// GetExitCode 返回退出码
func (r *Result) GetExitCode() int {
	return r.ExitCode
}
