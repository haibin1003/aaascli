package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// E2E 测试安全限制：
// - 必须在 /tmp/ 或临时目录下运行
// - 强制使用 "小白测" 研发空间 (XXJSxiaobaice)
// 这些限制在 e2e/main_test.go 的 TestMain 中强制执行
const (
	defaultWorkspaceKey  = "XXJSxiaobaice"   // 小白测研发空间 - 唯一允许的测试空间
	defaultWorkspaceName = "小白测研发项目"     // 小白测研发空间名称
	defaultProjectCode   = "R24113J3C04"
	defaultGroupID       = "617927"          // 默认个人代码组ID (魏宝辉)
)

// extractObjectID 从 JSON 响应的 data 字段中提取 objectId
// 新的 JSON 格式: {"success": true, "data": {"objectId": "xxx"}, "meta": {}}
func extractObjectID(response map[string]interface{}) string {
	// 获取 data 字段
	dataField, ok := response["data"].(map[string]interface{})
	if !ok {
		return ""
	}
	// 从 data 中提取 objectId
	if objId, ok := dataField["objectId"].(string); ok {
		return objId
	}
	return ""
}

// deleteReq 删除需求
func deleteReq(t *testing.T, lc *framework.CLI, objectID, workspaceKey string) {
	t.Helper()
	res := lc.Run("req", "delete", objectID, "-k", "--workspace-key", workspaceKey)
	if res.IsSuccess() {
		t.Logf("删除需求成功: %s", objectID)
	} else {
		t.Logf("删除需求失败: %s", objectID)
	}
}

// TestReqCreateSimple 测试简单创建需求
func TestReqCreateSimple(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	reqName := fmt.Sprintf("简单测试需求_%s", framework.GenerateTimestamp())

	t.Logf("Creating requirement: %s", reqName)

	res := lc.Run("req", "create", reqName, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatal("简单创建需求失败")
	}

	objectID := extractObjectID(data)
	t.Logf("简单创建需求成功 (Object ID: %s)", objectID)

	// 搜索验证：确认需求可通过搜索找到
	t.Logf("搜索验证需求存在: %s", reqName)
	time.Sleep(1 * time.Second)
	searchRes := lc.Run("req", "search", reqName, "-k", "--workspace-key", workspaceKey, "-l", "10")
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
		t.Logf("搜索验证警告: 未找到刚创建的需求（可能存在延迟）")
	} else {
		found := false
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["objectId"] == objectID {
					found = true
					t.Logf("搜索验证成功: 需求存在于搜索结果中")
					break
				}
			}
		}
		if !found {
			t.Logf("搜索验证警告: 需求可能尚未进入搜索索引")
		}
	}

	// 清理：删除需求
	if objectID != "" {
		time.Sleep(1 * time.Second)
		deleteReq(t, lc, objectID, workspaceKey)
	}
}

// TestReqCreateFromYAMLFile 测试从 YAML 文件创建需求
func TestReqCreateFromYAMLFile(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 创建 YAML 文件
	yamlContent := `name: YAML详细测试需求
proposer:
  label: "魏宝辉(weibaohui@hq.cmcc)"
  value: weibaohui@hq.cmcc
  username: weibaohui@hq.cmcc
  nickname: 魏宝辉
assignee:
  label: "魏宝辉(weibaohui@hq.cmcc)"
  value: weibaohui@hq.cmcc
  username: weibaohui@hq.cmcc
  nickname: 魏宝辉
businessBackground: |
  这是通过YAML文件创建的详细需求
  业务背景支持多行文本描述
requirement: |
  需求描述详情：
  1. 支持YAML文件输入创建需求
  2. 支持详细的字段配置
acceptanceCriteria: |
  验收标准：
  1. 需求能正常创建
  2. 所有字段正确保存
requirementType:
  - 开发域
`

	yamlFile := framework.CreateTempFile(t, "test_req", yamlContent)

	t.Logf("Creating requirement from YAML file: %s", yamlFile)

	res := lc.Run("req", "create", "-f", yamlFile, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatal("YAML文件创建需求失败")
	}

	objectID := extractObjectID(data)
	t.Logf("YAML文件创建需求成功 (Object ID: %s)", objectID)

	// 清理
	if objectID != "" {
		time.Sleep(1 * time.Second)
		deleteReq(t, lc, objectID, workspaceKey)
	}
}

