package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterWebhookTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getWebhookList",
		Description: "获取全部通知",
		InputSchema: utils.AutoSchema(types.GetWebhooksRequest{}),
	}, getWebhookList)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getWebhookDetail",
		Description: "获取通知",
		InputSchema: utils.AutoSchema(types.GetWebhookRequest{}),
	}, getWebhookDetail)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createWebhook",
		Description: "创建通知",
		InputSchema: utils.AutoSchema(types.CreateWebhookRequest{}),
	}, createWebhook)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateWebhook",
		Description: "更新通知",
		InputSchema: utils.AutoSchema(types.UpdateWebhookRequest{}),
	}, updateWebhook)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateWebhookActive",
		Description: "更新通知启用状态",
		InputSchema: utils.AutoSchema(types.UpdateWebhookActiveRequest{}),
	}, updateWebhookActive)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteWebhook",
		Description: "删除通知",
		InputSchema: utils.AutoSchema(types.DeleteWebhookRequest{}),
	}, deleteWebhook)
}

// 获取全部通知
func getWebhookList(ctx context.Context, req *mcp.CallToolRequest, params *types.GetWebhooksRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetWebhooks(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 获取通知
func getWebhookDetail(ctx context.Context, req *mcp.CallToolRequest, params *types.GetWebhookRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetWebhook(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 创建通知
func createWebhook(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateWebhookRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CreateWebhook(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 更新通知
func updateWebhook(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateWebhookRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateWebhook(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 更新通知启用状态
func updateWebhookActive(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateWebhookActiveRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateWebhookActive(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 删除通知
func deleteWebhook(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteWebhookRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.DeleteWebhook(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}
