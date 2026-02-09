package handlers

import (
	"net/http"
	"strings"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/provider/task"
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
//	@Success		200		{object}	types.APIResponse{}			"成功响应"
//	@Failure		400		{object}	types.APIResponse{}			"参数错误"
//	@Failure		404		{object}	types.APIResponse{}			"错误响应"
//	@Router			/esa-sites/sync [post]
func SyncESASites(c *gin.Context) {
	// 请求参数
	var requestParams types.SyncESASitesRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建 ESA站点同步 任务
	esaSiteSyncTask := task.NewESASiteSyncTaskProviderRaw(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
		global.GvaHttpLog,
	)

	if strings.EqualFold(requestParams.Type, "all") {

		// 添加 ESA站点同步 任务
		err := esaSiteSyncTask.AddTask(c, global.GvaEntDb, 0, 0, 0)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	if strings.EqualFold(requestParams.Type, "this-config") && requestParams.ConfigID > 0 {

		// 添加 ESA站点同步 任务
		err := esaSiteSyncTask.AddTask(c, global.GvaEntDb, requestParams.ConfigID, 0, 0)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	if strings.EqualFold(requestParams.Type, "this-site") && requestParams.ConfigID > 0 && requestParams.SiteID > 0 {

		// 添加 ESA站点同步 任务
		err := esaSiteSyncTask.AddTask(c, global.GvaEntDb, requestParams.ConfigID, requestParams.SiteID, 0)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	if strings.EqualFold(requestParams.Type, "this-pool") && requestParams.PoolID > 0 && requestParams.ConfigID > 0 && requestParams.SiteID > 0 {

		// 添加 ESA站点同步 任务
		err := esaSiteSyncTask.AddTask(c, global.GvaEntDb, requestParams.ConfigID, requestParams.SiteID, requestParams.PoolID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	successResponse(c, nil)
}
