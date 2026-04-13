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

//go:embed skills.zip
var skillsZip embed.FS

// DefaultSkillsDir 默认技能安装目录
func DefaultSkillsDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	var baseDir string
	switch runtime.GOOS {
	case "darwin", "linux":
		baseDir = os.Getenv("XDG_CONFIG_HOME")
		if baseDir == "" {
			baseDir = filepath.Join(homeDir, ".config")
		}
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = filepath.Join(homeDir, "AppData", "Roaming")
		}
	default: // others
		baseDir = os.Getenv("XDG_CONFIG_HOME")
		if baseDir == "" {
			baseDir = filepath.Join(homeDir, ".config")
		}
	}

	return filepath.Join(baseDir, "joinai-code", "skills")
}

// ExtractSkills 将嵌入的技能解压到指定目录
// targetDir: 目标目录，如果为空则使用默认目录
func ExtractSkills(targetDir string) error {
	// 如果未指定目标目录，使用默认目录
	if targetDir == "" {
		targetDir = DefaultSkillsDir()
		if targetDir == "" {
			return fmt.Errorf("无法确定默认技能目录")
		}
	}

	// 确保目标目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 读取嵌入的 zip 文件
	zipData, err := skillsZip.ReadFile("skills.zip")
	if err != nil {
		return fmt.Errorf("读取嵌入文件失败: %w", err)
	}

	// 检查是否是占位文件（小于 1KB 视为占位）
	if len(zipData) < 1024 {
		return fmt.Errorf("技能文件尚未打包，请先运行 'make build-skills'")
	}

	// 打开 zip 文件
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("打开 zip 文件失败: %w", err)
	}

	// 解压文件
	for _, file := range zipReader.File {
		// 跳过占位文件
		if file.Name == "placeholder.txt" {
			continue
		}

		targetPath := filepath.Join(targetDir, file.Name)

		// 防止 zip slip 攻击
		if !stringsHasPrefix(filepath.Clean(targetPath), filepath.Clean(targetDir)) {
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

// GetSkillsSize 返回嵌入的 zip 文件大小（用于检查是否已打包）
func GetSkillsSize() int64 {
	zipData, err := skillsZip.ReadFile("skills.zip")
	if err != nil {
		return 0
	}
	return int64(len(zipData))
}