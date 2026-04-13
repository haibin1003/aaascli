// Package common provides shared utilities for command execution.
package common

import (
	"fmt"
	"time"
)

// FormatTimestamp converts various timestamp formats to human-readable string
func FormatTimestamp(value interface{}) string {
	switch v := value.(type) {
	case float64:
		t := time.Unix(int64(v)/1000, 0)
		return t.Format("2006-01-02 15:04:05")
	case int64:
		t := time.Unix(v/1000, 0)
		return t.Format("2006-01-02 15:04:05")
	case string:
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
		return v
	}
	return ""
}

// ExtractTextFromRichText extracts plain text from rich text format
// Handles the []map[string]interface{} type returned by API
func ExtractTextFromRichText(value interface{}) string {
	// Handle []map[string]interface{} type
	if arr, ok := value.([]interface{}); ok {
		var result string
		for _, item := range arr {
			if block, ok := item.(map[string]interface{}); ok {
				if children, ok := block["children"].([]interface{}); ok {
					for _, child := range children {
						if childMap, ok := child.(map[string]interface{}); ok {
							if text, ok := childMap["text"].(string); ok {
								result += text
							}
						}
					}
				}
			}
			result += "\n"
		}
		return result
	}
	return ""
}

// DisplayRichTextField displays rich text field content
// Usage: DisplayRichTextField(values, "businessBackground", "业务背景")
func DisplayRichTextField(values map[string]interface{}, fieldKey, fieldName string) {
	if value, ok := values[fieldKey]; ok && value != nil {
		fmt.Printf("\n【%s】\n", fieldName)
		text := ExtractTextFromRichText(value)
		if text != "" {
			fmt.Printf("%s\n", text)
		} else {
			fmt.Printf("(空)\n")
		}
	}
}
