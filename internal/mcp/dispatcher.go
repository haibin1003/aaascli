package mcp

import (
	"context"
	"fmt"
	"sync"
)

// Dispatcher coordinates operations across all configured MCP servers.
type Dispatcher struct {
	config      *MCPConfig
	configPaths []string
}

// NewDispatcher creates a Dispatcher from a loaded MCPConfig.
func NewDispatcher(cfg *MCPConfig, configPaths []string) *Dispatcher {
	return &Dispatcher{config: cfg, configPaths: configPaths}
}

// buildServerInfo constructs a ServerInfo skeleton from a server config.
// It populates Transport, URL, and Command based on the transport type.
func buildServerInfo(name string, cfg ServerConfig) ServerInfo {
	transport := InferTransportType(cfg)
	info := ServerInfo{
		Name:      name,
		Transport: transport,
	}
	switch transport {
	case "stdio":
		info.Command = cfg.Command
	default:
		info.URL = cfg.URL
	}
	return info
}

// ConfigPaths returns the list of config files that were loaded and merged.
func (d *Dispatcher) ConfigPaths() []string {
	return d.configPaths
}

// ListServersConfig returns server configurations without connecting to fetch tools.
// This is fast and only shows server metadata (name, transport, url/command).
func (d *Dispatcher) ListServersConfig() []ServerInfo {
	servers := make([]ServerInfo, 0, len(d.config.MCPServers))
	for name, cfg := range d.config.MCPServers {
		servers = append(servers, buildServerInfo(name, cfg))
	}
	return servers
}

// ListAllServers connects to every configured server concurrently and returns
// a ServerInfo entry for each.  If a server is unreachable the entry will have
// a non-empty Error field; the rest of the entries are still returned.
func (d *Dispatcher) ListAllServers(ctx context.Context) []ServerInfo {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		servers []ServerInfo
	)

	for name, cfg := range d.config.MCPServers {
		name, cfg := name, cfg // capture loop variables
		wg.Add(1)
		go func() {
			defer wg.Done()

			client := newServerClient(name, cfg)
			info := buildServerInfo(name, cfg)

			tools, err := client.listTools(ctx)
			if err != nil {
				info.Error = err.Error()
			} else {
				info.Tools = make([]string, 0, len(tools))
				for _, t := range tools {
					info.Tools = append(info.Tools, t.Name)
				}
			}

			mu.Lock()
			servers = append(servers, info)
			mu.Unlock()
		}()
	}

	wg.Wait()
	return servers
}

// GetServerInfo connects to a single named server and returns its ServerInfo.
// Returns MCP_SERVER_NOT_FOUND if the server name is not in the config.
func (d *Dispatcher) GetServerInfo(ctx context.Context, serverName string) (ServerInfo, error) {
	cfg, ok := d.config.MCPServers[serverName]
	if !ok {
		return ServerInfo{}, &MCPError{
			Code:    ErrCodeServerNotFound,
			Message: fmt.Sprintf("配置中不存在 Server '%s'，请使用 `lc mcp` 查看可用 Server", serverName),
		}
	}

	client := newServerClient(serverName, cfg)
	info := buildServerInfo(serverName, cfg)

	tools, err := client.listTools(ctx)
	if err != nil {
		info.Error = err.Error()
	} else {
		info.Tools = make([]string, 0, len(tools))
		for _, t := range tools {
			info.Tools = append(info.Tools, t.Name)
		}
	}
	return info, nil
}

// GetToolInfo connects to the named server and returns the ToolInfo for the specified tool.
// Returns MCP_SERVER_NOT_FOUND if the server is not configured.
// Returns MCP_METHOD_NOT_FOUND if the tool is not exposed by that server.
func (d *Dispatcher) GetToolInfo(ctx context.Context, serverName, toolName string) (*ToolMatch, error) {
	cfg, ok := d.config.MCPServers[serverName]
	if !ok {
		return nil, &MCPError{
			Code:    ErrCodeServerNotFound,
			Message: fmt.Sprintf("配置中不存在 Server '%s'，请使用 `lc mcp` 查看可用 Server", serverName),
		}
	}

	client := newServerClient(serverName, cfg)
	tools, err := client.listTools(ctx)
	if err != nil {
		return nil, &MCPError{
			Code:    ErrCodeConnectFailed,
			Message: fmt.Sprintf("连接 Server '%s' 失败: %v", serverName, err),
		}
	}

	for _, t := range tools {
		if t.Name == toolName {
			return &ToolMatch{ServerName: serverName, Tool: t}, nil
		}
	}

	return nil, &MCPError{
		Code:    ErrCodeMethodNotFound,
		Message: fmt.Sprintf("Server '%s' 上未找到方法 '%s'，请使用 `lc mcp %s` 查看可用方法", serverName, toolName, serverName),
	}
}

