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

// 获取全部平台配置
//
//	@Description	获取全部平台配置
//	@Summary		获取全部平台配置
//	@Tags			平台配置
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	APIResponse{data=[]types.GetConfigResponse}	"成功响应"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/configs [get]
func GetConfigs(c *gin.Context) {

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetConfigs(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取平台配置详情
//
//	@Description	获取平台配置详情
//	@Summary		获取平台配置详情
//	@Tags			平台配置
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"配置序号"
//	@Success		200	{object}	APIResponse{data=types.GetConfigResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/configs/{id} [get]
func GetConfigDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetConfigRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetConfig(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 创建平台配置
//
//	@Description	创建平台配置
//	@Summary		创建平台配置
//	@Tags			平台配置
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateConfigRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetConfigRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/configs [post]
func CreateConfig(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateConfigRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.CreateConfig(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新平台配置
//
//	@Description	更新平台配置
//	@Summary		更新平台配置
//	@Tags			平台配置
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"配置序号"
//	@Param			request	body		types.UpdateConfigRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetConfigRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/configs/{id} [patch]
func UpdateConfig(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateConfigRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateConfig(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 删除平台配置
//
//	@Description	删除平台配置
//	@Summary		删除平台配置
//	@Tags			平台配置
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"配置序号"
//	@Success		200	{object}	APIResponse{data=types.GetConfigRequest}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/configs/{id} [delete]
func DeleteConfig(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteConfigRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.DeleteConfig(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
