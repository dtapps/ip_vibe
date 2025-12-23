package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func PortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.PortForwardRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.PortForwardResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetPortForwardResponse]{
			List: []types.GetPortForwardResponse{},
		},
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 查询全部端口转发
	allPortForward, err := global.GvaEntDb.DevicePortForward.Query().
		WithDevice().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allPortForward.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allPortForward.Page
	pageData.PageSize = allPortForward.Size
	pageData.Total = allPortForward.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "port-forward.html"), pageData)
}
