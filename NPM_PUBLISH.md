# LC CLI npm 发布系统

本文档说明如何使用 npm 发布系统来分发 LC CLI 工具。

## 架构概述

采用 Monorepo + 平台拆包发布策略：

- **1 个主包** (`@lingji/lc`): Wrapper 包，根据平台自动选择对应的二进制包
- **6 个平台包**: 包含各平台的二进制文件
  - `@lingji/lc-linux-x64`
  - `@lingji/lc-linux-arm64`
  - `@lingji/lc-darwin-x64`
  - `@lingji/lc-darwin-arm64`
  - `@lingji/lc-windows-x64`
  - `@lingji/lc-windows-arm64`

## 目录结构

```
packages/
├── lc/                    # 主包（wrapper）
│   ├── bin/lc.js         # 入口脚本
│   └── package.json
├── lc-linux-x64/         # Linux x64 平台包
│   ├── bin/.gitkeep
│   ├── index.js
│   └── package.json
├── lc-linux-arm64/       # Linux arm64 平台包
├── lc-darwin-x64/        # macOS x64 平台包
├── lc-darwin-arm64/      # macOS arm64 平台包
├── lc-windows-x64/       # Windows x64 平台包
└── lc-windows-arm64/     # Windows arm64 平台包
```

---

## 快速发布（推荐）

使用一键发布脚本：

```bash
./npm_release.sh v0.1.2
```

脚本会自动执行（按顺序）：
1. 检查工作区是否干净
2. 运行 `make check`（先检查，再打 tag，避免 tag 后才发现问题）
3. 创建 Git Tag
4. 同步版本号到 package.json
5. 提交版本更新
6. 推送代码和标签到远程
7. 检查 npm 登录状态
8. 执行 `make npm-release` 发布到 npm

---

## 完整发布流程（按顺序执行）

### 步骤 1: 代码检查与测试（先检查，再打 tag）

**重要**：先执行检查确保代码没有问题，再打 tag。避免代码有问题还要处理 tag。

```bash
# 运行代码检查（格式化 + 静态分析 + 测试）
make check

# 或者只运行测试
make test
```

### 步骤 2: 创建 Git Tag（版本号来源）

检查通过后，创建标签。**版本号以 Git Tag 为准**，不需要手动修改 `package.json`。

```bash
# 创建标签（格式：v + 版本号）
git tag v0.1.1

# 验证标签
git describe --tags
# 输出: v0.1.1
```

### 步骤 3: 推送到远程仓库

```bash
# 推送代码
git push origin master

# 推送标签（必须推送标签，版本号从 tag 读取）
git push origin v0.1.1
```

### 步骤 4: 登录 npm

```bash
# 检查是否已登录
npm whoami

# 如果未登录，执行登录
npm login

# 输入用户名、密码、邮箱、OTP（如有二次验证）
```

**注意**: 必须有 `@lingji` 组织的发布权限。

### 步骤 5: 执行发布

```bash
# 一键发布（推荐）
make npm-release
```

该命令会自动执行：
1. 从 git tag 获取版本号（如 v0.1.1 → 0.1.1）
2. 同步版本号到所有 7 个 package.json
3. 构建 6 个平台的二进制文件
4. 发布 6 个平台包到 npm
5. 发布主包到 npm
6. 清理构建产物

### 步骤 6: 验证发布

```bash
# 检查 npm 上是否已发布
npm view @lingji/lc versions

# 测试安装
npm install -g @lingji/lc

# 测试运行
lc --help
```

---

## 发布流程总结

```bash
# ====== 第 1 步：检查测试（先检查，避免打 tag 后才发现问题） ======
make check

# ====== 第 2 步：创建标签（版本号来源） ======
git tag v0.1.1

# ====== 第 3 步：推送 ======
git push origin master
git push origin v0.1.1

# ====== 第 4 步：登录 npm ======
npm login

# ====== 第 5 步：发布 ======
make npm-release

# ====== 第 6 步：验证 ======
npm install -g @lingji/lc
lc --help
```

---

## 常见问题

### Q: 版本号应该在哪里更新？

**只需要创建 Git Tag**，版本号从 Git Tag 自动读取。

