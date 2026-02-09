package mcp_resource

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cast"
)

func RegisterDNSRecordResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "dnsRecordList",
		Description: "获取全部DNS记录",
		MIMEType:    MIMETypeJSON,
		URI:         "dns://record/all",
	}, getDnsRecordList)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "dnsRecordDetail",
		Description: "获取DNS记录详情",
		MIMEType:    MIMETypeJSON,
		URITemplate: "dns://record/{id}",
	}, getDnsRecordDetail)
}

// 获取全部DNS记录
func getDnsRecordList(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	response, err := service.GetDnsRecords(ctx, &types.GetDnsRecordsRequest{})
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(req.Params.URI, response)
}

// 获取DNS记录
func getDnsRecordDetail(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	if err != nil {
		return errorResponseErr(err)
	}
	params := &types.GetDnsRecordRequest{}
	params.ID = cast.ToInt64(param["id"])

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	response, err := service.GetDnsRecord(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(req.Params.URI, response)
}
