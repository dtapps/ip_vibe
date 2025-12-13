package mcp_resource

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cast"
)

func RegisterDeviceResource(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "getDeviceList",
		Description: "获取全部设备",
		MIMEType:    MIMETypeJSON,
		URI:         "device://record/all",
	}, getDeviceList)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "getDeviceDetail",
		Description: "获取设备",
		MIMEType:    MIMETypeJSON,
		URITemplate: "device://record/{id}",
	}, getDeviceDetail)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "getDeviceIpHistory",
		Description: "获取设备IP历史记录",
		MIMEType:    MIMETypeJSON,
		URITemplate: "device://record/{id}/ipHistory",
	}, getDeviceIpHistory)
}

// 获取全部设备
func getDeviceList(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

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

	return successResponse(req.Params.URI, response)
}

// 获取设备
func getDeviceDetail(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	if err != nil {
		return errorResponse(err.Error())
	}
	params := &types.GetDeviceRequest{}
	params.ID = cast.ToInt64(param["id"])

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

	return successResponse(req.Params.URI, response)
}

// 获取设备的IP历史记录
func getDeviceIpHistory(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	if err != nil {
		return errorResponse(err.Error())
	}
	params := &types.GetDeviceRequest{}
	params.ID = cast.ToInt64(param["id"])

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

	return successResponse(req.Params.URI, response)
}
