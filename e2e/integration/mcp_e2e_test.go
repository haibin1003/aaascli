package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/lc/e2e/framework"
)

// mcpConfig 是测试用的 MCP 配置，指向真实可访问的 openDeepWiki Server。
const mcpConfig = `{
  "mcpServers": {
    "openDeepWiki": {
      "url": "https://opendeepwiki.k8m.site/mcp/sse",
      "timeout": 30000
    }
  }
}`

// setupMCPConfig 在 CLI 的临时 HOME 下创建 MCP 配置文件。
// 路径：<home>/.config/mcp/config.json（搜索优先级 #2）
func setupMCPConfig(t *testing.T, lc *framework.CLI) {
	t.Helper()
	dir := filepath.Join(lc.GetHome(), ".config", "mcp")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("创建 MCP 配置目录失败: %v", err)
	}
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(mcpConfig), 0644); err != nil {
		t.Fatalf("写入 MCP 配置文件失败: %v", err)
	}
}

// getDataMap 从 JSON 响应的 data 字段中提取 map。
func getDataMap(t *testing.T, data map[string]interface{}) map[string]interface{} {
	t.Helper()
	dm, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("响应中 data 字段缺失或类型错误: %v", data)
	}
	return dm
}

// ─────────────────────────────────────────────────────────────────────────────
// 帮助信息测试（无需网络）
// ─────────────────────────────────────────────────────────────────────────────

