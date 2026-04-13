// Package e2e 包含 LC CLI 的端到端测试
//
// 目录结构：
//   - framework/: 测试框架基础设施
//   - cli/: CLI 基础测试（无需外部服务）
//   - integration/: 集成测试（需要外部服务）
//
// 运行测试：
//   - 基础测试: go test -v -short ./e2e/cli/...
//   - 集成测试: go test -v ./e2e/integration/...
package e2e
