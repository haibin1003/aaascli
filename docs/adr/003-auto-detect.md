# ADR 003: Context Auto-Detection Design

## Status

Accepted

## Context

The CLI tool frequently requires workspace key and other context parameters for executing commands. These parameters are repetitive for users working within the same Git repository. We need a mechanism to automatically detect these values from the environment to improve user experience.

## Decision

We implement a context auto-detection system that reads configuration from Git repository root.

## Rationale

### Requirements

1. **Zero-config for frequent users**: Users working in the same repo shouldn't need to specify workspace key repeatedly
2. **Explicit override**: Users can still specify parameters explicitly when needed
3. **Git-native**: Store configuration in Git, following the principle of "configuration as code"
4. **Multi-repo support**: Support working across multiple repositories

### Design

```
Git Repository Root
├── .lc/
│   └── config.json          # Repository-specific configuration
│       {
│         "workspace_key": "XXJSLJCLIDEV",
│         "project_code": "XXJSLJCLIDEV"
│       }
```

### Implementation

The auto-detection follows a priority order:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Git repository config** (`.lc/config.json`)
4. **Global config** (`~/.lc/config.json`)
5. **Error if required and not found**

### Code Structure

```go
// AutoDetectContext holds detected values
type AutoDetectContext struct {
    WorkspaceKey string
    ProjectCode  string
    RepoName     string
    RepoURL      string
}

// Detect searches Git repository for .lc/config.json
func Detect() (*AutoDetectContext, error)

// ApplyTo applies detected values to command flags
func (ctx *AutoDetectContext) ApplyTo(cmd *cobra.Command)
```

## Consequences

### Positive

- Reduced repetition for daily operations
- Configuration is version-controlled with code
- Easy to share team settings
- Clear precedence rules avoid confusion

### Negative

- Requires `.lc/` directory in repository root
- May be confusing if multiple nested Git repositories exist
- Need to handle case where user is not in a Git repository

## References

- `internal/common/autodetect.go`
- `internal/git/git.go`
- `internal/config/config.go`
