package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	npmRegistryURL = "https://registry.npmjs.org/@lingji/lc/latest"
	npmPackageName = "@lingji/lc"
)

// npmPackageInfo 存储 npm 返回的包信息
type npmPackageInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
	Time struct {
		Modified string `json:"modified"`
	} `json:"time"`
	Repository struct {
		URL string `json:"url"`
	} `json:"repository"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "检查更新",
	Long: `检查灵畿 CLI 助手是否有新版本。

此命令会查询 npm registry 获取最新版本，并与当前版本进行比较。
如果发现有新版本，会提示用户如何更新。`,
	Run: func(cmd *cobra.Command, args []string) {
		checkUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func checkUpdate() {
	// 获取当前版本
	currentVersion := version
	if currentVersion == "dev" || currentVersion == "" {
		currentVersion = "0.0.0-dev"
	}

	// 获取最新版本
	latestVersion, info, err := fetchLatestVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 检查更新失败: %v\n", err)
		os.Exit(1)
	}

	// 比较版本
	comparison := compareVersions(currentVersion, latestVersion)

	fmt.Printf("当前版本: %s\n", currentVersion)
	fmt.Printf("最新版本: %s\n", latestVersion)

	if info.Time.Modified != "" {
		fmt.Printf("发布时间: %s\n", formatTime(info.Time.Modified))
	}

	if comparison == -1 {
		fmt.Println("\n⚠️  发现新版本!")
		fmt.Println("请运行: npm install -g @lingji/lc 更新")
	} else if comparison == 0 {
		fmt.Println("\n✅ 已是最新版本")
	} else {
		fmt.Println("\n✨ 当前版本比官方版本还新 (可能是开发版本)")
	}
}

// fetchLatestVersion 从 npm registry 获取最新版本信息
func fetchLatestVersion() (string, *npmPackageInfo, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(npmRegistryURL)
	if err != nil {
		return "", nil, fmt.Errorf("无法连接到 npm registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("npm registry 返回错误: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var info npmPackageInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return "", nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return info.Version, &info, nil
}

// compareVersions 比较两个版本号
// 返回: -1 表示 v1 < v2, 0 表示相等, 1 表示 v1 > v2
func compareVersions(v1, v2 string) int {
	// 去除前缀 v
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// 去除 -dev, -beta 等后缀进行比较
	v1Clean := strings.Split(v1, "-")[0]
	v2Clean := strings.Split(v2, "-")[0]

	parts1 := strings.Split(v1Clean, ".")
	parts2 := strings.Split(v2Clean, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int

		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	// 主版本号相同，检查后缀
	v1HasSuffix := strings.Contains(v1, "-")
	v2HasSuffix := strings.Contains(v2, "-")

	if v1HasSuffix && !v2HasSuffix {
		return -1 // v1 是开发版，v2 是正式版
	}
	if !v1HasSuffix && v2HasSuffix {
		return 1 // v1 是正式版，v2 是开发版
	}

	return 0
}

// formatTime 格式化时间字符串
func formatTime(t string) string {
	// 尝试解析 ISO8601 格式
	if parsed, err := time.Parse(time.RFC3339, t); err == nil {
		return parsed.Format("2006-01-02 15:04:05")
	}
	return t
}
