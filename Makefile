# SDP - 山东能力平台 CLI 工具 Makefile

# 变量定义
BINARY_NAME := sdp
BUILD_DIR := ./bin
OUTPUT := $(BUILD_DIR)/$(BINARY_NAME)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
NPM_VERSION := $(shell node -p "require('./packages/sdp/package.json').version" 2>/dev/null || echo "0.1.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date +%Y-%m-%d-%H:%M:%S)

# LDFLAGS
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"
NPM_LDFLAGS := -ldflags "-X main.Version=$(NPM_VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Go 命令
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

# 默认目标
.DEFAULT_GOAL := help

.PHONY: all build build-linux build-mac build-windows build-npm clean test fmt vet install uninstall help npm-publish npm-release clean-npm-bin sync-knowledge

## help: 显示帮助信息
help:
	@echo "SDP - 山东能力平台 CLI 工具"
	@echo ""
	@echo "使用方法: make [目标]"
	@echo ""
	@echo "目标列表:"
	@echo "  build         构建当前平台的二进制文件"
	@echo "  build-linux   构建 Linux 版本"
	@echo "  build-mac     构建 macOS 版本"
	@echo "  build-windows 构建 Windows 版本"
	@echo "  build-npm     构建所有 npm 平台包的二进制文件"
	@echo "  clean         清理构建产物"
	@echo "  test          运行测试"
	@echo "  fmt           格式化代码"
	@echo "  vet           静态分析"
	@echo "  npm-publish   发布 npm 包（需要先登录 npm）"
	@echo "  npm-release   完整 npm 发布流程"

## sync-knowledge: 同步知识库 Markdown 到 embed 目录
sync-knowledge:
	@echo "==> 同步知识库..."
	@cp docs/knowledge/*.md internal/knowledge/ 2>/dev/null || true
	@echo "==> 知识库同步完成"

## all: 清理、构建并运行测试
all: clean build test

## build: 构建当前平台的二进制文件（开发推荐）
build: sync-knowledge
	@echo "==> 构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(OUTPUT) .
	@echo "==> 构建完成: $(OUTPUT)"

## build-linux: 构建 Linux 平台二进制文件
build-linux: sync-knowledge
	@echo "==> 构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "==> Linux 构建完成"

## build-mac: 构建 macOS 平台二进制文件
build-mac: sync-knowledge
	@echo "==> 构建 macOS 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "==> macOS 构建完成"

## build-windows: 构建 Windows 平台二进制文件
build-windows: sync-knowledge
	@echo "==> 构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "==> Windows 构建完成"

## build-all: 构建所有平台
build-all: build-linux build-mac build-windows
	@echo "==> 所有平台构建完成"
	@ls -lh $(BUILD_DIR)/

## build-npm: 构建所有 npm 平台包的二进制文件
build-npm:
	@echo "==> 构建 npm 平台包..."
	@echo "==> 版本: $(NPM_VERSION)"
	@mkdir -p packages/sdp-linux-x64/bin
	@mkdir -p packages/sdp-linux-arm64/bin
	@mkdir -p packages/sdp-darwin-x64/bin
	@mkdir -p packages/sdp-darwin-arm64/bin
	@mkdir -p packages/sdp-windows-x64/bin
	@mkdir -p packages/sdp-windows-arm64/bin
	@echo "  -> Linux x64..."
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/sdp-linux-x64/bin/sdp .
	@echo "  -> Linux arm64..."
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/sdp-linux-arm64/bin/sdp .
	@echo "  -> macOS x64..."
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/sdp-darwin-x64/bin/sdp .
	@echo "  -> macOS arm64..."
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/sdp-darwin-arm64/bin/sdp .
	@echo "  -> Windows x64..."
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/sdp-windows-x64/bin/sdp.exe .
	@echo "  -> Windows arm64..."
	@GOOS=windows GOARCH=arm64 $(GOBUILD) $(NPM_LDFLAGS) -o packages/sdp-windows-arm64/bin/sdp.exe .
	@echo "==> npm 平台包构建完成"

## npm-publish: 发布 npm 包（需要先登录 npm）
npm-publish: npm-publish-platforms npm-publish-main

## npm-publish-platforms: 发布平台包
npm-publish-platforms:
	@echo "==> 发布平台包..."
	@cd packages/sdp-linux-x64 && npm publish --access public
	@cd packages/sdp-linux-arm64 && npm publish --access public
	@cd packages/sdp-darwin-x64 && npm publish --access public
	@cd packages/sdp-darwin-arm64 && npm publish --access public
	@cd packages/sdp-windows-x64 && npm publish --access public
	@cd packages/sdp-windows-arm64 && npm publish --access public
	@echo "==> 平台包发布完成"

## npm-publish-main: 发布主包
npm-publish-main:
	@echo "==> 发布主包..."
	@cd packages/sdp && npm publish --access public
	@echo "==> 主包发布完成"

## npm-release: 完整 npm 发布流程
npm-release: clean-npm-bin
	@echo "==> 开始 npm 发布流程..."
	@echo "==> 1. 构建二进制文件..."
	@make build-npm
	@echo ""
	@echo "==> 2. 发布平台包..."
	@make npm-publish-platforms
	@echo ""
	@echo "==> 3. 发布主包..."
	@make npm-publish-main
	@echo ""
	@echo "==> 4. 清理二进制文件..."
	@make clean-npm-bin
	@echo ""
	@echo "==> npm 发布流程完成!"

## clean: 清理构建产物和临时文件
clean:
	@echo "==> 清理构建产物..."
	@rm -rf $(BUILD_DIR)
	@$(GOCLEAN)
	@echo "==> 清理完成"

## clean-npm-bin: 清理 npm 包中的二进制文件
clean-npm-bin:
	@echo "==> 清理 npm 包二进制文件..."
	@rm -f packages/sdp-linux-x64/bin/sdp
	@rm -f packages/sdp-linux-arm64/bin/sdp
	@rm -f packages/sdp-darwin-x64/bin/sdp
	@rm -f packages/sdp-darwin-arm64/bin/sdp
	@rm -f packages/sdp-windows-x64/bin/sdp.exe
	@rm -f packages/sdp-windows-arm64/bin/sdp.exe
	@echo "==> 清理完成"

## test: 运行单元测试
test:
	@echo "==> 运行测试..."
	$(GOTEST) -v ./...

## fmt: 格式化 Go 代码
fmt:
	@echo "==> 格式化代码..."
	@gofmt -w -s ./
	@echo "==> 格式化完成"

## vet: 运行 go vet 静态分析
vet:
	@echo "==> 运行静态分析..."
	$(GOCMD) vet ./...

## install: 安装到 /usr/local/bin（Linux/Mac）或当前目录（Windows）
install: build
ifeq ($(OS),Windows_NT)
	@echo "==> Windows 系统，请手动将 $(OUTPUT).exe 添加到 PATH"
else
	@echo "==> 安装到 /usr/local/bin..."
	@install -d /usr/local/bin && cp $(OUTPUT) /usr/local/bin/$(BINARY_NAME)
	@echo "==> 安装完成: /usr/local/bin/$(BINARY_NAME)"
endif

## uninstall: 卸载
uninstall:
ifeq ($(OS),Windows_NT)
	@echo "==> Windows 系统，请手动删除"
else
	@echo "==> 卸载..."
	@rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "==> 卸载完成"
endif

## run: 构建并运行（带参数，如 make run ARGS="ability list"）
run: build
	@echo "==> 运行: $(OUTPUT) $(ARGS)"
	@$(OUTPUT) $(ARGS)

## deps: 下载依赖
deps:
	@echo "==> 下载依赖..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "==> 依赖整理完成"
