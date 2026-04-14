package knowledge

import (
	"embed"
	"fmt"
	"path"
	"strings"
)

//go:embed *.md
var knowledgeFS embed.FS

// Doc 知识文档
type Doc struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// List 列出所有知识文档
func List() ([]Doc, error) {
	files, err := knowledgeFS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("read knowledge dir failed: %w", err)
	}

	var docs []Doc
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		content, err := knowledgeFS.ReadFile(file.Name())
		if err != nil {
			continue
		}
		docs = append(docs, Doc{
			Name:    strings.TrimSuffix(file.Name(), ".md"),
			Title:   extractTitle(string(content), file.Name()),
			Content: string(content),
		})
	}
	return docs, nil
}

// Get 获取指定名称的文档
func Get(name string) (*Doc, error) {
	// 安全检查：防止路径遍历
	name = path.Base(name)
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}

	content, err := knowledgeFS.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("document not found: %s", strings.TrimSuffix(name, ".md"))
	}

	return &Doc{
		Name:    strings.TrimSuffix(name, ".md"),
		Title:   extractTitle(string(content), name),
		Content: string(content),
	}, nil
}

// Search 按关键词搜索文档标题和内容
func Search(keyword string) ([]Doc, error) {
	keyword = strings.ToLower(keyword)
	docs, err := List()
	if err != nil {
		return nil, err
	}

	var results []Doc
	for _, doc := range docs {
		if strings.Contains(strings.ToLower(doc.Title), keyword) ||
			strings.Contains(strings.ToLower(doc.Content), keyword) {
			results = append(results, doc)
		}
	}
	return results, nil
}

func extractTitle(content, fallback string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return strings.TrimSuffix(fallback, ".md")
}

// ExtractSnippets 从文档中提取包含关键词的上下文片段
func ExtractSnippets(docContent, keyword string, contextLines int) []string {
	keywordLower := strings.ToLower(keyword)
	lines := strings.Split(docContent, "\n")
	var snippets []string
	used := make(map[int]bool)

	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), keywordLower) {
			start := i - contextLines
			if start < 0 {
				start = 0
			}
			end := i + contextLines + 1
			if end > len(lines) {
				end = len(lines)
			}

			// 避免重复片段
			if used[start] {
				continue
			}
			used[start] = true

			snippet := strings.Join(lines[start:end], "\n")
			snippets = append(snippets, snippet)
		}
	}
	return snippets
}
