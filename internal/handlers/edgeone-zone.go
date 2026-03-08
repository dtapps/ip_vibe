package handlers

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/edgeonezone"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent_stat"
	"vibe_ip/internal/ent_stat/edgeonezoneanalysisdata"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gotime"
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
			zones := make([]ent.EdgeOneZone, 0)
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
					zones = append(zones, *zone)
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
	} else {
		slog.ErrorContext(c, "查询全部EdgeOne配置",
			slog.Any("request", requestParams),
			slog.String("platform", platformconfig.PlatformAliyunEsa.String()),
			slog.Any("err", err),
		)
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

func EdgeOneZoneAnalysis(c *gin.Context) {
	// 请求参数
	var requestParams types.EdgeOneZoneAnalysisRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 时间范围
	if requestParams.TimeRange == "" {
		requestParams.TimeRange = "today"
	}

	// 准备页面数据
	pageData := types.EdgeOneZoneAnalysisResponse{
		CommonPageWithItemAndStatResponse: types.CommonPageWithItemAndStatResponse[ent_stat.EdgeOneZoneAnalysisData, ent.EdgeOneZone, ent_stat.EdgeOneZoneAnalysisData, types.EdgeOneZoneAnalysisRequest]{
			Stat: ent_stat.EdgeOneZoneAnalysisData{},
			Item: ent.EdgeOneZone{},
			CommonPageResponse: types.CommonPageResponse[ent_stat.EdgeOneZoneAnalysisData, types.EdgeOneZoneAnalysisRequest]{
				Filter: requestParams,
				List:   []ent_stat.EdgeOneZoneAnalysisData{},
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

	// 查询站点统计分析数据
	startTime := gotime.Current().StartOfDay().Time
	endTime := gotime.Current().EndOfDay().Time
	switch requestParams.TimeRange {
	case "today":
		startTime = gotime.Current().StartOfDay().Time
		endTime = gotime.Current().EndOfDay().Time
	case "yesterday":
		startTime = gotime.Current().BeforeDay(1).StartOfDay().Time
		endTime = gotime.Current().BeforeDay(1).EndOfDay().Time
	case "before_yesterday":
		startTime = gotime.Current().BeforeDay(2).StartOfDay().Time
		endTime = gotime.Current().BeforeDay(2).EndOfDay().Time
	case "7d":
		startTime = gotime.Current().BeforeDay(7).StartOfDay().Time
		endTime = gotime.Current().EndOfDay().Time
	case "15d":
		startTime = gotime.Current().BeforeDay(15).StartOfDay().Time
		endTime = gotime.Current().EndOfDay().Time
	case "30d":
		startTime = gotime.Current().BeforeDay(30).StartOfDay().Time
		endTime = gotime.Current().EndOfDay().Time
	}
	analysisData, err := global.GvaEntStatDb.EdgeOneZoneAnalysisData.Query().
		Where(edgeonezoneanalysisdata.EdgeoneZoneIDEQ(firstZone.ID)).
		WhereIf(requestParams.TimeRange != "all", edgeonezoneanalysisdata.DateGTE(startTime)).
		WhereIf(requestParams.TimeRange != "all", edgeonezoneanalysisdata.DateLTE(endTime)).
		OrderByDateAsc().
		All(c)
	if err == nil {
		for _, item := range analysisData {
			pageData.Stat.L7FlowOutFlux += item.L7FlowOutFlux
			pageData.Stat.L7FlowInFlux += item.L7FlowInFlux
			pageData.Stat.L7FlowFlux += item.L7FlowFlux
			pageData.Stat.L7FlowOutBandwidth += item.L7FlowOutBandwidth
			pageData.Stat.L7FlowInBandwidth += item.L7FlowInBandwidth
			pageData.Stat.L7FlowBandwidth += item.L7FlowBandwidth
			pageData.Stat.L7FlowRequest += item.L7FlowRequest
			pageData.Stat.L7FlowAvgResponseTime += item.L7FlowAvgResponseTime
			pageData.Stat.L7FlowAvgFirstByteResponseTime += item.L7FlowAvgFirstByteResponseTime
			pageData.Stat.L4FlowFlux += item.L4FlowFlux
			pageData.Stat.L4FlowInFlux += item.L4FlowInFlux
			pageData.Stat.L4FlowOutFlux += item.L4FlowOutFlux
			pageData.Stat.L4FlowInBandwidth += item.L4FlowInBandwidth
			pageData.Stat.L4FlowOutBandwidth += item.L4FlowOutBandwidth
			pageData.Stat.L4FlowConnections += item.L4FlowConnections
			pageData.Stat.L4FlowNewConnectionsRate += item.L4FlowNewConnectionsRate
		}
	} else {
		slog.ErrorContext(c, "查询站点统计分析数据",
			slog.Any("request", requestParams),
			slog.Int64("id", firstZone.ID),
			slog.Any("err", err),
		)
	}

	// 查询站点历史分析数据
	listData, err := global.GvaEntStatDb.EdgeOneZoneAnalysisData.Query().
		Where(edgeonezoneanalysisdata.EdgeoneZoneIDEQ(firstZone.ID)).
		OrderByDateDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range listData.Items {
			pageData.List = append(pageData.List, *item)
		}
	} else {
		slog.ErrorContext(c, "查询站点历史分析数据",
			slog.Any("request", requestParams),
			slog.Int64("id", firstZone.ID),
			slog.Any("err", err),
		)
	}

	pageData.Page = listData.Page
	pageData.PageSize = listData.Size
	pageData.Total = listData.Total

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "edgeone-zone-analysis.html"),
		types.BasePageData[types.EdgeOneZoneAnalysisResponse]{
			Title: "EdgeOne 站点分析",
			Data:  pageData,
		},
		"edgeone-zone",
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
		// 查询站点DNS记录
		allRecord, err := firstZone.QueryDNSRecords().
			OrderByRemoteCreatedAtDesc().
			Paginate(c, requestParams.Page, requestParams.PageSize)
		if err == nil {
			for _, item := range allRecord.Items {
				pageData.List = append(pageData.List, *item)
			}
		} else {
			slog.ErrorContext(c, "查询站点DNS记录",
				slog.Any("request", requestParams),
				slog.Any("err", err),
			)
		}

		pageData.Page = allRecord.Page
		pageData.PageSize = allRecord.Size
		pageData.Total = allRecord.Total
	}
	if firstZone.Type == edgeonezone.TypePartial {
		// 查询站点加速域名
		allRecord, err := firstZone.QueryAccelerationDomains().
			OrderByRemoteCreatedAtDesc().
			Paginate(c, requestParams.Page, requestParams.PageSize)
		if err == nil {
			for _, item := range allRecord.Items {
				pageData.List = append(pageData.List, *item)
			}
		} else {
			slog.ErrorContext(c, "查询站点加速域名",
				slog.Any("request", requestParams),
				slog.Any("err", err),
			)
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

	// 查询站点源站组
	allOrigingroup, err := firstZone.QueryOriginGroups().
		OrderByRemoteCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allOrigingroup.Items {
			pageData.List = append(pageData.List, *item)
		}
	} else {
		slog.ErrorContext(c, "查询站点源站组",
			slog.Any("request", requestParams),
			slog.Any("err", err),
		)
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
