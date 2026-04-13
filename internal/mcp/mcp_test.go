package mcp

import (
	"os"
	"testing"
)

// ──────────────────────────────────────────────
// ParseKVArgs tests
// ──────────────────────────────────────────────

func TestParseKVArgs_valid(t *testing.T) {
	got, err := ParseKVArgs([]string{"id=10086", "name=tom", "age=25"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]any{"id": "10086", "name": "tom", "age": "25"}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("key %q: got %q, want %q", k, got[k], v)
		}
	}
}

func TestParseKVArgs_valueContainsEquals(t *testing.T) {
	got, err := ParseKVArgs([]string{"token=a=b=c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["token"] != "a=b=c" {
		t.Errorf("got %q, want %q", got["token"], "a=b=c")
	}
}

func TestParseKVArgs_noEquals_returnsError(t *testing.T) {
	_, err := ParseKVArgs([]string{"id=10086", "badparam"})
	if err == nil {
		t.Fatal("expected error for missing '=', got nil")
	}
	mcpErr, ok := err.(*MCPError)
	if !ok {
		t.Fatalf("expected *MCPError, got %T", err)
	}
	if mcpErr.Code != ErrCodeParamInvalid {
		t.Errorf("got code %q, want %q", mcpErr.Code, ErrCodeParamInvalid)
	}
}

func TestParseKVArgs_empty(t *testing.T) {
	got, err := ParseKVArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestParseKVArgs_emptyValue(t *testing.T) {
	got, err := ParseKVArgs([]string{"key="})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["key"] != "" {
		t.Errorf("got %q, want empty string", got["key"])
	}
}

// ──────────────────────────────────────────────
// SearchPaths tests
// ──────────────────────────────────────────────

func TestSearchPaths_notEmpty(t *testing.T) {
	paths := SearchPaths()
	if len(paths) == 0 {
		t.Error("SearchPaths should return at least one path")
	}
}

func TestSearchPaths_returnsACopy(t *testing.T) {
	p1 := SearchPaths()
	p2 := SearchPaths()
	p1[0] = "mutated"
	if p2[0] == "mutated" {
		t.Error("SearchPaths should return a copy, not the internal slice")
	}
}

// ──────────────────────────────────────────────
// LoadConfig tests
// ──────────────────────────────────────────────

func TestLoadConfig_noneExist(t *testing.T) {
	orig := configSearchPaths
	configSearchPaths = []string{"/nonexistent/a.json", "/nonexistent/b.json"}
	defer func() { configSearchPaths = orig }()

	cfg, paths, err := LoadConfig()
	if cfg != nil {
		t.Errorf("expected nil config, got %+v", cfg)
	}
	if len(paths) != 0 {
		t.Errorf("expected empty paths, got %v", paths)
	}
	mcpErr, ok := err.(*MCPError)
	if !ok {
		t.Fatalf("expected *MCPError, got %T: %v", err, err)
	}
	if mcpErr.Code != ErrCodeConfigNotFound {
		t.Errorf("got code %q, want %q", mcpErr.Code, ErrCodeConfigNotFound)
	}
}

func TestLoadConfig_validFile(t *testing.T) {
	tmpFile := t.TempDir() + "/mcp.json"
	content := `{"mcpServers":{"test":{"url":"http://example.com","timeout":5000}}}`
	if err := os.WriteFile(tmpFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	orig := configSearchPaths
	configSearchPaths = []string{tmpFile}
	defer func() { configSearchPaths = orig }()

	cfg, paths, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 || paths[0] != tmpFile {
		t.Errorf("got paths %v, want [%q]", paths, tmpFile)
	}
	if cfg.MCPServers["test"].URL != "http://example.com" {
		t.Errorf("got URL %q, want %q", cfg.MCPServers["test"].URL, "http://example.com")
	}
	if cfg.MCPServers["test"].Timeout != 5000 {
		t.Errorf("got timeout %d, want 5000", cfg.MCPServers["test"].Timeout)
	}
}

func TestLoadConfig_mergesAllFiles(t *testing.T) {
	dir := t.TempDir()
	file1 := dir + "/first.json"
	file2 := dir + "/second.json"
	_ = os.WriteFile(file1, []byte(`{"mcpServers":{"a":{"url":"http://first"}}}`), 0600)
	_ = os.WriteFile(file2, []byte(`{"mcpServers":{"b":{"url":"http://second"}}}`), 0600)

	orig := configSearchPaths
	configSearchPaths = []string{file1, file2}
	defer func() { configSearchPaths = orig }()

	cfg, paths, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	// 两个文件都应被加载
	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %v", paths)
	}
	// 两个 server 都应存在
	if _, ok := cfg.MCPServers["a"]; !ok {
		t.Error("expected server 'a' from first file")
	}
	if _, ok := cfg.MCPServers["b"]; !ok {
		t.Error("expected server 'b' from second file")
	}
}

func TestLoadConfig_highPriorityWinsOnConflict(t *testing.T) {
	dir := t.TempDir()
	file1 := dir + "/first.json"
	file2 := dir + "/second.json"
	// 两个文件都定义了 server "shared"，URL 不同
	_ = os.WriteFile(file1, []byte(`{"mcpServers":{"shared":{"url":"http://high-priority"}}}`), 0600)
	_ = os.WriteFile(file2, []byte(`{"mcpServers":{"shared":{"url":"http://low-priority"}}}`), 0600)

	orig := configSearchPaths
	configSearchPaths = []string{file1, file2}
	defer func() { configSearchPaths = orig }()

	cfg, paths, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %v", paths)
	}
	// 高优先级（file1）的 URL 应生效
	if cfg.MCPServers["shared"].URL != "http://high-priority" {
		t.Errorf("high-priority file should win, got URL %q", cfg.MCPServers["shared"].URL)
	}
}

func TestLoadConfig_emptymcpServers(t *testing.T) {
	tmpFile := t.TempDir() + "/mcp.json"
	_ = os.WriteFile(tmpFile, []byte(`{"mcpServers":{}}`), 0600)

	orig := configSearchPaths
	configSearchPaths = []string{tmpFile}
	defer func() { configSearchPaths = orig }()

	cfg, paths, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 path, got %v", paths)
	}
	if len(cfg.MCPServers) != 0 {
		t.Errorf("expected empty MCPServers, got %v", cfg.MCPServers)
	}
}

