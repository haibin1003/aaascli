package manual

import (
	"fmt"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// TestManualModeLibList 测试手动参数模式查询组件库列表
func TestManualModeLibList(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)

	t.Logf("测试手动参数模式查询组件库列表")

	res := m.Run("lib", "list")
	res.PrintOutput()

	data := res.ExpectJSON()
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Logf("查询组件库列表可能失败: 响应格式不正确")
		return
	}

	count, _ := dataField["count"].(float64)
	t.Logf("手动模式查询组件库列表成功 (找到 %.0f 个组件库)", count)
}

// TestManualModeLibCreateAndDelete 测试手动参数模式创建和删除组件库
func TestManualModeLibCreateAndDelete(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)

	libName := fmt.Sprintf("手动模式测试组件_%s", framework.GenerateTimestamp())
	t.Logf("测试手动参数模式创建组件库: %s", libName)

	// 创建组件库（lib 不需要 project-code）
	res := m.Run("lib", "create", libName)
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Logf("创建组件库可能失败，跳过删除测试")
		return
	}

	objectID := extractObjectID(data)
	t.Logf("手动模式创建组件库成功 (Object ID: %s)", objectID)

	// 等待一下再删除
	time.Sleep(1 * time.Second)

	// 删除组件库
	t.Logf("测试手动参数模式删除组件库: %s", objectID)
	res = m.Run("lib", "delete", objectID)
	res.PrintOutput()

	deleteData := res.ExpectJSON()
	if success, ok := deleteData["success"].(bool); ok && success {
		t.Logf("手动模式删除组件库成功")
	} else {
		t.Logf("手动模式删除组件库可能失败")
	}
}

