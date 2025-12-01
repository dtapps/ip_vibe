package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/ipgroup"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func IPGroup(c *gin.Context) {
	// 准备页面数据
	pageData := types.IPGroupResponse{
		List:     []types.GetIPGroupsResponse{},
		Page:     0,
		PageSize: 0,
		Total:    0,
	}

	// 查询全部IP组
	allIpGroup, err := global.GvaDb.IpGroup.Query().
		WithDevices().
		Order(ent.Desc(ipgroup.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allIpGroup {
			pageData.List = append(pageData.List, types.GetIPGroupsResponse{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				DeviceCount: len(item.Edges.Devices),
			})
		}
	}

	if global.GvaEnv.DEBUG {
		fmt.Printf("ip-group pageData: %+v \n", pageData)
	}

	// HTTP2 server 推送
	pusher(c)

	// 渲染模板
	c.HTML(http.StatusOK, "pages/ip-group.html", pageData)
}
