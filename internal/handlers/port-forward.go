package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/deviceportforward"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_param"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func PortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.PortForwardRequest
	if err := gin_param.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

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
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allPortForward.Items {
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

	pageData.Page = allPortForward.Page
	pageData.PageSize = allPortForward.Size
	pageData.Total = allPortForward.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, "pages/port-forward.html", pageData)
}
