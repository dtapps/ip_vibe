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

// 同步Esa站点
//
//	@Description	同步Esa站点
//	@Summary		同步Esa站点
//	@Tags			ESASite
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.SyncESASitesRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}				"成功响应"
//	@Failure		400		{object}	APIResponse{}				"参数错误"
//	@Failure		404		{object}	APIResponse{}				"错误响应"
//	@Router			/esa-sites/sync [post]
func SyncESASites(c *gin.Context) {
	// 请求参数
	var requestParams types.SyncESASitesRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建Esa站点同步
	esaSiteSync := syncs.NewESASiteSync(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
		global.GvaHttpLog,
	)

	// 同步全部
	if strings.EqualFold(requestParams.Type, "all") {

		// 调用 同步 全部账号 所有站点
		err := esaSiteSync.SyncAllSites(c)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 调用 同步 全部账号 所有站点 所有源地址池
		err = esaSiteSync.SyncAllSiteOriginPools(c)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 调用 同步 全部账号 所有站点 所有记录
		err = esaSiteSync.SyncAllSiteRecords(c)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	// 同步某个ESA站点
	if strings.EqualFold(requestParams.Type, "this-config") && requestParams.ConfigID > 0 {

		// 调用 同步 某个账号 所有站点
		err := esaSiteSync.SyncThisSites(c, requestParams.ConfigID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 调用 同步 某个账号 所有站点 所有源地址池
		err = esaSiteSync.SyncThisSiteOriginPools(c, requestParams.ConfigID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// 调用 同步 某个账号 所有站点 所有记录
		err = esaSiteSync.SyncThisSiteRecords(c, requestParams.ConfigID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	successResponse(c, nil)
}
