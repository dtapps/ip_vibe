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
		Name:        "listTask",
		Description: "列出所有任务",
	}, GetTasks)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "getTask",
		Description: "获取任务",
	}, GetTask)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "runTask",
		Description: "运行任务",
	}, RunTask)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "deleteTask",
		Description: "删除任务",
	}, DeleteTask)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "cleanTask",
		Description: "清理任务",
	}, CleanTask)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "startScheduler",
		Description: "启动调度器",
	}, StartScheduler)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "stopScheduler",
		Description: "停止调度器",
	}, StopScheduler)
}

// 获取全部任务
func GetTasks(ctx context.Context, req *mcp.CallToolRequest, params *types.GetTasksRequest) (*mcp.CallToolResult, any, error) {

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
func GetTask(ctx context.Context, req *mcp.CallToolRequest, params *types.GetTaskRequest) (*mcp.CallToolResult, any, error) {

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
func RunTask(ctx context.Context, req *mcp.CallToolRequest, params *types.RunTaskRequest) (*mcp.CallToolResult, any, error) {

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
func DeleteTask(ctx context.Context, req *mcp.CallToolRequest, params *types.DeleteTaskRequest) (*mcp.CallToolResult, any, error) {

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

// 清理任务
func CleanTask(ctx context.Context, req *mcp.CallToolRequest, params *any) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	err := service.CleanTask(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(nil)
}

// 启动调度器
func StartScheduler(ctx context.Context, req *mcp.CallToolRequest, params *any) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	err := service.StartTask(ctx)
	if err != nil {
		return errorResponse(err.Error())
	}

	return successResponse(nil)
}

// 停止调度器
func StopScheduler(ctx context.Context, req *mcp.CallToolRequest, params *any) (*mcp.CallToolResult, any, error) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	service.StopTask(ctx)
	return successResponse(nil)
}
