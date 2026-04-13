package service

import "fmt"

// DryRunService provides dry-run simulation for all operations
type DryRunService struct {
	formatter *ResponseFormatter
}

// NewDryRunService creates a new dry-run service
func NewDryRunService() *DryRunService {
	return &DryRunService{
		formatter: NewResponseFormatter(),
	}
}

// Operation represents the type of operation being simulated
type Operation string

const (
	OpCreate   Operation = "create"
	OpUpdate   Operation = "update"
	OpDelete   Operation = "delete"
	OpList     Operation = "list"
	OpMerge    Operation = "merge"
	OpReview   Operation = "review"
)

// Resource represents the type of resource being operated on
type Resource string

const (
	ResRequirement Resource = "requirement"
	ResTask        Resource = "task"
	ResBug         Resource = "bug"
	ResRepo        Resource = "repository"
	ResMR          Resource = "merge_request"
	ResLibrary     Resource = "library"
	ResFolder      Resource = "folder"
	ResComment     Resource = "comment"
)

// SimulationResult represents a dry-run simulation result
type SimulationResult struct {
	DryRun   bool                   `json:"dryRun"`
	Action   string                 `json:"action"`
	Resource string                 `json:"resource"`
	Summary  string                 `json:"summary"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// SimulateCreate simulates a create operation
func (s *DryRunService) SimulateCreate(resource Resource, name string, details map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"dryRun":   true,
		"action":   string(OpCreate),
		"resource": string(resource),
		"summary":  fmt.Sprintf("将创建%s: %s", s.getResourceName(resource), name),
	}

	if len(details) > 0 {
		result["details"] = details
	}

	return result
}

// SimulateDelete simulates a delete operation
func (s *DryRunService) SimulateDelete(resource Resource, identifiers []string) map[string]interface{} {
	var summary string
	if len(identifiers) == 1 {
		summary = fmt.Sprintf("将删除%s: %s", s.getResourceName(resource), identifiers[0])
	} else {
		summary = fmt.Sprintf("将删除%d个%s", len(identifiers), s.getResourceNamePlural(resource))
	}

	return map[string]interface{}{
		"dryRun":      true,
		"action":      string(OpDelete),
		"resource":    string(resource),
		"summary":     summary,
		"identifiers": identifiers,
	}
}

// SimulateList simulates a list operation
func (s *DryRunService) SimulateList(resource Resource, filters map[string]string) map[string]interface{} {
	return map[string]interface{}{
		"dryRun":   true,
		"action":   string(OpList),
		"resource": string(resource),
		"summary":  fmt.Sprintf("将查询%s列表", s.getResourceNamePlural(resource)),
		"filters":  filters,
	}
}

// SimulateUpdate simulates an update operation
func (s *DryRunService) SimulateUpdate(resource Resource, id string, changes map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"dryRun":   true,
		"action":   string(OpUpdate),
		"resource": string(resource),
		"summary":  fmt.Sprintf("将更新%s: %s", s.getResourceName(resource), id),
		"id":       id,
		"changes":  changes,
	}
}

// SimulateMerge simulates a merge operation
func (s *DryRunService) SimulateMerge(sourceBranch, targetBranch string, details map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"dryRun":       true,
		"action":       string(OpMerge),
		"resource":     string(ResMR),
		"summary":      fmt.Sprintf("将合并分支: %s → %s", sourceBranch, targetBranch),
		"sourceBranch": sourceBranch,
		"targetBranch": targetBranch,
	}

	if len(details) > 0 {
		result["details"] = details
	}

	return result
}

// SimulateReview simulates a code review operation
func (s *DryRunService) SimulateReview(mrID string, action string) map[string]interface{} {
	return map[string]interface{}{
		"dryRun":   true,
		"action":   string(OpReview),
		"resource": string(ResMR),
		"summary":  fmt.Sprintf("将对合并请求 #%s 执行 %s 操作", mrID, action),
		"mrId":     mrID,
		"reviewAction": action,
	}
}

// getResourceName returns the Chinese name for a resource
func (s *DryRunService) getResourceName(resource Resource) string {
	names := map[Resource]string{
		ResRequirement: "需求",
		ResTask:        "任务",
		ResBug:         "缺陷",
		ResRepo:        "仓库",
		ResMR:          "合并请求",
		ResLibrary:     "文档库",
		ResFolder:      "文件夹",
		ResComment:     "评论",
	}
	if name, ok := names[resource]; ok {
		return name
	}
	return string(resource)
}

// getResourceNamePlural returns the plural Chinese name for a resource
func (s *DryRunService) getResourceNamePlural(resource Resource) string {
	names := map[Resource]string{
		ResRequirement: "需求",
		ResTask:        "任务",
		ResBug:         "缺陷",
		ResRepo:        "仓库",
		ResMR:          "合并请求",
		ResLibrary:     "文档库",
		ResFolder:      "文件夹",
		ResComment:     "评论",
	}
	if name, ok := names[resource]; ok {
		return name
	}
	return string(resource)
}
