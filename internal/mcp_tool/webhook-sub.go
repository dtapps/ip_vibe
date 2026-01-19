package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterWebhookSubTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getWebhookSubList",
		Description: "获取全部通知订阅",
		InputSchema: utils.AutoSchema(types.GetWebhookSubsRequest{}),
	}, getWebhookSubList)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateWebhookSubActive",
		Description: "更新通知订阅启用状态",
		InputSchema: utils.AutoSchema(types.UpdateWebhookSubActiveRequest{}),
	}, updateWebhookSubActive)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteWebhookSub",
		Description: "删除通知订阅",
		InputSchema: utils.AutoSchema(types.DeleteWebhookSubRequest{}),
	}, deleteWebhookSub)
}

// 获取全部通知订阅
func getWebhookSubList(ctx context.Context, req *mcp.CallToolRequest, params *types.GetWebhookSubsRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetWebhookSubs(ctx, params)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 更新通知订阅启用状态
func updateWebhookSubActive(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateWebhookSubActiveRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.UpdateWebhookSubActive(ctx, params.ID, params.IsActive)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(nil, nil)
}

// 删除通知订阅
func deleteWebhookSub(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteWebhookSubRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.DeleteWebhookSub(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(nil, nil)
}
