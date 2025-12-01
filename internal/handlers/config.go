package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Config(c *gin.Context) {
	// 准备页面数据
	pageData := types.ConfigResponse{
		List:     []types.GetConfigsResponse{},
		Page:     0,
		PageSize: 0,
		Total:    0,
	}

	// 查询全部配置
	allConfig, err := global.GvaDb.PlatformConfig.Query().
		Order(ent.Desc(platformconfig.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allConfig {
			pageData.List = append(pageData.List, types.GetConfigsResponse{
				ID:          item.ID,
				Capability:  item.Capability.String(),
				Name:        item.Name,
				Description: item.Description,
				Platform:    item.Platform.String(),
			})
		}
	}

	if global.GvaEnv.DEBUG {
		fmt.Printf("config pageData: %+v \n", pageData)
	}

	// HTTP2 server 推送
	pusher(c)

	// 渲染模板
	c.HTML(http.StatusOK, "pages/config.html", pageData)
}
