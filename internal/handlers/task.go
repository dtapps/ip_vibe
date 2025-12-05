package handlers

import (
	"fmt"
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
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 准备页面数据
	pageData := types.TaskResponse{
		CommonPageResponse: types.CommonPageResponse[gocron_helper.JobData]{
			List: []gocron_helper.JobData{},
		},
	}

	// 任务助手
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)

	// 获取全部任务
	pageData.List = jobHelper.GetJobsData(5)

	pageData.Page = requestParams.Page
	pageData.PageSize = requestParams.PageSize
	pageData.Total = int64(len(pageData.List))

	// 查询任务状态
	if global.GvaGoCronRunning == 1 {
		pageData.Status = "running"
	}

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "task.html"), pageData)
}
