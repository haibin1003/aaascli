package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResponseFormatter(t *testing.T) {
	formatter := NewResponseFormatter()
	assert.NotNil(t, formatter)
}

func TestResponseFormatter_FormatRequirement(t *testing.T) {
	formatter := NewResponseFormatter()

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "complete requirement",
			input: map[string]interface{}{
				"objectId":     "req-123",
				"name":         "Test Requirement",
				"key":          "REQ-001",
				"status":       map[string]interface{}{"name": "进行中"},
				"assignee":     map[string]interface{}{"nickname": "John"},
				"creator":      map[string]interface{}{"nickname": "Admin"},
				"createdDate":  "2024-01-15T10:00:00Z",
				"updatedDate":  "2024-01-16T10:00:00Z",
				"priority":     "高",
				"itemType":     "功能需求",
			},
			expected: map[string]interface{}{
				"objectId":   "req-123",
				"name":       "Test Requirement",
				"key":        "REQ-001",
				"status":     "进行中",
				"assignee":   "John",
				"creator":    "Admin",
				"createTime": "2024-01-15 10:00",
				"updateTime": "2024-01-16 10:00",
				"priority":   "高",
				"type":       "功能需求",
			},
		},
		{
			name: "requirement without status",
			input: map[string]interface{}{
				"objectId": "req-456",
				"name":     "Another Requirement",
			},
			expected: map[string]interface{}{
				"objectId":   "req-456",
				"name":       "Another Requirement",
				"key":        "",
				"status":     "未知",
				"assignee":   "未分配",
				"creator":    "未知",
				"createTime": "",
				"updateTime": "",
				"priority":   "",
				"type":       "",
			},
		},
		{
			name:     "nil requirement",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatRequirement(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseFormatter_FormatRequirementList(t *testing.T) {
	formatter := NewResponseFormatter()

	items := []map[string]interface{}{
		{
			"objectId": "req-1",
			"name":     "First",
			"status":   map[string]interface{}{"name": "新建"},
		},
		{
			"objectId": "req-2",
			"name":     "Second",
			"status":   map[string]interface{}{"name": "进行中"},
		},
	}

	result := formatter.FormatRequirementList(items)
	assert.Len(t, result, 2)
	assert.Equal(t, "req-1", result[0]["objectId"])
	assert.Equal(t, "req-2", result[1]["objectId"])
}

func TestResponseFormatter_FormatTask(t *testing.T) {
	formatter := NewResponseFormatter()

	input := map[string]interface{}{
		"objectId":         "task-123",
		"name":             "Test Task",
		"key":              "TASK-001",
		"status":           map[string]interface{}{"name": "进行中"},
		"assignee":         map[string]interface{}{"nickname": "Developer"},
		"createdDate":      "2024-01-15T10:00:00Z",
		"plannedStartDate": "2024-01-16",
		"plannedEndDate":   "2024-01-20",
		"ancestors": []interface{}{
			map[string]interface{}{"name": "Parent Req"},
		},
	}

	result := formatter.FormatTask(input)
	assert.Equal(t, "task-123", result["objectId"])
	assert.Equal(t, "Test Task", result["name"])
	assert.Equal(t, "进行中", result["status"])
	assert.Equal(t, "Developer", result["assignee"])
	assert.Equal(t, "2024-01-15 10:00", result["createTime"])
	assert.Equal(t, "2024-01-16", result["plannedStart"])
	assert.Equal(t, "2024-01-20", result["plannedEnd"])

	ancestors, ok := result["ancestors"].([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"Parent Req"}, ancestors)
}

func TestResponseFormatter_FormatBug(t *testing.T) {
	formatter := NewResponseFormatter()

	input := map[string]interface{}{
		"id":             int64(123),
		"title":          "Bug Title",
		"statusName":     "待处理",
		"priority":       2,
		"level":          3,
		"assignUserName": "Tester",
		"createUserName": "Reporter",
		"createTime":     "2024-01-15T10:00:00Z",
	}

	result := formatter.FormatBug(input)
	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "Bug Title", result["title"])
	assert.Equal(t, "待处理", result["status"])
	assert.Equal(t, "主要", result["priority"])
	assert.Equal(t, "高", result["level"])
	assert.Equal(t, "Tester", result["assignee"])
	assert.Equal(t, "Reporter", result["creator"])
	assert.Equal(t, "2024-01-15 10:00", result["createTime"])
}

func TestResponseFormatter_FormatRepo(t *testing.T) {
	formatter := NewResponseFormatter()

	input := map[string]interface{}{
		"id":         int64(456),
		"name":       "my-repo",
		"path":       "group/my-repo",
		"httpUrl":    "https://git.example.com/group/my-repo.git",
		"sshUrl":     "git@git.example.com:group/my-repo.git",
		"creatorName": "Admin",
		"createTime": "2024-01-15T10:00:00Z",
	}

	result := formatter.FormatRepo(input)
	assert.Equal(t, int64(456), result["id"])
	assert.Equal(t, "my-repo", result["name"])
	assert.Equal(t, "group/my-repo", result["path"])
	assert.Equal(t, "https://git.example.com/group/my-repo.git", result["httpURL"])
	assert.Equal(t, "git@git.example.com:group/my-repo.git", result["sshURL"])
	assert.Equal(t, "Admin", result["creator"])
}

func TestResponseFormatter_FormatMR(t *testing.T) {
	formatter := NewResponseFormatter()

	input := map[string]interface{}{
		"id":           int64(789),
		"title":        "Fix bug",
		"sourceBranch": "feature/bug-fix",
		"targetBranch": "master",
		"authorName":   "Developer",
		"state":        "opened",
		"createdAt":    "2024-01-15T10:00:00Z",
	}

	result := formatter.FormatMR(input)
	assert.Equal(t, int64(789), result["id"])
	assert.Equal(t, "Fix bug", result["title"])
	assert.Equal(t, "feature/bug-fix", result["sourceBranch"])
	assert.Equal(t, "master", result["targetBranch"])
	assert.Equal(t, "Developer", result["author"])
	assert.Equal(t, "opened", result["status"])
}

func TestResponseFormatter_formatTime(t *testing.T) {
	formatter := NewResponseFormatter()

	tests := []struct {
		input    string
		expected string
	}{
		{"2024-01-15T10:30:00Z", "2024-01-15 10:30"},
		{"2024-01-15 10:30:00", "2024-01-15 10:30"},
		{"invalid", "invalid"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatter.formatTime(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPriorityName(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{0, "提示"},
		{1, "次要"},
		{2, "主要"},
		{3, "严重"},
		{4, "致命"},
		{99, "未知"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getPriorityName(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetLevelName(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{1, "低"},
		{2, "中"},
		{3, "高"},
		{99, "未知"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getLevelName(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "string value",
			m:        map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "int value",
			m:        map[string]interface{}{"key": 123},
			key:      "key",
			expected: "123",
		},
		{
			name:     "missing key",
			m:        map[string]interface{}{},
			key:      "missing",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getString(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
