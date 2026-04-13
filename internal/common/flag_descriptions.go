// Package common provides centralized flag descriptions for CLI help text.
// This is a simple map-based approach for easier maintenance.
package common

// FlagDescriptions contains all CLI flag descriptions in one place.
// Key: flag name (e.g., "workspace-key", "git-project-id")
// Value: description string
var FlagDescriptions = map[string]string{
	// Global flags
	"workspace-key":    "研发空间唯一标识，用于指定操作的工作空间（支持自动探测）",
	"workspace":        "研发空间唯一标识（workspace-key 的简写）",
	"project-code":     "项目代码，用于指定具体的项目",
	"project":          "项目代码（project-code 的简写）",
	"git-project-id":   "Git 项目 ID，用于 PR/MR 相关操作（支持自动探测）",
	"output":           "输出格式",
	"debug":            "启用调试模式，输出详细的请求/响应信息",
	"insecure":         "跳过 TLS 证书验证（仅用于测试环境）",
	"dry-run":          "试运行模式，显示将要执行的操作但不实际执行",
	"pretty":           "输出格式化的 JSON（默认输出紧凑 JSON）",
	"w":                "研发空间 key（workspace-key 的短形式）",
	"o":                "输出格式（output 的短形式）",
	"d":                "启用调试模式（debug 的短形式）",
	"cookie":           "直接传入认证 Cookie，覆盖本地配置（支持多人并发使用）",
	"c":                "传入认证 Cookie（cookie 的短形式）",

	// Common command flags
	"name":        "名称（必需）",
	"description": "描述",
	"type":        "类型",
	"assignee":    "负责人",
	"file":        "从 YAML 文件创建",
	"filename":    "YAML 文件路径（使用 '-' 从标准输入读取）",
	"keyword":     "搜索关键词",
	"status":      "按状态筛选",
	"limit":       "返回记录数限制",
	"offset":      "分页偏移量",
	"priority":    "优先级：高、中、低",
	"page":        "页码",
	"state":       "状态",
	"body":        "内容",
	"comments":    "显示评论列表",

	// Workspace flags
	"workspace-name": "研发空间名称（可选，支持自动探测）",
	"space":          "研发空间唯一标识（space 的简写）",
	"space-name":     "研发空间名称（space-name 的简写）",

	// Time related
	"planned-start-time": "计划开始时间（时间戳，毫秒）",
	"planned-end-time":   "计划完成时间（时间戳，毫秒）",

	// Requirement specific
	"acceptance-criteria": "验收标准",
	"business-background": "业务背景",
	"requirement":         "需求描述",
	"requirement-id":      "所属需求 ID（必需）",

	// Bug specific
	"bug-id": "缺陷 ID（必需）",

	// PR specific
	"title":          "MR 标题（必需）",
	"source":         "源分支",
	"target":         "目标分支",
	"remove-source":  "合并后删除源分支",
	"source-branch":  "源分支（支持自动探测，Git仓库目录下自动获取当前分支）",
	"target-branch":  "目标分支（默认: master）",

	// Repo specific
	"group-id":     "仓库组 ID",
	"parent-id":    "父组ID（创建子组时使用）",
	"visibility":   "可见性：private, internal, public",
	"init-readme":  "初始化 README 文件",
	"path":         "仓库组路径（默认与名称相同）",

	// MR/PR specific
	"merge-type":   "合并类型: merge、squash 或 rebase",
	"squash":       "合并时压缩提交",
	"comment-id":   "评论 ID（必填），对应 'lc pr view --comments' 输出中 'comments' 数组的 'id' 字段",
	"comment-state": "评论状态（必填）: active(活动中), fixed(已解决), wontFix(无法修复), closed(已关闭), pending(正在挂起)",

	// Task specific
	"task-id": "任务 ID（必需）",

	// Review specific
	"review-type":     "审核类型: approve（批准）或 reject（拒绝）",

	// Bug specific (additional)
	"project-id":      "项目 ID（必填，获取方式: lc space project linked -w <spaceCode>）",
	"handler-id":      "处理人 ID",
	"level":           "缺陷级别: 1(致命), 2(严重), 3(一般), 4(轻微)",
	"defect-type":     "缺陷类型",
	"template-id":     "模板 ID（自定义模板时使用）",
	"template-simple": "使用简洁模板（不关联版本、迭代等，默认使用完整模板）",
	"scene-id":        "场景 ID",
	"bug-status":      "缺陷状态ID列表，逗号分隔（可选）",

	// Repo/Library specific
	"prt-id":          "父文件夹 ID（必填，根目录使用 externalLibId，子目录使用 folderId）",
	"folder-id":       "目标文件夹 ID（必填）",
	"size":            "每页数量",
	"page-size":       "每页数量",

	// Detect specific
	"detect-path":     "指定探测路径（默认为当前目录）",

	// Skills/Hub specific
	"skill-type":      "插件类型",

	// PR specific (additional)
	"mr-state":        "MR 状态过滤: opened, merged, closed, all",

	// Upload specific
	"upload-name":     "上传后的文件名（可选，默认使用原文件名）",

	// CI specific
	"ci-status":       "构建状态过滤: 2=执行中, 3=成功, 4=失败, 5=已停止",

	// Artifact specific
	"artifact-type":   "制品库类型: Maven, Npm, Pypi, Docker, Debian, Composer, Rpm, Go, Conan, Nuget, Generic, Cocoapods, Helm, Cargo",
	"artifact-env":    "制品库环境: DEV(开发), TEST(测试), PROD(生产)",
	"artifact-group-id": "仓库组ID，可通过 'lc artifact group list' 获取",
}

// GetFlagDesc returns the description for a flag.
// Returns empty string if flag is not found.
func GetFlagDesc(name string) string {
	if desc, ok := FlagDescriptions[name]; ok {
		return desc
	}
	return ""
}

// GetFlagDescWithDefault returns the description for a flag,
// or defaultValue if not found.
func GetFlagDescWithDefault(name, defaultValue string) string {
	if desc, ok := FlagDescriptions[name]; ok {
		return desc
	}
	return defaultValue
}
