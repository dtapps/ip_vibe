package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/helper/sys_env"
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
		sys_env.GvaEnv.DEBUG,
	)

	// 查询全部配置
	allConfig, err := global.GvaEntDb.PlatformConfig.Query().
		WhereIf(requestParams.ID > 0, platformconfig.IDEQ(requestParams.ID)).
		WhereIf(requestParams.Capability != "", platformconfig.CapabilityEQ(platformconfig.Capability(requestParams.Capability))).
		WhereIf(requestParams.Name != "", platformconfig.NameContains(requestParams.Name)).
		WhereIf(requestParams.Platform != "", platformconfig.PlatformEQ(platformconfig.Platform(requestParams.Platform))).
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
			Debug: sys_env.GvaEnv.DEBUG,
			Title: "平台配置",
			Data:  pageData,
		},
	)
}
