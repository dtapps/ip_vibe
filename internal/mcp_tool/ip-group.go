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

func RegisterIPGroupTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getIPGroupList",
		Description: "获取全部IP组",
		InputSchema: utils.AutoSchema(types.GetIPGroupsRequest{}),
	}, getIPGroupList)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getIPGroupDetail",
		Description: "获取IP组",
		InputSchema: utils.AutoSchema(types.GetIPGroupRequest{}),
	}, getIPGroupDetail)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createIPGroup",
		Description: "创建IP组",
		InputSchema: utils.AutoSchema(types.CreateIPGroupRequest{}),
	}, createIPGroup)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateIPGroup",
		Description: "更新IP组",
		InputSchema: utils.AutoSchema(types.UpdateIPGroupRequest{}),
	}, updateIPGroup)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteIPGroup",
		Description: "删除IP组",
		InputSchema: utils.AutoSchema(types.DeleteIPGroupRequest{}),
	}, deleteIPGroup)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getIPGroupDevices",
		Description: "获取IP组的设备列表",
		InputSchema: utils.AutoSchema(types.GetIPGroupDevicesRequest{}),
	}, getIPGroupDevices)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "healthCheckIPGroup",
		Description: "检查IP组是否健康",
		InputSchema: utils.AutoSchema(types.HealthCheckIPGroupRequest{}),
	}, healthCheckIPGroup)
}

// 获取全部IP组
func getIPGroupList(ctx context.Context, req *mcp.CallToolRequest, params *types.GetIPGroupsRequest) (*mcp.CallToolResult, any, error) {

	service := service.NewIPGroupService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetIPGroups(ctx)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 获取IP组
func getIPGroupDetail(ctx context.Context, req *mcp.CallToolRequest, params *types.GetIPGroupRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetIPGroup(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 创建IP组
func createIPGroup(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateIPGroupRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CreateIPGroup(ctx, params)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 更新IP组
func updateIPGroup(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateIPGroupRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateIPGroup(ctx, params)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 删除IP组
func deleteIPGroup(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteIPGroupRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.DeleteIPGroup(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 获取IP组的设备列表
func getIPGroupDevices(ctx context.Context, req *mcp.CallToolRequest, params *types.GetIPGroupDevicesRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetIPGroupDevices(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 检查IP组健康状态
func healthCheckIPGroup(ctx context.Context, req *mcp.CallToolRequest, params *types.HealthCheckIPGroupRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.HealthcheckIPGroup(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}
