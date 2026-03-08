package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/esasite"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/ent_stat"
	"vibe_ip/internal/ent_stat/esasiteanalysisdata"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gotime"
)

func ESASite(c *gin.Context) {
	// 请求参数
	var requestParams types.ESASiteRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.ESASiteResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetESASiteResponse, types.ESASiteRequest]{
			Filter: requestParams,
			List:   []types.GetESASiteResponse{},
		},
	}

	// 查询全部ESA配置
	allPlatformConfig, err := global.GvaEntDb.PlatformConfig.Query().
		Where(platformconfig.PlatformEQ(platformconfig.PlatformAliyunEsa)).
		OrderByCreatedAtAsc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allPlatformConfig.Items {
			sites := make([]ent.EsaSite, 0)
			// 查询隧道列表
			allSite, err := global.GvaEntDb.EsaSite.Query().
				Where(esasite.ConfigID(item.ID)).
				WithRecords().
				WithOriginPools().
				OrderBySiteNameAsc().
				All(c)
			if err == nil {
				for _, site := range allSite {
					sites = append(sites, *site)
				}
			}
			pageData.List = append(pageData.List, types.GetESASiteResponse{
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
		filepath.Join("pages", "esa-site.html"),
		types.BasePageData[types.ESASiteResponse]{
			Title: "ESA 站点管理",
			Data:  pageData,
		},
	)
}

func ESASiteAnalysis(c *gin.Context) {
	// 请求参数
	var requestParams types.ESASiteAnalysisRequest
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
	pageData := types.ESASiteAnalysisResponse{
		CommonPageWithItemAndStatResponse: types.CommonPageWithItemAndStatResponse[ent_stat.EsaSiteAnalysisData, ent.EsaSite, ent_stat.EsaSiteAnalysisData, types.ESASiteAnalysisRequest]{
			Stat: ent_stat.EsaSiteAnalysisData{},
			Item: ent.EsaSite{},
			CommonPageResponse: types.CommonPageResponse[ent_stat.EsaSiteAnalysisData, types.ESASiteAnalysisRequest]{
				Filter: requestParams,
				List:   []ent_stat.EsaSiteAnalysisData{},
			},
		},
	}

	// 查询站点
	firstSite, err := global.GvaEntDb.EsaSite.Query().
		Where(esasite.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "未查询到站点:"+err.Error())
		return
	}
	pageData.Item = *firstSite

	// 查询站点统计分析数据
	pageData.Stat = ent_stat.EsaSiteAnalysisData{}
	startTime := gotime.Current().StartOfDay().Time
	endTime := gotime.Current().EndOfDay().Time
	if requestParams.TimeRange == "yesterday" {
		startTime = gotime.Current().BeforeDay(1).StartOfDay().Time
		endTime = gotime.Current().BeforeDay(1).EndOfDay().Time
	}
	if requestParams.TimeRange == "7d" {
		startTime = gotime.Current().BeforeDay(7).StartOfDay().Time
		endTime = gotime.Current().EndOfDay().Time
	}
	if requestParams.TimeRange == "30d" {
		startTime = gotime.Current().BeforeDay(30).StartOfDay().Time
		endTime = gotime.Current().EndOfDay().Time
	}
	analysisData, err := global.GvaEntStatDb.EsaSiteAnalysisData.Query().
		Where(esasiteanalysisdata.IDEQ(requestParams.ID)).
		WhereIf(requestParams.TimeRange != "all", esasiteanalysisdata.DateGTE(startTime)).
		WhereIf(requestParams.TimeRange != "all", esasiteanalysisdata.DateLTE(endTime)).
		All(c)
	if err == nil {
		for _, item := range analysisData {
			pageData.Stat.Traffic += item.Traffic
			pageData.Stat.Requests += item.Requests
			pageData.Stat.RequestTraffic += item.RequestTraffic
			pageData.Stat.PageView += item.PageView
		}
	}

	// 查询站点历史分析数据
	listData, err := global.GvaEntStatDb.EsaSiteAnalysisData.Query().
		Where(esasiteanalysisdata.IDEQ(requestParams.ID)).
		OrderByDateDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range listData.Items {
			pageData.List = append(pageData.List, *item)
		}
	}

	pageData.Page = listData.Page
	pageData.PageSize = listData.Size
	pageData.Total = listData.Total

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "esa-site-analysis.html"),
		types.BasePageData[types.ESASiteAnalysisResponse]{
			Title: "ESA 站点分析",
			Data:  pageData,
		},
		"esa-site",
	)
}

func ESASiteRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.ESASiteRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.ESASiteRecordResponse{
		CommonPageWithItemResponse: types.CommonPageWithItemResponse[ent.EsaSite, any, types.ESASiteRecordRequest]{
			Item: ent.EsaSite{},
			CommonPageResponse: types.CommonPageResponse[any, types.ESASiteRecordRequest]{
				Filter: requestParams,
				List:   []any{},
			},
		},
	}

	// 查询站点
	firstSite, err := global.GvaEntDb.EsaSite.Query().
		Where(esasite.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "未查询到站点:"+err.Error())
		return
	}
	pageData.Item = *firstSite

	allRecord, err := firstSite.QueryRecords().
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

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "esa-site-record.html"),
		types.BasePageData[types.ESASiteRecordResponse]{
			Title: "ESA 站点记录",
			Data:  pageData,
		},
		"esa-site",
	)
}

func ESASiteOriginPool(c *gin.Context) {
	// 请求参数
	var requestParams types.ESASiteOriginPoolRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.ESASiteOriginPoolResponse{
		CommonPageWithItemResponse: types.CommonPageWithItemResponse[ent.EsaSite, ent.EsaSiteOriginPool, types.ESASiteOriginPoolRequest]{
			Item: ent.EsaSite{},
			CommonPageResponse: types.CommonPageResponse[ent.EsaSiteOriginPool, types.ESASiteOriginPoolRequest]{
				Filter: requestParams,
				List:   []ent.EsaSiteOriginPool{},
			},
		},
	}

	// 查询站点
	firstSite, err := global.GvaEntDb.EsaSite.Query().
		Where(esasite.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "未查询到站点:"+err.Error())
		return
	}
	pageData.Item = *firstSite

	allOriginpool, err := firstSite.QueryOriginPools().
		OrderByNameAsc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allOriginpool.Items {
			pageData.List = append(pageData.List, *item)
		}
	}

	pageData.Page = allOriginpool.Page
	pageData.PageSize = allOriginpool.Size
	pageData.Total = allOriginpool.Total

	// 渲染模板
	renderHTMLWithMenu(
		c,
		filepath.Join("pages", "esa-site-origin-pool.html"),
		types.BasePageData[types.ESASiteOriginPoolResponse]{
			Title: "ESA 站点源地址池",
			Data:  pageData,
		},
		"esa-site",
	)
}
