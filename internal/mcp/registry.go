package mcp

// ToolDefinition 定义一个 MCP 工具静态命令的元数据
type ToolDefinition struct {
	Command       string   // 命令名，如 "get-repo-wiki"
	Server        string   // 依赖的 MCP Server 名
	Method        string   // MCP 方法名
	Description   string   // 命令描述
	RequiredArgs  []string // 必需参数列表
	OptionalArgs  []string // 可选参数列表
}

// PredefinedTools 是预定义的 MCP 工具命令表
// 这些命令在代码中写死，所有用户使用统一的命令集合
var PredefinedTools = []ToolDefinition{
	{
		Command:      "read-wiki",
		Server:       "openDeepWiki",
		Method:       "read_document",
		Description:  "读取 Wiki 文档内容",
		RequiredArgs: []string{"doc_id"},
	},
}
