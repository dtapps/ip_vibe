package handlers

import (
	"net/http"
	"strings"
	"vibe_ip/internal/backend/syncs"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// 同步EdgeOne站点
//
//	@Description	同步EdgeOne站点
//	@Summary		同步EdgeOne站点
//	@Tags			EdgeOneZone
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.SyncEdgeOneZonesRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}					"成功响应"
//	@Failure		400		{object}	APIResponse{}					"参数错误"
//	@Failure		404		{object}	APIResponse{}					"错误响应"
//	@Router			/edgeone-zones/sync [post]
func SyncEdgeOneZones(c *gin.Context) {
	// 请求参数
	var requestParams types.SyncEdgeOneZonesRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建EdgeOne站点同步
	edgeOneZoneSync := syncs.NewEdgeOneZoneSync(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
		global.GvaHttpLog,
	)

	// 同步全部
	if strings.EqualFold(requestParams.Type, "all") {

		// 调用 同步 全部账号 所有站点
		err := edgeOneZoneSync.SyncAllZones(c)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 调用 同步 全部账号 所有站点 所有源站组
		err = edgeOneZoneSync.SyncAllZoneOriginGroups(c)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 调用 同步 全部账号 所有站点 所有记录
		err = edgeOneZoneSync.SyncAllZoneRecords(c)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	// 同步某个EdgeOne站点
	if strings.EqualFold(requestParams.Type, "this-config") && requestParams.ConfigID > 0 {

		// 调用 同步 某个账号 所有站点
		err := edgeOneZoneSync.SyncThisZones(c, requestParams.ConfigID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 调用 同步 某个账号 所有站点 所有源站组
		err = edgeOneZoneSync.SyncThisZoneOriginGroups(c, requestParams.ConfigID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 调用 同步 某个账号 所有站点 所有记录
		err = edgeOneZoneSync.SyncThisZoneRecords(c, requestParams.ConfigID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	successResponse(c, nil)
}
