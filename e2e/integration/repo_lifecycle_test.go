package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// TestRepoLifecycleFull 测试完整的仓库生命周期
// 包括: 创建仓库 -> 禁用卡片 -> 克隆 -> 推送分支 -> 创建MR -> 合并MR -> 删除仓库
func TestRepoLifecycleFull(t *testing.T) {
	framework.SkipIfShort(t)

	// E2E 测试强制使用小白测研发空间 (XXJSxiaobaice)
	// 详细限制参见 e2e/main_test.go TestMain 函数
	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", "XXJSxiaobaice")
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建/删除操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 生成唯一仓库名称
	repoName := framework.GenerateRepoName()
	t.Logf("生成仓库名称: %s", repoName)

	var projectID string
	var mrIID string
	var tempDir string
	var repoDir string
	var defaultBranch string
	var cloneURL string

	// 清理函数
	defer func() {
		// 删除临时目录
		if tempDir != "" {
			os.RemoveAll(tempDir)
			t.Logf("删除临时目录: %s", tempDir)
		}
		// 删除远程仓库
		if projectID != "" {
			t.Logf("删除远程仓库 (Project ID: %s)...", projectID)
			res := lc.Run("repo", "delete", projectID, "-k", "--workspace-key", workspaceKey)
			if res.IsSuccess() {
				t.Logf("远程仓库删除成功: %s", repoName)
			} else {
				t.Logf("远程仓库删除失败: %s\n输出: %s", repoName, res.GetStdout())
			}
		}
	}()

	// 步骤 1: 创建仓库
	t.Run("CreateRepository", func(t *testing.T) {
		t.Logf("步骤 1: 创建仓库 %s", repoName)

		var res *framework.Result
		groupID := framework.GetEnvOrDefault("LC_GROUP_ID", defaultGroupID)
		framework.Retry(t, 3, 2*time.Second, func() error {
			res = lc.Run("repo", "create", repoName, "-k", "--workspace-key", workspaceKey, "--group-id", groupID)
			if res.ExitCode != 0 {
				return fmt.Errorf("exit code %d", res.ExitCode)
			}
			return nil
		})

		res.PrintOutput()

		data := res.ExpectJSON()
		// 新的 JSON 格式: data 字段中包含实际数据
		dataField, ok := data["data"].(map[string]interface{})
		if !ok {
			t.Fatal("无法从响应中获取 data 字段")
		}
		if gitProjectID, ok := dataField["gitProjectId"].(float64); ok {
			projectID = fmt.Sprintf("%.0f", gitProjectID)
			t.Logf("提取到项目 ID: %s", projectID)
		} else if id, ok := dataField["id"].(float64); ok {
			projectID = fmt.Sprintf("%.0f", id)
			t.Logf("提取到项目 ID: %s", projectID)
		} else {
			t.Fatal("无法从输出中提取项目 ID")
		}

		// 提取克隆 URL (优先使用 tenantHttpPath)
		if tenantPath, ok := dataField["tenantHttpPath"].(string); ok && tenantPath != "" {
			cloneURL = tenantPath
			t.Logf("提取到克隆 URL: %s", cloneURL)
		} else if httpPath, ok := dataField["httpPath"].(string); ok && httpPath != "" {
			cloneURL = httpPath
			t.Logf("提取到克隆 URL: %s", cloneURL)
		}

		time.Sleep(3 * time.Second) // 等待仓库初始化
	})

	// 步骤 2: 关闭工作项关联
	t.Run("DisableWorkItemLink", func(t *testing.T) {
		t.Logf("步骤 2: 关闭工作项关联功能")

		var res *framework.Result
		framework.Retry(t, 3, 2*time.Second, func() error {
			res = lc.Run("repo", "disable-work-item-link", projectID, "-k", "--workspace-key", workspaceKey)
			return nil
		})

		res.PrintOutput()
		if !res.IsSuccess() {
			t.Fatal("关闭工作项关联失败")
		}

		time.Sleep(2 * time.Second)
	})

	// 步骤 3: 克隆仓库并设置
	t.Run("CloneAndSetupRepo", func(t *testing.T) {
		t.Logf("步骤 3: 克隆仓库")

		if cloneURL == "" {
			t.Fatal("克隆 URL 为空，无法继续测试")
		}

		tempDir = framework.CreateTempDir(t)
		repoDir = filepath.Join(tempDir, repoName)

		t.Logf("使用克隆 URL: %s", cloneURL)

		// 克隆仓库
		cmd := exec.Command("git", "clone", cloneURL, repoDir)
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("克隆仓库失败: %v\n输出: %s", err, output)
		}
		t.Logf("仓库克隆到: %s", repoDir)

		// 配置 git
		framework.ExecGit(t, repoDir, "config", "user.email", "test@example.com")
		framework.ExecGit(t, repoDir, "config", "user.name", "Test User")

		// 创建初始提交
		readmePath := filepath.Join(repoDir, "README.md")
		if err := os.WriteFile(readmePath, []byte("# "+repoName+"\n"), 0644); err != nil {
			t.Fatalf("创建 README 失败: %v", err)
		}

		framework.ExecGit(t, repoDir, "add", ".")
		framework.ExecGit(t, repoDir, "commit", "-m", "Initial commit")

		// 推送 main 分支
		defaultBranch = "main"
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		output, err = cmd.CombinedOutput()
		if err != nil {
			// 尝试 master 分支
			defaultBranch = "master"
			framework.ExecGit(t, repoDir, "branch", "-m", "master")
			cmd = exec.Command("git", "push", "-u", "origin", "master")
			cmd.Dir = repoDir
			cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
			output, err = cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("推送分支失败: %v\n输出: %s", err, output)
			}
		}

		t.Logf("初始提交推送成功，默认分支: %s", defaultBranch)
		time.Sleep(2 * time.Second)
	})

	// 步骤 4: 创建功能分支并推送
	featureBranch := "test-feature-branch"
	t.Run("CreateFeatureBranch", func(t *testing.T) {
		t.Logf("步骤 4: 创建功能分支 %s", featureBranch)

		// 创建并切换到功能分支
		framework.ExecGit(t, repoDir, "checkout", "-b", featureBranch)

		// 创建测试文件
		testFile := filepath.Join(repoDir, "test_feature.go")
		if err := os.WriteFile(testFile, []byte("// Test feature\n"), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		framework.ExecGit(t, repoDir, "add", ".")
		framework.ExecGit(t, repoDir, "commit", "-m", "Add test feature")

		// 推送功能分支
		cmd := exec.Command("git", "push", "-u", "origin", featureBranch)
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("推送功能分支失败: %v\n输出: %s", err, output)
		}

		t.Logf("功能分支推送成功")
		time.Sleep(3 * time.Second)
	})

	// 步骤 5: 创建合并请求
	t.Run("CreateMergeRequest", func(t *testing.T) {
		t.Logf("步骤 5: 创建合并请求")

		var res *framework.Result
		framework.Retry(t, 5, 3*time.Second, func() error {
			res = lc.Run("pr", "create",
				"-t", "Test MR from E2E Test",
				"-b", "This is a test merge request created by Go E2E test",
				"-s", featureBranch,
				"--target", defaultBranch,
				"--git-project-id", projectID,
				"--workspace-key", workspaceKey,
				"-k",
			)
			data := res.ExpectJSON()
			// 新的 JSON 格式: data 字段中包含实际数据
			dataField, ok := data["data"].(map[string]interface{})
			if !ok {
				if strings.Contains(res.GetStdout(), "源分支不存在") {
					return fmt.Errorf("branch not found")
				}
			} else if _, ok := dataField["iid"].(float64); !ok {
				if strings.Contains(res.GetStdout(), "源分支不存在") {
					return fmt.Errorf("branch not found")
				}
			}
			return nil
		})

		res.PrintOutput()
		data := res.ExpectJSON()
		// 新的 JSON 格式: data 字段中包含实际数据
		dataField, ok := data["data"].(map[string]interface{})
		if !ok {
			t.Fatal("无法从响应中获取 data 字段")
		}
		if iid, ok := dataField["iid"].(float64); ok {
			mrIID = fmt.Sprintf("%.0f", iid)
			t.Logf("MR IID: %s", mrIID)
		} else {
			t.Fatal("无法从输出中提取 MR IID")
		}

		time.Sleep(2 * time.Second)
	})

	// 步骤 6: 合并合并请求
	t.Run("MergeMergeRequest", func(t *testing.T) {
		t.Logf("步骤 6: 合并合并请求 #%s", mrIID)

		var res *framework.Result
		framework.Retry(t, 3, 2*time.Second, func() error {
			res = lc.Run("pr", "merge", mrIID,
				"--git-project-id", projectID,
				"--workspace-key", workspaceKey,
				"--type", "merge",
				"-k",
			)
			return nil
		})

		res.PrintOutput()
		if !res.IsSuccess() {
			t.Fatal("合并请求失败")
		}
	})

	t.Logf("========================================")
	t.Logf("端到端测试完成！")
	t.Logf("仓库名称: %s", repoName)
	t.Logf("项目 ID: %s", projectID)
	t.Logf("MR IID: %s", mrIID)
	t.Logf("========================================")
}
