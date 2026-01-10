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
