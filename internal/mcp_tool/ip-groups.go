package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterIPGroupTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listIPGroup",
		Description: "列出所有IP组",
	}, GetIPGroups)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getIPGroup",
		Description: "获取IP组",
	}, GetIPGroup)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createIPGroup",
		Description: "创建IP组",
	}, CreateIPGroup)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateIPGroup",
		Description: "更新IP组",
	}, UpdateIPGroup)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteIPGroup",
		Description: "删除IP组",
	}, DeleteIPGroup)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getIPGroupDevices",
		Description: "获取IP组的设备列表",
	}, GetIPGroupDevices)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "healthCheckIPGroup",
		Description: "检查IP组是否健康",
	}, HealthCheckIPGroup)
}

// 获取全部IP组
func GetIPGroups(ctx context.Context, req *mcp.CallToolRequest, params *types.GetIPGroupsRequest) (*mcp.CallToolResult, any, error) {

	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetIPGroups(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 获取IP组
func GetIPGroup(ctx context.Context, req *mcp.CallToolRequest, params *types.GetIPGroupRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetIPGroup(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 创建IP组
func CreateIPGroup(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateIPGroupRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.CreateIPGroup(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 更新IP组
func UpdateIPGroup(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateIPGroupRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateIPGroup(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 删除IP组
func DeleteIPGroup(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteIPGroupRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.DeleteIPGroup(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 获取IP组的设备列表
func GetIPGroupDevices(ctx context.Context, req *mcp.CallToolRequest, params *types.GetIPGroupDevicesRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetIPGroupDevices(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 检查IP组健康状态
func HealthCheckIPGroup(ctx context.Context, req *mcp.CallToolRequest, params *types.HealthCheckIPGroupRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.HealthcheckIPGroup(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}
