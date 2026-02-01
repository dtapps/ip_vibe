package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cloudflarezone"
	"vibe_ip/internal/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
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
		WhereConfigJSONContainsFieldExistsAndNotEmpty("account_id").
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
		CommonPageWithItemResponse: types.CommonPageWithItemResponse[ent.CloudflareZone, any, types.CloudflareZoneRecordRequest]{
			Item: ent.CloudflareZone{},
			CommonPageResponse: types.CommonPageResponse[any, types.CloudflareZoneRecordRequest]{
				Filter: requestParams,
				List:   []any{},
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

	allRecord, err := firstZone.QueryRecords().
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
		filepath.Join("pages", "cloudflare-zone-record.html"),
		types.BasePageData[types.CloudflareZoneRecordResponse]{
			Title: "Cloudflare 站点记录",
			Data:  pageData,
		},
		"cloudflare-zone",
	)
}
