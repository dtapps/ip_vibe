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

// 获取全部DNS记录
//
//	@Description	获取全部DNS记录
//	@Summary		获取全部DNS记录
//	@Tags			DNS记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	query		types.GetDnsRecordsRequest						false	"请求参数"
//	@Success		200		{object}	APIResponse{data=[]types.GetDnsRecordResponse}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/dns-records [get]
func GetDnsRecords(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDnsRecordsRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDnsRecords(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取DNS记录详情
//
//	@Description	获取DNS记录详情
//	@Summary		获取DNS记录详情
//	@Tags			DNS记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64											true	"DNS记录序号"
//	@Success		200	{object}	APIResponse{data=types.GetDnsRecordResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}									"参数错误"
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Router			/dns-records/{id} [get]
func GetDnsRecordDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDnsRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDnsRecord(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 创建DNS记录
//
//	@Description	创建DNS记录
//	@Summary		创建DNS记录
//	@Tags			DNS记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateDnsRecordRequest				true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetDnsRecordRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/dns-records [post]
func CreateDnsRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateDnsRecordRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CreateDnsRecord(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新DNS记录
//
//	@Description	更新DNS记录
//	@Summary		更新DNS记录
//	@Tags			DNS记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"DNS记录序号"
//	@Param			request	body		types.UpdateDnsRecordRequest				true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetDnsRecordRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/dns-records/{id} [patch]
func UpdateDnsRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateDnsRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateDnsRecord(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新DNS记录启用状态
//
//	@Description	更新DNS记录启用状态
//	@Summary		更新DNS记录启用状态
//	@Tags			DNS记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64								true	"DNS记录序号"
//	@Param			request	body		types.UpdateDnsRecordActiveRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}						"成功响应"
//	@Failure		400		{object}	APIResponse{}						"参数错误"
//	@Failure		404		{object}	APIResponse{}						"错误响应"
//	@Router			/dns-records/{id}/active [patch]
func UpdateDnsRecordActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateDnsRecordActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.UpdateDnsRecordActive(c, requestParams.ID, requestParams.IsActive)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量更新DNS记录启用状态
//
//	@Description	批量更新DNS记录启用状态
//	@Summary		批量更新DNS记录启用状态
//	@Tags			DNS记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.BatchUpdateDnsRecordActiveRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}							"成功响应"
//	@Failure		400		{object}	APIResponse{}							"参数错误"
//	@Failure		404		{object}	APIResponse{}							"错误响应"
//	@Router			/dns-records/batch/active [patch]
func BatchUpdateDnsRecordActive(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchUpdateDnsRecordActiveRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.BatchUpdateDnsRecordActive(c, requestParams.IDS, requestParams.IsActive)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 删除DNS记录
//
//	@Description	删除DNS记录
//	@Summary		删除DNS记录
//	@Tags			DNS记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64			true	"DNS记录序号"
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		400	{object}	APIResponse{}	"参数错误"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/dns-records/{id} [delete]
func DeleteDnsRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteDnsRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.DeleteDnsRecord(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量删除DNS记录
//
//	@Description	批量删除DNS记录
//	@Summary		批量删除DNS记录
//	@Tags			DNS记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.BatchDeleteDnsRecordRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}						"成功响应"
//	@Failure		400		{object}	APIResponse{}						"参数错误"
//	@Failure		404		{object}	APIResponse{}						"错误响应"
//	@Router			/dns-records/batch [delete]
func BatchDeleteDnsRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchDeleteDnsRecordRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.BatchDeleteDnsRecord(c, requestParams.IDS)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 检查DNS记录健康状态
//
//	@Description	检查DNS记录健康状态
//	@Summary		检查DNS记录健康状态
//	@Tags			DNS记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"DNS记录序号"
//	@Success		200	{object}	APIResponse{data=[][]service.ToPairsStruct}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/dns-records/{id}/health-check [get]
func HealthCheckDnsRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthCheckDnsRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.HealthcheckDnsRecord(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
