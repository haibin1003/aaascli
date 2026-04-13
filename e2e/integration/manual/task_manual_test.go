package manual

import (
	"fmt"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// TestManualModeTaskCreate 测试手动参数模式创建任务
func TestManualModeTaskCreate(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)
	m.DisableReadonly() // 关闭只读模式以允许创建操作

	// 先创建需求
	reqName := fmt.Sprintf("手动模式任务测试需求_%s", framework.GenerateTimestamp())
	res := m.RunWithProject("req", "create", reqName)
	data := res.ExpectJSON()
	reqID := extractObjectID(data)

	if reqID == "" {
		t.Fatal("创建需求失败，无法测试创建任务")
	}

	t.Logf("创建需求成功: %s", reqID)

	// 创建任务
	taskName := fmt.Sprintf("手动模式测试任务_%s", framework.GenerateTimestamp())
	t.Logf("测试手动参数模式创建任务: %s", taskName)

	res = m.RunWithProject("task", "create", reqID, taskName)
	res.PrintOutput()

	taskData := res.ExpectJSON()
	if success, ok := taskData["success"].(bool); !ok || !success {
		t.Fatal("手动模式创建任务失败")
	}

	taskID := extractObjectID(taskData)
	t.Logf("手动模式创建任务成功 (Object ID: %s)", taskID)

	// 清理
	if taskID != "" {
		time.Sleep(1 * time.Second)
		deleteTask(t, m, taskID)
	}
	if reqID != "" {
		time.Sleep(1 * time.Second)
		deleteReq(t, m, reqID)
	}
}

// TestManualModeTaskList 测试手动参数模式查询任务列表
func TestManualModeTaskList(t *testing.T) {
	framework.SkipIfShort(t)

	m := framework.NewManualModeTest(t)

	t.Logf("测试手动参数模式查询任务列表")

	res := m.Run("task", "list", "-l", "10")
	res.PrintOutput()

	data := res.ExpectJSON()
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("查询任务列表失败: 响应格式不正确")
	}

	count, ok := dataField["count"].(float64)
	if !ok {
		t.Fatal("查询任务列表失败: 无法获取 count")
	}

	t.Logf("手动模式查询任务列表成功 (找到 %.0f 个任务)", count)
}

// deleteTask 删除任务
func deleteTask(t *testing.T, m *framework.ManualModeTest, objectID string) {
	t.Helper()
	res := m.Run("task", "delete", objectID)
	if res.IsSuccess() {
		t.Logf("删除任务成功: %s", objectID)
	} else {
		t.Logf("删除任务失败: %s", objectID)
	}
}
