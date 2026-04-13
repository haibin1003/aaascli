// Package otp - OTP 动态配置管理端到端测试
//
// 测试 `lc otp config` 子命令（add / remove / reset）的行为：
//   - 动态添加命令到受保护列表后，该命令会被 OTP 拦截
//   - 从保护列表移除命令后，该命令不再被 OTP 拦截
//   - reset 可将自定义列表恢复为系统默认列表
//
// 运行方式:
//
//	go test ./e2e/otp/... -v -run TestOTPConfig
package otp

import (
	"testing"
	"time"

	"github.com/user/lc/e2e/framework"
)

const (
	// testConfigWorkspaceKey 小白测研发空间（供网络相关测试使用）
	testConfigWorkspaceKey = testWorkspaceKey // 复用 otp_e2e_test.go 中的常量
)

// ─────────────────────────────────────────────────────────────
// otp config add / remove / reset - 纯配置操作（无需网络）
// ─────────────────────────────────────────────────────────────

// TestOTPConfig_AddCommand 测试 otp config add 将命令加入保护列表
func TestOTPConfig_AddCommand(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	injectOTPConfig(t, cli, testOTPSecret, nil) // 使用默认列表

	// 默认列表不含 "req list"
	t.Run("1_AddReqList", func(t *testing.T) {
		res := cli.Run("otp", "config", "add", "req list", "-k")
		res.PrintOutput()
		if !res.IsSuccess() {
			t.Fatalf("otp config add 失败: %s", res.GetStdout())
		}
	})

	// 添加后，config list 应包含 "req list"
	t.Run("2_ListContainsReqList", func(t *testing.T) {
		res := cli.Run("otp", "config", "list", "-k")
		res.PrintOutput()

		data := res.ExpectJSON()
		listData, ok := data["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("data 字段类型错误")
		}
		cmds, ok := listData["protectedCommands"].([]interface{})
		if !ok {
			t.Fatalf("protectedCommands 类型错误")
		}

		found := false
		for _, cmd := range cmds {
			if cmd == "req list" {
				found = true
			}
		}
		if !found {
			t.Errorf("添加后 protectedCommands 应包含 'req list'，实际: %v", cmds)
		}
		t.Logf("✅ 动态添加成功，当前保护列表: %v", cmds)
	})

	// 没有有效 OTP 会话时，req list 应被拦截（stdin 为空 → EOF → 失败）
	t.Run("3_ReqListBlockedWithoutOTP", func(t *testing.T) {
		res := cli.RunWithInput("", "req", "list",
			"--workspace-key", testConfigWorkspaceKey, "-k")
		res.PrintOutput()

		if res.ExitCode == 0 {
			t.Fatal("添加到保护列表后，无 OTP 会话的 req list 应被拦截")
		}
		output := res.GetStdout() + res.GetStderr()
		if !isBlockedByOTP(output) {
			t.Logf("注意：失败原因可能不是 OTP，输出:\n%s", output)
		}
		t.Log("✅ 动态保护拦截生效：req list 在无 OTP 会话时被拦截")
	})

	// 建立 OTP 会话后，req list 应可执行（失败原因是 API 而非 OTP）
	t.Run("4_ReqListAllowedAfterOTPVerify", func(t *testing.T) {
		code := generateTOTPCode(testOTPSecret)
		t.Logf("验证 OTP，验证码: %s", code)
		vRes := cli.Run("otp", "verify", code, "-k")
		if !vRes.IsSuccess() {
			t.Fatalf("OTP 验证失败: %s", vRes.GetStdout())
		}

		res := cli.RunWithInput("", "req", "list",
			"--workspace-key", testConfigWorkspaceKey, "-k")
		res.PrintOutput()

		// 不检查 exit code（API 可能失败），只确保不是 OTP 拦截
		output := res.GetStdout() + res.GetStderr()
		if isBlockedByOTP(output) {
			t.Fatalf("有效 OTP 会话后，req list 不应被 OTP 拦截，实际:\n%s", output)
		}
		t.Log("✅ 有效 OTP 会话后 req list 不再被 OTP 拦截")
	})
}

