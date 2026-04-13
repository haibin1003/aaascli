package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/user/lc/internal/common"
)

// requireWorkspaceKey 检查并返回workspace key，如果未设置则退出
func requireWorkspaceKey() string {
	if spaceWorkspaceKey == "" {
		common.PrintError(fmt.Errorf("workspace key is required"))
		os.Exit(1)
	}
	return spaceWorkspaceKey
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// containsString 检查字符串切片是否包含指定字符串
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// formatObjectID 格式化ObjectID为短格式
func formatObjectID(objectID string) string {
	if len(objectID) <= 8 {
		return objectID
	}
	return objectID[:8] + "..."
}

// cleanString 清理字符串中的特殊字符
func cleanString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	// 合并多个空格
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}
