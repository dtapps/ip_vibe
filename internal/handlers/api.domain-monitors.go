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

// 获取全部域名监控
//
//	@Description	获取全部域名监控
//	@Summary		获取全部域名监控
//	@Tags			域名监控
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	query		types.GetDomainMonitorsRequest						false	"请求参数"
//	@Success		200		{object}	APIResponse{data=[]types.GetDomainMonitorResponse}	"成功响应"
//	@Failure		400		{object}	APIResponse{}										"参数错误"
//	@Failure		404		{object}	APIResponse{}										"错误响应"
//	@Router			/domain-monitors [get]
func GetDomainMonitors(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDomainMonitorsRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDomainMonitors(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 获取域名监控详情
//
//	@Description	获取域名监控详情
//	@Summary		获取域名监控详情
//	@Tags			域名监控
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64												true	"域名监控记录序号"
//	@Success		200	{object}	APIResponse{data=types.GetDomainMonitorResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}										"参数错误"
//	@Failure		404	{object}	APIResponse{}										"错误响应"
//	@Router			/domain-monitors/{id} [get]
func GetDomainMonitorDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDomainMonitorRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.GetDomainMonitor(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 创建域名监控
//
//	@Description	创建域名监控
//	@Summary		创建域名监控
//	@Tags			域名监控
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateDomainMonitorRequest				true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetDomainMonitorRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/domain-monitors [post]
func CreateDomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateDomainMonitorRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CreateDomainMonitor(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新域名监控
//
//	@Description	更新域名监控
//	@Summary		更新域名监控
//	@Tags			域名监控
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64											true	"域名监控序号"
//	@Param			request	body		types.UpdateDomainMonitorRequest				true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetDomainMonitorRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/domain-monitors/{id} [patch]
func UpdateDomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateDomainMonitorRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.UpdateDomainMonitor(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}

// 更新域名监控启用状态
//
//	@Description	更新域名监控启用状态
//	@Summary		更新域名监控启用状态
//	@Tags			域名监控
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64											true	"域名监控序号"
//	@Param			request	body		types.UpdateDomainMonitorActiveRequest			true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetDomainMonitorRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/domain-monitors/{id}/active [patch]
func UpdateDomainMonitorActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateDomainMonitorActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.UpdateDomainMonitorActive(c, requestParams.ID, requestParams.IsActive)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量更新域名监控启用状态
//
//	@Description	批量更新域名监控启用状态
//	@Summary		批量更新域名监控启用状态
//	@Tags			域名监控
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.BatchUpdateDomainMonitorActiveRequest			true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetDomainMonitorRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/domain-monitors/batch/active [patch]
func BatchUpdateDomainMonitorActive(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchUpdateDomainMonitorActiveRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.BatchUpdateDomainMonitorActive(c, requestParams.IDS, requestParams.IsActive)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 删除域名监控
//
//	@Description	删除域名监控
//	@Summary		删除域名监控
//	@Tags			域名监控
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64			true	"域名监控序号"
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		400	{object}	APIResponse{}	"参数错误"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/domain-monitors/{id} [delete]
func DeleteDomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteDomainMonitorRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.DeleteDomainMonitor(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量删除域名监控
//
//	@Description	批量删除域名监控
//	@Summary		批量删除域名监控
//	@Tags			域名监控
//	@Accept			application/json
//	@Produce		application/json
//	@Param			ids	body		types.BatchDeleteDomainMonitorRequest	true	"域名监控序号列表"
//	@Success		200	{object}	APIResponse{}							"成功响应"
//	@Failure		400	{object}	APIResponse{}							"参数错误"
//	@Failure		404	{object}	APIResponse{}							"错误响应"
//	@Router			/domain-monitors/batch [delete]
func BatchDeleteDomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchDeleteDomainMonitorRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	err := service.BatchDeleteDomainMonitor(c, requestParams.IDS)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 检查域名监控
//
//	@Description	检查域名监控健康状态
//	@Summary		检查域名监控健康状态
//	@Tags			域名监控
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64											true	"域名监控序号"
//	@Success		200	{object}	APIResponse{data=types.GetDomainMonitorRequest}	"成功响应"
//	@Failure		400	{object}	APIResponse{}									"参数错误"
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Router			/domain-monitors/{id}/check [get]
func CheckDomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.CheckDomainMonitorRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
		sys_env.GvaEnv.DEBUG,
	)

	// 调用
	response, err := service.CheckDomainMonitor(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, response)
}
