package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterDNSRecordTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getDNSRecordList",
		Description: "获取全部DNS记录",
		InputSchema: utils.AutoSchema(types.GetDnsRecordsRequest{}),
	}, getDnsRecordList)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getDNSRecordDetail",
		Description: "获取DNS记录",
		InputSchema: utils.AutoSchema(types.GetDnsRecordRequest{}),
	}, getDnsRecordDetail)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createDNSRecord",
		Description: "创建DNS记录",
		InputSchema: utils.AutoSchema(types.CreateDnsRecordRequest{}),
	}, createDnsRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateDNSRecord",
		Description: "更新DNS记录",
		InputSchema: utils.AutoSchema(types.UpdateDnsRecordRequest{}),
	}, updateDnsRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateDNSRecordActive",
		Description: "更新DNS记录启用状态",
		InputSchema: utils.AutoSchema(types.UpdateDnsRecordActiveRequest{}),
	}, updateDnsRecordActive)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteDNSRecord",
		Description: "删除DNS记录",
		InputSchema: utils.AutoSchema(types.DeleteDnsRecordRequest{}),
	}, deleteDnsRecord)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "healthCheckDNSRecord",
		Description: "检查DNS记录是否健康",
		InputSchema: utils.AutoSchema(types.HealthCheckDnsRecordRequest{}),
	}, healthCheckDnsRecord)
}

// 获取全部DNS记录
func getDnsRecordList(ctx context.Context, req *mcp.CallToolRequest, params *types.GetDnsRecordsRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDNSRecordService(
		ctx,
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	response, err := service.GetDnsRecords(ctx, params)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}

// 获取DNS记录
func getDnsRecordDetail(ctx context.Context, req *mcp.CallToolRequest, params *types.GetDnsRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDNSRecordService(
		ctx,
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

	return successResponse(response, response)
}

// 创建DNS记录
func createDnsRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateDnsRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDNSRecordService(
		ctx,
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	err := service.CreateDnsRecord(ctx, params)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(nil, nil)
}

// 更新DNS记录
func updateDnsRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateDnsRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDNSRecordService(
		ctx,
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	err := service.UpdateDnsRecord(ctx, params)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(nil, nil)
}

// 更新DNS记录启用状态
func updateDnsRecordActive(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateDnsRecordActiveRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDNSRecordService(
		ctx,
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	err := service.UpdateDnsRecordActive(ctx, params.ID, params.IsActive)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(nil, nil)
}

// 删除DNS记录
func deleteDnsRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteDnsRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDNSRecordService(
		ctx,
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	err := service.DeleteDnsRecord(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(nil, nil)
}

// 检查DNS记录健康状态
func healthCheckDnsRecord(ctx context.Context, req *mcp.CallToolRequest, params *types.HealthCheckDnsRecordRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDNSRecordService(
		ctx,
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaMcpSlog,
		global.GvaHttpLog,
	)

	// 调用
	response, err := service.HealthcheckDnsRecord(ctx, params.ID)
	if err != nil {
		return errorResponseErr(err)
	}

	return successResponse(response, response)
}
