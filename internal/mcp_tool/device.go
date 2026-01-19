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

func RegisterDeviceTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getDeviceList",
		Description: "获取全部设备",
		InputSchema: utils.AutoSchema(types.GetDevicesRequest{}),
	}, getDeviceList)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getDeviceDetail",
		Description: "获取设备",
		InputSchema: utils.AutoSchema(types.GetDeviceRequest{}),
	}, getDeviceDetail)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createDevice",
		Description: "创建设备",
		InputSchema: utils.AutoSchema(types.CreateDeviceRequest{}),
	}, createDevice)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateDevice",
		Description: "更新设备",
		InputSchema: utils.AutoSchema(types.UpdateDeviceRequest{}),
	}, updateDevice)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteDevice",
		Description: "删除设备",
		InputSchema: utils.AutoSchema(types.DeleteDeviceRequest{}),
	}, deleteDevice)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getDeviceIpHistory",
		Description: "获取设备IP历史记录",
		InputSchema: utils.AutoSchema(types.GetDeviceIpHistoryRequest{}),
	}, getDeviceIpHistory)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "healthCheckDevice",
		Description: "检查设备是否健康",
		InputSchema: utils.AutoSchema(types.HealthCheckDeviceRequest{}),
	}, healthCheckDevice)
}

// 获取全部设备
func getDeviceList(ctx context.Context, req *mcp.CallToolRequest, params *types.GetDevicesRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDevices(ctx)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 获取设备
func getDeviceDetail(ctx context.Context, req *mcp.CallToolRequest, params *types.GetDeviceRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDevice(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 创建设备
func createDevice(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateDeviceRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CreateDevice(ctx, params)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 更新设备
func updateDevice(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateDeviceRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateDevice(ctx, params)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 删除设备
func deleteDevice(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteDeviceRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.DeleteDevice(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(nil, nil)
}

// 获取设备的IP历史记录
func getDeviceIpHistory(ctx context.Context, req *mcp.CallToolRequest, params *types.GetDeviceIpHistoryRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDeviceIpHistory(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 检查设备健康状态
func healthCheckDevice(ctx context.Context, req *mcp.CallToolRequest, params *types.HealthCheckDeviceRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.HealthcheckDevice(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}
