package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterWebhookTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listWebhook",
		Description: "列出所有通知",
	}, GetWebhooks)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getWebhook",
		Description: "获取通知",
	}, GetWebhook)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createWebhook",
		Description: "创建通知",
	}, CreateWebhook)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateWebhook",
		Description: "更新通知",
	}, UpdateWebhook)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateWebhookActive",
		Description: "更新通知启用状态",
	}, UpdateWebhookActive)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteWebhook",
		Description: "删除通知",
	}, DeleteWebhook)
}

// 获取全部通知
func GetWebhooks(ctx context.Context, req *mcp.CallToolRequest, params *types.GetWebhooksRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetWebhooks(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 获取通知
func GetWebhook(ctx context.Context, req *mcp.CallToolRequest, params *types.GetWebhookRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetWebhook(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 创建通知
func CreateWebhook(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateWebhookRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.CreateWebhook(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 更新通知
func UpdateWebhook(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateWebhookRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateWebhook(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 更新通知启用状态
func UpdateWebhookActive(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateWebhookActiveRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateWebhookActive(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 删除通知
func DeleteWebhook(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteWebhookRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.DeleteWebhook(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}
