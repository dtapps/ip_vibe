package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func Device(c *gin.Context) {
	// 请求参数
	var requestParams types.DeviceRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.DeviceResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetDeviceResponse, types.DeviceRequest]{
			Filter: requestParams,
			List:   []types.GetDeviceResponse{},
		},
	}

	// 创建服务
	service := service.NewDeviceService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 查询全部设备
	allDevice, err := global.GvaEntDb.Device.Query().
		WhereIf(requestParams.ID > 0, device.IDEQ(requestParams.ID)).
		WhereIf(requestParams.Code != "", device.CodeEQ(requestParams.Code)).
		WhereIf(requestParams.Name != "", device.NameContains(requestParams.Name)).
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allDevice.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allDevice.Page
	pageData.PageSize = allDevice.Size
	pageData.Total = allDevice.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "device.html"),
		types.BasePageData[types.DeviceResponse]{
			Title: "设备列表",
			Data:  pageData,
		},
	)
}

func DevicePortForwardStats(c *gin.Context) {
	// 请求参数
	var requestParams types.DevicePortForwardStatsRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.DevicePortForwardStatsResponse{}

	// 查询规则和关联设备
	firstRule, err := global.GvaEntDb.Device.Query().
		Where(device.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	pageData.Item = *firstRule

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "device_port-forward_stats.html"),
		types.BasePageData[types.DevicePortForwardStatsResponse]{
			Title: "设备端口转发统计",
			Data:  pageData,
		},
		"device",
	)
}
