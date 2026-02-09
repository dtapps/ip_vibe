package mcp_resource

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cast"
)

func RegisterWebhookResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "webhookList",
		Description: "获取全部通知",
		MIMEType:    MIMETypeJSON,
		URI:         "webhook://record/all",
	}, getWebhookList)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "webhookDetail",
		Description: "获取通知详情",
		MIMEType:    MIMETypeJSON,
		URITemplate: "webhook://record/{id}",
	}, getWebhookDetail)
}

// 获取全部通知
func getWebhookList(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 创建服务
	service := service.NewWebhookService(
		ctx,
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	response, err := service.GetWebhooks(ctx, &types.GetWebhooksRequest{})
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(req.Params.URI, response)
}

// 获取通知
func getWebhookDetail(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	if err != nil {
		return errorResponseErr(err)
	}
	params := &types.GetWebhookRequest{}
	params.ID = cast.ToInt64(param["id"])

	// 创建服务
	service := service.NewWebhookService(
		ctx,
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	response, err := service.GetWebhook(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(req.Params.URI, response)
}
