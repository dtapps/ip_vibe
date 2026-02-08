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

// 获取全部SSL记录
//
//	@Description	获取全部SSL记录
//	@Summary		获取全部SSL记录
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	query		types.GetSSLRecordsRequest								false	"请求参数"
//	@Success		200		{object}	types.APIResponse{data=[]types.GetSSLRecordResponse}	"成功响应"
//	@Failure		400		{object}	types.APIResponse{}										"参数错误"
//	@Failure		404		{object}	types.APIResponse{}										"错误响应"
//	@Router			/ssl-records [get]
func GetSSLRecords(c *gin.Context) {
	// 请求参数
	var requestParams types.GetSSLRecordsRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	response, err := sslService.GetSSLRecords(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取SSL记录详情
//
//	@Description	获取SSL记录详情
//	@Summary		获取SSL记录详情
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64												true	"SSL记录序号"
//	@Success		200	{object}	types.APIResponse{data=types.GetSSLRecordResponse}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}									"参数错误"
//	@Failure		404	{object}	types.APIResponse{}									"错误响应"
//	@Router			/ssl-records/{id} [get]
func GetSSLRecordDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetSSLRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	response, err := sslService.GetSSLRecord(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 导入SSL记录
//
//	@Description	导入SSL记录
//	@Summary		导入SSL记录
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.ImportSSLRecordRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}				"成功响应"
//	@Failure		400		{object}	types.APIResponse{}				"参数错误"
//	@Failure		404		{object}	types.APIResponse{}				"错误响应"
//	@Router			/ssl-records/import [post]
func ImportSSLRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.ImportSSLRecordRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := sslService.ImportSSLRecord(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 创建SSL记录
//
//	@Description	创建SSL记录
//	@Summary		创建SSL记录
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateSSLRecordRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}				"成功响应"
//	@Failure		400		{object}	types.APIResponse{}				"参数错误"
//	@Failure		404		{object}	types.APIResponse{}				"错误响应"
//	@Router			/ssl-records [post]
func CreateSSLRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateSSLRecordRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := sslService.CreateSSLRecord(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 更新SSL记录
//
//	@Description	更新SSL记录
//	@Summary		更新SSL记录
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64							true	"SSL记录序号"
//	@Param			request	body		types.UpdateSSLRecordRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}				"成功响应"
//	@Failure		400		{object}	types.APIResponse{}				"参数错误"
//	@Failure		404		{object}	types.APIResponse{}				"错误响应"
//	@Router			/ssl-records/{id} [patch]
func UpdateSSLRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateSSLRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := sslService.UpdateSSLRecord(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 更新SSL记录自动续费状态
//
//	@Description	更新SSL记录自动续费状态
//	@Summary		更新SSL记录自动续费状态
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64									true	"SSL记录序号"
//	@Param			request	body		types.UpdateSSLRecordAutoRenewalRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}						"成功响应"
//	@Failure		400		{object}	types.APIResponse{}						"参数错误"
//	@Failure		404		{object}	types.APIResponse{}						"错误响应"
//	@Router			/ssl-records/{id}/auto-renewal [patch]
func UpdateSSLRecordAutoRenewal(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateSSLRecordAutoRenewalRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := sslService.UpdateSSLRecordAutoRenewal(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 更新SSL记录启用状态
//
//	@Description	更新SSL记录启用状态
//	@Summary		更新SSL记录启用状态
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64								true	"SSL记录序号"
//	@Param			request	body		types.UpdateSSLRecordActiveRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}					"成功响应"
//	@Failure		400		{object}	types.APIResponse{}					"参数错误"
//	@Failure		404		{object}	types.APIResponse{}					"错误响应"
//	@Router			/ssl-records/{id}/active [patch]
func UpdateSSLRecordActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateSSLRecordActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := sslService.UpdateSSLRecordActive(c, requestParams.ID, requestParams.IsActive)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量更新SSL记录启用状态
//
//	@Description	批量更新SSL记录启用状态
//	@Summary		批量更新SSL记录启用状态
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.BatchUpdateSSLRecordActiveRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}						"成功响应"
//	@Failure		400		{object}	types.APIResponse{}						"参数错误"
//	@Failure		404		{object}	types.APIResponse{}						"错误响应"
//	@Router			/ssl-records/batch/active [patch]
func BatchUpdateSSLRecordActive(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchUpdateSSLRecordActiveRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := sslService.BatchUpdateSSLRecordActive(c, requestParams.IDS, requestParams.IsActive)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 删除SSL记录
//
//	@Description	删除SSL记录
//	@Summary		删除SSL记录
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64				true	"SSL记录序号"
//	@Success		200	{object}	types.APIResponse{}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}	"参数错误"
//	@Failure		404	{object}	types.APIResponse{}	"错误响应"
//	@Router			/ssl-records/{id} [delete]
func DeleteSSLRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteSSLRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := sslService.DeleteSSLRecord(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量删除SSL记录
//
//	@Description	批量删除SSL记录
//	@Summary		批量删除SSL记录
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			ids	body		types.BatchDeleteSSLRecordRequest	true	"SSL记录序号列表"
//	@Success		200	{object}	types.APIResponse{}					"成功响应"
//	@Failure		400	{object}	types.APIResponse{}					"参数错误"
//	@Failure		404	{object}	types.APIResponse{}					"错误响应"
//	@Router			/ssl-records/batch [delete]
func BatchDeleteSSLRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchDeleteSSLRecordRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := sslService.BatchDeleteSSLRecord(c, requestParams.IDS)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 申请新的SSL证书
//
//	@Description	申请新的SSL证书
//	@Summary		申请新的SSL证书
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64												true	"SSL记录序号"
//	@Success		200	{object}	types.APIResponse{data=types.GetSSLRecordRequest}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}									"参数错误"
//	@Failure		404	{object}	types.APIResponse{}									"错误响应"
//	@Router			/ssl-records/{id}/apply [post]
func ApplySSLRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.ApplySSLRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := sslService.ApplySSLRecord(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 获取SSL记录历史详情
//
//	@Description	获取SSL记录历史详情
//	@Summary		获取SSL记录历史详情
//	@Tags			SSLRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id			path		int64															true	"SSL记录序号"
//	@Param			history_id	path		int64															true	"SSL记录历史序号"
//	@Success		200			{object}	types.APIResponse{data=types.GetSSLRecordHistoryDetailResponse}	"成功响应"
//	@Failure		400			{object}	types.APIResponse{}												"参数错误"
//	@Failure		404			{object}	types.APIResponse{}												"错误响应"
//	@Router			/ssl-records/{id}/history/{history_id} [get]
func GetSSLRecordHistoryDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetSSLRecordHistoryDetailRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	requestParams.HistoryID = cast.ToInt64(c.Param("history_id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	sslService := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	response, err := sslService.GetSSLRecordHistoryDetail(c, requestParams.ID, requestParams.HistoryID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