```bash
# 创建标签（版本号会自动从 tag 获取）
git tag v0.1.1
```

`scripts/sync-version.js` 会自动从 git tag 读取版本号（去掉 v 前缀），并更新到所有 package.json。

### Q: 如果发布失败怎么办？

1. **检查 npm 登录状态**: `npm whoami`
2. **检查网络连接**
3. **检查是否有发布权限**
4. **重新执行发布**: `make npm-release`

如果平台包部分发布成功，可以单独发布主包：
```bash
make npm-publish-main
```

### Q: 如何跳过某些步骤？

如果确定代码没问题，可以跳过测试：
```bash
# 快速发布（跳过测试）
git tag v0.1.1
git push origin master v0.1.1
make npm-release
```

### Q: 标签和 package.json 版本不一致会怎样？

- Go 二进制文件会使用 **Git 标签**作为版本号
- npm 包会使用 **package.json** 中的版本号

**必须保持两者一致**，否则会造成版本混乱。

### Q: 如何发布 beta 版本？

```bash
# 创建 beta 标签（版本号包含 -beta.x）
git tag v0.2.0-beta.1

# 发布（npm-release 会正常发布，npm 会识别为预发布版本）
make npm-release

# 用户安装 beta 版本
npm install -g @lingji/lc@0.2.0-beta.1
```

---

## 分步发布（调试时使用）

如果需要调试发布过程，可以分步执行：

```bash
# 1. 同步版本号
node scripts/sync-version.js

# 2. 构建平台包
make build-npm

# 3. 发布平台包（先发布）
make npm-publish-platforms

# 4. 发布主包（后发布）
make npm-publish-main

# 5. 清理
make clean-npm-bin
```

---

## Make 命令参考

| 命令 | 说明 |
|------|------|
| `make build-npm` | 构建所有平台的 npm 包二进制文件 |
| `make npm-publish` | 发布所有 npm 包 |
| `make npm-publish-platforms` | 仅发布 6 个平台包 |
| `make npm-publish-main` | 仅发布主包 |
| `make npm-release` | **完整的发布流程（推荐）** |
| `make clean-npm-bin` | 清理 npm 包中的二进制文件 |
| `make check` | 代码检查（fmt + vet + test） |

---

## 一键发布脚本

### 用法

```bash
./npm_release.sh <版本号>
```

### 示例

```bash
# 发布正式版本
./npm_release.sh v0.1.2

# 发布 beta 版本
./npm_release.sh v0.2.0-beta.1
```

### 脚本执行流程

1. **检查工作区** - 确保没有未提交的更改
2. **创建 Git Tag** - 自动创建指定版本的 tag
3. **同步版本号** - 从 git tag 同步版本到所有 package.json
4. **代码检查** - 运行 `make check`
5. **提交更新** - 自动提交版本更新
6. **推送代码** - 推送代码和标签到远程
7. **检查 npm 登录** - 确保已登录 npm
8. **执行发布** - 运行 `make npm-release`

### 注意事项

- 版本号必须以 `v` 开头，如 `v0.1.2`
- 脚本会在关键步骤前提示确认
- 如果标签已存在，会提示是否重新创建
- 发布前会再次确认是否继续

---

## 用户安装

用户通过 npm 直接安装：

```bash
# 全局安装
npm install -g @lingji/lc

# 或使用 npx
npx @lingji/lc --help
```

npm 会根据用户的系统平台自动安装对应的平台包。

---

## 发布前检查清单

- [ ] 已创建 Git Tag（如 `v0.1.1`）
- [ ] 已运行 `make check` 且通过
- [ ] 已推送代码和标签到远程
- [ ] 已登录 npm（`npm whoami` 显示用户名）
- [ ] 有 `@lingji` 组织的发布权限

---

## 注意事项

1. **版本号管理**: 所有包的版本号必须保持一致，`make npm-release` 会自动同步
2. **二进制文件**: 发布前会自动构建，发布后自动清理，不需要手动管理
3. **平台检测**: 主包通过 `process.platform` 和 `process.arch` 检测平台
4. **scoped package**: 使用 `@lingji` scope，发布时需要 `--access public`
5. **发布顺序**: 必须先发布平台包，再发布主包（主包依赖平台包）
