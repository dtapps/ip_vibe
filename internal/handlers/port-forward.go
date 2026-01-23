package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent/deviceportforward"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func PortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.PortForwardRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.PortForwardResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetPortForwardResponse, types.PortForwardRequest]{
			Filter: requestParams,
			List:   []types.GetPortForwardResponse{},
		},
	}

	// 创建服务
	service := service.NewPortForwardService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
		sys_env.GvaEnv.DEBUG,
	)

	// 查询全部端口转发
	allPortForward, err := global.GvaEntDb.DevicePortForward.Query().
		WhereIf(requestParams.ID > 0, deviceportforward.IDEQ(requestParams.ID)).
		WhereIf(requestParams.Name != "", deviceportforward.NameContains(requestParams.Name)).
		WhereIf(!requestParams.AutoForward.IsEmpty(), deviceportforward.AutoForwardEQ(requestParams.AutoForward.Value())).
		WhereIf(!requestParams.IsActive.IsEmpty(), deviceportforward.IsActiveEQ(requestParams.IsActive.Value())).
		WithDevice().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allPortForward.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allPortForward.Page
	pageData.PageSize = allPortForward.Size
	pageData.Total = allPortForward.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "port-forward.html"),
		types.BasePageData[types.PortForwardResponse]{
			Debug: sys_env.GvaEnv.DEBUG,
			Title: "端口转发记录",
			Data:  pageData,
		},
	)
}

func PortForwardStats(c *gin.Context) {
	// 请求参数
	var requestParams types.PortForwardStatsRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.PortForwardStatsResponse{}

	// 查询规则和关联设备
	firstRule, err := global.GvaEntDb.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(requestParams.ID)).
		WithDevice().
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	pageData.Item = *firstRule
	pageData.Device = *firstRule.Edges.Device

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "port-forward_stats.html"),
		types.BasePageData[types.PortForwardStatsResponse]{
			Debug: sys_env.GvaEnv.DEBUG,
			Title: "端口转发统计",
			Data:  pageData,
		},
		"port-forward",
	)
}
