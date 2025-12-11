package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterCDNRecordTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "listCDNRecord",
		Description: "列出域名对应的CDN记录",
	}, GetCdnRecords)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getCDNRecord",
		Description: "获取CDN记录",
	}, GetCdnRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createCDNRecord",
		Description: "创建CDN记录",
	}, CreateCdnRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateCDNRecord",
		Description: "更新CDN记录",
	}, UpdateCdnRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateCDNRecordActive",
		Description: "更新CDN记录启用状态",
	}, UpdateCdnRecordActive)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteCDNRecord",
		Description: "删除CDN记录",
	}, DeleteCdnRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "healthCheckCDNRecord",
		Description: "检查CDN记录是否健康",
	}, HealthCheckCdnRecord)
}

// 获取全部CDN记录
func GetCdnRecords(ctx context.Context, req *mcp.CallToolRequest, params *types.GetCdnRecordsRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetCdnRecords(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 获取CDN记录
func GetCdnRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.GetCdnRecordRequest) (*mcp.CallToolResult, any, error) {

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
	return successResponse(response)
}

// 创建CDN记录
func CreateCdnRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateCdnRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.CreateCdnRecord(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}
	return successResponse(response)
}

// 更新CDN记录
func UpdateCdnRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateCdnRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateCdnRecord(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}
	return successResponse(response)
}

// 更新CDN记录启用状态
func UpdateCdnRecordActive(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateCdnRecordActiveRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateCdnRecordActive(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}
	return successResponse(response)
}

// 删除CDN记录
func DeleteCdnRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteCdnRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.DeleteCdnRecord(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}
	return successResponse(response)
}

// 检查CDN记录健康状态
func HealthCheckCdnRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.HealthCheckCdnRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.HealthcheckCdnRecord(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}
	return successResponse(response)
}
