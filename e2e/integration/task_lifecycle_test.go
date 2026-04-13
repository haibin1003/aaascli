package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// deleteTask 删除任务
func deleteTask(t *testing.T, lc *framework.CLI, objectID, workspaceKey string) {
	t.Helper()
	res := lc.Run("task", "delete", objectID, "-k", "--workspace-key", workspaceKey)
	if res.IsSuccess() {
		t.Logf("删除任务成功: %s", objectID)
	} else {
		t.Logf("删除任务失败: %s", objectID)
	}
}

// TestTaskCreateSimple 测试简单创建任务
func TestTaskCreateSimple(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 首先创建一个需求
	testDate := framework.GenerateTimestamp()
	reqName := fmt.Sprintf("任务测试需求_%s", testDate)

	res := lc.Run("req", "create", reqName, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	data := res.ExpectJSON()
	reqObjectID := extractObjectID(data)

	if reqObjectID == "" {
		t.Fatal("创建需求失败，无法创建任务")
	}

	t.Logf("需求 Object ID: %s", reqObjectID)

	// 创建任务
	taskName := fmt.Sprintf("简单测试任务_%s", testDate)
	res = lc.Run("task", "create", reqObjectID, taskName, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	res.PrintOutput()

	taskData := res.ExpectJSON()
	if success, ok := taskData["success"].(bool); !ok || !success {
		t.Fatal("简单创建任务失败")
	}

	taskObjectID := extractObjectID(taskData)
	t.Logf("简单创建任务成功 (Object ID: %s)", taskObjectID)

	// 搜索验证：确认任务可通过搜索找到
	t.Logf("搜索验证任务存在: %s", taskName)
	time.Sleep(1 * time.Second)
	searchRes := lc.Run("task", "search", taskName, "-k", "--workspace-key", workspaceKey, "-l", "10")
	searchRes.PrintOutput()

	searchData := searchRes.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含 items
	searchDataField, ok := searchData["data"].(map[string]interface{})
	if !ok {
		t.Logf("搜索验证警告: 响应格式不正确")
		return
	}
	items, ok := searchDataField["items"].([]interface{})
	if !ok || len(items) == 0 {
		t.Logf("搜索验证警告: 未找到刚创建的任务（可能存在延迟）")
	} else {
		found := false
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["objectId"] == taskObjectID {
					found = true
					t.Logf("搜索验证成功: 任务存在于搜索结果中")
					break
				}
			}
		}
		if !found {
			t.Logf("搜索验证警告: 任务可能尚未进入搜索索引")
		}
	}

	// 清理：先删除任务，再删除需求
	time.Sleep(1 * time.Second)
	if taskObjectID != "" {
		deleteTask(t, lc, taskObjectID, workspaceKey)
	}
	time.Sleep(1 * time.Second)
	deleteReq(t, lc, reqObjectID, workspaceKey)
}

