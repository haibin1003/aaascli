package common

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAutoDetectError(t *testing.T) {
	err := NewAutoDetectError("test error")
	require.NotNil(t, err)
	assert.Equal(t, "test error", err.Message)
}

func TestAutoDetectError_WithDetails(t *testing.T) {
	err := NewAutoDetectError("test error").
		WithDetails("additional details")

	assert.Equal(t, "additional details", err.Details)
}

func TestAutoDetectError_WithSuggestion(t *testing.T) {
	err := NewAutoDetectError("test error").
		WithSuggestion("try this")

	assert.Equal(t, "try this", err.Suggestion)
}

func TestAutoDetectError_Error(t *testing.T) {
	err := NewAutoDetectError("test error").
		WithDetails("details").
		WithSuggestion("suggestion")

	errStr := err.Error()
	assert.Contains(t, errStr, "test error")
	assert.Contains(t, errStr, "details")
	// Note: Suggestion is not included in Error() output, only Message and Details
}

func TestHandleAutoDetectError(t *testing.T) {
	originalErr := NewAutoDetectError("original error")
	handledErr := HandleAutoDetectError(originalErr, "-w, --workspace-key")

	autoErr, ok := handledErr.(*AutoDetectError)
	require.True(t, ok)
	assert.Contains(t, autoErr.Suggestion, "-w, --workspace-key")
}

func TestHandleAutoDetectError_NonAutoDetectError(t *testing.T) {
	originalErr := assert.AnError
	handledErr := HandleAutoDetectError(originalErr)

	// 非AutoDetectError会被包装
	assert.NotNil(t, handledErr)
}

func TestMakeAutoDetectFunc(t *testing.T) {
	var targetVar string
	fields := []AutoDetectField{
		{FlagName: "test-flag", TargetVar: &targetVar, ContextKey: "WorkspaceKey"},
	}

	fn := MakeAutoDetectFunc(fields)
	require.NotNil(t, fn)

	// 创建测试命令
	cmd := &cobra.Command{}
	cmd.Flags().String("test-flag", "", "test flag")

	// 设置标志值（不触发自动探测）
	cmd.Flags().Set("test-flag", "test-value")

	// 执行函数（忽略错误，只验证不panic）
	_ = fn(cmd)
	// 可能会失败因为没有Git仓库，但不会panic
	assert.NotPanics(t, func() { _ = fn(cmd) })
}

func TestAutoDetectField_Struct(t *testing.T) {
	var target string
	field := AutoDetectField{
		FlagName:   "workspace-key",
		TargetVar:  &target,
		ContextKey: "WorkspaceKey",
	}

	assert.Equal(t, "workspace-key", field.FlagName)
	assert.Equal(t, "WorkspaceKey", field.ContextKey)
	assert.Equal(t, &target, field.TargetVar)
}

func TestAutoDetectCache(t *testing.T) {
	// 清除缓存
	ClearAutoDetectCache()

	// 第一次调用应该执行探测
	result1 := TryAutoDetect(false)

	// 第二次调用应该返回缓存结果
	result2 := TryAutoDetect(false)

	// 由于我们没有Git仓库，两者都应该失败但结果一致
	assert.Equal(t, result1.Success, result2.Success)
}
