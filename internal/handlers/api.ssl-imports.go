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

// 创建SSL导入记录
//
//	@Description	创建SSL导入记录
//	@Summary		创建SSL导入记录
//	@Tags			SSLImport
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.ImportSSLImportRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}				"成功响应"
//	@Failure		400		{object}	types.APIResponse{}				"参数错误"
//	@Failure		404		{object}	types.APIResponse{}				"错误响应"
//	@Router			/ssl-imports/import [post]
func ImportSSLImport(c *gin.Context) {
	// 请求参数
	var requestParams types.ImportSSLImportRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewSSLImportService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.ImportSSLImport(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 创建SSL导入记录
//
//	@Description	创建SSL导入记录
//	@Summary		创建SSL导入记录
//	@Tags			SSLImport
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateSSLImportRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}				"成功响应"
//	@Failure		400		{object}	types.APIResponse{}				"参数错误"
//	@Failure		404		{object}	types.APIResponse{}				"错误响应"
//	@Router			/ssl-imports [post]
func CreateSSLImport(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateSSLImportRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewSSLImportService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.CreateSSLImport(c, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 删除SSL导入记录
//
//	@Description	删除SSL导入记录
//	@Summary		删除SSL导入记录
//	@Tags			SSLImport
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64				true	"SSL导入序号"
//	@Success		200	{object}	types.APIResponse{}	"成功响应"
//	@Failure		400	{object}	types.APIResponse{}	"参数错误"
//	@Failure		404	{object}	types.APIResponse{}	"错误响应"
//	@Router			/ssl-imports/{id} [delete]
func DeleteSSLImport(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteSSLImportRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewSSLImportService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.DeleteSSLImport(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}

// 批量删除SSL导入记录
//
//	@Description	批量删除SSL导入记录
//	@Summary		批量删除SSL导入记录
//	@Tags			SSLImport
//	@Accept			application/json
//	@Produce		application/json
//	@Param			ids	body		types.BatchDeleteSSLImportRequest	true	"SSL导入序号列表"
//	@Success		200	{object}	types.APIResponse{}					"成功响应"
//	@Failure		400	{object}	types.APIResponse{}					"参数错误"
//	@Failure		404	{object}	types.APIResponse{}					"错误响应"
//	@Router			/ssl-imports/batch [delete]
func BatchDeleteSSLImport(c *gin.Context) {
	// 请求参数
	var requestParams types.BatchDeleteSSLImportRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewSSLImportService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用
	err := service.BatchDeleteSSLImport(c, requestParams.IDS)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, nil)
}
