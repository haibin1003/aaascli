// Package embed 用于嵌入静态文件到二进制中
package embed

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed helper-extension.zip
var helperExtensionZip embed.FS

// ExtractHelperExtension 将嵌入的浏览器扩展解压到指定目录
// targetDir: 目标目录，如果为空则默认解压到桌面
func ExtractHelperExtension(targetDir string) error {
	// 如果未指定目标目录，使用桌面
	if targetDir == "" {
		targetDir = getDesktopPath()
	}

	// 确保目标目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 读取嵌入的 zip 文件
	zipData, err := helperExtensionZip.ReadFile("helper-extension.zip")
	if err != nil {
		return fmt.Errorf("读取嵌入文件失败: %w", err)
	}

	// 检查是否是占位文件（小于 1KB 视为占位）
	if len(zipData) < 1024 {
		return fmt.Errorf("扩展文件尚未打包，请先运行 'make build-helper'")
	}

	// 打开 zip 文件
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("打开 zip 文件失败: %w", err)
	}

	// 解压文件
	extensionDir := filepath.Join(targetDir, "lc-login-helper-extension")
	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return fmt.Errorf("创建扩展目录失败: %w", err)
	}

	for _, file := range zipReader.File {
		// 跳过占位文件
		if file.Name == "placeholder.txt" {
			continue
		}

		targetPath := filepath.Join(extensionDir, file.Name)

		// 防止 zip slip 攻击
		if !stringsHasPrefix(filepath.Clean(targetPath), filepath.Clean(extensionDir)) {
			return fmt.Errorf("非法文件路径: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.Mode()); err != nil {
				return fmt.Errorf("创建目录失败 %s: %w", targetPath, err)
			}
			continue
		}

		// 创建父目录
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("创建父目录失败 %s: %w", filepath.Dir(targetPath), err)
		}

		// 创建文件
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("打开 zip 内文件失败 %s: %w", file.Name, err)
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("创建目标文件失败 %s: %w", targetPath, err)
		}

		_, err = io.Copy(dstFile, srcFile)
		dstFile.Close()
		if err != nil {
			return fmt.Errorf("写入文件失败 %s: %w", targetPath, err)
		}
	}

	return nil
}

// GetHelperExtensionSize 返回嵌入的 zip 文件大小（用于检查是否已打包）
func GetHelperExtensionSize() int64 {
	zipData, err := helperExtensionZip.ReadFile("helper-extension.zip")
	if err != nil {
		return 0
	}
	return int64(len(zipData))
}

// getDesktopPath 获取当前用户的桌面路径
func getDesktopPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Desktop")
	case "windows":
		return filepath.Join(homeDir, "Desktop")
	default: // linux and others
		// 尝试常见的桌面路径
		desktop := filepath.Join(homeDir, "Desktop")
		if _, err := os.Stat(desktop); err == nil {
			return desktop
		}
		// 尝试中文桌面
		desktop = filepath.Join(homeDir, "桌面")
		if _, err := os.Stat(desktop); err == nil {
			return desktop
		}
		return homeDir
	}
}

// stringsHasPrefix 检查 s 是否以 prefix 开头
func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
