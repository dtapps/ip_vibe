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

// 获取全部TUN记录
//
//	@Description	获取全部TUN记录
//	@Summary		获取全部TUN记录
//	@Tags			TUNRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	query		types.GetTunRecordsRequest								false	"请求参数"
//	@Success		200		{object}	types.APIResponse{data=[]types.GetTunRecordResponse}	"成功响应"
//	@Failure		400		{object}	types.APIResponse{}										"参数错误"
//	@Failure		404		{object}	types.APIResponse{}										"错误响应"
//	@Router			/tun-records [get]
func GetTunRecords(c *gin.Context) {
	// 请求参数
	var requestParams types.GetTunRecordsRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewTUNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	response, err := service.GetTUNRecords(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取TUN记录详情
//
//	@Description	获取TUN记录详情
//	@Summary		获取TUN记录详情
//	@Tags			TUNRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64												true	"TUN记录序号"
//	@Success		200	{object}	types.APIResponse{data=types.GetTunRecordResponse}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}									"参数错误"
//	@Failure		404	{object}	types.APIResponse{}									"错误响应"
//	@Router			/tun-records/{id} [get]
func GetTunRecordDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetTunRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewTUNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	response, err := service.GetTUNRecord(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 创建TUN记录
//
//	@Description	创建TUN记录
//	@Summary		创建TUN记录
//	@Tags			TUNRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateTunRecordRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}				"成功响应"
//	@Failure		400		{object}	types.APIResponse{}				"参数错误"
//	@Failure		404		{object}	types.APIResponse{}				"错误响应"
//	@Router			/tun-records [post]
func CreateTunRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateTunRecordRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewTUNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.CreateTUNRecord(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 更新TUN记录
//
//	@Description	更新TUN记录
//	@Summary		更新TUN记录
//	@Tags			TUNRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64							true	"TUN记录序号"
//	@Param			request	body		types.UpdateTunRecordRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}				"成功响应"
//	@Failure		400		{object}	types.APIResponse{}				"参数错误"
//	@Failure		404		{object}	types.APIResponse{}				"错误响应"
//	@Router			/tun-records/{id} [patch]
func UpdateTunRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateTunRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewTUNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.UpdateTUNRecord(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 更新TUN记录启用状态
//
//	@Description	更新TUN记录启用状态
//	@Summary		更新TUN记录启用状态
//	@Tags			TUNRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64								true	"TUN记录序号"
//	@Param			request	body		types.UpdateTunRecordActiveRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}					"成功响应"
//	@Failure		400		{object}	types.APIResponse{}					"参数错误"
//	@Failure		404		{object}	types.APIResponse{}					"错误响应"
//	@Router			/tun-records/{id}/active [patch]
func UpdateTunRecordActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateTunRecordActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewTUNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.UpdateTUNRecordActive(c, requestParams.ID, requestParams.IsActive)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量更新TUN记录启用状态
//
//	@Description	批量更新TUN记录启用状态
//	@Summary		批量更新TUN记录启用状态
//	@Tags			TUNRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.BatchUpdateTunRecordActiveRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}						"成功响应"
//	@Failure		400		{object}	types.APIResponse{}						"参数错误"
//	@Failure		404		{object}	types.APIResponse{}						"错误响应"
//	@Router			/tun-records/batch/active [patch]
func BatchUpdateTunRecordActive(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchUpdateTunRecordActiveRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewTUNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.BatchUpdateTUNRecordActive(c, requestParams.IDS, requestParams.IsActive)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 删除TUN记录
//
//	@Description	删除TUN记录
//	@Summary		删除TUN记录
//	@Tags			TUNRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64				true	"TUN记录序号"
//	@Success		200	{object}	types.APIResponse{}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}	"参数错误"
//	@Failure		404	{object}	types.APIResponse{}	"错误响应"
//	@Router			/tun-records/{id} [delete]
func DeleteTunRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteTunRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewTUNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.DeleteTUNRecord(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量删除TUN记录
//
//	@Description	批量删除TUN记录
//	@Summary		批量删除TUN记录
//	@Tags			TUNRecord
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.BatchDeleteTunRecordRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}					"成功响应"
//	@Failure		400		{object}	types.APIResponse{}					"参数错误"
//	@Failure		404		{object}	types.APIResponse{}					"错误响应"
//	@Router			/tun-records/batch [delete]
func BatchDeleteTunRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchDeleteTunRecordRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewTUNRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.BatchDeleteTUNRecord(c, requestParams.IDS)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}
