// Package mcp provides types and utilities for interacting with MCP (Model Context Protocol) servers.
package mcp

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MCPConfig represents the top-level MCP configuration file structure.
// It maps server names to their individual configurations.
type MCPConfig struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// ServerConfig holds the connection configuration for a single MCP server.
type ServerConfig struct {
	// Transport specifies the transport type: "sse", "streamable", or "stdio".
	// Defaults to "sse" when omitted for backward compatibility.
	Transport string `json:"transport,omitempty"`

	// Type is an alias for Transport, used for compatibility with other MCP clients
	// like Claude Desktop (e.g., "type": "streamable-http").
	// Supports: "local" (→stdio), "remote" (→streamable/sse), "sse", "streamable", "stdio".
	// If Transport is set, it takes precedence over Type.
	Type string `json:"type,omitempty"`

	// Enabled controls whether this server is active. Defaults to true.
	Enabled *bool `json:"enabled,omitempty"`

	// URL is the HTTP endpoint for "sse" and "streamable" transports.
	URL string `json:"url,omitempty"`

	// Command is the executable path for the "stdio" transport.
	Command string `json:"command,omitempty"`
	// CommandSlice is the command as an array for the "stdio" transport.
	// Used for joinai-code format: ["npx", "-y", "my-mcp-command"].
	CommandSlice []string `json:"commandSlice,omitempty"`
	// Args is the argument list passed to Command for the "stdio" transport.
	Args []string `json:"args,omitempty"`
	// Env is a map of additional environment variables for the "stdio" transport.
	// Alias: environment (joinai-code format).
	Env map[string]string `json:"env,omitempty"`
	// Environment is an alias for Env (joinai-code format).
	Environment map[string]string `json:"environment,omitempty"`

	// Headers is a map of HTTP headers for remote transports.
	// Supports environment variable placeholders: "{env:VAR_NAME}".
	Headers map[string]string `json:"headers,omitempty"`

	// OAuth indicates OAuth authentication is required.
	OAuth any `json:"oauth,omitempty"`

	// Timeout is the connection/request timeout in milliseconds. Defaults to 30000.
	Timeout int `json:"timeout,omitempty"`
}

// ToolInfo holds the metadata for a single tool exposed by a MCP server.
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// InputSchema is the JSON Schema describing expected input parameters.
	InputSchema any `json:"inputSchema,omitempty"`
}

// ServerInfo is the list-mode output for a single server.
type ServerInfo struct {
	Name      string   `json:"name"`
	Transport string   `json:"transport"`
	URL       string   `json:"url,omitempty"`     // set for "sse" and "streamable"
	Command   string   `json:"command,omitempty"` // set for "stdio"
	Tools     []string `json:"tools,omitempty"`   // set when listing with tools, omitted when nil
	Error     string   `json:"error,omitempty"`   // set when connection fails
}

// ToolMatch holds a matched tool together with the server it was found on.
type ToolMatch struct {
	ServerName string
	Tool       *ToolInfo
}

// MCPError is a structured error carrying an error code, human-readable message,
// and optional key/value details.
type MCPError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e *MCPError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Error code constants used throughout the mcp package.
const (
	ErrCodeConfigNotFound  = "MCP_CONFIG_NOT_FOUND"
	ErrCodeConnectFailed   = "MCP_CONNECT_FAILED"
	ErrCodeServerNotFound  = "MCP_SERVER_NOT_FOUND"
	ErrCodeMethodNotFound  = "MCP_METHOD_NOT_FOUND"
	ErrCodeMethodAmbiguous = "MCP_METHOD_AMBIGUOUS"
	ErrCodeCallFailed      = "MCP_CALL_FAILED"
	ErrCodeParamInvalid    = "MCP_PARAM_INVALID"
)

// InferTransportType determines the transport type when not explicitly specified.
// Priority:
//   1. If config.Transport is set, use it directly
//   2. If config.Type is set, normalize it (e.g., "streamable-http" → "streamable")
//   3. If Command is provided, use "stdio"
//   4. If URL contains "sse" (case-insensitive), use "sse"
//   5. Default to "streamable" (modern protocol default)
func InferTransportType(cfg ServerConfig) string {
	if cfg.Transport != "" {
		return cfg.Transport
	}
	if cfg.Type != "" {
		return normalizeTransportType(cfg.Type)
	}
	if cfg.Command != "" {
		return "stdio"
	}
	if containsIgnoreCase(cfg.URL, "sse") {
		return "sse"
	}
	if containsIgnoreCase(cfg.URL, "stream") {
		return "streamable"
	}
	return "streamable"
}

