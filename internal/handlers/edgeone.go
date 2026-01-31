package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent/edgeonezone"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func EdgeOne(c *gin.Context) {
	// 请求参数
	var requestParams types.EdgeOneRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.EdgeOneResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetEdgeOneResponse, types.EdgeOneRequest]{
			Filter: requestParams,
			List:   []types.GetEdgeOneResponse{},
		},
	}

	// 查询全部EdgeOne配置
	allPlatformConfig, err := global.GvaEntDb.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		OrderByCreatedAtAsc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allPlatformConfig.Items {
			zones := make([]types.GetEdgeOneZone, 0)
			// 查询隧道列表
			allZone, err := global.GvaEntDb.EdgeOneZone.Query().
				Where(edgeonezone.ConfigID(item.ID)).
				WithAccelerationdomains().
				WithDnsrecords().
				WithOrigingroups().
				OrderByZoneNameAsc().
				All(c)
			if err == nil {
				for _, zone := range allZone {
					zones = append(zones, types.GetEdgeOneZone{
						ID:              zone.ID,
						RemoteCreatedAt: zone.RemoteCreatedAt,
						SyncAt:          zone.SyncAt,
						ZoneID:          zone.ZoneID,
						ZoneName:        zone.ZoneName,
						Area:            zone.Area.String(),
						Type:            zone.Type.String(),
						Status:          zone.Status.String(),
						CnameStatus:     zone.CnameStatus.String(),
						ActiveStatus:    zone.ActiveStatus.String(),
						Paused:          zone.Paused,

						AccelerationdomainCount: len(zone.Edges.Accelerationdomains),
						DnsrecordCount:          len(zone.Edges.Dnsrecords),
						OrigingroupCount:        len(zone.Edges.Origingroups),
					})
				}
			}
			pageData.List = append(pageData.List, types.GetEdgeOneResponse{
				ID:          item.ID,
				Capability:  item.Capability.String(),
				Name:        item.Name,
				Description: item.Description,
				Platform:    item.Platform.String(),
				Zones:       zones,
			})
		}
	}

	pageData.Page = allPlatformConfig.Page
	pageData.PageSize = allPlatformConfig.Size
	pageData.Total = allPlatformConfig.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "edgeone.html"),
		types.BasePageData[types.EdgeOneResponse]{
			Title: "EdgeOne 管理",
			Data:  pageData,
		},
	)
}
