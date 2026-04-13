package cli

import (
	"testing"

	"github.com/user/lc/e2e/framework"
)

// TestHelp 测试帮助命令
func TestHelp(t *testing.T) {
	lc := framework.NewCLI(t)

	res := lc.Run("--help")
	res.ExpectExit(0).
		ExpectContains("Usage").
		ExpectContains("lc")
}

// TestHelpSubcommand 测试子命令帮助
func TestHelpSubcommand(t *testing.T) {
	tests := []struct {
		name []string
	}{
		{name: []string{"repo", "--help"}},
		{name: []string{"req", "--help"}},
		{name: []string{"pr", "--help"}},
		{name: []string{"task", "--help"}},
		{name: []string{"login", "--help"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name[0], func(t *testing.T) {
			t.Parallel()
			lc := framework.NewCLI(t)
			res := lc.Run(tt.name...)
			res.ExpectSuccess().
				ExpectContains("Usage").
				ExpectContains(tt.name[0])
		})
	}
}

// TestRepoCommandsHelp 测试仓库命令帮助
func TestRepoCommandsHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"list", []string{"repo", "list", "--help"}},
		{"create", []string{"repo", "create", "--help"}},
		{"delete", []string{"repo", "delete", "--help"}},
		{"disable-work-item-link", []string{"repo", "disable-work-item-link", "--help"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lc := framework.NewCLI(t)
			res := lc.Run(tt.args...)
			res.ExpectSuccess().ExpectContains("Usage")
		})
	}
}

// TestReqCommandsHelp 测试需求命令帮助
func TestReqCommandsHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"list", []string{"req", "list", "--help"}},
		{"create", []string{"req", "create", "--help"}},
		{"update", []string{"req", "update", "--help"}},
		{"view", []string{"req", "view", "--help"}},
		{"delete", []string{"req", "delete", "--help"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lc := framework.NewCLI(t)
			res := lc.Run(tt.args...)
			res.ExpectSuccess().ExpectContains("Usage")
		})
	}
}

// TestPRCommandsHelp 测试 PR 命令帮助
func TestPRCommandsHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"list", []string{"pr", "list", "--help"}},
		{"create", []string{"pr", "create", "--help"}},
		{"merge", []string{"pr", "merge", "--help"}},
		{"review", []string{"pr", "review", "--help"}},
		{"patch", []string{"pr", "patch", "--help"}},
		{"patch-comment", []string{"pr", "patch-comment", "--help"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lc := framework.NewCLI(t)
			res := lc.Run(tt.args...)
			res.ExpectSuccess().ExpectContains("Usage")
		})
	}
}

// TestTaskCommandsHelp 测试任务命令帮助
func TestTaskCommandsHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"list", []string{"task", "list", "--help"}},
		{"create", []string{"task", "create", "--help"}},
		{"update", []string{"task", "update", "--help"}},
		{"delete", []string{"task", "delete", "--help"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lc := framework.NewCLI(t)
			res := lc.Run(tt.args...)
			res.ExpectSuccess().ExpectContains("Usage")
		})
	}
}

// TestLoginHelp 测试登录命令帮助
func TestLoginHelp(t *testing.T) {
	t.Parallel()
	lc := framework.NewCLI(t)

	res := lc.Run("login", "--help")
	res.ExpectSuccess().
		ExpectContains("Usage").
		ExpectContains("login")
}
