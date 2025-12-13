package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/platformconfig"
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
		CommonPageResponse: types.CommonPageResponse[types.GetPlatformConfigResponse]{
			List: []types.GetPlatformConfigResponse{},
		},
	}

	// 查询全部配置
	allConfig, err := global.GvaEntDb.PlatformConfig.Query().
		Order(ent.Desc(platformconfig.FieldCreatedAt)).
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allConfig.Items {
			pageData.List = append(pageData.List, types.GetPlatformConfigResponse{
				ID:          item.ID,
				Capability:  item.Capability.String(),
				Name:        item.Name,
				Description: item.Description,
				Platform:    item.Platform.String(),
			})
		}
	}

	pageData.Page = allConfig.Page
	pageData.PageSize = allConfig.Size
	pageData.Total = allConfig.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "platform-config.html"), pageData)
}