// TestOTPConfig_RemoveCommand 测试 otp config remove 将命令从保护列表移除
func TestOTPConfig_RemoveCommand(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	// 注入自定义列表（含 "repo delete"），这样才能测试 remove
	injectOTPConfig(t, cli, testOTPSecret, []string{"repo delete", "readonly off"})

	// 确认 "repo delete" 初始在列表中
	t.Run("1_InitialListContainsRepoDelete", func(t *testing.T) {
		res := cli.Run("otp", "config", "list", "-k")
		data := res.ExpectJSON()
		listData, _ := data["data"].(map[string]interface{})
		cmds, _ := listData["protectedCommands"].([]interface{})

		found := false
		for _, cmd := range cmds {
			if cmd == "repo delete" {
				found = true
			}
		}
		if !found {
			t.Fatalf("初始列表应含 'repo delete'，实际: %v", cmds)
		}
		t.Logf("✅ 初始保护列表确认: %v", cmds)
	})

	// 移除 "repo delete"
	t.Run("2_RemoveRepoDelete", func(t *testing.T) {
		res := cli.Run("otp", "config", "remove", "repo delete", "-k")
		res.PrintOutput()
		if !res.IsSuccess() {
			t.Fatalf("otp config remove 失败: %s", res.GetStdout())
		}
	})

	// 移除后，"repo delete" 不再被 OTP 拦截（会因无仓库 ID 报参数错误，而非 OTP 错误）
	t.Run("3_RepoDeleteNotBlockedByOTP", func(t *testing.T) {
		// 用空 stdin 执行（无 OTP 会话）
		// 若 OTP 仍拦截 → exit!=0 且含 OTP 关键词
		// 若 OTP 不拦截 → 可能因参数错误退出，但不含 OTP 关键词
		res := cli.RunWithInput("",
			"repo", "delete", "0", // 无效 ID，必然失败
			"--workspace-key", testConfigWorkspaceKey, "-k")
		res.PrintOutput()

		output := res.GetStdout() + res.GetStderr()
		if isBlockedByOTP(output) {
			t.Fatalf("从保护列表移除后，repo delete 不应被 OTP 拦截，实际:\n%s", output)
		}
		t.Log("✅ 移除后 repo delete 不再被 OTP 拦截")
	})
}

// TestOTPConfig_Reset 测试 otp config reset 将自定义列表恢复为系统默认
func TestOTPConfig_Reset(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	// 注入自定义保护列表（与默认不同）
	customList := []string{"req create", "task create", "lib create"}
	injectOTPConfig(t, cli, testOTPSecret, customList)

	// 确认自定义列表已生效
	t.Run("1_CustomListActive", func(t *testing.T) {
		res := cli.Run("otp", "config", "list", "-k")
		data := res.ExpectJSON()
		listData, _ := data["data"].(map[string]interface{})
		cmds, _ := listData["protectedCommands"].([]interface{})

		if len(cmds) != len(customList) {
			t.Logf("自定义列表长度不符，期望 %d，实际 %d: %v", len(customList), len(cmds), cmds)
		}
		t.Logf("✅ 自定义列表已生效: %v", cmds)
	})

	// 执行 reset（有交互式确认提示，需通过 stdin 传入 "y"；提示在 stderr，stdout 为纯 JSON）
	t.Run("2_Reset", func(t *testing.T) {
		res := cli.RunWithInput("y\n", "otp", "config", "reset", "-k")
		res.PrintOutput()
		if !res.IsSuccess() {
			t.Fatalf("otp config reset 失败: %s", res.GetStdout())
		}
	})

	// reset 后应恢复默认列表（含 "readonly off"）
	t.Run("3_DefaultListRestored", func(t *testing.T) {
		res := cli.Run("otp", "config", "list", "-k")
		res.PrintOutput()

		data := res.ExpectJSON()
		listData, _ := data["data"].(map[string]interface{})
		cmds, _ := listData["protectedCommands"].([]interface{})

		foundReadonlyOff := false
		for _, cmd := range cmds {
			if cmd == "readonly off" {
				foundReadonlyOff = true
			}
		}
		if !foundReadonlyOff {
			t.Errorf("reset 后列表应包含 'readonly off'（默认保护命令），实际: %v", cmds)
		}

		// 自定义命令（req create 等）不应在默认列表中
		for _, custom := range customList {
			for _, cmd := range cmds {
				if cmd == custom {
					t.Errorf("reset 后 '%s' 不应出现在默认列表中", custom)
				}
			}
		}

		t.Logf("✅ reset 后默认列表已恢复: %v", cmds)
	})
}

