package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/ent"
	"vibe_ip/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Config(c *gin.Context) {
	// 请求参数
	var requestParams types.ConfigRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 准备页面数据
	pageData := types.ConfigResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetConfigResponse]{
			List: []types.GetConfigResponse{},
		},
	}

	// 查询全部配置
	allConfig, err := global.GvaDb.PlatformConfig.Query().
		Order(ent.Desc(platformconfig.FieldCreatedAt)).
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allConfig.Items {
			pageData.List = append(pageData.List, types.GetConfigResponse{
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
	c.HTML(http.StatusOK, filepath.Join("pages", "config.html"), pageData)
}
