// Package common provides shared utilities for command execution.
package common

import (
	"fmt"
	"sort"
	"time"

	"github.com/user/lc/internal/config"
)

// isReadonlyByConfig 使用已加载的 Config 检查只读状态，不重新读盘。
// 若检测到临时关闭已过期，会就地恢复只读并保存配置。
func isReadonlyByConfig(cfg *config.Config) bool {
	if cfg.Readonly {
		return true
	}
	if cfg.TempOffExpireAt != nil && time.Now().After(*cfg.TempOffExpireAt) {
		cfg.Readonly = true
		cfg.TempOffExpireAt = nil
		config.SaveConfig(cfg)
		return true
	}
	return false
}

// IsReadonly 检查当前是否处于只读模式（自动加载配置）
// 默认返回 true（安全模式），只有在明确关闭只读模式后才返回 false
func IsReadonly() bool {
	cfg, err := config.LoadConfigWithDefaults(config.GetDefaultConfigPath())
	if err != nil {
		return true // 默认安全，出错时返回只读
	}
	return isReadonlyByConfig(cfg)
}

// CheckReadonlyForWrite 检查命令是否被只读模式拦截（自动加载配置）
// 只有 CommandRegistry 中 IsWrite=true 的命令才会被拦截；读操作直接放行
func CheckReadonlyForWrite(cmdName string) error {
	meta, ok := CommandRegistry[cmdName]
	if !ok || !meta.IsWrite {
		return nil
	}
	if IsReadonly() {
		return NewReadonlyError(cmdName)
	}
	return nil
}

// CheckReadonlyWithConfig 使用已加载的 Config 检查只读拦截，不重新读盘。
// 供 executeWithWrapper 等已持有 Config 的调用方使用。
func CheckReadonlyWithConfig(cfg *config.Config, cmdName string) error {
	meta, ok := CommandRegistry[cmdName]
	if !ok || !meta.IsWrite {
		return nil
	}
	if isReadonlyByConfig(cfg) {
		return NewReadonlyError(cmdName)
	}
	return nil
}

// ReadonlyError 只读模式错误
type ReadonlyError struct {
	Command string
}

// NewReadonlyError 创建只读模式错误
func NewReadonlyError(cmd string) *ReadonlyError {
	return &ReadonlyError{Command: cmd}
}

// Error 实现 error 接口
func (e *ReadonlyError) Error() string {
	return fmt.Sprintf("当前处于只读模式，命令 '%s' 被禁止", e.Command)
}

// GetReadonlyStatus 获取只读模式状态的详细描述
func GetReadonlyStatus(cfg *config.Config) map[string]interface{} {
	if cfg.Readonly {
		// 从 CommandRegistry 动态生成允许/拦截的命令列表，与注册表保持同步
		var writeCmds, readCmds []string
		for name, meta := range CommandRegistry {
			if meta.IsWrite {
				writeCmds = append(writeCmds, name)
			} else {
				readCmds = append(readCmds, name)
			}
		}
		sort.Strings(writeCmds)
		sort.Strings(readCmds)
		return map[string]interface{}{
			"readonly":          true,
			"description":       "当前处于只读模式，禁止执行创建、更新、删除等操作",
			"writable_commands": readCmds,
			"readonly_commands": writeCmds,
		}
	}

	// 非只读模式
	result := map[string]interface{}{
		"readonly":    false,
		"description": "当前处于读写模式，可以执行所有操作",
		"warning":     "请谨慎执行删除和修改操作，数据变更将直接生效",
	}

	// 添加临时关闭信息
	if cfg.TempOffExpireAt != nil {
		remaining := time.Until(*cfg.TempOffExpireAt)
		if remaining > 0 {
			result["temporary"] = true
			result["expireAt"] = cfg.TempOffExpireAt.Format(time.RFC3339)
			result["remainingMinutes"] = int(remaining.Minutes())
			result["description"] = fmt.Sprintf(
				"当前处于临时读写模式，将在 %d 分钟后自动恢复只读",
				int(remaining.Minutes()))
		}
	}

	return result
}

// ParseDuration 解析时长字符串，返回 time.Duration
// 支持格式: 10m, 30m, 1h, 2h
// 限制: 最小 1m，最大 24h
func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("时长不能为空")
	}

	// 使用 time.ParseDuration 解析
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("无效时长格式 %q，支持格式如: 10m, 30m, 1h", s)
	}

	// 验证范围
	if d < time.Minute {
		return 0, fmt.Errorf("时长不能小于 1 分钟")
	}
	if d > 24*time.Hour {
		return 0, fmt.Errorf("时长不能大于 24 小时")
	}

	return d, nil
}
