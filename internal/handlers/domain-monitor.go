package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent/domainmonitor"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/domainmonitorhistory"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func DomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.DomainMonitorRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.DomainMonitorResponse{
		CommonResponseWithRequest: types.CommonResponseWithRequest{
			Request: c.Request,
		},
		CommonPageResponse: types.CommonPageResponse[types.GetDomainMonitorResponse]{
			List: []types.GetDomainMonitorResponse{},
		},
	}

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaOperationLogHelper,
		global.GvaGoCron,
		global.GvaGocronSlog.WithLogger(),
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)

	// 查询全部域名监控
	allDomainMonitor, err := global.GvaEntDb.DomainMonitor.Query().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allDomainMonitor.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allDomainMonitor.Page
	pageData.PageSize = allDomainMonitor.Size
	pageData.Total = allDomainMonitor.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "domain-monitor.html"), pageData)
}

func DomainMonitorDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.DomainMonitorDetailRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	pageData := types.DomainMonitorDetailResponse{
		CommonResponseWithRequest: types.CommonResponseWithRequest{
			Request: c.Request,
		},
	}

	// 查询域名监控
	firstDomainMonitor, err := global.GvaEntDb.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询域名监控失败: "+err.Error())
		return
	}
	pageData.Item = *firstDomainMonitor

	// 查询最新的监控记录
	latestHistory, err := global.GvaEntLogDb.DomainMonitorHistory.Query().
		Where(domainmonitorhistory.MonitorIDEQ(requestParams.ID)).
		OrderByCheckTimeDesc().
		First(c)
	if err == nil {
		pageData.LatestHistory = *latestHistory
	}

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "domain-monitor_detail.html"), pageData)
}

func DomainMonitorHistory(c *gin.Context) {
	// 请求参数
	var requestParams types.DomainMonitorHistoryRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.DomainMonitorHistoryResponse{
		CommonResponseWithRequest: types.CommonResponseWithRequest{
			Request: c.Request,
		},
		CommonPageResponse: types.CommonPageResponse[ent_log.DomainMonitorHistory]{
			List: []ent_log.DomainMonitorHistory{},
		},
	}

	// 查询域名监控
	firstDomainMonitor, err := global.GvaEntDb.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询域名监控失败: "+err.Error())
		return
	}
	pageData.Item = *firstDomainMonitor

	// 查询全部域名监控
	allDomainMonitor, err := global.GvaEntLogDb.DomainMonitorHistory.Query().
		Where(domainmonitorhistory.MonitorIDEQ(requestParams.ID)).
		OrderByCheckTimeDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allDomainMonitor.Items {
			pageData.List = append(pageData.List, *item)
		}
	}

	pageData.Page = allDomainMonitor.Page
	pageData.PageSize = allDomainMonitor.Size
	pageData.Total = allDomainMonitor.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "domain-monitor_history.html"), pageData)
}
