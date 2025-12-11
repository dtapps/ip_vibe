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

// 获取全部通知订阅
//
//	@Description	获取全部通知订阅
//	@Summary		获取全部通知订阅
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.GetWebhookSubsRequest						false	"请求参数"
//	@Success		200		{object}	APIResponse{data=[]types.GetWebhookSubResponse}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/webhook-subs [get]
func GetWebhookSubs(c *gin.Context) {
	// 请求参数
	var requestParams types.GetWebhookSubsRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetWebhookSubs(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取通知订阅
//
//	@Description	获取通知订阅
//	@Summary		获取通知订阅
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64											true	"通知订阅序号"
//	@Success		200	{object}	APIResponse{data=types.GetWebhookSubResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"请求参数"
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Router			/webhook-subs/{id} [get]
func GetWebhookSub(c *gin.Context) {
	// 请求参数
	var requestParams types.GetWebhookSubRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetWebhookSub(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 创建通知订阅
//
//	@Description	创建通知订阅
//	@Summary		创建通知订阅
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateWebhookSubRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.CreateWebhookSubResponse}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"请求参数"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Failure		500		{object}	APIResponse{}									"服务错误"
//	@Router			/webhook-subs [post]
func CreateWebhookSub(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateWebhookSubRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.CreateWebhookSub(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新通知订阅
//
//	@Description	更新通知订阅
//	@Summary		更新通知订阅
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64											true	"通知订阅序号"
//	@Param			request	body		types.UpdateWebhookSubRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetWebhookSubRequest}	"成功响应"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Failure		500		{object}	APIResponse{}									"服务错误"
//	@Router			/webhook-subs/{id} [patch]
func UpdateWebhookSub(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateWebhookSubRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateWebhookSub(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新通知订阅启用状态
//
//	@Description	更新通知订阅启用状态
//	@Summary		更新通知订阅启用状态
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64											true	"通知订阅序号"
//	@Param			request	body		types.UpdateWebhookSubActiveRequest				true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetWebhookSubRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"请求参数"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Failure		500		{object}	APIResponse{}									"服务错误"
//	@Router			/webhook-subs/{id}/active [patch]
func UpdateWebhookSubActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateWebhookSubActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateWebhookSubActive(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 删除通知订阅
//
//	@Description	删除通知订阅
//	@Summary		删除通知订阅
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64											true	"通知订阅序号"
//	@Success		200	{object}	APIResponse{data=types.GetWebhookSubRequest}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"请求参数"
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Failure		500	{object}	APIResponse{}									"服务错误"
//	@Router			/webhook-subs/{id} [delete]
func DeleteWebhookSub(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteWebhookSubRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.DeleteWebhookSub(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
