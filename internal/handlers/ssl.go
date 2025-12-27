package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/backend/task"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Ssl(c *gin.Context) {
	// 请求参数
	var requestParams types.SSLRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.SSLResponse{
		CommonResponseWithRequest: types.CommonResponseWithRequest{
			Request: c.Request,
		},
		CommonPageResponse: types.CommonPageResponse[types.GetSSLRecordResponse]{
			List: []types.GetSSLRecordResponse{},
		},
	}

	// 创建SSL更新任务
	sslUpdateTask := task.NewSSLUpdateTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)

	// 创建服务
	service := service.NewSSLRecordService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		sslUpdateTask,
	)

	// 查询全部SSL记录
	allSSLRecord, err := global.GvaEntDb.SSLRecord.Query().
		WithConfig().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allSSLRecord.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allSSLRecord.Page
	pageData.PageSize = allSSLRecord.Size
	pageData.Total = allSSLRecord.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "ssl.html"), pageData)
}
