package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent/devicegroup"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func DeviceGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.DeviceGroupRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.DeviceGroupResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetDeviceGroupResponse, types.DeviceGroupRequest]{
			Filter: requestParams,
			List:   []types.GetDeviceGroupResponse{},
		},
	}

	// 创建服务
	service := service.NewDeviceGroupService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 查询全部IP组
	allDeviceGroup, err := global.GvaEntDb.DeviceGroup.Query().
		WhereIf(requestParams.ID > 0, devicegroup.IDEQ(requestParams.ID)).
		WhereIf(requestParams.Code != "", devicegroup.CodeEQ(requestParams.Code)).
		WhereIf(requestParams.Name != "", devicegroup.NameContains(requestParams.Name)).
		WithDevices().
		WithCdnRecords().
		WithDNSRecords().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allDeviceGroup.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allDeviceGroup.Page
	pageData.PageSize = allDeviceGroup.Size
	pageData.Total = allDeviceGroup.Total

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "device-group.html"),
		types.BasePageData[types.DeviceGroupResponse]{
			Title: "设备组记录",
			Data:  pageData,
		},
		"device",
	)
}

func DeviceGroupPortForwardStats(c *gin.Context) {
	// 请求参数
	var requestParams types.DeviceGroupPortForwardStatsRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.DeviceGroupPortForwardStatsResponse{}

	// 查询规则和关联设备
	firstRule, err := global.GvaEntDb.DeviceGroup.Query().
		Where(devicegroup.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	pageData.Item = *firstRule

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "device-group_port-forward_stats.html"),
		types.BasePageData[types.DeviceGroupPortForwardStatsResponse]{
			Title: "设备组端口转发统计",
			Data:  pageData,
		},
		"device",
	)
}
