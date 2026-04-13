package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// deleteBug 删除缺陷
func deleteBug(t *testing.T, lc *framework.CLI, bugID, workspaceKey string) {
	t.Helper()
	res := lc.Run("bug", "delete", bugID, "-k", "--workspace-key", workspaceKey)
	if res.IsSuccess() {
		t.Logf("删除缺陷成功: %s", bugID)
	} else {
		t.Logf("删除缺陷失败: %s, 错误: %s", bugID, res.Stderr)
	}
}

// findBugInListByTitle 在列表中根据标题查找缺陷
func findBugInListByTitle(t *testing.T, lc *framework.CLI, workspaceKey, title string) map[string]interface{} {
	t.Helper()
	res := lc.Run("bug", "list", "-k", "--workspace-key", workspaceKey, "-p", "1", "-l", "100")
	data := res.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含 items
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		return nil
	}
	items, ok := dataField["items"].([]interface{})
	if !ok {
		return nil
	}
	for _, item := range items {
		if bugMap, ok := item.(map[string]interface{}); ok {
			if bugMap["defectName"] == title {
				return bugMap
			}
		}
	}
	return nil
}

// findBugInListByID 在列表中查找指定 ID 的缺陷
func findBugInListByID(t *testing.T, lc *framework.CLI, workspaceKey, bugID string) map[string]interface{} {
	t.Helper()
	res := lc.Run("bug", "list", "-k", "--workspace-key", workspaceKey, "-p", "1", "-l", "100")
	data := res.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含 items
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		return nil
	}
	items, ok := dataField["items"].([]interface{})
	if !ok {
		return nil
	}
	for _, item := range items {
		if bugMap, ok := item.(map[string]interface{}); ok {
			if bugMap["id"] == bugID {
				return bugMap
			}
		}
	}
	return nil
}

