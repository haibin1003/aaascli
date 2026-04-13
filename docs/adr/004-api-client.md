# ADR 004: API Client Architecture

## Status

Accepted

## Context

The CLI tool needs to communicate with the Lingji (灵畿) platform's REST API. We need a clean, testable, and maintainable HTTP client architecture.

## Decision

We implement a layered API client architecture with separation of concerns.

## Rationale

### Architecture Layers

```
┌─────────────────────────────────────┐
│         Command Layer               │
│    (cmd/lc/*.go)                    │
├─────────────────────────────────────┤
│         Service Layer               │
│    (internal/service/*.go)          │
│    - Business logic                 │
│    - Data transformation            │
├─────────────────────────────────────┤
│         API Client Layer            │
│    (internal/api/*.go)              │
│    - HTTP communication             │
│    - Request/Response handling      │
│    - Authentication                 │
├─────────────────────────────────────┤
│         Transport Layer             │
│    (Go net/http)                    │
│    - Connection pooling             │
│    - Retry logic                    │
└─────────────────────────────────────┘
```

### Design Principles

1. **Separation of Concerns**: Each layer has a single responsibility
2. **Testability**: Easy to mock for unit testing
3. **Reusability**: Common patterns extracted into utilities
4. **Error Handling**: Consistent error propagation

### Key Components

#### Client (`internal/api/client.go`)

```go
type Client struct {
    httpClient *http.Client
    baseURL    string
}

func (c *Client) Send(req *Request) (*Response, error)
```

Features:
- Connection pooling via shared `http.Transport`
- Timeout configuration
- Request/Response interceptors

#### Request Builder (`internal/api/request.go`)

```go
type Request struct {
    Method  string
    URL     string
    Headers map[string]string
    Body    interface{}
}

func NewRequest(method, url string) *Request
func (r *Request) WithHeader(key, value string) *Request
func (r *Request) WithBody(body interface{}) *Request
```

#### Error Types (`internal/api/errors.go`)

```go
type APIError struct {
    StatusCode int
    Code       string
    Message    string
}

func (e *APIError) Error() string
```

## Consequences

### Positive

- Clean separation between business logic and HTTP details
- Easy to add new API endpoints
- Mock-friendly for testing
- Centralized error handling
- Connection reuse improves performance

### Negative

- More boilerplate code than direct HTTP calls
- Multiple layers to understand for new developers
- Need to maintain consistency across layers

## References

- `internal/api/client.go`
- `internal/api/request.go`
- `internal/api/errors.go`
- `internal/api/mock_test.go`
