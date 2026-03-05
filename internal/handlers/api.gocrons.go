package handlers

import (
	"net/http"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// 运行任务
//
//	@Description	运行任务
//	@Summary		运行任务
//	@Tags			Gocron
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		string				true	"任务 ID"
//	@Success		200	{object}	types.APIResponse{}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}	"参数错误"
//	@Failure		404	{object}	types.APIResponse{}	"错误响应"
//	@Router			/gocrons/{id} [post]
func RunGocron(c *gin.Context) {
	// 请求参数
	var requestParams types.RunGocronRequest
	requestParams.ID = c.Param("id")
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	gocronService := service.NewGocronService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGinSlog,
	)

	// 调用 运行任务
	err := gocronService.RunTask(c, requestParams.ID)
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
//	@Tags			Gocron
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		string				true	"任务 ID"
//	@Success		200	{object}	types.APIResponse{}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}	"参数错误"
//	@Failure		404	{object}	types.APIResponse{}	"错误响应"
//	@Router			/gocrons/{id} [delete]
func DeleteGocron(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteGocronRequest
	requestParams.ID = c.Param("id")
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	gocronService := service.NewGocronService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGinSlog,
	)

	// 调用 删除任务
	err := gocronService.DeleteTask(c, requestParams.ID)
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
//	@Tags			Gocron
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	types.APIResponse{}	"成功响应"
//	@Failure		404	{object}	types.APIResponse{}	"错误响应"
//	@Router			/gocrons/clean-all [post]
func CleanAllGocron(c *gin.Context) {

	// 创建服务
	gocronService := service.NewGocronService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGinSlog,
	)

	// 调用 清理全部任务
	err := gocronService.CleanAllTask(c)
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
//	@Tags			Gocron
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	types.APIResponse{}	"成功响应"
//	@Failure		404	{object}	types.APIResponse{}	"错误响应"
//	@Router			/gocrons/start-scheduler [post]
func StartGocronScheduler(c *gin.Context) {

	// 调用 启动调度器
	initialize.StartGoCron(c, global.GvaGocronSlog)

	successResponse(c, nil)
}

// 停止调度器
//
//	@Description	停止调度器
//	@Summary		停止调度器
//	@Tags			Gocron
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	types.APIResponse{}	"成功响应"
//	@Failure		404	{object}	types.APIResponse{}	"错误响应"
//	@Router			/gocrons/stop-scheduler [post]
func StopGocronScheduler(c *gin.Context) {

	// 调用 停止调度器
	initialize.StopGoCron(c, global.GvaGocronSlog)

	successResponse(c, nil)
}
