package mcp_resource

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cast"
)

func RegisterPlatformConfigResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "platformConfigList",
		Description: "获取全部配置",
		MIMEType:    MIMETypeJSON,
		URI:         "platformConfig://record/all",
	}, getPlatformConfigList)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "platformConfigDetail",
		Description: "获取配置详情",
		MIMEType:    MIMETypeJSON,
		URITemplate: "platformConfig://record/{id}",
	}, getPlatformConfigDetail)
}

// 获取全部配置
func getPlatformConfigList(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

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

	return successResponse(req.Params.URI, response)
}

// 获取配置
func getPlatformConfigDetail(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	if err != nil {
		return errorResponse(err.Error())
	}
	params := &types.GetPlatformConfigRequest{}
	params.ID = cast.ToInt64(param["id"])

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

	return successResponse(req.Params.URI, response)
}
