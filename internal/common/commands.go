package common

// CommandMeta 命令元数据
// 统一描述一条命令的属性，同时服务于只读检查和 OTP 检查
type CommandMeta struct {
	Description string // 命令的中文描述，用于 OTP 警告提示
	IsWrite     bool   // true = 写操作，在只读模式下会被拦截
	RiskLevel   string // 风险等级: critical / high / medium / low
	Reason      string // 需要 OTP 验证的原因说明
}

// CommandRegistry 统一命令注册表
// 所有需要参与只读检查或 OTP 检查的命令都在此处注册一次。
//
// 字段语义：
//   - IsWrite=true  → 只读模式下被拦截（等价于原 WriteCommands 集合）
//   - IsWrite=false → 读操作，只读模式放行；但用户可将其加入 OTP 保护列表
//   - RiskLevel      → 用于 OTP 警告提示，与 DefaultProtectedCommands 无强绑定关系
var CommandRegistry = map[string]CommandMeta{

	// ── 只读模式 ───────────────────────────────────────────────────────
	"readonly on": {
		Description: "开启只读模式",
		IsWrite:     false, // 安全操作，不需要只读拦截，也不纳入默认 OTP 保护
		RiskLevel:   "low",
		Reason:      "限制写入操作，提高安全性",
	},
	"readonly off": {
		Description: "关闭只读模式",
		IsWrite:     false, // 不受只读模式拦截（否则无法开启写权限）
		RiskLevel:   "medium",
		Reason:      "允许执行写入操作，增加误操作风险",
	},

	// ── 代码仓库 ───────────────────────────────────────────────────────
	"repo create": {
		Description: "创建代码仓库",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "创建新的代码仓库",
	},
	"repo delete": {
		Description: "删除代码仓库",
		IsWrite:     true,
		RiskLevel:   "critical",
		Reason:      "仓库删除后所有代码和历史记录将无法恢复",
	},
	"repo group add": {
		Description: "添加仓库代码组",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "修改仓库的代码组归属",
	},
	"repo disable-work-item-link": {
		Description: "关闭工作项关联",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "关闭后 MR 将不再要求关联工作项",
	},
	"repo list": {
		Description: "查询仓库列表",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取仓库列表数据",
	},
	"repo search": {
		Description: "搜索仓库",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "按关键词搜索仓库",
	},
	"repo group list": {
		Description: "查询代码组列表",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取代码组列表",
	},
	"repo personal-group": {
		Description: "查询个人代码组",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取当前用户的个人代码组",
	},

	// ── 合并请求 ───────────────────────────────────────────────────────
	"pr create": {
		Description: "创建合并请求",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "创建新的代码合并请求",
	},
	"pr review": {
		Description: "审核合并请求",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "批准或拒绝代码合并请求",
	},
	"pr merge": {
		Description: "合并代码请求",
		IsWrite:     true,
		RiskLevel:   "high",
		Reason:      "代码合并不可逆，可能影响生产环境",
	},
	"pr comment": {
		Description: "评论合并请求",
		IsWrite:     true,
		RiskLevel:   "low",
		Reason:      "在合并请求下发表评论",
	},
	"pr patch-comment": {
		Description: "修改评论状态",
		IsWrite:     true,
		RiskLevel:   "low",
		Reason:      "修改合并请求中评论的解决状态",
	},
	"pr list": {
		Description: "查询合并请求列表",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取 MR 列表数据",
	},
	"pr view": {
		Description: "查看合并请求详情",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取 MR 详情及评论",
	},

	// ── 需求 ───────────────────────────────────────────────────────────
	"req create": {
		Description: "创建需求",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "创建新的需求条目",
	},
	"req update": {
		Description: "更新需求",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "修改需求的标题、描述或状态",
	},
	"req delete": {
		Description: "删除需求",
		IsWrite:     true,
		RiskLevel:   "high",
		Reason:      "删除后数据无法恢复",
	},
	"req list": {
		Description: "查询需求列表",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取需求列表数据",
	},
	"req view": {
		Description: "查看需求详情",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取单条需求详情",
	},
	"req search": {
		Description: "搜索需求",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "按关键词搜索需求",
	},

	// ── 任务 ───────────────────────────────────────────────────────────
	"task create": {
		Description: "创建任务",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "在需求下创建新的任务",
	},
	"task delete": {
		Description: "删除任务",
		IsWrite:     true,
		RiskLevel:   "high",
		Reason:      "删除后数据无法恢复",
	},
	"task list": {
		Description: "查询任务列表",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取任务列表数据",
	},
	"task search": {
		Description: "搜索任务",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "按关键词搜索任务",
	},

	// ── 缺陷 ───────────────────────────────────────────────────────────
	"bug create": {
		Description: "创建缺陷",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "创建新的缺陷条目",
	},
	"bug update-status": {
		Description: "更新缺陷状态",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "修改缺陷的处理状态",
	},
	"bug delete": {
		Description: "删除缺陷",
		IsWrite:     true,
		RiskLevel:   "high",
		Reason:      "删除后数据无法恢复",
	},
	"bug list": {
		Description: "查询缺陷列表",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取缺陷列表数据",
	},
	"bug view": {
		Description: "查看缺陷详情",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取单条缺陷详情",
	},
	"bug list-statuses": {
		Description: "查询缺陷状态列表",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取可用缺陷状态",
	},

	// ── 知识库 ───────────────────────────────────────────────────────
	"lib create": {
		Description: "创建知识库",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "创建新的知识库",
	},
	"lib delete": {
		Description: "删除知识库",
		IsWrite:     true,
		RiskLevel:   "high",
		Reason:      "删除后库内所有文档将无法恢复",
	},
	"lib file delete": {
		Description: "删除知识库文件",
		IsWrite:     true,
		RiskLevel:   "high",
		Reason:      "文件删除后无法恢复",
	},
	"lib upload": {
		Description: "上传文件到知识库",
		IsWrite:     true,
		RiskLevel:   "medium",
		Reason:      "向知识库上传新文件",
	},
	"lib folder create": {
		Description: "创建知识库目录",
		IsWrite:     true,
		RiskLevel:   "low",
		Reason:      "在知识库中创建新目录",
	},
	"lib list": {
		Description: "查询知识库列表",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取知识库列表",
	},
	"lib folder tree": {
		Description: "查看知识库目录树",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取知识库目录结构",
	},
	"lib folder list": {
		Description: "查询知识库目录列表",
		IsWrite:     false,
		RiskLevel:   "low",
		Reason:      "读取目录下文件列表",
	},
}
