package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Gocron(c *gin.Context) {
	// 请求参数
	var requestParams types.GocronRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.GocronResponse{
		CommonPageResponse: types.CommonPageResponse[types.GocronJob, types.GocronRequest]{
			Filter: requestParams,
			List:   []types.GocronJob{},
		},
	}

	// 创建服务
	gocronService := service.NewGocronService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGinSlog,
	)

	// 获取全部任务
	jobs, _ := gocronService.GetTasks(c)
	for _, job := range jobs {
		if job != nil {
			pageData.List = append(pageData.List, *job)
		}
	}

	pageData.Page = requestParams.Page
	pageData.PageSize = requestParams.PageSize
	pageData.Total = int64(len(jobs))

	// 查询任务状态
	if global.GvaGoCronRunning == 1 {
		pageData.Status = "running"
	}

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "gocron.html"),
		types.BasePageData[types.GocronResponse]{
			Title: "任务记录",
			Data:  pageData,
		},
	)
}
