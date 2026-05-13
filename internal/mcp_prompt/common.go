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

// errorResponseErr 错误响应（支持错误）
func errorResponseErr(err error) (*mcp.GetPromptResult, error) {
	return nil, err
}

// successResponse 成功响应
func successResponse(text string) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: text},
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
