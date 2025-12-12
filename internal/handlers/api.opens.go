package handlers

import (
	"net/http"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/backend/webhook"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// 设备上报IP
//
//	@Description	设备上报IP
//	@Summary		设备上报IP
//	@Tags			开放
//	@Accept			application/json
//	@Produce		application/json
//	@Param			device_code	path		string													true	"设备编码"
//	@Param			request		body		types.DeviceReportDeviceIpRequest						true	"请求参数"
//	@Success		200			{object}	APIResponse{data=types.DeviceReportDeviceIpResponse}	"成功响应"
//	@Failure		400			{object}	APIResponse{}											"参数错误"
//	@Failure		404			{object}	APIResponse{}											"错误响应"
//	@Failure		500			{object}	APIResponse{}											"服务错误"
//	@Router			/open/device/{device_code}/ip-report [post]
func DeviceReportDeviceIp(c *gin.Context) {
	// 请求参数
	var requestParams types.DeviceReportDeviceIpRequest
	requestParams.DeviceCode = c.Param("device_code")
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewDeviceOpenService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		webhook.NewSend(global.GvaEntDb),
	)

	// 调用
	response, err := service.DeviceReportDeviceIp(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取设备端口转发
//
//	@Description	获取设备端口转发
//	@Summary		获取设备端口转发
//	@Tags			开放
//	@Accept			application/json
//	@Produce		application/json
//	@Param			device_code	path		string													true	"设备编码"
//	@Success		200			{object}	APIResponse{data=[]types.DevicePortForwardsResponse}	"成功响应"
//	@Failure		400			{object}	APIResponse{}											"参数错误"
//	@Failure		404			{object}	APIResponse{}											"错误响应"
//	@Router			/open/device/{device_code}/port-forwards [get]
func DevicePortForwards(c *gin.Context) {
	// 请求参数
	var requestParams types.DevicePortForwardsRequest
	requestParams.DeviceCode = c.Param("device_code")
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建服务
	service := service.NewDeviceOpenService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		webhook.NewSend(global.GvaEntDb),
	)

	// 调用
	response, err := service.DevicePortForwards(c, requestParams.DeviceCode)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
