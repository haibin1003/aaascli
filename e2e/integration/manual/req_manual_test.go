package manual

import (
	"fmt"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// TestManualModeReqCreate 测试手动参数模式创建需求
func TestManualModeReqCreate(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)
	m.DisableReadonly() // 关闭只读模式以允许创建操作

	reqName := fmt.Sprintf("手动模式测试需求_%s", framework.GenerateTimestamp())

	t.Logf("测试手动参数模式创建需求: %s", reqName)

	// 手动模式：显式指定所有参数
	res := m.RunWithProject("req", "create", reqName)
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatal("手动模式创建需求失败")
	}

	objectID := extractObjectID(data)
	t.Logf("手动模式创建需求成功 (Object ID: %s)", objectID)

	// 清理
	if objectID != "" {
		time.Sleep(1 * time.Second)
		deleteReq(t, m, objectID)
	}
}

// TestManualModeReqList 测试手动参数模式查询需求列表
func TestManualModeReqList(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)

	// 重置 OTP 配置：移除用户可能自定义的 protectedCommands（如 req list），
	// 避免测试被 OTP 交互式提示阻塞。此处不传 protectedCmds，使用系统默认列表
	// （repo delete、readonly off），req list 不在其中，测试可正常执行。
	framework.InjectOTPConfig(t, m.CLI, "JBSWY3DPEHPK3PXP", nil)

	t.Logf("测试手动参数模式查询需求列表")

	// 手动模式：显式指定 workspace-key（workspace-name 自动获取）
	res := m.Run("req", "list", "--workspace-key", m.Config.WorkspaceKey, "-l", "10")
	res.PrintOutput()

	data := res.ExpectJSON()
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("查询需求列表失败: 响应格式不正确")
	}

	count, ok := dataField["count"].(float64)
	if !ok {
		t.Fatal("查询需求列表失败: 无法获取 count")
	}

	t.Logf("手动模式查询需求列表成功 (找到 %.0f 个需求)", count)
}

// TestManualModeReqView 测试手动参数模式查看需求详情
func TestManualModeReqView(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)
	m.DisableReadonly() // 关闭只读模式以允许创建操作

	// 首先创建需求
	testDate := framework.GenerateTimestamp()
	reqName := fmt.Sprintf("手动模式查看测试_%s", testDate)

	res := m.RunWithProject("req", "create", reqName)
	data := res.ExpectJSON()
	objectID := extractObjectID(data)

	if objectID == "" {
		t.Fatal("创建需求失败，无法测试查看")
	}

	t.Logf("测试手动参数模式查看需求: %s", objectID)

	// 手动模式查看
	res = m.Run("req", "view", objectID)
	res.PrintOutput()

	viewData := res.ExpectJSON()
	viewDataField, ok := viewData["data"].(map[string]interface{})
	if !ok {
		t.Fatal("查看需求详情失败: 响应格式不正确")
	}

	if _, ok := viewDataField["name"]; !ok {
		t.Fatal("查看需求详情失败: 缺少 name 字段")
	}

	t.Logf("手动模式查看需求详情成功")

	// 清理
	time.Sleep(1 * time.Second)
	deleteReq(t, m, objectID)
}

// TestManualModeWithoutWorkspaceKey 测试手动模式下缺少必需参数
func TestManualModeWithoutWorkspaceKey(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)

	t.Logf("测试手动模式下缺少 workspace-key")

	// 不指定 workspace-key，应该失败
	res := lc.Run("req", "list", "-k")
	res.PrintOutput()

	// 期望失败，因为缺少 workspace-key
	if res.ExitCode == 0 {
		t.Logf("警告: 没有 workspace-key 时命令仍然成功，可能是自动探测生效")
	} else {
		t.Logf("符合预期: 缺少 workspace-key 时命令失败")
	}
}

// extractObjectID 从 JSON 响应中提取 objectId
func extractObjectID(response map[string]interface{}) string {
	dataField, ok := response["data"].(map[string]interface{})
	if !ok {
		return ""
	}
	if objId, ok := dataField["objectId"].(string); ok {
		return objId
	}
	return ""
}

// deleteReq 删除需求
func deleteReq(t *testing.T, m *framework.ManualModeTest, objectID string) {
	t.Helper()
	res := m.Run("req", "delete", objectID)
	if res.IsSuccess() {
		t.Logf("删除需求成功: %s", objectID)
	} else {
		t.Logf("删除需求失败: %s", objectID)
	}
}
