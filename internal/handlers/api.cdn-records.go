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

// 获取全部CDN记录
//
//	@Description	获取全部CDN记录
//	@Summary		获取全部CDN记录
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	query		types.GetCdnRecordsRequest						false	"请求参数"
//	@Success		200		{object}	APIResponse{data=[]types.GetCdnRecordResponse}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/cdn-records [get]
func GetCdnRecords(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCdnRecordsRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetCdnRecords(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取CDN记录详情
//
//	@Description	获取CDN记录详情
//	@Summary		获取CDN记录详情
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64											true	"CDN记录序号"
//	@Success		200	{object}	APIResponse{data=types.GetCdnRecordResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}									"参数错误"
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Router			/cdn-records/{id} [get]
func GetCdnRecordDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCdnRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetCdnRecord(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 创建CDN记录
//
//	@Description	创建CDN记录
//	@Summary		创建CDN记录
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateCdnRecordRequest				true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetCdnRecordRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/cdn-records [post]
func CreateCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateCdnRecordRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CreateCdnRecord(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新CDN记录
//
//	@Description	更新CDN记录
//	@Summary		更新CDN记录
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"CDN记录序号"
//	@Param			request	body		types.UpdateCdnRecordRequest				true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetCdnRecordRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/cdn-records/{id} [patch]
func UpdateCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateCdnRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateCdnRecord(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新CDN记录启用状态
//
//	@Description	更新CDN记录启用状态
//	@Summary		更新CDN记录启用状态
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"CDN记录序号"
//	@Param			request	body		types.UpdateCdnRecordActiveRequest			true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetCdnRecordRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数错误"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/cdn-records/{id}/active [patch]
func UpdateCdnRecordActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateCdnRecordActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateCdnRecordActive(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 删除CDN记录
//
//	@Description	删除CDN记录
//	@Summary		删除CDN记录
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64			true	"CDN记录序号"
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		400	{object}	APIResponse{}	"参数错误"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/cdn-records/{id} [delete]
func DeleteCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteCdnRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.DeleteCdnRecord(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量删除CDN记录
//
//	@Description	批量删除CDN记录
//	@Summary		批量删除CDN记录
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.DeleteCdnRecordBatchRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}						"成功响应"
//	@Failure		400		{object}	APIResponse{}						"参数错误"
//	@Failure		404		{object}	APIResponse{}						"错误响应"
//	@Router			/cdn-records/batch [delete]
func DeleteCdnRecordBatch(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteCdnRecordBatchRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.DeleteCdnRecords(c, requestParams.IDS)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 检查CDN记录健康状态
//
//	@Description	检查CDN记录健康状态
//	@Summary		检查CDN记录健康状态
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"CDN记录序号"
//	@Success		200	{object}	APIResponse{data=[][]service.ToPairsStruct}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/cdn-records/{id}/health-check [get]
func HealthCheckCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthCheckCdnRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCDNRecordService(
		global.GvaEntDb,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.HealthcheckCdnRecord(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
