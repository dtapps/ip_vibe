package mcp_prompt

import (
	"context"
	"fmt"
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

// 平台配置平台类型说明
func platformConfigPlatform(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {

	// 类型
	types := schematype.GetConfigPlatforms

	// 描述
	descriptions := schematype.GetConfigPlatformDescriptions

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