func TestMCPHelp(t *testing.T) {
	lc := framework.NewCLI(t)

	res := lc.Run("mcp", "--help")
	out := res.Stdout + res.Stderr

	for _, keyword := range []string{"mcp", "MCP", "serverName", "配置"} {
		if !strings.Contains(out, keyword) {
			t.Errorf("帮助信息中缺少关键词 %q\n输出: %s", keyword, out)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 错误处理测试（无需网络）
// ─────────────────────────────────────────────────────────────────────────────

func TestMCPNoConfig(t *testing.T) {
	// 不调用 setupMCPConfig，临时 HOME 下没有任何 MCP 配置文件。
	lc := framework.NewCLI(t)

	res := lc.Run("mcp")
	res.ExpectFailure()

	data := res.ExpectJSON()
	if success, _ := data["success"].(bool); success {
		t.Fatal("没有配置文件时 success 应为 false")
	}

	errObj, ok := data["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("响应中缺少 error 字段: %v", data)
	}
	code, _ := errObj["code"].(string)
	if code != "MCP_CONFIG_NOT_FOUND" {
		t.Errorf("错误码应为 MCP_CONFIG_NOT_FOUND，实际为 %q", code)
	}

	// 错误详情中应包含搜索路径
	details, _ := errObj["details"].(map[string]interface{})
	if details == nil {
		t.Error("error.details 字段应存在")
	} else if _, ok := details["searchPaths"]; !ok {
		t.Error("error.details 中应包含 searchPaths")
	}
}

func TestMCPParamInvalid(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)
	setupMCPConfig(t, lc)

	// "badparam" 没有等号，应触发 MCP_PARAM_INVALID
	res := lc.Run("mcp", "openDeepWiki", "list_repositories", "badparam")
	res.ExpectFailure()

	data := res.ExpectJSON()
	if success, _ := data["success"].(bool); success {
		t.Fatal("参数格式错误时 success 应为 false")
	}

	errObj, ok := data["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("响应中缺少 error 字段: %v", data)
	}
	code, _ := errObj["code"].(string)
	if code != "MCP_PARAM_INVALID" {
		t.Errorf("错误码应为 MCP_PARAM_INVALID，实际为 %q", code)
	}
	msg, _ := errObj["message"].(string)
	if !strings.Contains(msg, "badparam") {
		t.Errorf("错误信息应包含非法参数名 %q，实际为 %q", "badparam", msg)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 列出所有 Server 测试（需要网络 + 真实 Server）
// ─────────────────────────────────────────────────────────────────────────────

func TestMCPList(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)
	setupMCPConfig(t, lc)

	res := lc.Run("mcp")
	res.MustSucceed()

	data := res.ExpectJSON()
	dm := getDataMap(t, data)

	// 验证 configFiles 字段存在且非空
	configFiles, _ := dm["configFiles"].([]interface{})
	if len(configFiles) == 0 {
		t.Error("data.configFiles 字段应为非空数组")
	}
	t.Logf("配置文件路径: %v", configFiles)

	// 验证 servers 列表
	servers, ok := dm["servers"].([]interface{})
	if !ok || len(servers) == 0 {
		t.Fatalf("data.servers 应为非空数组，实际为: %v", dm["servers"])
	}

	// 找到 openDeepWiki
	var found bool
	for _, s := range servers {
		sv, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := sv["name"].(string)
		if name != "openDeepWiki" {
			continue
		}
		found = true

		// 验证 transport 存在
		transport, _ := sv["transport"].(string)
		if transport == "" {
			t.Error("openDeepWiki server 的 transport 字段应非空")
		}

		// 验证 URL 存在
		url, _ := sv["url"].(string)
		if url == "" {
			t.Error("openDeepWiki server 的 url 字段应非空")
		}

		// 注：0 参数模式不再返回 tools 列表，如需验证 tools 请使用 lc mcp <serverName>
		t.Logf("openDeepWiki server: name=%s, transport=%s, url=%s", name, transport, url)
	}

	if !found {
		t.Error("servers 列表中应包含 openDeepWiki")
	}
}

func TestMCPListOutputIsValidJSON(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)
	setupMCPConfig(t, lc)

	res := lc.Run("mcp")

	var raw interface{}
	if err := json.Unmarshal([]byte(res.Stdout), &raw); err != nil {
		t.Fatalf("输出应为合法 JSON，解析失败: %v\n输出: %s", err, res.Stdout)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 指定 Server 测试（需要网络 + 真实 Server）
// ─────────────────────────────────────────────────────────────────────────────

func TestMCPServerInfo(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)
	setupMCPConfig(t, lc)

	// 列出 openDeepWiki 的工具
	res := lc.Run("mcp", "openDeepWiki")
	res.MustSucceed()

	data := res.ExpectJSON()
	dm := getDataMap(t, data)

	// 验证 configFiles 字段存在且非空
	configFiles, _ := dm["configFiles"].([]interface{})
	if len(configFiles) == 0 {
		t.Error("data.configFiles 字段应为非空数组")
	}

	// 验证 server 字段
	server, ok := dm["server"].(map[string]interface{})
	if !ok {
		t.Fatalf("data.server 字段缺失或类型错误: %v", dm)
	}

	name, _ := server["name"].(string)
	if name != "openDeepWiki" {
		t.Errorf("server.name 应为 openDeepWiki，实际为 %q", name)
	}

	url, _ := server["url"].(string)
	if url == "" {
		t.Error("server.url 字段应非空")
	}

	tools, ok := server["tools"].([]interface{})
	if !ok || len(tools) == 0 {
		t.Error("server.tools 列表应非空")
	} else {
		t.Logf("openDeepWiki 工具数量: %d", len(tools))
	}
}

func TestMCPServerNotFound(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)
	setupMCPConfig(t, lc)

	res := lc.Run("mcp", "nonExistentServer_XYZ_12345")
	res.ExpectFailure()

	data := res.ExpectJSON()
	if success, _ := data["success"].(bool); success {
		t.Fatal("不存在的 Server 应返回 success=false")
	}

	errObj, ok := data["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("响应中缺少 error 字段: %v", data)
	}
	code, _ := errObj["code"].(string)
	if code != "MCP_SERVER_NOT_FOUND" {
		t.Errorf("错误码应为 MCP_SERVER_NOT_FOUND，实际为 %q", code)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 工具详情测试（需要网络 + 真实 Server）
// ─────────────────────────────────────────────────────────────────────────────

func TestMCPToolInfo(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)
	setupMCPConfig(t, lc)

	res := lc.Run("mcp", "openDeepWiki", "list_repositories")
	res.MustSucceed()

	data := res.ExpectJSON()
	dm := getDataMap(t, data)

	// 验证 server 字段
	serverName, _ := dm["server"].(string)
	if serverName != "openDeepWiki" {
		t.Errorf("server 字段应为 openDeepWiki，实际为 %q", serverName)
	}

	// 验证 tool 字段
	tool, ok := dm["tool"].(map[string]interface{})
	if !ok {
		t.Fatalf("data.tool 字段缺失或类型错误: %v", dm)
	}

	name, _ := tool["name"].(string)
	if name != "list_repositories" {
		t.Errorf("tool.name 应为 list_repositories，实际为 %q", name)
	}

	desc, _ := tool["description"].(string)
	if desc == "" {
		t.Error("tool.description 应非空")
	} else {
		t.Logf("工具描述: %s", desc)
	}

	// 验证新的国际化字段
	if paramFormat, ok := tool["param_format"].(string); !ok || paramFormat == "" {
		t.Error("tool.param_format 应存在且非空")
	}

	if paramExample, ok := tool["param_example"].([]interface{}); !ok || len(paramExample) == 0 {
		t.Error("tool.param_example 应存在且为非空数组")
	}

	if callExample, ok := tool["call_example"].(string); !ok || callExample == "" {
		t.Error("tool.call_example 应存在且非空")
	}

	// 注：有 param_format 时不会返回 inputSchema（已被转换为人类可读格式）
}

func TestMCPMethodNotFound(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)
	setupMCPConfig(t, lc)

	res := lc.Run("mcp", "openDeepWiki", "nonExistentMethodXYZ_12345")
	res.ExpectFailure()

	data := res.ExpectJSON()
	if success, _ := data["success"].(bool); success {
		t.Fatal("不存在的方法应返回 success=false")
	}

	errObj, ok := data["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("响应中缺少 error 字段: %v", data)
	}
	code, _ := errObj["code"].(string)
	if code != "MCP_METHOD_NOT_FOUND" {
		t.Errorf("错误码应为 MCP_METHOD_NOT_FOUND，实际为 %q", code)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 调用工具测试（需要网络 + 真实 Server）
// ─────────────────────────────────────────────────────────────────────────────

func TestMCPCallTool(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)
	setupMCPConfig(t, lc)

	// list_repositories 无必填参数，传入 limit 以验证参数传递
	res := lc.Run("mcp", "openDeepWiki", "list_repositories", "limit=3")
	res.MustSucceed()

	data := res.ExpectJSON()
	dm := getDataMap(t, data)

	// 验证 server 字段
	serverName, _ := dm["server"].(string)
	if serverName != "openDeepWiki" {
		t.Errorf("server 字段应为 openDeepWiki，实际为 %q", serverName)
	}

	// 验证 method 字段
	method, _ := dm["method"].(string)
	if method != "list_repositories" {
		t.Errorf("method 字段应为 list_repositories，实际为 %q", method)
	}

	// 验证 result 非空
	if dm["result"] == nil {
		t.Error("data.result 不应为 nil")
	}

	// result 是工具返回的字符串（JSON 文本），解析后验证结构
	resultStr, ok := dm["result"].(string)
	if !ok {
		t.Fatalf("data.result 应为字符串，实际类型: %T", dm["result"])
	}

	var resultData map[string]interface{}
	if err := json.Unmarshal([]byte(resultStr), &resultData); err != nil {
		t.Fatalf("result 内容应为合法 JSON，解析失败: %v\nresult: %s", err, resultStr)
	}

	// 验证返回结构包含 repositories 列表
	repos, ok := resultData["repositories"].([]interface{})
	if !ok {
		t.Fatalf("result 中应有 repositories 数组，实际: %v", resultData)
	}
	if len(repos) == 0 {
		t.Error("repositories 列表不应为空")
	}
	if len(repos) > 3 {
		t.Errorf("传入 limit=3，返回 %d 条，超出预期", len(repos))
	}
	t.Logf("返回 %d 个仓库（limit=3）", len(repos))
}

// ─────────────────────────────────────────────────────────────────────────────
// 输出格式测试
// ─────────────────────────────────────────────────────────────────────────────

func TestMCPOutputHasMeta(t *testing.T) {
	framework.SkipIfShort(t)

	lc := framework.NewCLI(t)
	setupMCPConfig(t, lc)

	res := lc.Run("mcp")

	data := res.ExpectJSON()

	meta, ok := data["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("响应应包含 meta 字段")
	}
	if meta["timestamp"] == "" || meta["timestamp"] == nil {
		t.Error("meta.timestamp 应非空")
	}
	if meta["version"] == "" || meta["version"] == nil {
		t.Error("meta.version 应非空")
	}
}

func TestMCPErrorOutputHasMeta(t *testing.T) {
	// 无配置 → 错误响应，也应有 meta 字段
	lc := framework.NewCLI(t)

	res := lc.Run("mcp")

	data := res.ExpectJSON()
	meta, ok := data["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("错误响应也应包含 meta 字段")
	}
	if meta["timestamp"] == nil {
		t.Error("meta.timestamp 应非空")
	}
}
