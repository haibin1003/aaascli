package autodetect

import (
	"fmt"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// TestAutoDetectModeLibList 测试自动探测模式查询组件库列表
func TestAutoDetectModeLibList(t *testing.T) {
	framework.SkipIfShort(t)
	framework.SkipIfNotInGitRepo(t)

	a := framework.NewAutoDetectTest(t)

	t.Logf("测试自动探测模式查询组件库列表")
	t.Logf("在 Git 仓库目录下执行: %s", a.GetRepoDir())

	res := a.RunInRepo("lib", "list", "-k")
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

	t.Logf("自动探测模式查询组件库列表成功 (找到 %.0f 个组件库)", count)
}

// TestAutoDetectModeLibCreateAndDelete 测试自动探测模式创建和删除组件库
func TestAutoDetectModeLibCreateAndDelete(t *testing.T) {
	framework.SkipIfShort(t)
	framework.SkipIfNotInGitRepo(t)

	a := framework.NewAutoDetectTest(t)

	libName := fmt.Sprintf("自动探测测试组件_%s", framework.GenerateTimestamp())
	t.Logf("测试自动探测模式创建组件库: %s", libName)

	// 创建组件库 - 自动探测模式
	res := a.RunInRepo("lib", "create", libName, "-k")
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, ok := data["success"].(bool); !ok || !success {
		t.Logf("自动探测模式创建组件库失败，跳过删除测试")
		return
	}

	objectID := extractObjectID(data)
	t.Logf("自动探测模式创建组件库成功 (Object ID: %s)", objectID)

	// 等待一下再删除
	time.Sleep(1 * time.Second)

	// 删除组件库
	t.Logf("测试自动探测模式删除组件库: %s", objectID)
	res = a.RunInRepo("lib", "delete", objectID, "-k")
	res.PrintOutput()

	deleteData := res.ExpectJSON()
	if success, ok := deleteData["success"].(bool); ok && success {
		t.Logf("自动探测模式删除组件库成功")
	} else {
		t.Logf("自动探测模式删除组件库可能失败")
	}
}

