# LC - 研发管理命令行工具 Makefile
# 统一构建入口，规范开发流程

# 变量定义
BINARY_NAME := lc
BUILD_DIR := ./bin
OUTPUT := $(BUILD_DIR)/$(BINARY_NAME)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
NPM_VERSION := $(shell node -p "require('./packages/lc/package.json').version" 2>/dev/null || echo "0.1.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date +%Y-%m-%d-%H:%M:%S)
GO_VERSION := $(shell go version | cut -d ' ' -f 3)

# 嵌入文件路径
EMBED_DIR := ./internal/embed
HELPER_EXTENSION_DIR := ./lc-login-helper-extension
EMBED_ZIP := $(EMBED_DIR)/helper-extension.zip
SKILLS_DIR := ./skills
SKILLS_ZIP := $(EMBED_DIR)/skills.zip

# LDFLAGS
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"
NPM_LDFLAGS := -ldflags "-X main.Version=$(NPM_VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Go 命令
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# 默认目标
.DEFAULT_GOAL := help

.PHONY: all build build-linux build-mac build-windows build-npm clean test test-short test-e2e coverage fmt vet lint install uninstall dev help npm-publish npm-release build-helper build-skills helper-status build-otp-gen install-otp-gen uninstall-otp-gen test-otp

## help: 显示帮助信息
help:
	@echo "LC - 研发管理命令行工具"
	@echo ""
	@echo "使用方法: make [目标]"
	@echo ""
	@echo "目标列表:"
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## /  /'

## all: 清理、构建并运行测试
all: clean build test

## build: 构建当前平台的二进制文件（开发推荐）
build: build-helper build-skills
	@echo "==> 构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(OUTPUT) .
	@echo "==> 构建完成: $(OUTPUT)"
	@echo "==> 文件大小: $$(ls -lh $(OUTPUT) | awk '{print $$5}')"

## build-linux: 构建 Linux 平台二进制文件
build-linux:
	@echo "==> 构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "==> Linux 构建完成"

## build-mac: 构建 macOS 平台二进制文件
build-mac:
	@echo "==> 构建 macOS 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "==> macOS 构建完成"

## build-windows: 构建 Windows 平台二进制文件
build-windows:
	@echo "==> 构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "==> Windows 构建完成"

## build-all: 构建所有平台
build-all: build-linux build-mac build-windows
	@echo "==> 所有平台构建完成"
	@ls -lh $(BUILD_DIR)/

## clean: 清理构建产物和临时文件
clean:
	@echo "==> 清理构建产物..."
	@rm -rf $(BUILD_DIR)
	@$(GOCLEAN)
	@echo "==> 清理完成"

## test: 运行单元测试
test:
	@echo "==> 运行测试..."
	$(GOTEST) -v ./...

## test-short: 运行短测试（跳过耗时测试）
test-short:
	@echo "==> 运行短测试..."
	$(GOTEST) -short -v ./...

## test-e2e: 运行端到端测试（需要配置）
test-e2e:
	@echo "==> 运行端到端测试..."
	$(GOTEST) -v ./e2e/integration/...

## coverage: 运行测试并生成覆盖率报告
coverage:
	@echo "==> 生成测试覆盖率报告..."
	@mkdir -p $(BUILD_DIR)
	$(GOTEST) -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "==> 覆盖率报告: $(BUILD_DIR)/coverage.html"

## fmt: 格式化 Go 代码
fmt:
	@echo "==> 格式化代码..."
	@gofmt -w -s ./
	@echo "==> 格式化完成"

## vet: 运行 go vet 静态分析
vet:
	@echo "==> 运行静态分析..."
	$(GOCMD) vet ./...

## lint: 运行 golangci-lint（如已安装）
lint:
	@echo "==> 运行 lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过"; \
	fi

## check: 运行所有代码检查（fmt + vet + lint + test）
check: fmt vet test
	@echo "==> 所有检查通过"

## install: 安装到 /usr/local/bin
install: build
	@echo "==> 安装到 /usr/local/bin..."
	@install -d /usr/local/bin && cp $(OUTPUT) /usr/local/bin/$(BINARY_NAME)
	@echo "==> 安装完成: /usr/local/bin/$(BINARY_NAME)"

## uninstall: 从 /usr/local/bin 卸载
uninstall:
	@echo "==> 卸载..."
	@rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "==> 卸载完成: /usr/local/bin/$(BINARY_NAME)"

## run: 构建并运行（带参数，如: make run ARGS="req list -k"）
run: build
	@echo "==> 运行: $(OUTPUT) $(ARGS)"
	@$(OUTPUT) $(ARGS)


## deps: 下载并整理依赖
deps:
	@echo "==> 下载依赖..."
	$(GOMOD) download
	$(GOMOD) tidy
	$(GOMOD) verify
	@echo "==> 依赖整理完成"

## update-deps: 更新所有依赖到最新版本
update-deps:
	@echo "==> 更新依赖..."
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "==> 依赖更新完成"

## version: 显示构建版本信息
version:
	@echo "版本: $(VERSION)"
	@echo "提交: $(COMMIT)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Go 版本: $(GO_VERSION)"

## build-npm: 构建所有 npm 平台包的二进制文件
build-npm:
	@echo "==> 构建 npm 平台包..."
	@echo "==> 版本: $(NPM_VERSION)"
	@mkdir -p packages/lc-linux-x64/bin
	@mkdir -p packages/lc-linux-arm64/bin
	@mkdir -p packages/lc-darwin-x64/bin
	@mkdir -p packages/lc-darwin-arm64/bin
	@mkdir -p packages/lc-windows-x64/bin
	@mkdir -p packages/lc-windows-arm64/bin
	@echo "  -> Linux x64..."
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/lc-linux-x64/bin/lc .
	@echo "  -> Linux arm64..."
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/lc-linux-arm64/bin/lc .
	@echo "  -> macOS x64..."
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/lc-darwin-x64/bin/lc .
	@echo "  -> macOS arm64..."
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/lc-darwin-arm64/bin/lc .
	@echo "  -> Windows x64..."
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/lc-windows-x64/bin/lc.exe .
	@echo "  -> Windows arm64..."
	@GOOS=windows GOARCH=arm64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/lc-windows-arm64/bin/lc.exe .
	@echo "==> npm 平台包构建完成"

## npm-publish: 发布 npm 包（需要先登录 npm）
npm-publish: npm-publish-platforms npm-publish-main

## npm-publish-platforms: 发布平台包
npm-publish-platforms:
	@echo "==> 发布平台包..."
	@cd packages/lc-linux-x64 && npm publish --access public
	@cd packages/lc-linux-arm64 && npm publish --access public
	@cd packages/lc-darwin-x64 && npm publish --access public
	@cd packages/lc-darwin-arm64 && npm publish --access public
	@cd packages/lc-windows-x64 && npm publish --access public
	@cd packages/lc-windows-arm64 && npm publish --access public
	@echo "==> 平台包发布完成"

## npm-publish-main: 发布主包
npm-publish-main:
	@echo "==> 发布主包..."
	@cd packages/lc && npm publish --access public
	@echo "==> 主包发布完成"

## npm-release: 完整的 npm 发布流程
npm-release: clean-npm-bin
	@echo "==> 开始 npm 发布流程..."
	@echo "==> 1. 同步版本号..."
	@node scripts/sync-version.js
	@echo ""
	@echo "==> 2. 构建二进制文件..."
	@make build-npm
	@echo ""
	@echo "==> 3. 发布平台包..."
	@make npm-publish-platforms
	@echo ""
	@echo "==> 4. 发布主包..."
	@make npm-publish-main
	@echo ""
	@echo "==> 5. 清理二进制文件..."
	@make clean-npm-bin
	@echo ""
	@echo "==> npm 发布流程完成!"

## clean-npm-bin: 清理 npm 包中的二进制文件
clean-npm-bin:
	@echo "==> 清理 npm 包二进制文件..."
	@rm -f packages/lc-linux-x64/bin/lc
	@rm -f packages/lc-linux-arm64/bin/lc
	@rm -f packages/lc-darwin-x64/bin/lc
	@rm -f packages/lc-darwin-arm64/bin/lc
	@rm -f packages/lc-windows-x64/bin/lc.exe
	@rm -f packages/lc-windows-arm64/bin/lc.exe
	@echo "==> 清理完成"

## watch: 使用 air 或 fresh 进行热重载开发（如已安装）
watch:
	@if command -v air >/dev/null 2>&1; then \
		air; \
	elif command -v fresh >/dev/null 2>&1; then \
		fresh; \
	else \
		echo "热重载工具未安装，推荐使用:"; \
		echo "  go install github.com/cosmtrek/air@latest"; \
	fi

## build-helper: 打包浏览器扩展到嵌入目录
build-helper:
	@echo "==> 打包浏览器扩展..."
	@mkdir -p $(EMBED_DIR)
	@rm -f $(EMBED_ZIP)
	@cd $(HELPER_EXTENSION_DIR) && zip -q -r ../$(EMBED_ZIP) . -x "*.git*" -x "node_modules/*" -x "*.md" -x "*.log"
	@echo "==> 扩展打包完成: $(EMBED_ZIP)"
	@echo "==> 文件大小: $$(ls -lh $(EMBED_ZIP) | awk '{print $$5}')"

## helper-status: 检查扩展打包状态
helper-status:
	@if [ -f $(EMBED_ZIP) ]; then \
		size=$$(stat -f%z "$(EMBED_ZIP)" 2>/dev/null || stat -c%s "$(EMBED_ZIP)" 2>/dev/null || echo "0"); \
		echo "扩展状态: 已打包"; \
		size_kb=$$((size / 1024)); \
		echo "文件大小: $$size bytes ($${size_kb} KB)"; \
	else \
		echo "扩展状态: 文件不存在"; \
		echo "提示: 运行 'make build' 或 'make build-helper' 打包扩展"; \
	fi

## build-skills: 打包技能到嵌入目录
build-skills:
	@echo "==> 打包技能..."
	@mkdir -p $(EMBED_DIR)
	@rm -f $(SKILLS_ZIP)
	@cd $(SKILLS_DIR) && zip -q -r ../$(SKILLS_ZIP) . -x "*.git*" -x "skill-creator/*"
	@echo "==> 技能打包完成: $(SKILLS_ZIP)"
	@echo "==> 文件大小: $$(ls -lh $(SKILLS_ZIP) | awk '{print $$5}')"

# ============================================================
# OTP 生成器工具 (lc-otp-gen)
# ============================================================

## build-otp-gen: 构建 OTP 生成器工具
build-otp-gen:
	@echo "==> 构建 lc-otp-gen..."
	@mkdir -p $(BUILD_DIR)
	@cd cmd/lc-otp-gen && go mod tidy && $(GOBUILD) -o ../../$(BUILD_DIR)/lc-otp-gen .
	@echo "==> 构建完成: $(BUILD_DIR)/lc-otp-gen"
	@echo "==> 文件大小: $$(ls -lh $(BUILD_DIR)/lc-otp-gen | awk '{print $$5}')"

## install-otp-gen: 安装 OTP 生成器到 /usr/local/bin
install-otp-gen: build-otp-gen
	@echo "==> 安装 lc-otp-gen 到 /usr/local/bin..."
	@install -d /usr/local/bin && cp $(BUILD_DIR)/lc-otp-gen /usr/local/bin/lc-otp-gen
	@echo "==> 安装完成: /usr/local/bin/lc-otp-gen"

## uninstall-otp-gen: 从 /usr/local/bin 卸载 OTP 生成器
uninstall-otp-gen:
	@echo "==> 卸载 lc-otp-gen..."
	@rm -f /usr/local/bin/lc-otp-gen
	@echo "==> 卸载完成: /usr/local/bin/lc-otp-gen"

## test-otp: 运行 OTP 功能自动化测试（需要 lc-otp-gen 已安装）
test-otp: install-otp-gen
	@echo "==> 运行 OTP 自动化测试..."
	@echo "==> 1. 生成验证码..."
	@CODE=$$(lc-otp-gen code 2>/dev/null | grep "验证码:" | head -1 | sed 's/.*验证码: //'); \
	echo "    验证码: $$CODE"; \
	if [ $${#CODE} -eq 6 ]; then \
		echo "==> 2. 验证 OTP..."; \
		$(OUTPUT) otp verify $$CODE; \
	else \
		echo "❌ 获取验证码失败"; \
		exit 1; \
	fi
	@echo "==> OTP 测试完成"

## setup-otp: 快速设置 OTP 测试环境（添加示例配置）
setup-otp: install-otp-gen
	@echo "==> 设置 OTP 测试环境..."
	@lc-otp-gen add test@example.com JXDW7TXQ2ZX2C676YWE7MK7EF5J5ZH2N 2>/dev/null || true
	@echo "==> 已添加测试配置: test@example.com"
	@echo "==> 查看配置: lc-otp-gen list"
