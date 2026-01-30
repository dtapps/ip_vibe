package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/edgeonezone"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

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
		allRecord, err := firstZone.QueryDnsrecords().
			OrderByCreatedAtAsc().
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
		allRecord, err := firstZone.QueryAccelerationdomains().
			OrderByCreatedAtAsc().
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
	renderHTML(
		c,
		filepath.Join("pages", "edgeone-zone-record.html"),
		types.BasePageData[types.EdgeOneZoneRecordResponse]{
			Title: "EdgeOne 站点记录",
			Data:  pageData,
		},
	)
}
