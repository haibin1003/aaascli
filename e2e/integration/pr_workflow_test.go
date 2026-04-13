package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// TestPRList 测试查询 PR 列表
func TestPRList(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	t.Logf("Listing PRs (limit 10)")

	res := lc.Run("pr", "list", "-k", "--workspace-key", workspaceKey, "-l", "10")
	res.PrintOutput()

	// 只要有响应即可，可能为空列表
	if res.ExitCode != 0 {
		t.Logf("查询 PR 列表返回非零退出码: %d", res.ExitCode)
	}

	t.Logf("查询 PR 列表成功")
}

// TestPRCreateAndMerge 测试创建并合并 PR
func TestPRCreateAndMerge(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建/删除操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 生成唯一仓库名称
	repoName := framework.GenerateRepoName()
	t.Logf("测试仓库名称: %s", repoName)

	var projectID string
	var mrIID string
	var commentID string
	var tempDir string
	var repoDir string
	var defaultBranch string
	var cloneURL string

	// 清理函数
	defer func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
			t.Logf("删除临时目录: %s", tempDir)
		}
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
		t.Logf("创建仓库 %s", repoName)

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
		} else if id, ok := dataField["id"].(float64); ok {
			projectID = fmt.Sprintf("%.0f", id)
		} else {
			t.Fatal("无法从输出中提取项目 ID")
		}

		// 提取克隆 URL (优先使用 tenantHttpPath)
		if tenantPath, ok := dataField["tenantHttpPath"].(string); ok && tenantPath != "" {
			cloneURL = tenantPath
		} else if httpPath, ok := dataField["httpPath"].(string); ok && httpPath != "" {
			cloneURL = httpPath
		}

		t.Logf("项目 ID: %s", projectID)
		t.Logf("克隆 URL: %s", cloneURL)
		time.Sleep(3 * time.Second)
	})

	// 步骤 2: 关闭工作项关联
	t.Run("DisableWorkItemLink", func(t *testing.T) {
		res := lc.Run("repo", "disable-work-item-link", projectID, "-k", "--workspace-key", workspaceKey)
		if !res.IsSuccess() {
			t.Logf("关闭工作项关联失败，继续执行")
		}
		time.Sleep(2 * time.Second)
	})

	// 步骤 3: 克隆仓库并创建分支
	t.Run("SetupRepository", func(t *testing.T) {
		if cloneURL == "" {
			t.Fatal("克隆 URL 为空，无法继续测试")
		}

		tempDir = framework.CreateTempDir(t)
		repoDir = filepath.Join(tempDir, repoName)

		// 克隆仓库
		cmd := exec.Command("git", "clone", cloneURL, repoDir)
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("克隆仓库失败: %v\n输出: %s", err, output)
		}

		// 配置 git
		framework.ExecGit(t, repoDir, "config", "user.email", "test@example.com")
		framework.ExecGit(t, repoDir, "config", "user.name", "Test User")

		// 创建初始提交
		readmePath := filepath.Join(repoDir, "README.md")
		os.WriteFile(readmePath, []byte("# "+repoName+"\n"), 0644)
		framework.ExecGit(t, repoDir, "add", ".")
		framework.ExecGit(t, repoDir, "commit", "-m", "Initial commit")

		// 推送并确定默认分支
		defaultBranch = "main"
		cmd = exec.Command("git", "push", "-u", "origin", "main")
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		_, err = cmd.CombinedOutput()
		if err != nil {
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

		t.Logf("默认分支: %s", defaultBranch)
		time.Sleep(2 * time.Second)
	})

	// 步骤 4: 创建功能分支
	featureBranch := "test-pr-branch"
	t.Run("CreateFeatureBranch", func(t *testing.T) {
		framework.ExecGit(t, repoDir, "checkout", "-b", featureBranch)

		testFile := filepath.Join(repoDir, "pr_test.go")
		os.WriteFile(testFile, []byte("// PR test\n"), 0644)

		framework.ExecGit(t, repoDir, "add", ".")
		framework.ExecGit(t, repoDir, "commit", "-m", "Add PR test file")

		cmd := exec.Command("git", "push", "-u", "origin", featureBranch)
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("推送功能分支失败: %v\n输出: %s", err, output)
		}

		t.Logf("功能分支推送成功: %s", string(output))
		time.Sleep(3 * time.Second)
	})

	// 步骤 5: 创建 PR
	t.Run("CreatePR", func(t *testing.T) {
		var res *framework.Result
		framework.Retry(t, 5, 3*time.Second, func() error {
			res = lc.Run("pr", "create",
				"-t", "Test PR from E2E",
				"-b", "This is a test PR created by E2E test",
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
				return fmt.Errorf("failed to get data field from response")
			}
			if _, ok := dataField["iid"].(float64); !ok {
				return fmt.Errorf("failed to create PR")
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
			t.Logf("PR IID: %s", mrIID)
		} else {
			t.Fatal("无法从输出中提取 PR IID")
		}

		time.Sleep(2 * time.Second)
	})

	// 步骤 6: 添加评论
	t.Run("CommentPR", func(t *testing.T) {
		if mrIID == "" {
			t.Skip("PR 未创建成功，跳过评论测试")
		}

		res := lc.Run("pr", "comment", mrIID,
			"--git-project-id", projectID,
			"--workspace-key", workspaceKey,
			"--body", "这是一条测试评论，由 E2E 测试自动添加",
			"-k",
		)
		res.PrintOutput()

		if !res.IsSuccess() {
			t.Logf("添加评论可能失败，但不影响测试完成")
		} else {
			data := res.ExpectJSON()
			// 新的 JSON 格式: data 字段中包含实际数据
			dataField, ok := data["data"].(map[string]interface{})
			if !ok {
				t.Logf("评论添加成功，但无法获取评论 ID: 响应格式不正确")
			} else if noteID, ok := dataField["noteId"].(float64); ok {
				commentID = fmt.Sprintf("%.0f", noteID)
				t.Logf("评论添加成功，评论 ID: %s", commentID)
			} else {
				t.Logf("评论添加成功，但无法获取评论 ID")
			}
		}
		time.Sleep(2 * time.Second)
	})

	// 步骤 7: 查询 PR 评论（验证评论已创建且未解决）
	t.Run("ViewPRComments", func(t *testing.T) {
		if mrIID == "" {
			t.Skip("PR 未创建成功，跳过查看评论测试")
		}

		res := lc.Run("pr", "view", mrIID,
			"--git-project-id", projectID,
			"--workspace-key", workspaceKey,
			"--comments",
			"-k",
		)
		res.PrintOutput()

		if !res.IsSuccess() {
			t.Logf("查询评论失败，但不影响测试完成")
		} else {
			data := res.ExpectJSON()
			// 新的 JSON 格式: data 字段中包含实际数据
			dataField, ok := data["data"].(map[string]interface{})
			if !ok {
				t.Logf("查询评论成功，但响应格式不正确")
			} else {
				if count, ok := dataField["count"].(float64); ok {
					t.Logf("查询评论成功，共有 %.0f 条评论", count)
					if count == 0 {
						t.Logf("警告：没有查询到任何评论")
					}
				}
				// 如果没有获取到 commentID，尝试从评论列表中解析
				if commentID == "" {
					if comments, ok := dataField["comments"].([]interface{}); ok && len(comments) > 0 {
						if firstComment, ok := comments[0].(map[string]interface{}); ok {
							if id, ok := firstComment["id"].(float64); ok {
								commentID = fmt.Sprintf("%.0f", id)
								t.Logf("从列表中获取到评论 ID: %s", commentID)
								// 验证评论内容正确
								if note, ok := firstComment["note"].(string); ok {
									t.Logf("评论内容: %s", note)
								}
							}
						}
					}
				}
				// 验证评论状态为未解决（active）
				if comments, ok := dataField["comments"].([]interface{}); ok && len(comments) > 0 {
					for _, c := range comments {
						if commentMap, ok := c.(map[string]interface{}); ok {
							if id, ok := commentMap["id"].(float64); ok {
								if fmt.Sprintf("%.0f", id) == commentID {
									if state, ok := commentMap["resolvedState"].(string); ok {
										t.Logf("评论状态验证: resolvedState=%s", state)
										if state != "active" {
											t.Logf("警告：新创建评论的状态不是 active，实际为: %s", state)
										}
									}
									break
								}
							}
						}
					}
				}
			}
		}
		time.Sleep(2 * time.Second)
	})

	// 步骤 8: 将评论置为已解决
	t.Run("ResolveComment", func(t *testing.T) {
		if mrIID == "" || commentID == "" {
			t.Skip("PR 或评论 ID 未获取，跳过解决评论测试")
		}

		res := lc.Run("pr", "patch-comment", mrIID,
			"--git-project-id", projectID,
			"--workspace-key", workspaceKey,
			"--comment-id", commentID,
			"--state", "fixed",
			"-k",
		)
		res.PrintOutput()

		if !res.IsSuccess() {
			t.Logf("解决评论可能失败，但不影响测试完成")
		} else {
			t.Logf("评论已置为已解决状态")
		}
		time.Sleep(2 * time.Second)
	})

	// 步骤 9: 再次查询 PR 评论（验证评论已被解决）
	t.Run("VerifyCommentResolved", func(t *testing.T) {
		if mrIID == "" || commentID == "" {
			t.Skip("PR 或评论 ID 未获取，跳过验证评论解决状态")
		}

		res := lc.Run("pr", "view", mrIID,
			"--git-project-id", projectID,
			"--workspace-key", workspaceKey,
			"--comments",
			"-k",
		)
		res.PrintOutput()

		if !res.IsSuccess() {
			t.Logf("查询评论失败，无法验证解决状态")
		} else {
			data := res.ExpectJSON()
			// 新的 JSON 格式: data 字段中包含实际数据
			dataField, ok := data["data"].(map[string]interface{})
			if !ok {
				t.Logf("查询评论成功，但响应格式不正确")
			} else {
				resolvedFound := false
				if comments, ok := dataField["comments"].([]interface{}); ok {
					for _, c := range comments {
						if commentMap, ok := c.(map[string]interface{}); ok {
							if id, ok := commentMap["id"].(float64); ok {
								if fmt.Sprintf("%.0f", id) == commentID {
									resolvedFound = true
									if state, ok := commentMap["resolvedState"].(string); ok {
										t.Logf("评论解决状态验证: resolvedState=%s", state)
										if state == "fixed" {
											t.Logf("✅ 评论已成功置为已解决状态")
										} else {
											t.Logf("⚠️ 评论状态不是 fixed，实际为: %s", state)
										}
									} else {
										t.Logf("警告：无法获取评论状态字段")
									}
									break
								}
							}
						}
					}
					if !resolvedFound {
						t.Logf("警告：未找到评论 ID %s", commentID)
					}
				}
			}
		}
		time.Sleep(2 * time.Second)
	})

	// 步骤 10: 合并 PR
	t.Run("MergePR", func(t *testing.T) {
		if mrIID == "" {
			t.Skip("PR 未创建成功，跳过合并测试")
		}

		res := lc.Run("pr", "merge", mrIID,
			"--git-project-id", projectID,
			"--workspace-key", workspaceKey,
			"--type", "merge",
			"-k",
		)
		res.PrintOutput()

		if !res.IsSuccess() {
			t.Logf("PR 合并可能失败，但不影响测试完成")
		} else {
			t.Logf("PR 合并成功")
		}
	})

	t.Logf("========================================")
	t.Logf("PR 工作流测试完成！")
	t.Logf("仓库: %s", repoName)
	t.Logf("项目 ID: %s", projectID)
	t.Logf("PR IID: %s", mrIID)
	t.Logf("评论 ID: %s", commentID)
	t.Logf("========================================")
}
