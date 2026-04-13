# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Add centralized flag descriptions (`internal/common/flag_descriptions.go`)
  - Simple map-based approach for easier maintenance
  - Avoid hardcoding descriptions in multiple files
  - Helper functions: `GetFlagDesc()`, `GetFlagDescWithDefault()`
  - `CobraFlagHelper` for easy integration with Cobra commands
- Add cache layer (`internal/cache/`)
  - Generic thread-safe cache with TTL support
  - Cache manager for workspace, user, config, API response caching
  - Auto-cleanup of expired items
- Add Service layer architecture (`internal/service/`)
  - `IQLService` - Fluent IQL query builder
  - `ResponseFormatter` - Unified response formatting
  - `DryRunService` - Dry-run operation simulation
  - `WorkspaceService` - Workspace resolution with caching
  - `RequirementService`, `TaskService`, `BugService` - Domain services
- Add Architecture Decision Records (ADR) documentation
- Add HTTP Mock testing framework for unit tests
- Add CI/CD workflows (GitHub Actions)
- Add unit tests for `internal/api`, `internal/cache`, `internal/config`, `internal/common`, `internal/service`

### Changed
- Simplify auto-detect functions by extracting common field configurations
- Extract common utility functions to `cmd/lc/utils.go`
- Optimize HTTP client with connection pool settings
- Enhance HTTP client with retry mechanism and improved timeout settings
  - Add retry support with exponential backoff and jitter
  - Add configurable retry for specific HTTP status codes (429, 502, 503, 504)
  - Add connection timeout (10s) and response header timeout (10s)
  - Enable HTTP/2 support (`ForceAttemptHTTP2`)
  - Add `ResilientClient` with `SendWithRetry` method

### Fixed
- Fix HTTP client resource leak (remove incorrect `defer logger.Sync()`)
- Unify execution mode: replace all `ExecuteWithResult` with `Execute`
- Fix E2E test failures (PR workflow, library lifecycle)

## [0.2.3] - 2024-03-15

### Added
- Add `HandleAutoDetectWithExit` helper function for unified error handling
- Add `utils.go` with common utility functions

### Changed
- Refactor auto-detect error handling to use JSON output consistently
- Remove direct stderr printing in favor of `common.PrintError()`

### Fixed
- Fix error output format: all errors now use unified JSON format
- Fix `--size` flag in autodetect tests (changed to `-l`)
- Fix Git clone URL extraction in E2E tests
- Fix missing `--group-id` parameter in E2E tests

## [0.2.2] - 2024-03-10

### Added
- Add `tryAutoDetectForXXX` functions for all commands
- Add `PrintAutoDetectError` for structured error output

### Changed
- Update error handling to support JSON format

## [0.2.1] - 2024-03-05

### Added
- Add library (文档库) command support
- Add folder management commands
- Add file upload/delete commands

### Changed
- Refactor command structure for better organization

## [0.2.0] - 2024-02-28

### Added
- Add PR (Merge Request) command support
- Add PR comment and review functionality
- Add `--git-project-id` auto-detection

### Changed
- Improve auto-detect reliability

## [0.1.9] - 2024-02-20

### Added
- Add bug (缺陷) command support
- Add bug status management
- Add task command support

## [0.1.0] - 2024-02-01

### Added
- Initial release
- Add req (需求) command support
- Add repo (仓库) command support
- Add space (研发空间) command support
- Add login and configuration commands
- Add auto-detect functionality for Git repositories

