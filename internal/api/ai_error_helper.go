package api

import (
	"fmt"
	"strings"
)

// FormatAIError generates a concise AI prompt for login/session errors
func FormatAIError(desc string) error {
	// Check if this is a session/login error
	if strings.Contains(desc, "登录") || strings.Contains(desc, "失效") || strings.Contains(desc, "过期") ||
		strings.Contains(desc, "Session") || strings.Contains(desc, "session") {
		return fmt.Errorf(`登录平台失败，需要重新认证。请安装灵畿CLI助手Chrome插件，在灵畿平台页面，点击插件图标，复制认证命令，在Console执行或者发给您的AI让他执行`)
	}
	return fmt.Errorf("API error: %s", desc)
}
