// E2E 测试模板
// 使用说明：
// 1. 复制此文件到 e2e/integration/xxx_test.go
// 2. 将所有 "xxx" 替换为你的业务名
// 3. 修改测试用例以匹配实际业务
// 4. 删除本注释

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// ==================== 辅助函数 ====================

// deleteXxx 删除 xxx（用于测试清理）
func deleteXxx(t *testing.T, lc *framework.CLI, objectID, workspaceKey string) {
	t.Helper()
	res := lc.Run("xxx", "delete", objectID, "-k", "--workspace-key", workspaceKey)
	if res.IsSuccess() {
		t.Logf("删除成功: %s", objectID)
	} else {
		t.Logf("删除失败（可忽略）: %s", objectID)
	}
}

// findXxxByName 根据名称查找
func findXxxByName(t *testing.T, lc *framework.CLI, workspaceKey, name string) map[string]interface{} {
	t.Helper()
	res := lc.Run("xxx", "list", "-k", "--workspace-key", workspaceKey, "-l", "100")
	data := res.ExpectJSON()

	if dataField, ok := data["data"].(map[string]interface{}); ok {
		if items, ok := dataField["items"].([]interface{}); ok {
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if itemMap["name"] == name {
						return itemMap
					}
				}
			}
		}
	}
	return nil
}

// extractObjectID 从响应中提取 objectId
func extractObjectID(response map[string]interface{}) string {
	if dataField, ok := response["data"].(map[string]interface{}); ok {
		if objId, ok := dataField["objectId"].(string); ok {
			return objId
		}
	}
	return ""
}

// ==================== 测试用例 ====================

// TestXxxList 测试列表查询
func TestXxxList(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	t.Logf("查询 xxx 列表")

	res := lc.Run("xxx", "list", "-k", "--workspace-key", workspaceKey, "-l", "10")
	res.PrintOutput()

	data := res.ExpectJSON()

	// 验证响应结构
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatal("查询列表失败")
	}

	// 验证 data 字段
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("响应中缺少 data 字段")
	}

	// 验证 items 字段
	items, ok := dataField["items"].([]interface{})
	if !ok {
		t.Logf("警告: items 字段格式不正确")
		return
	}

	t.Logf("查询成功，共 %d 条记录", len(items))
}

// TestXxxCreateSimple 测试简单创建
func TestXxxCreateSimple(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 生成唯一名称
	testName := fmt.Sprintf("测试xxx-%s", framework.GenerateTimestamp())
	t.Logf("创建 xxx: %s", testName)

	res := lc.Run("xxx", "create", testName,
		"-k", "--workspace-key", workspaceKey,
		"--description", "这是一个测试")
	res.PrintOutput()

	data := res.ExpectJSON()

	// 验证创建成功
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatal("创建失败")
	}

	// 提取 objectId
	objectID := extractObjectID(data)
	if objectID == "" {
		t.Fatal("无法获取创建的 objectId")
	}
	t.Logf("创建成功，objectId: %s", objectID)

	// 清理
	if objectID != "" {
		time.Sleep(1 * time.Second)
		deleteXxx(t, lc, objectID, workspaceKey)
	}
}

// TestXxxView 测试查看详情
func TestXxxView(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 先创建一个
	testName := fmt.Sprintf("查看测试-%s", framework.GenerateTimestamp())
	createRes := lc.Run("xxx", "create", testName, "-k", "--workspace-key", workspaceKey)
	createData := createRes.ExpectJSON()
	objectID := extractObjectID(createData)

	if objectID == "" {
		t.Fatal("创建失败，无法测试查看")
	}

	t.Logf("查看详情: %s", objectID)

	// 查看详情
	res := lc.Run("xxx", "view", objectID, "-k", "--workspace-key", workspaceKey)
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatal("查看详情失败")
	}

	// 验证返回的 name
	if dataField, ok := data["data"].(map[string]interface{}); ok {
		if name, ok := dataField["name"].(string); ok {
			if name != testName {
				t.Errorf("名称不匹配: 期望 %s, 实际 %s", testName, name)
			}
		}
	}

	// 清理
	time.Sleep(1 * time.Second)
	deleteXxx(t, lc, objectID, workspaceKey)
}

// TestXxxDelete 测试删除
func TestXxxDelete(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 先创建一个
	testName := fmt.Sprintf("删除测试-%s", framework.GenerateTimestamp())
	createRes := lc.Run("xxx", "create", testName, "-k", "--workspace-key", workspaceKey)
	createData := createRes.ExpectJSON()
	objectID := extractObjectID(createData)

	if objectID == "" {
		t.Fatal("创建失败，无法测试删除")
	}

	t.Logf("删除: %s", objectID)

	// 删除
	res := lc.Run("xxx", "delete", objectID, "-k", "--workspace-key", workspaceKey)
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatal("删除失败")
	}

	t.Logf("删除成功")
}

// TestXxxCreateAndView 测试创建并查看完整流程
func TestXxxCreateAndView(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	var objectID string
	defer func() {
		// 清理
		if objectID != "" {
			time.Sleep(1 * time.Second)
			deleteXxx(t, lc, objectID, workspaceKey)
		}
	}()

	// 步骤1: 创建
	testName := fmt.Sprintf("流程测试-%s", framework.GenerateTimestamp())
	t.Logf("步骤1: 创建 %s", testName)

	createRes := lc.Run("xxx", "create", testName,
		"-k", "--workspace-key", workspaceKey,
		"--description", "测试描述")
	createData := createRes.ExpectJSON()

	if success, ok := createData["success"].(bool); !ok || !success {
		t.Fatal("创建失败")
	}

	objectID = extractObjectID(createData)
	if objectID == "" {
		t.Fatal("无法获取 objectId")
	}

	t.Logf("创建成功，objectId: %s", objectID)

	// 步骤2: 查看
	t.Logf("步骤2: 查看详情")
	time.Sleep(1 * time.Second)

	viewRes := lc.Run("xxx", "view", objectID, "-k", "--workspace-key", workspaceKey)
	viewData := viewRes.ExpectJSON()

	if success, ok := viewData["success"].(bool); !ok || !success {
		t.Fatal("查看失败")
	}

	// 验证
	if dataField, ok := viewData["data"].(map[string]interface{}); ok {
		if name, ok := dataField["name"].(string); ok && name == testName {
			t.Logf("验证成功")
		} else {
			t.Errorf("名称不匹配")
		}
	}

	t.Logf("流程测试通过")
}
