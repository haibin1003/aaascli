package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/haibin1003/aaascli/internal/common"
	"github.com/haibin1003/aaascli/internal/knowledge"
)

var knowledgeCmd = &cobra.Command{
	Use:   "knowledge",
	Short: "知识库管理",
	Long:  "查询和管理内置知识库文档，帮助 AI 快速获取平台能力的使用建议、场景方案和代码示例",
}

var knowledgeListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有知识文档",
	Run: func(cmd *cobra.Command, args []string) {
		listKnowledgeDocs()
	},
}

var knowledgeViewCmd = &cobra.Command{
	Use:   "view [doc-name]",
	Short: "查看指定知识文档",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viewKnowledgeDoc(args[0])
	},
}

var knowledgeSearchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "搜索知识库内容",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchKnowledge(args[0])
	},
}

func init() {
	rootCmd.AddCommand(knowledgeCmd)
	knowledgeCmd.AddCommand(knowledgeListCmd)
	knowledgeCmd.AddCommand(knowledgeViewCmd)
	knowledgeCmd.AddCommand(knowledgeSearchCmd)
}

func listKnowledgeDocs() {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		docs, err := knowledge.List()
		if err != nil {
			return nil, fmt.Errorf("查询知识库失败: %w", err)
		}
		items := make([]map[string]interface{}, 0)
		for _, doc := range docs {
			items = append(items, map[string]interface{}{
				"name":  doc.Name,
				"title": doc.Title,
			})
		}
		return map[string]interface{}{
			"items": items,
			"total": len(items),
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func viewKnowledgeDoc(name string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		doc, err := knowledge.Get(name)
		if err != nil {
			return nil, fmt.Errorf("读取文档失败: %w", err)
		}
		return map[string]interface{}{
			"name":    doc.Name,
			"title":   doc.Title,
			"content": doc.Content,
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}

func searchKnowledge(keyword string) {
	common.Execute(func(ctx *common.CommandContext) (interface{}, error) {
		docs, err := knowledge.Search(keyword)
		if err != nil {
			return nil, fmt.Errorf("搜索知识库失败: %w", err)
		}
		
		items := make([]map[string]interface{}, 0)
		for _, doc := range docs {
			snippets := knowledge.ExtractSnippets(doc.Content, keyword, 2)
			if len(snippets) == 0 {
				// 如果标题匹配但内容中没有可提取的片段，展示开头
				lines := strings.Split(doc.Content, "\n")
				if len(lines) > 5 {
					snippets = []string{strings.Join(lines[:5], "\n")}
				} else {
					snippets = []string{doc.Content}
				}
			}
			items = append(items, map[string]interface{}{
				"name":     doc.Name,
				"title":    doc.Title,
				"snippets": snippets,
			})
		}
		
		return map[string]interface{}{
			"items": items,
			"total": len(items),
			"keyword": keyword,
		}, nil
	}, common.ExecuteOptions{
		DebugMode:   debugMode,
		Insecure:    insecure,
		DryRun:      dryRun,
		Cookie:      cookieFlag,
		PrettyPrint: prettyPrint,
	})
}