// normalizeTransportType converts various type aliases to internal transport names.
// Supported mappings:
//   - contains "stream" (case-insensitive) → "streamable" (e.g., "streamable-http", "streaming")
//   - contains "sse" (case-insensitive) → "sse" (e.g., "sse", "server-sent-events")
//   - contains "command" or "stdio" → "stdio"
func normalizeTransportType(t string) string {
	if containsIgnoreCase(t, "stream") {
		return "streamable"
	}
	if containsIgnoreCase(t, "sse") {
		return "sse"
	}
	if containsIgnoreCase(t, "command") || containsIgnoreCase(t, "stdio") {
		return "stdio"
	}
	return t
}

// containsIgnoreCase checks if s contains substr (case-insensitive).
func containsIgnoreCase(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// ParseKVArgs parses a slice of "key=value" or "key:type=value" strings into a map.
// The first "=" in each token is used as the delimiter, so values may contain "=".
// Supports type annotations: key:string=val, key:number=123, key:bool=true, key:float=1.5
// If type is omitted, defaults to string.
// Returns an MCPError with ErrCodeParamInvalid if any token lacks an "=" or has invalid type.
func ParseKVArgs(args []string) (map[string]any, error) {
	result := make(map[string]any, len(args))
	for _, arg := range args {
		idx := -1
		for i, ch := range arg {
			if ch == '=' {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil, &MCPError{
				Code:    ErrCodeParamInvalid,
				Message: fmt.Sprintf("参数格式错误：%q（应为 key=value 或 key:type=value 格式）", arg),
			}
		}
		keyPart := arg[:idx]
		valStr := arg[idx+1:]

		// Parse key and optional type annotation: key or key:type
		key, typeHint, err := parseKeyWithType(keyPart)
		if err != nil {
			return nil, err
		}

		// Convert value based on type hint
		convertedVal, err := convertValue(valStr, typeHint)
		if err != nil {
			return nil, &MCPError{
				Code:    ErrCodeParamInvalid,
				Message: fmt.Sprintf("参数 %q 的值 %q 无法转换为类型 %q: %v", key, valStr, typeHint, err),
			}
		}
		result[key] = convertedVal
	}
	return result, nil
}

// parseKeyWithType parses "key" or "key:type" and returns the key and type hint.
// If no type is specified, returns "string" as default.
func parseKeyWithType(keyPart string) (string, string, error) {
	// Find colon separator
	colonIdx := -1
	for i, ch := range keyPart {
		if ch == ':' {
			colonIdx = i
			break
		}
	}

	// No type annotation
	if colonIdx < 0 {
		return keyPart, "string", nil
	}

	key := keyPart[:colonIdx]
	typeHint := keyPart[colonIdx+1:]

	if key == "" {
		return "", "", &MCPError{
			Code:    ErrCodeParamInvalid,
			Message: fmt.Sprintf("参数键不能为空"),
		}
	}

	// Validate type hint
	validTypes := map[string]bool{"string": true, "number": true, "int": true, "float": true, "bool": true, "boolean": true}
	if !validTypes[typeHint] {
		return "", "", &MCPError{
			Code:    ErrCodeParamInvalid,
			Message: fmt.Sprintf("参数 %q 使用了不支持的类型 %q，支持的类型：string, number, int, float, bool", key, typeHint),
		}
	}

	return key, typeHint, nil
}

// ParamInfo holds information about a single parameter for tool info output.
type ParamInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

// FormatInputSchema converts JSON Schema to human-readable "key:type=value" format.
// Returns a slice of strings like ["a:number={value}", "b:number={value}"].
func FormatInputSchema(schema any) []string {
	if schema == nil {
		return nil
	}

	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil
	}

	properties, _ := schemaMap["properties"].(map[string]any)
	if properties == nil {
		return nil
	}

	requiredSet := make(map[string]bool)
	if required, ok := schemaMap["required"].([]any); ok {
		for _, r := range required {
			if s, ok := r.(string); ok {
				requiredSet[s] = true
			}
		}
	}

	var result []string
	for key, prop := range properties {
		propMap, ok := prop.(map[string]any)
		if !ok {
			continue
		}

		jsonType, _ := propMap["type"].(string)
		typeHint := jsonTypeToTypeHint(jsonType)

		// Build the parameter line with {value} format
		line := key + ":" + typeHint + "={value}"
		if desc, ok := propMap["description"].(string); ok && desc != "" {
			line += " // " + desc
		}
		result = append(result, line)
	}

	return result
}

// GetRequiredParams extracts the list of required parameter names from JSON Schema.
func GetRequiredParams(schema any) []string {
	if schema == nil {
		return nil
	}

	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil
	}

	var result []string
	if required, ok := schemaMap["required"].([]any); ok {
		for _, r := range required {
			if s, ok := r.(string); ok {
				result = append(result, s)
			}
		}
	}
	return result
}

