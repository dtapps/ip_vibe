package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/cronjoblog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func TaskLog(c *gin.Context) {
	// 请求参数
	var requestParams types.TaskLogRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.TaskLogResponse{
		CommonPageResponse: types.CommonPageResponse[ent_log.CronJobLog, types.TaskLogRequest]{
			Filter: requestParams,
			List:   []ent_log.CronJobLog{},
		},
	}

	// 查询全部SSL记录
	allTask, err := global.GvaEntLogDb.CronJobLog.Query().
		WhereIf(requestParams.ID > 0, cronjoblog.IDEQ(requestParams.ID)).
		WhereIf(requestParams.CronJobID > 0, cronjoblog.CronJobIDEQ(requestParams.CronJobID)).
		WhereIf(requestParams.Status != "", cronjoblog.StatusEQ(cronjoblog.Status(requestParams.Status))).
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allTask.Items {
			pageData.List = append(pageData.List, *item)
		}
	}

	pageData.Page = allTask.Page
	pageData.PageSize = allTask.Size
	pageData.Total = allTask.Total

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "task-log.html"),
		types.BasePageData[types.TaskLogResponse]{
			Title: "任务日志",
			Data:  pageData,
		},
		"task",
	)
}