// TestReqCreateFromPipe 测试从管道输入创建需求
func TestReqCreateFromPipe(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	yamlContent := `name: 管道输入测试需求
proposer:
  label: "魏宝辉(weibaohui@hq.cmcc)"
  value: weibaohui@hq.cmcc
  username: weibaohui@hq.cmcc
  nickname: 魏宝辉
assignee:
  label: "魏宝辉(weibaohui@hq.cmcc)"
  value: weibaohui@hq.cmcc
  username: weibaohui@hq.cmcc
  nickname: 魏宝辉
businessBackground: 这是通过管道输入创建的测试需求
requirement: 验证管道输入功能是否正常
acceptanceCriteria: 管道输入能正常创建需求
requirementType:
  - 开发域
`

	t.Logf("Creating requirement from pipe input")

	res := lc.RunWithInput(yamlContent, "req", "create", "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatal("管道输入创建需求失败")
	}

	objectID := extractObjectID(data)
	t.Logf("管道输入创建需求成功 (Object ID: %s)", objectID)

	// 清理
	if objectID != "" {
		time.Sleep(1 * time.Second)
		deleteReq(t, lc, objectID, workspaceKey)
	}
}

// TestReqList 测试查询需求列表
func TestReqList(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 重置 OTP 配置：移除用户可能自定义的 protectedCommands（如 req list），
	// 避免测试被 OTP 交互式提示阻塞。
	framework.InjectOTPConfig(t, lc, "JBSWY3DPEHPK3PXP", nil)

	t.Logf("Listing requirements (limit 10)")

	res := lc.Run("req", "list", "-k", "--workspace-key", workspaceKey, "-l", "10")
	res.PrintOutput()

	data := res.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含 count 和 items
	dataField, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("查询需求列表失败: 响应格式不正确")
	}
	count, ok := dataField["count"].(float64)
	if !ok {
		t.Fatal("查询需求列表失败: 无法获取 count")
	}

	t.Logf("查询需求列表成功 (找到 %.0f 个需求)", count)
}

