//go:build windows

package cmd

import (
	"golang.org/x/sys/windows"
)

func init() {
	// 设置 Windows 控制台输出为 UTF-8，避免中文乱码
	windows.SetConsoleOutputCP(65001)
}
