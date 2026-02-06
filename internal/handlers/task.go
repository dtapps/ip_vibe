package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cronjob"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Task(c *gin.Context) {
	// 请求参数
	var requestParams types.TaskRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 提前处理 EntryID 不为空的情况
	var entryID uuid.UUID
	if requestParams.EntryID != "" {
		// 解析 EntryID
		parsedEntryID, err := uuid.Parse(requestParams.EntryID)
		if err == nil {
			entryID = parsedEntryID
		} else {
			entryID = uuid.Nil
		}
	}

	// 准备页面数据
	pageData := types.TaskResponse{
		CommonPageResponse: types.CommonPageResponse[ent.CronJob, types.TaskRequest]{
			Filter: requestParams,
			List:   []ent.CronJob{},
		},
	}

	// 查询全部SSL记录
	allTask, err := global.GvaEntDb.CronJob.Query().
		WhereIf(requestParams.ID > 0, cronjob.IDEQ(requestParams.ID)).
		WhereIf(requestParams.EntryID != "", cronjob.EntryIDEQ(entryID)).
		WhereIf(requestParams.Name != "", cronjob.NameContains(requestParams.Name)).
		WhereIf(requestParams.TaskType != "", cronjob.TaskTypeEQ(cronjob.TaskType(requestParams.TaskType))).
		WhereIf(requestParams.JobType != "", cronjob.JobTypeEQ(cronjob.JobType(requestParams.JobType))).
		WhereIf(requestParams.JobCategory != "", cronjob.JobCategoryEQ(cronjob.JobCategory(requestParams.JobCategory))).
		WhereIf(requestParams.BusinessType != "", cronjob.BusinessTypeEQ(cronjob.BusinessType(requestParams.BusinessType))).
		WhereIf(requestParams.BusinessID > 0, cronjob.BusinessIDEQ(requestParams.BusinessID)).
		WhereIf(requestParams.Status != "", cronjob.StatusEQ(cronjob.Status(requestParams.Status))).
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allTask.Items {
			pageData.List = append(pageData.List, *item)
		}
	}

	pageData.Page = requestParams.Page
	pageData.PageSize = requestParams.PageSize
	pageData.Total = int64(len(pageData.List))

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "task.html"),
		types.BasePageData[types.TaskResponse]{
			Title: "任务管理",
			Data:  pageData,
		},
	)
}