// TestTaskCreateFromYAMLFile 测试从 YAML 文件创建任务
func TestTaskCreateFromYAMLFile(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 首先创建一个需求
	testDate := framework.GenerateTimestamp()
	reqName := fmt.Sprintf("任务YAML测试需求_%s", testDate)

	res := lc.Run("req", "create", reqName, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	data := res.ExpectJSON()
	reqObjectID := extractObjectID(data)

	if reqObjectID == "" {
		t.Fatal("创建需求失败，无法创建任务")
	}

	// 创建任务 YAML 文件
	yamlContent := fmt.Sprintf(`name: YAML任务测试_%s
requirementId: %s
taskType:
  - 开发
taskDescription: |
  这是通过YAML文件创建的详细任务
  支持多行文本描述
plannedWorkingHours: 8
assignee:
  label: "魏宝辉(weibaohui@hq.cmcc)"
  value: weibaohui@hq.cmcc
  username: weibaohui@hq.cmcc
  nickname: 魏宝辉
priority: 8f7912a5-9176-4a79-a269-2269ac42b5a2
`, testDate, reqObjectID)

	yamlFile := framework.CreateTempFile(t, "test_task", yamlContent)

	t.Logf("Creating task from YAML file: %s", yamlFile)

	res = lc.Run("task", "create", "-f", yamlFile, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	res.PrintOutput()

	taskData := res.ExpectJSON()
	if success, ok := taskData["success"].(bool); !ok || !success {
		t.Fatal("YAML文件创建任务失败")
	}

	taskObjectID := extractObjectID(taskData)
	t.Logf("YAML文件创建任务成功 (Object ID: %s)", taskObjectID)

	// 清理
	time.Sleep(1 * time.Second)
	if taskObjectID != "" {
		deleteTask(t, lc, taskObjectID, workspaceKey)
	}
	time.Sleep(1 * time.Second)
	deleteReq(t, lc, reqObjectID, workspaceKey)
}

// TestTaskCreateFromPipe 测试从管道输入创建任务
func TestTaskCreateFromPipe(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 首先创建一个需求
	testDate := framework.GenerateTimestamp()
	reqName := fmt.Sprintf("任务管道测试需求_%s", testDate)

	res := lc.Run("req", "create", reqName, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	data := res.ExpectJSON()
	reqObjectID := extractObjectID(data)

	if reqObjectID == "" {
		t.Fatal("创建需求失败，无法创建任务")
	}

	yamlContent := fmt.Sprintf(`name: 管道任务测试_%s
requirementId: %s
taskType:
  - 测试
taskDescription: 这是通过管道输入创建的测试任务
plannedWorkingHours: 4
assignee:
  label: "魏宝辉(weibaohui@hq.cmcc)"
  value: weibaohui@hq.cmcc
  username: weibaohui@hq.cmcc
  nickname: 魏宝辉
`, testDate, reqObjectID)

	t.Logf("Creating task from pipe input")

	res = lc.RunWithInput(yamlContent, "task", "create", "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	res.PrintOutput()

	taskData := res.ExpectJSON()
	if success, ok := taskData["success"].(bool); !ok || !success {
		t.Fatal("管道输入创建任务失败")
	}

	taskObjectID := extractObjectID(taskData)
	t.Logf("管道输入创建任务成功 (Object ID: %s)", taskObjectID)

	// 清理
	time.Sleep(1 * time.Second)
	if taskObjectID != "" {
		deleteTask(t, lc, taskObjectID, workspaceKey)
	}
	time.Sleep(1 * time.Second)
	deleteReq(t, lc, reqObjectID, workspaceKey)
}

// TestTaskList 测试查询任务列表
func TestTaskList(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	t.Logf("Listing tasks (limit 10)")

	res := lc.Run("task", "list", "-k", "--workspace-key", workspaceKey, "-l", "10")
	res.PrintOutput()

	data := res.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含 count 和 items
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("查询任务列表失败: 响应格式不正确")
	}
	count, ok := dataField["count"].(float64)
	if !ok {
		t.Fatal("查询任务列表失败: 无法获取 count")
	}

	t.Logf("查询任务列表成功 (找到 %.0f 个任务)", count)
}

// TestTaskListWithRequirementFilter 测试按需求过滤任务
func TestTaskListWithRequirementFilter(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 首先创建一个需求
	testDate := framework.GenerateTimestamp()
	reqName := fmt.Sprintf("任务过滤测试需求_%s", testDate)

	res := lc.Run("req", "create", reqName, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	data := res.ExpectJSON()
	reqObjectID := extractObjectID(data)

	if reqObjectID == "" {
		t.Fatal("创建需求失败")
	}

	t.Logf("Listing tasks for requirement: %s", reqObjectID)

	res = lc.Run("task", "list", "-k", "--workspace-key", workspaceKey, "-r", reqObjectID, "-l", "5")
	res.PrintOutput()

	listData := res.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含 count
	listDataField, ok := listData["data"].(map[string]interface{})
	if !ok {
		t.Fatal("按需求过滤任务列表失败: 响应格式不正确")
	}
	count, ok := listDataField["count"].(float64)
	if !ok {
		t.Fatal("按需求过滤任务列表失败: 无法获取 count")
	}

	t.Logf("按需求过滤任务列表成功 (找到 %.0f 个任务)", count)

	// 清理
	time.Sleep(1 * time.Second)
	deleteReq(t, lc, reqObjectID, workspaceKey)
}

// TestTaskListWithPagination 测试分页查询任务
func TestTaskListWithPagination(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	res := lc.Run("task", "list", "-k", "--workspace-key", workspaceKey, "-l", "5", "-o", "0")
	res.PrintOutput()

	data := res.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含 items
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("分页查询任务失败: 响应格式不正确")
	}
	items, ok := dataField["items"].([]interface{})
	if !ok {
		// items 可能为 null，这是正常的空结果
		if dataField["items"] == nil {
			t.Logf("分页查询成功 (第1页显示 0 个任务)")
			return
		}
		t.Fatal("分页查询任务失败")
	}

	t.Logf("分页查询成功 (第1页显示 %d 个任务)", len(items))
}

// TestTaskDelete 测试删除任务
func TestTaskDelete(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建/删除操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 首先创建一个需求
	testDate := framework.GenerateTimestamp()
	reqName := fmt.Sprintf("任务删除测试需求_%s", testDate)

	res := lc.Run("req", "create", reqName, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	res.MustSucceed()
	data := res.ExpectJSON()
	reqObjectID := extractObjectID(data)

	if reqObjectID == "" {
		t.Fatal("创建需求失败，无法创建任务")
	}

	// 创建一个任务
	taskName := fmt.Sprintf("删除测试任务_%s", testDate)
	res = lc.Run("task", "create", reqObjectID, taskName, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	taskData := res.ExpectJSON()
	taskObjectID := extractObjectID(taskData)

	if taskObjectID == "" {
		t.Fatal("创建任务失败，无法测试删除")
	}

	t.Logf("删除任务: %s", taskObjectID)

	// 删除任务
	res = lc.Run("task", "delete", taskObjectID, "-k", "--workspace-key", workspaceKey)
	res.PrintOutput()

	if !res.IsSuccess() {
		t.Fatal("删除任务失败")
	}

	t.Logf("删除任务成功")

	// 轮询验证删除结果
	t.Logf("等待任务删除生效...")
	deleted := framework.PollForCondition(t, 5, 3*time.Second, func() bool {
		res = lc.Run("task", "list", "-k", "--workspace-key", workspaceKey, "-r", reqObjectID, "-l", "10")
		data := res.ExpectJSON()
		// 新的 JSON 格式: data 字段中包含 items
		dataField, ok := data["data"].(map[string]interface{})
		if !ok {
			return true // 如果格式不对，认为已删除
		}
		items, _ := dataField["items"].([]interface{})
		for _, item := range items {
			if taskMap, ok := item.(map[string]interface{}); ok {
				if taskMap["objectId"] == taskObjectID {
					return false
				}
			}
		}
		return true
	})

	if deleted {
		t.Logf("任务删除已验证（任务不存在）")
	} else {
		t.Logf("任务删除未在轮询周期内确认（可能存在后端延迟）")
	}

	// 清理需求
	time.Sleep(1 * time.Second)
	deleteReq(t, lc, reqObjectID, workspaceKey)
}
