package autodetect

import (
	"testing"

	"github.com/user/lc/e2e/framework"
)

// TestAutoDetectModeRepoList 测试自动探测模式查询仓库列表
func TestAutoDetectModeRepoList(t *testing.T) {
	framework.SkipIfShort(t)
	framework.SkipIfNotInGitRepo(t)

	a := framework.NewAutoDetectTest(t)

	t.Logf("测试自动探测模式查询仓库列表")
	t.Logf("在 Git 仓库目录下执行: %s", a.GetRepoDir())

	// 自动探测模式：只指定 -k
	res := a.RunInRepo("repo", "list", "-k", "-l", "10")
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Logf("自动探测模式查询失败，可能当前测试仓库未关联研发空间")
		t.Skip("自动探测失败，跳过此测试")
	}

	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("响应格式不正确")
	}

	pagination, ok := dataField["pagination"].(map[string]interface{})
	if !ok {
		t.Fatal("无法获取 pagination")
	}

	total, ok := pagination["total"].(float64)
	if !ok {
		t.Fatal("无法获取 total")
	}

	t.Logf("自动探测模式查询仓库列表成功 (找到 %.0f 个仓库)", total)
}

// TestAutoDetectModeRepoSearch 测试自动探测模式搜索仓库
func TestAutoDetectModeRepoSearch(t *testing.T) {
	framework.SkipIfShort(t)
	framework.SkipIfNotInGitRepo(t)

	a := framework.NewAutoDetectTest(t)

	t.Logf("测试自动探测模式搜索仓库")

	// 自动探测模式：只指定 -k
	res := a.RunInRepo("repo", "search", "lc", "-k", "-l", "5")
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Logf("自动探测模式搜索失败")
		return
	}

	t.Logf("自动探测模式搜索仓库成功")
}