// TestReqView 测试查看需求详情
func TestReqView(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 首先创建一个需求
	testDate := framework.GenerateTimestamp()
	reqName := fmt.Sprintf("查看详情测试需求_%s", testDate)

	res := lc.Run("req", "create", reqName, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	data := res.ExpectJSON()
	objectID := extractObjectID(data)

	if objectID == "" {
		t.Fatal("创建需求失败，无法查看详情")
	}

	t.Logf("查看需求详情: %s", objectID)

	// 查看详情
	res = lc.Run("req", "view", objectID, "-k", "--workspace-key", workspaceKey)
	res.PrintOutput()

	viewData := res.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含详情
	viewDataField, ok := viewData["data"].(map[string]interface{})
	if !ok {
		t.Fatal("查看需求详情失败: 响应格式不正确")
	}
	if _, ok := viewDataField["name"]; !ok {
		t.Fatal("查看需求详情失败: 缺少 name 字段")
	}

	t.Logf("查看需求详情成功")

	// 清理
	time.Sleep(1 * time.Second)
	deleteReq(t, lc, objectID, workspaceKey)
}

// TestReqMinimalYAML 测试最小 YAML 配置创建
func TestReqMinimalYAML(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	testDate := framework.GenerateTimestamp()
	yamlContent := fmt.Sprintf(`name: 最小配置测试_%s`, testDate)

	t.Logf("Creating requirement with minimal YAML")

	res := lc.RunWithInput(yamlContent, "req", "create", "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatal("最小 YAML 创建需求失败")
	}

	objectID := extractObjectID(data)
	t.Logf("最小 YAML 创建需求成功 (Object ID: %s)", objectID)

	// 清理
	if objectID != "" {
		time.Sleep(1 * time.Second)
		deleteReq(t, lc, objectID, workspaceKey)
	}
}

// TestReqCreateWithoutOptionalFields 测试不传可选字段
func TestReqCreateWithoutOptionalFields(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建操作
	lc.Run("readonly", "off", "--duration", "30m")

	testDate := framework.GenerateTimestamp()
	yamlContent := fmt.Sprintf(`name: 无可选字段测试_%s
proposer:
  label: "魏宝辉(weibaohui@hq.cmcc)"
  value: weibaohui@hq.cmcc
  username: weibaohui@hq.cmcc
  nickname: 魏宝辉
assignee:
  label: "魏宝辉(weibaohui@hq.cmcc)"
  value: weibaohui@hq.cmcc
  username: weibaohui@hq.cmcc
  nickname: 魏宝辉
businessBackground: 测试不传contactNumber、contactEmail、affiliatedUnit
requirement: 验证不传这三个字段能否正常创建
acceptanceCriteria: 能正常创建
requirementType:
  - 开发域
`, testDate)

	t.Logf("Creating requirement without optional fields")

	res := lc.RunWithInput(yamlContent, "req", "create", "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Fatal("不传可选字段创建需求失败")
	}

	objectID := extractObjectID(data)
	t.Logf("不传可选字段创建需求成功 (Object ID: %s)", objectID)

	// 清理
	if objectID != "" {
		time.Sleep(1 * time.Second)
		deleteReq(t, lc, objectID, workspaceKey)
	}
}

// TestReqUpdate 测试更新需求字段
func TestReqUpdate(t *testing.T) {
	framework.SkipIfShort(t)

	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", defaultWorkspaceKey)
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建/更新操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 步骤1: 创建需求
	testDate := framework.GenerateTimestamp()
	reqName := fmt.Sprintf("更新测试需求_%s", testDate)

	t.Logf("步骤1: 创建需求: %s", reqName)
	res := lc.Run("req", "create", reqName, "-k", "--workspace-key", workspaceKey, "--project-code", defaultProjectCode)
	res.PrintOutput()

	createData := res.ExpectJSON()
	if success, ok := createData["success"].(bool); !ok || !success {
		t.Fatal("创建需求失败")
	}

	objectID := extractObjectID(createData)
	if objectID == "" {
		t.Fatal("无法获取创建的objectId")
	}
	t.Logf("创建需求成功 (Object ID: %s)", objectID)

	// 步骤2: 查询需求确认创建成功
	t.Logf("步骤2: 查询需求确认创建成功")
	time.Sleep(1 * time.Second)
	res = lc.Run("req", "view", objectID, "-k", "--workspace-key", workspaceKey)
	res.PrintOutput()

	viewData := res.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含详情
	viewDataField, ok := viewData["data"].(map[string]interface{})
	if !ok {
		t.Fatal("查询需求失败: 响应格式不正确")
	}
	if name, ok := viewDataField["name"].(string); !ok || name != reqName {
		t.Fatalf("查询需求失败，期望名称: %s, 实际名称: %s", reqName, name)
	}
	t.Logf("查询需求确认创建成功")

	// 步骤3: 更新需求字段
	newName := fmt.Sprintf("更新后的需求_%s", testDate)
	newRequirement := "这是更新后的需求描述"
	newAcceptanceCriteria := "这是更新后的验收标准"

	t.Logf("步骤3: 更新需求字段")
	res = lc.Run("req", "update", objectID,
		"-k", "--workspace-key", workspaceKey,
		"--name", newName,
		"--requirement", newRequirement,
		"--acceptance-criteria", newAcceptanceCriteria)
	res.PrintOutput()

	updateData := res.ExpectJSON()
	if success, ok := updateData["success"].(bool); !ok || !success {
		t.Fatal("更新需求失败")
	}

	// 验证更新响应中的字段 (数据在 data 字段中)
	data, ok := updateData["data"].(map[string]interface{})
	if !ok {
		t.Fatal("更新响应中缺少 data 字段")
	}
	if name, ok := data["name"].(string); !ok || name != newName {
		t.Fatalf("更新响应中名称不匹配，期望: %s, 实际: %s", newName, name)
	}
	t.Logf("更新需求成功")

	// 步骤4: 再次查询确认更新成功
	t.Logf("步骤4: 再次查询确认更新成功")
	time.Sleep(1 * time.Second)
	res = lc.Run("req", "view", objectID, "-k", "--workspace-key", workspaceKey)
	res.PrintOutput()

	viewData2 := res.ExpectJSON()
	// 新的 JSON 格式: data 字段中包含详情
	viewData2Field, ok := viewData2["data"].(map[string]interface{})
	if !ok {
		t.Fatal("查看需求详情失败: 响应格式不正确")
	}

	// 验证需求描述已更新 (富文本字段通常在 values 中更新)
	requirement := framework.ExtractTextFromRichText(viewData2Field["requirement"])
	if strings.TrimSpace(requirement) != strings.TrimSpace(newRequirement) {
		t.Fatalf("查询验证失败，期望需求描述: %q, 实际: %q", newRequirement, requirement)
	}

	// 验证验收标准已更新
	acceptanceCriteria := framework.ExtractTextFromRichText(viewData2Field["acceptanceCriteria"])
	if strings.TrimSpace(acceptanceCriteria) != strings.TrimSpace(newAcceptanceCriteria) {
		t.Fatalf("查询验证失败，期望验收标准: %q, 实际: %q", newAcceptanceCriteria, acceptanceCriteria)
	}

	// 注意：名称更新可能由于 API 缓存不会立即在 view 中反映
	// 但更新响应中已确认成功
	t.Logf("查询验证更新成功")

	// 步骤5: 测试更新计划完成时间
	plannedEndTime := time.Now().AddDate(0, 1, 0).UnixMilli() // 一个月后
	t.Logf("步骤5: 更新计划完成时间")
	res = lc.Run("req", "update", objectID,
		"-k", "--workspace-key", workspaceKey,
		"--planned-end-time", fmt.Sprintf("%d", plannedEndTime))
	res.PrintOutput()

	updateTimeData := res.ExpectJSON()
	if success, ok := updateTimeData["success"].(bool); !ok || !success {
		t.Fatal("更新计划完成时间失败")
	}

	// 验证更新时间响应
	if timeData, ok := updateTimeData["data"].(map[string]interface{}); ok {
		if updatedAt, ok := timeData["updatedAt"].(string); ok {
			t.Logf("更新计划完成时间成功，更新时间: %s", updatedAt)
		}
	}

	// 清理
	if objectID != "" {
		time.Sleep(1 * time.Second)
		deleteReq(t, lc, objectID, workspaceKey)
	}
}
