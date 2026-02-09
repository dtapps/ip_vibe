package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterCDNRecordTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getCdnRecordList",
		Description: "获取全部CDN记录",
		InputSchema: utils.AutoSchema(types.GetCdnRecordsRequest{}),
	}, getCdnRecordList)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getCdnRecordDetail",
		Description: "获取CDN记录详情",
		InputSchema: utils.AutoSchema(types.GetCdnRecordRequest{}),
	}, getCdnRecordDetail)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createCDNRecord",
		Description: "创建CDN记录",
		InputSchema: utils.AutoSchema(types.CreateCdnRecordRequest{}),
	}, createCdnRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateCDNRecord",
		Description: "更新CDN记录",
		InputSchema: utils.AutoSchema(types.UpdateCdnRecordRequest{}),
	}, updateCdnRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateCDNRecordActive",
		Description: "更新CDN记录启用状态",
		InputSchema: utils.AutoSchema(types.UpdateCdnRecordActiveRequest{}),
	}, updateCdnRecordActive)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteCDNRecord",
		Description: "删除CDN记录",
		InputSchema: utils.AutoSchema(types.DeleteCdnRecordRequest{}),
	}, deleteCdnRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "healthCheckCDNRecord",
		Description: "检查CDN记录是否健康",
		InputSchema: utils.AutoSchema(types.HealthCheckCdnRecordRequest{}),
	}, healthCheckCdnRecord)
}

// 获取全部CDN记录
func getCdnRecordList(ctx context.Context, req *mcp.CallToolRequest, params *types.GetCdnRecordsRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	response, err := service.GetCdnRecords(ctx, params)
	if err != nil {
		return errorResponseErr(err)
	}
	return successResponse(response, response)
}

// 获取CDN记录详情
func getCdnRecordDetail(ctx context.Context, req *mcp.CallToolRequest, params *types.GetCdnRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	response, err := service.GetCdnRecord(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}
	return successResponse(response, response)
}

// 创建CDN记录
func createCdnRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateCdnRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	err := service.CreateCdnRecord(ctx, params)
	if err != nil {
		return errorResponseErr(err)
	}
	return successResponse(nil, nil)
}

// 更新CDN记录
func updateCdnRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateCdnRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	err := service.UpdateCdnRecord(ctx, params)
	if err != nil {
		return errorResponseErr(err)
	}
	return successResponse(nil, nil)
}

// 更新CDN记录启用状态
func updateCdnRecordActive(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateCdnRecordActiveRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	err := service.UpdateCdnRecordActive(ctx, params.ID, params.IsActive)
	if err != nil {
		return errorResponseErr(err)
	}
	return successResponse(nil, nil)
}

// 删除CDN记录
func deleteCdnRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteCdnRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	err := service.DeleteCdnRecord(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(nil, nil)
}

// 检查CDN记录健康状态
func healthCheckCdnRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.HealthCheckCdnRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	response, err := service.HealthcheckCdnRecord(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}
	return successResponse(response, response)
}
