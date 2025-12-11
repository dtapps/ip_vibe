package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterWebhookSubTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listWebhookSub",
		Description: "列出所有通知订阅",
	}, GetWebhooks)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getWebhookSub",
		Description: "获取通知订阅",
	}, GetWebhook)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createWebhookSub",
		Description: "创建通知订阅",
	}, CreateWebhook)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateWebhookSub",
		Description: "更新通知订阅",
	}, UpdateWebhook)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateWebhookSubActive",
		Description: "更新通知订阅启用状态",
	}, UpdateWebhookActive)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteWebhookSub",
		Description: "删除通知订阅",
	}, DeleteWebhook)
}

// 获取全部通知订阅
func GetWebhookSubs(ctx context.Context, req *mcp.CallToolRequest, params *types.GetWebhookSubsRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetWebhookSubs(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 获取通知订阅
func GetWebhookSub(ctx context.Context, req *mcp.CallToolRequest, params *types.GetWebhookSubRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetWebhookSub(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 创建通知订阅
func CreateWebhookSub(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateWebhookSubRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.CreateWebhookSub(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 更新通知订阅
func UpdateWebhookSub(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateWebhookSubRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateWebhookSub(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 更新通知订阅启用状态
func UpdateWebhookSubActive(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateWebhookSubActiveRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateWebhookSubActive(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 删除通知订阅
func DeleteWebhookSub(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteWebhookSubRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.DeleteWebhookSub(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}
