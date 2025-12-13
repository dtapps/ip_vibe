package mcp_prompt

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterCDNRecordPrompt(server *mcp.Server) {
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

	var result string
	for _, recordType := range types {
		if desc, exists := descriptions[recordType]; exists {
			result += fmt.Sprintf("- %s: %s\n", recordType, desc)
		} else {
			result += fmt.Sprintf("- %s\n", recordType)
		}
	}
	return successResponse(result)
}

// CDN记录回源协议说明
func cdnRecordOriginProtocol(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetOriginProtocols

	// 描述
	descriptions := schematype.GetOriginProtocolDescriptions

	var result string
	for _, recordType := range types {
		if desc, exists := descriptions[recordType]; exists {
			result += fmt.Sprintf("- %s: %s\n", recordType, desc)
		} else {
			result += fmt.Sprintf("- %s\n", recordType)
		}
	}
	return successResponse(result)
}
