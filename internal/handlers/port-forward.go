package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/deviceportforward"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func PortForward(c *gin.Context) {
	// 准备页面数据
	pageData := types.PortForwardResponse{
		List:     []types.GetPortForwardsResponse{},
		Page:     0,
		PageSize: 0,
		Total:    0,
	}

	// 查询全部端口转发
	allPortForward, err := global.GvaDb.DevicePortForward.Query().
		WithDevice().
		Order(ent.Desc(deviceportforward.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allPortForward {
			pageData.List = append(pageData.List, types.GetPortForwardsResponse{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				LocalPort:   item.LocalPort,
				RemoteHost:  item.RemoteHost,
				RemotePort:  item.RemotePort,
				Protocols:   item.Protocols,
				AutoForward: item.AutoForward,
				IsActive:    item.IsActive,
				DeviceID:    item.Edges.Device.ID,
				DeviceName:  item.Edges.Device.Name,
			})
		}
	}

	if global.GvaEnv.DEBUG {
		fmt.Printf("port-forward pageData: %+v \n", pageData)
	}

	// HTTP2 server 推送
	pusher(c)

	// 渲染模板
	c.HTML(http.StatusOK, "pages/port-forward.html", pageData)
}
