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

// 获取全部端口转发
//
//	@Description	获取全部端口转发
//	@Summary		获取全部端口转发
//	@Tags			端口转发
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	query		types.GetPortForwardsRequest						false	"请求参数"
//	@Success		200		{object}	APIResponse{data=[]types.GetPortForwardResponse}	"成功响应"
//	@Failure		400		{object}	APIResponse{}										"参数错误"
//	@Failure		404		{object}	APIResponse{}										"错误响应"
//	@Router			/port-forwards [get]
func GetPortForwards(c *gin.Context) {
	// 请求参数
	var requestParams types.GetPortForwardsRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,

		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetPortForwards(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取端口转发详情
//
//	@Description	获取端口转发详情
//	@Summary		获取端口转发详情
//	@Tags			端口转发
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64											true	"设备端口转发序号"
//	@Success		200	{object}	APIResponse{data=types.GetPortForwardResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}									"参数错误"
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Router			/port-forwards/{id} [get]
func GetPortForwardDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetPortForwardRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetPortForward(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 创建端口转发
//
//	@Description	创建端口转发
//	@Summary		创建端口转发
//	@Tags			端口转发
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreatePortForwardRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetPortForwardRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/port-forwards [post]
func CreatePortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.CreatePortForwardRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CreatePortForward(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新端口转发
//
//	@Description	更新端口转发
//	@Summary		更新端口转发
//	@Tags			端口转发
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64											true	"端口转发序号"
//	@Param			request	body		types.UpdatePortForwardRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetPortForwardRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/port-forwards/{id} [patch]
func UpdatePortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdatePortForwardRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdatePortForward(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新端口转发防火墙状态
//
//	@Description	更新端口转发防火墙状态
//	@Summary		更新端口转发防火墙状态
//	@Tags			端口转发
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64											true	"端口转发序号"
//	@Param			request	body		types.UpdatePortForwardForwardRequest			true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetPortForwardRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/port-forwards/{id}/forward [patch]
func UpdatePortForwardForward(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdatePortForwardForwardRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdatePortForwardForward(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新端口转发启用状态
//
//	@Description	更新端口转发启用状态
//	@Summary		更新端口转发启用状态
//	@Tags			端口转发
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64											true	"端口转发序号"
//	@Param			request	body		types.UpdatePortForwardActiveRequest			true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetPortForwardRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/port-forwards/{id}/active [patch]
func UpdatePortForwardActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdatePortForwardActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.UpdatePortForwardActive(c, requestParams.ID, requestParams.IsActive)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量更新端口转发启用状态
//
//	@Description	批量更新端口转发启用状态
//	@Summary		批量更新端口转发启用状态
//	@Tags			端口转发
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.BatchUpdatePortForwardActiveRequest			true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetPortForwardRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/port-forwards/batch/active [patch]
func BatchUpdatePortForwardActive(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchUpdatePortForwardActiveRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.BatchUpdatePortForwardActive(c, requestParams.IDS, requestParams.IsActive)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 删除端口转发
//
//	@Description	删除端口转发
//	@Summary		删除端口转发
//	@Tags			端口转发
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64			true	"端口转发序号"
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		400	{object}	APIResponse{}	"参数错误"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/port-forwards/{id} [delete]
func DeletePortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.DeletePortForwardRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.DeletePortForward(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量删除端口转发
//
//	@Description	批量删除端口转发
//	@Summary		批量删除端口转发
//	@Tags			端口转发
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.BatchDeletePortForwardRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}						"成功响应"
//	@Failure		400		{object}	APIResponse{}						"参数错误"
//	@Failure		404		{object}	APIResponse{}						"错误响应"
//	@Router			/port-forwards/batch [delete]
func BatchDeletePortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchDeletePortForwardRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.BatchDeletePortForward(c, requestParams.IDS)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 检查端口转发健康状态
//
//	@Description	检查端口转发健康状态
//	@Summary		检查端口转发健康状态
//	@Tags			端口转发
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"端口转发序号"
//	@Success		200	{object}	APIResponse{data=[][]service.ToPairsStruct}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/port-forwards/{id}/health-check [get]
func HealthCheckPortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthCheckPortForwardRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.HealthcheckPortForward(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取端口转发统计数据
//
//	@Description	获取端口转发统计数据
//	@Summary		获取端口转发统计数据
//	@Tags			端口转发
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"端口转发序号"
//	@Success		200	{object}	APIResponse{data=[][]service.ToPairsStruct}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/port-forwards/{id}/stats [get]
func GetPortForwardStats(c *gin.Context) {
	// 请求参数
	var requestParams types.GetPortForwardStatsRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetPortForwardStats(c, requestParams.ID, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