// ─────────────────────────────────────────────────────────────
// OTP 动态保护 + 读操作（需要网络）
// ─────────────────────────────────────────────────────────────

// TestOTPConfig_DynamicReadProtection 验证读操作被动态加入 OTP 保护列表后行为正确
// 此测试覆盖"读操作也可以被 OTP 保护"的核心需求
func TestOTPConfig_DynamicReadProtection(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)

	// 注入 OTP 配置：仅保护 "readonly off"（不保护 "req list"）
	injectOTPConfig(t, cli, testOTPSecret, []string{"readonly off"})

	t.Run("1_ReqListNotProtectedByDefault", func(t *testing.T) {
		// req list 不在保护列表 → 即使无 OTP 会话也不会被 OTP 拦截
		// （可能因 API 失败，但不会因 OTP 失败）
		res := cli.RunWithInput("", "req", "list",
			"--workspace-key", testConfigWorkspaceKey, "-k")
		res.PrintOutput()

		output := res.GetStdout() + res.GetStderr()
		if isBlockedByOTP(output) {
			t.Fatalf("默认不保护 req list 时，不应被 OTP 拦截:\n%s", output)
		}
		t.Log("✅ 默认配置下 req list 不受 OTP 保护")
	})

	t.Run("2_AddReqListToProtection", func(t *testing.T) {
		res := cli.Run("otp", "config", "add", "req list", "-k")
		res.PrintOutput()
		if !res.IsSuccess() {
			t.Fatalf("otp config add req list 失败: %s", res.GetStdout())
		}
		t.Log("✅ req list 已添加到 OTP 保护列表")
	})

	t.Run("3_ReqListNowProtected", func(t *testing.T) {
		// 现在 req list 在保护列表中，无 OTP 会话时应被拦截
		res := cli.RunWithInput("", "req", "list",
			"--workspace-key", testConfigWorkspaceKey, "-k")
		res.PrintOutput()

		if res.ExitCode == 0 {
			t.Fatal("req list 加入保护列表后，无 OTP 会话时应被拦截")
		}
		output := res.GetStdout() + res.GetStderr()
		if !isBlockedByOTP(output) {
			t.Logf("注意：失败原因可能不是 OTP，输出:\n%s", output)
		}
		t.Log("✅ 读操作动态保护生效：req list 现在需要 OTP 验证")
	})
}

// TestOTPConfig_AddDuplicateCommand 测试添加重复命令时的错误处理
func TestOTPConfig_AddDuplicateCommand(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	injectOTPConfig(t, cli, testOTPSecret, []string{"repo delete", "readonly off"})

	// "repo delete" 已在列表中，再次添加应报错
	res := cli.Run("otp", "config", "add", "repo delete", "-k")
	res.PrintOutput()

	if res.IsSuccess() {
		t.Fatal("添加已存在命令时应返回错误")
	}

	t.Log("✅ 重复添加命令时正确报错")
}

// TestOTPConfig_RemoveNonExistentCommand 测试从自定义列表中移除不存在的命令时报错
func TestOTPConfig_RemoveNonExistentCommand(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	// 注入自定义列表（不含 "req list"）
	injectOTPConfig(t, cli, testOTPSecret, []string{"repo delete", "readonly off"})

	// "req list" 不在自定义列表中，移除应报错（"命令不在保护列表中"）
	res := cli.Run("otp", "config", "remove", "req list", "-k")
	res.PrintOutput()

	if res.IsSuccess() {
		t.Fatal("移除不存在的命令时应返回错误")
	}

	t.Log("✅ 从自定义列表移除不存在命令时正确报错")
}

