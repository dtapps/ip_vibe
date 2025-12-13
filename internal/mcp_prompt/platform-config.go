package mcp_prompt

import (
	"context"
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

	// 格式化
	var result = formatListWithDescriptions(types, descriptions)

	return successResponse(result)
}
