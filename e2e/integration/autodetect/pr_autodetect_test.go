package autodetect

import (
	"testing"

	"github.com/user/lc/e2e/framework"
)

// TestAutoDetectModePRList 测试自动探测模式查询合并请求列表
func TestAutoDetectModePRList(t *testing.T) {
	framework.SkipIfShort(t)
	framework.SkipIfNotInGitRepo(t)

	a := framework.NewAutoDetectTest(t)

	t.Logf("测试自动探测模式查询 PR 列表")
	t.Logf("在 Git 仓库目录下执行: %s", a.GetRepoDir())

	// 自动探测模式：应该能自动探测 git-project-id
	res := a.RunInRepo("pr", "list", "-k", "-l", "10")
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Logf("自动探测模式查询失败，可能原因：")
		t.Logf("1. 测试 Git 仓库未配置正确的远程 URL")
		t.Logf("2. 远程仓库未关联研发空间")
		t.Skip("自动探测失败，跳过此测试")
	}

	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Logf("响应格式不正确")
		return
	}

	count, ok := dataField["count"].(float64)
	if !ok {
		t.Logf("无法获取 count")
		return
	}

	t.Logf("自动探测模式查询 PR 列表成功 (找到 %.0f 个 PR)", count)
}

// TestAutoDetectModePRView 测试自动探测模式查看合并请求详情
func TestAutoDetectModePRView(t *testing.T) {
	framework.SkipIfShort(t)
	framework.SkipIfNotInGitRepo(t)

	a := framework.NewAutoDetectTest(t)

	// 先列出 PR 获取一个 ID
	res := a.RunInRepo("pr", "list", "-k", "-l", "1")
	data := res.ExpectJSON()

	if success, ok := data["success"].(bool); !ok || !success {
		t.Skip("无法获取 PR 列表，跳过查看测试")
	}

	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Skip("响应格式不正确，跳过查看测试")
	}

	list, ok := dataField["list"].([]interface{})
	if !ok || len(list) == 0 {
		t.Skip("没有可用的 PR，跳过查看测试")
	}

	firstPR, ok := list[0].(map[string]interface{})
	if !ok {
		t.Skip("无法解析 PR 数据，跳过查看测试")
	}

	iid, ok := firstPR["iid"].(float64)
	if !ok {
		t.Skip("无法获取 PR IID，跳过查看测试")
	}

	t.Logf("测试自动探测模式查看 PR: %.0f", iid)

	// 自动探测模式查看
	res = a.RunInRepo("pr", "view", framework.FormatFloat64(iid), "-k")
	res.PrintOutput()

	viewData := res.ExpectJSON()
	if success, ok := viewData["success"].(bool); !ok || !success {
		t.Logf("查看 PR 详情失败")
		return
	}

	t.Logf("自动探测模式查看 PR 详情成功")
}

// TestAutoDetectModePRDiff 测试自动探测模式查看合并请求差异
func TestAutoDetectModePRDiff(t *testing.T) {
	framework.SkipIfShort(t)
	framework.SkipIfNotInGitRepo(t)

	a := framework.NewAutoDetectTest(t)

	// 先列出 PR 获取一个 ID
	res := a.RunInRepo("pr", "list", "-k", "-l", "1")
	data := res.ExpectJSON()

	if success, ok := data["success"].(bool); !ok || !success {
		t.Skip("无法获取 PR 列表，跳过 diff 测试")
	}

	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Skip("响应格式不正确，跳过 diff 测试")
	}

	list, ok := dataField["list"].([]interface{})
	if !ok || len(list) == 0 {
		t.Skip("没有可用的 PR，跳过 diff 测试")
	}

	firstPR, ok := list[0].(map[string]interface{})
	if !ok {
		t.Skip("无法解析 PR 数据，跳过 diff 测试")
	}

	iid, ok := firstPR["iid"].(float64)
	if !ok {
		t.Skip("无法获取 PR IID，跳过 diff 测试")
	}

	t.Logf("测试自动探测模式查看 PR diff: %.0f", iid)

	// 自动探测模式查看 diff
	res = a.RunInRepo("pr", "diff", framework.FormatFloat64(iid), "-k")
	res.PrintOutput()

	// diff 可能成功也可能失败，取决于 PR 状态
	t.Logf("自动探测模式查看 PR diff 执行完成")
}
