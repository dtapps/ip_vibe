package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 获取全部设备
//
//	@Description	获取全部设备
//	@Summary		获取全部设备
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.GetDeviceResponse						false	"请求参数"
//	@Success		200		{object}	APIResponse{data=[]types.GetDeviceResponse}	"成功响应"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/devices [get]
func GetDevices(c *gin.Context) {
	devices, err := service.NewDeviceService(global.GvaEntDb).GetDevices(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}
	successResponse(c, devices)
}

// 获取设备
//
//	@Description	获取设备
//	@Summary		获取设备
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"设备序号"
//	@Success		200	{object}	APIResponse{data=types.GetDeviceResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}							"请求参数"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/devices/{id} [get]
func GetDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDeviceRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	response, err := service.NewDeviceService(global.GvaEntDb).GetDevice(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}

	successResponse(c, response)
}

// 创建设备
//
//	@Description	创建设备
//	@Summary		创建设备
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateDeviceRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetDeviceRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}							"请求参数"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Failure		500		{object}	APIResponse{}								"服务错误"
//	@Router			/devices [post]
func CreateDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateDeviceRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	response, err := service.NewDeviceService(global.GvaEntDb).CreateDevice(c, &requestParams)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("创建设备: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "创建设备失败: "+err.Error())
		return
	}

	successResponse(c, response)
}

// 更新设备
//
//	@Description	更新设备
//	@Summary		更新设备
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"设备序号"
//	@Param			request	body		types.UpdateDeviceRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetDeviceRequest}	"成功响应"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Failure		500		{object}	APIResponse{}								"服务错误"
//	@Router			/devices/{id} [patch]
func UpdateDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateDeviceRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	response, err := service.NewDeviceService(global.GvaEntDb).UpdateDevice(c, &requestParams)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新设备: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新设备失败: "+err.Error())
		return
	}

	successResponse(c, response)
}

// 删除设备
//
//	@Description	删除设备
//	@Summary		删除设备
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"设备序号"
//	@Success		200	{object}	APIResponse{data=types.GetDeviceResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}							"请求参数"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Failure		500	{object}	APIResponse{}								"服务错误"
//	@Router			/devices/{id} [delete]
func DeleteDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteDeviceRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	response, err := service.NewDeviceService(global.GvaEntDb).DeleteDevice(c, requestParams.ID)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("删除设备: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "删除设备失败: "+err.Error())
		return
	}

	successResponse(c, response)
}

// 获取设备的IP历史记录
//
//	@Description	获取设备的IP历史记录
//	@Summary		获取设备的IP历史记录
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64													true	"设备序号"
//	@Success		200	{object}	APIResponse{data=[]types.GetDeviceIpHistoryResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}										"请求参数"
//	@Failure		404	{object}	APIResponse{}											"错误响应"
//	@Router			/devices/{id}/ip-history [get]
func GetDeviceIpHistory(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDeviceIpHistoryRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	response, err := service.NewDeviceService(global.GvaEntDb).GetDeviceIpHistory(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备IP历史记录失败: "+err.Error())
		return
	}

	successResponse(c, response)
}

// 检查设备健康状态
//
//	@Description	检查设备健康状态
//	@Summary		检查设备健康状态
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"设备序号"
//	@Success		200	{object}	APIResponse{data=[][]service.ToPairsStruct}	"成功响应"
//	@Failure		400	{object}	APIResponse{}							"请求参数"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Failure		500	{object}	APIResponse{}								"服务错误"
//	@Router			/devices/{id}/healthcheck [get]
func HealthcheckDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthcheckDeviceRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 使用服务层检查健康状态
	results, err := service.NewDeviceService(global.GvaEntDb).HealthcheckDevice(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "设备健康状态检查失败: "+err.Error())
		return
	}

	successResponse(c, results)
}
