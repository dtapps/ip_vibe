package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/device"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Device(c *gin.Context) {
	// 准备页面数据
	pageData := types.DeviceResponse{
		List:     []types.GetDevicesResponse{},
		Page:     0,
		PageSize: 0,
		Total:    0,
	}

	// 查询全部设备
	allDevice, err := global.GvaDb.Device.Query().
		Order(ent.Desc(device.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allDevice {
			pageData.List = append(pageData.List, types.GetDevicesResponse{
				ID:              item.ID,
				Code:            item.Code,
				Name:            item.Name,
				Description:     item.Description,
				IPV4:            item.IPV4,
				IPV6:            item.IPV6,
				LastReportedAt:  item.LastReportedAt,
				LastHeartbeatAt: item.LastHeartbeatAt,
			})
		}
	}

	if global.GvaEnv.DEBUG {
		fmt.Printf("device pageData: %+v \n", pageData)
	}

	// HTTP2 server 推送
	pusher(c)

	// 渲染模板
	c.HTML(http.StatusOK, "pages/device.html", pageData)
}
