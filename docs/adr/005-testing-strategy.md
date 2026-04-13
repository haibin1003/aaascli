# ADR 005: Testing Strategy

## Status

Accepted

## Context

We need a comprehensive testing strategy for the CLI tool that balances speed, reliability, and coverage.

## Decision

We adopt a multi-layer testing approach with unit tests, integration tests with mocks, and end-to-end tests.

## Rationale

### Testing Pyramid

```
         /\
        /  \\     E2E Tests (Few, slow, real API)
       /____\\
      /      \\
     /   10%   \\   Integration Tests (Mock API)
    /____________\\
   /              \\
  /      70%       \\ Unit Tests (Fast, isolated)
 /__________________\\
```

### Test Types

#### 1. Unit Tests (`*_test.go` alongside source)

- **Scope**: Individual functions and methods
- **Speed**: Fast (< 100ms)
- **Dependencies**: All external dependencies mocked
- **Coverage Target**: 70%+

Example:
```go
func TestParseRequirementResponse(t *testing.T) {
    input := `{"success":true,"data":{"items":[]}}`
    result, err := ParseRequirementResponse([]byte(input))
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

#### 2. Integration Tests with Mocks (`internal/api/*_test.go`)

- **Scope**: API client interactions
- **Speed**: Medium (< 1s)
- **Dependencies**: HTTP server mocked via `httptest`

Example:
```go
func TestClientGetRequirements(t *testing.T) {
    mock := NewMockServer()
    defer mock.Close()

    mock.RegisterJSONHandler("GET", "/api/requirements", 200, mockResponse)

    client := NewClient()
    resp, err := client.GetRequirements(mock.URL())
    assert.NoError(t, err)
}
```

#### 3. End-to-End Tests (`e2e/integration/*_test.go`)

- **Scope**: Full command execution
- **Speed**: Slow (seconds to minutes)
- **Dependencies**: Real API, real authentication
- **Frequency**: Run before releases, not on every commit

Example:
```go
func TestRequirementLifecycle(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test")
    }

    // Create, read, update, delete requirement
    // Using actual CLI commands
}
```

### Test Utilities

#### Mock Server (`internal/api/mock_test.go`)

```go
type MockServer struct {
    Server   *httptest.Server
    Handlers map[string]http.HandlerFunc
}

func (m *MockServer) RegisterJSONHandler(method, path string, status int, response interface{})
```

#### Test Helpers (`internal/common/testutil/`)

- `NewTempConfig()` - Create temporary config files
- `CaptureOutput()` - Capture stdout/stderr
- `AssertJSONEqual()` - Compare JSON structures

### CI/CD Integration

```yaml
# .github/workflows/ci.yml
jobs:
  test:
    steps:
      - run: go test -short ./...        # Unit + Integration
      - run: go test ./e2e/...           # E2E (on main branch only)
```

## Consequences

### Positive

- Fast feedback loop for developers
- Reliable tests that don't depend on external services
- Comprehensive coverage from unit to E2E
- Easy to reproduce and debug failures

### Negative

- More test code to maintain
- Mocks need to be kept in sync with real API
- E2E tests require valid credentials and can be flaky

## References

- `internal/api/mock_test.go`
- `e2e/integration/*_test.go`
- `.github/workflows/ci.yml`
