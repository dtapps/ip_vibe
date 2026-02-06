package handlers

import (
	"net/http"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 更新任务
//
//	@Description	更新任务
//	@Summary		更新任务
//	@Tags			任务
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64							true	"任务序号"
//	@Param			request	body		types.UpdateTaskRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}					"成功响应"
//	@Failure		400		{object}	APIResponse{}					"参数错误"
//	@Failure		404		{object}	APIResponse{}					"错误响应"
//	@Router			/tasks/{id} [patch]
func UpdateTask(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateTaskRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	taskService := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用 更新 任务
	err := taskService.UpdateTask(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, nil)
}

// 更新任务状态
//
//	@Description	更新任务状态
//	@Summary		更新任务状态
//	@Tags			任务
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64								true	"任务序号"
//	@Param			request	body		types.UpdateTaskStatusRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}						"成功响应"
//	@Failure		400		{object}	APIResponse{}						"参数错误"
//	@Failure		404		{object}	APIResponse{}						"错误响应"
//	@Router			/tasks/{id}/active [patch]
func UpdateTaskStatus(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateTaskStatusRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	taskService := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用 更新 任务状态
	err := taskService.UpdateTaskActive(c, requestParams.ID, requestParams.Status)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, nil)
}

// 批量更新任务状态
//
//	@Description	批量更新任务状态
//	@Summary		批量更新任务状态
//	@Tags			任务
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.BatchUpdateTaskStatusRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}							"成功响应"
//	@Failure		400		{object}	APIResponse{}							"参数错误"
//	@Failure		404		{object}	APIResponse{}							"错误响应"
//	@Router			/tasks/batch/active [patch]
func BatchUpdateTaskStatus(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchUpdateTaskStatusRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	taskService := service.NewTaskService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用 批量更新 任务状态
	err := taskService.BatchUpdateTaskActive(c, requestParams.IDS, requestParams.Status)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, nil)
}
