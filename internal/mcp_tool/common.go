package mcp_tool

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// errorResponse 错误响应（支持格式化）
func errorResponse(format string, args ...any) (*mcp.CallToolResult, any, error) {
	return nil, nil, fmt.Errorf(format, args...)
}

// errorResponseErr 错误响应（支持错误）
func errorResponseErr(err error) (*mcp.CallToolResult, any, error) {
	return nil, nil, err
}

// successResponse 成功响应（自动处理多种数据类型）
func successResponse(data any, response any) (*mcp.CallToolResult, any, error) {
	var content string

	switch v := data.(type) {
	case string:
		content = v
	case []byte:
		content = string(v)
	case json.RawMessage:
		content = string(v)
	default:
		// 尝试JSON序列化
		if jsonData, err := json.Marshal(data); err == nil {
			content = string(jsonData)
		} else {
			content = fmt.Sprintf("%v", data)
		}
		// if jsonData, err := json.MarshalIndent(data, "", "  "); err == nil {
		// 	content = string(jsonData)
		// } else {
		// 	content = fmt.Sprintf("%v", data)
		// }
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: content,
			},
		},
	}, response, nil
}
