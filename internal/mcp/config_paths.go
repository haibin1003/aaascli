package mcp

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func init() {
	configSearchPaths = buildSearchPaths()
}

// findGitRoot 查找最近的 git 仓库根目录
func findGitRoot(startDir string) string {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return ""
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// buildSearchPaths returns the ordered list of MCP config search paths
// for the current operating system.
func buildSearchPaths() []string {
	if runtime.GOOS == "windows" {
		return buildWindowsSearchPaths()
	}
	return buildUnixSearchPaths()
}

func buildUnixSearchPaths() []string {
	home, _ := os.UserHomeDir()
	expand := func(p string) string {
		if strings.HasPrefix(p, "~/") {
			return filepath.Join(home, p[2:])
		}
		return p
	}

	// 获取 git 根目录（如果存在）
	gitRoot := findGitRoot(".")

	paths := []string{
		// joinai-code 配置（优先级最高）
	}

	// 项目级 joinai-code 配置
	if gitRoot != "" {
		paths = append(paths,
			filepath.Join(gitRoot, "joinai-code.json"),
			filepath.Join(gitRoot, ".joinai-code", "joinai-code.json"),
		)
	}

	// 全局级 joinai-code 配置
	if home != "" {
		paths = append(paths,
			filepath.Join(home, ".config", "joinai-code", "joinai-code.json"),
		)
	}

	// 标准 MCP 配置路径
	paths = append(paths,
		expand("~/.config/modelcontextprotocol/mcp.json"), // Priority 3
		expand("~/.config/mcp/config.json"),               // Priority 4
		"./mcp.json",                                      // Priority 5: current directory
		"./.mcp/config.json",                              // Priority 6
		"/etc/mcp/config.json",                            // Priority 7: system-wide
	)

	return paths
}

func buildWindowsSearchPaths() []string {
	var paths []string

	// 获取 git 根目录（如果存在）
	gitRoot := findGitRoot(".")

	// 项目级 joinai-code 配置（优先级最高）
	if gitRoot != "" {
		paths = append(paths,
			filepath.Join(gitRoot, "joinai-code.json"),
			filepath.Join(gitRoot, ".joinai-code", "joinai-code.json"),
		)
	}

	// 全局级 joinai-code 配置
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths,
			filepath.Join(home, ".config", "joinai-code", "joinai-code.json"),
		)
	}

	// %APPDATA%\modelcontextprotocol\mcp.json  (Priority 3)
	// %APPDATA%\mcp\config.json                (Priority 4)
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		paths = append(paths,
			filepath.Join(appdata, "modelcontextprotocol", "mcp.json"),
			filepath.Join(appdata, "mcp", "config.json"),
		)
	}

	// %USERPROFILE%\.mcp\config.json           (Priority 5)
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".mcp", "config.json"))
	}

	// .\mcp.json          (Priority 6: current directory)
	// .\.mcp\config.json  (Priority 7)
	paths = append(paths,
		filepath.Join(".", "mcp.json"),
		filepath.Join(".", ".mcp", "config.json"),
	)

	// %ProgramData%\mcp\config.json            (Priority 8: system-wide)
	if progData := os.Getenv("ProgramData"); progData != "" {
		paths = append(paths, filepath.Join(progData, "mcp", "config.json"))
	}

	return paths
}
