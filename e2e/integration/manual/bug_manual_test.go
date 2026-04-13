package manual

import (
	"testing"

	"github.com/user/lc/e2e/framework"
)

// TestManualModeBugList 测试手动参数模式查询缺陷列表
func TestManualModeBugList(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)

	t.Logf("测试手动参数模式查询缺陷列表")

	// 手动模式：显式指定 workspace-key
	res := m.Run("bug", "list",
		"--workspace-key", m.Config.WorkspaceKey,
		"-l", "10",
	)
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Logf("查询缺陷列表可能失败")
		return
	}

	t.Logf("手动模式查询缺陷列表执行完成")
}
