package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent/dnsrecord"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Dns(c *gin.Context) {
	// 请求参数
	var requestParams types.DnsRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.DnsResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetDnsRecordResponse, types.DnsRequest]{
			Filter: requestParams,
			List:   []types.GetDnsRecordResponse{},
		},
	}

	// 创建服务
	service := service.NewDNSRecordService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
		sys_env.GvaEnv.DEBUG,
	)

	// 查询全部DNS记录
	allDnsRecord, err := global.GvaEntDb.DNSRecord.Query().
		WhereIf(requestParams.ID > 0, dnsrecord.IDEQ(requestParams.ID)).
		WhereIf(requestParams.ConfigID > 0, dnsrecord.ConfigIDEQ(requestParams.ConfigID)).
		WhereIf(requestParams.RecordID != "", dnsrecord.RecordIDEQ(requestParams.RecordID)).
		WhereIf(requestParams.SubDomain != "", dnsrecord.SubDomainContains(requestParams.SubDomain)).
		WhereIf(requestParams.Type != "", dnsrecord.TypeEQ(dnsrecord.Type(requestParams.Type))).
		WhereIf(requestParams.IPGroupID > 0, dnsrecord.IPGroupID(requestParams.IPGroupID)).
		WhereIf(requestParams.CdnRecordID > 0, dnsrecord.CdnRecordID(requestParams.CdnRecordID)).
		WhereIf(!requestParams.IsActive.IsEmpty(), dnsrecord.IsActiveEQ(requestParams.IsActive.Value())).
		WithConfig().
		WithIPGroup().
		WithCdnRecord().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allDnsRecord.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allDnsRecord.Page
	pageData.PageSize = allDnsRecord.Size
	pageData.Total = allDnsRecord.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "dns.html"),
		types.BasePageData[types.DnsResponse]{
			Debug: sys_env.GvaEnv.DEBUG,
			Title: "DNS记录",
			Data:  pageData,
		},
	)
}
