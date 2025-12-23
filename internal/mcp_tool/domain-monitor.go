package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterDomainMonitorTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getDomainMonitorList",
		Description: "获取全部域名监控",
		InputSchema: utils.AutoSchema(types.GetDomainMonitorsRequest{}),
	}, getDomainMonitorList)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getDomainMonitorDetail",
		Description: "获取域名监控",
		InputSchema: utils.AutoSchema(types.GetDomainMonitorRequest{}),
	}, getDomainMonitorDetail)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "createDomainMonitor",
		Description: "创建域名监控",
		InputSchema: utils.AutoSchema(types.CreateDomainMonitorRequest{}),
	}, createDomainMonitor)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateDomainMonitor",
		Description: "更新域名监控",
		InputSchema: utils.AutoSchema(types.UpdateDomainMonitorRequest{}),
	}, updateDomainMonitor)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "updateDomainMonitorActive",
		Description: "更新域名监控启用状态",
		InputSchema: utils.AutoSchema(types.UpdateDomainMonitorActiveRequest{}),
	}, updateDomainMonitorActive)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteDomainMonitor",
		Description: "删除域名监控",
		InputSchema: utils.AutoSchema(types.DeleteDomainMonitorRequest{}),
	}, deleteDomainMonitor)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "checkDomainMonitor",
		Description: "检查域名监控",
		InputSchema: utils.AutoSchema(types.CheckDomainMonitorRequest{}),
	}, checkDomainMonitor)
}

// 获取全部域名监控
func getDomainMonitorList(ctx context.Context, req *mcp.CallToolRequest, params *types.GetDomainMonitorsRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
		global.GvaGocronSlog.WithLogger(),
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDomainMonitors(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 获取域名监控
func getDomainMonitorDetail(ctx context.Context, req *mcp.CallToolRequest, params *types.GetDomainMonitorRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
		global.GvaGocronSlog.WithLogger(),
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDomainMonitor(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 创建域名监控
func createDomainMonitor(ctx context.Context, req *mcp.CallToolRequest, params *types.CreateDomainMonitorRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
		global.GvaGocronSlog.WithLogger(),
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CreateDomainMonitor(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 更新域名监控
func updateDomainMonitor(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateDomainMonitorRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
		global.GvaGocronSlog.WithLogger(),
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateDomainMonitor(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 更新域名监控启用状态
func updateDomainMonitorActive(ctx context.Context, req *mcp.CallToolRequest, params *types.UpdateDomainMonitorActiveRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
		global.GvaGocronSlog.WithLogger(),
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateDomainMonitorActive(ctx, params)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 删除域名监控
func deleteDomainMonitor(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteDomainMonitorRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
		global.GvaGocronSlog.WithLogger(),
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.DeleteDomainMonitor(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}

// 检查域名监控
func checkDomainMonitor(ctx context.Context, req *mcp.CallToolRequest, params *types.CheckDomainMonitorRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
		global.GvaGocronSlog.WithLogger(),
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CheckDomainMonitor(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response, response)
}