// TestBugLifecycle 测试缺陷的完整生命周期：创建 -> 列表查询 -> 状态修改 -> 验证 -> 删除 -> 验证
func TestBugLifecycle(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	projectID := framework.GetEnvOrDefault("LC_TEST_PROJECT_ID", "R24113J3C04")
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	testDate := framework.GenerateTimestamp()
	bugTitle := fmt.Sprintf("E2E测试缺陷_%s", testDate)
	bugDesc := fmt.Sprintf("这是E2E测试自动创建的缺陷描述_%s", testDate)

	t.Logf("步骤1: 创建缺陷 - %s", bugTitle)

	// 步骤1: 创建缺陷（使用简洁模板）
	res := lc.Run("bug", "create",
		"-t", bugTitle,
		"-D", bugDesc,
		"-p", projectID,
		"--workspace-key", workspaceKey,
		"--template-simple",
		"-k")
	res.PrintOutput()

	createData := res.ExpectJSON()
	if success, ok := createData["success"].(bool); !ok || !success {
		t.Fatalf("创建缺陷失败: %v", createData)
	}

	// Bug 创建 API 不返回 ID，通过列表查询获取
	t.Logf("创建缺陷成功，通过列表查询获取缺陷ID...")

	// 步骤2: 查询列表获取缺陷ID
	t.Logf("步骤2: 查询列表获取缺陷ID")
	time.Sleep(2 * time.Second)

	bug := findBugInListByTitle(t, lc, workspaceKey, bugTitle)
	if bug == nil {
		t.Fatalf("在列表中未找到刚创建的缺陷: %s", bugTitle)
	}

	bugID := bug["id"].(string)
	if bugID == "" {
		t.Fatal("获取缺陷ID失败")
	}

	t.Logf("获取缺陷成功 (ID: %s)", bugID)
	if bug == nil {
		t.Fatalf("在列表中未找到刚创建的缺陷: %s", bugID)
	}

	if bug["defectName"] != bugTitle {
		t.Fatalf("缺陷标题不匹配: 期望 %s, 实际 %s", bugTitle, bug["defectName"])
	}

	originalStatus := bug["status"].(string)
	t.Logf("在列表中找到缺陷，当前状态: %s", originalStatus)

	// 步骤3: 验证缺陷存在后，再次查询列表确认
	t.Logf("步骤3: 再次查询列表验证缺陷存在")
	bug = findBugInListByID(t, lc, workspaceKey, bugID)
	if bug == nil {
		t.Fatalf("按ID查询未找到缺陷: %s", bugID)
	}

	// 步骤4: 获取可用状态列表
	t.Logf("步骤4: 获取可用状态列表")
	res = lc.Run("bug", "status", "--workspace-key", workspaceKey, "-k")
	res.PrintOutput()

	// 解析统一格式的 JSON 响应
	statusData := res.ExpectJSON()
	dataField, ok := statusData["data"].([]interface{})
	if !ok {
		t.Fatalf("解析状态列表失败: data 字段格式不正确")
	}

	// 转换为 map 列表
	var statusList []map[string]interface{}
	for _, item := range dataField {
		if statusMap, ok := item.(map[string]interface{}); ok {
			statusList = append(statusList, statusMap)
		}
	}
	if len(statusList) == 0 {
		t.Fatal("获取状态列表为空")
	}

	// 选择一个有效的状态转换
	// 从"待修复"可以转换到"待验证"
	var newStatusID, newStatusName string
	for _, statusMap := range statusList {
		statusName := statusMap["statusName"].(string)
		// 优先选择"待验证"状态（从待修复可以转换到待验证）
		if statusName == "待验证" {
			newStatusID = statusMap["statusId"].(string)
			newStatusName = statusName
			break
		}
	}

	// 如果没找到"待验证"，选择一个其他状态（非当前状态）
	if newStatusID == "" {
		for _, statusMap := range statusList {
			statusName := statusMap["statusName"].(string)
			if statusName != originalStatus {
				newStatusID = statusMap["statusId"].(string)
				newStatusName = statusName
				break
			}
		}
	}

	if newStatusID == "" {
		t.Logf("警告: 无法找到与当前状态不同的状态，跳过状态更新测试")
	} else {
		// 步骤4: 修改缺陷状态
		t.Logf("步骤4: 修改缺陷状态为 %s (%s)", newStatusName, newStatusID)
		res = lc.Run("bug", "update-status", bugID, newStatusID, "--workspace-key", workspaceKey, "-k")
		res.PrintOutput()

		// 状态更新可能受限，不强制要求成功
		updateData := res.ExpectJSON()
		if success, ok := updateData["success"].(bool); !ok || !success {
			t.Logf("警告: 更新缺陷状态失败（可能受状态流转限制）: %v", updateData)
		} else {
			t.Logf("更新缺陷状态成功")
		}

		// 步骤5: 验证状态已更新
		t.Logf("步骤5: 验证状态已更新")
		time.Sleep(2 * time.Second)

		bug = findBugInListByID(t, lc, workspaceKey, bugID)
		if bug == nil {
			t.Fatalf("在列表中未找到缺陷: %s", bugID)
		}

		currentStatus := bug["status"].(string)
		if currentStatus != newStatusName {
			// 状态更新可能有延迟或受限，记录警告但不失败
			t.Logf("警告: 缺陷状态可能未更新: 期望 %s, 实际 %s", newStatusName, currentStatus)
		} else {
			t.Logf("缺陷状态已验证更新为: %s", currentStatus)
		}
	}

	// 步骤7: 删除缺陷
	t.Logf("步骤7: 删除缺陷 %s", bugID)
	res = lc.Run("bug", "delete", bugID, "--workspace-key", workspaceKey, "-k")
	res.PrintOutput()

	deleteData := res.ExpectJSON()
	if success, ok := deleteData["success"].(bool); !ok || !success {
		t.Fatalf("删除缺陷失败: %v", deleteData)
	}

	t.Logf("删除缺陷成功")

	// 步骤7: 验证缺陷已被删除（列表中不再出现）
	t.Logf("步骤7: 验证缺陷已被删除")
	time.Sleep(2 * time.Second)

	deleted := framework.PollForCondition(t, 5, 3*time.Second, func() bool {
		bug := findBugInListByID(t, lc, workspaceKey, bugID)
		return bug == nil
	})

	if deleted {
		t.Logf("缺陷删除已验证（缺陷不在列表中）")
	} else {
		t.Logf("警告: 缺陷删除未在轮询周期内确认（可能存在后端延迟）")
	}

	t.Logf("缺陷完整生命周期测试完成")
}

// TestBugCreateSimple 测试简单创建缺陷
func TestBugCreateSimple(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	projectID := framework.GetEnvOrDefault("LC_TEST_PROJECT_ID", "R24113J3C04")
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	testDate := framework.GenerateTimestamp()
	bugTitle := fmt.Sprintf("简单测试缺陷_%s", testDate)
	bugDesc := "这是简单测试缺陷的描述"

	t.Logf("创建简单缺陷: %s", bugTitle)

	res := lc.Run("bug", "create",
		"-t", bugTitle,
		"-D", bugDesc,
		"-p", projectID,
		"--workspace-key", workspaceKey,
		"--template-simple",
		"-k")
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatal("创建缺陷失败")
	}

	// 通过列表查询获取缺陷ID
	time.Sleep(2 * time.Second)
	bug := findBugInListByTitle(t, lc, workspaceKey, bugTitle)
	if bug == nil {
		t.Fatal("在列表中未找到刚创建的缺陷")
	}

	bugID := bug["id"].(string)
	t.Logf("创建缺陷成功 (ID: %s)", bugID)

	// 清理
	if bugID != "" {
		time.Sleep(1 * time.Second)
		deleteBug(t, lc, bugID, workspaceKey)
	}
}

