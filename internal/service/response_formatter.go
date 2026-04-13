// Package service provides business logic layer between commands and API clients
package service

import (
	"fmt"
	"time"
)

// ResponseFormatter provides standardized response formatting across all commands
type ResponseFormatter struct{}

// NewResponseFormatter creates a new response formatter
func NewResponseFormatter() *ResponseFormatter {
	return &ResponseFormatter{}
}

// FormatRequirementList formats requirement items for consistent output
func (f *ResponseFormatter) FormatRequirementList(items []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		formatted := f.FormatRequirement(item)
		if formatted != nil {
			result = append(result, formatted)
		}
	}

	return result
}

// FormatRequirement formats a single requirement item
func (f *ResponseFormatter) FormatRequirement(item map[string]interface{}) map[string]interface{} {
	if item == nil {
		return nil
	}

	return map[string]interface{}{
		"objectId":    getString(item, "objectId"),
		"name":        getString(item, "name"),
		"key":         getString(item, "key"),
		"status":      f.getStatusName(item),
		"assignee":    f.getAssigneeName(item),
		"creator":     f.getCreatorName(item),
		"createTime":  f.formatTime(getString(item, "createdDate")),
		"updateTime":  f.formatTime(getString(item, "updatedDate")),
		"priority":    getString(item, "priority"),
		"type":        getString(item, "itemType"),
	}
}

// FormatTaskList formats task items for consistent output
func (f *ResponseFormatter) FormatTaskList(items []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		formatted := f.FormatTask(item)
		if formatted != nil {
			result = append(result, formatted)
		}
	}

	return result
}

// FormatTask formats a single task item
func (f *ResponseFormatter) FormatTask(item map[string]interface{}) map[string]interface{} {
	if item == nil {
		return nil
	}

	return map[string]interface{}{
		"objectId":     getString(item, "objectId"),
		"name":         getString(item, "name"),
		"key":          getString(item, "key"),
		"status":       f.getStatusName(item),
		"assignee":     f.getAssigneeName(item),
		"createTime":   f.formatTime(getString(item, "createdDate")),
		"plannedStart": getString(item, "plannedStartDate"),
		"plannedEnd":   getString(item, "plannedEndDate"),
		"ancestors":    getAncestorNames(item),
	}
}

// FormatBugList formats bug items for consistent output
func (f *ResponseFormatter) FormatBugList(items []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		formatted := f.FormatBug(item)
		if formatted != nil {
			result = append(result, formatted)
		}
	}

	return result
}

// FormatBug formats a single bug item
func (f *ResponseFormatter) FormatBug(item map[string]interface{}) map[string]interface{} {
	if item == nil {
		return nil
	}

	// Extract priority name
	priorityName := getPriorityName(getInt(item, "priority"))
	levelName := getLevelName(getInt(item, "level"))

	return map[string]interface{}{
		"id":          getString(item, "id"),
		"title":       getString(item, "title"),
		"status":      getString(item, "statusName"),
		"priority":    priorityName,
		"level":       levelName,
		"assignee":    getString(item, "assignUserName"),
		"creator":     getString(item, "createUserName"),
		"createTime":  f.formatTime(getString(item, "createTime")),
	}
}

// FormatRepoList formats repository items for consistent output
func (f *ResponseFormatter) FormatRepoList(items []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		formatted := f.FormatRepo(item)
		if formatted != nil {
			result = append(result, formatted)
		}
	}

	return result
}

// FormatRepo formats a single repository item
func (f *ResponseFormatter) FormatRepo(item map[string]interface{}) map[string]interface{} {
	if item == nil {
		return nil
	}

	return map[string]interface{}{
		"id":       getInt64(item, "id"),
		"name":     getString(item, "name"),
		"path":     getString(item, "path"),
		"httpURL":  getString(item, "httpUrl"),
		"sshURL":   getString(item, "sshUrl"),
		"creator":  getString(item, "creatorName"),
		"createTime": f.formatTime(getString(item, "createTime")),
	}
}

// FormatMRList formats merge request items for consistent output
func (f *ResponseFormatter) FormatMRList(items []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		formatted := f.FormatMR(item)
		if formatted != nil {
			result = append(result, formatted)
		}
	}

	return result
}

// FormatMR formats a single merge request item
func (f *ResponseFormatter) FormatMR(item map[string]interface{}) map[string]interface{} {
	if item == nil {
		return nil
	}

	return map[string]interface{}{
		"id":           getInt64(item, "id"),
		"title":        getString(item, "title"),
		"sourceBranch": getString(item, "sourceBranch"),
		"targetBranch": getString(item, "targetBranch"),
		"author":       getString(item, "authorName"),
		"status":       getString(item, "state"),
		"createdAt":    f.formatTime(getString(item, "createdAt")),
	}
}

// Helper methods

func (f *ResponseFormatter) getStatusName(item map[string]interface{}) string {
	if status, ok := item["status"].(map[string]interface{}); ok {
		return getString(status, "name")
	}
	return "未知"
}

func (f *ResponseFormatter) getAssigneeName(item map[string]interface{}) string {
	if assignee, ok := item["assignee"].(map[string]interface{}); ok {
		return getString(assignee, "nickname")
	}
	return "未分配"
}

func (f *ResponseFormatter) getCreatorName(item map[string]interface{}) string {
	if creator, ok := item["creator"].(map[string]interface{}); ok {
		return getString(creator, "nickname")
	}
	return "未知"
}

func (f *ResponseFormatter) formatTime(timeStr string) string {
	if timeStr == "" {
		return ""
	}

	// Try to parse common time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t.Format("2006-01-02 15:04")
		}
	}

	// Return original if parsing fails
	return timeStr
}

// getPriorityName returns priority name from code
func getPriorityName(code int) string {
	names := map[int]string{
		0: "提示",
		1: "次要",
		2: "主要",
		3: "严重",
		4: "致命",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return "未知"
}

// getLevelName returns level name from code
func getLevelName(code int) string {
	names := map[int]string{
		1: "低",
		2: "中",
		3: "高",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return "未知"
}

// Helper functions for safe type extraction

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case fmt.Stringer:
			return val.String()
		default:
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int64:
			return val
		case int:
			return int64(val)
		case float64:
			return int64(val)
		}
	}
	return 0
}

func getAncestorNames(item map[string]interface{}) []string {
	if ancestors, ok := item["ancestors"].([]interface{}); ok {
		names := make([]string, 0, len(ancestors))
		for _, ancestor := range ancestors {
			if a, ok := ancestor.(map[string]interface{}); ok {
				names = append(names, getString(a, "name"))
			}
		}
		return names
	}
	return nil
}
