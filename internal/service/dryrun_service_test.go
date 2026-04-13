package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDryRunService(t *testing.T) {
	service := NewDryRunService()
	assert.NotNil(t, service)
}

func TestDryRunService_SimulateCreate(t *testing.T) {
	service := NewDryRunService()

	tests := []struct {
		name     string
		resource Resource
		resName  string
		details  map[string]interface{}
	}{
		{
			name:     "create requirement",
			resource: ResRequirement,
			resName:  "New Feature",
			details:  map[string]interface{}{"priority": "高"},
		},
		{
			name:     "create task",
			resource: ResTask,
			resName:  "Implement API",
			details:  nil,
		},
		{
			name:     "create bug",
			resource: ResBug,
			resName:  "Fix Crash",
			details:  map[string]interface{}{"priority": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.SimulateCreate(tt.resource, tt.resName, tt.details)

			assert.True(t, result["dryRun"].(bool))
			assert.Equal(t, "create", result["action"])
			assert.Equal(t, string(tt.resource), result["resource"])
			assert.Contains(t, result["summary"], tt.resName)

			if tt.details != nil {
				assert.Equal(t, tt.details, result["details"])
			}
		})
	}
}

func TestDryRunService_SimulateDelete(t *testing.T) {
	service := NewDryRunService()

	t.Run("single item", func(t *testing.T) {
		result := service.SimulateDelete(ResRequirement, []string{"req-123"})

		assert.True(t, result["dryRun"].(bool))
		assert.Equal(t, "delete", result["action"])
		assert.Contains(t, result["summary"], "req-123")
		assert.Equal(t, []string{"req-123"}, result["identifiers"])
	})

	t.Run("multiple items", func(t *testing.T) {
		ids := []string{"req-1", "req-2", "req-3"}
		result := service.SimulateDelete(ResTask, ids)

		assert.True(t, result["dryRun"].(bool))
		assert.Contains(t, result["summary"], "3")
		assert.Equal(t, ids, result["identifiers"])
	})
}

func TestDryRunService_SimulateList(t *testing.T) {
	service := NewDryRunService()

	filters := map[string]string{
		"status":   "进行中",
		"assignee": "john",
	}

	result := service.SimulateList(ResRequirement, filters)

	assert.True(t, result["dryRun"].(bool))
	assert.Equal(t, "list", result["action"])
	assert.Equal(t, string(ResRequirement), result["resource"])
	assert.Contains(t, result["summary"], "列表")
	assert.Equal(t, filters, result["filters"])
}

func TestDryRunService_SimulateUpdate(t *testing.T) {
	service := NewDryRunService()

	changes := map[string]interface{}{
		"status":   "已完成",
		"assignee": "jane",
	}

	result := service.SimulateUpdate(ResTask, "task-123", changes)

	assert.True(t, result["dryRun"].(bool))
	assert.Equal(t, "update", result["action"])
	assert.Equal(t, "task-123", result["id"])
	assert.Equal(t, changes, result["changes"])
}

func TestDryRunService_SimulateMerge(t *testing.T) {
	service := NewDryRunService()

	details := map[string]interface{}{
		"title":       "Feature merge",
		"description": "Merging feature branch",
	}

	result := service.SimulateMerge("feature/new", "master", details)

	assert.True(t, result["dryRun"].(bool))
	assert.Equal(t, "merge", result["action"])
	assert.Equal(t, "feature/new", result["sourceBranch"])
	assert.Equal(t, "master", result["targetBranch"])
	assert.Contains(t, result["summary"], "feature/new")
	assert.Contains(t, result["summary"], "master")
	assert.Equal(t, details, result["details"])
}

func TestDryRunService_SimulateReview(t *testing.T) {
	service := NewDryRunService()

	result := service.SimulateReview("42", "approve")

	assert.True(t, result["dryRun"].(bool))
	assert.Equal(t, "review", result["action"])
	assert.Equal(t, "42", result["mrId"])
	assert.Equal(t, "approve", result["reviewAction"])
	assert.Contains(t, result["summary"], "42")
	assert.Contains(t, result["summary"], "approve")
}

func TestDryRunService_getResourceName(t *testing.T) {
	service := NewDryRunService()

	tests := []struct {
		resource Resource
		expected string
	}{
		{ResRequirement, "需求"},
		{ResTask, "任务"},
		{ResBug, "缺陷"},
		{ResRepo, "仓库"},
		{ResMR, "合并请求"},
		{ResLibrary, "文档库"},
		{ResFolder, "文件夹"},
		{ResComment, "评论"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.resource), func(t *testing.T) {
			result := service.getResourceName(tt.resource)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDryRunService_getResourceNamePlural(t *testing.T) {
	service := NewDryRunService()

	// Verify plural names are Chinese (same as singular in Chinese)
	assert.Equal(t, "需求", service.getResourceNamePlural(ResRequirement))
	assert.Equal(t, "任务", service.getResourceNamePlural(ResTask))
	assert.Equal(t, "缺陷", service.getResourceNamePlural(ResBug))
}
