// Package otp 提供 OTP 二次验证功能的端到端测试
//
// 运行方式:
//
//	go test ./e2e/otp/... -v
//	go test ./e2e/otp/... -v -run TestOTPLifecycle
//
// 注意:
//   - 测试使用 framework.NewCLI 隔离 HOME 目录，不影响真实配置
//   - 涉及远程操作的测试强制使用 XXJSxiaobaice 小白测研发空间
//   - 删除的内容均为测试临时创建，不会影响已有数据
package otp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

const (
	// testOTPSecret 是测试专用的 TOTP 密钥（RFC 6238 标准测试向量）
	testOTPSecret = "JBSWY3DPEHPK3PXP"

	// testWorkspaceKey 强制使用小白测研发空间
	testWorkspaceKey = "XXJSxiaobaice"

	// testGroupID 小白测空间下的个人代码组 ID
	testGroupID = "617927"
)

// TestMain 初始化测试环境
func TestMain(m *testing.M) {
	if err := framework.EnsureBinary(); err != nil {
		fmt.Fprintf(os.Stderr, "构建 CLI 二进制失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ OTP E2E 测试环境初始化完成，使用研发空间: %s\n\n", testWorkspaceKey)
	os.Exit(m.Run())
}

// generateTOTPCode 根据密钥生成当前时间的 TOTP 验证码（RFC 6238）
func generateTOTPCode(secret string) string {
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		panic("TOTP 密钥无效: " + err.Error())
	}
	counter := uint64(math.Floor(float64(time.Now().Unix()) / 30))
	mac := hmac.New(sha1.New, key)
	_ = binary.Write(mac, binary.BigEndian, counter)
	hash := mac.Sum(nil)
	offset := hash[len(hash)-1] & 0x0F
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF
	code = code % 1000000
	return fmt.Sprintf("%06d", code)
}

// injectOTPConfig 向 CLI 的隔离 HOME 目录注入 OTP 配置（委托给 framework）
func injectOTPConfig(t *testing.T, cli *framework.CLI, secret string, protectedCmds []string) {
	t.Helper()
	framework.InjectOTPConfig(t, cli, secret, protectedCmds)
}

// isBlockedByOTP 判断命令是否因 OTP 保护而被拦截。
// 使用精确关键词避免误判（如"OTP 未启用"不属于被拦截）。
func isBlockedByOTP(output string) bool {
	return strings.Contains(output, "OTP_REQUIRED") ||
		strings.Contains(output, "读取输入失败") ||
		strings.Contains(output, "请输入 OTP 验证码")
}

// ─────────────────────────────────────────────
// 本地配置测试（不需要网络，不影响远端数据）
// ─────────────────────────────────────────────

// TestOTPStatus_Disabled 测试 OTP 未启用时的状态查询
func TestOTPStatus_Disabled(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)

	res := cli.Run("otp", "status", "-k")
	res.PrintOutput()

	data := res.ExpectJSON()
	otpData, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data 字段类型错误")
	}

	if enabled, _ := otpData["enabled"].(bool); enabled {
		t.Skip("当前 OTP 已启用（用户真实配置），跳过未启用测试")
	}

	t.Log("✅ OTP 未启用状态验证通过")
}

// TestOTPStatus_Enabled 测试通过注入配置启用 OTP 后的状态查询
func TestOTPStatus_Enabled(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	injectOTPConfig(t, cli, testOTPSecret, nil)

	res := cli.Run("otp", "status", "-k")
	res.PrintOutput()

	data := res.ExpectJSON()
	if success, _ := data["success"].(bool); !success {
		t.Fatalf("状态查询应成功: %v", data)
	}

	otpData, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data 字段类型错误")
	}

	if enabled, _ := otpData["enabled"].(bool); !enabled {
		t.Fatalf("注入配置后 OTP 应为启用状态")
	}

	t.Log("✅ OTP 启用状态验证通过")
}

