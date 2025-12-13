package mcp_prompt

import (
	"context"
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterDNSRecordPrompts(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "dnsRecordType",
		Description: "DNS记录类型说明",
	}, dnsRecordType)
	server.AddPrompt(&mcp.Prompt{
		Name:        "dnsRecordOriginProtocol",
		Description: "DNS记录回源协议说明",
	}, dnsRecordOriginProtocol)
}

// DNS记录类型说明
func dnsRecordType(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetDNSRecordTypes

	// 描述
	descriptions := schematype.GetDNSRecordTypeDescriptions

	// 格式化
	var result = formatListWithDescriptions(types, descriptions)

	return successResponse(result)
}

// DNS记录回源协议说明
func dnsRecordOriginProtocol(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetOriginProtocols

	// 描述
	descriptions := schematype.GetOriginProtocolDescriptions

	// 格式化
	var result = formatListWithDescriptions(types, descriptions)

	return successResponse(result)
}
