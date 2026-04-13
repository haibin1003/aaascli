// Package common provides shared utilities for command execution.
package common

import (
	"fmt"
	"strings"

	"github.com/user/lc/internal/api"
	"github.com/user/lc/internal/config"
	"go.uber.org/zap"
)

// CommandContext holds all dependencies needed for command execution.
// It provides a unified way to initialize config, HTTP client, and logger.
type CommandContext struct {
	Config    *config.Config
	Client    *api.Client
	Logger    *zap.Logger
	cleanup   func()
	DebugMode bool
	Insecure  bool
	DryRun    bool
}

// NewCommandContext creates a new command context with all dependencies initialized.
// It handles config loading, logger setup, and HTTP client creation.
//
// Usage:
//
//	ctx, err := common.NewCommandContext(debugMode, insecureSkipVerify, dryRun, cookie)
//	if err != nil {
//	    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
//	    os.Exit(1)
//	}
//	defer ctx.Close()
func NewCommandContext(debugMode, insecure bool, dryRun bool, cookie string) (*CommandContext, error) {
	ctx := &CommandContext{
		DebugMode: debugMode,
		Insecure:  insecure,
		DryRun:    dryRun,
	}

	// Initialize logger
	if err := ctx.initLogger(); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize config
	if err := ctx.initConfig(cookie); err != nil {
		ctx.cleanup()
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	// Initialize HTTP client
	ctx.initClient()

	return ctx, nil
}

// initLogger initializes the zap logger based on debug mode.
func (c *CommandContext) initLogger() error {
	if !c.DebugMode {
		c.Logger = zap.NewNop()
		c.cleanup = func() {}
		return nil
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	c.Logger = logger
	c.cleanup = func() {
		_ = logger.Sync()
	}

	return nil
}

// initConfig loads configuration from default path.
func (c *CommandContext) initConfig(cookie string) error {
	cfg := config.NewConfig()

	// Try to load from default config path
	if loadedCfg, err := config.LoadConfigWithDefaults(config.GetDefaultConfigPath()); err == nil {
		cfg = loadedCfg
	}

	// Override cookie if provided via flag
	if cookie != "" {
		// Clean up the cookie value (trim whitespace, quotes, and MOSS_SESSION= prefix if present)
		cookie = strings.TrimSpace(cookie)
		cookie = strings.TrimPrefix(cookie, "MOSS_SESSION=")
		cookie = strings.Trim(cookie, `"'`)

		// Set the cookie with MOSS_SESSION= prefix
		cfg.Cookie = "MOSS_SESSION=" + cookie
		cfg.API.Headers["Cookie"] = cfg.Cookie
	}

	c.Config = cfg
	return nil
}

// initClient creates the HTTP client based on insecure flag.
func (c *CommandContext) initClient() {
	if c.Insecure {
		c.Client = api.NewInsecureClientWithLogger(c.Logger)
	} else {
		c.Client = api.NewClientWithLogger(c.Logger)
	}
}

// Close cleans up resources. Should be called with defer.
func (c *CommandContext) Close() {
	if c.cleanup != nil {
		c.cleanup()
	}
}

// GetHeaders returns HTTP headers with workspace key.
func (c *CommandContext) GetHeaders(workspaceKey string) map[string]string {
	return c.Config.GetHeadersWithWorkspace(workspaceKey)
}

// Debug logs a debug message if debug mode is enabled.
func (c *CommandContext) Debug(msg string, fields ...zap.Field) {
	if c.DebugMode && c.Logger != nil {
		c.Logger.Debug(msg, fields...)
	}
}
