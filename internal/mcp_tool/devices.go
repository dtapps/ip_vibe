package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterDeviceTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listDevice",
		Description: "列出所有设备",
	}, GetDevices)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getDevice",
		Description: "获取设备",
	}, GetDevice)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createDevice",
		Description: "创建设备",
	}, CreateDevice)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateDevice",
		Description: "更新设备",
	}, UpdateDevice)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteDevice",
		Description: "删除设备",
	}, DeleteDevice)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getDeviceIpHistory",
		Description: "获取设备IP历史记录",
	}, GetDeviceIpHistory)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "healthCheckDevice",
		Description: "检查设备是否健康",
	}, HealthCheckDevice)
}

// 获取全部设备
func GetDevices(ctx context.Context, req *mcp.CallToolRequest, params *types.GetDevicesRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetDevices(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 获取设备
func GetDevice(ctx context.Context, req *mcp.CallToolRequest, params *types.GetDeviceRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetDevice(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 创建设备
func CreateDevice(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateDeviceRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.CreateDevice(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 更新设备
func UpdateDevice(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateDeviceRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateDevice(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 删除设备
func DeleteDevice(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteDeviceRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.DeleteDevice(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 获取设备的IP历史记录
func GetDeviceIpHistory(ctx context.Context, req *mcp.CallToolRequest, params *types.GetDeviceIpHistoryRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetDeviceIpHistory(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 检查设备健康状态
func HealthCheckDevice(ctx context.Context, req *mcp.CallToolRequest, params *types.HealthCheckDeviceRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.HealthcheckDevice(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}
