package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

// TestLibraryLifecycleFull 测试完整的文档库生命周期（增删改查闭环）
// 包括: 创建文档库 -> 创建文件夹(查) -> 上传文件(查) -> 删除文件(查) -> 删除文件夹(查) -> 删除文档库
func TestLibraryLifecycleFull(t *testing.T) {
	framework.SkipIfShort(t)

	// E2E 测试强制使用小白测研发空间 (XXJSxiaobaice)
	// 详细限制参见 e2e/main_test.go TestMain 函数
	workspaceKey := framework.GetEnvOrDefault("LC_WORKSPACE_KEY", "XXJSxiaobaice")
	lc := framework.NewCLI(t)

	// 关闭只读模式以允许创建/删除操作
	lc.Run("readonly", "off", "--duration", "30m")

	// 生成唯一文档库名称
	libName := fmt.Sprintf("e2e-lib-%d", time.Now().UnixNano())
	var libID int64
	var externalLibID int64
	var folderID int64
	var docID int64
	var skipRemaining bool // 标记是否跳过后续测试

	t.Logf("生成文档库名称: %s", libName)

	// 前置清理：删除之前可能残留的测试文档库
	t.Logf("前置清理: 检查并删除残留的测试文档库...")
	res := lc.Run("lib", "list", "--workspace-key", workspaceKey, "-k")
	deletedCount := 0
	if res.IsSuccess() {
		result := res.ExpectJSON()
		if data, ok := result["data"].(map[string]interface{}); ok {
			if libraries, ok := data["libraries"].([]interface{}); ok {
				for _, lib := range libraries {
					if libMap, ok := lib.(map[string]interface{}); ok {
						if name, ok := libMap["libName"].(string); ok {
							// 删除所有 e2e-lib- 开头的残留文档库
							if strings.HasPrefix(name, "e2e-lib-") {
								if extID, ok := libMap["externalLibId"].(float64); ok {
									t.Logf("发现残留文档库 '%s' (ID: %.0f)，正在删除...", name, extID)
									delRes := lc.Run("lib", "delete", fmt.Sprintf("%.0f", extID), "-k")
									if delRes.IsSuccess() {
										t.Logf("残留文档库 '%s' 删除成功", name)
										deletedCount++
									} else {
										t.Logf("残留文档库 '%s' 删除失败: %s", name, delRes.GetStdout())
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if deletedCount > 0 {
		t.Logf("共删除 %d 个残留文档库，等待 5 秒让服务端同步...", deletedCount)
		time.Sleep(5 * time.Second)
	}

	// 清理函数：确保测试结束后删除文档库
	defer func() {
		if externalLibID != 0 {
			t.Logf("清理: 删除文档库 (External Lib ID: %d)...", externalLibID)
			res := lc.Run("lib", "delete", fmt.Sprintf("%d", externalLibID), "-k")
			if res.IsSuccess() {
				t.Logf("文档库删除成功")
			} else {
				t.Logf("文档库删除失败或已删除: %s", res.GetStdout())
			}
		}
	}()

	// 步骤 1: 创建文档库
	t.Run("CreateLibrary", func(t *testing.T) {
		t.Logf("步骤 1: 创建文档库 %s", libName)

		var res *framework.Result
		framework.Retry(t, 3, 2*time.Second, func() error {
			res = lc.Run("lib", "create", libName, "--workspace-key", workspaceKey, "-k")
			if res.ExitCode != 0 {
				return fmt.Errorf("exit code %d", res.ExitCode)
			}
			return nil
		})

		res.PrintOutput()

		// 检查是否是重复CODE错误，如果是则跳过所有后续测试
		if !res.IsSuccess() {
			stdout := res.GetStdout()
			if strings.Contains(stdout, "重复CODE") {
				skipRemaining = true
				t.Skip("工作空间下已存在文档库（libCode冲突），跳过文档库生命周期测试")
			}
			t.Fatal("创建文档库失败，无法继续后续测试")
		}

		result := res.ExpectJSON()
		data, ok := result["data"].(map[string]interface{})
		if !ok {
			t.Fatal("无法获取 data 对象")
		}
		if id, ok := data["libId"].(float64); ok {
			libID = int64(id)
			t.Logf("提取到文档库 ID: %d", libID)
		}

		// 等待并查询获取 externalLibId
		time.Sleep(2 * time.Second)
		res = lc.Run("lib", "list", "--workspace-key", workspaceKey, "-k")
		if res.IsSuccess() {
			result := res.ExpectJSON()
			// 解析 data.libraries 数组
			if data, ok := result["data"].(map[string]interface{}); ok {
				if libraries, ok := data["libraries"].([]interface{}); ok {
					for _, lib := range libraries {
						if libMap, ok := lib.(map[string]interface{}); ok {
							if name, ok := libMap["libName"].(string); ok && name == libName {
								if extID, ok := libMap["externalLibId"].(float64); ok {
									externalLibID = int64(extID)
									t.Logf("提取到 External Lib ID: %d", externalLibID)
								}
							}
						}
					}
				}
			}
		}

		if externalLibID == 0 {
			t.Fatal("无法获取 externalLibId")
		}

		time.Sleep(2 * time.Second)
	})

	// 步骤 2: 创建文件夹
	t.Run("CreateFolder", func(t *testing.T) {
		if skipRemaining {
			t.Skip("前置步骤失败，跳过此测试")
		}
		t.Logf("步骤 2: 在文档库 %d 中创建文件夹", externalLibID)

		folderName := fmt.Sprintf("e2e-folder-%d", time.Now().Unix())

		var res *framework.Result
		framework.Retry(t, 3, 2*time.Second, func() error {
			res = lc.Run("lib", "folder", "create", folderName, "--prt-id", fmt.Sprintf("%d", externalLibID), "-k")
			if res.ExitCode != 0 {
				return fmt.Errorf("exit code %d", res.ExitCode)
			}
			return nil
		})

		res.PrintOutput()

		result := res.ExpectJSON()
		// 获取嵌套的 data 对象
		data, ok := result["data"].(map[string]interface{})
		if !ok {
			t.Fatal("无法获取 data 对象")
		}
		if id, ok := data["folderId"].(float64); ok {
			folderID = int64(id)
			t.Logf("提取到文件夹 ID: %d", folderID)
		} else {
			t.Fatal("无法从输出中提取文件夹 ID")
		}

		time.Sleep(1 * time.Second)

		// 查询验证：检查文件夹是否已创建
		t.Logf("步骤 2.1: 查询验证文件夹已创建")
		res = lc.Run("lib", "folder", "list", "--prt-id", fmt.Sprintf("%d", externalLibID), "-k")
		if !res.IsSuccess() {
			t.Fatalf("查询文件夹列表失败: %s", res.GetStdout())
		}
		result = res.ExpectJSON()
		if data, ok := result["data"].(map[string]interface{}); ok {
			if items, ok := data["items"].([]interface{}); ok {
				found := false
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if id, ok := itemMap["id"].(string); ok && id == fmt.Sprintf("%d", folderID) {
							found = true
							if name, ok := itemMap["name"].(string); ok {
								t.Logf("查询验证成功: 文件夹 '%s' (ID: %d) 已存在于列表中", name, folderID)
							}
							break
						}
					}
				}
				if !found {
					t.Fatalf("查询验证失败: 文件夹 ID %d 未在列表中找到", folderID)
				}
			}
		}
	})

	// 步骤 3: 上传文件
	t.Run("UploadFile", func(t *testing.T) {
		if skipRemaining {
			t.Skip("前置步骤失败，跳过此测试")
		}
		t.Logf("步骤 3: 上传文件到文件夹 %d", folderID)

		// 创建测试文件
		tempDir := framework.CreateTempDir(t)
		testFile := filepath.Join(tempDir, "test_upload.txt")
		content := fmt.Sprintf("This is a test file for E2E testing.\nCreated at: %s\n", time.Now().Format(time.RFC3339))
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		var res *framework.Result
		framework.Retry(t, 3, 2*time.Second, func() error {
			res = lc.Run("lib", "upload", testFile, "--folder-id", fmt.Sprintf("%d", folderID), "--name", "e2e_test_file.txt", "-k")
			if res.ExitCode != 0 {
				return fmt.Errorf("exit code %d", res.ExitCode)
			}
			return nil
		})

		res.PrintOutput()

		result := res.ExpectJSON()
		// 获取嵌套的 data 对象
		data, ok := result["data"].(map[string]interface{})
		if !ok {
			t.Fatal("无法获取 data 对象")
		}
		if id, ok := data["docId"].(float64); ok {
			docID = int64(id)
			t.Logf("提取到文档 ID: %d", docID)
		} else {
			t.Fatal("无法从输出中提取文档 ID")
		}

		// 清理临时文件
		os.RemoveAll(tempDir)

		time.Sleep(1 * time.Second)

		// 查询验证：检查文件是否已上传
		t.Logf("步骤 3.1: 查询验证文件已上传")
		res = lc.Run("lib", "folder", "list", "--prt-id", fmt.Sprintf("%d", folderID), "-k")
		if !res.IsSuccess() {
			t.Fatalf("查询文件列表失败: %s", res.GetStdout())
		}
		result = res.ExpectJSON()
		if data, ok := result["data"].(map[string]interface{}); ok {
			if items, ok := data["items"].([]interface{}); ok {
				found := false
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if id, ok := itemMap["id"].(string); ok && id == fmt.Sprintf("%d", docID) {
							found = true
							if name, ok := itemMap["name"].(string); ok {
								t.Logf("查询验证成功: 文件 '%s' (ID: %d) 已存在于文件夹中", name, docID)
							}
							break
						}
					}
				}
				if !found {
					t.Fatalf("查询验证失败: 文件 ID %d 未在文件夹中找到", docID)
				}
			}
		}
	})

	// 步骤 4: 删除文件
	t.Run("DeleteFile", func(t *testing.T) {
		if skipRemaining {
			t.Skip("前置步骤失败，跳过此测试")
		}
		t.Logf("步骤 4: 删除文件 %d", docID)

		var res *framework.Result
		framework.Retry(t, 3, 2*time.Second, func() error {
			res = lc.Run("lib", "file", "delete", fmt.Sprintf("%d", docID), "--folder-id", fmt.Sprintf("%d", folderID), "-k")
			if res.ExitCode != 0 {
				return fmt.Errorf("exit code %d", res.ExitCode)
			}
			return nil
		})

		res.PrintOutput()

		if !res.IsSuccess() {
			t.Fatal("删除文件失败")
		}

		result := res.ExpectJSON()
		// 获取嵌套的 data 对象
		data, ok := result["data"].(map[string]interface{})
		if !ok {
			t.Fatal("无法获取 data 对象")
		}
		if deletedCount, ok := data["deletedCount"].(float64); ok {
			t.Logf("成功删除 %d 个文件", int(deletedCount))
		}

		time.Sleep(1 * time.Second)

		// 查询验证：检查文件是否已删除
		t.Logf("步骤 4.1: 查询验证文件已删除")
		res = lc.Run("lib", "folder", "list", "--prt-id", fmt.Sprintf("%d", folderID), "-k")
		if !res.IsSuccess() {
			t.Fatalf("查询文件列表失败: %s", res.GetStdout())
		}
		result = res.ExpectJSON()
		if data, ok := result["data"].(map[string]interface{}); ok {
			if items, ok := data["items"].([]interface{}); ok {
				found := false
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if id, ok := itemMap["id"].(string); ok && id == fmt.Sprintf("%d", docID) {
							found = true
							break
						}
					}
				}
				if found {
					t.Fatalf("查询验证失败: 文件 ID %d 仍然存在，删除未生效", docID)
				} else {
					t.Logf("查询验证成功: 文件 ID %d 已不在列表中，删除生效", docID)
				}
			} else {
				t.Logf("查询验证成功: 文件夹为空，文件已删除")
			}
		}
	})

	// 步骤 5: 删除文件夹
	t.Run("DeleteFolder", func(t *testing.T) {
		if skipRemaining {
			t.Skip("前置步骤失败，跳过此测试")
		}
		t.Logf("步骤 5: 删除文件夹 %d", folderID)

		var res *framework.Result
		framework.Retry(t, 3, 2*time.Second, func() error {
			res = lc.Run("lib", "file", "delete", fmt.Sprintf("%d", folderID), "--folder-id", fmt.Sprintf("%d", externalLibID), "-k")
			if res.ExitCode != 0 {
				return fmt.Errorf("exit code %d", res.ExitCode)
			}
			return nil
		})

		res.PrintOutput()

		if !res.IsSuccess() {
			t.Fatal("删除文件夹失败")
		}

		result := res.ExpectJSON()
		// 获取嵌套的 data 对象
		data, ok := result["data"].(map[string]interface{})
		if !ok {
			t.Fatal("无法获取 data 对象")
		}
		if deletedCount, ok := data["deletedCount"].(float64); ok {
			t.Logf("成功删除 %d 个文件夹", int(deletedCount))
		}

		time.Sleep(1 * time.Second)

		// 查询验证：检查文件夹是否已删除
		t.Logf("步骤 5.1: 查询验证文件夹已删除")
		res = lc.Run("lib", "folder", "list", "--prt-id", fmt.Sprintf("%d", externalLibID), "-k")
		if !res.IsSuccess() {
			t.Fatalf("查询文件夹列表失败: %s", res.GetStdout())
		}
		result = res.ExpectJSON()
		if data, ok := result["data"].(map[string]interface{}); ok {
			if items, ok := data["items"].([]interface{}); ok {
				found := false
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if id, ok := itemMap["id"].(string); ok && id == fmt.Sprintf("%d", folderID) {
							found = true
							break
						}
					}
				}
				if found {
					t.Fatalf("查询验证失败: 文件夹 ID %d 仍然存在，删除未生效", folderID)
				} else {
					t.Logf("查询验证成功: 文件夹 ID %d 已不在列表中，删除生效", folderID)
				}
			} else {
				t.Logf("查询验证成功: 文档库根目录为空，文件夹已删除")
			}
		}
	})

	// 步骤 6: 删除文档库
	t.Run("DeleteLibrary", func(t *testing.T) {
		if skipRemaining {
			t.Skip("前置步骤失败，跳过此测试")
		}
		t.Logf("步骤 6: 删除文档库 %d", externalLibID)

		var res *framework.Result
		framework.Retry(t, 3, 2*time.Second, func() error {
			res = lc.Run("lib", "delete", fmt.Sprintf("%d", externalLibID), "-k")
			if res.ExitCode != 0 {
				return fmt.Errorf("exit code %d", res.ExitCode)
			}
			return nil
		})

		res.PrintOutput()

		if !res.IsSuccess() {
			t.Fatal("删除文档库失败")
		}

		t.Logf("文档库删除成功")

		// 清理标记，避免 defer 重复删除
		externalLibID = 0
	})

	t.Logf("========================================")
	t.Logf("文档库端到端测试完成！")
	t.Logf("文档库名称: %s", libName)
	t.Logf("文档库 ID: %d", libID)
	t.Logf("External Lib ID: %d", externalLibID)
	t.Logf("创建的文件夹 ID: %d", folderID)
	t.Logf("上传的文档 ID: %d", docID)
	t.Logf("========================================")
}
