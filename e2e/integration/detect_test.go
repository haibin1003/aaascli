package integration

import (
	"os"
	"testing"

	"github.com/user/lc/e2e/framework"
)

// TestDetectInGitRepo 测试在 Git 仓库中探测上下文
func TestDetectInGitRepo(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)

	t.Logf("测试在 Git 仓库中探测上下文")

	// 当前目录应该是 Git 仓库
	res := lc.Run("detect", "-k")
	res.PrintOutput()

	data := res.ExpectJSON()

	// 验证基本结构
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatalf("detect 命令失败: %v", data)
	}

	result, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("响应中缺少 data 字段")
	}

	// 验证 matched 字段
	matched, ok := result["matched"].(bool)
	if !ok {
		t.Fatal("响应中缺少 matched 字段")
	}

	if !matched {
		t.Logf("警告: 未能匹配到远程仓库，可能当前目录不是 lc 仓库")
		return
	}

	// 验证关键字段存在
	if workspaceKey, ok := result["workspaceKey"].(string); !ok || workspaceKey == "" {
		t.Error("workspaceKey 字段为空或不存在")
	} else {
		t.Logf("Workspace Key: %s", workspaceKey)
	}

	if workspaceName, ok := result["workspaceName"].(string); !ok || workspaceName == "" {
		t.Logf("警告: workspaceName 字段为空")
	} else {
		t.Logf("Workspace Name: %s", workspaceName)
	}

	if tenantId, ok := result["tenantId"].(string); !ok || tenantId == "" {
		t.Error("tenantId 字段为空或不存在")
	} else {
		t.Logf("Tenant ID: %s", tenantId)
	}

	// 验证 repository 字段
	repo, ok := result["repository"].(map[string]interface{})
	if !ok {
		t.Error("repository 字段为空或格式不正确")
	} else {
		if gitProjectId, ok := repo["gitProjectId"]; ok {
			t.Logf("Git Project ID: %v", gitProjectId)
		}
		if repoName, ok := repo["name"].(string); ok {
			t.Logf("Repository Name: %s", repoName)
		}
	}

	// 验证 gitInfo 字段
	gitInfo, ok := result["gitInfo"].(map[string]interface{})
	if !ok {
		t.Error("gitInfo 字段为空或格式不正确")
	} else {
		if isGitRepo, ok := gitInfo["IsGitRepo"].(bool); !ok || !isGitRepo {
			t.Error("IsGitRepo 应该为 true")
		}
		if repoName, ok := gitInfo["RepoName"].(string); ok {
			t.Logf("Git Repo Name: %s", repoName)
		}
	}

	t.Logf("detect 命令测试通过")
}

// TestDetectWithPath 测试指定路径探测
func TestDetectWithPath(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)

	// 使用当前工作目录（应该是 Git 仓库）
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前目录失败: %v", err)
	}

	t.Logf("测试指定路径探测: %s", currentDir)

	res := lc.Run("detect", "--path", currentDir, "-k")
	res.PrintOutput()

	data := res.ExpectJSON()

	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatalf("detect 命令失败: %v", data)
	}

	result := data["data"].(map[string]interface{})

	// 验证 gitInfo 中的路径
	gitInfo, ok := result["gitInfo"].(map[string]interface{})
	if !ok {
		t.Fatal("gitInfo 字段不存在")
	}

	if currentPath, ok := gitInfo["CurrentPath"].(string); ok {
		if currentPath != currentDir {
			t.Errorf("CurrentPath 不匹配: 期望 %s, 实际 %s", currentDir, currentPath)
		}
	}

	t.Logf("指定路径探测测试通过")
}

// TestDetectInNonGitDir 测试在非 Git 目录中探测
func TestDetectInNonGitDir(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)

	// 使用 /tmp 目录（通常不是 Git 仓库）
	tmpDir := "/tmp"

	t.Logf("测试在非 Git 目录探测: %s", tmpDir)

	res := lc.Run("detect", "--path", tmpDir, "-k")
	res.PrintOutput()

	data := res.ExpectJSON()

	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatalf("detect 命令失败: %v", data)
	}

	result := data["data"].(map[string]interface{})

	// 验证 matched 为 false
	matched, ok := result["matched"].(bool)
	if !ok {
		t.Fatal("响应中缺少 matched 字段")
	}

	if matched {
		t.Error("在非 Git 目录中 matched 应该为 false")
	}

	// 验证 gitInfo 中的 IsGitRepo 为 false
	gitInfo, ok := result["gitInfo"].(map[string]interface{})
	if !ok {
		t.Fatal("gitInfo 字段不存在")
	}

	if isGitRepo, ok := gitInfo["IsGitRepo"].(bool); ok && isGitRepo {
		t.Error("在非 Git 目录中 IsGitRepo 应该为 false")
	}

	t.Logf("非 Git 目录探测测试通过")
}

