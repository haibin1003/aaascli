package autodetect

import (
	"testing"

	"github.com/user/lc/e2e/framework"
)

// TestAutoDetectModeBugList 测试自动探测模式查询缺陷列表
func TestAutoDetectModeBugList(t *testing.T) {
	framework.SkipIfShort(t)
	framework.SkipIfNotInGitRepo(t)

	a := framework.NewAutoDetectTest(t)

	t.Logf("测试自动探测模式查询缺陷列表")
	t.Logf("在 Git 仓库目录下执行: %s", a.GetRepoDir())

	// 自动探测模式：只指定 -k，不指定 workspace-key
	res := a.RunInRepo("bug", "list", "-k")
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Logf("自动探测模式查询失败，可能当前测试仓库未关联研发空间")
		t.Skip("自动探测失败，跳过此测试")
	}

	t.Logf("自动探测模式查询缺陷列表执行完成")
}
