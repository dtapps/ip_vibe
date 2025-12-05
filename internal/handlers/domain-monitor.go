package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/ent"
	"vibe_ip/ent/domainmonitor"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func DomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.CdnRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 准备页面数据
	pageData := types.DomainMonitorResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetDomainMonitorResponse]{
			List: []types.GetDomainMonitorResponse{},
		},
	}

	// 查询全部域名监控
	allDomainMonitor, err := global.GvaDb.DomainMonitor.Query().
		Order(ent.Desc(domainmonitor.FieldCreatedAt)).
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allDomainMonitor.Items {
			list := types.GetDomainMonitorResponse{
				ID:            item.ID,
				Name:          item.Name,
				Domain:        item.Domain,
				CheckInterval: item.CheckInterval,
				CheckProtocol: item.CheckProtocol.String(),
				IsActive:      item.IsActive,
				CurrentStatus: item.CurrentStatus.String(),
				LastSuccessAt: item.LastSuccessAt,
				LastFailureAt: item.LastFailureAt,
			}
			pageData.List = append(pageData.List, list)
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
