package mcp_prompt

import (
	"context"
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterPortForwardPrompts(server *mcp.Server) {
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

	// 格式化
	var result = formatListWithDescriptions(types, descriptions)

	return successResponse(result)
}
