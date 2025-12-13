package mcp_prompt

import (
	"context"
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterCDNRecordPrompts(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "cdnRecordType",
		Description: "CDN记录类型说明",
	}, cdnRecordType)
	server.AddPrompt(&mcp.Prompt{
		Name:        "cdnRecordOriginProtocol",
		Description: "CDN记录回源协议说明",
	}, cdnRecordOriginProtocol)
}

// CDN记录类型说明
func cdnRecordType(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetCDNRecordTypes

	// 描述
	descriptions := schematype.GetCDNRecordTypeDescriptions

	// 格式化
	var result = formatListWithDescriptions(types, descriptions)

	return successResponse(result)
}

// CDN记录回源协议说明
func cdnRecordOriginProtocol(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetOriginProtocols

	// 描述
	descriptions := schematype.GetOriginProtocolDescriptions

	// 格式化
	var result = formatListWithDescriptions(types, descriptions)

	return successResponse(result)
}
