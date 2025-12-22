package mcp_resource

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cast"
)

func RegisterDomainMonitorResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "domainMonitorList",
		Description: "获取全部域名监控",
		MIMEType:    MIMETypeJSON,
		URI:         "domainMonitor://record/all",
	}, getDomainMonitorList)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "domainMonitorDetail",
		Description: "获取域名监控详情",
		MIMEType:    MIMETypeJSON,
		URITemplate: "domainMonitor://record/{id}",
	}, getDomainMonitorDetail)
}

// 获取全部域名监控
func getDomainMonitorList(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	response, err := service.GetDomainMonitors(ctx, &types.GetDomainMonitorsRequest{})
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(req.Params.URI, response)
}

// 获取域名监控
func getDomainMonitorDetail(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	if err != nil {
		return errorResponse(err.Error())
	}
	params := &types.GetDomainMonitorRequest{}
	params.ID = cast.ToInt64(param["id"])

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	response, err := service.GetDomainMonitor(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(req.Params.URI, response)
}
