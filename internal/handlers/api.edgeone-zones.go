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

// 同步EdgeOne站点
//
//	@Description	同步EdgeOne站点
//	@Summary		同步EdgeOne站点
//	@Tags			EdgeOneZone
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.SyncEdgeOneZonesRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}				"成功响应"
//	@Failure		400		{object}	types.APIResponse{}				"参数错误"
//	@Failure		404		{object}	types.APIResponse{}				"错误响应"
//	@Router			/edgeone-zones/sync [post]
func SyncEdgeOneZones(c *gin.Context) {
	// 请求参数
	var requestParams types.SyncEdgeOneZonesRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建 EdgeOne站点同步 任务
	edgeOneZoneSyncTask := task.NewEdgeOneZoneSyncTaskProviderRaw(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	if strings.EqualFold(requestParams.Type, "all") {

		// 添加 EdgeOne站点同步 一次性任务
		err := edgeOneZoneSyncTask.AddTask(c, global.GvaEntDb, 0, 0, 0)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	if strings.EqualFold(requestParams.Type, "this-config") && requestParams.ConfigID > 0 {

		// 添加 EdgeOne站点同步 一次性任务
		err := edgeOneZoneSyncTask.AddTask(c, global.GvaEntDb, requestParams.ConfigID, 0, 0)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	if strings.EqualFold(requestParams.Type, "this-zone") && requestParams.ZoneID > 0 && requestParams.ConfigID > 0 {

		// 添加 EdgeOne站点同步 一次性任务
		err := edgeOneZoneSyncTask.AddTask(c, global.GvaEntDb, requestParams.ConfigID, requestParams.ZoneID, 0)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	if strings.EqualFold(requestParams.Type, "this-group") && requestParams.GroupID > 0 && requestParams.ZoneID > 0 && requestParams.ConfigID > 0 {

		// 添加 EdgeOne站点同步 一次性任务
		err := edgeOneZoneSyncTask.AddTask(c, global.GvaEntDb, requestParams.ConfigID, requestParams.ZoneID, requestParams.GroupID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	successResponse(c, nil)
}
