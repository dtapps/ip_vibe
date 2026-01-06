package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterPlatformConfigTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getPlatformConfigList",
		Description: "获取全部配置",
		InputSchema: utils.AutoSchema(types.GetPlatformConfigsRequest{}),
	}, getPlatformConfigList)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getPlatformConfigDetail",
		Description: "获取配置",
		InputSchema: utils.AutoSchema(types.GetPlatformConfigRequest{}),
	}, getPlatformConfigDetail)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createPlatformConfig",
		Description: "创建配置",
		InputSchema: utils.AutoSchema(types.CreatePlatformConfigRequest{}),
	}, createPlatformConfig)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updatePlatformConfig",
		Description: "更新配置",
		InputSchema: utils.AutoSchema(types.UpdatePlatformConfigRequest{}),
	}, updatePlatformConfig)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deletePlatformConfig",
		Description: "删除配置",
		InputSchema: utils.AutoSchema(types.DeletePlatformConfigRequest{}),
	}, deletePlatformConfig)
}

// 获取全部配置
func getPlatformConfigList(ctx context.Context, req *mcp.CallToolRequest, params *types.GetPlatformConfigsRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetConfigs(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 获取配置
func getPlatformConfigDetail(ctx context.Context, req *mcp.CallToolRequest, params *types.GetPlatformConfigRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetConfig(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 创建配置
func createPlatformConfig(ctx context.Context, req *mcp.CallToolRequest, params *types.CreatePlatformConfigRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CreateConfig(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 更新配置
func updatePlatformConfig(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdatePlatformConfigRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateConfig(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 删除配置
func deletePlatformConfig(ctx context.Context, req *mcp.CallToolRequest, params *types.DeletePlatformConfigRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.DeleteConfig(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}
