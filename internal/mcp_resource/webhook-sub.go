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

func RegisterWebhookSubResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "webhookSubList",
		Description: "获取全部通知订阅",
		MIMEType:    MIMETypeJSON,
		URI:         "webhookSub://record/all",
	}, getWebhookSubList)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "webhookSubDetail",
		Description: "获取通知订阅详情",
		MIMEType:    MIMETypeJSON,
		URITemplate: "webhookSub://record/{id}",
	}, getWebhookSubDetail)
}

// 获取全部通知订阅
func getWebhookSubList(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetWebhookSubs(ctx, &types.GetWebhookSubsRequest{})
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(req.Params.URI, response)
}

// 获取通知订阅
func getWebhookSubDetail(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	if err != nil {
		return errorResponseErr(err)
	}
	params := &types.GetWebhookSubRequest{}
	params.ID = cast.ToInt64(param["id"])

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetWebhookSub(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(req.Params.URI, response)
}
