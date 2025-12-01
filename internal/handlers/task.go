package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Task(c *gin.Context) {
	// 准备页面数据
	pageData := types.TaskResponse{
		List:     []gocron_helper.JobData{},
		Page:     0,
		PageSize: 0,
		Total:    0,
	}

	// 任务助手
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)

	// 获取全部任务
	pageData.List = jobHelper.GetJobsData(5)

	// 查询任务状态
	if global.GvaGoCronRunning == 1 {
		pageData.Status = "running"
	}

	if global.GvaEnv.DEBUG {
		fmt.Printf("task pageData: %+v \n", pageData)
	}

	// HTTP2 server 推送
	pusher(c)

	// 渲染模板
	c.HTML(http.StatusOK, "pages/task.html", pageData)
}
