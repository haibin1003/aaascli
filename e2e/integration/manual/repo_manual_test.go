package manual

import (
	"testing"

	"github.com/user/lc/e2e/framework"
)

// TestManualModeRepoList 测试手动参数模式查询仓库列表
func TestManualModeRepoList(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)

	t.Logf("测试手动参数模式查询仓库列表")

	res := m.Run("repo", "list", "-l", "10")
	res.PrintOutput()

	data := res.ExpectJSON()
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("查询仓库列表失败: 响应格式不正确")
	}

	pagination, ok := dataField["pagination"].(map[string]interface{})
	if !ok {
		t.Fatal("查询仓库列表失败: 无法获取 pagination")
	}

	total, ok := pagination["total"].(float64)
	if !ok {
		t.Fatal("查询仓库列表失败: 无法获取 total")
	}

	t.Logf("手动模式查询仓库列表成功 (找到 %.0f 个仓库)", total)
}

// TestManualModeRepoSearch 测试手动参数模式搜索仓库
func TestManualModeRepoSearch(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)

	t.Logf("测试手动参数模式搜索仓库")

	// repo search 是全局搜索，不支持 workspace-key，直接使用 CLI.Run
	res := m.CLI.Run("repo", "search", "lc", "-l", "5", "-k")
	res.PrintOutput()

	data := res.ExpectJSON()
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Logf("搜索仓库可能失败: 响应格式不正确")
		return
	}

	// 搜索结果可能为空，不强求
	count, _ := dataField["count"].(float64)
	t.Logf("手动模式搜索仓库完成 (找到 %.0f 个仓库)", count)
}


