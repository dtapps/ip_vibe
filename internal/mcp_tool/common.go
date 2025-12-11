package mcp_tool

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func successResponse(data any) (*mcp.CallToolResult, any, error) {
	// 判断可不可以转Json字符串
	text, err := json.Marshal(data)
	if err != nil {
		text = []byte(fmt.Sprintf("%v", data))
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(text),
			},
		},
	}, nil, nil
}

func errorResponse(message string) (*mcp.CallToolResult, any, error) {
	return nil, nil, fmt.Errorf("%s", message)
}