// ──────────────────────────────────────────────
// Joinai-code format tests
// ──────────────────────────────────────────────

func TestLoadConfig_mcpRootKey(t *testing.T) {
	tmpFile := t.TempDir() + "/joinai-code.json"
	content := `{"mcp":{"myServer":{"type":"remote","url":"http://example.com"}}}`
	_ = os.WriteFile(tmpFile, []byte(content), 0600)

	orig := configSearchPaths
	configSearchPaths = []string{tmpFile}
	defer func() { configSearchPaths = orig }()

	cfg, _, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cfg.MCPServers["myServer"]; !ok {
		t.Error("expected server 'myServer' from mcp root key")
	}
}

func TestLoadConfig_enabledFilter(t *testing.T) {
	tmpFile := t.TempDir() + "/mcp.json"
	content := `{"mcpServers":{"enabledServer":{"url":"http://enabled.com"},"disabledServer":{"url":"http://disabled.com","enabled":false}}}`
	_ = os.WriteFile(tmpFile, []byte(content), 0600)

	orig := configSearchPaths
	configSearchPaths = []string{tmpFile}
	defer func() { configSearchPaths = orig }()

	cfg, _, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := cfg.MCPServers["enabledServer"]; !ok {
		t.Error("expected enabledServer to be present")
	}
	if _, ok := cfg.MCPServers["disabledServer"]; ok {
		t.Error("expected disabledServer to be filtered out")
	}
}

