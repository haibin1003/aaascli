#!/usr/bin/env bash

# LC CLI npm 发布脚本
# 用法: ./npm_release.sh v0.1.2

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的信息
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查参数
if [ $# -ne 1 ]; then
    error "请指定版本号"
    echo "用法: $0 v0.1.2"
    echo "示例:"
    echo "  $0 v0.1.2        # 发布正式版本"
    echo "  $0 v0.2.0-beta.1 # 发布 beta 版本"
    exit 1
fi

VERSION_TAG=$1

# 验证版本号格式（必须以 v 开头）
if [[ ! $VERSION_TAG =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    error "版本号格式错误: $VERSION_TAG"
    error "版本号必须以 'v' 开头，如 v0.1.2, v0.2.0-beta.1"
    exit 1
fi

# 提取版本号（去掉 v 前缀）
VERSION=${VERSION_TAG#v}

info "========================================"
info "开始发布 LC CLI"
info "版本号: $VERSION_TAG (npm: $VERSION)"
info "========================================"
echo ""

# 步骤 0: 检查当前目录
cd "$(dirname "$0")"
info "工作目录: $(pwd)"

# 检查必要的文件是否存在
if [ ! -f "Makefile" ] || [ ! -f "packages/lc/package.json" ]; then
    error "当前目录不是项目根目录，或者缺少必要文件"
    exit 1
fi

# 步骤 1: 检查工作区是否干净
info "步骤 1/8: 检查工作区状态..."
if [ -n "$(git status --porcelain)" ]; then
    error "工作区有未提交的更改，请先提交或清理"
    git status --short
    exit 1
fi
success "工作区干净"
echo ""



# 步骤 3: 创建 Git Tag（检查通过后再打 tag）
info "步骤 3/8: 创建 Git Tag: $VERSION_TAG..."
if git rev-parse "$VERSION_TAG" >/dev/null 2>&1; then
    warning "标签 $VERSION_TAG 已存在，删除并重新创建"
    git tag -d "$VERSION_TAG"
    git tag "$VERSION_TAG"
    success "标签已重新创建"
else
    git tag "$VERSION_TAG"
    success "标签 $VERSION_TAG 创建成功"
fi
echo ""

# 步骤 4: 同步版本号
info "步骤 4/8: 同步版本号到 package.json..."
node scripts/sync-version.js
if [ $? -ne 0 ]; then
    error "版本同步失败"
    exit 1
fi
success "版本同步完成"
echo ""

# 步骤 5: 提交版本更新（如果 sync-version.js 修改了文件）
info "步骤 5/8: 检查并提交版本更新..."
if [ -n "$(git status --porcelain packages/)" ]; then
    info "检测到 package.json 有更新，自动提交..."
    git add packages/
    git commit -m "chore: bump version to $VERSION"
    success "版本更新已提交"
else
    info "package.json 无需更新"
fi
echo ""

# 步骤 6: 推送到远程仓库
info "步骤 6/8: 推送到远程仓库..."
info "推送代码和标签到 origin..."
git push origin HEAD
git push origin "$VERSION_TAG"
success "推送完成"
echo ""

# 步骤 7: 检查 npm 登录状态
info "步骤 7/8: 检查 npm 登录状态..."
if ! npm whoami >/dev/null 2>&1; then
    error "未登录 npm"
    info "请执行: npm login"
    npm login
fi
NPM_USER=$(npm whoami)
success "已登录 npm: $NPM_USER"
echo ""

# 步骤 8: 执行发布
info "步骤 8/8: 执行 npm 发布..."
info "执行: make npm-release"
echo ""

if make npm-release; then
    echo ""
    success "========================================"
    success "发布成功!"
    success "版本: $VERSION_TAG"
    success "========================================"
    echo ""
    info "验证发布:"
    echo "  npm view @lingji/lc versions | tail -5"
    echo ""
    info "安装测试:"
    echo "  npm install -g @lingji/lc"
    echo "  lc --help"
else
    error "========================================"
    error "发布失败!"
    error "========================================"
    echo ""
    info "常见解决方法:"
    echo "  1. 检查网络连接"
    echo "  2. 检查 npm 登录状态: npm whoami"
    echo "  3. 检查是否有发布权限"
    echo "  4. 如果部分包已发布，可以单独发布主包: make npm-publish-main"
    exit 1
fi