// FindTool searches all servers for a tool with the given name.
//
//   - If exactly one server has the tool, it is returned as *ToolMatch.
//   - If no server has the tool, an MCPError with ErrCodeMethodNotFound is returned.
//   - If multiple servers have the tool and preferServer is empty, an MCPError with
//     ErrCodeMethodAmbiguous is returned together with the slice of all matches.
//   - If multiple servers have the tool and preferServer is set, the match on that
//     server is returned (or ErrCodeMethodNotFound if that server does not have it).
func (d *Dispatcher) FindTool(ctx context.Context, method, preferServer string) (*ToolMatch, []*ToolMatch, error) {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		matches []*ToolMatch
	)

	for name, cfg := range d.config.MCPServers {
		name, cfg := name, cfg
		wg.Add(1)
		go func() {
			defer wg.Done()

			client := newServerClient(name, cfg)
			tools, err := client.listTools(ctx)
			if err != nil {
				// Skip unreachable servers silently during discovery.
				return
			}

			for _, t := range tools {
				if t.Name == method {
					mu.Lock()
					matches = append(matches, &ToolMatch{ServerName: name, Tool: t})
					mu.Unlock()
					break
				}
			}
		}()
	}
	wg.Wait()

	if len(matches) == 0 {
		return nil, nil, &MCPError{
			Code:    ErrCodeMethodNotFound,
			Message: fmt.Sprintf("未在任何 Server 中找到方法 '%s'，请使用 `lc mcp server` 查看可用方法", method),
		}
	}

	if len(matches) == 1 {
		return matches[0], nil, nil
	}

	// Multiple matches – try to resolve via preferServer.
	if preferServer != "" {
		for _, m := range matches {
			if m.ServerName == preferServer {
				return m, nil, nil
			}
		}
		// preferServer was given but is not among the matching servers.
		return nil, nil, &MCPError{
			Code:    ErrCodeMethodNotFound,
			Message: fmt.Sprintf("Server '%s' 上未找到方法 '%s'", preferServer, method),
		}
	}

	// Ambiguous – caller must specify --server.
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, m.ServerName)
	}
	return nil, matches, &MCPError{
		Code:    ErrCodeMethodAmbiguous,
		Message: fmt.Sprintf("方法 '%s' 存在于多个 Server 中，请使用 --server 参数指定", method),
		Details: map[string]any{"matches": names},
	}
}

// CallTool resolves which server hosts method, then calls it with the provided args.
// It returns the server name and the tool result.
//
// When preferServer is non-empty and the server exists in the config, the tool is
// called directly without a discovery step (one fewer round-trip).
func (d *Dispatcher) CallTool(ctx context.Context, method, preferServer string, args map[string]any) (string, any, error) {
	// Fast path: caller knows the target server.
	if preferServer != "" {
		cfg, ok := d.config.MCPServers[preferServer]
		if !ok {
			return "", nil, &MCPError{
				Code:    ErrCodeServerNotFound,
				Message: fmt.Sprintf("配置中不存在 Server '%s'", preferServer),
			}
		}
		client := newServerClient(preferServer, cfg)
		result, err := client.callTool(ctx, method, args)
		if err != nil {
			return "", nil, err
		}
		return preferServer, result, nil
	}

	// Discovery path: find the server that has this tool, then call it.
	match, _, err := d.FindTool(ctx, method, "")
	if err != nil {
		return "", nil, err
	}

	client := newServerClient(match.ServerName, d.config.MCPServers[match.ServerName])
	result, callErr := client.callTool(ctx, method, args)
	if callErr != nil {
		return "", nil, callErr
	}
	return match.ServerName, result, nil
}
