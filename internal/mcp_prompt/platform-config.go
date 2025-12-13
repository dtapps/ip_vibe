package mcp_prompt

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterPlatformConfigPrompt(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "webhookSubObjectType",
		Description: "通知订阅订阅对象类型说明",
	}, webhookSubObjectType)
}

// 通知订阅订阅对象类型说明
func webhookSubObjectType(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetWebhookSubscriptionObjTypes

	// 描述
	descriptions := schematype.GetWebhookSubscriptionObjTypeDescriptions

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
