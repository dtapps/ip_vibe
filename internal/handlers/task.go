package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Task(c *gin.Context) {
	// 请求参数
	var requestParams types.TaskRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.TaskResponse{
		CommonPageResponse: types.CommonPageResponse[types.JobData, types.TaskRequest]{
			Filter: requestParams,
			List:   []types.JobData{},
		},
	}

	// 任务助手
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron, global.GvaGocronSlog.WithLogger())

	// 获取全部任务
	jobs := jobHelper.GetJobsData(5)
	for _, job := range jobs {
		if job != nil {
			pageData.List = append(pageData.List, *job)
		}
	}

	pageData.Page = requestParams.Page
	pageData.PageSize = requestParams.PageSize
	pageData.Total = int64(len(pageData.List))

	// 查询任务状态
	if global.GvaGoCronRunning == 1 {
		pageData.Status = "running"
	}

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "task.html"),
		types.BasePageData[types.TaskResponse]{
			Debug: global.GvaEnv.DEBUG,
			Title: "任务记录",
			Data:  pageData,
		},
	)
}