// GetParamInfoList extracts structured parameter information from JSON Schema.
func GetParamInfoList(schema any) []ParamInfo {
	if schema == nil {
		return nil
	}

	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil
	}

	properties, _ := schemaMap["properties"].(map[string]any)
	if properties == nil {
		return nil
	}

	requiredSet := make(map[string]bool)
	if required, ok := schemaMap["required"].([]any); ok {
		for _, r := range required {
			if s, ok := r.(string); ok {
				requiredSet[s] = true
			}
		}
	}

	var result []ParamInfo
	for key, prop := range properties {
		propMap, ok := prop.(map[string]any)
		if !ok {
			continue
		}

		jsonType, _ := propMap["type"].(string)
		typeHint := jsonTypeToTypeHint(jsonType)

		info := ParamInfo{
			Name:        key,
			Type:        typeHint,
			Required:    requiredSet[key],
			Description: "",
		}
		if desc, ok := propMap["description"].(string); ok {
			info.Description = desc
		}
		result = append(result, info)
	}

	return result
}

// jsonTypeToTypeHint converts JSON Schema type to our type hint.
func jsonTypeToTypeHint(jsonType string) string {
	switch jsonType {
	case "number":
		return "number"
	case "integer":
		return "int"
	case "boolean":
		return "bool"
	case "array":
		return "array"
	case "object":
		return "object"
	case "string":
		return "string"
	default:
		return "string"
	}
}

// convertValue converts a string value to the specified type.
func convertValue(valStr, typeHint string) (any, error) {
	switch typeHint {
	case "string":
		return valStr, nil
	case "number", "int":
		// Try integer first
		if intVal, err := strconv.ParseInt(valStr, 10, 64); err == nil {
			return float64(intVal), nil
		}
		// Try float
		return strconv.ParseFloat(valStr, 64)
	case "float":
		return strconv.ParseFloat(valStr, 64)
	case "bool", "boolean":
		lower := strings.ToLower(valStr)
		if lower == "true" || lower == "1" || lower == "yes" {
			return true, nil
		}
		if lower == "false" || lower == "0" || lower == "no" {
			return false, nil
		}
		return nil, fmt.Errorf("无法将 %q 解析为布尔值", valStr)
	default:
		return valStr, nil
	}
}

// ──────────────────────────────────────────────
// Environment variable resolution and header helpers
// ──────────────────────────────────────────────

// resolveEnvPlaceholder replaces {env:VAR_NAME} with the value of the environment variable.
// If the variable is not set and required is true, returns an error.
// If the variable is not set and required is false, returns the original string.
func resolveEnvPlaceholder(s string, required bool) (string, error) {
	const prefix = "{env:"
	const suffix = "}"

	if !strings.Contains(s, prefix) {
		return s, nil
	}

	result := s
	for {
		start := strings.Index(result, prefix)
		if start < 0 {
			break
		}

		end := strings.Index(result[start:], suffix)
		if end < 0 {
			break
		}
		end += start

		varName := result[start+len(prefix) : end]
		varValue := os.Getenv(varName)

		if varValue == "" && required {
			return "", fmt.Errorf("required environment variable %q is not set", varName)
		}

		result = result[:start] + varValue + result[end+len(suffix):]
	}

	return result, nil
}

// ResolveHeaders returns a copy of headers with environment variable placeholders resolved.
// Supports format: "Bearer {env:API_KEY}".
func (cfg ServerConfig) ResolveHeaders() (map[string]string, error) {
	if len(cfg.Headers) == 0 {
		return nil, nil
	}

	resolved := make(map[string]string, len(cfg.Headers))
	for key, value := range cfg.Headers {
		resolvedValue, err := resolveEnvPlaceholder(value, false)
		if err != nil {
			return nil, fmt.Errorf("header %q: %w", key, err)
		}
		resolved[key] = resolvedValue
	}
	return resolved, nil
}

// IsEnabled returns true if the server is enabled (default is true when not specified).
func (cfg ServerConfig) IsEnabled() bool {
	if cfg.Enabled == nil {
		return true
	}
	return *cfg.Enabled
}

// GetEffectiveCommand returns the effective command and arguments for stdio transport.
// Handles both "command" + "args" and "commandSlice" formats.
func (cfg ServerConfig) GetEffectiveCommand() (string, []string) {
	// commandSlice takes precedence
	if len(cfg.CommandSlice) > 0 {
		return cfg.CommandSlice[0], cfg.CommandSlice[1:]
	}
	return cfg.Command, cfg.Args
}

// GetEffectiveEnv returns the effective environment variables map.
// Merges both "env" and "environment" fields (env takes precedence).
func (cfg ServerConfig) GetEffectiveEnv() map[string]string {
	if len(cfg.Env) == 0 && len(cfg.Environment) == 0 {
		return nil
	}

	result := make(map[string]string)
	for k, v := range cfg.Environment {
		result[k] = v
	}
	for k, v := range cfg.Env {
		result[k] = v
	}
	return result
}
