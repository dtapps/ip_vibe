package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterWebhookSubTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getWebhookSubList",
		Description: "获取全部通知订阅",
	}, getWebhookSubList)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getWebhookSubDetail",
		Description: "获取通知订阅",
	}, getWebhookSubDetail)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createWebhookSub",
		Description: "创建通知订阅",
	}, createWebhookSub)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateWebhookSub",
		Description: "更新通知订阅",
	}, updateWebhookSub)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateWebhookSubActive",
		Description: "更新通知订阅启用状态",
	}, updateWebhookSubActive)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteWebhookSub",
		Description: "删除通知订阅",
	}, deleteWebhookSub)
}

// 获取全部通知订阅
func getWebhookSubList(ctx context.Context, req *mcp.CallToolRequest, params *types.GetWebhookSubsRequest) (*mcp.CallToolResult, any, error) {

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

	return successResponse(response, response)
}

// 获取通知订阅
func getWebhookSubDetail(ctx context.Context, req *mcp.CallToolRequest, params *types.GetWebhookSubRequest) (*mcp.CallToolResult, any, error) {

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

	return successResponse(response, response)
}

// 创建通知订阅
func createWebhookSub(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateWebhookSubRequest) (*mcp.CallToolResult, any, error) {

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

	return successResponse(response, response)
}

// 更新通知订阅
func updateWebhookSub(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateWebhookSubRequest) (*mcp.CallToolResult, any, error) {

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

	return successResponse(response, response)
}

// 更新通知订阅启用状态
func updateWebhookSubActive(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateWebhookSubActiveRequest) (*mcp.CallToolResult, any, error) {

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

	return successResponse(response, response)
}

// 删除通知订阅
func deleteWebhookSub(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteWebhookSubRequest) (*mcp.CallToolResult, any, error) {

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

	return successResponse(response, response)
}
