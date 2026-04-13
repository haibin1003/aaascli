package mcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// serverClient wraps the MCP SDK for a single named server.
type serverClient struct {
	name   string
	config ServerConfig
}

// newServerClient constructs a client for a specific server.
func newServerClient(name string, cfg ServerConfig) *serverClient {
	return &serverClient{name: name, config: cfg}
}

// buildTransport constructs the appropriate SDK transport based on the server config.
// Supported transport types: "sse" (default), "streamable", "stdio".
// Transport type is auto-detected from URL (contains "streamable") or Command presence
// when not explicitly specified.
func (c *serverClient) buildTransport() (sdkmcp.Transport, error) {
	transportType := InferTransportType(c.config)

	switch transportType {
	case "sse":
		httpClient, err := c.buildHTTPClient()
		if err != nil {
			return nil, err
		}
		transport := &sdkmcp.SSEClientTransport{
			Endpoint: c.config.URL,
		}
		if httpClient != nil {
			transport.HTTPClient = httpClient
		}
		return transport, nil

	case "streamable":
		httpClient, err := c.buildHTTPClient()
		if err != nil {
			return nil, err
		}
		transport := &sdkmcp.StreamableClientTransport{
			Endpoint: c.config.URL,
		}
		if httpClient != nil {
			transport.HTTPClient = httpClient
		}
		return transport, nil

	case "stdio":
		cmd, args := c.config.GetEffectiveCommand()
		if cmd == "" {
			return nil, &MCPError{
				Code:    ErrCodeConnectFailed,
				Message: fmt.Sprintf("Server '%s' 使用 stdio 传输但未指定 command 字段", c.name),
			}
		}
		execCmd := exec.Command(cmd, args...)
		// Use GetEffectiveEnv to support both "env" and "environment" fields
		if envVars := c.config.GetEffectiveEnv(); len(envVars) > 0 {
			execCmd.Env = os.Environ()
			for k, v := range envVars {
				execCmd.Env = append(execCmd.Env, k+"="+v)
			}
		}
		return &sdkmcp.CommandTransport{Command: execCmd}, nil

	default:
		return nil, &MCPError{
			Code:    ErrCodeConnectFailed,
			Message: fmt.Sprintf("Server '%s' 使用了不支持的传输类型 '%s'，支持的类型：sse、streamable、stdio", c.name, transportType),
		}
	}
}

// connect establishes an MCP session with the server.
// It applies the timeout from the server config (defaulting to 30 s).
// The returned context and session share the same deadline;
// callers must invoke cleanup() when done to release resources.
func (c *serverClient) connect(parent context.Context) (context.Context, *sdkmcp.ClientSession, func(), error) {
	timeout := time.Duration(c.config.Timeout) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(parent, timeout)

	transport, err := c.buildTransport()
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}

	sdkClient := sdkmcp.NewClient(&sdkmcp.Implementation{
		Name:    "lc-cli",
		Version: "1.0.0",
	}, nil)

	session, err := sdkClient.Connect(ctx, transport, nil)
	if err != nil {
		cancel()
		transportType := InferTransportType(c.config)
		msg := fmt.Sprintf("连接 MCP Server '%s' 失败 (使用 %s 传输): %v\n\n", c.name, transportType, err)
		msg += "请检查您的配置是否正确。以下是各传输类型的配置示例：\n\n"
		msg += "1. SSE (传统协议):\n"
		msg += `   { "url": "http://host/mcp/sse" }` + "\n\n"
		msg += "2. Streamable HTTP (现代协议，默认):\n"
		msg += `   { "url": "http://host/mcp" }` + " 或 " + `{ "type": "streamable-http", "url": "http://host/mcp" }` + "\n\n"
		msg += "3. stdio (本地命令):\n"
		msg += `   { "command": "npx", "args": ["-y", "@server/mcp"] }` + "\n\n"
		msg += "4. joinai-code 格式 (本地):\n"
		msg += `   { "type": "local", "command": ["npx", "-y", "@server/mcp"] }` + "\n\n"
		msg += "5. joinai-code 格式 (远程，带认证):\n"
		msg += `   { "type": "remote", "url": "https://api.example.com/mcp", "headers": { "Authorization": "Bearer {env:API_KEY}" } }` + "\n\n"
		msg += "如果 URL 包含 'sse' 会自动使用 SSE 传输，包含 'stream' 会自动使用 Streamable HTTP。"

		details := map[string]any{
			"transport": transportType,
			"hint":      "请检查配置或显式指定 type/transport 字段",
		}
		if c.config.URL != "" {
			details["url"] = c.config.URL
		}
		if c.config.Command != "" {
			details["command"] = c.config.Command
		}
		return nil, nil, nil, &MCPError{
			Code:    ErrCodeConnectFailed,
			Message: msg,
			Details: details,
		}
	}

	cleanup := func() {
		_ = session.Close()
		cancel()
	}
	return ctx, session, cleanup, nil
}

