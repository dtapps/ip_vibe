package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/cdnrecord"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Cdn(c *gin.Context) {
	// 准备页面数据
	pageData := types.CdnResponse{
		List:     []types.GetCdnRecordsResponse{},
		Page:     0,
		PageSize: 0,
		Total:    0,
	}

	// 查询全部CDN记录
	allCdnRecord, err := global.GvaDb.CDNRecord.Query().
		WithConfig().
		WithIPGroup().
		Order(ent.Desc(cdnrecord.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allCdnRecord {
			list := types.GetCdnRecordsResponse{
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

	if global.GvaEnv.DEBUG {
		fmt.Printf("cdn pageData: %+v \n", pageData)
	}

	// HTTP2 server 推送
	pusher(c)

	// 渲染模板
	c.HTML(http.StatusOK, "pages/cdn.html", pageData)
}
