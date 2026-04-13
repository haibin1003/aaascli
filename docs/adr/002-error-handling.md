# ADR 002: Error Handling Strategy

## Status

Accepted

## Context

The CLI tool needs to provide errors in a machine-readable format for AI systems to process. The errors should be consistent across all commands.

## Decision

We use unified JSON format for all error outputs.

## Rationale

### Requirements

1. **Machine readable**: AI systems can parse the error
2. **Human readable**: Developers can understand the error
3. **Consistent**: Same format across all commands
4. **Actionable**: Includes suggestions for fixing

### Format

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": "Additional context",
    "suggestion": "How to fix this"
  },
  "meta": {
    "timestamp": "2024-03-15T10:00:00Z",
    "version": "v0.2.3"
  }
}
```

### Implementation

- `common.PrintError()`: Output error as JSON
- `common.HandleAutoDetectError()`: Wrap auto-detect errors
- `AutoDetectError` struct: Structured error type

## Consequences

### Positive
- Consistent error format
- Easy for AI to parse
- Includes helpful suggestions

### Negative
- Cannot use standard Go error handling
- Must remember to use `PrintError` instead of `fmt.Println`

## References
- `internal/common/errors.go`
- `internal/common/executor.go`
