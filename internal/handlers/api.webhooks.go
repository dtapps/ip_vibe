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

// 获取全部通知
//
//	@Description	获取全部通知
//	@Summary		获取全部通知
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.GetWebhooksRequest						false	"请求参数"
//	@Success		200		{object}	APIResponse{data=[]types.GetWebhookResponse}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/webhooks [get]
func GetWebhooks(c *gin.Context) {
	// 请求参数
	var requestParams types.GetWebhooksRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetWebhooks(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取通知
//
//	@Description	获取通知
//	@Summary		获取通知
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"通知序号"
//	@Success		200	{object}	APIResponse{data=types.GetWebhookResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/webhooks/{id} [get]
func GetWebhook(c *gin.Context) {
	// 请求参数
	var requestParams types.GetWebhookRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetWebhook(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 创建通知
//
//	@Description	创建通知
//	@Summary		创建通知
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateWebhookRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetWebhookRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Failure		500		{object}	APIResponse{}								"服务错误"
//	@Router			/webhooks [post]
func CreateWebhook(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateWebhookRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.CreateWebhook(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新通知
//
//	@Description	更新通知
//	@Summary		更新通知
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"通知序号"
//	@Param			request	body		types.UpdateWebhookRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetWebhookRequest}	"成功响应"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Failure		500		{object}	APIResponse{}								"服务错误"
//	@Router			/webhooks/{id} [patch]
func UpdateWebhook(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateWebhookRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateWebhook(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新通知启用状态
//
//	@Description	更新通知启用状态
//	@Summary		更新通知启用状态
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"通知序号"
//	@Param			request	body		types.UpdateWebhookActiveRequest			true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetWebhookRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Failure		500		{object}	APIResponse{}								"服务错误"
//	@Router			/webhooks/{id}/active [patch]
func UpdateWebhookActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateWebhookActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateWebhookActive(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 删除通知
//
//	@Description	删除通知
//	@Summary		删除通知
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"通知序号"
//	@Success		200	{object}	APIResponse{data=types.GetWebhookRequest}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Failure		500	{object}	APIResponse{}								"服务错误"
//	@Router			/webhooks/{id} [delete]
func DeleteWebhook(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteWebhookRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.DeleteWebhook(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
