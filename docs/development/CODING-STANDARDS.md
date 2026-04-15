# 开发规范（Coding Standards）

> 版本：v1.0  
> 生效日期：2026-04-15  
> 语言：Go 1.26+

---

## 一、项目结构规范

```
cmd/sdp/              # Cobra 命令定义（薄层，仅负责参数解析和调用）
internal/
  api/                # HTTP 客户端、平台接口封装、领域模型
  common/             # 命令执行器、结果打印、上下文管理
  config/             # 配置文件读写（~/.sdp/config.json）
  knowledge/          # 内置知识库（embed Markdown）
docs/
  knowledge/          # 知识库源文件 Markdown
  ai-usage-guide.md   # AI 使用手册
  user-prompt-template.md  # Prompt 约束模板
  development/        # 研发规范与路线图
release/
  bin/                # 构建产物（不提交 Git）
  sdp-login-helper/   # Chrome 插件源文件
```

### 原则
- `cmd/sdp/*.go` 中的 `Run` 函数不得超过 30 行，禁止直接写 HTTP 请求。
- 所有业务逻辑必须下沉到 `internal/api/` 或 `internal/common/`。

---

## 二、Go 代码规范

### 2.1 格式与检查
所有提交前必须执行：

```bash
go fmt ./...
go vet ./...
go test ./...
```

推荐安装 `staticcheck`：
```bash
staticcheck ./...
```

### 2.2 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | 全小写，无下划线 | `api`, `common`, `config` |
| 导出符号 | PascalCase | `AbilityService`, `ListMyApplies` |
| 未导出符号 | camelCase | `abilityPage`, `ensureCrypto` |
| 接口名 | 动词+er 或 名词 | `CommandFunc` |
| 测试函数 | `Test` + 被测函数名 | `TestAbilityService_ListAll` |
| Mock 服务 | `mock` + 被测服务名 | `mockAbilityServer` |

### 2.3 错误处理
- 禁止裸 `return err`，必须带上下文：
  ```go
  // 推荐
  return nil, fmt.Errorf("list abilities failed: %w", err)
  
  // 禁止
  return nil, err
  ```
- API 层错误必须包含调用的接口路径信息。

### 2.4 常量与配置抽离
禁止在业务代码中硬编码以下内容，必须放在 `internal/api/consts.go`：
- `BaseURL`
- 成功判断码（如 `"00000"`）
- 通用 HTTP Header（`User-Agent`、`Referer`、`Origin`）
- 超时时间默认值

```go
// internal/api/consts.go
package api

const (
    BaseURL          = "https://service.sd.10086.cn"
    DefaultTimeout   = 60 * time.Second
    SuccessCode      = "00000"
    DefaultUserAgent = "Mozilla/5.0 (...)"
)
```

---

## 三、HTTP 客户端规范

### 3.1 请求封装
所有 HTTP 请求统一通过 `Client` 的方法发送：
- `Get(url string)`
- `Post(url string, body []byte)`
- `PostMultipart(url string, contentType string, body []byte)`

新增请求方法需经过 review，禁止在 service 里直接调用 `http.DefaultClient`。

### 3.2 响应解析
所有 JSON 响应必须通过统一的 `ParseJSON` 处理：

```go
func ParseJSON(data []byte, v interface{}) error {
    if err := json.Unmarshal(data, v); err != nil {
        return fmt.Errorf("unmarshal failed: %w, raw: %s", err, string(data))
    }
    return nil
}
```

### 3.3 测试中的 HTTP 模拟
必须使用 `httptest` 创建本地服务器，禁止测试访问真实平台：

```go
func TestAbilityService_ListAll(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/openportalsrv/...", r.URL.Path)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"code":"00000","data":{...}}`))
    }))
    defer server.Close()

    client := NewClient("fake-cookie", true)
    client.BaseURL = server.URL  // 或重构成可注入 BaseURL
    svc := NewAbilityService(client)
    _, err := svc.ListAll()
    require.NoError(t, err)
}
```

---

## 四、CLI 命令规范

### 4.1 命令分层
```
能力管理      sdp ability <subcommand>
服务管理      sdp service <subcommand>
应用管理      sdp app <subcommand>
订购申请       sdp order <subcommand>
知识库        sdp knowledge <subcommand>
辅助工具      sdp helper <subcommand>
```

### 4.2 Flag 规范
- 分页统一使用 `--page` / `-p` 和 `--size` / `-s`。
- 输出格式统一使用 `--format` / `-f`，可选值：`json`（默认）、`table`、`yaml`、`markdown`。
- 调试用 `--debug` / `-d`。
- 跳过 TLS 验证用 `--insecure`。
- 禁用缓存用 `--no-cache`。

### 4.3 输出规范
- 成功输出统一使用 `CommandResult` 结构：
  ```go
  type CommandResult struct {
      Success bool        `json:"success"`
      Data    interface{} `json:"data,omitempty"`
      Error   interface{} `json:"error,omitempty"`
      Meta    MetaInfo    `json:"meta"`
  }
  ```
- 错误输出必须包含 `suggestion` 字段，提示用户下一步怎么做。

---

## 五、测试规范

### 5.1 测试文件位置
- 与被测文件同包，文件名后缀 `_test.go`。
- 示例：`internal/api/ability.go` → `internal/api/ability_test.go`

### 5.2 测试覆盖率目标
- `internal/api/*`：≥ 60%
- `cmd/sdp/*`：核心路径有集成测试即可
- 整体项目：≥ 40%

### 5.3 测试工具
- 推荐使用 `testify/require` 和 `testify/assert`。
- HTTP Mock 使用标准库 `net/http/httptest`。
- 文件系统 Mock 使用 `testing/fstest` 或临时目录 `t.TempDir()`。

### 5.4 禁止事项
- 禁止测试依赖真实网络或真实平台环境。
- 禁止测试修改用户真实配置文件（`~/.sdp/config.json`）。
- 禁止测试遗留未清理的临时文件。

---

## 六、文档同步规范

### 6.1 代码变更 → 文档检查清单
每完成一个功能，开发者必须勾选以下清单：

- [ ] `docs/ai-usage-guide.md` 已更新（新增/修改命令）
- [ ] `cmd/sdp/onboard.go` 已更新（新增/修改命令）
- [ ] `docs/user-prompt-template.md` 已更新（新增场景或约束）
- [ ] `docs/knowledge/` 已更新（新增业务知识）
- [ ] `README.md` 已更新（安装方式、功能列表变化）
- [ ] `docs/development/ROADMAP.md` / `TODO.md` 已更新

### 6.2 知识库更新
- 新增 Markdown 必须同时放入 `docs/knowledge/` 和 `internal/knowledge/`（或通过构建脚本同步）。
- 知识库标题必须以 `# ` 开头，否则 `extractTitle` 会回退到文件名。

---

## 七、Git 提交规范

### 7.1 Commit Message 格式
```
<type>(<scope>): <subject>

<body>

Test Report:
- go test ./...: PASS (coverage: xx%)
- go vet ./...: PASS
- build: PASS
- manual check: PASS (平台: xxx, 命令: xxx)
```

### 7.2 常用 scope
- `api`：`internal/api/*` 变更
- `cmd`：`cmd/sdp/*` 变更
- `knowledge`：知识库变更
- `docs`：文档变更
- `test`：测试补充
- `build`：Makefile / 构建脚本变更

### 7.3 提交粒度
- 一个 commit 只做一个逻辑变更。
- 禁止"大杂烩" commit（同时改功能、改文档、调格式）。
