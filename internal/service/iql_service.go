package service

import (
	"fmt"
	"strings"
)

// IQLService provides Item Query Language (IQL) query building
type IQLService struct{}

// NewIQLService creates a new IQL service
func NewIQLService() *IQLService {
	return &IQLService{}
}

// QueryBuilder provides a fluent interface for building IQL queries
type QueryBuilder struct {
	service    *IQLService
	conditions []string
	orderBy    string
	desc       bool
	limit      int
}

// NewQuery creates a new IQL query builder
func (s *IQLService) NewQuery() *QueryBuilder {
	return &QueryBuilder{
		service:    s,
		conditions: make([]string, 0),
	}
}

// WithWorkspace adds workspace filter condition
func (qb *QueryBuilder) WithWorkspace(workspaceName string) *QueryBuilder {
	if workspaceName != "" {
		qb.conditions = append(qb.conditions,
			fmt.Sprintf("所属空间 = '%s'", qb.service.escapeString(workspaceName)))
	}
	return qb
}

// WithTypeIn adds type filter with IN operator
func (qb *QueryBuilder) WithTypeIn(types []string) *QueryBuilder {
	if len(types) > 0 {
		escaped := make([]string, len(types))
		for i, t := range types {
			escaped[i] = fmt.Sprintf("\"%s\"", qb.service.escapeString(t))
		}
		qb.conditions = append(qb.conditions,
			fmt.Sprintf("类型 in [%s]", strings.Join(escaped, ",")))
	}
	return qb
}

// WithKeyword adds keyword search condition (matches name)
func (qb *QueryBuilder) WithKeyword(keyword string) *QueryBuilder {
	if keyword != "" {
		qb.conditions = append(qb.conditions,
			fmt.Sprintf("名称包含 '%s'", qb.service.escapeString(keyword)))
	}
	return qb
}

// WithStatus adds status filter
func (qb *QueryBuilder) WithStatus(status string) *QueryBuilder {
	if status != "" {
		qb.conditions = append(qb.conditions,
			fmt.Sprintf("状态 = '%s'", qb.service.escapeString(status)))
	}
	return qb
}

// WithAssignee adds assignee filter
func (qb *QueryBuilder) WithAssignee(assignee string) *QueryBuilder {
	if assignee != "" {
		qb.conditions = append(qb.conditions,
			fmt.Sprintf("负责人 = '%s'", qb.service.escapeString(assignee)))
	}
	return qb
}

// WithCondition adds a custom condition
func (qb *QueryBuilder) WithCondition(condition string) *QueryBuilder {
	if condition != "" {
		qb.conditions = append(qb.conditions, condition)
	}
	return qb
}

// OrderBy sets the order by field
func (qb *QueryBuilder) OrderBy(field string, desc bool) *QueryBuilder {
	qb.orderBy = field
	qb.desc = desc
	return qb
}

// OrderByCreateTime orders by creation time (descending by default)
func (qb *QueryBuilder) OrderByCreateTime() *QueryBuilder {
	return qb.OrderBy("创建时间", true)
}

// Limit sets the result limit
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Build constructs the final IQL query string
func (qb *QueryBuilder) Build() string {
	var parts []string

	// Combine conditions with AND
	if len(qb.conditions) > 0 {
		if len(qb.conditions) == 1 {
			parts = append(parts, qb.conditions[0])
		} else {
			parts = append(parts, fmt.Sprintf("(%s)",
				strings.Join(qb.conditions, " and ")))
		}
	}

	query := ""
	if len(parts) > 0 {
		query = parts[0]
	}

	// Add order by
	if qb.orderBy != "" {
		direction := "asc"
		if qb.desc {
			direction = "desc"
		}
		query = fmt.Sprintf("%s order by %s %s", query, qb.orderBy, direction)
	}

	// Add limit
	if qb.limit > 0 {
		query = fmt.Sprintf("%s limit %d", query, qb.limit)
	}

	return strings.TrimSpace(query)
}

// escapeString escapes single quotes in strings for IQL
func (s *IQLService) escapeString(str string) string {
	return strings.ReplaceAll(str, "'", "\\'")
}

// BuildRequirementListQuery builds a query for listing requirements
func (s *IQLService) BuildRequirementListQuery(workspace, keyword string) string {
	return s.NewQuery().
		WithWorkspace(workspace).
		WithTypeIn([]string{"需求"}).
		WithKeyword(keyword).
		OrderByCreateTime().
		Build()
}

// BuildTaskListQuery builds a query for listing tasks
func (s *IQLService) BuildTaskListQuery(workspace, keyword, status string) string {
	return s.NewQuery().
		WithWorkspace(workspace).
		WithTypeIn([]string{"任务"}).
		WithKeyword(keyword).
		WithStatus(status).
		OrderByCreateTime().
		Build()
}

// BuildRequirementSearchQuery builds a query for searching requirements with advanced filters
func (s *IQLService) BuildRequirementSearchQuery(workspace, keyword, status, assignee string) string {
	qb := s.NewQuery().
		WithWorkspace(workspace).
		WithTypeIn([]string{"需求"})

	if keyword != "" {
		// Search in both name and description
		qb = qb.WithCondition(fmt.Sprintf("(名称包含 '%s' or 描述包含 '%s')",
			s.escapeString(keyword), s.escapeString(keyword)))
	}

	if status != "" {
		qb = qb.WithStatus(status)
	}

	if assignee != "" {
		qb = qb.WithAssignee(assignee)
	}

	return qb.OrderByCreateTime().Build()
}

// BuildAssignedToMeQuery builds a query for items assigned to a specific user
func (s *IQLService) BuildAssignedToMeQuery(workspace string, types []string, username string) string {
	return s.NewQuery().
		WithWorkspace(workspace).
		WithTypeIn(types).
		WithAssignee(username).
		OrderByCreateTime().
		Build()
}

// BuildActiveItemsQuery builds a query for active (non-closed) items
func (s *IQLService) BuildActiveItemsQuery(workspace string, types []string) string {
	return s.NewQuery().
		WithWorkspace(workspace).
		WithTypeIn(types).
		WithCondition("状态 != '已关闭'").
		OrderByCreateTime().
		Build()
}
