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

// 获取全部IP组
//
//	@Description	获取全部IP组
//	@Summary		获取全部IP组
//	@Tags			IP组
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	APIResponse{data=[]types.GetIPGroupResponse}	"成功响应"
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Router			/ip-groups [get]
func GetIPGroups(c *gin.Context) {

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetIPGroups(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取IP组
//
//	@Description	获取IP组
//	@Summary		获取IP组
//	@Tags			IP组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"IP组序号"
//	@Success		200	{object}	APIResponse{data=types.GetIPGroupResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/ip-groups/{id} [get]
func GetIPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.GetIPGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetIPGroup(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 创建IP组
//
//	@Description	创建IP组
//	@Summary		创建IP组
//	@Tags			IP组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateIPGroupRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetIPGroupResponse}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Failure		500		{object}	APIResponse{}								"服务错误"
//	@Router			/ip-groups [post]
func CreateIPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateIPGroupRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.CreateIPGroup(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新IP组
//
//	@Description	更新IP组
//	@Summary		更新IP组
//	@Tags			IP组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"IP组序号"
//	@Param			request	body		types.UpdateIPGroupRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetIPGroupRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Failure		500		{object}	APIResponse{}								"服务错误"
//	@Router			/ip-groups/{id}/active [patch]
func UpdateIPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateIPGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.UpdateIPGroup(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 删除IP组
//
//	@Description	删除IP组
//	@Summary		删除IP组
//	@Tags			IP组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"IP组序号"
//	@Success		200	{object}	APIResponse{data=types.GetIPGroupResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Failure		500	{object}	APIResponse{}								"服务错误"
//	@Router			/ip-groups/{id} [delete]
func DeleteIPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteIPGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.DeleteIPGroup(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取IP组的设备列表
//
//	@Description	获取IP组的设备列表
//	@Summary		获取IP组的设备列表
//	@Tags			IP组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"IP组序号"
//	@Success		200	{object}	APIResponse{data=[]types.GetDeviceResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/ip-groups/{id}/devices [get]
func GetIPGroupDevices(c *gin.Context) {
	// 请求参数
	var requestParams types.GetIPGroupDevicesRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.GetIPGroupDevices(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 检查IP组健康状态
//
//	@Description	检查IP组健康状态
//	@Summary		检查IP组健康状态
//	@Tags			IP组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"IP组序号"
//	@Success		200	{object}	APIResponse{data=[][]service.ToPairsStruct}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Failure		500	{object}	APIResponse{}								"服务错误"
//	@Router			/ip-groups/{id}/health-check [get]
func HealthCheckIPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthCheckIPGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewIPGroupService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 调用
	response, err := service.HealthcheckIPGroup(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
