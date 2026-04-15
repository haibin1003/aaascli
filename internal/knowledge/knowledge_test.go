package knowledge

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func listMarkdownNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func TestKnowledgeSync(t *testing.T) {
	srcDir := filepath.Join("..", "..", "docs", "knowledge")
	embedDir := "."

	srcNames, err := listMarkdownNames(srcDir)
	if err != nil {
		t.Fatalf("read docs/knowledge failed: %v", err)
	}

	embedNames, err := listMarkdownNames(embedDir)
	if err != nil {
		t.Fatalf("read internal/knowledge failed: %v", err)
	}

	if len(srcNames) != len(embedNames) {
		t.Errorf("knowledge count mismatch: docs=%d, internal=%d", len(srcNames), len(embedNames))
	}

	max := len(srcNames)
	if len(embedNames) > max {
		max = len(embedNames)
	}
	for i := 0; i < max; i++ {
		var src, emb string
		if i < len(srcNames) {
			src = srcNames[i]
		}
		if i < len(embedNames) {
			emb = embedNames[i]
		}
		if src != emb {
			t.Errorf("knowledge file mismatch at index %d: docs=%q, internal=%q", i, src, emb)
		}
	}
}

func TestListAndGet(t *testing.T) {
	docs, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(docs) == 0 {
		t.Fatal("expected at least one knowledge doc")
	}

	first := docs[0]
	if first.Name == "" {
		t.Error("doc name should not be empty")
	}
	if first.Title == "" {
		t.Error("doc title should not be empty")
	}
	if first.Content == "" {
		t.Error("doc content should not be empty")
	}

	got, err := Get(first.Name)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != first.Name {
		t.Errorf("Get name mismatch: expected %q, got %q", first.Name, got.Name)
	}
}

func TestSearch(t *testing.T) {
	results, err := Search("短信")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Log("no results for '短信', may be expected depending on knowledge base content")
	}
}

func TestExtractSnippets(t *testing.T) {
	content := "line1\nline2 关键词 line3\nline4\nline5\n"
	snippets := ExtractSnippets(content, "关键词", 1)
	if len(snippets) == 0 {
		t.Error("expected at least one snippet")
	}
	if !strings.Contains(snippets[0], "关键词") {
		t.Errorf("snippet should contain keyword: %q", snippets[0])
	}
}
