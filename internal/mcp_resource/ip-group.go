package mcp_resource

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cast"
)

func RegisterIPGroupResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "ipGroupList",
		Description: "获取全部IP组",
		MIMEType:    MIMETypeJSON,
		URI:         "ipGroup://record/all",
	}, getIPGroupList)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "ipGroupDetail",
		Description: "获取IP组详情",
		MIMEType:    MIMETypeJSON,
		URITemplate: "ipGroup://record/{id}",
	}, getIPGroupDetail)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "ipGroupDevices",
		Description: "获取IP组的设备列表",
		MIMEType:    MIMETypeJSON,
		URITemplate: "ipGroup://record/{id}/devices",
	}, getIPGroupDevices)
}

// 获取全部IP组
func getIPGroupList(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetIPGroups(ctx)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(req.Params.URI, response)
}

// 获取IP组
func getIPGroupDetail(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	if err != nil {
		return errorResponseErr(err)
	}
	params := &types.GetIPGroupRequest{}
	params.ID = cast.ToInt64(param["id"])

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetIPGroup(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(req.Params.URI, response)
}

// 获取IP组的设备列表
func getIPGroupDevices(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	if err != nil {
		return errorResponseErr(err)
	}
	params := &types.GetIPGroupRequest{}
	params.ID = cast.ToInt64(param["id"])

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetIPGroupDevices(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(req.Params.URI, response)
}
