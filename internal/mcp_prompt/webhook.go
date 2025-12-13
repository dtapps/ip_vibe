package mcp_prompt

import (
	"context"
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterWebhookPrompt(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "webhookWebhookPlatform",
		Description: "通知平台通知类型说明",
	}, webhookWebhookPlatform)
}

// 通知平台通知类型说明
func webhookWebhookPlatform(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetWebhookPlatforms

	// 描述
	descriptions := schematype.GetWebhookPlatformDescriptions

	// 格式化
	var result = formatListWithDescriptions(types, descriptions)

	return successResponse(result)
}
