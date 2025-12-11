package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterPortForwardTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listPortForward",
		Description: "列出所有端口转发",
	}, GetPortForwards)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getPortForward",
		Description: "获取端口转发",
	}, GetPortForward)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createPortForward",
		Description: "创建端口转发",
	}, CreatePortForward)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updatePortForward",
		Description: "更新端口转发",
	}, UpdatePortForward)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updatePortForwardForward",
		Description: "更新端口转发防火墙状态",
	}, UpdatePortForwardForward)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updatePortForwardActive",
		Description: "更新端口转发启用状态",
	}, UpdatePortForwardActive)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deletePortForward",
		Description: "删除端口转发",
	}, DeletePortForward)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "healthCheckPortForward",
		Description: "检查端口转发是否健康",
	}, HealthCheckPortForward)
}

// 获取全部端口转发
func GetPortForwards(ctx context.Context, req *mcp.CallToolRequest, params *types.GetPortForwardsRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetPortForwards(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 获取端口转发
func GetPortForward(ctx context.Context, req *mcp.CallToolRequest, params *types.GetPortForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetPortForward(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 创建端口转发
func CreatePortForward(ctx context.Context, req *mcp.CallToolRequest, params *types.CreatePortForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.CreatePortForward(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 更新端口转发
func UpdatePortForward(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdatePortForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdatePortForward(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 更新端口转发防火墙状态
func UpdatePortForwardForward(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdatePortForwardForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdatePortForwardForward(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 更新端口转发启用状态
func UpdatePortForwardActive(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdatePortForwardActiveRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdatePortForwardActive(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 删除端口转发
func DeletePortForward(ctx context.Context, req *mcp.CallToolRequest, params *types.DeletePortForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.DeletePortForward(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 检查端口转发健康状态
func HealthCheckPortForward(ctx context.Context, req *mcp.CallToolRequest, params *types.HealthCheckPortForwardRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.HealthcheckPortForward(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}
