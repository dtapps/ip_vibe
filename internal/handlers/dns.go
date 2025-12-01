package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/dnsrecord"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Dns(c *gin.Context) {
	// 准备页面数据
	pageData := types.DnsResponse{
		List:     []types.GetDnsRecordsResponse{},
		Page:     0,
		PageSize: 0,
		Total:    0,
	}

	// 查询全部DNS记录
	allDnsRecord, err := global.GvaDb.DNSRecord.Query().
		WithConfig().
		WithIPGroup().
		WithCdnRecord().
		Order(ent.Desc(dnsrecord.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allDnsRecord {
			list := types.GetDnsRecordsResponse{
				ID:            item.ID,
				ConfigID:      item.ConfigID,
				Platform:      item.Edges.Config.Platform.String(),
				RecordID:      item.RecordID,
				Domain:        item.Domain,
				Type:          item.Type.String(),
				IsActive:      item.IsActive,
				LastRefreshAt: item.LastRefreshAt,
			}
			if ipGroup, err := item.Edges.IPGroupOrErr(); err == nil {
				list.IPGroupID = ipGroup.ID
				list.IPGroupName = ipGroup.Name
			}
			if cdnRecord, err := item.Edges.CdnRecordOrErr(); err == nil {
				list.CdnRecordID = cdnRecord.ID
				list.CdnRecordCdnDomain = cdnRecord.CdnDomain
			}
			pageData.List = append(pageData.List, list)
		}
	}

	if global.GvaEnv.DEBUG {
		fmt.Printf("dns pageData: %+v \n", pageData)
	}

	// HTTP2 server 推送
	pusher(c)

	// 渲染模板
	c.HTML(http.StatusOK, "pages/dns.html", pageData)
}