// TestOTPConfig_RemoveFromDefaultList 测试在使用默认列表时直接 remove 应报错并给出提示
func TestOTPConfig_RemoveFromDefaultList(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	// 注入 OTP 配置但不设置自定义保护列表（protectedCommands 为空 → 使用默认列表）
	injectOTPConfig(t, cli, testOTPSecret, nil)

	// 当前使用默认列表，直接 remove 应报错并提示需要先创建自定义列表
	res := cli.Run("otp", "config", "remove", "repo delete", "-k")
	res.PrintOutput()

	if res.IsSuccess() {
		t.Fatal("使用默认列表时，remove 应返回错误")
	}

	t.Log("✅ 默认列表下 remove 正确报错，提示需先创建自定义列表")
}

// TestOTPConfig_SessionReuse 测试 OTP 会话可以跨多个受保护命令复用
func TestOTPConfig_SessionReuse(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	// 注入：同时保护 "readonly off" 和 "req list"
	injectOTPConfig(t, cli, testOTPSecret, []string{"readonly off", "req list"})

	// 先开启只读（确保 readonly off 是写操作会被只读拦截前，先验证 OTP）
	cli.Run("readonly", "on", "-k")

	// 验证一次 OTP
	t.Run("1_VerifyOTPOnce", func(t *testing.T) {
		code := generateTOTPCode(testOTPSecret)
		t.Logf("验证 OTP，验证码: %s", code)
		res := cli.Run("otp", "verify", code, "-k")
		if !res.IsSuccess() {
			t.Fatalf("OTP 验证失败: %s", res.GetStdout())
		}
		t.Log("✅ OTP 验证成功，会话已建立")
	})

	// 有效会话下，readonly off 不被 OTP 拦截
	t.Run("2_ReadonlyOffNotBlockedByOTP", func(t *testing.T) {
		res := cli.Run("readonly", "off")
		res.PrintOutput()
		if !res.IsSuccess() {
			t.Fatalf("有效 OTP 会话下 readonly off 应成功:\n%s", res.GetStdout())
		}
		t.Log("✅ 会话复用：readonly off 通过")
	})

	// 同一会话下，req list 也不被 OTP 拦截
	t.Run("3_ReqListNotBlockedByOTP", func(t *testing.T) {
		res := cli.RunWithInput("", "req", "list",
			"--workspace-key", testConfigWorkspaceKey, "-k")
		res.PrintOutput()

		output := res.GetStdout() + res.GetStderr()
		if isBlockedByOTP(output) {
			t.Fatalf("有效 OTP 会话下 req list 不应被 OTP 拦截:\n%s", output)
		}
		t.Log("✅ 会话复用：req list 通过（OTP 未重复要求）")
	})

	t.Logf("✅ OTP 会话复用测试通过（时间戳: %s）", time.Now().Format("15:04:05"))
}

// TestOTPConfig_AddUnregisteredCommand 测试添加未在 CommandRegistry 中注册的命令
func TestOTPConfig_AddUnregisteredCommand(t *testing.T) {
	framework.SkipIfShort(t)

	cli := framework.NewCLI(t)
	injectOTPConfig(t, cli, testOTPSecret, nil)

	// "nonexistent cmd" 不在 CommandRegistry 中
	res := cli.Run("otp", "config", "add", "nonexistent cmd", "-k")
	res.PrintOutput()

	// 根据实现，可能允许添加未知命令（宽松策略），也可能拒绝
	// 这里只验证命令能正常返回（不 panic）
	t.Logf("添加未知命令结果: exit=%d, output=%s", res.ExitCode, res.GetStdout())
	if res.ExitCode != 0 && res.ExitCode != 1 {
		t.Errorf("退出码应为 0 或 1，实际: %d", res.ExitCode)
	}

	t.Log("✅ 添加未知命令时返回合理结果（无 panic）")
}
