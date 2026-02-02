package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/edgeonezone"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func EdgeOneZone(c *gin.Context) {
	// 请求参数
	var requestParams types.EdgeOneZoneRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.EdgeOneZoneResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetEdgeOneZoneResponse, types.EdgeOneZoneRequest]{
			Filter: requestParams,
			List:   []types.GetEdgeOneZoneResponse{},
		},
	}

	// 查询全部EdgeOne配置
	allPlatformConfig, err := global.GvaEntDb.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformTencentEo)).
		OrderByCreatedAtAsc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allPlatformConfig.Items {
			zones := make([]types.GetEdgeOneZoneZone, 0)
			// 查询隧道列表
			allZone, err := global.GvaEntDb.EdgeOneZone.Query().
				Where(edgeonezone.ConfigID(item.ID)).
				WithAccelerationDomains().
				WithDNSRecords().
				WithOriginGroups().
				OrderByZoneNameAsc().
				All(c)
			if err == nil {
				for _, zone := range allZone {
					zones = append(zones, types.GetEdgeOneZoneZone{
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

						AccelerationdomainCount: len(zone.Edges.AccelerationDomains),
						DnsrecordCount:          len(zone.Edges.DNSRecords),
						OrigingroupCount:        len(zone.Edges.OriginGroups),
					})
				}
			}
			pageData.List = append(pageData.List, types.GetEdgeOneZoneResponse{
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
		filepath.Join("pages", "edgeone-zone.html"),
		types.BasePageData[types.EdgeOneZoneResponse]{
			Title: "EdgeOne 站点管理",
			Data:  pageData,
		},
	)
}

func EdgeOneZoneRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.EdgeOneZoneRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.EdgeOneZoneRecordResponse{
		CommonPageWithItemResponse: types.CommonPageWithItemResponse[ent.EdgeOneZone, any, types.EdgeOneZoneRecordRequest]{
			Item: ent.EdgeOneZone{},
			CommonPageResponse: types.CommonPageResponse[any, types.EdgeOneZoneRecordRequest]{
				Filter: requestParams,
				List:   []any{},
			},
		},
	}

	// 查询站点
	firstZone, err := global.GvaEntDb.EdgeOneZone.Query().
		Where(edgeonezone.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "未查询到站点:"+err.Error())
		return
	}
	pageData.Item = *firstZone

	if firstZone.Type == edgeonezone.TypeFull {
		allRecord, err := firstZone.QueryDNSRecords().
			OrderByRemoteCreatedAtDesc().
			Paginate(c, requestParams.Page, requestParams.PageSize)
		if err == nil {
			for _, item := range allRecord.Items {
				pageData.List = append(pageData.List, *item)
			}
		}

		pageData.Page = allRecord.Page
		pageData.PageSize = allRecord.Size
		pageData.Total = allRecord.Total
	}
	if firstZone.Type == edgeonezone.TypePartial {
		allRecord, err := firstZone.QueryAccelerationDomains().
			OrderByRemoteCreatedAtDesc().
			Paginate(c, requestParams.Page, requestParams.PageSize)
		if err == nil {
			for _, item := range allRecord.Items {
				pageData.List = append(pageData.List, *item)
			}
		}

		pageData.Page = allRecord.Page
		pageData.PageSize = allRecord.Size
		pageData.Total = allRecord.Total
	}

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "edgeone-zone-record.html"),
		types.BasePageData[types.EdgeOneZoneRecordResponse]{
			Title: "EdgeOne 站点记录",
			Data:  pageData,
		},
		"edgeone-zone",
	)
}

func EdgeOneZoneOriginGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.EdgeOneZoneOriginGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.EdgeOneZoneOriginGroupResponse{
		CommonPageWithItemResponse: types.CommonPageWithItemResponse[ent.EdgeOneZone, ent.EdgeOneZoneOriginGroup, types.EdgeOneZoneOriginGroupRequest]{
			Item: ent.EdgeOneZone{},
			CommonPageResponse: types.CommonPageResponse[ent.EdgeOneZoneOriginGroup, types.EdgeOneZoneOriginGroupRequest]{
				Filter: requestParams,
				List:   []ent.EdgeOneZoneOriginGroup{},
			},
		},
	}

	// 查询站点
	firstZone, err := global.GvaEntDb.EdgeOneZone.Query().
		Where(edgeonezone.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "未查询到站点:"+err.Error())
		return
	}
	pageData.Item = *firstZone

	allOrigingroup, err := firstZone.QueryOriginGroups().
		OrderByRemoteCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allOrigingroup.Items {
			pageData.List = append(pageData.List, *item)
		}
	}

	pageData.Page = allOrigingroup.Page
	pageData.PageSize = allOrigingroup.Size
	pageData.Total = allOrigingroup.Total

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "edgeone-zone-origin-group.html"),
		types.BasePageData[types.EdgeOneZoneOriginGroupResponse]{
			Title: "EdgeOne 站点源组",
			Data:  pageData,
		},
		"edgeone-zone",
	)
}
