package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterPortForwardTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getPortForwardList",
		Description: "获取全部端口转发",
		InputSchema: utils.AutoSchema(types.GetPortForwardsRequest{}),
	}, getPortForwardList)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getPortForwardDetail",
		Description: "获取端口转发",
		InputSchema: utils.AutoSchema(types.GetPortForwardRequest{}),
	}, getPortForwardDetail)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createPortForward",
		Description: "创建端口转发",
		InputSchema: utils.AutoSchema(types.CreatePortForwardRequest{}),
	}, createPortForward)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updatePortForward",
		Description: "更新端口转发",
		InputSchema: utils.AutoSchema(types.UpdatePortForwardRequest{}),
	}, updatePortForward)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updatePortForwardForward",
		Description: "更新端口转发防火墙状态",
		InputSchema: utils.AutoSchema(types.UpdatePortForwardForwardRequest{}),
	}, updatePortForwardForward)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updatePortForwardActive",
		Description: "更新端口转发启用状态",
		InputSchema: utils.AutoSchema(types.UpdatePortForwardActiveRequest{}),
	}, updatePortForwardActive)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deletePortForward",
		Description: "删除端口转发",
		InputSchema: utils.AutoSchema(types.DeletePortForwardRequest{}),
	}, deletePortForward)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "healthCheckPortForward",
		Description: "检查端口转发是否健康",
		InputSchema: utils.AutoSchema(types.HealthCheckPortForwardRequest{}),
	}, healthCheckPortForward)
}

// 获取全部端口转发
func getPortForwardList(ctx context.Context, req *mcp.CallToolRequest, params *types.GetPortForwardsRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,

		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetPortForwards(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 获取端口转发
func getPortForwardDetail(ctx context.Context, req *mcp.CallToolRequest, params *types.GetPortForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,

		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetPortForward(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 创建端口转发
func createPortForward(ctx context.Context, req *mcp.CallToolRequest, params *types.CreatePortForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,

		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CreatePortForward(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 更新端口转发
func updatePortForward(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdatePortForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,

		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdatePortForward(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 更新端口转发防火墙状态
func updatePortForwardForward(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdatePortForwardForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,

		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdatePortForwardForward(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 更新端口转发启用状态
func updatePortForwardActive(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdatePortForwardActiveRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,

		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdatePortForwardActive(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 删除端口转发
func deletePortForward(ctx context.Context, req *mcp.CallToolRequest, params *types.DeletePortForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,

		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.DeletePortForward(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 检查端口转发健康状态
func healthCheckPortForward(ctx context.Context, req *mcp.CallToolRequest, params *types.HealthCheckPortForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,

		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.HealthcheckPortForward(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}
