package mcp_tool

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterTaskTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getTaskList",
		Description: "获取全部任务",
	}, getTaskList)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getTaskDetail",
		Description: "获取任务",
	}, getTaskDetail)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "runTask",
		Description: "运行任务",
	}, runTask)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteTask",
		Description: "删除任务",
	}, deleteTask)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "cleanAllTask",
		Description: "清理全部任务",
	}, cleanAllTask)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "startScheduler",
		Description: "启动调度器",
	}, startScheduler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "stopScheduler",
		Description: "停止调度器",
	}, stopScheduler)
}

// 获取全部任务
func getTaskList(ctx context.Context, req *mcp.CallToolRequest, params *types.GetTasksRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	response, err := service.GetTasks(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 获取任务
func getTaskDetail(ctx context.Context, req *mcp.CallToolRequest, params *types.GetTaskRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	response, err := service.GetTask(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(response)
}

// 运行任务
func runTask(ctx context.Context, req *mcp.CallToolRequest, params *types.RunTaskRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	err := service.RunTask(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(nil)
}

// 删除任务
func deleteTask(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteTaskRequest) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	err := service.DeleteTask(ctx, params.ID)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(nil)
}

// 清理全部任务
func cleanAllTask(ctx context.Context, req *mcp.CallToolRequest, params *any) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	err := service.CleanAllTask(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(nil)
}

// 启动调度器
func startScheduler(ctx context.Context, req *mcp.CallToolRequest, params *any) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	err := service.StartScheduler(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(nil)
}

// 停止调度器
func stopScheduler(ctx context.Context, req *mcp.CallToolRequest, params *any) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	service.StopScheduler(ctx)
	return successResponse(nil)
}
