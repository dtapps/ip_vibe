package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent/domainmonitor"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/domainmonitorhistory"
	"vibe_ip/internal/ent_log/operationlog"
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
		CommonPageResponse: types.CommonPageResponse[types.GetDomainMonitorResponse, types.DomainMonitorRequest]{
			Filter: requestParams,
			List:   []types.GetDomainMonitorResponse{},
		},
	}

	// 创建服务
	service := service.NewDomainMonitorService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
	)

	// 查询全部域名监控
	allDomainMonitor, err := global.GvaEntDb.DomainMonitor.Query().
		WhereIf(requestParams.ID > 0, domainmonitor.IDEQ(requestParams.ID)).
		WhereIf(requestParams.Name != "", domainmonitor.NameContains(requestParams.Name)).
		WhereIf(requestParams.Domain != "", domainmonitor.DomainContains(requestParams.Domain)).
		WhereIf(requestParams.CheckProtocol != "", domainmonitor.CheckProtocolEQ(domainmonitor.CheckProtocol(requestParams.CheckProtocol))).
		WhereIf(!requestParams.IsActive.IsEmpty(), domainmonitor.IsActiveEQ(requestParams.IsActive.Value())).
		WhereIf(requestParams.CurrentStatus != "", domainmonitor.CurrentStatusEQ(domainmonitor.CurrentStatus(requestParams.CurrentStatus))).
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

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "domain-monitor.html"),
		types.BasePageData[types.DomainMonitorResponse]{
			Title: "域名监控记录",
			Data:  pageData,
		},
	)
}

func DomainMonitorDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.DomainMonitorDetailRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.DomainMonitorDetailResponse{}

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
		Where(domainmonitorhistory.MonitorIDEQ(firstDomainMonitor.ID)).
		OrderByCheckTimeDesc().
		First(c)
	if err == nil {
		pageData.LatestHistory = *latestHistory
	}

	// 查询通知订阅
	allWebhookSubscriptions, err := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDomainMonitor)).
		Where(webhooksubscription.ObjIDEQ(firstDomainMonitor.ID)).
		All(c)
	if err == nil {
		for _, item := range allWebhookSubscriptions {
			pageData.WebhookSubs = append(pageData.WebhookSubs, *item)
		}
	}

	// 最近操作日志
	allOperationLogs, err := global.GvaEntLogDb.OperationLog.Query().
		Where(operationlog.SourceIn(operationlog.SourceUser)).
		Where(operationlog.ObjTypeEQ(operationlog.ObjTypeDomainMonitor)).
		// Where(operationlog.ObjIDEQ(cast.ToString(firstDomainMonitor.ID))).
		WhereObjIDContains(firstDomainMonitor.ID).
		OrderByCreatedAtDesc().
		Limit(5).
		All(c)
	if err == nil {
		for _, item := range allOperationLogs {
			pageData.OperationLogs = append(pageData.OperationLogs, *item)
		}
	}

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "domain-monitor_detail.html"),
		types.BasePageData[types.DomainMonitorDetailResponse]{
			Title: "域名监控详情",
			Data:  pageData,
		},
		"domain-monitor",
	)
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
		CommonPageResponse: types.CommonPageResponse[ent_log.DomainMonitorHistory, types.DomainMonitorHistoryRequest]{
			Filter: requestParams,
			List:   []ent_log.DomainMonitorHistory{},
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

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "domain-monitor_history.html"),
		types.BasePageData[types.DomainMonitorHistoryResponse]{
			Title: "域名监控记录",
			Data:  pageData,
		},
		"domain-monitor",
	)
}
