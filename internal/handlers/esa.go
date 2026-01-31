package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent/esasite"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func ESA(c *gin.Context) {
	// 请求参数
	var requestParams types.ESARequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.ESAResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetESAResponse, types.ESARequest]{
			Filter: requestParams,
			List:   []types.GetESAResponse{},
		},
	}

	// 查询全部ESA配置
	allPlatformConfig, err := global.GvaEntDb.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformAliyunEsa)).
		OrderByCreatedAtAsc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allPlatformConfig.Items {
			sites := make([]types.GetESASite, 0)
			// 查询隧道列表
			allSite, err := global.GvaEntDb.EsaSite.Query().
				Where(esasite.ConfigID(item.ID)).
				WithRecords().
				WithOriginpools().
				OrderBySiteNameAsc().
				All(c)
			if err == nil {
				for _, site := range allSite {
					sites = append(sites, types.GetESASite{
						ID:              site.ID,
						RemoteCreatedAt: site.RemoteCreatedAt,
						SyncAt:          site.SyncAt,
						AccessType:      site.AccessType.String(),
						CnameZone:       site.CnameZone,
						Coverage:        site.Coverage.String(),
						InstanceID:      site.InstanceID,
						NameServerList:  site.NameServerList,
						PlanName:        site.PlanName,
						PlanSpecName:    site.PlanSpecName,
						SiteID:          site.SiteID,
						SiteName:        site.SiteName,
						Status:          site.Status.String(),
						OfflineReason:   site.OfflineReason,

						RecordCount:     len(site.Edges.Records),
						OriginpoolCount: len(site.Edges.Originpools),
					})
				}
			}
			pageData.List = append(pageData.List, types.GetESAResponse{
				ID:          item.ID,
				Capability:  item.Capability.String(),
				Name:        item.Name,
				Description: item.Description,
				Platform:    item.Platform.String(),
				Sites:       sites,
			})
		}
	}

	pageData.Page = allPlatformConfig.Page
	pageData.PageSize = allPlatformConfig.Size
	pageData.Total = allPlatformConfig.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "esa.html"),
		types.BasePageData[types.ESAResponse]{
			Title: "ESA 管理",
			Data:  pageData,
		},
	)
}