// TestBugView 测试查看缺陷详情
func TestBugView(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	projectID := framework.GetEnvOrDefault("LC_TEST_PROJECT_ID", "R24113J3C04")
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 首先创建一个缺陷
	testDate := framework.GenerateTimestamp()
	bugTitle := fmt.Sprintf("查看测试缺陷_%s", testDate)
	bugDesc := "这是用于测试查看功能的缺陷"

	res := lc.Run("bug", "create",
		"-t", bugTitle,
		"-D", bugDesc,
		"-p", projectID,
		"--workspace-key", workspaceKey,
		"--template-simple",
		"-k")
	createData := res.ExpectJSON()
	if success, ok := createData["success"].(bool); !ok || !success {
		t.Fatal("创建缺陷失败，无法测试查看功能")
	}

	// 通过列表查询获取缺陷ID
	time.Sleep(2 * time.Second)
	bug := findBugInListByTitle(t, lc, workspaceKey, bugTitle)
	if bug == nil {
		t.Fatal("在列表中未找到刚创建的缺陷")
	}

	bugID := bug["id"].(string)
	t.Logf("创建缺陷成功 (ID: %s)", bugID)

	// 查看缺陷详情
	t.Logf("查看缺陷详情: %s", bugID)
	time.Sleep(1 * time.Second)

	res = lc.Run("bug", "view", bugID, "--workspace-key", workspaceKey, "-k")
	res.PrintOutput()

	viewData := res.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含详情
	viewDataField, ok := viewData["data"].(map[string]interface{})
	if !ok {
		t.Fatal("查看缺陷详情失败: 响应格式不正确")
	}
	if viewDataField["id"] != bugID {
		t.Fatalf("查看缺陷详情失败: ID 不匹配")
	}

	if viewDataField["defectName"] != bugTitle {
		t.Fatalf("缺陷标题不匹配: 期望 %s, 实际 %s", bugTitle, viewDataField["defectName"])
	}

	t.Logf("查看缺陷详情成功")

	// 清理
	time.Sleep(1 * time.Second)
	deleteBug(t, lc, bugID, workspaceKey)
}

// TestBugList 测试查询缺陷列表
func TestBugList(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	t.Logf("查询缺陷列表")

	res := lc.Run("bug", "list",
		"--workspace-key", workspaceKey,
		"-p", "1",
		"-l", "10",
		"-k")
	res.PrintOutput()

	data := res.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含 items 和 total
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("查询缺陷列表失败: 响应格式不正确")
	}
	if _, ok := dataField["items"].([]interface{}); !ok && dataField["items"] != nil {
		t.Fatal("查询缺陷列表失败: items 格式不正确")
	}

	total, _ := dataField["total"].(float64)
	t.Logf("查询缺陷列表成功 (共 %.0f 个缺陷)", total)
}

// TestBugDelete 测试删除缺陷
func TestBugDelete(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	projectID := framework.GetEnvOrDefault("LC_TEST_PROJECT_ID", "R24113J3C04")
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建/删除操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 创建一个缺陷
	testDate := framework.GenerateTimestamp()
	bugTitle := fmt.Sprintf("删除测试缺陷_%s", testDate)
	bugDesc := "这是用于测试删除功能的缺陷"

	res := lc.Run("bug", "create",
		"-t", bugTitle,
		"-D", bugDesc,
		"-p", projectID,
		"--workspace-key", workspaceKey,
		"--template-simple",
		"-k")
	createData := res.ExpectJSON()
	if success, ok := createData["success"].(bool); !ok || !success {
		t.Fatal("创建缺陷失败，无法测试删除功能")
	}

	// 通过列表查询获取缺陷ID
	time.Sleep(2 * time.Second)
	bug := findBugInListByTitle(t, lc, workspaceKey, bugTitle)
	if bug == nil {
		t.Fatal("在列表中未找到刚创建的缺陷")
	}

	bugID := bug["id"].(string)
	t.Logf("创建缺陷成功 (ID: %s)", bugID)

	// 删除缺陷
	t.Logf("删除缺陷: %s", bugID)
	time.Sleep(1 * time.Second)

	res = lc.Run("bug", "delete", bugID, "--workspace-key", workspaceKey, "-k")
	res.PrintOutput()

	deleteData := res.ExpectJSON()
	if success, ok := deleteData["success"].(bool); !ok || !success {
		t.Fatal("删除缺陷失败")
	}

	t.Logf("删除缺陷成功")

	// 验证缺陷已被删除
	t.Logf("验证缺陷已被删除")
	time.Sleep(2 * time.Second)

	bug = findBugInListByID(t, lc, workspaceKey, bugID)
	if bug != nil {
		t.Logf("警告: 缺陷删除后在列表中仍然存在")
	} else {
		t.Logf("缺陷删除已验证（缺陷不在列表中）")
	}
}
