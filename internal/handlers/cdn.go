package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cdnrecord"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Cdn(c *gin.Context) {
	// 请求参数
	var requestParams types.CdnRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.CdnResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetCdnRecordResponse]{
			List: []types.GetCdnRecordResponse{},
		},
	}

	// 查询全部CDN记录
	allCdnRecord, err := global.GvaEntDb.CDNRecord.Query().
		WithConfig().
		WithIPGroup().
		Order(ent.Desc(cdnrecord.FieldCreatedAt)).
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allCdnRecord.Items {
			list := types.GetCdnRecordResponse{
				ID:             item.ID,
				ConfigID:       item.ConfigID,
				Platform:       item.Edges.Config.Platform.String(),
				RecordID:       item.RecordID,
				Domain:         item.Domain,
				Type:           item.Type.String(),
				OriginProtocol: item.OriginProtocol.String(),
				HTTPPort:       item.HTTPPort,
				HTTPSPort:      item.HTTPSPort,
				CdnDomain:      item.CdnDomain,
				IsActive:       item.IsActive,
				LastRefreshAt:  item.LastRefreshAt,
			}
			if ipGroup, err := item.Edges.IPGroupOrErr(); err == nil {
				list.IPGroupID = ipGroup.ID
				list.IPGroupName = ipGroup.Name
			}
			pageData.List = append(pageData.List, list)
		}
	}

	pageData.Page = allCdnRecord.Page
	pageData.PageSize = allCdnRecord.Size
	pageData.Total = allCdnRecord.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "cdn.html"), pageData)
}
