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

func PlatformConfig(c *gin.Context) {
	// 请求参数
	var requestParams types.PlatformConfigRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.PlatformConfigResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetPlatformConfigResponse, types.PlatformConfigRequest]{
			Filter: requestParams,
			List:   []types.GetPlatformConfigResponse{},
		},
	}

	// 创建服务
	service := service.NewPlatformConfigService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 查询全部配置
	allConfig, err := global.GvaEntDb.PlatformConfig.Query().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allConfig.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allConfig.Page
	pageData.PageSize = allConfig.Size
	pageData.Total = allConfig.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "platform-config.html"),
		types.BasePageData[types.PlatformConfigResponse]{
			Debug: global.GvaEnv.DEBUG,
			Title: "平台配置",
			Data:  pageData,
		},
	)
}