func TestLoadConfig_typeMapping(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedType string
	}{
		{
			name:         "local type maps to stdio",
			content:      `{"mcpServers":{"s":{"type":"local","command":"echo"}}}`,
			expectedType: "stdio",
		},
		{
			name:         "remote type maps to streamable",
			content:      `{"mcpServers":{"s":{"type":"remote","url":"http://example.com"}}}`,
			expectedType: "streamable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := t.TempDir() + "/mcp.json"
			_ = os.WriteFile(tmpFile, []byte(tt.content), 0600)

			orig := configSearchPaths
			configSearchPaths = []string{tmpFile}
			defer func() { configSearchPaths = orig }()

			cfg, _, err := LoadConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			server := cfg.MCPServers["s"]
			transport := InferTransportType(server)
			if transport != tt.expectedType {
				t.Errorf("expected transport %q, got %q", tt.expectedType, transport)
			}
		})
	}
}

// ──────────────────────────────────────────────
// ServerConfig helper tests
// ──────────────────────────────────────────────

func TestServerConfig_GetEffectiveCommand(t *testing.T) {
	tests := []struct {
		name        string
		cfg         ServerConfig
		wantCmd     string
		wantArgs    []string
	}{
		{
			name:        "commandSlice takes precedence",
			cfg:         ServerConfig{CommandSlice: []string{"npx", "-y", "pkg"}, Command: "ignored", Args: []string{"ignored"}},
			wantCmd:     "npx",
			wantArgs:    []string{"-y", "pkg"},
		},
		{
			name:        "fallback to command and args",
			cfg:         ServerConfig{Command: "echo", Args: []string{"hello"}},
			wantCmd:     "echo",
			wantArgs:    []string{"hello"},
		},
		{
			name:        "empty returns empty",
			cfg:         ServerConfig{},
			wantCmd:     "",
			wantArgs:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs := tt.cfg.GetEffectiveCommand()
			if gotCmd != tt.wantCmd {
				t.Errorf("GetEffectiveCommand() cmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("GetEffectiveCommand() args = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestServerConfig_IsEnabled(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name    string
		enabled *bool
		want    bool
	}{
		{"nil defaults to true", nil, true},
		{"explicit true", &trueVal, true},
		{"explicit false", &falseVal, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ServerConfig{Enabled: tt.enabled}
			if got := cfg.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerConfig_ResolveHeaders(t *testing.T) {
	t.Run("no headers returns nil", func(t *testing.T) {
		cfg := ServerConfig{}
		headers, err := cfg.ResolveHeaders()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if headers != nil {
			t.Errorf("expected nil, got %v", headers)
		}
	})

	t.Run("static headers unchanged", func(t *testing.T) {
		cfg := ServerConfig{Headers: map[string]string{"X-Key": "value"}}
		headers, err := cfg.ResolveHeaders()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if headers["X-Key"] != "value" {
			t.Errorf("expected X-Key=value, got %v", headers)
		}
	})

	t.Run("env placeholder resolved", func(t *testing.T) {
		os.Setenv("TEST_API_KEY", "secret123")
		defer os.Unsetenv("TEST_API_KEY")

		cfg := ServerConfig{Headers: map[string]string{"Authorization": "Bearer {env:TEST_API_KEY}"}}
		headers, err := cfg.ResolveHeaders()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if headers["Authorization"] != "Bearer secret123" {
			t.Errorf("expected 'Bearer secret123', got %v", headers["Authorization"])
		}
	})
}

func TestServerConfig_GetEffectiveEnv(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		cfg := ServerConfig{}
		if env := cfg.GetEffectiveEnv(); env != nil {
			t.Errorf("expected nil, got %v", env)
		}
	})

	t.Run("env takes precedence over environment", func(t *testing.T) {
		cfg := ServerConfig{
			Environment: map[string]string{"KEY1": "from-env", "KEY2": "from-env2"},
			Env:         map[string]string{"KEY1": "from-Env"},
		}
		env := cfg.GetEffectiveEnv()
		if env["KEY1"] != "from-Env" {
			t.Errorf("expected KEY1=from-Env, got %v", env["KEY1"])
		}
		if env["KEY2"] != "from-env2" {
			t.Errorf("expected KEY2=from-env2, got %v", env["KEY2"])
		}
	})
}
