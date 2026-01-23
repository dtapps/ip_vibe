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

func RegisterPortForwardResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "portForwardList",
		Description: "获取全部端口转发",
		MIMEType:    MIMETypeJSON,
		URI:         "portForward://record/all",
	}, getPortForwardList)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "portForwardDetail",
		Description: "获取端口转发详情",
		MIMEType:    MIMETypeJSON,
		URITemplate: "portForward://record/{id}",
	}, getPortForwardDetail)
}

// 获取全部端口转发
func getPortForwardList(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetPortForwards(ctx, &types.GetPortForwardsRequest{})
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(req.Params.URI, response)
}

// 获取端口转发
func getPortForwardDetail(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	if err != nil {
		return errorResponseErr(err)
	}
	params := &types.GetPortForwardRequest{}
	params.ID = cast.ToInt64(param["id"])

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetPortForward(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(req.Params.URI, response)
}
