---
name: lc-artifact
description: |
  管理制品库。当用户提到 artifact、制品库、制品、maven、npm、pypi、docker、仓库、发布、或需要操作制品仓库时触发。支持多种仓库类型。
metadata:
  {
    "joinai-code":
      {
        "requires": { "bins": ["lc"] },
      },
  }
---

# 制品库管理 (Artifact)

## 支持的仓库类型

| 类型 | 说明 |
|------|------|
| Maven | Java 依赖库 |
| Npm | Node.js 包 |
| Pypi | Python 包 |
| Docker | 容器镜像 |
| Go | Go 模块 |
| Debian | Debian 包 |
| Composer | PHP 包 |
| Rpm | RedHat 包 |
| Conan | C/C++ 包 |
| Nuget | .NET 包 |
| Generic | 通用文件 |
| Cocoapods | iOS/macOS 包 |
| Helm | Kubernetes Helm |
| Cargo | Rust 包 |

## 查询仓库列表
```bash
lc artifact list -w <workspace-key>
lc artifact list -w XXJSLJCLIDEV --pretty
```

## 查询仓库组
```bash
lc artifact group list -w <workspace-key>
```

## 创建仓库

**Maven 仓库**：
```bash
lc artifact create my-maven-repo -w XXJSLJCLIDEV -t Maven -e DEV -g com.example
```

**Npm 仓库**：
```bash
lc artifact create my-npm-repo -w XXJSLJCLIDEV -t Npm -e DEV
```

**Docker 仓库**：
```bash
lc artifact create my-docker-repo -w XXJSLJCLIDEV -t Docker -e DEV
```

## 参数说明

| 参数 | 说明 |
|------|------|
| `-t, --type` | 仓库类型 |
| `-e, --env` | 环境（DEV、TEST、PROD） |
| `-g, --group` | Group ID（Maven 等需要） |

## 工作流程

```bash
# 1. 查看现有仓库
lc artifact list -w XXJSLJCLIDEV

# 2. 查看仓库组
lc artifact group list -w XXJSLJCLIDEV

# 3. 创建仓库
lc artifact create my-app -w XXJSLJCLIDEV -t Maven -e DEV -g com.mycompany

# 4. 获取仓库配置信息（用于配置本地构建工具）
```

## 注意事项

- 仓库名称在同一类型下需唯一
- Group ID 格式通常为反域名（如 com.example）
- 环境选项：DEV、TEST、PROD
