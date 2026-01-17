package handlers

import (
	"net/http"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/helper/sys_env"
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
//	@Success		200	{object}	APIResponse{data=[]types.GetDeviceResponse}	"成功响应"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/devices [get]
func GetDevices(c *gin.Context) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDevices(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取设备详情
//
//	@Description	获取设备详情
//	@Summary		获取设备详情
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"设备序号"
//	@Success		200	{object}	APIResponse{data=types.GetDeviceResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/devices/{id} [get]
func GetDeviceDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDeviceRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDevice(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
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
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/devices [post]
func CreateDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateDeviceRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CreateDevice(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
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
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/devices/{id} [patch]
func UpdateDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateDeviceRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateDevice(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
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
//	@Success		200	{object}	APIResponse{data=types.GetDeviceRequest}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/devices/{id} [delete]
func DeleteDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteDeviceRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.DeleteDevice(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取设备IP历史记录
//
//	@Description	获取设备IP历史记录
//	@Summary		获取设备IP历史记录
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64													true	"设备序号"
//	@Success		200	{object}	APIResponse{data=[]types.GetDeviceIpHistoryResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}											"参数错误"
//	@Failure		404	{object}	APIResponse{}											"错误响应"
//	@Router			/devices/{id}/ip-history [get]
func GetDeviceIpHistory(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDeviceIpHistoryRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDeviceIpHistory(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
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
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/devices/{id}/health-check [get]
func HealthCheckDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthCheckDeviceRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.HealthcheckDevice(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取设备端口转发统计数据
//
//	@Description	获取设备端口转发统计数据
//	@Summary		获取设备端口转发统计数据
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"设备序号"
//	@Success		200	{object}	APIResponse{data=[][]service.ToPairsStruct}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/devices/{id}/port-forward-stats [get]
func GetDevicePortForwardStats(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDevicePortForwardStatsRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDevicePortForwardStats(c, requestParams.ID, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
