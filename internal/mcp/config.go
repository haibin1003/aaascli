package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ErrConfigNotFound is returned when none of the standard config paths contain a valid file.
var ErrConfigNotFound = errors.New("未找到 MCP 配置文件")

// configSearchPaths lists the MCP configuration file locations in priority order.
// The actual value is populated by init() in config_paths.go, which calls
// buildSearchPaths() to select platform-specific paths at runtime via runtime.GOOS.
// All existing files are loaded and merged; higher-priority files win on name conflicts.
var configSearchPaths []string

// SearchPaths returns the ordered list of paths that LoadConfig searches.
func SearchPaths() []string {
	// Return a copy so callers cannot mutate the package-level slice.
	result := make([]string, len(configSearchPaths))
	copy(result, configSearchPaths)
	return result
}

// rawMCPConfig is used for unmarshalling config files that may use either
// "mcpServers" (legacy) or "mcp" (joinai-code format) as the root key.
type rawMCPConfig struct {
	MCPServers map[string]json.RawMessage `json:"mcpServers"`
	MCP        map[string]json.RawMessage `json:"mcp"`
}

// LoadConfig searches ALL standard paths and merges every MCP server it finds into
// a single MCPConfig.  When the same server name appears in more than one file the
// higher-priority file (lower index in configSearchPaths) wins.
//
// Supports both legacy "mcpServers" and joinai-code "mcp" root keys.
// Filters out disabled servers (enabled: false).
// Maps type aliases: "local" → "stdio", "remote" → "streamable".
//
// Returns the merged config, the list of files that were actually loaded, and an error.
// Returns ErrConfigNotFound (wrapped in MCPError) when no file exists at any path.
func LoadConfig() (*MCPConfig, []string, error) {
	merged := &MCPConfig{MCPServers: make(map[string]ServerConfig)}
	var foundPaths []string

	for _, path := range configSearchPaths {
		expanded := path // paths are fully expanded by buildSearchPaths() in config_paths.go

		data, err := os.ReadFile(expanded)
		if err != nil {
			// File not accessible – try the next path.
			continue
		}

		servers, err := parseConfigFile(data)
		if err != nil {
			// Unparseable file – report it as a hard error to avoid silent data loss.
			return nil, nil, fmt.Errorf("配置文件解析失败 (%s): %w", expanded, err)
		}

		foundPaths = append(foundPaths, expanded)

		// Merge servers: first-found (higher priority) wins on name conflicts.
		for name, serverCfg := range servers {
			if _, exists := merged.MCPServers[name]; !exists {
				merged.MCPServers[name] = serverCfg
			}
		}
	}

	if len(foundPaths) == 0 {
		return nil, nil, &MCPError{
			Code:    ErrCodeConfigNotFound,
			Message: ErrConfigNotFound.Error(),
			Details: map[string]any{"searchPaths": SearchPaths()},
		}
	}

	return merged, foundPaths, nil
}

// parseConfigFile parses a single config file and returns the map of server configurations.
// It supports both "mcpServers" and "mcp" root keys, with "mcpServers" taking precedence.
func parseConfigFile(data []byte) (map[string]ServerConfig, error) {
	var raw rawMCPConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	servers := make(map[string]ServerConfig)

	// Determine which root key to use (mcpServers takes precedence over mcp)
	var source map[string]json.RawMessage
	if len(raw.MCPServers) > 0 {
		source = raw.MCPServers
	} else {
		source = raw.MCP
	}

	for name, rawCfg := range source {
		var cfg ServerConfig
		if err := json.Unmarshal(rawCfg, &cfg); err != nil {
			return nil, fmt.Errorf("server %q: %w", name, err)
		}

		// Skip disabled servers
		if !cfg.IsEnabled() {
			continue
		}

		// Map type aliases for compatibility
		cfg = normalizeServerConfig(cfg)

		servers[name] = cfg
	}

	return servers, nil
}

// normalizeServerConfig maps joinai-code type aliases to internal transport types.
//   - "local" → "stdio"
//   - "remote" → "streamable" (or "sse" if URL contains "sse")
func normalizeServerConfig(cfg ServerConfig) ServerConfig {
	// Get the effective type from Type or Transport field
	effectiveType := cfg.Type
	if cfg.Transport != "" {
		effectiveType = cfg.Transport
	}

	// Map joinai-code aliases to internal transport types
	switch strings.ToLower(effectiveType) {
	case "local":
		cfg.Transport = "stdio"
	case "remote":
		// Infer from URL: contains "sse" → "sse", otherwise "streamable"
		if containsIgnoreCase(cfg.URL, "sse") {
			cfg.Transport = "sse"
		} else {
			cfg.Transport = "streamable"
		}
	}

	// Clear Type field to avoid confusion (Transport now has the canonical value)
	cfg.Type = ""

	return cfg
}

// HasServer 检查指定的 MCP Server 是否已配置
// 用于工具命令执行前检查依赖的 Server 是否可用
func (cfg *MCPConfig) HasServer(name string) bool {
	if cfg == nil || cfg.MCPServers == nil {
		return false
	}
	_, exists := cfg.MCPServers[name]
	return exists
}
