package handlers

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cloudflarezone"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent_stat"
	"vibe_ip/internal/ent_stat/cloudflarezoneanalysisdata"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gotime"
)

func CloudflareZone(c *gin.Context) {
	// 请求参数
	var requestParams types.CloudflareZoneRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.CloudflareZoneResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetCloudflareZoneResponse, types.CloudflareZoneRequest]{
			Filter: requestParams,
			List:   []types.GetCloudflareZoneResponse{},
		},
	}

	// 查询全部Cloudflare配置
	allPlatformConfig, err := global.GvaEntDb.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformCloudflare)).
		Where(platformconfig.CapabilityIn(
			platformconfig.CapabilityDNS,
			platformconfig.CapabilityCdn,
			platformconfig.CapabilityAll,
		)).
		// WhereConfigJSONContainsFieldExistsAndNotEmpty("account_id").
		OrderByCreatedAtAsc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allPlatformConfig.Items {
			zones := make([]ent.CloudflareZone, 0)
			// 查询站点列表
			allZone, err := global.GvaEntDb.CloudflareZone.Query().
				Where(cloudflarezone.ConfigID(item.ID)).
				WithRecords().
				All(c)
			if err == nil {
				for _, zone := range allZone {
					zones = append(zones, *zone)
				}
			} else {
				slog.ErrorContext(c, "查询全部Cloudflare站点",
					slog.Any("request", requestParams),
					slog.Any("err", err),
				)
			}
			pageData.List = append(pageData.List, types.GetCloudflareZoneResponse{
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
		filepath.Join("pages", "cloudflare-zone.html"),
		types.BasePageData[types.CloudflareZoneResponse]{
			Title: "Cloudflare 站点管理",
			Data:  pageData,
		},
	)
}

func CloudflareZoneAnalysis(c *gin.Context) {
	// 请求参数
	var requestParams types.CloudflareZoneAnalysisRequest
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
	pageData := types.CloudflareZoneAnalysisResponse{
		CommonPageWithItemAndStatResponse: types.CommonPageWithItemAndStatResponse[ent_stat.CloudflareZoneAnalysisData, ent.CloudflareZone, ent_stat.CloudflareZoneAnalysisData, types.CloudflareZoneAnalysisRequest]{
			Stat: ent_stat.CloudflareZoneAnalysisData{},
			Item: ent.CloudflareZone{},
			CommonPageResponse: types.CommonPageResponse[ent_stat.CloudflareZoneAnalysisData, types.CloudflareZoneAnalysisRequest]{
				Filter: requestParams,
				List:   []ent_stat.CloudflareZoneAnalysisData{},
			},
		},
	}

	// 查询站点
	firstZone, err := global.GvaEntDb.CloudflareZone.Query().
		Where(cloudflarezone.IDEQ(requestParams.ID)).
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
	analysisData, err := global.GvaEntStatDb.CloudflareZoneAnalysisData.Query().
		Where(cloudflarezoneanalysisdata.CloudflareZoneIDEQ(firstZone.ID)).
		WhereIf(requestParams.TimeRange != "all", cloudflarezoneanalysisdata.DateGTE(startTime)).
		WhereIf(requestParams.TimeRange != "all", cloudflarezoneanalysisdata.DateLTE(endTime)).
		OrderByDateAsc().
		All(c)
	if err == nil {
		for _, item := range analysisData {
			pageData.Stat.Requests += item.Requests
			pageData.Stat.CachedRequests += item.CachedRequests
			pageData.Stat.UncachedRequests += item.UncachedRequests
			pageData.Stat.Bandwidth += item.Bandwidth
			pageData.Stat.CachedBandwidth += item.CachedBandwidth
			pageData.Stat.UncachedBandwidth += item.UncachedBandwidth
			pageData.Stat.Uniques += item.Uniques
		}
	} else {
		slog.ErrorContext(c, "查询站点统计分析数据",
			slog.Any("request", requestParams),
			slog.Int64("id", firstZone.ID),
			slog.Any("err", err),
		)
	}

	// 查询站点历史分析数据
	listData, err := global.GvaEntStatDb.CloudflareZoneAnalysisData.Query().
		Where(cloudflarezoneanalysisdata.CloudflareZoneIDEQ(firstZone.ID)).
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
		filepath.Join("pages", "cloudflare-zone-analysis.html"),
		types.BasePageData[types.CloudflareZoneAnalysisResponse]{
			Title: "Cloudflare 站点记录",
			Data:  pageData,
		},
		"cloudflare-zone",
	)
}

func CloudflareZoneRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.CloudflareZoneRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.CloudflareZoneRecordResponse{
		CommonPageWithItemResponse: types.CommonPageWithItemResponse[ent.CloudflareZone, ent.CloudflareZoneRecord, types.CloudflareZoneRecordRequest]{
			Item: ent.CloudflareZone{},
			CommonPageResponse: types.CommonPageResponse[ent.CloudflareZoneRecord, types.CloudflareZoneRecordRequest]{
				Filter: requestParams,
				List:   []ent.CloudflareZoneRecord{},
			},
		},
	}

	// 查询站点
	firstZone, err := global.GvaEntDb.CloudflareZone.Query().
		Where(cloudflarezone.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "未查询到站点:"+err.Error())
		return
	}
	pageData.Item = *firstZone

	// 查询站点DNS记录
	allRecord, err := firstZone.QueryRecords().
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

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "cloudflare-zone-record.html"),
		types.BasePageData[types.CloudflareZoneRecordResponse]{
			Title: "Cloudflare 站点记录",
			Data:  pageData,
		},
		"cloudflare-zone",
	)
}