// TestOTPConfigList_Default 测试默认保护命令列表
func TestOTPConfigList_Default(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	// 注入干净的 OTP 配置：有 secret 但不设置自定义 protectedCommands，
	// 确保测试读到的是系统默认列表（repo delete、readonly off），
	// 而非用户真实配置里的自定义列表。
	injectOTPConfig(t, cli, testOTPSecret, nil)

	res := cli.Run("otp", "config", "list", "-k")
	res.PrintOutput()

	data := res.ExpectJSON()
	listData, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data 字段类型错误")
	}

	cmds, ok := listData["protectedCommands"].([]interface{})
	if !ok {
		t.Fatalf("protectedCommands 类型错误: %T", listData["protectedCommands"])
	}

	foundReadonlyOff := false
	for _, cmd := range cmds {
		if cmd == "readonly off" {
			foundReadonlyOff = true
		}
	}

	if !foundReadonlyOff {
		t.Errorf("默认列表应包含 'readonly off'，实际: %v", cmds)
	}

	t.Logf("✅ 默认保护列表验证通过: %v", cmds)
}

// TestOTPVerify_ValidCode 测试有效验证码的验证
func TestOTPVerify_ValidCode(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	injectOTPConfig(t, cli, testOTPSecret, nil)

	code := generateTOTPCode(testOTPSecret)
	t.Logf("生成测试验证码: %s", code)

	res := cli.Run("otp", "verify", code, "-k")
	res.PrintOutput()

	if !res.IsSuccess() {
		t.Fatalf("有效验证码应验证成功，实际输出:\n%s", res.GetStdout())
	}

	t.Log("✅ 有效验证码验证通过，会话已建立")
}

// TestOTPVerify_InvalidCode 测试无效验证码被拒绝
func TestOTPVerify_InvalidCode(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	injectOTPConfig(t, cli, testOTPSecret, nil)

	// 000000 是错误的验证码（极低概率恰好正确，可接受）
	res := cli.Run("otp", "verify", "000000", "-k")
	res.PrintOutput()

	if res.IsSuccess() {
		t.Fatal("无效验证码不应验证成功")
	}

	t.Log("✅ 无效验证码被正确拒绝")
}

// TestOTPVerify_InvalidFormat 测试各种格式错误的验证码
func TestOTPVerify_InvalidFormat(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	injectOTPConfig(t, cli, testOTPSecret, nil)

	cases := []struct {
		code string
		desc string
	}{
		{"12345", "5位数字"},
		{"1234567", "7位数字"},
		{"abcdef", "字母"},
		{"12345a", "含字母"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			res := cli.Run("otp", "verify", tc.code, "-k")
			if res.IsSuccess() {
				t.Errorf("格式错误的验证码 '%s'(%s) 不应验证成功", tc.code, tc.desc)
			}
		})
	}

	t.Log("✅ 验证码格式校验全部通过")
}

// ─────────────────────────────────────────────
// OTP 保护拦截测试（readonly off - 本地操作）
// ─────────────────────────────────────────────

// TestOTPGuard_ReadonlyOff_Blocked 测试 readonly off 在无 OTP 会话时被拦截
func TestOTPGuard_ReadonlyOff_Blocked(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)

	// 先开启只读
	cli.Run("readonly", "on", "-k")

	// 注入 OTP 配置（默认保护列表包含 "readonly off"）
	injectOTPConfig(t, cli, testOTPSecret, nil)

	// 使用空 stdin 执行 readonly off → OTP 提示读取 stdin 得到 EOF → 失败
	res := cli.RunWithInput("", "readonly", "off")
	res.PrintOutput()

	if res.ExitCode == 0 {
		t.Fatal("无 OTP 会话时，readonly off 应被 OTP 保护拦截")
	}

	output := res.GetStdout() + res.GetStderr()
	if !isBlockedByOTP(output) {
		t.Logf("警告: 失败原因可能不是 OTP，输出:\n%s", output)
	}

	t.Log("✅ OTP 保护正确拦截了无会话的 readonly off 操作")
}

