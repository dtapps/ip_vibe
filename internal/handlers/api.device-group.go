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

// 获取全部设备组
//
//	@Description	获取全部设备组
//	@Summary		获取全部设备组
//	@Tags			设备组
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	APIResponse{data=[]types.GetDeviceGroupResponse}	"成功响应"
//	@Failure		404	{object}	APIResponse{}										"错误响应"
//	@Router			/device-groups [get]
func GetDeviceGroups(c *gin.Context) {

	// 创建服务
	service := service.NewDeviceGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	response, err := service.GetDeviceGroups(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取设备组详情
//
//	@Description	获取设备组详情
//	@Summary		获取设备组详情
//	@Tags			设备组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64											true	"设备组序号"
//	@Success		200	{object}	APIResponse{data=types.GetDeviceGroupResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}									"参数错误"
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Router			/device-groups/{id} [get]
func GetDeviceGroupDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDeviceGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	response, err := service.GetDeviceGroup(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 创建设备组
//
//	@Description	创建设备组
//	@Summary		创建设备组
//	@Tags			设备组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateDeviceGroupRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}					"成功响应"
//	@Failure		400		{object}	APIResponse{}					"参数错误"
//	@Failure		404		{object}	APIResponse{}					"错误响应"
//	@Router			/device-groups [post]
func CreateDeviceGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateDeviceGroupRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.CreateDeviceGroup(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 更新设备组
//
//	@Description	更新设备组
//	@Summary		更新设备组
//	@Tags			设备组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64							true	"设备组序号"
//	@Param			request	body		types.UpdateDeviceGroupRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}					"成功响应"
//	@Failure		400		{object}	APIResponse{}					"参数错误"
//	@Failure		404		{object}	APIResponse{}					"错误响应"
//	@Router			/device-groups/{id}/active [patch]
func UpdateDeviceGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateDeviceGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.UpdateDeviceGroup(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 删除设备组
//
//	@Description	删除设备组
//	@Summary		删除设备组
//	@Tags			设备组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64			true	"设备组序号"
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		400	{object}	APIResponse{}	"参数错误"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/device-groups/{id} [delete]
func DeleteDeviceGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteDeviceGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.DeleteDeviceGroup(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量删除设备组
//
//	@Description	批量删除设备组
//	@Summary		批量删除设备组
//	@Tags			设备组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.BatchDeleteDeviceGroupRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}						"成功响应"
//	@Failure		400		{object}	APIResponse{}						"参数错误"
//	@Failure		404		{object}	APIResponse{}						"错误响应"
//	@Router			/device-groups/batch [delete]
func BatchDeleteDeviceGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchDeleteDeviceGroupRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.BatchDeleteDeviceGroup(c, requestParams.IDS)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 获取设备组设备
//
//	@Description	获取设备组设备
//	@Summary		获取设备组设备
//	@Tags			设备组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"设备组序号"
//	@Success		200	{object}	APIResponse{data=[]types.GetDeviceResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/device-groups/{id}/devices [get]
func GetDeviceGroupDevices(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDeviceGroupDevicesRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	response, err := service.GetDeviceGroupDevices(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 检查设备组健康状态
//
//	@Description	检查设备组健康状态
//	@Summary		检查设备组健康状态
//	@Tags			设备组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"设备组序号"
//	@Success		200	{object}	APIResponse{data=[][]service.ToPairsStruct}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/device-groups/{id}/health-check [get]
func HealthCheckDeviceGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthCheckDeviceGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	response, err := service.HealthcheckDeviceGroup(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取设备组设备端口转发统计数据
//
//	@Description	获取设备组设备端口转发统计数据
//	@Summary		获取设备组设备端口转发统计数据
//	@Tags			设备组
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64																	true	"设备组序号"
//	@Success		200	{object}	APIResponse{data=types.GetDeviceGroupDevicePortForwardStatsResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}															"参数错误"
//	@Failure		404	{object}	APIResponse{}															"错误响应"
//	@Router			/device-groups/{id}/port-forward/stats [get]
func GetDeviceGroupDevicePortForwardStats(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDeviceGroupDevicePortForwardStatsRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	response, err := service.GetDeviceGroupDevicePortForwardStats(c, requestParams.ID, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
