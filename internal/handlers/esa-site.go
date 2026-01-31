package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/esasite"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

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
		"esa",
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

	allOriginpool, err := firstSite.QueryOriginpools().
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
		"esa",
	)
}
