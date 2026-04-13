# Architecture Decision Records (ADR)

This directory contains Architecture Decision Records for the Lingji CLI project.

## What is an ADR?

An Architecture Decision Record (ADR) captures an important architectural decision made along with its context and consequences. ADRs help team members understand why certain decisions were made and provide historical context for future developers.

## Index

| ADR | Title | Status | Description |
|-----|-------|--------|-------------|
| [001](001-cli-framework.md) | CLI Framework Selection | Accepted | Why we chose Cobra |
| [002](002-error-handling.md) | Error Handling Strategy | Accepted | Unified JSON error format |
| [003](003-auto-detect.md) | Context Auto-Detection | Accepted | Git-based configuration detection |
| [004](004-api-client.md) | API Client Architecture | Accepted | Layered HTTP client design |
| [005](005-testing-strategy.md) | Testing Strategy | Accepted | Multi-layer testing approach |

## Status Meanings

- **Proposed**: Under discussion, not yet decided
- **Accepted**: Decision made and implemented
- **Deprecated**: Decision was valid but is no longer relevant
- **Superseded**: Decision was replaced by a newer ADR

## Creating New ADRs

When creating a new ADR:

1. Use the next available number (e.g., `006-feature-name.md`)
2. Follow the template in existing ADRs
3. Update this README index
4. Link to relevant code files

## Template

```markdown
# ADR XXX: Title

## Status

Proposed / Accepted / Deprecated / Superseded

## Context

What is the issue that we're seeing that is motivating this decision?

## Decision

What is the change that we're proposing or have agreed to implement?

## Rationale

Why was this decision made? What alternatives were considered?

## Consequences

### Positive
- Benefit 1
- Benefit 2

### Negative
- Trade-off 1
- Trade-off 2

## References
- Link to relevant files
- Link to related issues/PRs
```
