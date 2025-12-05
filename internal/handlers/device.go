package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/ent"
	"vibe_ip/ent/device"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Device(c *gin.Context) {
	// 请求参数
	var requestParams types.DeviceRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 准备页面数据
	pageData := types.DeviceResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetDeviceResponse]{
			List: []types.GetDeviceResponse{},
		},
	}

	// 查询全部设备
	allDevice, err := global.GvaDb.Device.Query().
		Order(ent.Desc(device.FieldCreatedAt)).
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allDevice.Items {
			pageData.List = append(pageData.List, types.GetDeviceResponse{
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

	pageData.Page = allDevice.Page
	pageData.PageSize = allDevice.Size
	pageData.Total = allDevice.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "device.html"), pageData)
}