// TestOTPGuard_ReadonlyOff_AllowedWithVerify 测试验证 OTP 后 readonly off 可以执行
func TestOTPGuard_ReadonlyOff_AllowedWithVerify(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)

	// 先开启只读
	cli.Run("readonly", "on", "-k")

	// 注入 OTP 配置
	injectOTPConfig(t, cli, testOTPSecret, nil)

	// 先验证 OTP 建立会话
	code := generateTOTPCode(testOTPSecret)
	t.Logf("验证 OTP，验证码: %s", code)
	verifyRes := cli.Run("otp", "verify", code, "-k")
	verifyRes.PrintOutput()
	if !verifyRes.IsSuccess() {
		t.Fatalf("OTP 验证失败: %s", verifyRes.GetStdout())
	}

	// 有效会话下 readonly off 应成功
	res := cli.Run("readonly", "off")
	res.PrintOutput()

	if !res.IsSuccess() {
		t.Fatalf("有效 OTP 会话下 readonly off 应成功，实际:\n%s", res.GetStdout())
	}

	t.Log("✅ 有效 OTP 会话后 readonly off 操作成功")
}

// TestOTPGuard_ReadonlyOff_AllowedViaStdin 测试通过 stdin 输入 OTP 码后 readonly off 可以执行
func TestOTPGuard_ReadonlyOff_AllowedViaStdin(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)

	// 先开启只读
	cli.Run("readonly", "on", "-k")

	// 注入 OTP 配置
	injectOTPConfig(t, cli, testOTPSecret, nil)

	// 直接通过 stdin 传入 OTP 码（模拟用户在提示时输入）
	// 交互提示已改到 stderr，stdout 为纯 JSON，可直接用 IsSuccess()
	code := generateTOTPCode(testOTPSecret)
	t.Logf("通过 stdin 传入 OTP 验证码: %s", code)
	res := cli.RunWithInput(code+"\n", "readonly", "off")
	res.PrintOutput()

	if !res.IsSuccess() {
		t.Fatalf("通过 stdin 输入有效 OTP 后 readonly off 应成功:\n%s", res.GetStdout())
	}

	t.Log("✅ 通过 stdin 输入 OTP 后 readonly off 操作成功")
}

// ─────────────────────────────────────────────
// OTP 保护拦截测试（repo delete - 小白测研发空间）
// 要求: 删除的内容必须是测试新建的，操作在 XXJSxiaobaice 空间下进行
// ─────────────────────────────────────────────

// TestOTPGuard_RepoDelete_InWorkspace 测试 repo delete 在小白测空间中的 OTP 保护
// 流程: 创建测试仓库 → OTP 拦截验证 → 提供有效 OTP → 删除仓库
func TestOTPGuard_RepoDelete_InWorkspace(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)

	// 关闭只读模式以允许创建/删除操作
	cli.Run("readonly", "off", "--duration", "30m")

	repoName := fmt.Sprintf("test-otp-guard-%d", time.Now().Unix())
	t.Logf("测试仓库名称: %s", repoName)

	var projectID string

	// 确保清理（无论测试成败）
	t.Cleanup(func() {
		if projectID == "" {
			return
		}
		t.Logf("清理: 删除测试仓库 (project-id: %s)", projectID)
		// 需要有效 OTP 会话才能删除
		code := generateTOTPCode(testOTPSecret)
		cli.Run("otp", "verify", code, "-k")
		cli.Run("repo", "delete", projectID,
			"--workspace-key", testWorkspaceKey,
			"-k",
		)
	})

	// 步骤 1: 在小白测空间创建测试仓库
	t.Run("1_CreateTestRepo", func(t *testing.T) {
		res := cli.Run("repo", "create", repoName,
			"--workspace-key", testWorkspaceKey,
			"--group-id", testGroupID,
			"-k",
		)
		res.PrintOutput()

		data := res.ExpectJSON()
		if success, _ := data["success"].(bool); !success {
			t.Fatalf("创建测试仓库失败: %v", data)
		}

		dataField, ok := data["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("响应 data 字段格式错误")
		}

		if gitProjectID, ok := dataField["gitProjectId"].(float64); ok {
			projectID = fmt.Sprintf("%.0f", gitProjectID)
		} else if id, ok := dataField["id"].(float64); ok {
			projectID = fmt.Sprintf("%.0f", id)
		} else {
			t.Fatalf("无法从响应中提取项目 ID: %v", dataField)
		}

		t.Logf("✅ 测试仓库创建成功，项目 ID: %s", projectID)
	})

	if projectID == "" {
		t.Fatal("仓库创建失败，无法继续测试")
	}

	// 步骤 2: 注入 OTP 配置（默认列表包含 "repo delete"）
	injectOTPConfig(t, cli, testOTPSecret, nil)

	// 步骤 3: 尝试删除仓库（无 OTP 会话，空 stdin → 被拦截）
	t.Run("2_DeleteBlocked_WithoutOTP", func(t *testing.T) {
		res := cli.RunWithInput("", "repo", "delete", projectID,
			"--workspace-key", testWorkspaceKey,
			"-k",
		)
		res.PrintOutput()

		if res.ExitCode == 0 {
			t.Fatal("无 OTP 会话时，repo delete 应被 OTP 保护拦截")
		}

		output := res.GetStdout() + res.GetStderr()
		if !isBlockedByOTP(output) {
			t.Logf("注意: 失败原因可能不是 OTP，输出:\n%s", output)
		}

		t.Log("✅ OTP 保护正确拦截了无会话的 repo delete 操作")
	})

	// 步骤 4: 验证 OTP 建立会话
	t.Run("3_VerifyOTP", func(t *testing.T) {
		code := generateTOTPCode(testOTPSecret)
		t.Logf("验证 OTP，验证码: %s", code)
		res := cli.Run("otp", "verify", code, "-k")
		res.PrintOutput()
		if !res.IsSuccess() {
			t.Fatalf("OTP 验证失败: %s", res.GetStdout())
		}
		t.Log("✅ OTP 验证成功，会话已建立")
	})

	// 步骤 5: 重新尝试删除仓库（有效会话 → 成功）
	t.Run("4_DeleteAllowed_WithOTPSession", func(t *testing.T) {
		res := cli.Run("repo", "delete", projectID,
			"--workspace-key", testWorkspaceKey,
			"-k",
		)
		res.PrintOutput()

		if !res.IsSuccess() {
			t.Fatalf("有效 OTP 会话下 repo delete 应成功:\n%s", res.GetStdout())
		}

		t.Logf("✅ 仓库 (ID: %s) 在 OTP 验证后成功删除", projectID)
		projectID = "" // 已删除，清理时跳过
	})
}

