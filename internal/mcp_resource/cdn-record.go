package mcp_resource

import (
	"context"
	"fmt"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cast"
)

func RegisterCDNRecordResource(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "cdnRecordsList",
		Description: "获取全部CDN记录",
		MIMEType:    MIMETypeJSON,
		URI:         "cdn://records/all",
	}, getCdnRecordsList)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "cdnRecordDetail",
		Description: "获取CDN记录详情",
		MIMEType:    MIMETypeJSON,
		URITemplate: "cdn://records/{id}",
	}, getCdnRecordDetail)

}

// 获取全部CDN记录
func getCdnRecordsList(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetCdnRecords(ctx, nil)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(req.Params.URI, response)
}

// 获取CDN记录详情
func getCdnRecordDetail(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	fmt.Printf("param: %+v \n", param)
	if err != nil {
		return errorResponse(err.Error())
	}
	params := &types.GetCdnRecordRequest{}
	params.ID = cast.ToInt64(param["id"])

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetCdnRecord(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}
	return successResponse(req.Params.URI, response)
}
