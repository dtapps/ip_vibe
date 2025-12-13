package mcp_prompt

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterDomainMonitorPrompt(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "domainMonitorCheckProtocol",
		Description: "域名监控检查类型说明",
	}, domainMonitorCheckProtocol)
}

// 域名监控检查类型说明
func domainMonitorCheckProtocol(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetDomainMonitorCheckProtocols

	// 描述
	descriptions := schematype.GetDomainMonitorCheckProtocolDescriptions

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