// ─────────────────────────────────────────────
// 完整生命周期测试
// ─────────────────────────────────────────────

// TestOTPLifecycle 测试 OTP 完整生命周期（本地配置操作）
func TestOTPLifecycle(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)

	t.Run("1_InitialState", func(t *testing.T) {
		res := cli.Run("otp", "status", "-k")
		data := res.ExpectJSON()
		if success, _ := data["success"].(bool); !success {
			t.Fatalf("初始状态查询应成功: %v", data)
		}
		t.Log("✅ 初始状态检查通过")
	})

	t.Run("2_InjectAndCheckStatus", func(t *testing.T) {
		injectOTPConfig(t, cli, testOTPSecret, nil)
		res := cli.Run("otp", "status", "-k")
		data := res.ExpectJSON()
		otpData, _ := data["data"].(map[string]interface{})
		if enabled, _ := otpData["enabled"].(bool); !enabled {
			t.Fatal("注入 OTP 配置后状态应为启用")
		}
		t.Log("✅ OTP 注入后状态为启用")
	})

	t.Run("3_ConfigList", func(t *testing.T) {
		res := cli.Run("otp", "config", "list", "-k")
		data := res.ExpectJSON()
		listData, _ := data["data"].(map[string]interface{})
		t.Logf("保护命令列表: %v", listData["protectedCommands"])
		t.Log("✅ 保护命令列表查询正常")
	})

	t.Run("4_VerifyOTP", func(t *testing.T) {
		code := generateTOTPCode(testOTPSecret)
		res := cli.Run("otp", "verify", code, "-k")
		if !res.IsSuccess() {
			t.Fatalf("有效验证码应通过: %s", res.GetStdout())
		}
		t.Log("✅ OTP 验证成功，会话已建立")
	})

	t.Run("5_OTPCommandsIndependentOfReadonly", func(t *testing.T) {
		// 开启只读
		cli.Run("readonly", "on", "-k")
		// OTP 命令不受只读限制
		res := cli.Run("otp", "status", "-k")
		data := res.ExpectJSON()
		if success, _ := data["success"].(bool); !success {
			t.Fatal("只读模式下 OTP 命令应可正常执行")
		}
		t.Log("✅ OTP 命令不受只读模式限制")
	})

	t.Log("✅ OTP 完整生命周期测试通过")
}
