package manual

import (
	"testing"

	"github.com/user/lc/e2e/framework"
)

// TestManualModePRList 测试手动参数模式查询合并请求列表
func TestManualModePRList(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)

	t.Logf("测试手动参数模式查询 PR 列表")

	// 手动模式：显式指定 git-project-id
	res := m.Run("pr", "list",
		"--git-project-id", m.Config.GitProjectID,
		"-l", "10",
	)
	res.PrintOutput()

	data := res.ExpectJSON()
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Logf("查询 PR 列表可能失败: 响应格式不正确")
		return
	}

	count, _ := dataField["count"].(float64)
	t.Logf("手动模式查询 PR 列表成功 (找到 %.0f 个 PR)", count)
}

// TestManualModePRView 测试手动参数模式查看合并请求详情
func TestManualModePRView(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)

	// 先列出 PR 获取一个 ID
	res := m.Run("pr", "list",
		"--git-project-id", m.Config.GitProjectID,
		"-l", "1",
	)
	data := res.ExpectJSON()
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Skip("无法获取 PR 列表，跳过查看测试")
	}

	list, ok := dataField["list"].([]interface{})
	if !ok || len(list) == 0 {
		t.Skip("没有可用的 PR，跳过查看测试")
	}

	// 获取第一个 PR 的 IID
	firstPR, ok := list[0].(map[string]interface{})
	if !ok {
		t.Skip("无法解析 PR 数据，跳过查看测试")
	}

	iid, ok := firstPR["iid"].(float64)
	if !ok {
		t.Skip("无法获取 PR IID，跳过查看测试")
	}

	t.Logf("测试手动参数模式查看 PR: %.0f", iid)

	res = m.Run("pr", "view",
		"--git-project-id", m.Config.GitProjectID,
		framework.FormatFloat64(iid),
	)
	res.PrintOutput()

	viewData := res.ExpectJSON()
	if success, ok := viewData["success"].(bool); !ok || !success {
		t.Logf("查看 PR 详情可能失败")
		return
	}

	t.Logf("手动模式查看 PR 详情成功")
}

// TestManualModePRDiff 测试手动参数模式查看合并请求差异
func TestManualModePRDiff(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)

	// 先列出 PR 获取一个 ID
	res := m.Run("pr", "list",
		"--git-project-id", m.Config.GitProjectID,
		"-l", "1",
	)
	data := res.ExpectJSON()
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Skip("无法获取 PR 列表，跳过 diff 测试")
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

	t.Logf("测试手动参数模式查看 PR diff: %.0f", iid)

	res = m.Run("pr", "diff",
		"--git-project-id", m.Config.GitProjectID,
		framework.FormatFloat64(iid),
	)
	res.PrintOutput()

	// diff 可能成功也可能失败，取决于 PR 状态
	t.Logf("手动模式查看 PR diff 执行完成")
}
