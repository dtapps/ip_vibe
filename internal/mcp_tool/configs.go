package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterPlatformConfigTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listPlatformConfig",
		Description: "列出全部配置",
	}, GetConfigs)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getPlatformConfig",
		Description: "获取配置",
	}, GetConfig)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createPlatformConfig",
		Description: "创建配置",
	}, CreateConfig)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updatePlatformConfig",
		Description: "更新配置",
	}, UpdateConfig)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deletePlatformConfig",
		Description: "删除配置",
	}, DeleteConfig)
}

// 获取全部配置
func GetConfigs(ctx context.Context, req *mcp.CallToolRequest, params *types.GetConfigsRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetConfigs(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 获取配置
func GetConfig(ctx context.Context, req *mcp.CallToolRequest, params *types.GetConfigRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetConfig(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 创建配置
func CreateConfig(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateConfigRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.CreateConfig(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 更新配置
func UpdateConfig(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateConfigRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateConfig(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 删除配置
func DeleteConfig(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteConfigRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.DeleteConfig(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}
