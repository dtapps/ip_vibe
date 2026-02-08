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

// 同步Cloudflare站点
//
//	@Description	同步Cloudflare站点
//	@Summary		同步Cloudflare站点
//	@Tags			CloudflareZone
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.SyncCloudflareZonesRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}					"成功响应"
//	@Failure		400		{object}	types.APIResponse{}					"参数错误"
//	@Failure		404		{object}	types.APIResponse{}					"错误响应"
//	@Router			/cloudflare-zones/sync [post]
func SyncCloudflareZones(c *gin.Context) {
	// 请求参数
	var requestParams types.SyncCloudflareZonesRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建Cloudflare站点同步
	cloudflareZoneSync := syncs.NewCloudflareZoneSync(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
		global.GvaHttpLog,
	)

	if strings.EqualFold(requestParams.Type, "all") {

		// 调用 同步 全部账号 所有站点
		err := cloudflareZoneSync.SyncAllZones(c)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 调用 同步 全部账号 所有站点 所有记录
		err = cloudflareZoneSync.SyncAllZoneRecords(c)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	if strings.EqualFold(requestParams.Type, "this-config") && requestParams.ConfigID > 0 {

		// 调用 同步 某个账号 所有站点
		err := cloudflareZoneSync.SyncThisZones(c, requestParams.ConfigID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 调用 同步 某个账号 所有站点 所有记录
		err = cloudflareZoneSync.SyncThisZoneRecords(c, requestParams.ConfigID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	if strings.EqualFold(requestParams.Type, "this-zone") && requestParams.ConfigID > 0 && requestParams.ZoneID > 0 {

		// 调用 同步 某个账号 某个站点
		err := cloudflareZoneSync.SyncThisSite(c, requestParams.ConfigID, requestParams.ZoneID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 调用 同步 某个账号 某个站点 所有记录
		err = cloudflareZoneSync.SyncThisZoneRecord(c, requestParams.ConfigID, requestParams.ZoneID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	successResponse(c, nil)
}
