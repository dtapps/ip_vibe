package mcp_prompt

import (
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// errorResponse 错误响应（支持格式化）
func errorResponse(format string, args ...any) (*mcp.GetPromptResult, error) {
	return nil, fmt.Errorf(format, args...)
}

// successResponse 成功响应（自动处理多种数据类型）
func successResponse(format string, args ...any) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: fmt.Sprintf(format, args...)},
			},
		},
	}, nil
}

func formatListWithDescriptions(items []string, descriptions map[string]string) string {
	var lines []string
	for _, item := range items {
		if desc, exists := descriptions[item]; exists {
			lines = append(lines, fmt.Sprintf("- %s: %s", item, desc))
		} else {
			lines = append(lines, fmt.Sprintf("- %s", item))
		}
	}
	return strings.Join(lines, "\n")
}
