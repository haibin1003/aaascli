package framework

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// InjectOTPConfig 向 CLI 的隔离 HOME 目录注入 OTP 配置。
// 保留原有 cookie 等配置，仅添加/覆盖 OTP 部分。
//
// secret 为空时禁用 OTP（设置 enabled=false）。
// protectedCmds 为 nil 时不写入 protectedCommands 字段，使用服务端默认列表。
func InjectOTPConfig(t *testing.T, cli *CLI, secret string, protectedCmds []string) {
	t.Helper()
	configPath := filepath.Join(cli.GetHome(), ".lc", "config.json")

	var cfgMap map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, &cfgMap)
	}
	if cfgMap == nil {
		cfgMap = make(map[string]interface{})
	}

	if secret == "" {
		// 禁用 OTP
		cfgMap["otp"] = map[string]interface{}{
			"enabled": false,
		}
	} else {
		otpCfg := map[string]interface{}{
			"enabled":              true,
			"secret":               secret,
			"sessionExpiryMinutes": 5,
		}
		if len(protectedCmds) > 0 {
			otpCfg["protectedCommands"] = protectedCmds
		}
		cfgMap["otp"] = otpCfg
	}

	data, err := json.MarshalIndent(cfgMap, "", "  ")
	if err != nil {
		t.Fatalf("序列化 OTP 配置失败: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("写入 OTP 配置失败: %v", err)
	}

	if secret == "" {
		t.Logf("已禁用 OTP 配置")
	} else {
		t.Logf("已注入测试 OTP 配置 (secret: %s...)", secret[:8])
	}
}
