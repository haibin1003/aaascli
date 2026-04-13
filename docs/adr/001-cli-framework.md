# ADR 001: CLI Framework Selection

## Status

Accepted

## Context

We needed to build a command-line interface (CLI) tool for the Lingji (灵畿) platform. The CLI should support multiple commands, subcommands, flags, and have good documentation and help text generation.

## Decision

We chose [Cobra](https://github.com/spf13/cobra) as our CLI framework.

## Rationale

### Why Cobra?

1. **Standard in Go ecosystem**: Cobra is used by Kubernetes, Hugo, and many other major Go projects
2. **Built-in features**:
   - Automatic help generation
   - Bash/Zsh/Fish shell completion
   - Nested subcommands support
   - Flag parsing (POSIX & GNU style)
   - Integration with Viper for configuration
3. **Good documentation**: Extensive documentation and community support
4. **MIT License**: Compatible with our project

### Alternatives Considered

| Framework | Pros | Cons |
|-----------|------|------|
| urfave/cli | Simpler API | Less feature-rich |
| flag (stdlib) | No dependency | Manual help generation |
| kong | Powerful struct tags | Smaller community |

## Consequences

### Positive
- Fast development with standard patterns
- Easy to add new commands
- Good user experience with auto-completion

### Negative
- Dependency on external library
- Learning curve for team members unfamiliar with Cobra

## References
- [Cobra GitHub](https://github.com/spf13/cobra)
- [Cobra Documentation](https://cobra.dev/)
