package mcp_prompt

import (
	"context"
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterWebhookSubPrompt(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "platformConfigCapability",
		Description: "平台配置服务能力说明",
	}, platformConfigCapability)
	server.AddPrompt(&mcp.Prompt{
		Name:        "platformConfigPlatform",
		Description: "平台配置平台类型说明",
	}, platformConfigPlatform)
}

// 平台配置服务能力说明
func platformConfigCapability(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetConfigCapabilitys

	// 描述
	descriptions := schematype.GetConfigCapabilityDescriptions

	// 格式化
	var result = formatListWithDescriptions(types, descriptions)

	return successResponse(result)
}

// 平台配置平台类型说明
func platformConfigPlatform(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetConfigPlatforms

	// 描述
	descriptions := schematype.GetConfigPlatformDescriptions

	// 格式化
	var result = formatListWithDescriptions(types, descriptions)

	return successResponse(result)
}
