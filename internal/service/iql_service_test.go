package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIQLService(t *testing.T) {
	service := NewIQLService()
	assert.NotNil(t, service)
}

func TestQueryBuilder_WithWorkspace(t *testing.T) {
	service := NewIQLService()

	tests := []struct {
		name      string
		workspace string
		want      string
	}{
		{
			name:      "with workspace",
			workspace: "TestSpace",
			want:      "所属空间 = 'TestSpace'",
		},
		{
			name:      "empty workspace",
			workspace: "",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := service.NewQuery().WithWorkspace(tt.workspace).Build()
			assert.Equal(t, tt.want, query)
		})
	}
}

func TestQueryBuilder_WithTypeIn(t *testing.T) {
	service := NewIQLService()

	tests := []struct {
		name  string
		types []string
		want  string
	}{
		{
			name:  "single type",
			types: []string{"需求"},
			want:  `类型 in ["需求"]`,
		},
		{
			name:  "multiple types",
			types: []string{"需求", "任务"},
			want:  `类型 in ["需求","任务"]`,
		},
		{
			name:  "empty types",
			types: []string{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := service.NewQuery().WithTypeIn(tt.types).Build()
			assert.Equal(t, tt.want, query)
		})
	}
}

func TestQueryBuilder_WithKeyword(t *testing.T) {
	service := NewIQLService()

	query := service.NewQuery().WithKeyword("test keyword").Build()
	assert.Equal(t, "名称包含 'test keyword'", query)
}

func TestQueryBuilder_CombinedConditions(t *testing.T) {
	service := NewIQLService()

	query := service.NewQuery().
		WithWorkspace("MySpace").
		WithTypeIn([]string{"需求"}).
		WithKeyword("search").
		OrderByCreateTime().
		Build()

	assert.Equal(t, "(所属空间 = 'MySpace' and 类型 in [\"需求\"] and 名称包含 'search') order by 创建时间 desc", query)
}

func TestQueryBuilder_WithLimit(t *testing.T) {
	service := NewIQLService()

	query := service.NewQuery().
		WithWorkspace("Space").
		Limit(10).
		Build()

	assert.Equal(t, "所属空间 = 'Space' limit 10", query)
}

func TestQueryBuilder_EscapeQuotes(t *testing.T) {
	service := NewIQLService()

	query := service.NewQuery().
		WithKeyword("it's a test").
		Build()

	assert.Equal(t, `名称包含 'it\'s a test'`, query)
}

func TestIQLService_BuildRequirementListQuery(t *testing.T) {
	service := NewIQLService()

	tests := []struct {
		name      string
		workspace string
		keyword   string
		want      string
	}{
		{
			name:      "basic requirement list",
			workspace: "DevSpace",
			keyword:   "",
			want:      "(所属空间 = 'DevSpace' and 类型 in [\"需求\"]) order by 创建时间 desc",
		},
		{
			name:      "requirement list with keyword",
			workspace: "DevSpace",
			keyword:   "feature",
			want:      "(所属空间 = 'DevSpace' and 类型 in [\"需求\"] and 名称包含 'feature') order by 创建时间 desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := service.BuildRequirementListQuery(tt.workspace, tt.keyword)
			assert.Equal(t, tt.want, query)
		})
	}
}

func TestIQLService_BuildTaskListQuery(t *testing.T) {
	service := NewIQLService()

	tests := []struct {
		name      string
		workspace string
		keyword   string
		status    string
		want      string
	}{
		{
			name:      "basic task list",
			workspace: "DevSpace",
			keyword:   "",
			status:    "",
			want:      "(所属空间 = 'DevSpace' and 类型 in [\"任务\"]) order by 创建时间 desc",
		},
		{
			name:      "task list with filters",
			workspace: "DevSpace",
			keyword:   "fix",
			status:    "进行中",
			want:      "(所属空间 = 'DevSpace' and 类型 in [\"任务\"] and 名称包含 'fix' and 状态 = '进行中') order by 创建时间 desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := service.BuildTaskListQuery(tt.workspace, tt.keyword, tt.status)
			assert.Equal(t, tt.want, query)
		})
	}
}

func TestIQLService_BuildAssignedToMeQuery(t *testing.T) {
	service := NewIQLService()

	query := service.BuildAssignedToMeQuery("DevSpace", []string{"需求", "任务"}, "john")
	assert.Equal(t, "(所属空间 = 'DevSpace' and 类型 in [\"需求\",\"任务\"] and 负责人 = 'john') order by 创建时间 desc", query)
}

func TestIQLService_BuildActiveItemsQuery(t *testing.T) {
	service := NewIQLService()

	query := service.BuildActiveItemsQuery("DevSpace", []string{"需求"})
	assert.Equal(t, "(所属空间 = 'DevSpace' and 类型 in [\"需求\"] and 状态 != '已关闭') order by 创建时间 desc", query)
}
