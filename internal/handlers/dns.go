package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/ent"
	"vibe_ip/ent/dnsrecord"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Dns(c *gin.Context) {
	// 请求参数
	var requestParams types.DnsRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 准备页面数据
	pageData := types.DnsResponse{
		List:     []types.GetDnsRecordResponse{},
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
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allDnsRecord.Items {
			list := types.GetDnsRecordResponse{
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

	pageData.Page = allDnsRecord.Page
	pageData.PageSize = allDnsRecord.Size
	pageData.Total = allDnsRecord.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "dns.html"), pageData)
}
