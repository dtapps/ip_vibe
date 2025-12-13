package mcp_resource

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterTaskResource(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "getTaskList",
		Description: "获取全部任务",
		MIMEType:    MIMETypeJSON,
		URI:         "task://record/all",
	}, getTaskList)
	AddResourceTemplateEx(server, &mcp.ResourceTemplate{
		Name:        "getTaskDetail",
		Description: "获取任务",
		MIMEType:    MIMETypeJSON,
		URITemplate: "task://record/{id}",
	}, getTaskDetail)
}

// 获取全部任务
func getTaskList(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

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

	return successResponse(req.Params.URI, response)
}

// 获取任务
func getTaskDetail(ctx context.Context, tpl *mcp.ResourceTemplate, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	// 请求参数
	param, err := ExtractURIParams(tpl, req.Params.URI)
	if err != nil {
		return errorResponse(err.Error())
	}
	params := &types.GetTaskRequest{}
	params.ID = param["id"]

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

	return successResponse(req.Params.URI, response)
}