// listTools connects to the server and returns all available tools.
func (c *serverClient) listTools(parent context.Context) ([]*ToolInfo, error) {
	ctx, session, cleanup, err := c.connect(parent)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		return nil, &MCPError{
			Code:    ErrCodeConnectFailed,
			Message: fmt.Sprintf("获取 Server '%s' 工具列表失败: %v", c.name, err),
		}
	}

	tools := make([]*ToolInfo, 0, len(result.Tools))
	for _, t := range result.Tools {
		tools = append(tools, &ToolInfo{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}
	return tools, nil
}

// callTool connects to the server and calls the named tool with the given arguments.
func (c *serverClient) callTool(parent context.Context, method string, args map[string]any) (any, error) {
	ctx, session, cleanup, err := c.connect(parent)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	params := &sdkmcp.CallToolParams{
		Name:      method,
		Arguments: args,
	}

	res, err := session.CallTool(ctx, params)
	if err != nil {
		return nil, &MCPError{
			Code:    ErrCodeCallFailed,
			Message: fmt.Sprintf("调用工具 '%s' 失败: %v", method, err),
		}
	}

	if res.IsError {
		content := extractContent(res.Content)
		return nil, &MCPError{
			Code:    ErrCodeCallFailed,
			Message: fmt.Sprintf("工具 '%s' 返回错误", method),
			Details: map[string]any{"toolError": content},
		}
	}

	return extractContent(res.Content), nil
}

// extractContent converts a slice of MCP Content items into a JSON-friendly value.
// A single TextContent item is unwrapped to just its text string.
// Multiple items or non-text items are returned as a slice of objects.
func extractContent(contents []sdkmcp.Content) any {
	if len(contents) == 0 {
		return nil
	}

	// Single text item – return the raw text for readability.
	if len(contents) == 1 {
		if tc, ok := contents[0].(*sdkmcp.TextContent); ok {
			return tc.Text
		}
	}

	// Multiple items or non-text – wrap each in a typed object.
	result := make([]map[string]any, 0, len(contents))
	for _, c := range contents {
		switch tc := c.(type) {
		case *sdkmcp.TextContent:
			result = append(result, map[string]any{"type": "text", "text": tc.Text})
		case *sdkmcp.ImageContent:
			result = append(result, map[string]any{"type": "image", "mimeType": tc.MIMEType})
		default:
			result = append(result, map[string]any{"type": "unknown"})
		}
	}
	return result
}

// buildHTTPClient creates an HTTP client with custom headers if configured.
// Returns nil if no headers are configured, which lets the transport use the default client.
func (c *serverClient) buildHTTPClient() (*http.Client, error) {
	headers, err := c.config.ResolveHeaders()
	if err != nil {
		return nil, &MCPError{
			Code:    ErrCodeConnectFailed,
			Message: fmt.Sprintf("Server '%s' headers 解析失败: %v", c.name, err),
		}
	}

	// No headers configured, use default
	if len(headers) == 0 {
		return nil, nil
	}

	// Create a custom transport that adds headers to each request
	baseTransport := http.DefaultTransport
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		// Clone to avoid modifying the global default
		baseTransport = t.Clone()
	}

	transport := &headerTransport{
		base:    baseTransport,
		headers: headers,
	}

	return &http.Client{Transport: transport}, nil
}

// headerTransport wraps an http.RoundTripper and adds custom headers to each request.
type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	newReq := req.Clone(req.Context())
	for key, value := range t.headers {
		newReq.Header.Set(key, value)
	}
	return t.base.RoundTrip(newReq)
}
