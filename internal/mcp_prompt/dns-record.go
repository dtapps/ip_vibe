package mcp_prompt

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterDNSRecordPrompt(server *mcp.Server) {
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

// DNS记录回源协议说明
func dnsRecordOriginProtocol(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

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
