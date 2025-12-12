package handlers

import (
	"net/http"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// 获取全部任务
//
//	@Description	获取全部任务
//	@Summary		获取全部任务
//	@Tags			任务
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	APIResponse{data=[]types.JobData}	"成功响应"
//	@Failure		404	{object}	APIResponse{}						"错误响应"
//	@Router			/tasks [get]
func GetTasks(c *gin.Context) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	response, err := service.GetTasks(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取任务详情
//
//	@Description	获取任务详情
//	@Summary		获取任务详情
//	@Tags			任务
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		string							true	"任务 ID"
//	@Success		200	{object}	APIResponse{data=types.JobData}	"成功响应"
//	@Failure		400	{object}	APIResponse{}					"参数错误"
//	@Failure		404	{object}	APIResponse{}					"错误响应"
//	@Router			/tasks/{id} [get]
func GetTaskDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetTaskRequest
	requestParams.ID = c.Param("id")
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	response, err := service.GetTask(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 运行任务
//
//	@Description	运行任务
//	@Summary		运行任务
//	@Tags			任务
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		string			true	"任务 ID"
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		400	{object}	APIResponse{}	"参数错误"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/tasks/{id} [post]
func RunTask(c *gin.Context) {
	// 请求参数
	var requestParams types.RunTaskRequest
	requestParams.ID = c.Param("id")
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	err := service.RunTask(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 删除任务
//
//	@Description	删除任务
//	@Summary		删除任务
//	@Tags			任务
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		string			true	"任务 ID"
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		400	{object}	APIResponse{}	"参数错误"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/tasks/{id} [delete]
func DeleteTask(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteTaskRequest
	requestParams.ID = c.Param("id")
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	err := service.DeleteTask(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 清理所有任务
//
//	@Description	清理所有任务
//	@Summary		清理所有任务
//	@Tags			任务
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/tasks/clean-all [post]
func CleanAllTask(c *gin.Context) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	err := service.CleanAllTask(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 启动调度器
//
//	@Description	启动调度器
//	@Summary		启动调度器
//	@Tags			任务
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/tasks/start-scheduler [post]
func StartScheduler(c *gin.Context) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	err := service.StartScheduler(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 停止调度器
//
//	@Description	停止调度器
//	@Summary		停止调度器
//	@Tags			任务
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/tasks/stop-scheduler [post]
func StopScheduler(c *gin.Context) {

	// 创建服务
	service := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
	)

	// 调用
	service.StopScheduler(c)
	successResponse(c, nil)
}
