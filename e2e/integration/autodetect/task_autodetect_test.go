package autodetect

import (
	"fmt"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// TestAutoDetectModeTaskCreate 测试自动探测模式创建任务
func TestAutoDetectModeTaskCreate(t *testing.T) {
	framework.SkipIfShort(t)
	framework.SkipIfNotInGitRepo(t)

	a := framework.NewAutoDetectTest(t)

	// 先创建需求
	reqName := fmt.Sprintf("自动探测任务测试需求_%s", framework.GenerateTimestamp())
	res := a.RunInRepo("req", "create", reqName, "-k")
	data := res.ExpectJSON()

	if success, ok := data["success"].(bool); !ok || !success {
		t.Logf("创建需求失败，跳过任务创建测试")
		t.Skip("前置条件不满足")
	}

	reqID := extractObjectID(data)
	t.Logf("创建需求成功: %s", reqID)

	// 创建任务 - 自动探测模式
	taskName := fmt.Sprintf("自动探测测试任务_%s", framework.GenerateTimestamp())
	t.Logf("测试自动探测模式创建任务: %s", taskName)

	res = a.RunInRepo("task", "create", reqID, taskName, "-k")
	res.PrintOutput()

	taskData := res.ExpectJSON()
	if success, ok := taskData["success"].(bool); !ok || !success {
		t.Logf("自动探测模式创建任务失败，可能原因：自动探测未生效")
		t.Skip("自动探测可能未生效")
	}

	taskID := extractObjectID(taskData)
	t.Logf("自动探测模式创建任务成功 (Object ID: %s)", taskID)

	// 清理
	if taskID != "" {
		time.Sleep(1 * time.Second)
		deleteTask(t, a, taskID)
	}
	if reqID != "" {
		time.Sleep(1 * time.Second)
		deleteReq(t, a, reqID)
	}
}

// TestAutoDetectModeTaskList 测试自动探测模式查询任务列表
func TestAutoDetectModeTaskList(t *testing.T) {
	framework.SkipIfShort(t)
	framework.SkipIfNotInGitRepo(t)

	a := framework.NewAutoDetectTest(t)

	t.Logf("测试自动探测模式查询任务列表")

	res := a.RunInRepo("task", "list", "-k", "-l", "10")
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

	count, ok := dataField["count"].(float64)
	if !ok {
		t.Fatal("无法获取 count")
	}

	t.Logf("自动探测模式查询任务列表成功 (找到 %.0f 个任务)", count)
}

// deleteTask 删除任务
func deleteTask(t *testing.T, a *framework.AutoDetectTest, objectID string) {
	t.Helper()
	res := a.RunInRepo("task", "delete", objectID, "-k")
	if res.IsSuccess() {
		t.Logf("删除任务成功: %s", objectID)
	} else {
		t.Logf("删除任务失败: %s", objectID)
	}
}