// TestDetectSpaceDetails 测试研发空间详情获取
func TestDetectSpaceDetails(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)

	t.Logf("测试研发空间详情获取")

	res := lc.Run("detect", "-k")
	res.PrintOutput()

	data := res.ExpectJSON()
	result := data["data"].(map[string]interface{})

	// 如果匹配成功，验证 spaceDetails
	if matched, ok := result["matched"].(bool); ok && matched {
		spaceDetails, ok := result["spaceDetails"].(map[string]interface{})
		if !ok {
			t.Logf("警告: spaceDetails 字段为空")
			return
		}

		if spaceName, ok := spaceDetails["spaceName"].(string); ok {
			t.Logf("Space Name: %s", spaceName)
		}

		if spaceDesc, ok := spaceDetails["spaceDesc"].(string); ok {
			t.Logf("Space Description: %s", spaceDesc)
		}
	}

	t.Logf("研发空间详情获取测试通过")
}

// TestDetectHelp 测试帮助信息
func TestDetectHelp(t *testing.T) {
	lc := framework.NewCLI(t)

	t.Logf("测试 detect 命令帮助信息")

	res := lc.Run("detect", "--help")

	output := res.Stdout + res.Stderr

	// 验证帮助信息中包含关键内容
	expectedStrings := []string{
		"detect",
		"检测",
		"workspaceKey",
		"脚本自动化",
		"--path",
	}

	for _, str := range expectedStrings {
		if !contains(output, str) {
			t.Errorf("帮助信息中缺少: %s", str)
		}
	}

	t.Logf("帮助信息测试通过")
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestDetectConsistency 测试多次执行结果一致性
func TestDetectConsistency(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)

	t.Logf("测试多次执行结果一致性")

	// 执行两次
	res1 := lc.Run("detect", "-k")
	data1 := res1.ExpectJSON()

	res2 := lc.Run("detect", "-k")
	data2 := res2.ExpectJSON()

	result1 := data1["data"].(map[string]interface{})
	result2 := data2["data"].(map[string]interface{})

	// 验证 workspaceKey 一致
	wk1, _ := result1["workspaceKey"].(string)
	wk2, _ := result2["workspaceKey"].(string)

	if wk1 != wk2 {
		t.Errorf("两次执行 workspaceKey 不一致: %s vs %s", wk1, wk2)
	}

	// 验证 matched 一致
	m1, _ := result1["matched"].(bool)
	m2, _ := result2["matched"].(bool)

	if m1 != m2 {
		t.Errorf("两次执行 matched 不一致: %v vs %v", m1, m2)
	}

	t.Logf("多次执行一致性测试通过")
}

// TestDetectWorkspaceKeyMatch 测试 workspaceKey 与仓库 spaceCode 匹配
func TestDetectWorkspaceKeyMatch(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)

	t.Logf("测试 workspaceKey 与仓库 spaceCode 匹配")

	res := lc.Run("detect", "-k")
	data := res.ExpectJSON()
	result := data["data"].(map[string]interface{})

	if matched, ok := result["matched"].(bool); !ok || !matched {
		t.Skip("未匹配到仓库，跳过此测试")
	}

	workspaceKey, _ := result["workspaceKey"].(string)

	repo, ok := result["repository"].(map[string]interface{})
	if !ok {
		t.Fatal("repository 字段不存在")
	}

	spaceCode, ok := repo["spaceCode"].(string)
	if !ok {
		t.Fatal("repository.spaceCode 字段不存在")
	}

	if workspaceKey != spaceCode {
		t.Errorf("workspaceKey (%s) 与 repository.spaceCode (%s) 不匹配", workspaceKey, spaceCode)
	}

	t.Logf("workspaceKey 与 spaceCode 匹配: %s", workspaceKey)
}
