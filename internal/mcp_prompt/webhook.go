package mcp_prompt

import (
	"context"
	"fmt"
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
