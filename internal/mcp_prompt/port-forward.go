package mcp_prompt

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterPortForwardPrompt(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "portForwardProtocols",
		Description: "端口转发协议列表说明",
	}, portForwardProtocols)
}

// portForwardProtocols
func portForwardProtocols(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetDevicePortForwardProtocols

	// 描述
	descriptions := schematype.GetDevicePortForwardProtocolDescriptions

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
