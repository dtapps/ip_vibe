package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_server"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gotime"
)

// 获取全部设备
//
//	@Description	获取全部设备
//	@Summary		获取全部设备
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	types.APIResponse{data=[]types.GetDeviceResponse}	"成功响应"
//	@Failure		404	{object}	types.APIResponse{}									"错误响应"
//	@Router			/devices [get]
func GetDevices(c *gin.Context) {

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
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
//	@Param			id	path		int64											true	"设备序号"
//	@Success		200	{object}	types.APIResponse{data=types.GetDeviceResponse}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}								"参数错误"
//	@Failure		404	{object}	types.APIResponse{}								"错误响应"
//	@Router			/devices/{id} [get]
func GetDeviceDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDeviceRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
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
//	@Param			request	body		types.CreateDeviceRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}			"成功响应"
//	@Failure		400		{object}	types.APIResponse{}			"参数错误"
//	@Failure		404		{object}	types.APIResponse{}			"错误响应"
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
		global.GvaGinSlog,
	)

	// 调用
	err := service.CreateDevice(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 更新设备
//
//	@Description	更新设备
//	@Summary		更新设备
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64						true	"设备序号"
//	@Param			request	body		types.UpdateDeviceRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}			"成功响应"
//	@Failure		400		{object}	types.APIResponse{}			"参数错误"
//	@Failure		404		{object}	types.APIResponse{}			"错误响应"
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
		global.GvaGinSlog,
	)

	// 调用
	err := service.UpdateDevice(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 删除设备
//
//	@Description	删除设备
//	@Summary		删除设备
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64				true	"设备序号"
//	@Success		200	{object}	types.APIResponse{}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}	"参数错误"
//	@Failure		404	{object}	types.APIResponse{}	"错误响应"
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
		global.GvaGinSlog,
	)

	// 调用
	err := service.DeleteDevice(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量删除设备
//
//	@Description	批量删除设备
//	@Summary		批量删除设备
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.BatchDeleteDeviceRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}				"成功响应"
//	@Failure		400		{object}	types.APIResponse{}				"参数错误"
//	@Failure		404		{object}	types.APIResponse{}				"错误响应"
//	@Router			/devices/batch [delete]
func BatchDeleteDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchDeleteDeviceRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.BatchDeleteDevice(c, requestParams.IDS)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 获取设备IP历史记录
//
//	@Description	获取设备IP历史记录
//	@Summary		获取设备IP历史记录
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64														true	"设备序号"
//	@Success		200	{object}	types.APIResponse{data=[]types.GetDeviceIpHistoryResponse}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}											"参数错误"
//	@Failure		404	{object}	types.APIResponse{}											"错误响应"
//	@Router			/devices/{id}/ip-history [get]
func GetDeviceIpHistory(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDeviceIpHistoryRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
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
//	@Param			id	path		int64												true	"设备序号"
//	@Success		200	{object}	types.APIResponse{data=[][]service.ToPairsStruct}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}									"参数错误"
//	@Failure		404	{object}	types.APIResponse{}									"错误响应"
//	@Router			/devices/{id}/health-check [get]
func HealthCheckDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthCheckDeviceRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
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
//	@Param			id	path		int64												true	"设备序号"
//	@Success		200	{object}	types.APIResponse{data=[][]service.ToPairsStruct}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}									"参数错误"
//	@Failure		404	{object}	types.APIResponse{}									"错误响应"
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
		global.GvaGinSlog,
	)

	// 调用
	response, err := service.GetDevicePortForwardStats(c, requestParams.ID, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取设备状态
//
//	@Description	获取设备状态
//	@Summary		获取设备状态
//	@Tags			设备
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64													true	"设备序号"
//	@Success		200	{object}	types.APIResponse{data=[]types.GetDeviceStatusResponse}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}										"参数错误"
//	@Failure		404	{object}	types.APIResponse{}										"错误响应"
//	@Router			/devices/{id}/status [get]
func GetDeviceStatus(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDeviceStatusRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用 获取单个设备
	device, err := service.GetDevice(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	// 创建 gRPC 服务
	grpcMgr := grpc_server.NewManager(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGrpcSlog,
		grpc_server.Config{},
	)

	// 获取 DeviceService 实例
	deviceService := grpcMgr.GetDeviceService()

	// 请求全部设备状态
	deviceService.RequestFullReportToDevice(device.Code)

	// 查询全部设备状态
	clientStates := deviceService.GetClients()

	// 查找该设备的所有角色实例
	var roles []types.GetDeviceStatusResponse
	prefix := device.Code + "|"

	for key, state := range clientStates {
		if strings.HasPrefix(key, prefix) {
			info := state.Info
			role := strings.TrimPrefix(key, prefix)
			roles = append(roles, types.GetDeviceStatusResponse{
				Role:      role,
				Version:   info.Version,
				GitCommit: info.GitCommit,
				BuildTime: info.BuildTime,
				GoVersion: info.GoVersion,
				Os:        info.Os,
				Arch:      info.Arch,
				RunTime:   gotime.SetCurrent(info.RunTime.AsTime()).Format(),
				StartTime: gotime.SetCurrent(info.StartTime.AsTime()).Format(),
				GetTime:   gotime.SetCurrent(info.GetTime.AsTime()).Format(),
			})
		}
	}

	if len(roles) == 0 {
		errorResponse(c, http.StatusNotFound, fmt.Sprintf("设备 %s 无在线客户端", device.Code))
		return
	}

	successResponse(c, roles)
}
